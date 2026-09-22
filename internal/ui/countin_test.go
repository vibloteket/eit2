package ui

import (
	"encoding/json"
	"github.com/hajimehoshi/ebiten/v2"
	core "github.com/vibloteket/eit2/internal/game"
	"github.com/vibloteket/eit2/internal/lobby"
	"github.com/vibloteket/eit2/internal/sound"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestCountInThreeBeatsAndOneDownbeat(t *testing.T) {
	for _, fps := range []int{5, 10, 30, 60, 120} {
		c := countInState{active: true}
		click, goNow := c.step(0)
		if !click || goNow {
			t.Fatal("missing initial preparation beat")
		}
		clicks, starts := 1, 0
		elapsed := time.Duration(0)
		dt := time.Second / time.Duration(fps)
		for c.active {
			elapsed += dt
			click, goNow = c.step(dt)
			if click {
				clicks++
			}
			if goNow {
				starts++
			}
			if elapsed > 3*time.Second {
				t.Fatal("slow-TPS countdown stretched")
			}
		}
		if clicks != 3 || starts != 1 || elapsed < 3*countInBeat || elapsed > 3*(countInBeat+dt) {
			t.Fatalf("fps%d clicks%d starts%d elapsed%v", fps, clicks, starts, elapsed)
		}
		c.step(time.Second)
		if c.active || c.goFlash != 0 {
			t.Fatal("count-in did not settle")
		}
	}
}

func TestCountInStallCannotBurstClicks(t *testing.T) {
	c := countInState{active: true}
	c.step(0)
	click, goNow := c.step(10 * time.Second)
	if !click || goNow || c.strikes != 2 {
		t.Fatal("stall skipped preparation")
	}
	for i := 0; i < 100; i++ {
		click, goNow = c.step(0)
		if click || goNow {
			t.Fatal("catch-up burst")
		}
	}
	c.step(countInBeat)
	_, goNow = c.step(countInBeat)
	if !goNow || c.strikes != 3 {
		t.Fatal("recovery did not finish")
	}
}

func TestCountInMatchesGameplayTempo(t *testing.T) {
	data, err := os.ReadFile("../../music/classical/manifest.json")
	if err != nil {
		t.Fatal(err)
	}
	var manifest struct {
		BPM int `json:"bpm"`
	}
	if err = json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.BPM != countInBPM {
		t.Fatal("count-in and gameplay tempos disagree")
	}
}

func TestCountInFreezesAllBoardsAndStartsClockOnDownbeat(t *testing.T) {
	for n := 1; n <= 4; n++ {
		g := Game{view: viewLobby}
		for i := 0; i < n; i++ {
			g.Lobby.Join(lobby.Device{Kind: lobby.DeviceGamepad, ID: i + 1, Name: "test"})
		}
		g.start()
		base := g.countIn.last
		before := make([]core.Game, len(g.players))
		for i, p := range g.players {
			before[i] = *p
		}
		if g.countIn.strikes != 1 || !g.matchStarted.IsZero() || g.elapsedMatchTime() != 0 {
			t.Fatal("initial count/clock")
		}
		for frame := 1; frame < 99; frame++ {
			if !g.updateCountInAt(base.Add(time.Duration(frame) * time.Second / 60)) {
				t.Fatal("early GO")
			}
			for i, p := range g.players {
				if !reflect.DeepEqual(*p, before[i]) {
					t.Fatalf("player%d moved before GO", i)
				}
			}
			if !g.matchStarted.IsZero() {
				t.Fatal("clock includes preparation")
			}
		}
		now := base.Add(99 * time.Second / 60)
		if !g.updateCountInAt(now) || g.countIn.active || !g.matchStarted.Equal(now) || g.countIn.goFlash != countInGoFlash {
			t.Fatal("GO not synchronized")
		}
		if g.updateCountInAt(now.Add(time.Second / 60)) {
			t.Fatal("post-GO input gated")
		}
	}
}

func TestCountInPauseDisconnectRestartAndCancel(t *testing.T) {
	g := Game{view: viewLobby}
	g.Lobby.Join(lobby.Device{Kind: lobby.DeviceTouch, Name: "test"})
	g.start()
	base := g.countIn.last
	g.updateCountInAt(base.Add(200 * time.Millisecond))
	at := g.countIn.elapsed
	g.paused = true
	g.updateCountInAt(base.Add(10 * time.Second))
	g.updateCountInAt(base.Add(11 * time.Second))
	if g.countIn.elapsed != at {
		t.Fatal("pause advanced count")
	}
	g.paused = false
	g.disconnectedPlayer = 0
	g.updateCountInAt(base.Add(12 * time.Second))
	if g.countIn.elapsed != at {
		t.Fatal("disconnect advanced count")
	}
	g.disconnectedPlayer = -1
	g.updateCountInAt(base.Add(12*time.Second + time.Second/60))
	if g.countIn.elapsed != at+time.Second/60 {
		t.Fatal("resume counted paused time")
	}
	track, next := g.selectedMusicTrack(), g.nextMatchMusic
	g.restart()
	if g.countIn.elapsed != 0 || g.countIn.strikes != 1 || g.selectedMusicTrack() != track || g.nextMatchMusic != next {
		t.Fatal("Restart changed selection/count")
	}
	base = g.countIn.last
	for i := 1; i <= 100; i++ {
		g.updateCountInAt(base.Add(time.Duration(i) * time.Second / 60))
	}
	g.setPaused(true)
	g.setPaused(false)
	if g.countIn.active {
		t.Fatal("ordinary Resume starts new count")
	}
	g.restart()
	g.backToLobby()
	g.updateCountInAt(time.Now().Add(10 * time.Second))
	if g.countIn.active || g.countIn.goFlash != 0 || g.selectedMusicTrack() != sound.LobbyMusic {
		t.Fatal("cancel leaked count-in")
	}
}

func TestCountInHeldControlsRequireReleaseWithoutBlockingOtherPlayers(t *testing.T) {
	g := Game{countIn: countInState{active: true}, countDownKeys: map[ebiten.Key]bool{ebiten.KeyS: true}, countTouches: map[ebiten.TouchID]bool{7: true}}
	if !g.blockPadUntilNeutral(0, true) || !g.blockPadUntilNeutral(1, false) {
		t.Fatal("pads not gated during preparation")
	}
	g.countIn.active = false
	if !g.blockPadUntilNeutral(0, true) || g.blockPadUntilNeutral(1, true) {
		t.Fatal("held pad blocked wrong player")
	}
	if g.blockPadUntilNeutral(0, false) || g.blockPadUntilNeutral(0, true) {
		t.Fatal("fresh pad input not released")
	}
	if !g.blockDownUntilReleased(ebiten.KeyS, true) || g.blockDownUntilReleased(ebiten.KeyArrowDown, true) {
		t.Fatal("held down key blocked wrong layout")
	}
	if g.blockDownUntilReleased(ebiten.KeyS, false) || g.blockDownUntilReleased(ebiten.KeyS, true) {
		t.Fatal("fresh down key not released")
	}
	if !g.touchBlockedByCountIn(7) || g.touchBlockedByCountIn(8) {
		t.Fatal("held touch blocked wrong contact")
	}
}

func TestCountInBatonRemainsOnPaperCard(t *testing.T) {
	c := countInState{active: true}
	c.step(0)
	for i := 0; i < 120; i++ {
		c.step(time.Second / 60)
		x, y := countInTip(c)
		if x < 580 || x > 710 || y < 309 || y > 394 {
			t.Fatal("baton leaves its area or overlaps the READY label")
		}
	}
}

func TestLateBeatDoesNotRushFollowingBeat(t *testing.T) {
	c := countInState{active: true}
	c.step(0)
	c.step(countInBeat - 20*time.Millisecond)
	click, start := c.step(200 * time.Millisecond)
	if !click || start || c.strikes != 2 {
		t.Fatal("late second beat")
	}
	click, start = c.step(40 * time.Millisecond)
	if click || start {
		t.Fatal("late beat caused immediate catch-up click")
	}
	click, start = c.step(countInBeat - 40*time.Millisecond)
	if !click || start || c.strikes != 3 {
		t.Fatal("third beat not separated")
	}
	_, start = c.step(40 * time.Millisecond)
	if start {
		t.Fatal("GO rushed after final click")
	}
}
