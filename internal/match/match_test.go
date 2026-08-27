package match

import (
	"testing"

	"github.com/vibloteket/eit2/internal/game"
)

func TestInitialTargetsPointToNextPlayer(t *testing.T) {
	m := New(4)
	want := []int{1, 2, 3, 0}
	for player, target := range want {
		if m.Target(player) != target {
			t.Fatalf("P%d target = %d, want %d", player+1, m.Target(player)+1, target+1)
		}
	}
}

func TestCycleTargetSkipsSelfAndWraps(t *testing.T) {
	m := New(3)
	if got := m.CycleTarget(0); got != 2 {
		t.Fatalf("first cycle = %d, want 2", got)
	}
	if got := m.CycleTarget(0); got != 1 {
		t.Fatalf("wrapped cycle = %d, want 1", got)
	}
}

func TestEliminatedTargetIsReassigned(t *testing.T) {
	m := New(4)
	m.Players[1].GameOver = true
	m.UpdateStatus()
	if got := m.Target(0); got != 2 {
		t.Fatalf("P1 target = %d, want P3", got+1)
	}
	if m.Target(1) != -1 {
		t.Fatal("eliminated player should have no target")
	}
}

func TestGarbageRoutesToSelectedTarget(t *testing.T) {
	m := New(3)
	m.Players[0].QueueAttack(2)
	m.routeGarbage()
	occupied := 0
	for _, row := range m.Players[1].Board {
		for _, value := range row {
			if value != 0 {
				occupied++
			}
		}
	}
	if occupied != 18 {
		t.Fatalf("target garbage cells = %d, want 18", occupied)
	}
	for _, row := range m.Players[2].Board {
		for _, value := range row {
			if value != 0 {
				t.Fatal("non-target received garbage")
			}
		}
	}
}

func TestSpecialRoutesToSelectedTarget(t *testing.T) {
	m := New(3)
	m.Players[0].QueueSpecial(game.SpecialBlind)
	m.routeSpecials()
	if !m.Players[1].Blind {
		t.Fatal("selected target did not receive blind")
	}
	if m.Players[2].Blind {
		t.Fatal("non-target received blind")
	}
}

func TestBridgeAndQuestionRouteToTarget(t *testing.T) {
	m := New(2)
	for y := game.BoardHeight - 2; y < game.BoardHeight; y++ {
		for x := 0; x < game.BoardWidth; x++ {
			m.Players[1].Board[y][x] = 1
		}
	}
	m.Players[0].QueueSpecial(game.SpecialQuestion)
	m.routeSpecials()
	occupied := 0
	for _, row := range m.Players[1].Board {
		for _, value := range row {
			if value != 0 {
				occupied++
			}
		}
	}
	if occupied != 10 {
		t.Fatalf("question left %d blocks on target", occupied)
	}
	m.Players[0].QueueSpecial(game.SpecialBridge)
	m.routeSpecials()
	occupied = 0
	for _, row := range m.Players[1].Board {
		for _, value := range row {
			if value != 0 {
				occupied++
			}
		}
	}
	if occupied != 28 {
		t.Fatalf("bridge left %d blocks on target, want 28", occupied)
	}
}

func TestStairAndFillRouteToTarget(t *testing.T) {
	m := New(2)
	m.Players[0].QueueSpecial(game.SpecialStair)
	m.routeSpecials()
	if m.Players[1].PendingPatternRows() != 10 {
		t.Fatal("selected target did not queue Stair")
	}
	m = New(2)
	m.Players[0].QueueSpecial(game.SpecialFill)
	m.routeSpecials()
	if m.Players[1].PendingPatternRows() != 10 {
		t.Fatal("selected target did not queue Fill")
	}
}

func TestFasterRoutesToTarget(t *testing.T) {
	m := New(2)
	m.Players[0].QueueSpecial(game.SpecialFaster)
	m.routeSpecials()
	if m.Players[1].FasterStacks != 1 {
		t.Fatal("selected target did not receive Faster")
	}
	if m.Players[0].FasterStacks != 0 {
		t.Fatal("attacker received its own Faster")
	}
}

func TestLastAlivePlayerWins(t *testing.T) {
	m := New(3)
	m.Players[0].GameOver = true
	m.Players[2].GameOver = true
	m.UpdateStatus()
	if !m.Over || m.Winner != 1 {
		t.Fatalf("over=%v winner=%d", m.Over, m.Winner)
	}
}

func TestSoloTargetsSelfForSpecials(t *testing.T) {
	m := New(1)
	if m.Target(0) != 0 {
		t.Fatalf("solo target = %d, want self", m.Target(0))
	}
	m.Players[0].QueueSpecial(game.SpecialBlind)
	m.routeSpecials()
	if !m.Players[0].Blind {
		t.Fatal("solo target special did not hit self")
	}
}

func TestSoloGarbageAttackHasNoRecipient(t *testing.T) {
	m := New(1)
	m.Players[0].QueueAttack(2)
	m.routeGarbage()
	for _, row := range m.Players[0].Board {
		for _, value := range row {
			if value != 0 {
				t.Fatal("solo four-line garbage attacked self")
			}
		}
	}
}

func TestSoloDoesNotDeclareMultiplayerWinner(t *testing.T) {
	m := New(1)
	m.Players[0].GameOver = true
	m.UpdateStatus()
	if m.Over || m.Winner != -1 {
		t.Fatalf("solo match over=%v winner=%d", m.Over, m.Winner)
	}
	if m.Target(0) != -1 {
		t.Fatal("eliminated solo player should have no target")
	}
}
