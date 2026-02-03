// Package game provides the core Connect 4 game logic.
// This file defines all shared types and constants used across the game engine.
package game

import (
	"sync"
	"time"
)

// Board dimensions - standard Connect 4 is 7 columns × 6 rows
const (
	Cols = 7
	Rows = 6
)

// Cell represents the state of a single board cell
type Cell int

const (
	Empty  Cell = 0
	Red    Cell = 1 // Player 1
	Yellow Cell = 2 // Player 2 / Bot
)

// GameStatus represents the current state of a game
type GameStatus string

const (
	StatusWaiting  GameStatus = "waiting"
	StatusActive   GameStatus = "active"
	StatusFinished GameStatus = "finished"
	StatusForfeit  GameStatus = "forfeit"
)

// WinResult represents the outcome of a game
type WinResult int

const (
	NoWinner WinResult = 0
	Player1  WinResult = 1
	Player2  WinResult = 2
	Draw     WinResult = 3
)

// Board represents the 7×6 game grid
// Indexed as Board[row][col] where row 0 is top, row 5 is bottom
type Board [Rows][Cols]Cell

// Player holds information about a game participant
type Player struct {
	ID       string
	Username string
	Symbol   Cell
	IsBot    bool
}

// Game represents an active game session with all state
type Game struct {
	ID        string
	Board     Board
	Player1   Player
	Player2   Player
	Turn      Cell // Whose turn (Red or Yellow)
	Status    GameStatus
	Winner    WinResult
	CreatedAt time.Time
	UpdatedAt time.Time

	// Reconnection tracking
	Player1LastSeen time.Time
	Player2LastSeen time.Time

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// Move represents a player's move
type Move struct {
	GameID   string
	PlayerID string
	Column   int
	Row      int // Filled after drop
	Player   Cell
}

// MoveResult contains the outcome of a move attempt
type MoveResult struct {
	Valid    bool
	Row      int
	Col      int
	GameOver bool
	Winner   WinResult
	Board    Board
	NextTurn Cell
	Error    string
}
