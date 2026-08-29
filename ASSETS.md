# Asset licensing

Eit 2 currently contains no imported music, sound effects, backgrounds or game
art from the original Eit distribution.

Current visual assets are project-owned source files:

| Asset | Source | License |
|---|---|---|
| Favicon and retained favicon concepts | Created for Eit 2 | AGPL-3.0-or-later |
| Generated game graphics | Drawn in code | AGPL-3.0-or-later |
| Go Regular font embedded through `golang.org/x/image/font/gofont/goregular` | The Go Authors | BSD 3-Clause |
| Six prototype WAV effects and one background loop under `internal/sound/audio/` | Generated specifically for Eit 2 by `scripts/generate-audio` | AGPL-3.0-or-later |

The generated effects are 44.1 kHz, 16-bit stereo PCM WAV files for lock, line
clear, four-line clear, special pickup, incoming attack and game over. There are
intentionally no move or rotation sounds. `music-loop.wav` is a quiet four-second
prototype melody intended for continuous audio/backend testing, not final music.

Before adding an asset, record its author, source URL, exact license and any
required attribution here. Do not copy legacy Eit assets whose rights or source
license have not been verified.
