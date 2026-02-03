// rules.go implements win detection for Connect 4.
// Uses optimized checking that only examines cells around the last move.
package game

// Direction vectors for checking 4-in-a-row
// Each direction is paired with its opposite for bidirectional counting
type direction struct {
	dRow, dCol int
}

var (
	// Horizontal: left-right
	horizontal = [2]direction{{0, -1}, {0, 1}}
	// Vertical: up-down
	vertical = [2]direction{{-1, 0}, {1, 0}}
	// Diagonal: top-left to bottom-right
	diagonal1 = [2]direction{{-1, -1}, {1, 1}}
	// Diagonal: top-right to bottom-left
	diagonal2 = [2]direction{{-1, 1}, {1, -1}}

	allDirections = [][2]direction{horizontal, vertical, diagonal1, diagonal2}
)

// CheckWinAt checks if the last move at (row, col) creates a winning condition.
// This is an optimized approach - only check lines that include the last move.
// Returns the winning player (Red/Yellow) or Empty if no win.
func (b *Board) CheckWinAt(row, col int) Cell {
	player := b[row][col]
	if player == Empty {
		return Empty
	}

	for _, dirs := range allDirections {
		count := 1 // Start with the placed disc

		// Count in first direction
		count += b.countInDirection(row, col, dirs[0].dRow, dirs[0].dCol, player)

		// Count in opposite direction
		count += b.countInDirection(row, col, dirs[1].dRow, dirs[1].dCol, player)

		if count >= 4 {
			return player
		}
	}

	return Empty
}

// countInDirection counts consecutive matching cells in one direction
func (b *Board) countInDirection(startRow, startCol, dRow, dCol int, player Cell) int {
	count := 0
	r, c := startRow+dRow, startCol+dCol

	for r >= 0 && r < Rows && c >= 0 && c < Cols && b[r][c] == player {
		count++
		r += dRow
		c += dCol
	}

	return count
}

// CheckWinFull does a complete board scan for winners.
// Used when we don't know the last move (e.g., loading saved game).
// Less efficient but comprehensive.
func (b *Board) CheckWinFull() Cell {
	// Check horizontal wins
	for r := 0; r < Rows; r++ {
		for c := 0; c <= Cols-4; c++ {
			player := b[r][c]
			if player != Empty &&
				player == b[r][c+1] &&
				player == b[r][c+2] &&
				player == b[r][c+3] {
				return player
			}
		}
	}

	// Check vertical wins
	for r := 0; r <= Rows-4; r++ {
		for c := 0; c < Cols; c++ {
			player := b[r][c]
			if player != Empty &&
				player == b[r+1][c] &&
				player == b[r+2][c] &&
				player == b[r+3][c] {
				return player
			}
		}
	}

	// Check diagonal (top-left to bottom-right)
	for r := 0; r <= Rows-4; r++ {
		for c := 0; c <= Cols-4; c++ {
			player := b[r][c]
			if player != Empty &&
				player == b[r+1][c+1] &&
				player == b[r+2][c+2] &&
				player == b[r+3][c+3] {
				return player
			}
		}
	}

	// Check diagonal (top-right to bottom-left)
	for r := 0; r <= Rows-4; r++ {
		for c := 3; c < Cols; c++ {
			player := b[r][c]
			if player != Empty &&
				player == b[r+1][c-1] &&
				player == b[r+2][c-2] &&
				player == b[r+3][c-3] {
				return player
			}
		}
	}

	return Empty
}

// GetWinResult evaluates the board state and returns the game result
func (b *Board) GetWinResult() WinResult {
	winner := b.CheckWinFull()
	switch winner {
	case Red:
		return Player1
	case Yellow:
		return Player2
	default:
		if b.IsFull() {
			return Draw
		}
		return NoWinner
	}
}

// HasWinningMove checks if the given player can win with the next move
// Returns the winning column or -1 if no immediate win
func (b *Board) HasWinningMove(player Cell) int {
	for col := 0; col < Cols; col++ {
		if !b.CanDrop(col) {
			continue
		}

		// Clone and simulate
		clone := b.Clone()
		row := clone.Drop(col, player)
		if row >= 0 && clone.CheckWinAt(row, col) == player {
			return col
		}
	}
	return -1
}
