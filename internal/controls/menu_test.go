package controls

import "testing"

func TestMenuFocusWraps(t *testing.T) {
	menu := MenuState{Count: 5}
	menu.Move(-1)
	if menu.Focus != 4 {
		t.Fatalf("backward focus = %d", menu.Focus)
	}
	menu.Move(2)
	if menu.Focus != 1 {
		t.Fatalf("forward focus = %d", menu.Focus)
	}
}
