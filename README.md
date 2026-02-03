# Connect 4 - Real-Time Multiplayer

A fast, real-time Connect 4 game built with Go and WebSockets. Play against other players online or challenge the bot.

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![WebSocket](https://img.shields.io/badge/WebSocket-Real--Time-4A90E2?style=flat)](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-Optional-336791?style=flat&logo=postgresql)](https://www.postgresql.org)

---

## Features

- **Real-Time Multiplayer** - Play with anyone, anywhere via WebSockets
- **AI Opponent** - Minimax bot that thinks 6 moves ahead
- **Quick Matchmaking** - Find a match in 10 seconds or play the bot
- **Reconnection Support** - 30-second grace period if you disconnect
- **Leaderboard** - Track your wins and climb the rankings
- **Clean UI** - Dark theme with smooth animations

---

## Quick Start

### Run Locally
```bash
git clone https://github.com/prokriti11/connect4-multiplayer.git
cd connect4-multiplayer
go mod download
go run ./cmd/server
```

Open `http://localhost:8080` and start playing!

### With Docker
```bash
docker build -t connect4 .
docker run -p 8080:8080 connect4
```

### With PostgreSQL
```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/connect4?sslmode=disable"
go run ./cmd/server
```

---

## Project Structure
Connect4/
├── cmd/server/           # Main entry point
├── internal/
│   ├── game/             # Game logic and AI
│   ├── websocket/        # WebSocket handler
│   ├── matchmaking/      # Player queue system
│   ├── state/            # Game state manager
│   ├── storage/          # PostgreSQL integration
│   ├── leaderboard/      # Stats tracking
│   └── analytics/        # Event logging
├── frontend/             # HTML/CSS/JS UI
├── config/               # Server config
└── Dockerfile            # Container build

---

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | Server port |
| `DATABASE_URL` | - | PostgreSQL connection |
| `MATCH_TIMEOUT` | 10 | Seconds before bot match |
| `RECONNECT_WINDOW` | 30 | Reconnection time limit |
| `KAFKA_BROKERS` | - | Event streaming (optional) |

---

## How to Play

1. Enter your username
2. Click "Play Now" to join matchmaking
3. Get matched with a player or the bot in 10 seconds
4. Click columns to drop your pieces
5. Connect four in a row to win!

---

## API Reference

### HTTP Endpoints

- `GET /` - Game interface
- `GET /leaderboard` - Top 10 players (JSON)
- `GET /health` - Server status
- `WebSocket /ws` - Game connection

### WebSocket Messages

**Join matchmaking:**
```json
{"type": "join_queue", "payload": {"username": "Player1"}}
```

**Make a move:**
```json
{"type": "move", "payload": {"game_id": "abc123", "column": 3}}
```

**Game started:**
```json
{
  "type": "game_start",
  "payload": {
    "game_id": "abc123",
    "opponent_name": "Player2",
    "is_bot": false,
    "symbol": 1,
    "your_turn": true
  }
}
```

**Board update:**
```json
{
  "type": "game_state",
  "payload": {
    "board": [[0,0,0,...], ...],
    "turn": 1,
    "winner": 0,
    "status": "active"
  }
}
```

---

## How the Bot Works

The AI uses Minimax with alpha-beta pruning:

1. Check if it can win this turn → take it
2. Check if opponent can win → block them
3. Search 6 moves deep to find the best move
4. Prefer center columns (stronger position)
5. Evaluate threats and opportunities

**Position scoring:**
- Three in a row with space: +100 points
- Two in a row with spaces: +10 points
- Center columns: bonus points

**Performance:**
- Average decision time: 50-100ms
- Search depth: 6 moves

---

## Deploy to Production

### Render

This project includes `render.yaml` for one-click deployment:

1. Fork this repository
2. Go to [Render](https://render.com)
3. Click "New" → "Blueprint"
4. Connect your GitHub repo
5. Click "Apply"

### Railway
```bash
npm install -g railway
railway login
railway init
railway up
```

### Fly.io
```bash
fly launch
fly deploy
```

---

## Tech Stack

**Backend:**
- Go 1.21+ (concurrency, performance)
- Gorilla WebSocket (real-time communication)
- PostgreSQL (persistent storage)
- Docker (containerization)

**Frontend:**
- Vanilla JavaScript (ES6+)
- WebSocket API
- CSS3 (animations, dark theme)
- Responsive design

**Algorithms:**
- Minimax with alpha-beta pruning (depth 6)
- Position evaluation and pattern recognition

---

## Performance

- **Bot Response:** 50-100ms per move
- **WebSocket Latency:** <10ms on local networks
- **Matchmaking:** 10-second timeout
- **Reconnection Window:** 30 seconds
- **Memory:** ~15MB per 1,000 active games

---

## Future Plans

- [ ] Tournament mode with brackets
- [ ] ELO rating system
- [ ] Game replay functionality
- [ ] Mobile app
- [ ] Spectator mode
- [ ] In-game chat
- [ ] Multiple board sizes
- [ ] Custom game rules

---

## Why This Project?

Connect 4 is a simple game, but building it right is surprisingly complex. This project tackles real problems: how do you keep two players in sync across the internet? How do you handle someone's WiFi cutting out mid-game? How do you make an AI that's challenging but not frustrating?

The answers involve WebSockets, careful state management, and a chess-like algorithm called Minimax. But more than that, this project is about taking something familiar and making it work flawlessly in the real world. It's production-grade code that handles edge cases, recovers from failures, and scales gracefully.

Whether you're here to play, learn from the code, or build something similar, I hope this shows that even classic games can teach us something new about software architecture, concurrency, and real-time systems.

Thanks for checking it out. Now go connect four. 🎮

---

## License

MIT

---

**Made by Kriti Porwal**

[![GitHub](https://img.shields.io/badge/GitHub-@prokriti11-181717?style=flat&logo=github)](https://github.com/prokriti11)
