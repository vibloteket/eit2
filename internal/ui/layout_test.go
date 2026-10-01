package ui

import "testing"

func TestLayoutKeepsDesignWidthAtSixteenNine(t *testing.T) {
	g := &Game{}
	if w, h := g.Layout(1280, 720); w != 1280 || h != 720 {
		t.Fatalf("16:9 outside = %dx%d, want 1280x720", w, h)
	}
	if g.L() != layoutFor(1280) {
		t.Fatal("layout must stay at the 1280 design width")
	}
}

func TestLayoutWidensOnPhoneLandscape(t *testing.T) {
	g := &Game{}
	// A 2532x1170 phone in landscape is ~19.5:9.
	w, h := g.Layout(2532, 1170)
	if h != 720 || w <= 1280 || w > maxLogicalWidth {
		t.Fatalf("phone landscape = %dx%d, want 1280 < w <= %d, h=720", w, h, maxLogicalWidth)
	}
	// The returned logical aspect must match the outside aspect (±1 px rounding).
	if got, want := w*1170, 2532*720; got < want-2532 || got > want {
		t.Fatalf("logical aspect drifted: got %d, want ~%d", got, want)
	}
}

func TestLayoutKeepsLetterboxInPortrait(t *testing.T) {
	g := &Game{}
	if w, h := g.Layout(1170, 2532); w != 1280 || h != 720 {
		t.Fatalf("portrait outside = %dx%d, want 1280x720 letterbox", w, h)
	}
}

func TestLayoutCapsUltrawide(t *testing.T) {
	g := &Game{}
	if w, _ := g.Layout(5120, 1440); w != maxLogicalWidth {
		t.Fatalf("32:9 outside = %d, want cap %d", w, maxLogicalWidth)
	}
}

func TestWideLayoutMovesRightAnchoredControls(t *testing.T) {
	base := layoutFor(logicalWidth)
	wide := layoutFor(1560)
	if wide == base {
		t.Fatal("wide layout must differ")
	}
	baseButtons := base.touchButtons()
	wideButtons := wide.touchButtons()
	for i := range baseButtons {
		if baseButtons[i].Rect.Y != wideButtons[i].Rect.Y || baseButtons[i].Rect.W != wideButtons[i].Rect.W {
			t.Fatal("touch control geometry must only move horizontally")
		}
	}
	// Left cluster stays at the left edge; right cluster follows the right edge.
	if wideButtons[0].Rect != baseButtons[0].Rect || wideButtons[2].Rect != baseButtons[2].Rect {
		t.Fatal("left touch cluster must not move")
	}
	if got := wideButtons[6].Rect.X + wideButtons[6].Rect.W; got != wide.w-45 {
		t.Fatalf("right touch cluster must keep its 45 px right margin, got edge %d for width %d", got, wide.w)
	}
	// Centered groups shift by half the extra width.
	if got := wide.startButton().X - base.startButton().X; got != wide.shift() {
		t.Fatalf("start button shift = %d, want %d", got, wide.shift())
	}
	row := wide.settingsRow(0)
	if row.X+row.W/2 != wide.cx() {
		t.Fatal("settings rows must stay centred")
	}
}
