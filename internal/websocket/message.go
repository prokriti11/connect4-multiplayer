// Package websocket defines message types for client-server communication.
package websocket

// Message types for WebSocket protocol
const (
	MsgJoinQueue          = "join_queue"
	MsgGameStart          = "game_start"
	MsgMove               = "move"
	MsgGameState          = "game_state"
	MsgGameOver           = "game_over"
	MsgError              = "error"
	MsgOpponentDisconnect = "opponent_disconnect"
	MsgOpponentReconnect  = "opponent_reconnect"
)

// IncomingMessage represents messages from client to server
type IncomingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// OutgoingMessage represents messages from server to client
type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// JoinQueuePayload - client requests to join matchmaking
type JoinQueuePayload struct {
	Username string `json:"username"`
}

// MovePayload - client makes a move
type MovePayload struct {
	GameID string `json:"game_id"`
	Column int    `json:"column"`
}

// GameStartPayload - server notifies game has started
type GameStartPayload struct {
	GameID       string `json:"game_id"`
	OpponentName string `json:"opponent_name"`
	IsBot        bool   `json:"is_bot"`
	Symbol       int    `json:"symbol"` // 1=Red, 2=Yellow
	YourTurn     bool   `json:"your_turn"`
}

// GameStatePayload - server sends current game state
type GameStatePayload struct {
	GameID      string      `json:"game_id"`
	Board       interface{} `json:"board"`
	Turn        int         `json:"turn"`
	Winner      int         `json:"winner"`
	Status      string      `json:"status"`
	Player1Name string      `json:"player1_name"`
	Player2Name string      `json:"player2_name"`
	LastMove    *LastMove   `json:"last_move,omitempty"`
}

// LastMove contains info about the most recent move
type LastMove struct {
	Row    int `json:"row"`
	Col    int `json:"col"`
	Player int `json:"player"`
}

// ErrorPayload - server sends an error
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// DisconnectPayload - notify about opponent disconnect
type DisconnectPayload struct {
	Message       string `json:"message"`
	TimeRemaining int    `json:"time_remaining"` // seconds
}
