// handlers.go implements WebSocket message handling and client connection management.
package websocket

import (
	"encoding/json"
	"log"
	"time"

	"connect4/internal/game"
	"connect4/internal/matchmaking"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message
	pongWait = 60 * time.Second

	// Send pings with this period (must be less than pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size
	maxMessageSize = 1024
)

// Client represents a single WebSocket connection
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	ID       string
	Username string
	GameID   string
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		c.handleMessage(message)
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Drain queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage routes incoming messages to appropriate handlers
func (c *Client) handleMessage(rawMessage []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		log.Printf("Error parsing message: %v", err)
		c.sendError("Invalid message format")
		return
	}

	msgType, _ := msg["type"].(string)
	payload, _ := msg["payload"].(map[string]interface{})

	log.Printf("Received %s from %s", msgType, c.ID)

	switch msgType {
	case MsgJoinQueue:
		c.handleJoinQueue(payload)
	case MsgMove:
		c.handleMove(payload)
	default:
		c.sendError("Unknown message type: " + msgType)
	}
}

// handleJoinQueue processes matchmaking requests
func (c *Client) handleJoinQueue(payload map[string]interface{}) {
	username, _ := payload["username"].(string)
	if username == "" {
		c.sendError("Username is required")
		return
	}

	c.Username = username

	// Check for reconnection to existing game
	if existingGame, found := c.hub.gameStore.GetGameByUsername(username); found {
		c.handleReconnection(existingGame)
		return
	}

	// Join matchmaking queue
	resultChan := c.hub.matchmaker.Add(c.ID, username)

	go func() {
		result := <-resultChan
		c.onMatchFound(result)
	}()
}

// onMatchFound handles successful matchmaking
func (c *Client) onMatchFound(result matchmaking.MatchResult) {
	// Create or join game
	g, exists := c.hub.gameStore.GetGame(result.GameID)

	if !exists {
		// We are the first player - create the game
		g = game.NewGame(
			c.ID, c.Username,
			result.OpponentID, result.OpponentName,
			result.IsBot,
		)
		c.hub.gameStore.CreateGame(g)
	} else {
		// We are the second player - update our info
		c.hub.gameStore.UpdatePlayerID(result.GameID, c.Username, c.ID)
	}

	c.GameID = result.GameID

	// Send game start notification
	msg := OutgoingMessage{
		Type: MsgGameStart,
		Payload: GameStartPayload{
			GameID:       result.GameID,
			OpponentName: result.OpponentName,
			IsBot:        result.IsBot,
			Symbol:       result.Symbol,
			YourTurn:     result.Symbol == 1, // Red goes first
		},
	}

	msgBytes, _ := json.Marshal(msg)
	c.send <- msgBytes

	// Send initial game state
	c.hub.BroadcastGameState(g, nil)

	log.Printf("Game started for %s: %s (Symbol %d)", c.Username, result.GameID, result.Symbol)
}

// handleReconnection restores a player to their existing game
func (c *Client) handleReconnection(g *game.Game) {
	log.Printf("Reconnecting %s to game %s", c.Username, g.ID)

	// Update player connection ID
	c.hub.gameStore.UpdatePlayerID(g.ID, c.Username, c.ID)
	c.hub.gameStore.ClearDisconnect(c.ID)
	c.GameID = g.ID

	// Determine player info
	var symbol int
	var opponentName string
	isBot := false

	if g.Player1.Username == c.Username {
		symbol = 1
		opponentName = g.Player2.Username
		isBot = g.Player2.IsBot
	} else {
		symbol = 2
		opponentName = g.Player1.Username
	}

	// Send game start (reconnection)
	startMsg := OutgoingMessage{
		Type: MsgGameStart,
		Payload: GameStartPayload{
			GameID:       g.ID,
			OpponentName: opponentName,
			IsBot:        isBot,
			Symbol:       symbol,
			YourTurn:     int(g.Turn) == symbol,
		},
	}
	startBytes, _ := json.Marshal(startMsg)
	c.send <- startBytes

	// Send current game state
	state := g.GetState()
	stateMsg := OutgoingMessage{
		Type: MsgGameState,
		Payload: GameStatePayload{
			GameID:      g.ID,
			Board:       state.Board,
			Turn:        int(state.Turn),
			Winner:      int(state.Winner),
			Status:      string(state.Status),
			Player1Name: state.Player1Name,
			Player2Name: state.Player2Name,
		},
	}
	stateBytes, _ := json.Marshal(stateMsg)
	c.send <- stateBytes

	// Notify opponent about reconnection
	reconnectMsg := OutgoingMessage{
		Type: MsgOpponentReconnect,
		Payload: map[string]string{
			"message": "Opponent has reconnected!",
		},
	}
	reconnectBytes, _ := json.Marshal(reconnectMsg)

	c.hub.clientsMu.RLock()
	for client := range c.hub.clients {
		if client.GameID == g.ID && client.ID != c.ID {
			select {
			case client.send <- reconnectBytes:
			default:
			}
		}
	}
	c.hub.clientsMu.RUnlock()
}

// handleMove processes a player's move
func (c *Client) handleMove(payload map[string]interface{}) {
	gameID, _ := payload["game_id"].(string)
	colFloat, _ := payload["column"].(float64)
	col := int(colFloat)

	// Validate request
	if gameID == "" || gameID != c.GameID {
		c.sendError("Invalid game ID")
		return
	}

	// Get game
	g, exists := c.hub.gameStore.GetGame(gameID)
	if !exists {
		c.sendError("Game not found")
		return
	}

	// Make the move (game engine validates turn order, column, etc.)
	result := g.MakeMove(c.ID, col)

	if !result.Valid {
		c.sendError(result.Error)
		return
	}

	// Broadcast updated state
	c.hub.BroadcastGameState(g, &LastMove{
		Row:    result.Row,
		Col:    result.Col,
		Player: int(g.GetPlayerByID(c.ID).Symbol),
	})

	// Handle game over
	if result.GameOver {
		c.hub.handleGameOver(g)
		return
	}

	// Trigger bot move if applicable
	if g.IsBotTurn() {
		c.hub.TriggerBotMove(g)
	}
}

// sendError sends an error message to the client
func (c *Client) sendError(message string) {
	msg := OutgoingMessage{
		Type: MsgError,
		Payload: ErrorPayload{
			Message: message,
		},
	}
	msgBytes, _ := json.Marshal(msg)
	c.send <- msgBytes
}
