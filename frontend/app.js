/**
 * Connect 4 Frontend Application
 * Handles WebSocket communication, game rendering, and user interactions
 */

// Configuration
const CONFIG = {
    WS_URL: `ws://${window.location.host}/ws`,
    LEADERBOARD_URL: '/leaderboard',
    RECONNECT_DELAY: 2000,
    MAX_RECONNECT_ATTEMPTS: 5
};

// Game state
const state = {
    socket: null,
    username: '',
    gameId: null,
    mySymbol: 0, // 1 = Red, 2 = Yellow
    isMyTurn: false,
    board: Array(6).fill(null).map(() => Array(7).fill(0)),
    gameStatus: 'idle', // idle, matchmaking, playing, ended
    reconnectAttempts: 0,
    matchmakingTimer: null,
    disconnectTimer: null
};

// DOM Elements
const elements = {
    // Screens
    loginScreen: document.getElementById('login-screen'),
    matchmakingScreen: document.getElementById('matchmaking-screen'),
    gameScreen: document.getElementById('game-screen'),
    leaderboardScreen: document.getElementById('leaderboard-screen'),

    // Login
    usernameInput: document.getElementById('username-input'),
    playBtn: document.getElementById('play-btn'),

    // Matchmaking
    matchmakingStatus: document.getElementById('matchmaking-status'),
    countdownTimer: document.getElementById('countdown-timer'),
    cancelBtn: document.getElementById('cancel-btn'),

    // Game
    gameBoard: document.getElementById('game-board'),
    columnIndicators: document.getElementById('column-indicators'),
    player1Name: document.getElementById('player1-name'),
    player2Name: document.getElementById('player2-name'),
    gameStatus: document.getElementById('game-status'),
    gameOverBanner: document.getElementById('game-over-banner'),
    winnerText: document.getElementById('winner-text'),
    playAgainBtn: document.getElementById('play-again-btn'),
    disconnectBanner: document.getElementById('disconnect-banner'),
    reconnectTimer: document.getElementById('reconnect-timer'),

    // Leaderboard
    leaderboardLogin: document.getElementById('leaderboard-login'),
    leaderboardFull: document.getElementById('leaderboard-full')
};

// ============= Screen Management =============

function showScreen(screenId) {
    document.querySelectorAll('.screen').forEach(screen => {
        screen.classList.remove('active');
    });
    document.getElementById(screenId).classList.add('active');
}

// ============= WebSocket Connection =============

function connectWebSocket() {
    if (state.socket && state.socket.readyState === WebSocket.OPEN) {
        return;
    }

    console.log('Connecting to WebSocket...');
    state.socket = new WebSocket(CONFIG.WS_URL);

    state.socket.onopen = () => {
        console.log('WebSocket connected');
        state.reconnectAttempts = 0;

        // If we have a stored game, try to reconnect
        const storedGameId = localStorage.getItem('connect4_game_id');
        const storedUsername = localStorage.getItem('connect4_username');
        if (storedGameId && storedUsername) {
            state.username = storedUsername;
            joinQueue();
        }
    };

    state.socket.onmessage = (event) => {
        // Handle multiple messages separated by newlines
        const messages = event.data.split('\n');
        messages.forEach(msg => {
            if (msg.trim()) {
                handleMessage(JSON.parse(msg));
            }
        });
    };

    state.socket.onclose = () => {
        console.log('WebSocket disconnected');
        if (state.gameStatus === 'playing') {
            // Try to reconnect
            attemptReconnect();
        }
    };

    state.socket.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
}

function attemptReconnect() {
    if (state.reconnectAttempts >= CONFIG.MAX_RECONNECT_ATTEMPTS) {
        alert('Connection lost. Please refresh the page.');
        return;
    }

    state.reconnectAttempts++;
    console.log(`Reconnecting... (attempt ${state.reconnectAttempts})`);

    setTimeout(() => {
        connectWebSocket();
    }, CONFIG.RECONNECT_DELAY);
}

// ============= Message Handling =============

function handleMessage(message) {
    console.log('Received:', message.type, message.payload);

    switch (message.type) {
        case 'game_start':
            handleGameStart(message.payload);
            break;
        case 'game_state':
            handleGameState(message.payload);
            break;
        case 'game_over':
            handleGameOver(message.payload);
            break;
        case 'error':
            handleError(message.payload);
            break;
        case 'opponent_disconnect':
            handleOpponentDisconnect(message.payload);
            break;
        case 'opponent_reconnect':
            handleOpponentReconnect(message.payload);
            break;
    }
}

function handleGameStart(payload) {
    state.gameId = payload.game_id;
    state.mySymbol = payload.symbol;
    state.isMyTurn = payload.your_turn;
    state.gameStatus = 'playing';

    // Store for reconnection
    localStorage.setItem('connect4_game_id', state.gameId);
    localStorage.setItem('connect4_username', state.username);

    // Clear matchmaking timer
    if (state.matchmakingTimer) {
        clearInterval(state.matchmakingTimer);
        state.matchmakingTimer = null;
    }

    // Update UI
    const opponentName = payload.opponent_name + (payload.is_bot ? ' 🤖' : '');

    if (state.mySymbol === 1) {
        elements.player1Name.textContent = state.username + ' (You)';
        elements.player2Name.textContent = opponentName;
    } else {
        elements.player1Name.textContent = opponentName;
        elements.player2Name.textContent = state.username + ' (You)';
    }

    updateTurnIndicator();
    showScreen('game-screen');
}

function handleGameState(payload) {
    state.board = payload.board;
    state.isMyTurn = payload.turn === state.mySymbol;

    renderBoard(payload.last_move);
    updateTurnIndicator();

    // Check for game over
    if (payload.status === 'finished' || payload.status === 'forfeit') {
        handleGameOver(payload);
    }
}

function handleGameOver(payload) {
    state.gameStatus = 'ended';
    localStorage.removeItem('connect4_game_id');

    // Hide disconnect banner if visible
    elements.disconnectBanner.classList.add('hidden');
    if (state.disconnectTimer) {
        clearInterval(state.disconnectTimer);
    }

    // Determine winner message
    let message = '';
    let winnerClass = '';

    if (payload.winner === 3) {
        message = "It's a Draw!";
    } else if (payload.winner === state.mySymbol) {
        message = '🎉 You Win!';
        winnerClass = `winner-${state.mySymbol}`;
    } else {
        message = 'You Lose 😔';
        winnerClass = `winner-${payload.winner}`;
    }

    elements.winnerText.textContent = message;
    elements.gameOverBanner.className = winnerClass;
    elements.gameOverBanner.classList.remove('hidden');
}

function handleError(payload) {
    console.error('Server error:', payload.message);
    // Could show a toast notification here
}

function handleOpponentDisconnect(payload) {
    elements.disconnectBanner.classList.remove('hidden');

    let remaining = payload.time_remaining || 30;
    elements.reconnectTimer.textContent = `${remaining}s remaining`;

    state.disconnectTimer = setInterval(() => {
        remaining--;
        elements.reconnectTimer.textContent = `${remaining}s remaining`;
        if (remaining <= 0) {
            clearInterval(state.disconnectTimer);
        }
    }, 1000);
}

function handleOpponentReconnect(payload) {
    elements.disconnectBanner.classList.add('hidden');
    if (state.disconnectTimer) {
        clearInterval(state.disconnectTimer);
    }
}

// ============= Send Messages =============

function sendMessage(type, payload) {
    if (state.socket && state.socket.readyState === WebSocket.OPEN) {
        state.socket.send(JSON.stringify({ type, payload }));
    }
}

function joinQueue() {
    sendMessage('join_queue', { username: state.username });
}

function makeMove(column) {
    if (!state.isMyTurn || state.gameStatus !== 'playing') {
        return;
    }

    sendMessage('move', {
        game_id: state.gameId,
        column: column
    });
}

// ============= Board Rendering =============

function initializeBoard() {
    elements.gameBoard.innerHTML = '';
    elements.columnIndicators.innerHTML = '';

    // Create cells (6 rows x 7 columns)
    for (let row = 0; row < 6; row++) {
        for (let col = 0; col < 7; col++) {
            const cell = document.createElement('div');
            cell.className = 'cell';
            cell.dataset.row = row;
            cell.dataset.col = col;
            cell.addEventListener('click', () => makeMove(col));
            elements.gameBoard.appendChild(cell);
        }
    }

    // Create column indicators
    for (let col = 0; col < 7; col++) {
        const indicator = document.createElement('div');
        indicator.className = 'col-indicator';
        indicator.addEventListener('click', () => makeMove(col));
        elements.columnIndicators.appendChild(indicator);
    }
}

function renderBoard(lastMove = null) {
    const cells = elements.gameBoard.querySelectorAll('.cell');

    cells.forEach((cell, index) => {
        const row = Math.floor(index / 7);
        const col = index % 7;
        const value = state.board[row][col];

        // Remove old classes
        cell.classList.remove('red', 'yellow', 'drop-animation');

        // Add appropriate class
        if (value === 1) {
            cell.classList.add('red');
        } else if (value === 2) {
            cell.classList.add('yellow');
        }

        // Animate last move
        if (lastMove && lastMove.row === row && lastMove.col === col) {
            cell.classList.add('drop-animation');
        }
    });
}

function updateTurnIndicator() {
    if (state.isMyTurn) {
        elements.gameStatus.textContent = 'Your turn';
        elements.gameStatus.classList.add('your-turn');
    } else {
        elements.gameStatus.textContent = "Opponent's turn";
        elements.gameStatus.classList.remove('your-turn');
    }
}

// ============= Leaderboard =============

async function fetchLeaderboard() {
    try {
        const response = await fetch(CONFIG.LEADERBOARD_URL);
        const data = await response.json();
        return data || [];
    } catch (error) {
        console.error('Failed to fetch leaderboard:', error);
        return [];
    }
}

function renderLeaderboard(container, scores) {
    if (!scores || scores.length === 0) {
        container.innerHTML = '<p class="loading">No scores yet. Be the first!</p>';
        return;
    }

    container.innerHTML = scores.map((score, index) => {
        let rankClass = '';
        if (index === 0) rankClass = 'gold';
        else if (index === 1) rankClass = 'silver';
        else if (index === 2) rankClass = 'bronze';

        return `
            <div class="leaderboard-item">
                <span class="rank ${rankClass}">#${index + 1}</span>
                <span class="name">${escapeHtml(score.username)}</span>
                <span class="wins">${score.wins} wins</span>
            </div>
        `;
    }).join('');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// ============= Event Listeners =============

function setupEventListeners() {
    // Play button
    elements.playBtn.addEventListener('click', () => {
        const username = elements.usernameInput.value.trim();
        if (!username) {
            elements.usernameInput.focus();
            return;
        }

        state.username = username;
        localStorage.setItem('connect4_username', username);

        state.gameStatus = 'matchmaking';
        showScreen('matchmaking-screen');

        // Start countdown
        let countdown = 10;
        elements.countdownTimer.textContent = countdown;
        state.matchmakingTimer = setInterval(() => {
            countdown--;
            elements.countdownTimer.textContent = countdown;
            if (countdown <= 0) {
                clearInterval(state.matchmakingTimer);
            }
        }, 1000);

        connectWebSocket();

        // Wait for connection before joining
        const checkConnection = setInterval(() => {
            if (state.socket && state.socket.readyState === WebSocket.OPEN) {
                clearInterval(checkConnection);
                joinQueue();
            }
        }, 100);
    });

    // Enter key in username input
    elements.usernameInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            elements.playBtn.click();
        }
    });

    // Cancel matchmaking
    elements.cancelBtn.addEventListener('click', () => {
        if (state.matchmakingTimer) {
            clearInterval(state.matchmakingTimer);
        }
        state.gameStatus = 'idle';
        if (state.socket) {
            state.socket.close();
        }
        showScreen('login-screen');
    });

    // Play again
    elements.playAgainBtn.addEventListener('click', () => {
        elements.gameOverBanner.classList.add('hidden');
        state.gameStatus = 'idle';
        state.board = Array(6).fill(null).map(() => Array(7).fill(0));
        initializeBoard();
        showScreen('login-screen');
        fetchLeaderboard().then(scores => {
            renderLeaderboard(elements.leaderboardLogin, scores);
        });
    });
}

// ============= Initialization =============

async function init() {
    // Initialize board
    initializeBoard();

    // Set up event listeners
    setupEventListeners();

    // Load leaderboard
    const scores = await fetchLeaderboard();
    renderLeaderboard(elements.leaderboardLogin, scores);

    // Check for stored username
    const storedUsername = localStorage.getItem('connect4_username');
    if (storedUsername) {
        elements.usernameInput.value = storedUsername;
    }

    console.log('Connect 4 initialized');
}

// Start the app
init();
