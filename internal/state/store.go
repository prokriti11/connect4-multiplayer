// Package state manages active game sessions in memory.
// It provides thread-safe access to ongoing games and handles lifecycle events.
package state

import (
	"log"
	"sync"
	"time"

	"connect4/internal/game"
)

const (
	ReconnectWindow = 30 * time.Second // Time allowed for reconnection
)

// DisconnectInfo tracks when a player disconnected
type DisconnectInfo struct {
	GameID       string
	PlayerID     string
	DisconnectAt time.Time
}

// Store manages all active games in memory
type Store struct {
	games       map[string]*game.Game
	userToGame  map[string]string          // Username -> GameID for reconnection
	disconnects map[string]*DisconnectInfo // PlayerID -> disconnect info
	mu          sync.RWMutex
}

// NewStore creates a new game state store
func NewStore() *Store {
	return &Store{
		games:       make(map[string]*game.Game),
		userToGame:  make(map[string]string),
		disconnects: make(map[string]*DisconnectInfo),
	}
}

// CreateGame adds a new game to the store
func (s *Store) CreateGame(g *game.Game) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.games[g.ID] = g
	s.userToGame[g.Player1.Username] = g.ID
	s.userToGame[g.Player2.Username] = g.ID

	log.Printf("Game %s created: %s vs %s", g.ID, g.Player1.Username, g.Player2.Username)
}

// GetGame retrieves a game by ID
func (s *Store) GetGame(gameID string) (*game.Game, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	g, exists := s.games[gameID]
	return g, exists
}

// GetGameByUsername finds a game by player username (for reconnection)
func (s *Store) GetGameByUsername(username string) (*game.Game, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	gameID, exists := s.userToGame[username]
	if !exists {
		return nil, false
	}

	g, exists := s.games[gameID]
	if !exists || g.Status != game.StatusActive {
		return nil, false
	}

	return g, true
}

// RemoveGame deletes a game from active storage
func (s *Store) RemoveGame(gameID string) *game.Game {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, exists := s.games[gameID]
	if !exists {
		return nil
	}

	delete(s.games, gameID)
	delete(s.userToGame, g.Player1.Username)
	delete(s.userToGame, g.Player2.Username)
	delete(s.disconnects, g.Player1.ID)
	delete(s.disconnects, g.Player2.ID)

	log.Printf("Game %s removed from store", gameID)
	return g
}

// RecordDisconnect marks a player as disconnected
func (s *Store) RecordDisconnect(playerID, gameID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.disconnects[playerID] = &DisconnectInfo{
		GameID:       gameID,
		PlayerID:     playerID,
		DisconnectAt: time.Now(),
	}

	log.Printf("Player %s disconnected from game %s", playerID, gameID)
}

// ClearDisconnect removes disconnect tracking (player reconnected)
func (s *Store) ClearDisconnect(playerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.disconnects, playerID)
	log.Printf("Player %s reconnected", playerID)
}

// IsDisconnected checks if a player is currently disconnected
func (s *Store) IsDisconnected(playerID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.disconnects[playerID]
	return exists
}

// GetDisconnectInfo returns disconnect details for a player
func (s *Store) GetDisconnectInfo(playerID string) (*DisconnectInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.disconnects[playerID]
	return info, exists
}

// CheckTimeouts checks for expired disconnections and forfeits games
// Returns list of games that were forfeited
func (s *Store) CheckTimeouts() []*game.Game {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	var forfeited []*game.Game

	for playerID, info := range s.disconnects {
		if now.Sub(info.DisconnectAt) >= ReconnectWindow {
			// Player timed out - forfeit the game
			g, exists := s.games[info.GameID]
			if exists && g.Status == game.StatusActive {
				g.Forfeit(playerID)
				forfeited = append(forfeited, g)
				log.Printf("Game %s forfeited due to %s timeout", info.GameID, playerID)
			}
			delete(s.disconnects, playerID)
		}
	}

	return forfeited
}

// ActiveGameCount returns number of active games
func (s *Store) ActiveGameCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.games)
}

// GetAllActiveGames returns a snapshot of all active games
func (s *Store) GetAllActiveGames() []*game.Game {
	s.mu.RLock()
	defer s.mu.RUnlock()

	games := make([]*game.Game, 0, len(s.games))
	for _, g := range s.games {
		games = append(games, g)
	}
	return games
}

// UpdatePlayerID updates the player's connection ID after reconnection
func (s *Store) UpdatePlayerID(gameID, username, newPlayerID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	g, exists := s.games[gameID]
	if !exists {
		return false
	}

	// Use the Game's thread-safe method to update player ID
	oldID := g.UpdatePlayerID(username, newPlayerID)
	if oldID != "" {
		delete(s.disconnects, oldID)
		log.Printf("Updated player ID for %s: %s -> %s", username, oldID, newPlayerID)
		return true
	}

	return false
}

// StartTimeoutChecker runs a goroutine that periodically checks for timeouts
func (s *Store) StartTimeoutChecker(onForfeit func(*game.Game)) {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			forfeited := s.CheckTimeouts()
			for _, g := range forfeited {
				if onForfeit != nil {
					onForfeit(g)
				}
			}
		}
	}()
}
