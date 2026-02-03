// bot.go implements a competitive Connect 4 AI using Minimax with alpha-beta pruning.
// The bot never makes random moves - it always plays optimally within its search depth.
package game

import (
	"math"
)

// Bot AI configuration
const (
	MaxDepth    = 6 // Search depth - higher = stronger but slower
	WinScore    = 100000
	ThreeScore  = 100
	TwoScore    = 10
	CenterBonus = 3
)

// Column ordering for better alpha-beta pruning (center first)
var columnOrder = []int{3, 2, 4, 1, 5, 0, 6}

// GetBotMove calculates the best move for the bot (Yellow/Player2).
// Uses Minimax with alpha-beta pruning for competitive play.
// CRITICAL: This operates on a CLONE of the board to avoid mutating game state.
func GetBotMove(g *Game) int {
	g.mu.RLock()
	board := g.Board.Clone()
	g.mu.RUnlock()

	// Priority 1: Win immediately if possible
	if winCol := board.HasWinningMove(Yellow); winCol >= 0 {
		return winCol
	}

	// Priority 2: Block opponent's immediate win
	if blockCol := board.HasWinningMove(Red); blockCol >= 0 {
		return blockCol
	}

	// Priority 3: Use Minimax to find best strategic move
	_, bestCol := minimax(&board, MaxDepth, math.MinInt32, math.MaxInt32, true)

	// Fallback to center if something went wrong
	if bestCol == -1 {
		for _, col := range columnOrder {
			if board.CanDrop(col) {
				return col
			}
		}
	}

	return bestCol
}

// minimax implements the Minimax algorithm with alpha-beta pruning.
// Parameters:
//   - board: current board state (CLONED, safe to mutate)
//   - depth: remaining search depth
//   - alpha: best score for maximizer
//   - beta: best score for minimizer
//   - maximizing: true if bot's turn (Yellow), false if opponent's turn (Red)
//
// Returns: (score, bestColumn)
func minimax(board *Board, depth int, alpha, beta int, maximizing bool) (int, int) {
	// Terminal conditions
	winner := board.CheckWinFull()
	if winner == Yellow {
		return WinScore + depth, -1 // Prefer faster wins
	}
	if winner == Red {
		return -WinScore - depth, -1 // Opponent winning is bad
	}
	if board.IsFull() {
		return 0, -1 // Draw
	}
	if depth == 0 {
		return evaluateBoard(board), -1
	}

	validCols := getOrderedMoves(board)
	if len(validCols) == 0 {
		return 0, -1
	}

	bestCol := validCols[0]

	if maximizing {
		// Bot's turn (Yellow) - maximize score
		maxEval := math.MinInt32

		for _, col := range validCols {
			// Clone board for simulation
			clone := board.Clone()
			clone.Drop(col, Yellow)

			eval, _ := minimax(&clone, depth-1, alpha, beta, false)

			if eval > maxEval {
				maxEval = eval
				bestCol = col
			}

			alpha = max(alpha, eval)
			if beta <= alpha {
				break // Beta cutoff
			}
		}

		return maxEval, bestCol

	} else {
		// Opponent's turn (Red) - minimize score
		minEval := math.MaxInt32

		for _, col := range validCols {
			clone := board.Clone()
			clone.Drop(col, Red)

			eval, _ := minimax(&clone, depth-1, alpha, beta, true)

			if eval < minEval {
				minEval = eval
				bestCol = col
			}

			beta = min(beta, eval)
			if beta <= alpha {
				break // Alpha cutoff
			}
		}

		return minEval, bestCol
	}
}

// getOrderedMoves returns valid columns ordered for better pruning
// Center columns first (more strategically valuable)
func getOrderedMoves(board *Board) []int {
	var moves []int
	for _, col := range columnOrder {
		if board.CanDrop(col) {
			moves = append(moves, col)
		}
	}
	return moves
}

// evaluateBoard calculates a heuristic score for the current board position.
// Positive scores favor Yellow (bot), negative scores favor Red (opponent).
func evaluateBoard(board *Board) int {
	score := 0

	// Center column control is strategically important
	for row := 0; row < Rows; row++ {
		if board[row][3] == Yellow {
			score += CenterBonus
		} else if board[row][3] == Red {
			score -= CenterBonus
		}
	}

	// Evaluate all possible windows of 4
	score += evaluateWindows(board, Yellow) - evaluateWindows(board, Red)

	return score
}

// evaluateWindows scores all 4-cell windows for a player
func evaluateWindows(board *Board, player Cell) int {
	score := 0
	opponent := Red
	if player == Red {
		opponent = Yellow
	}

	// Horizontal windows
	for r := 0; r < Rows; r++ {
		for c := 0; c <= Cols-4; c++ {
			score += evaluateWindow(board, r, c, 0, 1, player, opponent)
		}
	}

	// Vertical windows
	for r := 0; r <= Rows-4; r++ {
		for c := 0; c < Cols; c++ {
			score += evaluateWindow(board, r, c, 1, 0, player, opponent)
		}
	}

	// Diagonal (down-right)
	for r := 0; r <= Rows-4; r++ {
		for c := 0; c <= Cols-4; c++ {
			score += evaluateWindow(board, r, c, 1, 1, player, opponent)
		}
	}

	// Diagonal (down-left)
	for r := 0; r <= Rows-4; r++ {
		for c := 3; c < Cols; c++ {
			score += evaluateWindow(board, r, c, 1, -1, player, opponent)
		}
	}

	return score
}

// evaluateWindow scores a single 4-cell window
func evaluateWindow(board *Board, startR, startC, dR, dC int, player, opponent Cell) int {
	playerCount := 0
	emptyCount := 0

	for i := 0; i < 4; i++ {
		cell := board[startR+i*dR][startC+i*dC]
		switch cell {
		case player:
			playerCount++
		case Empty:
			emptyCount++
		case opponent:
			// Window blocked by opponent, no potential
			return 0
		}
	}

	// Score based on how many of our pieces and empty spaces
	if playerCount == 4 {
		return WinScore
	} else if playerCount == 3 && emptyCount == 1 {
		return ThreeScore
	} else if playerCount == 2 && emptyCount == 2 {
		return TwoScore
	}

	return 0
}

// Helper functions for Go versions < 1.21
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
