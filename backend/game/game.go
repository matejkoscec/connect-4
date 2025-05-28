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

func (g *Game) Make(move Move) (int, bool, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	board := g.State
	lastI := Rows - 1
	if len(g.Moves) == 0 {
		if move.Color == ColorYellow {
			return lastI, false, fmt.Errorf("not %d turn", move.Color)
		}
		board[lastI][move.Column] = move.Color
		g.Moves = append(g.Moves, move)
		return lastI, false, nil
	}

	lastMove := g.Moves[len(g.Moves)-1]
	if move.Color == lastMove.Color {
		return 0, false, fmt.Errorf("not %d turn", move.Color)
	}
	if board[0][move.Column] != ColorNone {
		return 0, false, fmt.Errorf("column %d is full", move.Column)
	}

	for board[lastI][move.Column] != ColorNone {
		lastI--
	}
	board[lastI][move.Column] = move.Color
	g.Moves = append(g.Moves, move)

	return lastI, isWinningMove(board, lastI, move), nil
}

func isWinningMove(board *Board, lastI int, move Move) bool {
	color := move.Color
	col := int(move.Column)
	row := lastI

	hCount := 1
	for c := col - 1; c >= 0 && board[row][c] == color; c-- {
		hCount++
	}

	for c := col + 1; c < Cols && board[row][c] == color; c++ {
		hCount++
	}
	if hCount >= 4 {
		return true
	}

	vCount := 1
	for r := row + 1; r < Rows && board[r][col] == color; r++ {
		vCount++
	}
	if vCount >= 4 {
		return true
	}

	d1Count := 1
	for r, c := row-1, col+1; r >= 0 && c < Cols && board[r][c] == color; r, c = r-1, c+1 {
		d1Count++
	}

	for r, c := row+1, col-1; r < Rows && c >= 0 && board[r][c] == color; r, c = r+1, c-1 {
		d1Count++
	}
	if d1Count >= 4 {
		return true
	}

	d2Count := 1
	for r, c := row-1, col-1; r >= 0 && c >= 0 && board[r][c] == color; r, c = r-1, c-1 {
		d2Count++
	}

	for r, c := row+1, col+1; r < Rows && c < Cols && board[r][c] == color; r, c = r+1, c+1 {
		d2Count++
	}
	if d2Count >= 4 {
		return true
	}

	return false
}
