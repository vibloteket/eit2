# Gameplay music — Beethoven G1 and G2

Selected by the user after comparing144BPM and126BPM. The106.67-second126BPM
loop replaces Wooden Bounce in matches. Badinerie remains unchanged in the
lobby and the full-screen Credits view.

## Additional G2 loop (v0.3.5)

G2 uses previously unused source beats112–314 of the same PD Beethoven work,
not a new composition. The G-minor and E-major episodes give it a contrasting
character. At112BPM its202 beats last108.21429s. G1 stays unchanged at126BPM.
Both share the light flute/pizzicato/pluck palette and similar mastered levels;
G2 adds only the lowest CC0 cello pizzicato region from the same pinned library.

New matches from the lobby alternate G1/G2, starting with G1 after launch.
Restart resets the current choice without advancing. Credits/lobby keep Bach.
This deterministic playlist does not consume the gameplay RNG. The loop is
built from source repeats and varied episodes, not copies of G1.

`render-beethoven-g2.ts` reproduces the new asset; `check-beethoven.ts --g2`
checks its504 lead events and source reduction. Raw PCM is deliberately retained
for verified fidelity/looping; a future compressed/streamed format requires its
own tests. The additional track increases the web download size.

## G1 audio and source

- Asset: `internal/sound/audio/gameplay-beethoven.wav`.
- SHA-256: `52285ceff9158e6fb683af7c25d58c898344e5a1c4f6532e7dcd0a8e9ea3470a`.
- PCM16 stereo,44.1kHz,4,704,000 frames,126BPM,224 quarter-note beats.
- Same pitches, instrument mix, low-flute attack correction and master gain as
  the selected listening candidate; not a resampled/slowed recording.
- Excerpt from Beethoven Op.129, bars1–56, arranged AABB. It is not the whole
  piano work. Lowest RH figures use the original synthetic pluck; the outgoing
  pickup into the next episode is omitted to make a G-major loop closure.
- Mutopia score and generated MIDI explicitly public domain;9 source files,
  notices and hashes are archived in `music/beethoven/`. The source checker
  reproduces the242 lead events and LH reduction from the archived MIDI.
- Shared CC0 VSCO sample manifest;11 of the existing12 samples used. No new
  instrument library or external performance recording.

## Runtime mix

The selected file is unchanged. Its mastered level is higher than the former
procedural track, so playback uses`.08` rather than the lobby's`.16`. Effects
remain`.36`. Tests require the drop/line/four-line/attack cues to have at least
4dB whole-file RMS margin over the music at those playback settings and check
single-effect-plus-music peak headroom. These are signal measurements, not a
claim that every possible busy multiplayer mix has been subjectively auditioned.

Scene switching, pause/results behavior, mute/music-off and browser interaction
gating reuse the established manager. Since v0.3.4, explicit Restart resets
match music to its beginning; Resume does not. Since v0.3.5 a fresh match
explicitly starts the next selection from its beginning.
Restart cannot unmute or enable music, and it does not reset the lobby track.
No new controls or automatic ducking were added.

## Reproduction and checks

`make music` rebuilds both tracks with expected-hash checks before replacing
assets. `bun scripts/music/check-beethoven.ts` validates the public-domain
source declaration, all source hashes and the documented note reduction.
Normal builds use the committed WAVs and do not need Bun or sample downloads.

The SFX generator now emits only14 effects by default. Its historical
`-legacy-music` option can export Wooden Bounce into a separate directory;
that old track must not be reintroduced into the embedded audio set. All14
current effects remain byte-identical, and the source audit inventory is tested.

Credits on-screen and in the distribution include Beethoven and the Mutopia
typesetter without removing Bach/VSCO credits. The screen still fits without
scrolling; layout tests and browser checks cover this.
