package controls

// KeyboardLayout is engine-independent metadata for lobby assignment and docs.
type KeyboardLayout struct {
	ID   int
	Name string
}

var KeyboardLayouts = []KeyboardLayout{
	{ID: 1, Name: "Keyboard 1 · A/D"},
	{ID: 2, Name: "Keyboard 2 · Arrows"},
	{ID: 3, Name: "Keyboard 3 · J/L"},
}
