# Lobby music integration

Selected sound: full Badinerie B2 with the preferred low-flute sample-attack
correction. The 2026-09-16 source review and its precise scope are recorded in
`MUSIC-SOURCES.md`; the self-contained source bundle is in `music/badinerie/`.
This was introduced with v0.3.2 and its separate Credits screen. The lobby
track remains unchanged in v0.3.3; only gameplay music is replaced.

## Behavior

- `LobbyMusic`: embedded `audio/lobby-badinerie.wav`, AABB,160 beats at108 BPM,
  88.88889 seconds. This is the actual loop, not the90.69-second listening file.
- `MatchMusic`: `audio/gameplay-beethoven.wav` since v0.3.3; see
  [gameplay-music.md](gameplay-music.md). The lobby asset itself is unchanged.
- The separate Credits view keeps `LobbyMusic` playing without restarting it.
- UI selects music after processing input, so start/back changes take effect in
  the same update. Pause/results remain on the match track, as before.
- Switching scenes pauses the old player before the new one may play. No music
  overlap. Repeated lobby visits do not restart Badinerie. Since v0.3.5,
  fresh matches from the lobby alternate G1/G2 and start the selected track at
  its beginning. **Restart** rewinds that same selection without advancing the
  playlist. Normal Resume does not rewind.
  Restart while muted/music-off still resets the position, but remains silent
  until playback is permitted; the lobby track and all settings are unchanged.
- Master mute and the independent MUSIC ON/OFF setting persist across scenes.
  They never get reset by track selection. Effects are unaffected by music-off.
- Browser audio readiness is still required; scene changes cannot bypass the
  first-user-interaction requirement.
- Lobby/Credits retain `.16` playback volume. Gameplay uses `.08` to leave
  room for important effect cues; effects stay `.36`. Exported music files are
  unchanged. No new volume UI, fade system or adaptive music was added.

## Validation

- Baseline tests passed before edits.
- `xvfb-run -a make check`: Go vet/format, all tests, Linux build and WebAssembly
  build passed with the new track.
- Fake-player tests cover lobby→match→lobby, retained positions, same-view
  idempotence, exclusive playback, readiness/mute/enabled combinations, changing
  scenes before readiness, independent toggles, unknown tracks and absent audio.
- Embedded WAV test pins the selected full corrected asset's SHA-256, checks its
  exact3,920,000-frame length and tiny1/0 PCM-unit loop boundary. Separate
  gameplay tests pin the selected126BPM asset and its effect-level margin.
- UI test covers lobby/match selection and paused-match behavior.
- Chromium/SwiftShader browser smoke test used a Web Audio analyser to confirm
  actual nonzero signal in lobby and match, silence on music-off/master mute,
  no mute bypass from music toggles, and signal after returning to lobby.
  Browser audio context was running; no page errors. This is signal-level
  verification, not a subjective listening/mixing review.
- Firefox under this container could display the game with Xvfb/software GL,
  but AudioContext stayed suspended. The same failure was reproduced with an
  untouched HEAD build. Do not claim Firefox audio was verified or change game
  code to hide this environment/baseline limitation.

## Source review and reproduction

`MUSIC-SOURCES.md` distinguishes the public-domain composition from modern
reference engravings, records the independent230-note comparison and explains
why the unlicensed reference files themselves are not distributed. All12 used
CC0 samples have pinned source hashes. `make music` reproduces the approved
loop without any private workspace files; CI also runs that check.

The raw88.89-second lobby WAV uses about15MB; the106.67-second gameplay WAV
uses about19MB. Both remain the exact tested PCM and HTTP compression is used
for web delivery. A future compressed/streamed format needs its own decoding
and loop tests; lossy encoder boundaries must not be assumed seamless.
