package controls

import "testing"

func TestLobbyNavigationFollowsGeometry(t *testing.T) {
	if got := NavigateLobby(1, MenuDown, true); got != 0 {
		t.Fatalf("Join down = %d, want Start", got)
	}
	if got := NavigateLobby(0, MenuLeft, true); got != 2 {
		t.Fatalf("Start left = %d, want Sound", got)
	}
	if got := NavigateLobby(0, MenuRight, true); got != 3 {
		t.Fatalf("Start right = %d, want Music", got)
	}
	if got := NavigateLobby(4, MenuRight, true); got != 2 {
		t.Fatalf("Controller right = %d, want Sound", got)
	}
}

func TestWebNavigationNeverSelectsExit(t *testing.T) {
	for focus := 0; focus < 6; focus++ {
		for direction := MenuLeft; direction <= MenuDown; direction++ {
			if got := NavigateLobby(focus, direction, false); got == 6 {
				t.Fatalf("focus %d direction %d selected web Exit", focus, direction)
			}
		}
	}
}
