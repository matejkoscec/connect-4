package game

import (
	"fmt"
	"github.com/google/uuid"
	"sync"
)

type Color uint8

const (
	ColorNone Color = iota
	ColorRed
	ColorYellow
)

type Move struct {
	Column uint8
	Color  Color
}

const (
	Rows = 6
	Cols = 7
)

type Game struct {
	mu sync.Mutex

	Id    uuid.UUID
	Moves []Move
	State *Board
}

type Board [Rows][Cols]Color

func New() (*Game, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}

	var state Board
	for i := 0; i < Rows; i++ {
		var row [Cols]Color
		for j := 0; j < Cols; j++ {
			row[j] = ColorNone
		}
		state[i] = row
	}

	return &Game{
		mu:    sync.Mutex{},
		Id:    id,
		Moves: []Move{},
		State: &state,
	}, nil
}

func (g *Game) Make(move Move) (bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	board := g.State
	lastI := Rows - 1
	if len(g.Moves) == 0 {
		if move.Color == ColorYellow {
			return false, fmt.Errorf("not %d turn", move.Color)
		}
		board[lastI][move.Column] = move.Color
		g.Moves = append(g.Moves, move)
		for _, row := range board {
			fmt.Println(row)
		}
		fmt.Println()
		return false, nil
	}

	lastMove := g.Moves[len(g.Moves)-1]
	if move.Color == lastMove.Color {
		return false, fmt.Errorf("not %d turn", move.Color)
	}
	if board[0][move.Column] != ColorNone {
		return false, fmt.Errorf("column %d is full", move.Column)
	}

	for board[lastI][move.Column] != ColorNone {
		lastI--
	}
	board[lastI][move.Column] = move.Color
	g.Moves = append(g.Moves, move)

	for _, row := range board {
		fmt.Println(row)
	}
	fmt.Println()

	return isWinningMove(board, lastI, move), nil
}

func isWinningMove(board *Board, lastI int, move Move) bool {
	row := lastI
	col := move.Column
	playerColor := move.Color

	for startCol := col - 3; startCol <= col; startCol++ {
		if startCol < 0 {
			continue
		}
		if startCol+3 >= Cols {
			continue
		}

		if board[row][startCol] == playerColor &&
			board[row][startCol+1] == playerColor &&
			board[row][startCol+2] == playerColor &&
			board[row][startCol+3] == playerColor {
			return true
		}
	}

	count := 0
	for i := row; i < Rows; i++ {
		if board[i][col] == playerColor {
			count++
		} else {
			break
		}
	}
	if count >= 4 {
		return true
	}

	r, c := row, col
	for r > 0 && c > 0 {
		r--
		c--
	}

	count = 0
	for i, j := r, c; i < Rows && j < Cols; i, j = i+1, j+1 {
		if board[i][j] == playerColor {
			count++
			if count >= 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	r, c = row, col
	for r > 0 && c < Cols-1 {
		r--
		c++
	}

	count = 0
	for i, j := r, c; i < Rows; i, j = i+1, j-1 {
		if board[i][j] == playerColor {
			count++
			if count >= 4 {
				return true
			}
		} else {
			count = 0
		}
	}

	return false
}
