# Connect 4 - Real-Time Multiplayer Game

A production-grade Connect 4 game server built with Go, featuring WebSocket multiplayer, competitive AI bot, and persistent leaderboard.

## Features

- **Real-Time Multiplayer**: WebSocket-based gameplay with instant state synchronization
- **Competitive AI Bot**: Minimax algorithm with alpha-beta pruning (depth 6)
- **Smart Matchmaking**: 10-second queue with automatic bot fallback
- **Reconnection Support**: 30-second grace period for disconnected players
- **Persistent Leaderboard**: PostgreSQL storage with in-memory fallback
- **Modern Frontend**: Dark theme UI with animations

## Architecture

```
Connect4/
├── cmd/server/           # Entry point
├── internal/
│   ├── game/             # Core game engine
│   │   ├── types.go      # Type definitions
│   │   ├── board.go      # Grid & gravity logic
│   │   ├── rules.go      # Win detection
│   │   ├── engine.go     # Turn management
│   │   └── bot.go        # Minimax AI
│   ├── matchmaking/      # Player queue
│   ├── websocket/        # Real-time communication
│   ├── state/            # Active game storage
│   ├── storage/          # PostgreSQL persistence
│   ├── leaderboard/      # Leaderboard service
│   └── analytics/        # Kafka event producer (optional)
├── config/               # Configuration
└── frontend/             # Web UI
```

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL (optional, for persistence)

### Run Locally

```bash
cd Connect4

# Download dependencies
go mod tidy

# Run the server
go run ./cmd/server

# Server starts at http://localhost:8080
```

### With Docker

```bash
# Build image
docker build -t connect4 .

# Run container
docker run -p 8080:8080 connect4
```

### With PostgreSQL

```bash
# Set database URL
export DATABASE_URL="postgres://user:pass@localhost:5432/connect4?sslmode=disable"

# Run server
go run ./cmd/server
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DATABASE_URL` | - | PostgreSQL connection string |
| `MATCH_TIMEOUT` | `10` | Seconds before bot match |
| `RECONNECT_WINDOW` | `30` | Seconds for reconnection |
| `KAFKA_BROKERS` | - | Kafka broker addresses (optional) |

## API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/ws` | WebSocket | Game communication |
| `/leaderboard` | GET | Top 10 players |
| `/health` | GET | Health check |
| `/` | GET | Frontend UI |

## WebSocket Protocol

### Client → Server

```json
// Join matchmaking
{"type": "join_queue", "payload": {"username": "player1"}}

// Make a move
{"type": "move", "payload": {"game_id": "...", "column": 3}}
```

### Server → Client

```json
// Game started
{"type": "game_start", "payload": {"game_id": "...", "opponent_name": "...", "is_bot": false, "symbol": 1, "your_turn": true}}

// Game state update
{"type": "game_state", "payload": {"board": [[...]], "turn": 1, "winner": 0, "status": "active"}}

// Error
{"type": "error", "payload": {"message": "Not your turn"}}
```

## Bot Algorithm

The AI uses Minimax with alpha-beta pruning:

1. **Immediate Win**: Win if possible
2. **Block**: Block opponent's winning move
3. **Minimax Search**: Evaluate positions to depth 6
4. **Heuristics**:
   - Center column preference
   - 3-in-a-row with open space: +100
   - 2-in-a-row with open spaces: +10

## Concurrency Safety

- **Game State**: Protected by `sync.RWMutex`
- **Client Registry**: Mutex-protected map
- **Matchmaking Queue**: Channel-based with mutex
- **Board Simulation**: Cloned boards for bot calculation

## Production Deployment

Recommended platforms: Render, Railway, Fly.io

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

## License

MIT
