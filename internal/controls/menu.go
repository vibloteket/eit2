package controls

type MenuDirection int

const (
	MenuLeft MenuDirection = iota
	MenuRight
	MenuUp
	MenuDown
)

// NavigateLobby follows the visual layout. Indices: Start, Sound, Music,
// Controller debug, Debug mode, and optional native Exit.
func NavigateLobby(focus int, direction MenuDirection, nativeExit bool) int {
	left := map[int]int{0: 1, 1: 3, 2: 0, 3: 4, 4: 2, 5: 4}
	right := map[int]int{0: 2, 1: 0, 2: 4, 3: 1, 4: 3, 5: 4}
	up := map[int]int{0: 2, 1: 0, 2: 0, 3: 0, 4: 5, 5: 4}
	down := map[int]int{0: 2, 1: 3, 2: 4, 3: 1, 4: 5, 5: 4}
	if !nativeExit {
		up[4], down[4] = 0, 2
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
	if !nativeExit && next == 5 {
		return 4
	}
	return next
}
