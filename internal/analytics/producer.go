// Package analytics provides Kafka-based event streaming for game analytics.
// This is designed to be non-blocking - events are sent asynchronously.
package analytics

import (
	"encoding/json"
	"log"
	"time"
)

// Event types for analytics
const (
	EventGameStarted = "GAME_STARTED"
	EventMoveMade    = "MOVE_MADE"
	EventGameEnded   = "GAME_ENDED"
)

// GameEvent represents an analytics event
type GameEvent struct {
	Type      string                 `json:"type"`
	GameID    string                 `json:"game_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Producer sends events to Kafka (or logs them when Kafka is not configured)
type Producer struct {
	topic   string
	enabled bool
	events  chan GameEvent
	// kafka   *kafka.Writer // Uncomment when adding real Kafka
}

// NewProducer creates a new analytics producer
// In production, pass Kafka broker addresses
func NewProducer(brokers []string, topic string) *Producer {
	p := &Producer{
		topic:   topic,
		enabled: len(brokers) > 0,
		events:  make(chan GameEvent, 1000),
	}

	// Start background event processor
	go p.processEvents()

	return p
}

// processEvents handles events in the background
// This ensures gameplay is never blocked by analytics
func (p *Producer) processEvents() {
	for event := range p.events {
		if p.enabled {
			// TODO: Send to Kafka
			// err := p.kafka.WriteMessages(context.Background(), kafka.Message{
			//     Key:   []byte(event.GameID),
			//     Value: eventBytes,
			// })
			p.logEvent(event)
		} else {
			// Log to stdout as fallback
			p.logEvent(event)
		}
	}
}

// logEvent outputs event to log
func (p *Producer) logEvent(event GameEvent) {
	bytes, _ := json.Marshal(event)
	log.Printf("[ANALYTICS] %s", string(bytes))
}

// GameStarted emits a game start event
func (p *Producer) GameStarted(gameID, player1, player2 string, isBot bool) {
	event := GameEvent{
		Type:      EventGameStarted,
		GameID:    gameID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"player1": player1,
			"player2": player2,
			"is_bot":  isBot,
		},
	}

	select {
	case p.events <- event:
	default:
		// Channel full, drop event (non-blocking)
	}
}

// MoveMade emits a move event
func (p *Producer) MoveMade(gameID string, player string, column, row int, moveNumber int) {
	event := GameEvent{
		Type:      EventMoveMade,
		GameID:    gameID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"player":      player,
			"column":      column,
			"row":         row,
			"move_number": moveNumber,
		},
	}

	select {
	case p.events <- event:
	default:
	}
}

// GameEnded emits a game end event
func (p *Producer) GameEnded(gameID, winner string, reason string, totalMoves int, duration time.Duration) {
	event := GameEvent{
		Type:      EventGameEnded,
		GameID:    gameID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"winner":      winner,
			"reason":      reason,
			"total_moves": totalMoves,
			"duration_ms": duration.Milliseconds(),
		},
	}

	select {
	case p.events <- event:
	default:
	}
}

// Close shuts down the producer
func (p *Producer) Close() {
	close(p.events)
}
