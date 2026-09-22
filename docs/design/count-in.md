# Musical count-in (v0.3.7)

## Approved behavior

User approved three gentle wooden preparation clicks and a simple drawn baton,
then a clear downstroke that starts gameplay and match music together. Apply
on a new match and Restart, not on ordinary Resume. Keep it short (~1.6 seconds
at the gameplay music's nominal112 quarter BPM). No spoken numbers, new music,
full conductor character or new setting is needed.

## Presentation and timing

The boards and first pieces are prepared immediately but do not accept gameplay
commands or tick before GO. A shared paper card shows READY, a baton tapping a
small stand and three filled/outlined markers. The downstroke changes READY to
GO briefly; the overlay then clears. It is visible equally in solo and couch
layouts, including when muted. Small geometry is drawn directly in the existing
paper/ink palette; there are no imported images.

Timing uses active monotonic time, not a fixed number of gameplay updates. This
matters on low-TPS software/phone renderers. Each next beat is anchored to the
beat actually presented, so a late frame cannot squeeze the following click or
GO into an immediate catch-up update. Normal display quantization adds at most
one update per beat (about1.65s total at60UPS). An extreme stall preserves beat
order instead of bursting or skipping clicks. These are UI scheduling guarantees,
not sample-accurate promises about every operating system/audio backend.

Pause/disconnect freezes the preparation and animation. Resume during an
unfinished preparation continues it; Resume after GO does not create a new one.
Returning to the lobby cancels the preparation. Restart resets it and keeps the
same music selection; only fresh matches advance A–B–C–D.

## Audio and input

A transient `musicSuspended` gate pauses music during preparation without changing
Mute or Music-enabled. GO releases that gate in the same update that ticks the
new match and starts its elapsed-time clock. Lobby playback position is retained.
The three taps are ordinary effects: Music-off still allows taps, master Mute
silences them. Missing/unready audio does not prevent the visual start.

`count-in.wav` is generated with the existing oscillator/noise renderer and is
55ms long. No external recording or licensing dependency is introduced. All
five music files and the original14 effects remain byte-identical.

Gameplay input is consumed during preparation, including the GO update. Held
soft-drop keys, gamepad controls and touch contacts from preparation must release
before they can act; this is per key/player/contact, not a global blockade. Fresh
input afterward works normally. Pause/menu controls remain usable. Menu-triggered
Restart/Back returns immediately rather than continuing to process the replaced
match in the same update.

## Refinement / scope decisions

- Goal/users/success: shared, readable start for1–4 local players; no hidden early
  movement, input buffering, music start or match-time advancement.
- Scope/surfaces: UI clock/drawing/input gate, temporary music gate, one procedural
  effect, tests and notices. No gameplay rules, RNG behavior, playlist order,
  existing music/SFX bytes, persistent settings or new controls changed.
- Platforms: native and WASM, keyboard/gamepad/touch/mouse, muted/no-audio paths.
- Error/security behavior: visual timing remains usable without audio; normal
  source/hash checks remain; no remote services or recordings needed at runtime.
- Data/delivery: transient state only; new effect and generator hash audited;
  increment VERSION and publish only after standard checks/CI and live smoke.

## Verification

- Controlled-time tests at5/10/30/60/120UPS, late frames, zero-delta catch-up,
  freeze of all1–4 boards, GO/clock timing, pause/disconnect/restart/cancel and
  per-player input-release behavior.
- All four music tracks × readiness/mute/music-enabled combinations verify the
  presentation gate never changes preferences or lobby position.
- Browser verification uses an audio-thread tap to count short pulses reliably
  even when compositor screenshots stall the main thread. It checks three
  distinct taps, input rejection, Music-off/master Mute, post-GO music, Resume,
  Restart, pause and cancel. AudioContext timing in headless software graphics
  is not treated as an exact wall-clock metronome measurement.
- Existing full browser smoke waits for preparation to finish before testing
  gameplay music/effects. Unit tests separately verify tempo/scheduling.
- Full test/race/build/package flow, unchanged music checks and byte comparison
  of the original14 effects; no subjective assistant listening claim.
