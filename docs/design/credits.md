# Credits / About screen

## Current scope (v0.3.9)

Credits are reached from **Settings → About & Credits**, not from a separate
lobby utility button. The screen remains a separate full-screen view with the
project, music/audio, technology and license credits. Full legal/source details
ship in `CREDITS.md`, `MUSIC-SOURCES.md`, `ASSETS.md` and `LICENSES/`.

- Settings and Credits keep the lobby music position; no match music starts.
- Closing Credits returns to Settings with About & Credits focused. Closing
  Settings returns to the lobby with Settings focused.
- Only fresh key/button/pointer input dismisses Credits; the held input that
  opened it cannot immediately close it. Return input is consumed and cannot
  join/start/change audio.
- Keyboard, mouse, touch and any raw gamepad button can return, including an
  unjoined controller. Analog stick movement alone is not a press.
- The static text must fit the 1280×720 design space without scrolling.

Historical note: Credits was introduced as a direct lobby button in v0.3.2.
The move under Settings in v0.3.9 is intentional to keep the lobby focused on
joining and Start while retaining an in-game About location.

## Verification

`internal/ui/credits_test.go` covers Settings→Credits→Settings return, text
bounds and lobby button geometry. Browser smoke exercises keyboard, pointer,
touch and synthetic gamepad paths through Settings, held-opening input, and
return without leaking into lobby actions.
