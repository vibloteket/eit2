package controls

import "testing"

func TestAxisDirectionUsesHysteresis(t *testing.T) {
	if got := AxisDirection(.5, 0); got != 0 {
		t.Fatalf("early direction = %d", got)
	}
	if got := AxisDirection(.8, 0); got != 1 {
		t.Fatalf("positive direction = %d", got)
	}
	if got := AxisDirection(.7, 1); got != 1 {
		t.Fatalf("held direction = %d", got)
	}
	if got := AxisDirection(.1, 1); got != 0 {
		t.Fatalf("recentering = %d", got)
	}
	if got := AxisDirection(-.8, 0); got != -1 {
		t.Fatalf("negative direction = %d", got)
	}
}
