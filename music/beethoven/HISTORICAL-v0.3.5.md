# Historical Beethoven gameplay music (through v0.3.5)

Not embedded or selected in v0.3.6. Source files and reproduction remain available.
Legacy renderers now write loops under `dist/audio/legacy/`, never into embedded audio.

## 3. Match music: Beethoven, Rondo a capriccio, Op.129

**Composer:** Ludwig van Beethoven (1770–1827), *Rage Over a Lost Penny*.
The composition is public domain. This track uses an explicitly public-domain
Mutopia transcription, not an unlicensed commercial MIDI performance.

- [Mutopia item498](https://www.mutopiaproject.org/cgibin/piece-info.cgi?id=498).
- Typesetter/maintainer: **Magnus Lewis-Smith**. LilyPond update by
  **Javier Ruiz-Alma**,2015. The transcription identifies Augener's Edition
  as its source.
- Pinned source revision: `2144afd6f52d56c5b6995b8b589ef1268b3139f0` in
  [MutopiaProject](https://github.com/MutopiaProject/MutopiaProject/tree/2144afd6f52d56c5b6995b8b589ef1268b3139f0/ftp/BeethovenLv/O129/beethoven_rondo_op129).
- The main LilyPond file explicitly declares `license = "Public Domain"` and
  says it is placed in the public domain by the typesetter, **free to distribute,
  modify and perform**. This is a PD declaration, not a CC0 label.
- All8 LilyPond source files and Mutopia's accompanying LilyPond-generated MIDI
  are bundled under `music/beethoven/sources/mutopia/`. Source URLs and SHA-256
  hashes are recorded in `music/beethoven/sources.json`.

The first project arrangement is **G1 at126BPM**, not the earlier144BPM test.
It uses bars1–56: A=1–24, B=25–56, with AABB repeats for224 quarter-note beats
and106.66667 seconds (~1:47). It is an arranged excerpt, not the entire work.
The final outgoing D pickup is omitted to close on G before the loop repeats.
The highest RH note at each onset supplies the lead;11 low RH figures use the
project pluck at their written pitches rather than forcing them into flute
range. LH notes are reduced to a quarter-note pizzicato bass and an eighth-note
pluck from sounding source pitches. Repeats thin a few accompaniment notes.

`scripts/music/check-beethoven.ts` verifies all9 archived input hashes, the
explicit PD statement, MIDI format and the complete reduction into242 lead
events plus accompaniment. It ignores the source piano patch, tempo and
performance dynamics: the selected126BPM, mix and render are project choices.
No pianist's recording or external synthesized audio is used.

The selected PCM file is reproduced **unchanged**, including the B2 low-flute
50ms lead-in correction and original G1 master gain. Only its in-game playback
volume is lower than the lobby track (`.08` versus `.16`, effects remain`.36`).
This leaves a measured RMS margin above4dB for drop/line/four-line/attack cues;
that is a signal-level check, not a universal perceptual guarantee.

### Additional gameplay loop G2 (v0.3.5)

G2 uses different, previously unused sections of the **same Beethoven work**:
unfolded source beats112–314 (written bars57–128 with the source repeats).
It does not reuse G1's source range0–112 and is not a tempo-only variant of G1.
The G-minor episodes, varied return and E-major episode create a contrasting
loop; the final D harmony resolves to the G-minor opening. No new composition
or external arrangement is claimed.

G2 runs at112BPM,202 beats,108.21429 seconds. It uses the same relative mix,
articulations and reduction algorithm as G1;83 low RH figures use the pluck.
The new master is leveled close to G1 with peak headroom, and playback remains
at.08. The existing G1 and Badinerie WAVs are unchanged.


Every new match started from the lobby alternates G1/G2 (G1 first); each starts
at its beginning. Restart resets the current selection without advancing the
playlist. Returning to Credits/lobby still preserves Badinerie's position.

The same explicitly public-domain Mutopia inputs are used. Run
`bun scripts/music/check-beethoven.ts --g2` to reproduce the504 lead events
and accompaniment in `music/beethoven/score-g2.json`. Source hashes, waveform
hashes, note ranges and loop boundaries are verified before publication.
