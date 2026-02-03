# Connect 4 - Real-Time Multiplayer

A fast, real-time Connect 4 game built with Go and WebSockets. Play against other players online or challenge the bot.

## What's Inside

- **WebSocket Multiplayer** - Play with anyone, anywhere in real-time
- **AI Opponent** - Minimax bot that thinks 6 moves ahead
- **Quick Matchmaking** - Find a match in 10 seconds or play the bot
- **Reconnection** - Lost connection? You have 30 seconds to rejoin
- **Leaderboard** - Track your wins and compete with others
- **Clean UI** - Dark theme with smooth animations

## Project Structure
Connect4/
├── cmd/server/           # Main entry point
├── internal/
│   ├── game/             # Game logic and AI
│   ├── matchmaking/      # Player queue system
│   ├── websocket/        # WebSocket handler
│   ├── state/            # Game state manager
│   ├── storage/          # PostgreSQL integration
│   ├── leaderboard/      # Stats tracking
│   └── analytics/        # Event logging
├── config/               # Server config
└── frontend/             # HTML/CSS/JS UI

## Getting Started

**Requirements:**
- Go 1.21 or higher
- PostgreSQL (optional)

**Run it:**
```bash
cd Connect4
go mod tidy
go run ./cmd/server
```

Open `http://localhost:8080` and start playing.

**With Docker:**
```bash
docker build -t connect4 .
docker run -p 8080:8080 connect4
```

**With Database:**
```bash
export DATABASE_URL="postgres://user:pass@localhost:5432/connect4?sslmode=disable"
go run ./cmd/server
```

## Configuration

| Variable | Default | What it does |
|----------|---------|--------------|
| `PORT` | 8080 | Server port |
| `DATABASE_URL` | - | PostgreSQL connection |
| `MATCH_TIMEOUT` | 10 | Seconds before bot match |
| `RECONNECT_WINDOW` | 30 | Reconnection time limit |
| `KAFKA_BROKERS` | - | Event streaming (optional) |

## How to Play

1. Enter your username
2. Click "Find Match"
3. Wait for an opponent (or get matched with the bot)
4. Drop your pieces by clicking columns
5. Connect four in a row to win!

## API Reference

**HTTP Endpoints:**
- `GET /` - Game interface
- `GET /leaderboard` - Top 10 players
- `GET /health` - Server status
- `WebSocket /ws` - Game connection

**WebSocket Messages:**

Join matchmaking:
```json
{"type": "join_queue", "payload": {"username": "player1"}}
```

Make a move:
```json
{"type": "move", "payload": {"game_id": "abc123", "column": 3}}
```

Game started:
```json
{
  "type": "game_start",
  "payload": {
    "game_id": "abc123",
    "opponent_name": "player2",
    "is_bot": false,
    "symbol": 1,
    "your_turn": true
  }
}
```

Board update:
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

## How the Bot Works

The AI uses Minimax with alpha-beta pruning:

1. Check if it can win this turn → take it
2. Check if opponent can win → block them
3. Search 6 moves deep to find the best move
4. Prefer center columns (stronger position)
5. Evaluate threats and opportunities

Position scoring:
- Three in a row with space: +100 points
- Two in a row with spaces: +10 points
- Center columns: bonus points

## Thread Safety

All shared data is protected:
- Game state uses `sync.RWMutex`
- Client connections stored in locked maps
- Matchmaking queue uses channels
- Bot calculations work on copied boards

## Deploy to Production

Works on Render, Railway, Fly.io, and similar platforms.

Example `render.yaml`:
```yaml
services:
  - type: web
    name: connect4
    env: docker
    envVars:
      - key: DATABASE_URL
        fromDatabase:
          name: connect4-db
          property: connectionString
```

## Why This Project Exists

Connect 4 is a simple game, but building it right is surprisingly complex. This project tackles real problems: how do you keep two players in sync across the internet? How do you handle someone's WiFi cutting out mid-game? How do you make an AI that's challenging but not frustrating?

The answers involve WebSockets, careful state management, and a chess-like algorithm called Minimax. But more than that, this project is about taking something familiar and making it work flawlessly in the real world. It's production-grade code that handles edge cases, recovers from failures, and scales gracefully.

Whether you're here to play, learn from the code, or build something similar, I hope this shows that even classic games can teach us something new about software architecture, concurrency, and real-time systems.

Thanks for checking it out. Now go connect four. 🎮

## License

MIT

---

**Made by Kriti Porwal**
