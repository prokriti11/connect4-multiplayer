// Connect 4 Real-Time Multiplayer Server
// Entry point that initializes all components and starts the HTTP/WebSocket server.
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"connect4/config"
	"connect4/internal/leaderboard"
	"connect4/internal/matchmaking"
	"connect4/internal/state"
	"connect4/internal/storage"
	"connect4/internal/websocket"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting Connect 4 Server on port %s", cfg.Port)

	// Initialize PostgreSQL (optional - works without it)
	postgres, err := storage.New()
	if err != nil {
		log.Printf("Database initialization failed: %v (continuing without persistence)", err)
	} else if postgres != nil {
		defer postgres.Close()
	}

	// Initialize core components
	matchmaker := matchmaking.NewQueue()
	gameStore := state.NewStore()
	leaderboardService := leaderboard.NewService(postgres)

	// Start matchmaking goroutine
	go matchmaker.Run()

	// Initialize WebSocket hub
	hub := websocket.NewHub(matchmaker, gameStore, leaderboardService)
	go hub.Run()

	// Set up HTTP routes
	setupRoutes(hub, leaderboardService)

	// Start server
	addr := ":" + cfg.Port
	log.Printf("Server listening on %s", addr)
	log.Printf("WebSocket endpoint: ws://localhost%s/ws", addr)
	log.Printf("Leaderboard API: http://localhost%s/leaderboard", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// setupRoutes configures all HTTP endpoints
func setupRoutes(hub *websocket.Hub, lb *leaderboard.Service) {
	// WebSocket endpoint
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.ServeWs(hub, w, r)
	})

	// Leaderboard API
	http.HandleFunc("/leaderboard", func(w http.ResponseWriter, r *http.Request) {
		handleLeaderboard(w, r, lb)
	})

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Serve frontend static files
	setupStaticFiles()
}

// handleLeaderboard returns the current leaderboard
func handleLeaderboard(w http.ResponseWriter, r *http.Request, lb *leaderboard.Service) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	scores := lb.GetLeaderboard(10)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scores)
}

// setupStaticFiles serves the frontend from the frontend directory
func setupStaticFiles() {
	// Try to find frontend directory
	frontendDirs := []string{
		"./frontend",
		"../frontend",
		"./Connect4/frontend",
	}

	var frontendDir string
	for _, dir := range frontendDirs {
		if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
			frontendDir = dir
			break
		}
	}

	if frontendDir != "" {
		log.Printf("Serving frontend from %s", frontendDir)
		fs := http.FileServer(http.Dir(frontendDir))
		http.Handle("/", fs)
	} else {
		// Fallback: simple message
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/" {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
<!DOCTYPE html>
<html>
<head><title>Connect 4</title></head>
<body>
	<h1>Connect 4 Server Running</h1>
	<p>WebSocket endpoint: /ws</p>
	<p>Leaderboard: <a href="/leaderboard">/leaderboard</a></p>
</body>
</html>
			`))
		})
	}
}
