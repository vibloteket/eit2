package controls

type MenuDirection int

const (
	MenuLeft MenuDirection = iota
	MenuRight
	MenuUp
	MenuDown
)

// NavigateLobby follows the lobby's visual two-row layout rather than a linear
// tab order. Indices: Start, Join keyboard, Sound, Music, Controller debug,
// Debug mode, and optional native Exit.
func NavigateLobby(focus int, direction MenuDirection, nativeExit bool) int {
	left := map[int]int{0: 2, 1: 4, 2: 4, 3: 0, 4: 5, 5: 3, 6: 1}
	right := map[int]int{0: 3, 1: 6, 2: 0, 3: 5, 4: 2, 5: 4, 6: 1}
	up := map[int]int{0: 1, 1: 0, 2: 1, 3: 1, 4: 1, 5: 6, 6: 5}
	down := map[int]int{0: 1, 1: 0, 2: 4, 3: 5, 4: 2, 5: 3, 6: 5}
	if !nativeExit {
		left[1], right[1], up[5], down[5] = 4, 5, 1, 3
	}
	var next int
	switch direction {
	case MenuLeft:
		next = left[focus]
	case MenuRight:
		next = right[focus]
	case MenuUp:
		next = up[focus]
	case MenuDown:
		next = down[focus]
	}
	if !nativeExit && next == 6 {
		return 5
	}
	return next
}
