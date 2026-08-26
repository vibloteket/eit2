// Package controls contains engine-independent input behavior.
package controls

type Action int

const (
	Left Action = iota
	Right
	Down
	RotateCCW
	RotateCW
	HardDrop
)

// ShouldRepeat reports whether a held action should fire on this tick.
func ShouldRepeat(action Action, heldTicks int) bool {
	if heldTicks == 0 {
		return true
	}
	switch action {
	case Left, Right:
		return heldTicks >= 12 && (heldTicks-12)%4 == 0
	case Down:
		return heldTicks%2 == 0
	default:
		return false
	}
}
