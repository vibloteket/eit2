// Package game implements deterministic falling-block rules without depending
// on rendering or physical input APIs.
package game

import "math/rand/v2"

const (
	BoardWidth  = 10
	BoardHeight = 22
)

type Point struct{ X, Y int }

type Piece struct {
	Kind     int
	Rotation int
	X, Y     int
}

var shapes = [7][4][4]Point{
	{ // I
		{{0, 1}, {1, 1}, {2, 1}, {3, 1}}, {{2, 0}, {2, 1}, {2, 2}, {2, 3}}, {{0, 2}, {1, 2}, {2, 2}, {3, 2}}, {{1, 0}, {1, 1}, {1, 2}, {1, 3}},
	},
	{ // O
		{{1, 0}, {2, 0}, {1, 1}, {2, 1}}, {{1, 0}, {2, 0}, {1, 1}, {2, 1}}, {{1, 0}, {2, 0}, {1, 1}, {2, 1}}, {{1, 0}, {2, 0}, {1, 1}, {2, 1}},
	},
	{ // T
		{{1, 0}, {0, 1}, {1, 1}, {2, 1}}, {{1, 0}, {1, 1}, {2, 1}, {1, 2}}, {{0, 1}, {1, 1}, {2, 1}, {1, 2}}, {{1, 0}, {0, 1}, {1, 1}, {1, 2}},
	},
	{ // L
		{{2, 0}, {0, 1}, {1, 1}, {2, 1}}, {{1, 0}, {1, 1}, {1, 2}, {2, 2}}, {{0, 1}, {1, 1}, {2, 1}, {0, 2}}, {{0, 0}, {1, 0}, {1, 1}, {1, 2}},
	},
	{ // J
		{{0, 0}, {0, 1}, {1, 1}, {2, 1}}, {{1, 0}, {2, 0}, {1, 1}, {1, 2}}, {{0, 1}, {1, 1}, {2, 1}, {2, 2}}, {{1, 0}, {1, 1}, {0, 2}, {1, 2}},
	},
	{ // S
		{{1, 0}, {2, 0}, {0, 1}, {1, 1}}, {{1, 0}, {1, 1}, {2, 1}, {2, 2}}, {{1, 1}, {2, 1}, {0, 2}, {1, 2}}, {{0, 0}, {0, 1}, {1, 1}, {1, 2}},
	},
	{ // Z
		{{0, 0}, {1, 0}, {1, 1}, {2, 1}}, {{2, 0}, {1, 1}, {2, 1}, {1, 2}}, {{0, 1}, {1, 1}, {1, 2}, {2, 2}}, {{1, 0}, {0, 1}, {1, 1}, {0, 2}},
	},
}

type Game struct {
	Board    [BoardHeight][BoardWidth]int
	Active   Piece
	NextKind int
	Score    int
	Lines    int
	GameOver bool
	random   *rand.Rand
	fallTick int
}

func New(seed uint64) *Game {
	g := &Game{random: rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15)), NextKind: -1}
	g.Spawn()
	return g
}

func (g *Game) Cells(piece Piece) [4]Point {
	return PieceCells(piece)
}

func PieceCells(piece Piece) [4]Point {
	var cells [4]Point
	for i, point := range shapes[piece.Kind][piece.Rotation%4] {
		cells[i] = Point{X: piece.X + point.X, Y: piece.Y + point.Y}
	}
	return cells
}

func (g *Game) valid(piece Piece) bool {
	for _, cell := range g.Cells(piece) {
		if cell.X < 0 || cell.X >= BoardWidth || cell.Y >= BoardHeight {
			return false
		}
		if cell.Y >= 0 && g.Board[cell.Y][cell.X] != 0 {
			return false
		}
	}
	return true
}

func (g *Game) Spawn() {
	kind := g.NextKind
	if kind < 0 {
		kind = g.random.IntN(len(shapes))
	}
	g.Active = Piece{Kind: kind, X: 3, Y: -1}
	g.NextKind = g.random.IntN(len(shapes))
	if !g.valid(g.Active) {
		g.GameOver = true
	}
}

func (g *Game) Move(dx int) bool {
	if g.GameOver {
		return false
	}
	candidate := g.Active
	candidate.X += dx
	if !g.valid(candidate) {
		return false
	}
	g.Active = candidate
	return true
}

func (g *Game) Rotate(direction int) bool {
	if g.GameOver {
		return false
	}
	candidate := g.Active
	candidate.Rotation = (candidate.Rotation + direction + 4) % 4
	for _, kick := range []int{0, -1, 1, -2, 2} {
		kicked := candidate
		kicked.X += kick
		if g.valid(kicked) {
			g.Active = kicked
			return true
		}
	}
	return false
}

func (g *Game) StepDown() bool {
	if g.GameOver {
		return false
	}
	candidate := g.Active
	candidate.Y++
	if g.valid(candidate) {
		g.Active = candidate
		return true
	}
	g.lock()
	return false
}

func (g *Game) HardDrop() {
	for g.StepDown() {
	}
}

func (g *Game) Tick() {
	if g.GameOver {
		return
	}
	g.fallTick++
	if g.fallTick >= 30 {
		g.StepDown()
		g.fallTick = 0
	}
}

func (g *Game) lock() {
	for _, cell := range g.Cells(g.Active) {
		if cell.Y < 0 {
			g.GameOver = true
			return
		}
		g.Board[cell.Y][cell.X] = g.Active.Kind + 1
	}
	cleared := g.clearLines()
	level := g.Lines / 5
	g.Lines += cleared
	scores := [...]int{0, 40, 100, 300, 1200}
	g.Score += scores[cleared] * (level + 1)
	g.Spawn()
}

func (g *Game) clearLines() int {
	write := BoardHeight - 1
	cleared := 0
	for read := BoardHeight - 1; read >= 0; read-- {
		full := true
		for x := 0; x < BoardWidth; x++ {
			if g.Board[read][x] == 0 {
				full = false
				break
			}
		}
		if full {
			cleared++
			continue
		}
		g.Board[write] = g.Board[read]
		write--
	}
	for write >= 0 {
		g.Board[write] = [BoardWidth]int{}
		write--
	}
	return cleared
}
