# Eit 2

## Purpose
- TV-first falling-block game for 1–4 local players, focused on simultaneous couch multiplayer with a real solo practice mode.

## Stack
- Language/runtime: Go 1.26+
- Game library: Ebitengine 2.x
- Initial targets: Linux x86_64 and browser/WebAssembly
- Deferred target: Android/Android TV

## Common commands
- Install: `make install`
- Run locally: `make run`
- Test: `make test`
- Lint: `make lint`
- Build native and web: `make build build-web`
- Full verification: `make check`

## Key paths
- Entry point: `cmd/eit2/`
- Input assignment: `internal/lobby/`
- Ebitengine adapter/UI: `internal/ui/`
- Browser loader: `web/`
- Generated artifacts: `dist/`

## Notes for future agents
- Keep gameplay rules independent of Ebitengine and physical input APIs.
- Controller-first means all menus must eventually work without keyboard/mouse.
- Support 1–4 local players on one screen; solo play is a supported practice/exploration mode.
- Keep touch as a lightweight input adapter for browser-based mobile testing.
- Do not add Android build complexity until Linux and web targets are stable.
- Do not copy legacy assets without checking their licensing status.
