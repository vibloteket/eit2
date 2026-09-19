# Asset licensing

Eit 2 currently contains no imported music, sound effects, backgrounds or game
art from the original Eit distribution.

Assets and sources:

| Asset | Source | License |
|---|---|---|
| Favicon and retained favicon concepts | Created for Eit 2 | AGPL-3.0-or-later |
| Generated game graphics | Drawn in code | AGPL-3.0-or-later |
| Go Regular font embedded through `golang.org/x/image/font/gofont/goregular` | The Go Authors | BSD 3-Clause |
| Fourteen Doodle Party WAV effects | Generated specifically for Eit 2 by `scripts/generate-audio` | AGPL-3.0-or-later |
| `internal/sound/audio/lobby-badinerie.wav` | J.S. Bach, BWV 1067/VII; project B2 rendering with VSCO 2 CE flute/cello and original synthetic pluck | Public-domain composition; CC0-1.0 samples; project contributions AGPL-3.0-or-later. See the source audit below. |
| `internal/sound/audio/gameplay-*.wav` (four loops) | Bach BWV1067 Bourrées and BWV1035 II; Handel HWV369 IV; Vivaldi RV428 III; project arrangements of public-domain historical material | Public-domain music; CC0-1.0 samples; project contributions AGPL-3.0-or-later. See MUSIC-SOURCES.md for historical-source and pilot-license distinctions. |

The generated audio is 44.1 kHz, 16-bit stereo PCM WAV. The effects cover menu
focus/selection, join/leave, rotate, lock, hard drop, line and four-line clears,
special pickup, incoming attack, Antidote, game over and winner. They combine
procedural mallet, wood, bell, pop and noise layers to match the hand-made
Doodle Party theme. The source parameters are kept in `scripts/generate-audio`
so those generated files are reproducible and project-owned. The former
Wooden Bounce music is retained only as an optional historical generator export,
not as an embedded game asset.

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
- byte-exact regeneration of the lobby loop and all14 procedural effects.

The reproducible source bundle is now in `music/badinerie/` and
`scripts/music/render-badinerie.ts`; it has no private workspace dependency.
Use `make music` to fetch verified sources and reproduce the lobby and four gameplay loops.
The audit does not grant rights to the modern reference repositories or imply
that their score files are CC0; only the pre-existing melody is retained.

## Four classical gameplay loops

All four are PCM16 stereo44.1kHz,96 quarter beats at112BPM (51.428571s).
File identities, hashes, score/source snapshots and audits are in
`music/classical/manifest.json`. `render-classical.ts` reproduces them;
`check-classical.ts` checks the source and signal contract.

The four arrangements retain the approved light flute/pizzicato/pluck sound,
with sustained flute for long notes and continuous quiet backing. Release tails
and ambience wrap without silence/fades. They replace Beethoven G1/G2; lobby
Badinerie and all14 effects are unchanged. Runtime levels remain gameplay.08,
lobby.16 and effects.36. Restart preserves the current track; fresh matches
cycle through the four in order.

Handel production is based on a new, independent transcription of Chrysander1879,
checked for all146 events against the audition. The private pilot's modern
Mutopia engraving remains CC BY-SA3.0; it is neither relabeled CC0 nor shipped
as a production input. This independent historical derivation is documented
in MUSIC-SOURCES.md and the source audits.

Historical Beethoven sources and their original notices remain in
`music/beethoven/`, but its WAVs are no longer embedded. Optional legacy
renderers write to `dist/audio/legacy/` only.

Before adding an asset, record its author, source URL, exact license and any
required attribution here. Do not copy legacy Eit assets whose rights or source
license have not been verified.
