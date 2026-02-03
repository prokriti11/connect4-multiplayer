// Package storage provides PostgreSQL persistence for completed games and leaderboard.
package storage

import (
	"database/sql"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

// GameRecord represents a completed game stored in the database
type GameRecord struct {
	ID          string
	Player1     string
	Player2     string
	Winner      string // "player1", "player2", "draw"
	Moves       int
	Duration    time.Duration
	CompletedAt time.Time
}

// Postgres handles database operations
type Postgres struct {
	db *sql.DB
}

// New creates a new Postgres storage connection
func New() (*Postgres, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Println("DATABASE_URL not set, running without persistence")
		return nil, nil
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		log.Printf("Database connection failed: %v", err)
		return nil, err
	}

	p := &Postgres{db: db}

	// Initialize schema
	if err := p.initSchema(); err != nil {
		return nil, err
	}

	log.Println("Connected to PostgreSQL database")
	return p, nil
}

// initSchema creates the required database tables
func (p *Postgres) initSchema() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS players (
			username VARCHAR(50) PRIMARY KEY,
			wins INT DEFAULT 0,
			losses INT DEFAULT 0,
			draws INT DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW(),
			updated_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS games (
			id VARCHAR(100) PRIMARY KEY,
			player1 VARCHAR(50) NOT NULL,
			player2 VARCHAR(50) NOT NULL,
			winner VARCHAR(10),
			moves INT DEFAULT 0,
			duration_ms BIGINT DEFAULT 0,
			completed_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_games_player1 ON games(player1)`,
		`CREATE INDEX IF NOT EXISTS idx_games_player2 ON games(player2)`,
		`CREATE INDEX IF NOT EXISTS idx_players_wins ON players(wins DESC)`,
	}

	for _, query := range queries {
		if _, err := p.db.Exec(query); err != nil {
			log.Printf("Schema error: %v", err)
			return err
		}
	}

	return nil
}

// RecordWin increments the win count for a player
func (p *Postgres) RecordWin(username string) error {
	if p == nil || p.db == nil {
		return nil
	}

	query := `
		INSERT INTO players (username, wins, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (username)
		DO UPDATE SET wins = players.wins + 1, updated_at = NOW()
	`

	_, err := p.db.Exec(query, username)
	if err != nil {
		log.Printf("Error recording win: %v", err)
	}
	return err
}

// RecordLoss increments the loss count for a player
func (p *Postgres) RecordLoss(username string) error {
	if p == nil || p.db == nil {
		return nil
	}

	query := `
		INSERT INTO players (username, losses, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (username)
		DO UPDATE SET losses = players.losses + 1, updated_at = NOW()
	`

	_, err := p.db.Exec(query, username)
	return err
}

// RecordDraw increments the draw count for a player
func (p *Postgres) RecordDraw(username string) error {
	if p == nil || p.db == nil {
		return nil
	}

	query := `
		INSERT INTO players (username, draws, updated_at)
		VALUES ($1, 1, NOW())
		ON CONFLICT (username)
		DO UPDATE SET draws = players.draws + 1, updated_at = NOW()
	`

	_, err := p.db.Exec(query, username)
	return err
}

// PlayerStats represents aggregated player statistics
type PlayerStats struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Draws    int    `json:"draws"`
}

// GetLeaderboard returns top players sorted by wins
func (p *Postgres) GetLeaderboard(limit int) ([]PlayerStats, error) {
	if p == nil || p.db == nil {
		return []PlayerStats{}, nil
	}

	query := `
		SELECT username, wins, losses, draws
		FROM players
		ORDER BY wins DESC, losses ASC
		LIMIT $1
	`

	rows, err := p.db.Query(query, limit)
	if err != nil {
		log.Printf("Error fetching leaderboard: %v", err)
		return nil, err
	}
	defer rows.Close()

	var stats []PlayerStats
	for rows.Next() {
		var s PlayerStats
		if err := rows.Scan(&s.Username, &s.Wins, &s.Losses, &s.Draws); err != nil {
			continue
		}
		stats = append(stats, s)
	}

	return stats, nil
}

// SaveGame persists a completed game record
func (p *Postgres) SaveGame(record GameRecord) error {
	if p == nil || p.db == nil {
		return nil
	}

	query := `
		INSERT INTO games (id, player1, player2, winner, moves, duration_ms, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO NOTHING
	`

	_, err := p.db.Exec(query,
		record.ID,
		record.Player1,
		record.Player2,
		record.Winner,
		record.Moves,
		record.Duration.Milliseconds(),
		record.CompletedAt,
	)

	return err
}

// GetPlayerStats retrieves stats for a specific player
func (p *Postgres) GetPlayerStats(username string) (*PlayerStats, error) {
	if p == nil || p.db == nil {
		return nil, nil
	}

	query := `SELECT username, wins, losses, draws FROM players WHERE username = $1`

	var s PlayerStats
	err := p.db.QueryRow(query, username).Scan(&s.Username, &s.Wins, &s.Losses, &s.Draws)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// Close closes the database connection
func (p *Postgres) Close() error {
	if p != nil && p.db != nil {
		return p.db.Close()
	}
	return nil
}
