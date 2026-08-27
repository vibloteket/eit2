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
