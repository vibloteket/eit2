package controls

// MenuState implements wrapping focus movement independently of Ebitengine.
type MenuState struct {
	Focus int
	Count int
}

func (m *MenuState) Move(delta int) {
	if m.Count <= 0 {
		m.Focus = 0
		return
	}
	m.Focus = (m.Focus + delta) % m.Count
	if m.Focus < 0 {
		m.Focus += m.Count
	}
}
