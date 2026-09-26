package controls

type MenuDirection int

const (
	MenuLeft MenuDirection = iota
	MenuRight
	MenuUp
	MenuDown
)

// NavigateLobby follows the visible geometry. Indices: Start, Settings, and
// optional native Exit. Start sits above the utility row.
func NavigateLobby(focus int, direction MenuDirection, nativeExit bool) int {
	row := []int{1}
	if nativeExit {
		row = append(row, 2)
	}
	if focus == 0 {
		return 1
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
