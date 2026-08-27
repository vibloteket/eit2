package match

import (
	"testing"

	"github.com/vibloteket/eit2/internal/game"
)

func TestSeededMatchesAreReproducible(t *testing.T) {
	a := NewSeeded(2, 12345)
	b := NewSeeded(2, 12345)
	for i := range a.Players {
		if a.Players[i].Active.Kind != b.Players[i].Active.Kind || a.Players[i].NextKind != b.Players[i].NextKind {
			t.Fatalf("player %d seeded state differs", i)
		}
	}
}

func TestPlayersUseDistinctRandomStreams(t *testing.T) {
	m := NewSeeded(4, 99)
	same := true
	for i := 1; i < len(m.Players); i++ {
		if m.Players[i].Active.Kind != m.Players[0].Active.Kind || m.Players[i].NextKind != m.Players[0].NextKind {
			same = false
		}
	}
	if same {
		t.Fatal("all players started with the same random stream")
	}
}

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

func TestFlipAndSwitchRouteToTarget(t *testing.T) {
	m := New(2)
	m.Players[1].Board[18][2] = 3
	m.Players[0].QueueSpecial(game.SpecialFlip)
	m.routeSpecials()
	if m.Players[1].Board[21][2] != 3 {
		t.Fatal("selected target board was not flipped")
	}

	m = New(2)
	m.Players[0].Board[21][0] = 1
	m.Players[1].Board[20][9] = 2
	m.Players[0].QueueSpecial(game.SpecialSwitch)
	m.routeSpecials()
	if m.Players[0].Board[20][9] != 2 || m.Players[1].Board[21][0] != 1 {
		t.Fatal("attacker and selected target did not swap boards")
	}
}

func TestSoloSwitchIsNoOp(t *testing.T) {
	m := New(1)
	m.Players[0].Board[21][0] = 1
	m.Players[0].QueueSpecial(game.SpecialSwitch)
	m.routeSpecials()
	if m.Players[0].Board[21][0] != 1 {
		t.Fatal("solo switch changed its own board")
	}
}

func TestRingRoutesToTarget(t *testing.T) {
	m := New(2)
	m.Players[0].QueueSpecial(game.SpecialRing)
	m.routeSpecials()
	if m.Players[1].PendingPatternRows() != 10 {
		t.Fatal("selected target did not queue Ring")
	}
}

func TestMiniAndBlinkRouteToTarget(t *testing.T) {
	m := New(2)
	m.Players[0].QueueSpecial(game.SpecialMini)
	m.Players[0].QueueSpecial(game.SpecialBlink)
	m.routeSpecials()
	if !m.Players[1].Mini || !m.Players[1].Blink {
		t.Fatal("selected target did not receive Mini and Blink")
	}
	if m.Players[0].Mini || m.Players[0].Blink {
		t.Fatal("attacker received its own Mini or Blink")
	}
}

func TestSZAndTransRouteToTarget(t *testing.T) {
	m := New(2)
	m.Players[0].QueueSpecial(game.SpecialSZ)
	m.Players[0].QueueSpecial(game.SpecialTrans)
	m.routeSpecials()
	if !m.Players[1].SZ || !m.Players[1].Trans {
		t.Fatal("selected target did not receive SZ and Trans")
	}
	if m.Players[0].SZ || m.Players[0].Trans {
		t.Fatal("attacker received its own SZ or Trans")
	}
}

func TestCastleAndBlackoutRouteToTarget(t *testing.T) {
	m := New(2)
	m.Players[1].Board[5][0] = 1
	m.Players[0].QueueSpecial(game.SpecialCastle)
	m.Players[0].QueueSpecial(game.SpecialColor)
	m.routeSpecials()
	if m.Players[1].PendingPatternRows() != 11 || !m.Players[1].Blackout {
		t.Fatal("selected target did not receive Castle and Blackout")
	}
	if m.Players[1].Board[5][0] != 0 {
		t.Fatal("castle did not clear selected target")
	}
}

func TestRumbleAndBackgroundRouteToTarget(t *testing.T) {
	m := New(2)
	m.Players[1].Board[18][4] = 1
	m.Players[0].QueueSpecial(game.SpecialRumble)
	m.Players[0].QueueSpecial(game.SpecialBackground)
	m.routeSpecials()
	if m.Players[1].RumbleRounds != 5 || m.Players[1].BackgroundVariant == 0 {
		t.Fatal("selected target did not receive Rumble and Background")
	}
	if m.Players[0].RumbleRounds != 0 || m.Players[0].BackgroundVariant != 0 {
		t.Fatal("attacker received its own Rumble or Background")
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

func TestSpecialFeedbackNamesSenderAndTarget(t *testing.T) {
	m := New(2)
	m.Players[0].QueueSpecial(game.SpecialBlind)
	m.routeSpecials()
	if m.Players[0].LastEvent != "SENT Blind TO P2" {
		t.Fatalf("sender event = %q", m.Players[0].LastEvent)
	}
	if m.Players[1].LastEvent != "P1: Blind" {
		t.Fatalf("target event = %q", m.Players[1].LastEvent)
	}
}

func TestSoloSpecialFeedbackSaysSelf(t *testing.T) {
	m := New(1)
	m.Players[0].QueueSpecial(game.SpecialInverse)
	m.routeSpecials()
	if m.Players[0].LastEvent != "SELF: Inverse" {
		t.Fatalf("solo event = %q", m.Players[0].LastEvent)
	}
}

func TestDebugCollectUsesNormalRouting(t *testing.T) {
	m := New(2)
	if !m.DebugCollect(0, game.SpecialBlind) {
		t.Fatal("debug collect failed")
	}
	if !m.Players[1].Blind || m.Players[0].Blind {
		t.Fatal("debug special did not follow target routing")
	}
	if m.DebugCollect(-1, game.SpecialBlind) || m.DebugCollect(2, game.SpecialBlind) || m.DebugCollect(0, game.SpecialNone) {
		t.Fatal("debug collect accepted invalid input")
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
