# Credits screen — introduced in v0.3.2, updated in v0.3.3

## Scope and refinement

The requested workflow is explicit: a Credits button in the lobby opens a
**separate full-screen view**, not a popup. Use static text if it fits; otherwise
scroll. A fresh key/button press returns to the lobby. Existing players, audio
preferences and gameplay must not change.

- Audience/platforms: existing TV-first keyboard, mouse, touch and gamepad users;
  Linux/native and browser builds.
- MVP: one static screen using the current paper palette and readable type.
  The complete short credits fit the1280x720 design space; automated text-width
  and vertical-bound checks prove no scrolling is needed for this version.
- Contents: project/original game, Bach, Beethoven, project audio work, Mutopia
  and its typesetter, VSCO recordists/sample editor, Ebitengine, Go font,
  licenses and full-source link.
- No new persistence, external page load, authentication, data sharing,
  interactive links, animations or modal stack. Detailed dependency/legal
  notices ship as `CREDITS.md`, `MUSIC-SOURCES.md` and `LICENSES/`.
- Input: keyboard, mouse button, tap or any raw gamepad button, including an
  unjoined controller. Analog stick movement alone is not a button press.
- Guard: only *just-pressed* input dismisses. The held input that opened Credits
  cannot close it immediately. The return input is consumed in the Credits
  view and cannot join/leave a player, start a match or toggle audio.
- Return: preserve lobby players/settings, restore focus to Credits, and keep
  the same lobby music position throughout. No match music in Credits.
- Layout: Credits occupies the next utility-row slot; native Exit moves right.
  Keyboard/controller traversal follows the visible order; web cannot select
  native Exit.

## Acceptance / verification

1. Credits is a distinct view with its own draw/update path; no lobby boards or
   gameplay behind it.
2. Button, navigation and return work with keyboard, pointer/touch and gamepad.
3. Holding the opening button does not immediately dismiss the screen.
4. Any fresh return press does not leak into lobby actions.
5. Text/button bounds fit, credits are complete, and music/settings are retained.
6. All Go tests, native/Wasm builds, browser checks and distribution notice checks
   pass before publishing.

Tests are in `internal/ui/credits_test.go`, `internal/ui/music_test.go` and
`internal/controls/menu_test.go`. Browser smoke checks exercise the real view
transitions, held input, touch and an unjoined synthetic gamepad.
