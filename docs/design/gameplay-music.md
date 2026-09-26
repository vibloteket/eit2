# Gameplay music — four classical loops (v0.3.6)

## Approved scope

User approved all four flute/pizzicato/pluck auditions on2026-09-19, then
explicitly approved replacing Beethoven while keeping title/lobby music.
The finished loops retain those motif arrangements, not newly expanded full
movements. Historical G1/G2 design is in `gameplay-music-v0.3.5.md`.

1. Bach BWV1067 Bourrée I/II: first strains, I–II–II–I.
2. Handel HWV369 IV Allegro: first8 bars repeated.
3. Bach BWV1035 II Allegro: first16 full bars three times, no unmatched pickup.
4. Vivaldi RV428 III Allegro: solo episode bars16–31 repeated; not the opening
   tutti where flute is frequently resting.

All four:112 quarter-note BPM,96 quarter beats,2,268,000 stereo frames at44.1kHz
(51.428571s). The approved Badinerie title/lobby/Credits PCM is unchanged.

## Arrangement and boundaries

Short flute articulations and longer sustained samples share the approved
palette. Pizzicato bass and a quiet synthetic pluck continue through melodic
breathing spaces. No bowed-string pad, piano, percussion addition or external
recording. Bass register/rhythm and ornaments are explicitly adapted, not a
literal full orchestral/continuo performance.

The audition's ending fade/tail is not embedded. Sample/pluck releases fold
into the next cycle and ambience is calculated periodically. C omits its
initial standalone eighth-note pickup; its final D-sharp leads back to the
first E on a proper bar boundary. No clock correction, silence padding or
in-game crossfade is required. Rendering pins expected hashes before writing.

## Playback and settings

Fresh lobby matches cycle A→B→C→D, starting with A after launch. Each fresh match
rewinds that selection. Restart rewinds the current track without advancing;
Resume does not rewind. No gameplay RNG is consumed. Lobby/Credits use their
original player and preserve playback position. User mute/music-off and browser
interaction readiness remain independent gates. No new controls or persistence
format changes.

All four loops are RMS-matched near.06. Runtime gameplay gain stays.08; lobby.16
and effects.36. Drop/line/four-line/attack tests require at least4dB signal-RMS
margin and single-effect-plus-music peak headroom. This is not a subjective
claim about every multiplayer mix.

## Sources and licensing

`music/classical/` contains production notation, historical-source snapshots,
source/output hashes, sample provenance and audit records. The new Handel
production notation is independently read from Chrysander1879, with all146
note/rest events agreeing with the pilot. Its modern CC BY-SA pilot engraving
is not a production input, and its old license is not retroactively changed.
See `MUSIC-SOURCES.md` for the detailed, limited claims for each edition.

## Refinement decisions and acceptance

The previously discussed requirements answer the refinement checklist:
- Problem/users/outcome: local couch gameplay needs unobtrusive musical drive
  without conspicuous holes; all four auditions musically accepted.
- Scope: finished loops, deterministic playlist, credits/provenance, tests and
  deployment; no new game rules, controls, sample library or lobby changes.
- Inputs/outputs: pinned historical notation + CC0 samples → reproducible PCM.
- Platforms/performance: Go/Ebitengine native and WASM; committed PCM needs no
  runtime downloads. Four short gameplay loops replace two longer ones.
- Failures/security: source/output mismatches fail; never substitute unreviewed
  MIDI, recordings or license grants. Existing initialization cleanup retained.
- Settings/state: unchanged audio preferences/input gating; same-track Restart;
  deterministic per-launch playlist counter, no new persisted fields.
- Delivery: original auditions/old release remain archived; public packages
  retain credits and notices; increment VERSION for public deployment.
- Proof/closure: music reproducibility, source/timing/PCM/hash/seam checks,
  all-five-player gating, multi-cycle UI selection, unchanged lobby/effects,
  full build/test/package verification, browser smoke and green release CI.

## Reproduction

`make music` renders the lobby and four gameplay loops, then checks the new
music contract. `make check` and `make verify-packages` test/build/package the
normal game. Normal builds use committed WAVs and do not invoke AI, download
samples, or require private workspace files. `check-classical.ts` can rerun the
source/signal audit after rendering.

Legacy Beethoven renderers are optional historical tools and now write their
loop WAVs to `dist/audio/legacy/`, not to embedded game audio.

## Start preparation (v0.3.7)

New matches and Restart now include the brief musical preparation documented
in `count-in.md`. Music is rewound but temporarily held while boards wait;
the downbeat releases music and simulation in the same update. Normal Resume
does not introduce a new count-in. The five music WAVs and their levels remain
unchanged; one separately generated wooden-tap effect is added.

## Now-playing label (v0.3.8)

A small static `♪ Composer — work` label appears at the top right of the lobby
and match views. It identifies the selected track without opening Credits and
avoids the bottom-right touch/status areas. If music is muted, disabled or set
to 0% volume, the same selected title is shown with `· OFF`; it does not claim
audible playback. Credits remains the full source/license list. The label has
no effect on input, playlist order or audio state.

## Settings and volume (v0.3.9)

Sound, Music, session-scoped Music volume, debug tools and About/Credits live
under the lobby's Settings button; see `settings.md`. Music volume defaults to
100%, preserving the established lobby/match balance, and resets on application
start until persistence is added later.
