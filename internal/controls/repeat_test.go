package controls

import "testing"

func TestTouchRepeatTiming(t *testing.T) {
	if !ShouldRepeat(Left, 0) || !ShouldRepeat(Down, 0) {
		t.Fatal("new touch must act immediately")
	}
	for tick := 1; tick < 12; tick++ {
		if ShouldRepeat(Left, tick) {
			t.Fatalf("horizontal movement repeated too early at %d", tick)
		}
	}
	if !ShouldRepeat(Left, 12) || !ShouldRepeat(Left, 16) {
		t.Fatal("horizontal movement did not repeat after delay")
	}
	if !ShouldRepeat(Down, 2) || ShouldRepeat(Down, 3) {
		t.Fatal("soft drop should repeat every other tick")
	}
	if ShouldRepeat(RotateCW, 10) || ShouldRepeat(HardDrop, 10) {
		t.Fatal("rotation and hard drop must not repeat while held")
	}
}
