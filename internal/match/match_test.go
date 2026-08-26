package match

import "testing"

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

func TestLastAlivePlayerWins(t *testing.T) {
	m := New(3)
	m.Players[0].GameOver = true
	m.Players[2].GameOver = true
	m.UpdateStatus()
	if !m.Over || m.Winner != 1 {
		t.Fatalf("over=%v winner=%d", m.Over, m.Winner)
	}
}

func TestSoloDoesNotDeclareMultiplayerWinner(t *testing.T) {
	m := New(1)
	m.Players[0].GameOver = true
	m.UpdateStatus()
	if m.Over || m.Winner != -1 {
		t.Fatalf("solo match over=%v winner=%d", m.Over, m.Winner)
	}
}
