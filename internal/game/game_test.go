package game

import "testing"

func TestPieceCannotMoveThroughWall(t *testing.T) {
	g := New(1)
	for g.Move(-1) {
	}
	for _, cell := range g.Cells(g.Active) {
		if cell.X < 0 {
			t.Fatalf("cell outside board: %+v", cell)
		}
	}
}

func TestHardDropLocksPieceAndSpawnsNext(t *testing.T) {
	g := New(1)
	first := g.Active
	g.HardDrop()
	occupied := 0
	for _, row := range g.Board {
		for _, cell := range row {
			if cell != 0 {
				occupied++
			}
		}
	}
	if occupied != 4 {
		t.Fatalf("occupied = %d", occupied)
	}
	if g.Active == first {
		t.Fatal("next piece was not spawned")
	}
}

func TestGroundedPieceWaitsForLockDelay(t *testing.T) {
	g := New(3)
	for g.StepDown() {
	}
	landed := g.Active
	for range LockDelayTicks - 1 {
		g.Tick()
	}
	if g.Active != landed {
		t.Fatal("piece locked before lock delay elapsed")
	}
	g.Tick()
	if g.Active == landed {
		t.Fatal("piece did not lock after lock delay")
	}
}

func TestHardDropBypassesLockDelay(t *testing.T) {
	g := New(4)
	g.HardDrop()
	occupied := 0
	for _, row := range g.Board {
		for _, value := range row {
			if value != 0 {
				occupied++
			}
		}
	}
	if occupied != 4 {
		t.Fatalf("occupied after hard drop = %d", occupied)
	}
}

func TestFasterAndSlowerChangeGravity(t *testing.T) {
	g := New(18)
	base := g.GravityTicks()
	g.ApplySpecial(SpecialFaster)
	if g.GravityTicks() != base*3/4 {
		t.Fatalf("faster gravity = %d, want %d", g.GravityTicks(), base*3/4)
	}
	g.Antidotes = 1
	if !g.UseAntidote() || g.FasterStacks != 0 || g.GravityTicks() != base {
		t.Fatal("antidote did not remove Faster")
	}
	g.activateSpecial(SpecialSlower)
	if g.GravityTicks() != base+1 {
		t.Fatalf("slower gravity = %d, want %d", g.GravityTicks(), base+1)
	}
	g.Antidotes = 1
	if !g.UseAntidote() || g.SlowerBonus != 0 || g.GravityTicks() != base {
		t.Fatal("antidote did not remove Slower")
	}
}

func TestFasterStacksLikeOriginal(t *testing.T) {
	g := New(19)
	g.ApplySpecial(SpecialFaster)
	g.ApplySpecial(SpecialFaster)
	if g.FasterStacks != 2 || g.GravityTicks() != 24 {
		t.Fatalf("stacks=%d gravity=%d", g.FasterStacks, g.GravityTicks())
	}
}

func TestGravitySpeedsUpWithLevel(t *testing.T) {
	g := New(5)
	initial := g.GravityTicks()
	if initial != 45 {
		t.Fatalf("level 0 gravity = %d, wanted 45 ticks", initial)
	}
	g.Lines = 5
	if g.GravityTicks() >= initial {
		t.Fatalf("level 1 gravity = %d, initial = %d", g.GravityTicks(), initial)
	}
}

func TestFourLineClearEarnsTwoGarbageRows(t *testing.T) {
	g := New(6)
	for y := BoardHeight - 4; y < BoardHeight; y++ {
		for x := 1; x < BoardWidth; x++ {
			g.Board[y][x] = 1
		}
	}
	g.Active = Piece{Kind: 0, Rotation: 1, X: -2, Y: BoardHeight - 4}
	g.lock()
	if got := g.ConsumeGarbage(); got != 2 {
		t.Fatalf("garbage = %d, want 2", got)
	}
	if got := g.ConsumeGarbage(); got != 0 {
		t.Fatalf("garbage was not consumed: %d", got)
	}
}

func TestGarbageRowHasOneHole(t *testing.T) {
	g := New(7)
	g.AddGarbage(1)
	occupied := 0
	for _, value := range g.Board[BoardHeight-1] {
		if value != 0 {
			occupied++
		}
	}
	if occupied != BoardWidth-1 {
		t.Fatalf("garbage occupied cells = %d", occupied)
	}
}

func TestGarbageAboveFullStackEndsGame(t *testing.T) {
	g := New(8)
	g.Board[0][0] = 1
	g.AddGarbage(1)
	if !g.GameOver {
		t.Fatal("full stack should be eliminated by garbage")
	}
}

func TestFullLineClearsAndScores(t *testing.T) {
	g := New(2)
	for x := 0; x < BoardWidth; x++ {
		g.Board[BoardHeight-1][x] = 1
	}
	if cleared, _ := g.clearLines(); cleared != 1 {
		t.Fatalf("cleared = %d", cleared)
	}
	for x := 0; x < BoardWidth; x++ {
		if g.Board[BoardHeight-1][x] != 0 {
			t.Fatal("bottom line was not cleared")
		}
	}
}

func TestSpecialSpawnIsSkippedOnEmptyBoard(t *testing.T) {
	g := New(33)
	g.spawnSpecial()
	if g.SpecialLifetimeTicks() != 0 {
		t.Fatalf("empty-board lifetime = %d", g.SpecialLifetimeTicks())
	}
	for _, row := range g.Specials {
		for _, special := range row {
			if special != SpecialNone {
				t.Fatal("special spawned on empty board")
			}
		}
	}
}

func TestSpecialLifetimeScalesWithOccupiedBlocks(t *testing.T) {
	tests := []struct {
		blocks  int
		seconds int
	}{{1, 18}, {10, 19}, {30, 21}, {80, 26}, {120, 30}, {180, 30}}
	for _, test := range tests {
		g := New(uint64(100 + test.blocks))
		for i := 0; i < test.blocks && i < BoardWidth*BoardHeight; i++ {
			g.Board[i/BoardWidth][i%BoardWidth] = 1
		}
		g.spawnSpecial()
		if got := g.SpecialLifetimeTicks(); got != test.seconds*60 {
			t.Errorf("%d blocks: lifetime = %d ticks, want %d", test.blocks, got, test.seconds*60)
		}
	}
}

func TestSpecialExpiresAtItsSpawnLifetime(t *testing.T) {
	g := New(34)
	for x := 0; x < 10; x++ {
		g.Board[BoardHeight-1][x] = 1
	}
	g.spawnSpecial()
	lifetime := g.SpecialLifetimeTicks()
	g.specialTick = lifetime - 1
	g.Tick()
	if g.SpecialLifetimeTicks() != 0 {
		t.Fatal("special did not expire at fixed spawn lifetime")
	}
}

func TestAntidoteSpecialActivatesWhenItsRowClears(t *testing.T) {
	g := New(10)
	for x := 0; x < BoardWidth; x++ {
		g.Board[BoardHeight-1][x] = 1
	}
	if !g.SpawnSpecial(SpecialAntidote, Point{X: 3, Y: BoardHeight - 1}) {
		t.Fatal("could not place special")
	}
	_, activated := g.clearLines()
	for _, special := range activated {
		g.activateSpecial(special)
	}
	if g.Antidotes != 1 {
		t.Fatalf("antidotes = %d", g.Antidotes)
	}
}

func TestBlindSpecialQueuesTargetEffect(t *testing.T) {
	g := New(13)
	g.activateSpecial(SpecialBlind)
	specials := g.ConsumeSpecials()
	if len(specials) != 1 || specials[0] != SpecialBlind {
		t.Fatalf("queued specials = %v", specials)
	}
	if len(g.ConsumeSpecials()) != 0 {
		t.Fatal("special was not consumed")
	}
}

func TestInverseReversesPlayerMovementAndRotation(t *testing.T) {
	g := New(16)
	g.ApplySpecial(SpecialInverse)
	startX := g.Active.X
	if !g.MoveInput(-1) || g.Active.X != startX+1 {
		t.Fatalf("inverse left moved x from %d to %d", startX, g.Active.X)
	}
	startRotation := g.Active.Rotation
	if !g.RotateInput(1) || g.Active.Rotation != (startRotation+3)%4 {
		t.Fatalf("inverse CW rotation = %d", g.Active.Rotation)
	}
}

func TestAntidoteClearsAllNegativeEffects(t *testing.T) {
	g := New(17)
	g.ApplySpecial(SpecialBlind)
	g.ApplySpecial(SpecialInverse)
	g.Antidotes = 1
	if !g.UseAntidote() || g.Blind || g.Inverse || g.Antidotes != 0 {
		t.Fatalf("blind=%v inverse=%v antidotes=%d", g.Blind, g.Inverse, g.Antidotes)
	}
}

func TestAntidoteClearsBlindEffect(t *testing.T) {
	g := New(14)
	g.ApplySpecial(SpecialBlind)
	g.Antidotes = 1
	if !g.UseAntidote() || g.Blind || g.Antidotes != 0 {
		t.Fatalf("blind=%v antidotes=%d", g.Blind, g.Antidotes)
	}
}

func TestAntidoteClearsMixedPositiveAndNegativeEffects(t *testing.T) {
	g := New(20)
	g.ApplySpecial(SpecialBlind)
	g.ApplySpecial(SpecialFaster)
	g.activateSpecial(SpecialSlower)
	g.Antidotes = 1
	if !g.UseAntidote() || g.Blind || g.FasterStacks != 0 || g.SlowerBonus != 0 {
		t.Fatalf("blind=%v faster=%d slower=%d", g.Blind, g.FasterStacks, g.SlowerBonus)
	}
}

func TestAntidoteIsNotSpentWithoutEffect(t *testing.T) {
	g := New(15)
	g.Antidotes = 1
	if g.UseAntidote() || g.Antidotes != 1 {
		t.Fatal("antidote should only be spent on an active negative effect")
	}
}

func TestBridgeAddsTwoRows(t *testing.T) {
	g := New(21)
	g.ApplySpecial(SpecialBridge)
	occupied := 0
	for _, row := range g.Board {
		for _, value := range row {
			if value != 0 {
				occupied++
			}
		}
	}
	if occupied != 2*(BoardWidth-1) {
		t.Fatalf("bridge occupied cells = %d", occupied)
	}
}

func TestQuestionRemovesHalfOfPlacedBlocks(t *testing.T) {
	g := New(22)
	for y := BoardHeight - 2; y < BoardHeight; y++ {
		for x := 0; x < BoardWidth; x++ {
			g.Board[y][x] = 1
		}
	}
	g.SpawnSpecial(SpecialAntidote, Point{X: 0, Y: BoardHeight - 1})
	g.ApplySpecial(SpecialQuestion)
	occupied := 0
	for _, row := range g.Board {
		for _, value := range row {
			if value != 0 {
				occupied++
			}
		}
	}
	if occupied != 10 {
		t.Fatalf("question left %d blocks, want 10", occupied)
	}
	if g.Specials[BoardHeight-1][0] != SpecialNone && g.Board[BoardHeight-1][0] == 0 {
		t.Fatal("question left an orphaned special")
	}
}

func advanceAllPatternRows(g *Game) {
	for g.PendingPatternRows() > 0 {
		for range 3 {
			g.advancePattern()
		}
	}
}

func TestStairBuildsFromBottomUp(t *testing.T) {
	g := New(23)
	g.ApplySpecial(SpecialStair)
	if g.PendingPatternRows() != 10 {
		t.Fatalf("queued stair rows = %d", g.PendingPatternRows())
	}
	for range 3 {
		g.advancePattern()
	}
	if g.Board[21][0] == 0 || g.Board[20][1] != 0 {
		t.Fatal("stair did not start at bottom")
	}
	advanceAllPatternRows(g)
	for x := 0; x <= 9; x++ {
		y := 21 - x
		if g.Board[y][x] == 0 {
			t.Fatalf("stair missing block at (%d,%d)", x, y)
		}
	}
}

func TestFillAddsTenRowsBottomUpWithOneHoleEach(t *testing.T) {
	g := New(24)
	g.ApplySpecial(SpecialFill)
	for range 3 {
		g.advancePattern()
	}
	if g.PendingPatternRows() != 9 {
		t.Fatalf("pending rows after first fill row = %d", g.PendingPatternRows())
	}
	advanceAllPatternRows(g)
	for y := 12; y <= 21; y++ {
		occupied := 0
		for _, value := range g.Board[y] {
			if value != 0 {
				occupied++
			}
		}
		if occupied != BoardWidth-1 {
			t.Fatalf("row %d has %d blocks, want %d", y, occupied, BoardWidth-1)
		}
	}
}

func TestFillDoesNotOverwriteActivePiece(t *testing.T) {
	g := New(25)
	g.Active = Piece{Kind: 1, X: 3, Y: 20}
	active := g.Cells(g.Active)
	g.ApplySpecial(SpecialFill)
	for range 3 {
		g.advancePattern()
	}
	if g.GameOver {
		t.Fatal("fill overlap should not immediately eliminate player")
	}
	for _, point := range active {
		if point.Y == 21 && g.Board[point.Y][point.X] != 0 {
			t.Fatalf("fill overwrote active piece at %+v", point)
		}
	}
	if !g.grounded() {
		t.Fatal("new fill row should act as ground for active piece")
	}
}

func TestFlipReversesOccupiedVerticalExtentAndSpecials(t *testing.T) {
	g := New(26)
	g.Board[18][1] = 1
	g.Board[21][7] = 2
	g.Specials[18][1] = SpecialAntidote
	g.FlipBoard()
	if g.Board[21][1] != 1 || g.Board[18][7] != 2 {
		t.Fatalf("flip positions incorrect: top=%d bottom=%d", g.Board[18][7], g.Board[21][1])
	}
	if g.Specials[21][1] != SpecialAntidote {
		t.Fatal("special marker did not flip with settled block")
	}
}

func TestFlipOfEmptyBoardIsNoOp(t *testing.T) {
	g := New(27)
	active := g.Active
	g.FlipBoard()
	if g.Active != active {
		t.Fatal("flip moved active piece")
	}
}

func TestSwapExchangesOnlySettledBoards(t *testing.T) {
	a, b := New(28), New(29)
	a.Board[21][0] = 1
	a.Specials[21][0] = SpecialClear
	b.Board[20][9] = 2
	aActive, bActive := a.Active, b.Active
	aNext, bNext := a.NextKind, b.NextKind
	a.SwapBoard(b)
	if a.Board[20][9] != 2 || b.Board[21][0] != 1 || b.Specials[21][0] != SpecialClear {
		t.Fatal("settled boards were not swapped")
	}
	if a.Active != aActive || b.Active != bActive || a.NextKind != aNext || b.NextKind != bNext {
		t.Fatal("swap changed active or next pieces")
	}
}

func TestPacketSendsOneGarbageRowPerClearedLine(t *testing.T) {
	g := New(30)
	g.PacketTicks = 20 * 60
	for x := 0; x < BoardWidth; x++ {
		g.Board[BoardHeight-1][x] = 1
	}
	g.Active = Piece{Kind: 1, X: 3, Y: 0}
	g.lock()
	if got := g.ConsumeGarbage(); got != 1 {
		t.Fatalf("packet garbage = %d, want 1", got)
	}
}

func TestPacketExpiresAndAntidoteClearsIt(t *testing.T) {
	g := New(31)
	g.PacketTicks = 2
	g.Tick()
	g.Tick()
	if g.PacketTicks != 0 {
		t.Fatalf("packet ticks = %d", g.PacketTicks)
	}
	g.PacketTicks = 100
	g.Antidotes = 1
	if !g.UseAntidote() || g.PacketTicks != 0 {
		t.Fatal("antidote did not clear Packet")
	}
}

func TestRingBuildsHollowPatternBottomUp(t *testing.T) {
	g := New(32)
	g.ApplySpecial(SpecialRing)
	if g.PendingPatternRows() != 10 {
		t.Fatalf("ring rows = %d", g.PendingPatternRows())
	}
	advanceAllPatternRows(g)
	if g.Board[20][3] == 0 || g.Board[20][6] == 0 || g.Board[20][4] == 0 {
		t.Fatal("ring bottom missing")
	}
	if g.Board[17][0] == 0 || g.Board[17][9] == 0 || g.Board[17][5] != 0 {
		t.Fatal("ring side or hollow center incorrect")
	}
	if g.Board[11][3] == 0 || g.Board[11][6] == 0 {
		t.Fatal("ring top missing")
	}
}

func TestBlinkAlternatesActivePieceVisibility(t *testing.T) {
	g := New(35)
	g.ApplySpecial(SpecialBlink)
	if !g.Blink || !g.BlinkVisible {
		t.Fatal("blink did not start visible")
	}
	for range 6 {
		g.Tick()
	}
	if g.BlinkVisible {
		t.Fatal("blink did not hide after six ticks")
	}
	for range 6 {
		g.Tick()
	}
	if !g.BlinkVisible {
		t.Fatal("blink did not become visible again")
	}
}

func TestMiniAndBlinkAreClearedByAntidote(t *testing.T) {
	g := New(36)
	g.ApplySpecial(SpecialMini)
	g.ApplySpecial(SpecialBlink)
	g.Antidotes = 1
	if !g.UseAntidote() || g.Mini || g.Blink {
		t.Fatalf("mini=%v blink=%v", g.Mini, g.Blink)
	}
}

func TestSZRestrictsFuturePiecesToSAndZ(t *testing.T) {
	g := New(37)
	g.ApplySpecial(SpecialSZ)
	for range 30 {
		if g.NextKind != 5 && g.NextKind != 6 {
			t.Fatalf("SZ generated kind %d", g.NextKind)
		}
		g.Board = [BoardHeight][BoardWidth]int{}
		g.HardDrop()
		if g.Active.Kind != 5 && g.Active.Kind != 6 {
			t.Fatalf("SZ spawned kind %d", g.Active.Kind)
		}
	}
}

func TestTransAndSZAreClearedByAntidote(t *testing.T) {
	g := New(38)
	g.ApplySpecial(SpecialSZ)
	g.ApplySpecial(SpecialTrans)
	g.Antidotes = 1
	if !g.UseAntidote() || g.SZ || g.Trans {
		t.Fatalf("sz=%v trans=%v", g.SZ, g.Trans)
	}
}

func TestClearSpecialEmptiesBoard(t *testing.T) {
	g := New(11)
	for x := 0; x < BoardWidth; x++ {
		g.Board[BoardHeight-1][x] = 1
	}
	g.Board[BoardHeight-2][0] = 2
	g.SpawnSpecial(SpecialClear, Point{X: 3, Y: BoardHeight - 1})
	_, activated := g.clearLines()
	for _, special := range activated {
		g.activateSpecial(special)
	}
	for _, row := range g.Board {
		for _, value := range row {
			if value != 0 {
				t.Fatal("clear special left occupied cell")
			}
		}
	}
}

func TestSpecialMovesDownWithRowsAboveClearedLine(t *testing.T) {
	g := New(12)
	for x := 0; x < BoardWidth; x++ {
		g.Board[BoardHeight-1][x] = 1
	}
	g.Board[BoardHeight-2][0] = 2
	g.SpawnSpecial(SpecialAntidote, Point{X: 0, Y: BoardHeight - 2})
	g.clearLines()
	if g.Specials[BoardHeight-1][0] != SpecialAntidote {
		t.Fatal("special did not fall with its block")
	}
}

func TestNextPieceIsKnownBeforeSpawn(t *testing.T) {
	g := New(9)
	next := g.NextKind
	g.HardDrop()
	if g.Active.Kind != next {
		t.Fatalf("active kind = %d, wanted previewed %d", g.Active.Kind, next)
	}
}

func TestSameSeedProducesSamePieces(t *testing.T) {
	a, b := New(42), New(42)
	for range 20 {
		if a.Active.Kind != b.Active.Kind {
			t.Fatal("seeded sequence differs")
		}
		a.HardDrop()
		b.HardDrop()
		if a.GameOver || b.GameOver {
			break
		}
	}
}
