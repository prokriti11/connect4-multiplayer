// engine.go provides the high-level game engine that orchestrates game flow.
// It handles turn management, move validation, and game lifecycle.
package game

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrGameNotActive   = errors.New("game is not active")
	ErrNotYourTurn     = errors.New("not your turn")
	ErrInvalidColumn   = errors.New("invalid column")
	ErrColumnFull      = errors.New("column is full")
	ErrPlayerNotInGame = errors.New("player not in this game")
)

// NewGame creates a new game session between two players
func NewGame(p1ID, p1Name string, p2ID, p2Name string, p2IsBot bool) *Game {
	now := time.Now()
	return &Game{
		ID:    uuid.New().String(),
		Board: NewBoard(),
		Player1: Player{
			ID:       p1ID,
			Username: p1Name,
			Symbol:   Red,
			IsBot:    false,
		},
		Player2: Player{
			ID:       p2ID,
			Username: p2Name,
			Symbol:   Yellow,
			IsBot:    p2IsBot,
		},
		Turn:            Red, // Player 1 always starts
		Status:          StatusActive,
		Winner:          NoWinner,
		CreatedAt:       now,
		UpdatedAt:       now,
		Player1LastSeen: now,
		Player2LastSeen: now,
	}
}

// MakeMove attempts to play a disc in the specified column.
// Thread-safe: acquires lock for the duration of the move.
func (g *Game) MakeMove(playerID string, col int) MoveResult {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Validate game state
	if g.Status != StatusActive {
		return MoveResult{Valid: false, Error: ErrGameNotActive.Error()}
	}

	// Identify the player
	var player *Player
	if g.Player1.ID == playerID {
		player = &g.Player1
	} else if g.Player2.ID == playerID {
		player = &g.Player2
	} else {
		return MoveResult{Valid: false, Error: ErrPlayerNotInGame.Error()}
	}

	// Validate turn
	if g.Turn != player.Symbol {
		return MoveResult{Valid: false, Error: ErrNotYourTurn.Error()}
	}

	// Validate column
	if col < 0 || col >= Cols {
		return MoveResult{Valid: false, Error: ErrInvalidColumn.Error()}
	}

	if !g.Board.CanDrop(col) {
		return MoveResult{Valid: false, Error: ErrColumnFull.Error()}
	}

	// Execute the move
	row := g.Board.Drop(col, player.Symbol)
	g.UpdatedAt = time.Now()

	// Check for win
	winner := g.Board.CheckWinAt(row, col)
	gameOver := false

	if winner != Empty {
		gameOver = true
		g.Status = StatusFinished
		if winner == Red {
			g.Winner = Player1
		} else {
			g.Winner = Player2
		}
	} else if g.Board.IsFull() {
		gameOver = true
		g.Status = StatusFinished
		g.Winner = Draw
	} else {
		// Switch turn
		g.switchTurn()
	}

	return MoveResult{
		Valid:    true,
		Row:      row,
		Col:      col,
		GameOver: gameOver,
		Winner:   g.Winner,
		Board:    g.Board,
		NextTurn: g.Turn,
	}
}

// switchTurn alternates between players
func (g *Game) switchTurn() {
	if g.Turn == Red {
		g.Turn = Yellow
	} else {
		g.Turn = Red
	}
}

// GetCurrentPlayer returns the player whose turn it is
func (g *Game) GetCurrentPlayer() *Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.Turn == Red {
		return &g.Player1
	}
	return &g.Player2
}

// GetPlayerByID returns the player with matching ID
func (g *Game) GetPlayerByID(playerID string) *Player {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if g.Player1.ID == playerID {
		return &g.Player1
	}
	if g.Player2.ID == playerID {
		return &g.Player2
	}
	return nil
}

// UpdateLastSeen updates the last seen timestamp for a player
func (g *Game) UpdateLastSeen(playerID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now()
	if g.Player1.ID == playerID {
		g.Player1LastSeen = now
	} else if g.Player2.ID == playerID {
		g.Player2LastSeen = now
	}
}

// Forfeit ends the game due to player timeout/disconnect
func (g *Game) Forfeit(loserID string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Status != StatusActive {
		return
	}

	g.Status = StatusForfeit
	if g.Player1.ID == loserID {
		g.Winner = Player2
	} else {
		g.Winner = Player1
	}
	g.UpdatedAt = time.Now()
}

// GetState returns a snapshot of the current game state (thread-safe)
func (g *Game) GetState() GameState {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return GameState{
		ID:          g.ID,
		Board:       g.Board,
		Turn:        g.Turn,
		Status:      g.Status,
		Winner:      g.Winner,
		Player1Name: g.Player1.Username,
		Player2Name: g.Player2.Username,
		IsP2Bot:     g.Player2.IsBot,
	}
}

// GameState is a read-only snapshot for broadcasting
type GameState struct {
	ID          string     `json:"game_id"`
	Board       Board      `json:"board"`
	Turn        Cell       `json:"turn"`
	Status      GameStatus `json:"status"`
	Winner      WinResult  `json:"winner"`
	Player1Name string     `json:"player1_name"`
	Player2Name string     `json:"player2_name"`
	IsP2Bot     bool       `json:"is_bot"`
}

// IsBotTurn returns true if it's currently the bot's turn
func (g *Game) IsBotTurn() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return g.Status == StatusActive && g.Player2.IsBot && g.Turn == Yellow
}

// IsActive returns true if the game is still in progress
func (g *Game) IsActive() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Status == StatusActive
}

// UpdatePlayerID updates a player's connection ID (for reconnection)
// Returns the old player ID if update was successful, empty string otherwise
func (g *Game) UpdatePlayerID(username, newPlayerID string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.Player1.Username == username {
		oldID := g.Player1.ID
		g.Player1.ID = newPlayerID
		return oldID
	}

	if g.Player2.Username == username {
		oldID := g.Player2.ID
		g.Player2.ID = newPlayerID
		return oldID
	}

	return ""
}
