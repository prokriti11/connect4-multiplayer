// Package matchmaking implements the player queue and matchmaking logic.
// It handles pairing players together or matching with a bot after timeout.
package matchmaking

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	MatchTimeout = 10 * time.Second // Wait time before bot fallback
)

// MatchResult contains the result of a matchmaking request
type MatchResult struct {
	GameID       string
	OpponentID   string
	OpponentName string
	IsBot        bool
	Symbol       int // 1 = Red (first), 2 = Yellow (second)
}

// PlayerRequest represents a player waiting in queue
type PlayerRequest struct {
	ID           string
	Username     string
	ResponseChan chan MatchResult
	EnteredAt    time.Time
	ctx          context.Context
	cancel       context.CancelFunc
}

// Queue manages the matchmaking queue
type Queue struct {
	waiting  []*PlayerRequest
	mu       sync.Mutex
	register chan *PlayerRequest
	quit     chan struct{}
}

// NewQueue creates a new matchmaking queue
func NewQueue() *Queue {
	return &Queue{
		waiting:  make([]*PlayerRequest, 0),
		register: make(chan *PlayerRequest, 100),
		quit:     make(chan struct{}),
	}
}

// Add puts a player in the matchmaking queue.
// Returns a channel that will receive the match result.
func (q *Queue) Add(playerID, username string) chan MatchResult {
	respChan := make(chan MatchResult, 1)
	ctx, cancel := context.WithTimeout(context.Background(), MatchTimeout)

	req := &PlayerRequest{
		ID:           playerID,
		Username:     username,
		ResponseChan: respChan,
		EnteredAt:    time.Now(),
		ctx:          ctx,
		cancel:       cancel,
	}

	q.register <- req
	return respChan
}

// Run starts the matchmaking loop. Should be called in a goroutine.
func (q *Queue) Run() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case req := <-q.register:
			q.handleNewPlayer(req)

		case <-ticker.C:
			q.checkTimeouts()

		case <-q.quit:
			return
		}
	}
}

// Stop gracefully shuts down the queue
func (q *Queue) Stop() {
	close(q.quit)
}

// handleNewPlayer adds a player to queue and attempts matching
func (q *Queue) handleNewPlayer(newPlayer *PlayerRequest) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// First, remove any old entries for this username (prevents self-matching on refresh)
	var cleanQueue []*PlayerRequest
	for _, player := range q.waiting {
		if player.Username != newPlayer.Username {
			cleanQueue = append(cleanQueue, player)
		} else {
			player.cancel() // Cancel old entry
			log.Printf("Removed old queue entry for %s", player.Username)
		}
	}
	q.waiting = cleanQueue

	// Try to find an opponent in the queue
	for i, waiting := range q.waiting {
		// Don't match with yourself
		if waiting.ID == newPlayer.ID {
			continue
		}

		// Check if waiting player is still valid (not timed out)
		select {
		case <-waiting.ctx.Done():
			continue // Skip expired entries
		default:
		}

		// Found a match!
		q.waiting = append(q.waiting[:i], q.waiting[i+1:]...) // Remove from queue

		gameID := generateGameID()
		log.Printf("Match found: %s (%s) vs %s (%s) -> Game %s",
			waiting.ID, waiting.Username, newPlayer.ID, newPlayer.Username, gameID)

		// Notify both players
		waiting.ResponseChan <- MatchResult{
			GameID:       gameID,
			OpponentID:   newPlayer.ID,
			OpponentName: newPlayer.Username,
			IsBot:        false,
			Symbol:       1, // First player is Red
		}
		waiting.cancel()

		newPlayer.ResponseChan <- MatchResult{
			GameID:       gameID,
			OpponentID:   waiting.ID,
			OpponentName: waiting.Username,
			IsBot:        false,
			Symbol:       2, // Second player is Yellow
		}
		newPlayer.cancel()

		return
	}

	// No opponent found, add to queue
	q.waiting = append(q.waiting, newPlayer)
	log.Printf("Player %s (%s) added to queue. Queue size: %d",
		newPlayer.ID, newPlayer.Username, len(q.waiting))
}

// checkTimeouts handles players who have waited too long
func (q *Queue) checkTimeouts() {
	q.mu.Lock()
	defer q.mu.Unlock()

	now := time.Now()
	var remaining []*PlayerRequest

	for _, player := range q.waiting {
		// Check if context timed out
		select {
		case <-player.ctx.Done():
			// Time's up - match with bot
			gameID := generateGameID()
			log.Printf("Timeout: Matching %s (%s) with bot -> Game %s",
				player.ID, player.Username, gameID)

			player.ResponseChan <- MatchResult{
				GameID:       gameID,
				OpponentID:   "bot",
				OpponentName: "Bot",
				IsBot:        true,
				Symbol:       1, // Human plays first against bot
			}
			player.cancel()

		default:
			// Check if approaching timeout (for logging)
			elapsed := now.Sub(player.EnteredAt)
			if elapsed > MatchTimeout-time.Second {
				log.Printf("Player %s waiting for %.1fs...", player.Username, elapsed.Seconds())
			}
			remaining = append(remaining, player)
		}
	}

	q.waiting = remaining
}

// RemovePlayer removes a player from the queue (e.g., if they disconnect)
func (q *Queue) RemovePlayer(playerID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for i, player := range q.waiting {
		if player.ID == playerID {
			player.cancel()
			q.waiting = append(q.waiting[:i], q.waiting[i+1:]...)
			log.Printf("Player %s removed from queue", playerID)
			return
		}
	}
}

// QueueSize returns number of players waiting
func (q *Queue) QueueSize() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.waiting)
}

// generateGameID creates a unique game identifier
func generateGameID() string {
	return fmt.Sprintf("game-%d", time.Now().UnixNano())
}
