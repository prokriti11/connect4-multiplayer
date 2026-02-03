// Package leaderboard provides the leaderboard service.
// It wraps storage operations and can work without a database (in-memory fallback).
package leaderboard

import (
	"log"
	"sort"
	"sync"

	"connect4/internal/storage"
)

// PlayerScore represents a player's win count
type PlayerScore struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
}

// Service manages leaderboard operations
type Service struct {
	postgres *storage.Postgres

	// In-memory fallback when no database
	memoryScores map[string]int
	mu           sync.RWMutex
}

// NewService creates a new leaderboard service
func NewService(pg *storage.Postgres) *Service {
	return &Service{
		postgres:     pg,
		memoryScores: make(map[string]int),
	}
}

// RecordWin records a win for a player
func (s *Service) RecordWin(username string) {
	if username == "" || username == "Bot" {
		return
	}

	// Try database first
	if s.postgres != nil {
		if err := s.postgres.RecordWin(username); err != nil {
			log.Printf("DB error, falling back to memory: %v", err)
		} else {
			return
		}
	}

	// In-memory fallback
	s.mu.Lock()
	defer s.mu.Unlock()
	s.memoryScores[username]++
	log.Printf("Recorded win for %s (in-memory)", username)
}

// RecordLoss records a loss for a player
func (s *Service) RecordLoss(username string) {
	if username == "" || username == "Bot" {
		return
	}

	// Try database first
	if s.postgres != nil {
		if err := s.postgres.RecordLoss(username); err != nil {
			log.Printf("DB error recording loss: %v", err)
		} else {
			return
		}
	}

	// In-memory fallback (losses not tracked in simple memory mode)
	log.Printf("Recorded loss for %s (in-memory)", username)
}

// RecordDraw records a draw for a player
func (s *Service) RecordDraw(username string) {
	if username == "" || username == "Bot" {
		return
	}

	// Try database first
	if s.postgres != nil {
		if err := s.postgres.RecordDraw(username); err != nil {
			log.Printf("DB error recording draw: %v", err)
		} else {
			return
		}
	}

	// In-memory fallback (draws not tracked in simple memory mode)
	log.Printf("Recorded draw for %s (in-memory)", username)
}

// GetLeaderboard returns the top players
func (s *Service) GetLeaderboard(limit int) []PlayerScore {
	// Try database first
	if s.postgres != nil {
		stats, err := s.postgres.GetLeaderboard(limit)
		if err == nil && len(stats) > 0 {
			scores := make([]PlayerScore, len(stats))
			for i, stat := range stats {
				scores[i] = PlayerScore{
					Username: stat.Username,
					Wins:     stat.Wins,
				}
			}
			return scores
		}
	}

	// In-memory fallback
	s.mu.RLock()
	defer s.mu.RUnlock()

	scores := make([]PlayerScore, 0, len(s.memoryScores))
	for username, wins := range s.memoryScores {
		scores = append(scores, PlayerScore{
			Username: username,
			Wins:     wins,
		})
	}

	// Sort by wins descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].Wins > scores[j].Wins
	})

	// Apply limit
	if limit > 0 && len(scores) > limit {
		scores = scores[:limit]
	}

	return scores
}

// HasDatabase returns true if connected to a database
func (s *Service) HasDatabase() bool {
	return s.postgres != nil
}
