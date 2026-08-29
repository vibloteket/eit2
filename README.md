# Eit 2

Eit 2 is a TV-first couch multiplayer falling-block game for two to four local
players. It is a clean successor to the original
[`vibloteket/eit`](https://github.com/vibloteket/eit), built in Go with
[Ebitengine](https://ebitengine.org/).

## License

Eit 2 is free software licensed under
[`AGPL-3.0-or-later`](LICENSE). Network users of modified hosted versions must
be offered the corresponding source code. See [`NOTICE.md`](NOTICE.md) for
origin and third-party notices and [`ASSETS.md`](ASSETS.md) for asset licensing.

The current published version is stored in `VERSION` and shown in the lobby.
Every public deployment must increment it so testers can identify the exact
build. The web loader also appends the version to the Wasm URL to avoid stale
browser caches.

## Current milestone

This initial technical spike proves the project structure and its first targets:

- native Linux build;
- browser/WebAssembly build;
- deterministic tests for assigning up to four input devices;
- a 16:9 lobby showing four player slots and allowing solo play;
- keyboard, standard gamepad and touch join input;
- gamepad Start launches/pauses play, and disconnects pause the match until a
  controller presses A to reclaim the affected player slot;
- a first playable solo falling-block core with movement, rotation, gravity,
  soft/hard drop, a short lock delay, line clearing and scoring;
- multi-touch controls with hold/repeat and visual pressed feedback;
- pause/resume, restart, lobby return and a usable game-over flow;
- level-based gravity following the original Eit acceleration;
- independent random piece selection, retaining the original game's possibility
  of lucky and unlucky streaks;
- a fresh production seed for every match, while tests and recorded scenarios
  keep explicit reproducible seeds;
- specials skip their spawn cycle on an empty board and otherwise remain for
  18 seconds plus one second per ten settled blocks, capped at 30 seconds;
- four-second activation messages identify the special, sender and target;
- an opt-in Debug mode pauses play and can trigger any of all 22 specials using
  normal player and target routing;
- short project-owned WAV effects for lock, line clears, special pickup,
  incoming attacks and game over, plus a quiet looping prototype melody, with
  separate Mute/Music controls and Debug-mode sound status/tests.

Two to four joined players receive separate boards, device-routed input,
cycleable targets, elimination retargeting and a last-player-standing winner.
Clearing four rows sends two incomplete garbage rows to the selected target,
matching the original game's basic multiplayer attack. The first two special
blocks are implemented: Antidote is stored when collected, Clear empties the
collector's board, Blind hides the selected target's Next preview, Inverse
reverses the target's horizontal movement and rotation, Rabbit/Faster speeds up
the target, Turtle/Slower slows the collector, Bridge adds two disruption rows,
Question removes half of the target's placed blocks, Stair builds a diagonal,
Fill packs the target's lower ten rows, Flip vertically reverses the target's
settled structure, Switch exchanges settled structures, Packet sends one row
per line cleared for 20 seconds, Ring builds a hollow grey ring, Mini shrinks
settled blocks visually, Blink flashes the active piece, SZ restricts upcoming
pieces to S/Z, Trans/Ice makes settled blocks translucent, Castle replaces the
board with the original grey castle silhouette, Color/Blackout darkens the whole playfield
except for a circular spotlight around the active piece, Rumble shakes a
selection of settled blocks, and Background changes the target's
playfield colour. Pattern attacks animate from the bottom upward; they never overwrite the active piece, which instead settles on
the newly created structure using normal lock delay. One Antidote clears all active
effects, including helpful Slower stacks, making its
use a tactical trade-off. Solo mode targets Self for special effects so every
special can be tested, while ordinary four-line garbage attacks remain disabled
without opponents. Android is deliberately deferred; native Linux and web are
the active targets.

## Develop

Requires Go 1.26 or newer.

```sh
make install
make check
make run
```

The lobby is fully keyboard/gamepad navigable: arrows or D-pad move visible
focus, Enter or a joined gamepad's A activates, and Escape or Back exits. An
unjoined gamepad's A joins it. On a phone, tap once to join and then tap
**Start**. Keyboard layouts 1, 2 and 3 join with the corresponding number key,
or use **Join next keyboard**; Enter only activates the focused menu action.
Mouse clicks operate visible UI buttons and never create a touch player.
Browsers only expose connected gamepads after user interaction.

Keyboard 1 uses A/D, S, Q/W, left Shift, E and Tab. Keyboard 2 uses arrows,
comma/period, right Shift, slash and Enter. Emergency Keyboard 3 uses J/L, K,
U/I, Space, O and P; three-player keyboard use depends on the hardware's key
rollover. Touch screens use controls around the board for movement, rotation,
hard drop and Antidote.

The lobby's **Controller debug** panel shows connected controller names/IDs,
player assignment, standard mapping availability, axis count, SDL GUID where
available and live pressed buttons.

For gameplay testing, enable **Debug mode** in the lobby. During a match,
open **Debug**, choose a source player and tap a special. Direct/self effects
apply to that player while offensive effects follow that player's current
target. The panel pauses simulation while open.

## Build outputs

```sh
make build             # dist/native/eit2
make build-web         # dist/web/
make verify-packages   # build and verify Linux, Windows and web archives
```

Versioned archives and checksums are written under `dist/packages/`. CI uploads
separate `eit2-linux-x86_64`, `eit2-windows-x86_64` and `eit2-web` artifacts so
testers only download the target they need. Desktop builds support
`--fullscreen`, `--windowed` and `--version`. Serve the web
directory over HTTP; browsers do not normally allow the Wasm loader to run
directly from `file://`.

## Architecture

- `internal/lobby`: engine-independent local device assignment
- `internal/ui`: Ebitengine presentation and physical input adapter
- `cmd/eit2`: native and WebAssembly entry point
- `web`: browser loader

Gameplay rules live in the engine-independent `internal/game` package so they
can be tested without opening a window.
