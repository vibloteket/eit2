package controls

type MenuDirection int

const (
	MenuLeft MenuDirection = iota
	MenuRight
	MenuUp
	MenuDown
)

// NavigateLobby follows the visible geometry. Indices: Start, Sound, Music,
// Controller debug, Debug mode, and optional native Exit. The utility buttons
// form one left-to-right row; Start sits above its centre.
func NavigateLobby(focus int, direction MenuDirection, nativeExit bool) int {
	row := []int{3, 1, 2, 4}
	if nativeExit {
		row = append(row, 5)
	}
	if focus == 0 {
		switch direction {
		case MenuLeft:
			return 2
		case MenuRight, MenuUp, MenuDown:
			return 4
		}
	}
	position := -1
	for i, item := range row {
		if item == focus {
			position = i
			break
		}
	}
	if position < 0 {
		return 0
	}
	switch direction {
	case MenuLeft:
		return row[(position-1+len(row))%len(row)]
	case MenuRight:
		return row[(position+1)%len(row)]
	case MenuUp:
		return 0
	case MenuDown:
		return focus
	}
	return focus
}
