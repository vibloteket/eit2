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

func (s Special) Name() string {
	switch s {
	case SpecialAntidote:
		return "Antidote"
	case SpecialClear:
		return "Clear"
	case SpecialBlind:
		return "Blind"
	case SpecialInverse:
		return "Inverse"
	case SpecialFaster:
		return "Rabbit / Faster"
	case SpecialSlower:
		return "Turtle / Slower"
	case SpecialBridge:
		return "Bridge"
	case SpecialQuestion:
		return "Question"
	case SpecialStair:
		return "Stair"
	case SpecialFill:
		return "Fill"
	case SpecialFlip:
		return "Flip"
	case SpecialSwitch:
		return "Switch"
	case SpecialPacket:
		return "Packet"
	case SpecialRing:
		return "Ring"
	case SpecialMini:
		return "Mini"
	case SpecialBlink:
		return "Blink"
	case SpecialSZ:
		return "SZ"
	case SpecialTrans:
		return "Ice"
	case SpecialCastle:
		return "Castle"
	case SpecialColor:
		return "Blackout"
	case SpecialRumble:
		return "Rumble"
	case SpecialBackground:
		return "Background"
	default:
		return "Unknown"
	}
}

const (
	SpecialNone Special = iota
	SpecialAntidote
	SpecialClear
	SpecialBlind
	SpecialInverse
	SpecialFaster     // Rabbit: speeds up the selected target.
	SpecialSlower     // Turtle: slows down the collector.
	SpecialBridge     // Adds two disruptive rows to the selected target.
	SpecialQuestion   // Removes half of the selected target's placed blocks.
	SpecialStair      // Builds a diagonal staircase on the selected target.
	SpecialFill       // Fills the selected target's lower ten rows with one hole each.
	SpecialFlip       // Vertically flips the selected target's placed structure.
	SpecialSwitch     // Swaps placed structures with the selected target.
	SpecialPacket     // Sends one garbage row per cleared row for 20 seconds.
	SpecialRing       // Builds a hollow ring on the selected target.
	SpecialMini       // Renders the selected target's settled blocks smaller.
	SpecialBlink      // Makes the selected target's active piece blink.
	SpecialSZ         // Restricts the selected target's future pieces to S and Z.
	SpecialTrans      // Renders the selected target's settled blocks translucent.
	SpecialCastle     // Replaces the selected target's board with a castle.
	SpecialColor      // Blackout: renders the selected target's settled blocks dark.
	SpecialRumble     // Shakes a selection of the target's settled blocks.
	SpecialBackground // Changes the selected target's playfield background.
)

type patternCell struct {
	X        int
	Occupied bool
	Value    int // 0 chooses a random standard colour; 8 is construction grey.
}

type patternRow struct {
	Y     int
	Cells []patternCell
}

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
	Board                [BoardHeight][BoardWidth]int
	Specials             [BoardHeight][BoardWidth]Special
	Active               Piece
	NextKind             int
	Score                int
	Lines                int
	GameOver             bool
	Antidotes            int
	Blind                bool
	Inverse              bool
	FasterStacks         int
	SlowerBonus          int
	PacketTicks          int
	Mini                 bool
	Blink                bool
	BlinkVisible         bool
	SZ                   bool
	Trans                bool
	Blackout             bool
	BackgroundVariant    int
	RumbleRounds         int
	rumbleTick           int
	rumblePoints         []Point
	blinkTick            int
	pendingSpecial       []Special
	pendingGarbage       int
	random               *rand.Rand
	fallTick             int
	lockTick             int
	specialTick          int
	specialLifetimeTicks int
	patternTick          int
	patternRows          []patternRow
	LastEvent            string
	EventTicks           int
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
	if kind < 0 || (g.SZ && kind != 5 && kind != 6) {
		kind = g.randomPieceKind()
	}
	g.Active = Piece{Kind: kind, X: 3, Y: -1}
	g.NextKind = g.randomPieceKind()
	if !g.valid(g.Active) {
		g.GameOver = true
	}
}

// MoveInput applies player-directed horizontal movement, including effects.
func (g *Game) randomPieceKind() int {
	if g.SZ {
		return 5 + g.random.IntN(2)
	}
	return g.random.IntN(len(shapes))
}

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
	if g.EventTicks > 0 {
		g.EventTicks--
		if g.EventTicks == 0 {
			g.LastEvent = ""
		}
	}
	g.advancePattern()
	g.advanceRumble()
	if g.PacketTicks > 0 {
		g.PacketTicks--
	}
	if g.Blink {
		g.blinkTick++
		if g.blinkTick >= 6 {
			g.BlinkVisible = !g.BlinkVisible
			g.blinkTick = 0
		}
	} else {
		g.BlinkVisible = true
		g.blinkTick = 0
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
	if g.specialLifetimeTicks > 0 && g.specialTick >= g.specialLifetimeTicks {
		g.removeSpecial()
	}
	if g.specialTick >= 30*60 {
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
	if g.PacketTicks > 0 {
		g.pendingGarbage += cleared
	}
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
	switch g.random.IntN(22) {
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
	case 10:
		special = SpecialFlip
	case 11:
		special = SpecialSwitch
	case 12:
		special = SpecialPacket
	case 13:
		special = SpecialRing
	case 14:
		special = SpecialMini
	case 15:
		special = SpecialBlink
	case 16:
		special = SpecialSZ
	case 17:
		special = SpecialTrans
	case 18:
		special = SpecialCastle
	case 19:
		special = SpecialColor
	case 20:
		special = SpecialRumble
	case 21:
		special = SpecialBackground
	}
	g.SpawnSpecial(special, occupied[g.random.IntN(len(occupied))])
	seconds := 18 + len(occupied)/10
	if seconds > 30 {
		seconds = 30
	}
	g.specialLifetimeTicks = seconds * 60
}

func (g *Game) removeSpecial() {
	g.Specials = [BoardHeight][BoardWidth]Special{}
	g.specialLifetimeTicks = 0
}

func (g *Game) SpecialLifetimeTicks() int {
	return g.specialLifetimeTicks
}

func (g *Game) ShowEvent(message string) {
	g.LastEvent = message
	g.EventTicks = 4 * 60
}

func (g *Game) activateSpecial(special Special) {
	if special == SpecialAntidote {
		g.ShowEvent("COLLECTED: " + special.Name())
	} else {
		g.ShowEvent("ACTIVATED: " + special.Name())
	}
	switch special {
	case SpecialAntidote:
		if g.Antidotes < 4 {
			g.Antidotes++
		}
	case SpecialClear:
		g.Board = [BoardHeight][BoardWidth]int{}
		g.Specials = [BoardHeight][BoardWidth]Special{}
	case SpecialBlind, SpecialInverse, SpecialFaster, SpecialBridge, SpecialQuestion, SpecialStair, SpecialFill, SpecialFlip, SpecialSwitch, SpecialRing, SpecialMini, SpecialBlink, SpecialSZ, SpecialTrans, SpecialCastle, SpecialColor, SpecialRumble, SpecialBackground:
		g.pendingSpecial = append(g.pendingSpecial, special)
	case SpecialSlower:
		// Original Eit's Turtle is a small permanent slowdown for the collector.
		g.SlowerBonus++
	case SpecialPacket:
		g.PacketTicks = 20 * 60
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
	case SpecialFlip:
		g.FlipBoard()
	case SpecialRing:
		g.addRing()
	case SpecialMini:
		g.Mini = true
	case SpecialBlink:
		g.Blink = true
		g.BlinkVisible = true
	case SpecialSZ:
		g.SZ = true
		if g.NextKind != 5 && g.NextKind != 6 {
			g.NextKind = g.randomPieceKind()
		}
	case SpecialTrans:
		g.Trans = true
	case SpecialCastle:
		g.addCastle()
	case SpecialColor:
		g.Blackout = true
	case SpecialRumble:
		g.startRumble()
	case SpecialBackground:
		g.BackgroundVariant = 1 + g.random.IntN(6)
	}
}

// FlipBoard reverses the occupied vertical extent of the settled structure,
// matching the original game's Flip behavior. The active piece is not moved.
func (g *Game) FlipBoard() {
	top := BoardHeight
	for y, row := range g.Board {
		for _, value := range row {
			if value != 0 {
				top = y
				break
			}
		}
		if top != BoardHeight {
			break
		}
	}
	if top == BoardHeight {
		return
	}
	for upper, lower := top, BoardHeight-1; upper < lower; upper, lower = upper+1, lower-1 {
		g.Board[upper], g.Board[lower] = g.Board[lower], g.Board[upper]
		g.Specials[upper], g.Specials[lower] = g.Specials[lower], g.Specials[upper]
	}
}

// SwapBoard exchanges settled blocks and their attached special markers. Match
// state, active pieces, score, effects, next pieces and queued patterns stay
// with their original players.
func (g *Game) SwapBoard(other *Game) {
	if other == nil || other == g {
		return
	}
	g.Board, other.Board = other.Board, g.Board
	g.Specials, other.Specials = other.Specials, g.Specials
}

func (g *Game) queuePattern(rows []patternRow) {
	g.patternRows = append(g.patternRows, rows...)
	g.patternTick = 0
}

func (g *Game) addStair() {
	rows := make([]patternRow, 0, 10)
	for x := 0; x < BoardWidth; x++ {
		y := BoardHeight - 1 - x
		cells := []patternCell{{X: x, Occupied: true}}
		if x > 0 {
			cells = append(cells, patternCell{X: x - 1})
		}
		if x < BoardWidth-1 {
			cells = append(cells, patternCell{X: x + 1})
		}
		rows = append(rows, patternRow{Y: y, Cells: cells})
	}
	g.queuePattern(rows)
}

func (g *Game) addFill() {
	rows := make([]patternRow, 0, 10)
	for y := BoardHeight - 1; y >= BoardHeight-10; y-- {
		hole := g.random.IntN(BoardWidth)
		cells := make([]patternCell, 0, BoardWidth)
		for x := 0; x < BoardWidth; x++ {
			cells = append(cells, patternCell{X: x, Occupied: x != hole})
		}
		rows = append(rows, patternRow{Y: y, Cells: cells})
	}
	g.queuePattern(rows)
}

func (g *Game) addCastle() {
	// Castle first clears the settled structure, then rebuilds the original
	// eleven-row grey castle silhouette from bottom to top.
	g.Board = [BoardHeight][BoardWidth]int{}
	g.Specials = [BoardHeight][BoardWidth]Special{}
	occupiedByY := map[int][]int{
		21: {2, 3, 4, 5, 6, 7},
		20: {2, 3, 4, 6, 7},
		19: {2, 3, 4, 6, 7},
		18: {2, 4, 5, 6, 7},
		17: {2, 4, 5, 6, 7},
		16: {2, 3, 4, 5, 7},
		15: {2, 3, 4, 5, 7},
		14: {2, 3, 5, 6, 7},
		13: {1, 2, 3, 4, 5, 6, 7, 8},
		12: {1, 2, 3, 4, 5, 6, 7, 8},
		11: {1, 2, 4, 5, 7, 8},
	}
	rows := make([]patternRow, 0, len(occupiedByY))
	for y := 21; y >= 11; y-- {
		occupied := make(map[int]bool)
		for _, x := range occupiedByY[y] {
			occupied[x] = true
		}
		cells := make([]patternCell, 0, BoardWidth)
		for x := 0; x < BoardWidth; x++ {
			cells = append(cells, patternCell{X: x, Occupied: occupied[x], Value: 8})
		}
		rows = append(rows, patternRow{Y: y, Cells: cells})
	}
	g.patternRows = nil
	g.queuePattern(rows)
}

func (g *Game) addRing() {
	// Hollow oval based on the original Ring coordinates, queued bottom-up.
	occupiedByY := map[int][]int{
		20: {3, 4, 5, 6},
		19: {1, 2, 7, 8},
		18: {1, 8},
		17: {0, 9},
		16: {0, 9},
		15: {0, 9},
		14: {0, 9},
		13: {1, 8},
		12: {1, 2, 7, 8},
		11: {3, 4, 5, 6},
	}
	rows := make([]patternRow, 0, len(occupiedByY))
	for y := 20; y >= 11; y-- {
		occupied := make(map[int]bool)
		for _, x := range occupiedByY[y] {
			occupied[x] = true
		}
		cells := make([]patternCell, 0, BoardWidth)
		for x := 0; x < BoardWidth; x++ {
			cells = append(cells, patternCell{X: x, Occupied: occupied[x], Value: 8})
		}
		rows = append(rows, patternRow{Y: y, Cells: cells})
	}
	g.queuePattern(rows)
}

func (g *Game) advancePattern() {
	if len(g.patternRows) == 0 {
		return
	}
	g.patternTick++
	if g.patternTick < 3 { // roughly one row per 50 ms at 60 Hz
		return
	}
	g.patternTick = 0
	row := g.patternRows[0]
	g.patternRows = g.patternRows[1:]
	active := make(map[Point]bool, 4)
	for _, point := range g.Cells(g.Active) {
		active[point] = true
	}
	for _, cell := range row.Cells {
		point := Point{X: cell.X, Y: row.Y}
		if active[point] {
			// Do not overwrite the falling piece. New surrounding blocks become
			// normal ground, so the piece settles through the usual lock delay.
			continue
		}
		g.Specials[row.Y][cell.X] = SpecialNone
		if cell.Occupied {
			value := cell.Value
			if value == 0 {
				value = 1 + g.random.IntN(len(shapes))
			}
			g.Board[row.Y][cell.X] = value
		} else {
			g.Board[row.Y][cell.X] = 0
		}
	}
}

func (g *Game) PendingPatternRows() int {
	return len(g.patternRows)
}

func (g *Game) startRumble() {
	occupied := make([]Point, 0)
	for y, row := range g.Board {
		for x, value := range row {
			if value != 0 {
				occupied = append(occupied, Point{X: x, Y: y})
			}
		}
	}
	g.random.Shuffle(len(occupied), func(i, j int) {
		occupied[i], occupied[j] = occupied[j], occupied[i]
	})
	if len(occupied) > 6 {
		occupied = occupied[:6]
	}
	g.rumblePoints = occupied
	g.RumbleRounds = 5
	g.rumbleTick = 0
}

func (g *Game) advanceRumble() {
	if g.RumbleRounds == 0 || len(g.rumblePoints) == 0 {
		return
	}
	g.rumbleTick++
	if g.rumbleTick < 3 {
		return
	}
	g.rumbleTick = 0
	for i, point := range g.rumblePoints {
		if g.Board[point.Y][point.X] == 0 {
			continue
		}
		dx := g.random.IntN(3) - 1
		dy := -g.random.IntN(2)
		destination := Point{X: point.X + dx, Y: point.Y + dy}
		if destination.X < 0 || destination.X >= BoardWidth || destination.Y < 1 || destination.Y >= BoardHeight || g.Board[destination.Y][destination.X] != 0 {
			continue
		}
		g.Board[destination.Y][destination.X] = g.Board[point.Y][point.X]
		g.Specials[destination.Y][destination.X] = g.Specials[point.Y][point.X]
		g.Board[point.Y][point.X] = 0
		g.Specials[point.Y][point.X] = SpecialNone
		g.rumblePoints[i] = destination
	}
	g.RumbleRounds--
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
	return g.Blind || g.Inverse || g.FasterStacks > 0 || g.SlowerBonus > 0 || g.PacketTicks > 0 || g.Mini || g.Blink || g.SZ || g.Trans || g.Blackout || g.RumbleRounds > 0 || g.BackgroundVariant > 0
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
	g.PacketTicks = 0
	g.Mini = false
	g.Blink = false
	g.BlinkVisible = true
	g.blinkTick = 0
	g.SZ = false
	g.Trans = false
	g.Blackout = false
	g.BackgroundVariant = 0
	g.RumbleRounds = 0
	g.rumbleTick = 0
	g.rumblePoints = nil
	g.ShowEvent("USED: Antidote")
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
