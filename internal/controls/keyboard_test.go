package controls

import "testing"

func TestKeyboardLayoutsAreThreeUniqueDevices(t *testing.T) {
	if len(KeyboardLayouts) != 3 {
		t.Fatalf("layouts = %d", len(KeyboardLayouts))
	}
	ids := make(map[int]bool)
	for _, layout := range KeyboardLayouts {
		if layout.ID <= 0 || layout.Name == "" || ids[layout.ID] {
			t.Fatalf("invalid layout: %+v", layout)
		}
		ids[layout.ID] = true
	}
}
