// board.go implements the Connect 4 board grid with gravity logic.
// All board operations are pure functions that work on Board copies to prevent mutation.
package game

// NewBoard creates an empty game board
func NewBoard() Board {
	return Board{} // All cells default to Empty (0)
}

// Clone creates a deep copy of the board for simulation
// This is critical for bot AI to simulate moves without mutating game state
func (b Board) Clone() Board {
	var clone Board
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			clone[r][c] = b[r][c]
		}
	}
	return clone
}

// CanDrop checks if a column has space for a disc
func (b *Board) CanDrop(col int) bool {
	if col < 0 || col >= Cols {
		return false
	}
	// Column is playable if top row is empty
	return b[0][col] == Empty
}

// Drop places a disc in the specified column with gravity.
// Returns the row where the disc landed, or -1 if column is full/invalid.
// IMPORTANT: This mutates the board - only call on actual game state or clones.
func (b *Board) Drop(col int, player Cell) int {
	if col < 0 || col >= Cols {
		return -1
	}

	// Find the lowest empty row (gravity simulation)
	for row := Rows - 1; row >= 0; row-- {
		if b[row][col] == Empty {
			b[row][col] = player
			return row
		}
	}

	return -1 // Column is full
}

// GetValidColumns returns all columns that can accept a disc
func (b *Board) GetValidColumns() []int {
	var valid []int
	for col := 0; col < Cols; col++ {
		if b.CanDrop(col) {
			valid = append(valid, col)
		}
	}
	return valid
}

// IsFull checks if the entire board is filled (draw condition)
func (b *Board) IsFull() bool {
	for col := 0; col < Cols; col++ {
		if b[0][col] == Empty {
			return false
		}
	}
	return true
}

// GetCell safely retrieves a cell value, returns Empty for out-of-bounds
func (b *Board) GetCell(row, col int) Cell {
	if row < 0 || row >= Rows || col < 0 || col >= Cols {
		return Empty
	}
	return b[row][col]
}

// CountDiscs returns total number of discs on the board
func (b *Board) CountDiscs() int {
	count := 0
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			if b[r][c] != Empty {
				count++
			}
		}
	}
	return count
}

// String returns a visual representation of the board for debugging
func (b *Board) String() string {
	result := ""
	for r := 0; r < Rows; r++ {
		for c := 0; c < Cols; c++ {
			switch b[r][c] {
			case Empty:
				result += ". "
			case Red:
				result += "R "
			case Yellow:
				result += "Y "
			}
		}
		result += "\n"
	}
	result += "0 1 2 3 4 5 6\n"
	return result
}
