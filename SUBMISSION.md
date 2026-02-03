# Quick Start - Connect 4 Submission

## 📋 Pre-Submission Checklist

✅ All code in `Connect4/` folder  
✅ Backend: Production-grade Go with competitive AI  
✅ Frontend: Modern UI with real-time updates  
✅ README.md with setup instructions  
✅ Dockerfile for deployment  
✅ Successfully builds and runs  

---

## 🚀 Quick Deploy to Render (Recommended)

### Step 1: Push to GitHub

```bash
cd Connect4

# Initialize git
git init

# Add all files
git add .

# Commit
git commit -m "Connect 4 real-time multiplayer game"

# Add your GitHub repo as remote
git remote add origin https://github.com/YOUR_USERNAME/connect4-game.git

# Push
git push -u origin main
```

### Step 2: Deploy on Render

1. Go to [render.com](https://render.com) and sign up/login
2. Click **"New +"** → **"Blueprint"**
3. Connect your GitHub repository
4. Render will detect `render.yaml` automatically
5. Click **"Apply"**
6. Wait 3-5 minutes for deployment

**Your app will be live at**: `https://connect4-server.onrender.com`

---

## 📝 Submission Format

**Email/Form submission should include**:

```
GitHub Repository: https://github.com/YOUR_USERNAME/connect4-game
Live URL: https://connect4-server.onrender.com
```

---

## 🎮 Quick Test Instructions

Once deployed, test your live app:

1. **Open two browser tabs** with your app URL
2. **Tab 1**: Enter username "Player1" → Click "Play Now"
3. **Tab 2**: Enter username "Player2" → Click "Play Now"
4. You should be matched together
5. Play a game and verify moves sync in real-time
6. Check the leaderboard updates

**Test bot gameplay**:
1. Open one tab, enter username, click Play
2. Wait 10 seconds
3. You'll be matched with the competitive AI bot

---

## ⚡ Alternative: Railway (Even Faster)

```bash
# Install Railway CLI
npm install -g railway

# Login
railway login

# Deploy from Connect4 folder
cd Connect4
railway init
railway up

# Get URL
railway domain
```

---

## 📂 What Reviewers Will See

```
connect4-game/  (GitHub repo)
├── README.md              ← Setup instructions
├── Dockerfile             ← Docker deployment
├── render.yaml            ← Auto-deployment config
├── cmd/server/main.go     ← Entry point
├── internal/              ← Backend packages
│   ├── game/              ← Game engine + AI bot
│   ├── websocket/         ← Real-time communication
│   ├── matchmaking/       ← Queue system
│   ├── storage/           ← PostgreSQL
│   └── leaderboard/       ← Scoring
└── frontend/              ← Web UI
    ├── index.html
    ├── style.css
    └── app.js
```

---

## 🎯 Key Features to Highlight

When submitting, you can mention:

✅ **Backend-Authoritative Design** - Prevents cheating  
✅ **Competitive AI** - Minimax algorithm (depth 6)  
✅ **Real-Time Multiplayer** - WebSocket with 10s matchmaking  
✅ **Reconnection Support** - 30-second grace period  
✅ **Production-Ready** - Thread-safe, tested, documented  
✅ **Cloud-Deployable** - Docker + render.yaml included  

---

## 🆘 Troubleshooting

**Port already in use?**
```bash
export PORT=9090
go run ./cmd/server
```

**Database connection fails?**
- App works without database (in-memory fallback)
- Render will auto-provision PostgreSQL if using render.yaml

**Frontend not loading?**
- Check that `frontend/` folder is in the same directory as the binary
- Server logs will show: "Serving frontend from ./frontend"

---

## ✨ You're Ready!

Your Connect 4 game is submission-ready with:
- ✅ Clean, organized code
- ✅ Comprehensive README
- ✅ One-click deployment
- ✅ Live demo capability

Good luck with your submission! 🚀
