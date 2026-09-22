package ui

import (
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/vibloteket/eit2/internal/sound"
)

const countInBPM = 112
const countInBeat = time.Minute / countInBPM
const countInGoFlash = 250 * time.Millisecond
const countInPivotX, countInPivotY, countInBatonLength = 596, 353, 66

type countInState struct {
	active                    bool
	strikes                   int
	elapsed, beatAge, goFlash time.Duration
	last                      time.Time
}

// Presentation follows active monotonic time, not gameplay TPS. A slow device
// must not turn a short musical preparation into a long wait. Cap a stalled
// update to one beat so recovery cannot burst all three clicks together.
func (c *countInState) step(delta time.Duration) (click, goNow bool) {
	if delta < 0 {
		delta = 0
	}
	if !c.active {
		c.goFlash = max(0, c.goFlash-delta)
		return false, false
	}
	delta = min(delta, countInBeat)
	c.elapsed += delta
	c.beatAge += delta
	if c.strikes > 0 && c.beatAge < countInBeat {
		return false, false
	}
	// Anchor the next beat to the one actually presented. Carrying lateness
	// forward could squeeze two clicks (or the GO) into adjacent updates.
	c.beatAge = 0
	if c.strikes == 3 {
		c.active = false
		c.goFlash = countInGoFlash
		return false, true
	}
	c.strikes++
	return true, false
}

// The GO update advances simulation once, but consumes gameplay input in that
// update. Music is released by Update's ordinary audio synchronization below.
func (g *Game) updateCountIn() bool { return g.updateCountInAt(time.Now()) }

func (g *Game) updateCountInAt(now time.Time) bool {
	delta := time.Duration(0)
	if !g.countIn.last.IsZero() {
		delta = now.Sub(g.countIn.last)
	}
	g.countIn.last = now
	if g.view != viewPlay || g.paused || g.debugOpen || g.disconnectedPlayer >= 0 {
		return g.countIn.active
	}
	wasActive := g.countIn.active
	click, goNow := g.countIn.step(delta)
	if click && g.sound != nil {
		g.sound.Play(sound.CountIn)
	}
	if goNow {
		g.matchStarted = now
		g.pausedAt, g.pausedDuration = time.Time{}, 0
		clear(g.heldActions)
		clear(g.padHeld)
		if g.match != nil {
			g.match.Tick()
		}
	}
	return wasActive
}

func (g *Game) consumeCountInInputs() {
	if g.countDownKeys == nil {
		g.countDownKeys = make(map[ebiten.Key]bool)
	}
	for _, layout := range keyboardLayouts {
		g.countDownKeys[layout.Down] = ebiten.IsKeyPressed(layout.Down)
	}
	if g.countTouches == nil {
		g.countTouches = make(map[ebiten.TouchID]bool)
	}
	for _, id := range g.touchIDs {
		g.countTouches[id] = true
	}
	clear(g.heldActions)
	clear(g.padHeld)
}

func (g *Game) touchBlockedByCountIn(id ebiten.TouchID) bool {
	return g.countTouches[id]
}

func (g *Game) blockDownUntilReleased(key ebiten.Key, held bool) bool {
	if !held {
		delete(g.countDownKeys, key)
	}
	return g.countDownKeys[key]
}

func gamepadGameplayHeld(id ebiten.GamepadID) bool {
	if math.Abs(ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickHorizontal)) > .55 || ebiten.StandardGamepadAxisValue(id, ebiten.StandardGamepadAxisLeftStickVertical) > .55 {
		return true
	}
	for _, button := range []ebiten.StandardGamepadButton{
		ebiten.StandardGamepadButtonLeftLeft, ebiten.StandardGamepadButtonLeftRight, ebiten.StandardGamepadButtonLeftBottom,
		ebiten.StandardGamepadButtonRightLeft, ebiten.StandardGamepadButtonRightBottom, ebiten.StandardGamepadButtonRightRight,
		ebiten.StandardGamepadButtonFrontTopLeft, ebiten.StandardGamepadButtonFrontTopRight,
	} {
		if ebiten.IsStandardGamepadButtonPressed(id, button) {
			return true
		}
	}
	return false
}

func (g *Game) countInBlocksPad(player int, id ebiten.GamepadID) bool {
	return g.blockPadUntilNeutral(player, gamepadGameplayHeld(id))
}

func (g *Game) blockPadUntilNeutral(player int, held bool) bool {
	if g.countPads == nil {
		g.countPads = make(map[int]bool)
	}
	if g.countIn.active {
		g.countPads[player] = held
		return true
	}
	if g.countPads[player] {
		if held {
			return true
		}
		delete(g.countPads, player)
	}
	return false
}

func countInAngle(c countInState) float64 {
	if !c.active {
		return .60
	}
	phase := float64(c.beatAge%countInBeat) / float64(countInBeat)
	stroke := math.Max(0, 1-math.Min(phase, 1-phase)/.22)
	return -.65 + 1.25*stroke*stroke
}

func countInTip(c countInState) (float32, float32) {
	angle := countInAngle(c)
	return float32(countInPivotX + countInBatonLength*math.Cos(angle)), float32(countInPivotY + countInBatonLength*math.Sin(angle))
}

func (g *Game) drawCountIn(screen *ebiten.Image) {
	if !g.countIn.active && g.countIn.goFlash == 0 {
		return
	}
	// One shared paper card is equally visible from all four couch positions.
	// Simple batched geometry keeps the small moving marks lightweight.
	ebitenutil.DrawRect(screen, 456, 258, 368, 188, boardInk)
	ebitenutil.DrawRect(screen, 461, 263, 358, 178, paperLight)
	label := "READY"
	if !g.countIn.active {
		label = "GO!"
	}
	drawCenteredText(screen, label, g.face(24), 640, 272, accent)
	angle := countInAngle(g.countIn)
	x, y := countInTip(g.countIn)
	vector.StrokeLine(screen, countInPivotX-2, countInPivotY+2, x-2, y+2, 6, muted, false)
	vector.StrokeLine(screen, countInPivotX, countInPivotY, x, y, 4, white, false)
	vector.DrawFilledCircle(screen, countInPivotX, countInPivotY, 7, accent, false)
	vector.StrokeLine(screen, 575, 392, 715, 392, 3, boardInk, false)
	vector.StrokeLine(screen, 640, 392, 633, 400, 2, muted, false)
	if angle > .20 {
		for i := 0; i < 3; i++ {
			sx := float32(638 + i*10)
			vector.StrokeLine(screen, sx, 389, sx+4, 381-float32(i%2)*4, 2, accent, false)
		}
	}
	for i := 0; i < 3; i++ {
		x := float32(596 + i*44)
		vector.StrokeCircle(screen, x, 421, 8, 2, muted, false)
		if i < g.countIn.strikes {
			vector.DrawFilledCircle(screen, x, 421, 5, accent, false)
		}
	}
}

// Remove old touch IDs once released; a held Start/Restart finger cannot fall
// through to a newly appearing play control when the countdown ends.
func (g *Game) releaseCountInTouches() {
	// Browser touch identifiers may be reused for a genuinely new press.
	for _, id := range g.pressedIDs {
		delete(g.countTouches, id)
	}
	for id := range g.countTouches {
		if inpututil.TouchPressDuration(id) == 0 {
			delete(g.countTouches, id)
		}
	}
}
