// Package game implements deterministic falling-block rules without depending
// on rendering or physical input APIs.
package game

import "math/rand/v2"

const (
	BoardWidth     = 10
	BoardHeight    = 22
	LockDelayTicks = 24 // 0.4 seconds at 60 updates/second
)

type Point struct{ X, Y int }

type Special int

const (
	SpecialNone Special = iota
	SpecialAntidote
	SpecialClear
	SpecialBlind
	SpecialInverse
	SpecialFaster   // Rabbit: speeds up the selected target.
	SpecialSlower   // Turtle: slows down the collector.
	SpecialBridge   // Adds two disruptive rows to the selected target.
	SpecialQuestion // Removes half of the selected target's placed blocks.
	SpecialStair    // Builds a diagonal staircase on the selected target.
	SpecialFill     // Fills the selected target's lower ten rows with one hole each.
)

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
	Board          [BoardHeight][BoardWidth]int
	Specials       [BoardHeight][BoardWidth]Special
	Active         Piece
	NextKind       int
	Score          int
	Lines          int
	GameOver       bool
	Antidotes      int
	Blind          bool
	Inverse        bool
	FasterStacks   int
	SlowerBonus    int
	pendingSpecial []Special
	pendingGarbage int
	random         *rand.Rand
	fallTick       int
	lockTick       int
	specialTick    int
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

// MoveInput applies player-directed horizontal movement, including effects.
func (g *Game) MoveInput(dx int) bool {
	if g.Inverse {
		dx = -dx
	}
	return g.Move(dx)
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
	g.resetLockDelayIfAirborne()
	return true
}

// RotateInput applies player-directed rotation, including effects.
func (g *Game) RotateInput(direction int) bool {
	if g.Inverse {
		direction = -direction
	}
	return g.Rotate(direction)
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
			g.resetLockDelayIfAirborne()
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
		g.lockTick = 0
		return true
	}
	return false
}

func (g *Game) HardDrop() {
	if g.GameOver {
		return
	}
	for g.StepDown() {
	}
	g.lock()
}

func (g *Game) Tick() {
	if g.GameOver {
		return
	}
	if g.grounded() {
		g.lockTick++
		if g.lockTick >= LockDelayTicks {
			g.lock()
			return
		}
	} else {
		g.lockTick = 0
	}
	g.specialTick++
	if g.specialTick == 22*60 {
		g.removeSpecial()
	} else if g.specialTick >= 30*60 {
		g.spawnSpecial()
		g.specialTick = 0
	}
	g.fallTick++
	if g.fallTick >= g.GravityTicks() {
		g.StepDown()
		g.fallTick = 0
	}
}

func (g *Game) GravityTicks() int {
	// Start at a relaxed 750 ms for level 0, then keep the original Eit
	// acceleration factor of 1.07 per level. Ebitengine updates at 60 Hz.
	ticks := 45
	for level := 0; level < g.Lines/5; level++ {
		ticks = ticks * 100 / 107
	}
	for range g.FasterStacks {
		ticks = ticks * 3 / 4
	}
	ticks += g.SlowerBonus
	if ticks < 3 {
		return 3
	}
	return ticks
}

func (g *Game) grounded() bool {
	candidate := g.Active
	candidate.Y++
	return !g.valid(candidate)
}

func (g *Game) resetLockDelayIfAirborne() {
	if !g.grounded() {
		g.lockTick = 0
	}
}

func (g *Game) lock() {
	g.lockTick = 0
	g.fallTick = 0
	for _, cell := range g.Cells(g.Active) {
		if cell.Y < 0 {
			g.GameOver = true
			return
		}
		g.Board[cell.Y][cell.X] = g.Active.Kind + 1
	}
	cleared, activated := g.clearLines()
	level := g.Lines / 5
	g.Lines += cleared
	scores := [...]int{0, 40, 100, 300, 1200}
	g.Score += scores[cleared] * (level + 1)
	if cleared == 4 {
		g.pendingGarbage += 2
	}
	for _, special := range activated {
		g.activateSpecial(special)
	}
	g.Spawn()
}

// QueueAttack adds outgoing garbage rows. Besides four-line clears, later
// special-block effects can use the same match-facing mechanism.
func (g *Game) QueueAttack(rows int) {
	if rows > 0 {
		g.pendingGarbage += rows
	}
}

// ConsumeGarbage returns and clears rows earned by line clears. In original
// Eit, clearing four rows attacks the selected opponent with two rows.
func (g *Game) ConsumeGarbage() int {
	rows := g.pendingGarbage
	g.pendingGarbage = 0
	return rows
}

// AddGarbage adds incomplete rows immediately above the current stack, matching
// the original game's disruptive top-of-stack attack rather than pushing the
// entire board upward from the bottom.
func (g *Game) AddGarbage(rows int) {
	for range rows {
		row := BoardHeight - 1
		for y := 0; y < BoardHeight; y++ {
			occupied := false
			for x := 0; x < BoardWidth; x++ {
				if g.Board[y][x] != 0 {
					occupied = true
					break
				}
			}
			if occupied {
				row = y - 1
				break
			}
		}
		if row < 0 {
			g.GameOver = true
			return
		}
		hole := g.random.IntN(BoardWidth)
		for x := 0; x < BoardWidth; x++ {
			if x != hole {
				g.Board[row][x] = 1 + g.random.IntN(len(shapes))
			}
		}
		for _, cell := range g.Cells(g.Active) {
			if cell.Y == row && g.Board[cell.Y][cell.X] != 0 {
				g.GameOver = true
				return
			}
		}
	}
}

func (g *Game) SpawnSpecial(special Special, point Point) bool {
	if special == SpecialNone || point.X < 0 || point.X >= BoardWidth || point.Y < 0 || point.Y >= BoardHeight || g.Board[point.Y][point.X] == 0 {
		return false
	}
	g.removeSpecial()
	g.Specials[point.Y][point.X] = special
	return true
}

func (g *Game) spawnSpecial() {
	occupied := make([]Point, 0)
	for y, row := range g.Board {
		for x, value := range row {
			if value != 0 {
				occupied = append(occupied, Point{X: x, Y: y})
			}
		}
	}
	if len(occupied) == 0 {
		return
	}
	special := SpecialAntidote
	switch g.random.IntN(10) {
	case 1:
		special = SpecialClear
	case 2:
		special = SpecialBlind
	case 3:
		special = SpecialInverse
	case 4:
		special = SpecialFaster
	case 5:
		special = SpecialSlower
	case 6:
		special = SpecialBridge
	case 7:
		special = SpecialQuestion
	case 8:
		special = SpecialStair
	case 9:
		special = SpecialFill
	}
	g.SpawnSpecial(special, occupied[g.random.IntN(len(occupied))])
}

func (g *Game) removeSpecial() {
	g.Specials = [BoardHeight][BoardWidth]Special{}
}

func (g *Game) activateSpecial(special Special) {
	switch special {
	case SpecialAntidote:
		if g.Antidotes < 4 {
			g.Antidotes++
		}
	case SpecialClear:
		g.Board = [BoardHeight][BoardWidth]int{}
		g.Specials = [BoardHeight][BoardWidth]Special{}
	case SpecialBlind, SpecialInverse, SpecialFaster, SpecialBridge, SpecialQuestion, SpecialStair, SpecialFill:
		g.pendingSpecial = append(g.pendingSpecial, special)
	case SpecialSlower:
		// Original Eit's Turtle is a small permanent slowdown for the collector.
		g.SlowerBonus++
	}
}

// QueueSpecial adds an outgoing effect for the selected target.
func (g *Game) QueueSpecial(special Special) {
	if special != SpecialNone {
		g.pendingSpecial = append(g.pendingSpecial, special)
	}
}

func (g *Game) ConsumeSpecials() []Special {
	result := append([]Special(nil), g.pendingSpecial...)
	g.pendingSpecial = g.pendingSpecial[:0]
	return result
}

func (g *Game) ApplySpecial(special Special) {
	switch special {
	case SpecialBlind:
		g.Blind = true
	case SpecialInverse:
		g.Inverse = true
	case SpecialFaster:
		g.FasterStacks++
	case SpecialBridge:
		g.AddGarbage(2)
	case SpecialQuestion:
		g.removeRandomHalf()
	case SpecialStair:
		g.addStair()
	case SpecialFill:
		g.addFill()
	}
}

func (g *Game) setPatternCell(x, y int, occupied bool) {
	if !occupied {
		g.Board[y][x] = 0
		g.Specials[y][x] = SpecialNone
		return
	}
	g.Board[y][x] = 1 + g.random.IntN(len(shapes))
	g.Specials[y][x] = SpecialNone
}

func (g *Game) addStair() {
	// Ported from the original pattern: one diagonal block per row, with the
	// neighbouring cells explicitly cleared so the staircase stays readable.
	g.setPatternCell(0, 21, true)
	g.setPatternCell(1, 21, false)
	for x := 1; x <= 8; x++ {
		y := 21 - x
		g.setPatternCell(x-1, y, false)
		g.setPatternCell(x, y, true)
		g.setPatternCell(x+1, y, false)
	}
	g.setPatternCell(8, 12, false)
	g.setPatternCell(9, 12, true)
	g.checkActiveCollision()
}

func (g *Game) addFill() {
	for y := 21; y >= 12; y-- {
		hole := g.random.IntN(BoardWidth)
		for x := 0; x < BoardWidth; x++ {
			g.setPatternCell(x, y, x != hole)
		}
	}
	g.checkActiveCollision()
}

func (g *Game) checkActiveCollision() {
	if !g.valid(g.Active) {
		g.GameOver = true
	}
}

func (g *Game) removeRandomHalf() {
	occupied := make([]Point, 0)
	for y, row := range g.Board {
		for x, value := range row {
			if value != 0 {
				occupied = append(occupied, Point{X: x, Y: y})
			}
		}
	}
	remove := len(occupied) / 2
	g.random.Shuffle(len(occupied), func(i, j int) {
		occupied[i], occupied[j] = occupied[j], occupied[i]
	})
	for _, point := range occupied[:remove] {
		g.Board[point.Y][point.X] = 0
		g.Specials[point.Y][point.X] = SpecialNone
	}
}

func (g *Game) HasActiveEffect() bool {
	return g.Blind || g.Inverse || g.FasterStacks > 0 || g.SlowerBonus > 0
}

func (g *Game) UseAntidote() bool {
	if g.Antidotes == 0 || !g.HasActiveEffect() {
		return false
	}
	g.Antidotes--
	g.Blind = false
	g.Inverse = false
	g.FasterStacks = 0
	g.SlowerBonus = 0
	return true
}

func (g *Game) clearLines() (int, []Special) {
	write := BoardHeight - 1
	cleared := 0
	activated := make([]Special, 0)
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
			for _, special := range g.Specials[read] {
				if special != SpecialNone {
					activated = append(activated, special)
				}
			}
			continue
		}
		g.Board[write] = g.Board[read]
		g.Specials[write] = g.Specials[read]
		write--
	}
	for write >= 0 {
		g.Board[write] = [BoardWidth]int{}
		g.Specials[write] = [BoardWidth]Special{}
		write--
	}
	return cleared, activated
}
