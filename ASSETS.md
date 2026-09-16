# Asset licensing

Eit 2 currently contains no imported music, sound effects, backgrounds or game
art from the original Eit distribution.

Assets and sources:

| Asset | Source | License |
|---|---|---|
| Favicon and retained favicon concepts | Created for Eit 2 | AGPL-3.0-or-later |
| Generated game graphics | Drawn in code | AGPL-3.0-or-later |
| Go Regular font embedded through `golang.org/x/image/font/gofont/goregular` | The Go Authors | BSD 3-Clause |
| Fourteen Doodle Party WAV effects and the match track `music-loop.wav` | Generated specifically for Eit 2 by `scripts/generate-audio` | AGPL-3.0-or-later |
| `internal/sound/audio/lobby-badinerie.wav` | J.S. Bach, BWV 1067/VII; project B2 rendering with VSCO 2 CE flute/cello and original synthetic pluck | Public-domain composition; CC0-1.0 samples; project contributions AGPL-3.0-or-later. See the source audit below. |

The generated audio is 44.1 kHz, 16-bit stereo PCM WAV. The effects cover menu
focus/selection, join/leave, rotate, lock, hard drop, line and four-line clears,
special pickup, incoming attack, Antidote, game over and winner. They combine
procedural mallet, wood, bell, pop and noise layers to match the hand-made
Doodle Party theme. `music-loop.wav` is a roughly 30-second, 128 BPM procedural
Wooden Bounce loop with short marimba notes, syncopated plucked bass, eighth-note
shaker, wooden beat pulses and cardboard claps. It intentionally has no pads,
drones or sustained bell tones. The source parameters are kept in `scripts/generate-audio`
so those generated files are reproducible and project-owned.

## Badinerie lobby music

The full AABB B2 rendering includes the selected50ms low-flute sample lead-in
correction. It is88.88889 seconds at108BPM, stereo PCM16/44.1kHz, SHA-256:
`d41d29242ea46c8df0014eb84910dc72ced90f1ecd7149d06346e943d95361b5`.

The source audit is in [MUSIC-SOURCES.md](MUSIC-SOURCES.md), with credits in
[CREDITS.md](CREDITS.md). It records:

- the1885 public-domain score and Bach composition;
- the exact230 principal-note comparison against an independent transcription;
- the exclusion of unlicensed modern score files, added arrangements and recordings;
- all12 CC0 samples, mapping files, author credits and pinned upstream hashes;
- project-created accompaniment, performance and synthetic pluck;
- byte-exact regeneration of the lobby loop and all15 procedural audio files.

The reproducible source bundle is now in `music/badinerie/` and
`scripts/music/render-badinerie.ts`; it has no private workspace dependency.
Use `make music` to fetch verified sources and reproduce the approved loop.
The audit does not grant rights to the modern reference repositories or imply
that their score files are CC0; only the pre-existing melody is retained.

Before adding an asset, record its author, source URL, exact license and any
required attribution here. Do not copy legacy Eit assets whose rights or source
license have not been verified.
