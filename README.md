<div align="center">

# zsm

> This fork keeps zsm working with newer zmx releases. See [About this fork](#about-this-fork) for compatibility and preview-rendering notes.

TUI session manager for [zmx](https://github.com/neurosnap/zmx)

[<img src="https://img.shields.io/github/actions/workflow/status/khangkontum/zmx-session-manager/ci.yaml?label=build&logo=github" />](https://github.com/khangkontum/zmx-session-manager/actions)
[<img src="https://img.shields.io/github/v/release/khangkontum/zmx-session-manager?label=release&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIgc3Ryb2tlPSJ3aGl0ZSIgc3Ryb2tlLXdpZHRoPSIxLjUiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgc3Ryb2tlLWxpbmVqb2luPSJyb3VuZCI%2BCiAgPHBhdGggZD0iTTIgNyBMNyAyIEgxNCBWOSBMOSAxNCBaIi8%2BCiAgPGNpcmNsZSBjeD0iMTEiIGN5PSI1IiByPSIxIi8%2BCjwvc3ZnPg%3D%3D" />](https://github.com/khangkontum/zmx-session-manager/releases/latest)
[<img src="https://img.shields.io/github/downloads/khangkontum/zmx-session-manager/total?label=downloads&logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxNiAxNiIgZmlsbD0ibm9uZSIgc3Ryb2tlPSJ3aGl0ZSIgc3Ryb2tlLXdpZHRoPSIxLjUiIHN0cm9rZS1saW5lY2FwPSJyb3VuZCIgc3Ryb2tlLWxpbmVqb2luPSJyb3VuZCI%2BCiAgPHBhdGggZD0iTTggMiBWMTAiLz4KICA8cGF0aCBkPSJNNSA3IEw4IDEwIEwxMSA3Ii8%2BCiAgPHBhdGggZD0iTTMgMTMgSDEzIi8%2BCjwvc3ZnPg%3D%3D" />](https://github.com/khangkontum/zmx-session-manager/releases)
[<img src="https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go&logoColor=white" />](https://go.dev)
[<img src="https://img.shields.io/badge/platform-macOS-lightgrey?logo=apple" />](https://github.com/mdsakalu/zmx-session-manager)
[<img src="https://img.shields.io/badge/platform-Linux-lightgrey?logo=linux&logoColor=white" />](https://github.com/mdsakalu/zmx-session-manager)
[<img src="https://img.shields.io/github/license/mdsakalu/zmx-session-manager?logo=data%3Aimage%2Fsvg%2Bxml%3Bbase64%2CPHN2ZyB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHZpZXdCb3g9IjAgMCAxNCAxNiI%2BPHBhdGggZmlsbD0id2hpdGUiIGZpbGwtcnVsZT0iZXZlbm9kZCIgZD0iTTcgNGMtLjgzIDAtMS41LS42Ny0xLjUtMS41UzYuMTcgMSA3IDFzMS41LjY3IDEuNSAxLjVTNy44MyA0IDcgNHptNyA2YzAgMS4xMS0uODkgMi0yIDJoLTFjLTEuMTEgMC0yLS44OS0yLTJsMi00aC0xYy0uNTUgMC0xLS40NS0xLTFIOHY4Yy40MiAwIDEgLjQ1IDEgMWgxYy40MiAwIDEgLjQ1IDEgMUgzYzAtLjU1LjU4LTEgMS0xaDFjMC0uNTUuNTgtMSAxLTFoLjAzTDYgNUg1YzAgLjU1LS40NSAxLTEgMUgzbDIgNGMwIDEuMTEtLjg5IDItMiAySDJjLTEuMTEgMC0yLS44OS0yLTJsMi00SDFWNWgzYzAtLjU1LjQ1LTEgMS0xaDRjLjU1IDAgMSAuNDUgMSAxaDN2MWgtMWwyIDR6TTIuNSA3TDEgMTBoM0wyLjUgN3pNMTMgMTBsLTEuNS0zLTEuNSAzaDN6Ii8%2BPC9zdmc%2B" />](LICENSE)
[<img src="https://img.shields.io/badge/Built_With-Bubble_Tea-blue" />](https://github.com/charmbracelet/bubbletea)

<img src="assets/screenshot.png" alt="zsm screenshot" width="600" />

</div>

## Install

```
go install github.com/khangkontum/zmx-session-manager/cmd/zsm@latest
```

## Requirements

[zmx](https://github.com/neurosnap/zmx) must be installed and available in your `PATH`.

## About this fork

This fork tracks compatibility with newer `zmx` releases:

- Parses the newer `zmx list` keys (`name`, `start_dir`) while preserving support for the older keys (`session_name`, `started_in`).
- Avoids double-applying `ZMX_SESSION_PREFIX` when opening preview history, attaching, or killing sessions returned by `zmx list`.
- Renders preview history from `zmx history --vt` with [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Charm ANSI](https://github.com/charmbracelet/x/tree/main/ansi) helpers so ANSI colors are preserved and preview scrolling/cropping stays terminal-width aware.

## Key Bindings

Press `?` in the app to open the keyboard shortcut help popup.

| Key | Action |
|-----|--------|
| `↑` `↓` / `j` `k` | Navigate sessions |
| `space` | Toggle selection |
| `ctrl+a` | Select / deselect all |
| `enter` | Attach to session |
| `x` | Kill selected session(s) |
| `c` | Copy attach command |
| `s` | Cycle sort mode (name / clients / newest) |
| `g` | Toggle local / global sessions |
| `/` | Filter sessions |
| `[` `]` | Scroll preview |
| `{` `}` | Jump preview to top / bottom |
| `?` | Show / hide help popup |
| `q` | Quit |

## License

[MIT](LICENSE)
