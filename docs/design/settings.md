# Lobby Settings (v0.3.9)

## Decided scope

The lobby stays focused on joining and starting. Its only utility entries are
**Settings** and, on native builds, **Exit**. Sound, Music, debug tools and
About/Credits live inside Settings.

Settings contains:

- **Sound**: master sound on/off.
- **Music**: music on/off without changing master sound.
- **Music volume**: session-scoped 0–100% in 10% steps. 100% is the existing
  internal mix; the user control scales lobby and match music together while
  preserving their relative balance. It does not change effects.
- **Controller Debug**: opens the existing live controller diagnostic overlay.
- **Debug Mode**: enables/disables the gameplay debug entry point.
- **About & Credits**: opens the existing Credits screen and returns to
  Settings, not directly to the lobby.
- **Back**: returns to the lobby with Settings focused.

Persistence is deliberately out of scope for this version. Sound, Music and
Music volume reset on application start. A later version can add versioned
settings storage (`localStorage` on web, a small config file native) without
changing the menu structure.

## Input behavior

- Keyboard: Up/Down select, Left/Right adjust or toggle, Enter/Space activate,
  Escape returns.
- Gamepad: D-pad/left stick navigates and adjusts, A activates, B returns.
  Unlike the lobby's join protection, any connected controller may operate this
  global settings menu.
- Pointer/touch: rows activate directly. On Music volume, the left half of the
  row decreases and the right half increases by 10%.
- Controller Debug remains available from Settings and closes with tap/key/B.
- Opening Settings/About consumes its initiating input; the same held input
  cannot immediately dismiss the destination screen.

## Runtime/audio behavior

Settings remains on the lobby music track. Music volume is applied to every
existing music player immediately. Music-off and master mute remain independent;
volume zero is treated as inaudible by the Now playing label, but does not
change the Music enabled setting. The temporary count-in music suspension still
does not change user-facing settings.

No new assets, licenses, persistence format, match rules or controller mapping
are introduced. Future button remapping should live under a Controls row in this
same Settings view.
