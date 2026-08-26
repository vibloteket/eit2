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
	if cleared := g.clearLines(); cleared != 1 {
		t.Fatalf("cleared = %d", cleared)
	}
	for x := 0; x < BoardWidth; x++ {
		if g.Board[BoardHeight-1][x] != 0 {
			t.Fatal("bottom line was not cleared")
		}
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
