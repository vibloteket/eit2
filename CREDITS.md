# Eit 2 — credits

## Game

**Victor Blomqvist and the Eit 2 contributors.** Based on the original Eit by
Victor Blomqvist. Project code and original contributions: AGPL-3.0-or-later;
see `LICENSE` and `NOTICE.md`.

Source: <https://github.com/vibloteket/eit2>

## Music and sound

- **Lobby composition:** Johann Sebastian Bach, *Badinerie*, Orchestral Suite
  No.2 in B minor, BWV1067/VII. Public-domain composition.
- **Match composition:** Ludwig van Beethoven, *Rondo a capriccio*, Op.129
  (*Rage Over a Lost Penny*). Public-domain composition.
- **Arrangements and rendering:** Eit 2 project, full Badinerie B2 and the
  Beethoven gameplay arrangements G1 at126BPM and G2 at112BPM. These are our generated
  performances, not recordings by an external orchestra or pianist.
- **Beethoven score:** [Mutopia Project, item498](https://www.mutopiaproject.org/cgibin/piece-info.cgi?id=498),
  typeset by **Magnus Lewis-Smith**, with a2015 LilyPond update by
  **Javier Ruiz-Alma**. The typesetter explicitly placed the transcription in
  the public domain; the credit is voluntary.
- **Flute and cello samples:** VSCO2 Community Edition, recorded by **Sam Gossner
  and Simon Dalzell**; sample cutting by **Elan Hickler / Soundemote**.
  <https://github.com/sgossner/VSCO-2-CE> — CC0 1.0 Universal.
  The credit is voluntary; retain the included license/disclaimer when
  redistributing the sample sources.
- **Sound effects and synthetic pluck:** generated specifically for Eit 2.

Full source review, historical edition, hashes, exact sample manifest, score
comparison and reproduction instructions: **[MUSIC-SOURCES.md](MUSIC-SOURCES.md)**.
No YouTube or other third-party music recording is included.

## Technology and type

- **Ebitengine:** Hajime Hoshi and contributors — Apache License 2.0.
- **Oto audio backend and related Ebitengine libraries:** their respective
  contributors — Apache License 2.0.
- **Go Regular:** The Go Authors — BSD 3-Clause.
- Additional resolved dependencies are listed below. Their complete license
  texts and copyright notices are retained under `LICENSES/dependencies/`.

## Dependency notices

Versions recorded from the module graph for this release; see `go.mod` and
`go.sum`. These licenses govern software, not the public-domain compositions
or the CC0 sample library.

| Module / source | Version | License | Retained notice |
|---|---|---|---|
| [github.com/ebitengine/hideconsole](https://pkg.go.dev/github.com/ebitengine/hideconsole@v1.0.0) | v1.0.0 | Apache-2.0 | `LICENSES/dependencies/github.com_ebitengine_hideconsole/` |
| [github.com/ebitengine/oto/v3](https://pkg.go.dev/github.com/ebitengine/oto/v3@v3.4.1) | v3.4.1 | Apache-2.0 | `LICENSES/dependencies/github.com_ebitengine_oto_v3/` |
| [github.com/ebitengine/purego](https://pkg.go.dev/github.com/ebitengine/purego@v0.9.0) | v0.9.0 | Apache-2.0 | `LICENSES/dependencies/github.com_ebitengine_purego/` |
| [github.com/go-text/typesetting](https://pkg.go.dev/github.com/go-text/typesetting@v0.3.0) | v0.3.0 | Unlicense OR BSD-3-Clause | `LICENSES/dependencies/github.com_go-text_typesetting/` |
| [github.com/hajimehoshi/ebiten/v2](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2@v2.9.10) | v2.9.10 | Apache-2.0 | `LICENSES/dependencies/github.com_hajimehoshi_ebiten_v2/` |
| [github.com/jezek/xgb](https://pkg.go.dev/github.com/jezek/xgb@v1.1.1) | v1.1.1 | BSD-3-Clause | `LICENSES/dependencies/github.com_jezek_xgb/` |
| [github.com/rivo/uniseg](https://pkg.go.dev/github.com/rivo/uniseg@v0.4.7) | v0.4.7 | MIT | `LICENSES/dependencies/github.com_rivo_uniseg/` |
| [golang.org/x/image](https://pkg.go.dev/golang.org/x/image@v0.43.0) | v0.43.0 | BSD-3-Clause | `LICENSES/dependencies/golang.org_x_image/` |
| [golang.org/x/sync](https://pkg.go.dev/golang.org/x/sync@v0.21.0) | v0.21.0 | BSD-3-Clause | `LICENSES/dependencies/golang.org_x_sync/` |
| [golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys@v0.44.0) | v0.44.0 | BSD-3-Clause | `LICENSES/dependencies/golang.org_x_sys/` |
| [golang.org/x/text](https://pkg.go.dev/golang.org/x/text@v0.38.0) | v0.38.0 | BSD-3-Clause | `LICENSES/dependencies/golang.org_x_text/` |
