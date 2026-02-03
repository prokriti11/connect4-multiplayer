// Package websocket provides the WebSocket server implementation.
// hub.go contains the central Hub that manages all client connections.
package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"connect4/internal/game"
	"connect4/internal/leaderboard"
	"connect4/internal/matchmaking"
	"connect4/internal/state"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins (configure for production)
	},
}

// Hub maintains the set of active clients and manages game coordination
type Hub struct {
	// Thread-safe client registry
	clients   map[*Client]bool
	clientsMu sync.RWMutex

	// Channels for client lifecycle
	register   chan *Client
	unregister chan *Client

	// Core components
	matchmaker  *matchmaking.Queue
	gameStore   *state.Store
	leaderboard *leaderboard.Service

	// Broadcast channel for game state updates
	broadcast chan *BroadcastMessage
}

// BroadcastMessage contains a message to send to specific clients
type BroadcastMessage struct {
	GameID  string
	Message []byte
}

// NewHub creates a new Hub instance
func NewHub(mm *matchmaking.Queue, gs *state.Store, lb *leaderboard.Service) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		matchmaker:  mm,
		gameStore:   gs,
		leaderboard: lb,
		broadcast:   make(chan *BroadcastMessage, 256),
	}
}

// Run starts the Hub's main event loop
func (h *Hub) Run() {
	// Start timeout checker
	h.gameStore.StartTimeoutChecker(h.handleForfeit)

	for {
		select {
		case client := <-h.register:
			h.clientsMu.Lock()
			h.clients[client] = true
			h.clientsMu.Unlock()
			log.Printf("Client registered: %s", client.ID)

		case client := <-h.unregister:
			h.handleClientDisconnect(client)

		case msg := <-h.broadcast:
			h.broadcastToGame(msg.GameID, msg.Message)
		}
	}
}

// handleClientDisconnect processes a client leaving
func (h *Hub) handleClientDisconnect(client *Client) {
	h.clientsMu.Lock()
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
	}
	h.clientsMu.Unlock()

	log.Printf("Client disconnected: %s (%s)", client.ID, client.Username)

	// Check if client was in a game
	if client.GameID != "" {
		g, exists := h.gameStore.GetGame(client.GameID)
		if exists && g.Status == game.StatusActive {
			// Record disconnect and notify opponent
			h.gameStore.RecordDisconnect(client.ID, client.GameID)
			h.notifyOpponentDisconnect(g, client.ID)
		}
	}

	// Remove from matchmaking queue if waiting
	h.matchmaker.RemovePlayer(client.ID)
}

// notifyOpponentDisconnect sends disconnect notification to the other player
func (h *Hub) notifyOpponentDisconnect(g *game.Game, disconnectedID string) {
	msg := OutgoingMessage{
		Type: MsgOpponentDisconnect,
		Payload: DisconnectPayload{
			Message:       "Opponent disconnected. Waiting for reconnection...",
			TimeRemaining: 30,
		},
	}

	msgBytes, _ := json.Marshal(msg)

	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for client := range h.clients {
		if client.GameID == g.ID && client.ID != disconnectedID {
			select {
			case client.send <- msgBytes:
			default:
			}
		}
	}
}

// handleForfeit is called when a game is forfeited due to timeout
func (h *Hub) handleForfeit(g *game.Game) {
	state := g.GetState()

	// Record win for the player who stayed and loss for who left
	if h.leaderboard != nil {
		if state.Winner == game.Player1 {
			h.leaderboard.RecordWin(g.Player1.Username)
			if !g.Player2.IsBot {
				h.leaderboard.RecordLoss(g.Player2.Username)
			}
		} else if state.Winner == game.Player2 && !g.Player2.IsBot {
			h.leaderboard.RecordWin(g.Player2.Username)
			h.leaderboard.RecordLoss(g.Player1.Username)
		}
	}

	// Broadcast final state
	msg := OutgoingMessage{
		Type: MsgGameOver,
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

	msgBytes, _ := json.Marshal(msg)
	h.broadcastToGame(g.ID, msgBytes)

	// Clean up game
	h.gameStore.RemoveGame(g.ID)
}

// broadcastToGame sends a message to all clients in a specific game
func (h *Hub) broadcastToGame(gameID string, message []byte) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for client := range h.clients {
		if client.GameID == gameID {
			select {
			case client.send <- message:
			default:
				// Client buffer full, skip
			}
		}
	}
}

// BroadcastGameState sends current game state to all players in the game
func (h *Hub) BroadcastGameState(g *game.Game, lastMove *LastMove) {
	state := g.GetState()

	msg := OutgoingMessage{
		Type: MsgGameState,
		Payload: GameStatePayload{
			GameID:      g.ID,
			Board:       state.Board,
			Turn:        int(state.Turn),
			Winner:      int(state.Winner),
			Status:      string(state.Status),
			Player1Name: state.Player1Name,
			Player2Name: state.Player2Name,
			LastMove:    lastMove,
		},
	}

	msgBytes, _ := json.Marshal(msg)

	h.broadcast <- &BroadcastMessage{
		GameID:  g.ID,
		Message: msgBytes,
	}
}

// GetClientByID finds a client by their ID
func (h *Hub) GetClientByID(clientID string) *Client {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for client := range h.clients {
		if client.ID == clientID {
			return client
		}
	}
	return nil
}

// ServeWs handles WebSocket upgrade requests
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		ID:   uuid.New().String(),
	}

	hub.register <- client

	// Start client goroutines
	go client.writePump()
	go client.readPump()
}

// TriggerBotMove initiates a bot move after a short delay
func (h *Hub) TriggerBotMove(g *game.Game) {
	if !g.IsBotTurn() {
		return
	}

	go func() {
		time.Sleep(700 * time.Millisecond) // Small delay for realism

		if !g.IsBotTurn() {
			return
		}

		col := game.GetBotMove(g)
		result := g.MakeMove(g.Player2.ID, col)

		if result.Valid {
			h.BroadcastGameState(g, &LastMove{
				Row:    result.Row,
				Col:    result.Col,
				Player: int(game.Yellow),
			})

			// Check for game over
			if result.GameOver {
				h.handleGameOver(g)
			}
		}
	}()
}

// handleGameOver processes game completion
func (h *Hub) handleGameOver(g *game.Game) {
	state := g.GetState()

	// Record game results for both players
	if h.leaderboard != nil {
		switch state.Winner {
		case game.Player1:
			// Player 1 wins
			h.leaderboard.RecordWin(g.Player1.Username)
			if !g.Player2.IsBot {
				h.leaderboard.RecordLoss(g.Player2.Username)
			}
		case game.Player2:
			// Player 2 wins
			if !g.Player2.IsBot {
				h.leaderboard.RecordWin(g.Player2.Username)
			}
			h.leaderboard.RecordLoss(g.Player1.Username)
		case game.Draw:
			// Draw - both players get draw
			h.leaderboard.RecordDraw(g.Player1.Username)
			if !g.Player2.IsBot {
				h.leaderboard.RecordDraw(g.Player2.Username)
			}
		}
	}

	// Remove from active games after a delay
	go func() {
		time.Sleep(5 * time.Second)
		h.gameStore.RemoveGame(g.ID)
	}()
}
