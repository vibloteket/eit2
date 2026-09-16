package controls

import "testing"

func TestLobbyNavigationFollowsGeometry(t *testing.T) {
	if got := NavigateLobby(3, MenuRight, true); got != 1 {
		t.Fatalf("Controller right = %d, want Sound", got)
	}
	if got := NavigateLobby(1, MenuRight, true); got != 2 {
		t.Fatalf("Sound right = %d, want Music", got)
	}
	if got := NavigateLobby(2, MenuRight, true); got != 4 {
		t.Fatalf("Music right = %d, want Debug", got)
	}
	if got := NavigateLobby(4, MenuRight, true); got != 5 {
		t.Fatalf("Debug right = %d, want Credits", got)
	}
	if got := NavigateLobby(5, MenuRight, true); got != 6 {
		t.Fatalf("Credits right = %d, want Exit", got)
	}
	if got := NavigateLobby(6, MenuRight, true); got != 3 {
		t.Fatalf("Exit right = %d, want Controller debug", got)
	}
	if got := NavigateLobby(5, MenuRight, false); got != 3 {
		t.Fatalf("Web Credits right = %d, want Controller debug", got)
	}
	if got := NavigateLobby(3, MenuLeft, false); got != 5 {
		t.Fatalf("Web Controller left = %d, want Credits", got)
	}
	if got := NavigateLobby(2, MenuUp, true); got != 0 {
		t.Fatalf("Music up = %d, want Start", got)
	}
	if got := NavigateLobby(0, MenuDown, true); got != 4 {
		t.Fatalf("Start down = %d, want Debug", got)
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
