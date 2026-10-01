package ui

import (
	core "github.com/vibloteket/eit2/internal/game"
	"github.com/vibloteket/eit2/internal/sound"
)

// The logical design space is logicalWidth x logicalHeight (16:9). Phone
// screens in landscape are typically 19.5:9 or wider, so on wider outside
// aspects the logical width follows the outside aspect (capped at 21:9) and
// the game fills the whole screen instead of pillarboxing. Narrower aspects
// (portrait phones) keep the 1280x720 letterbox unchanged.
const maxLogicalWidth = 1680 // 21:9

type screenLayout struct{ w int }

func layoutFor(width int) screenLayout {
	if width < logicalWidth {
		width = logicalWidth
	}
	if width > maxLogicalWidth {
		width = maxLogicalWidth
	}
	return screenLayout{w: width}
}

func (l screenLayout) cx() int { return l.w / 2 }

// shift recentres the 1280-wide design on wider screens.
func (l screenLayout) shift() int { return (l.w - logicalWidth) / 2 }

// right offsets elements anchored to the right edge of the 1280 design.
func (l screenLayout) right() int { return l.w - logicalWidth }

func (l screenLayout) startButton() imageRect {
	return imageRect{X: 490 + l.shift(), Y: 550, W: 300, H: 64}
}
func (l screenLayout) settingsButton() imageRect {
	return imageRect{X: 470 + l.shift(), Y: 632, W: 340, H: 60}
}
func (l screenLayout) exitButton() imageRect {
	return imageRect{X: 985 + l.shift(), Y: 632, W: 130, H: 60}
}

func (l screenLayout) lobbyMenuButtons() []imageRect {
	buttons := []imageRect{l.startButton(), l.settingsButton()}
	if !isWeb() {
		buttons = append(buttons, l.exitButton())
	}
	return buttons
}

func (l screenLayout) settingsRow(index int) imageRect {
	return imageRect{X: l.cx() - 310, Y: 165 + index*64, W: 620, H: 54}
}

// Left-anchored play controls keep their positions on all widths.
func (l screenLayout) debugPlayButton() imageRect { return imageRect{X: 45, Y: 280, W: 160, H: 62} }
func (l screenLayout) touchMenuButton() imageRect { return imageRect{X: 45, Y: 205, W: 160, H: 62} }

func (l screenLayout) resumeButton() imageRect {
	return imageRect{X: 375 + l.shift(), Y: 340, W: 160, H: 72}
}

func (l screenLayout) menuButtons(gameOver bool) (restart, back imageRect) {
	if gameOver {
		return imageRect{X: 455 + l.shift(), Y: 340, W: 180, H: 72}, imageRect{X: 650 + l.shift(), Y: 340, W: 180, H: 72}
	}
	return imageRect{X: 560 + l.shift(), Y: 340, W: 160, H: 72}, imageRect{X: 745 + l.shift(), Y: 340, W: 160, H: 72}
}

func (l screenLayout) debugCloseButton() imageRect {
	return imageRect{X: 1035 + l.shift(), Y: 110, W: 150, H: 60}
}
func (l screenLayout) debugPrevPlayer() imageRect {
	return imageRect{X: 150 + l.shift(), Y: 110, W: 90, H: 60}
}
func (l screenLayout) debugNextPlayer() imageRect {
	return imageRect{X: 390 + l.shift(), Y: 110, W: 90, H: 60}
}

func (l screenLayout) debugSoundButtons() []struct {
	Rect   imageRect
	Label  string
	Effect sound.Effect
} {
	shift := l.shift()
	return []struct {
		Rect   imageRect
		Label  string
		Effect sound.Effect
	}{
		{imageRect{105 + shift, 590, 160, 48}, "Lock", sound.Lock},
		{imageRect{275 + shift, 590, 160, 48}, "Line", sound.Line},
		{imageRect{445 + shift, 590, 160, 48}, "Four-line", sound.FourLine},
		{imageRect{615 + shift, 590, 160, 48}, "Pickup", sound.Pickup},
		{imageRect{785 + shift, 590, 160, 48}, "Attack", sound.Attack},
		{imageRect{955 + shift, 590, 160, 48}, "Game over", sound.GameOver},
	}
}

func (l screenLayout) debugSpecialButtons() []struct {
	Rect    imageRect
	Special core.Special
} {
	buttons := make([]struct {
		Rect    imageRect
		Special core.Special
	}, 0, len(core.AllSpecials))
	const columns, width, height, gapX, gapY = 4, 255, 52, 18, 8
	for i, special := range core.AllSpecials {
		row, column := i/columns, i%columns
		buttons = append(buttons, struct {
			Rect    imageRect
			Special core.Special
		}{Rect: imageRect{X: 105 + l.shift() + column*(width+gapX), Y: 190 + row*(height+gapY), W: width, H: height}, Special: special})
	}
	return buttons
}

// touchButtons keeps the left cluster anchored to the left edge and moves the
// right cluster with the right edge, so widened screens put the rotate/drop
// controls under the player's right thumb.
func (l screenLayout) touchButtons() []button {
	right := l.right()
	return []button{
		{Rect: imageRect{X: 15, Y: 420, W: 175, H: 120}, Label: "LEFT", Do: actionLeft},
		{Rect: imageRect{X: 210, Y: 420, W: 175, H: 120}, Label: "RIGHT", Do: actionRight},
		{Rect: imageRect{X: 125, Y: 570, W: 145, H: 120}, Label: "DOWN", Do: actionDown},
		{Rect: imageRect{X: 895 + right, Y: 420, W: 175, H: 120}, Label: "CCW", Do: actionCCW},
		{Rect: imageRect{X: 1090 + right, Y: 420, W: 175, H: 120}, Label: "CW", Do: actionCW},
		{Rect: imageRect{X: 970 + right, Y: 570, W: 150, H: 120}, Label: "DROP", Do: actionDrop},
		{Rect: imageRect{X: 1135 + right, Y: 570, W: 100, H: 120}, Label: "ANTI", Do: actionAnti},
	}
}
