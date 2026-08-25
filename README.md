# Eit 2

Eit 2 is a TV-first couch multiplayer falling-block game for two to four local
players. It is a clean successor to the original
[`vibloteket/eit`](https://github.com/vibloteket/eit), built in Go with
[Ebitengine](https://ebitengine.org/).

## Current milestone

This initial technical spike proves the project structure and its first targets:

- native Linux build;
- browser/WebAssembly build;
- deterministic tests for assigning up to four input devices;
- a 16:9 lobby showing four player slots;
- keyboard and standard gamepad join input.

There is no gameplay yet. Android is deliberately deferred; native Linux and
web are the active targets.

## Develop

Requires Go 1.26 or newer.

```sh
make install
make check
make run
```

In the lobby, press the standard south/A button on a gamepad to join. Enter or
Space joins one keyboard player. Browsers only expose connected gamepads after
the user has interacted with them.

## Build outputs

```sh
make build       # dist/native/eit2
make build-web   # dist/web/
```

Serve the web directory over HTTP; browsers do not normally allow the Wasm
loader to run directly from `file://`.

## Architecture

- `internal/lobby`: engine-independent local device assignment
- `internal/ui`: Ebitengine presentation and physical input adapter
- `cmd/eit2`: native and WebAssembly entry point
- `web`: browser loader

Future gameplay rules will live in an engine-independent `internal/game`
package so they can be tested without opening a window.
