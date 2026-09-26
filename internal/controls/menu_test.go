package controls

import "testing"

func TestLobbyNavigationFollowsGeometry(t *testing.T) {
	if got := NavigateLobby(0, MenuDown, true); got != 1 {
		t.Fatalf("Start down = %d, want Settings", got)
	}
	if got := NavigateLobby(1, MenuUp, true); got != 0 {
		t.Fatalf("Settings up = %d, want Start", got)
	}
	if got := NavigateLobby(1, MenuRight, true); got != 2 {
		t.Fatalf("Settings right = %d, want Exit", got)
	}
	if got := NavigateLobby(2, MenuRight, true); got != 1 {
		t.Fatalf("Exit right = %d, want Settings", got)
	}
	if got := NavigateLobby(1, MenuRight, false); got != 1 {
		t.Fatalf("web Settings right = %d, want Settings", got)
	}
	if got := NavigateLobby(1, MenuLeft, false); got != 1 {
		t.Fatalf("web Settings left = %d, want Settings", got)
	}
	if got := NavigateLobby(99, MenuDown, false); got != 0 {
		t.Fatalf("invalid focus = %d, want Start", got)
	}
}

func TestWebNavigationNeverSelectsExit(t *testing.T) {
	for focus := 0; focus < 3; focus++ {
		for direction := MenuLeft; direction <= MenuDown; direction++ {
			if got := NavigateLobby(focus, direction, false); got == 2 {
				t.Fatalf("focus %d direction %d selected web Exit", focus, direction)
			}
		}
	}
}
