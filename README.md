# xmux

A tmux session launcher and service monitor for the terminal.

## Features

- **Session picker TUI**: fuzzy-filter and switch between active, inactive (tmuxinator), and remote (SSH) tmux sessions
- **Service monitor sidebar** (`xmux bar`): a narrow 4-column pane showing background service health with Nerd Font icons colored by state
- **Process watcher** (`xmux watch`): wraps a background process, tees output to a log file, and tracks service state (starting → running → activity → alert → exited)

## Installation

### From Source

```bash
go install github.com/mvgrimes/xmux/cmd/xmux@latest
```

Or clone and build:

```bash
git clone https://github.com/mvgrimes/xmux.git
cd xmux
go build -o xmux ./cmd/xmux
```

### Homebrew

```bash
brew install mvgrimes/tap/xmux
```

## Usage

### Session picker (default)

Run `xmux` with no arguments to open the session picker TUI:

```
xmux
```

- **Tab / Right** — cycle to the next session list (Active → Inactive → Remote)
- **Shift+Tab / Left** — cycle to the previous list
- **↑ / ↓** — move selection
- **PgUp / PgDn** — page through items
- **Type** — fuzzy-filter the current list
- **Backspace** — delete last filter character
- **Enter** — switch to / start / connect to the selected session
- **Esc / Ctrl+C** — quit

### Service monitor sidebar

Add a sidebar pane to your tmuxinator YAML:

```yaml
windows:
  - editor: vim .
  - sidebar: >
      xmux bar
      --spawn "dev 󰎙 --alert 'error|Error|failed' -- npm run dev"
      --spawn "gen --alert 'error' -- npm run codegen --watch"
```

Then split a narrow pane for the bar:

```bash
tmux split-window -h -l 4 'xmux bar'
```

#### `xmux watch <name> <icon> [--alert <regex>] -- <command...>`

Wraps a background process. Writes status JSON to
`~/.local/state/xmux/<session>/<name>.json` and a log to
`~/.local/state/xmux/<session>/<name>.log`.

State transitions: `starting` → `running` → `activity` (on output) →
`running` (3 s of silence) → `alert` (regex match, sticky).

#### `xmux bar`

BubbleTea TUI for the 4-column sidebar. Polls the state directory every 500 ms
and renders stacked service icons colored by state:

`--spawn` is repeatable and each value should be the exact args you'd pass to
`xmux watch`.

```bash
xmux bar \
  --spawn "dev 󰎙 --alert 'error|Error|failed' -- npm run dev" \
  --spawn "gen --alert 'error' -- npm run codegen --watch"
```

| State      | Color            |
|------------|------------------|
| starting   | dim gray         |
| running    | green `#00d700`  |
| activity   | yellow `#ffaf00` |
| alert      | red `#ff0000`    |
| exited     | dim gray         |

Keyboard controls when the pane is focused:

- **j / ↓** — select next service
- **k / ↑** — select previous service
- **Enter / Space** — open a tmux popup with `less +G <logfile>`
- **q / Ctrl+C** — quit

## Development

```bash
just build   # build binary
just test    # run tests
just fmt     # format code
just lint    # run linters
just install # cross-compile and install to dotfiles
```

## License

MIT
