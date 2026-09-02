# Asset licensing

Eit 2 currently contains no imported music, sound effects, backgrounds or game
art from the original Eit distribution.

Current visual assets are project-owned source files:

| Asset | Source | License |
|---|---|---|
| Favicon and retained favicon concepts | Created for Eit 2 | AGPL-3.0-or-later |
| Generated game graphics | Drawn in code | AGPL-3.0-or-later |
| Go Regular font embedded through `golang.org/x/image/font/gofont/goregular` | The Go Authors | BSD 3-Clause |
| Fourteen Doodle Party WAV effects and one background loop under `internal/sound/audio/` | Generated specifically for Eit 2 by `scripts/generate-audio` | AGPL-3.0-or-later |

The generated audio is 44.1 kHz, 16-bit stereo PCM WAV. The effects cover menu
focus/selection, join/leave, rotate, lock, hard drop, line and four-line clears,
special pickup, incoming attack, Antidote, game over and winner. They combine
procedural mallet, wood, bell, pop and noise layers to match the hand-made
Doodle Party theme. `music-loop.wav` is a roughly 36-second, 108 BPM procedural
mallet/tabletop loop. The source parameters are kept in `scripts/generate-audio`
so every shipped file is reproducible and project-owned.

Before adding an asset, record its author, source URL, exact license and any
required attribution here. Do not copy legacy Eit assets whose rights or source
license have not been verified.
