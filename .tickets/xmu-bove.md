---
id: xmu-bove
status: open
deps: []
links: []
created: 2026-03-29T20:37:41Z
type: feature
priority: 1
---
# xmux sidebar: vertical service monitor with popup

Extend xmux with two new Cobra subcommands to support a narrow (~4-col) right-side tmux sidebar pane showing background service health (dev server, codegen, etc.). Services are defined CLI-style in tmuxinator YAML — no separate config. The sidebar discovers services by polling a shared state directory.

## Design

## Architecture

### State directory
`~/.local/state/xmux/<session-name>/<name>.json` — status written by `xmux watch`
`~/.local/state/xmux/<session-name>/<name>.log` — full output log (appended, tee'd)

### New subcommands (Cobra)

#### `xmux watch <name> <icon> [--alert <regex>] -- <command...>`
Wraps a background process. Behavior:
1. Resolve session: `tmux display-message -p '#S'`
2. Ensure state dir exists
3. Write initial status (state=starting)
4. Run command via os/exec, pipe combined stdout+stderr
5. Per-line: tee to log + stdout, update status file, check alert regex
6. State machine: starting → running → activity (output) → running (3s no output) → alert (regex match, sticky until restart)
7. On exit: state=exited, exit_code=N

Status JSON schema:
```json
{"name": "dev", "icon": "󰎙", "state": "alert", "last_line": "...", "alert_line": "...", "pid": 123, "ts": 1234567890, "exit_code": 0}
```

#### `xmux bar`
BubbleTea TUI for the narrow sidebar. Behavior:
- Reads session name via tmux, polls `~/.local/state/xmux/<session>/` every 500ms
- Renders stacked service icons (Nerd Font) colored by state:
  - starting: dim gray
  - running: green #00d700
  - activity: yellow #ffaf00
  - alert: red #ff0000 bold
  - exited: dim gray
- Designed for 4-col pane: [space][2-col icon][space] per service, 3 rows each
- When pane is focused (receives keyboard input):
  - j/↓ and k/↑ to move selection (▶ prefix on selected icon)
  - Enter/Space → `tmux popup -E -w 80% -h 80% -T " <name>" "less +G <logfile>"`

### Cobra wiring (main.go)
Replace direct BubbleTea launch with a Cobra root command:
- `xmux` (no args) → existing session picker TUI (unchanged)
- `xmux watch` → new watch subcommand
- `xmux bar` → new bar subcommand

### File layout
```
cmd/
  watch/
    watch.go        # xmux watch subcommand
  bar/
    bar.go          # xmux bar TUI subcommand
state/
  state.go          # Status struct, Dir(), Read(), Write(), ReadAll()
main.go             # Cobra root cmd, wire subcommands
```

### New dependency
`github.com/spf13/cobra` — for subcommand routing

### Typical tmuxinator YAML usage
```yaml
windows:
  - editor: vim .
  - dev: xmux watch dev 󰎙 --alert 'error|Error|failed' -- npm run dev
  - codegen: xmux watch gen  --alert 'error' -- npm run codegen --watch
  - sidebar: xmux bar
```

## Acceptance Criteria

1. Watcher state transitions:
   - Run: xmux watch dev 󰎙 --alert 'ERROR' -- bash -c 'echo ok; sleep 2; echo ERROR bad; sleep 99'
   - Verify ~/.local/state/xmux/<session>/dev.json transitions: starting → running → activity → alert

2. Sidebar renders in 4-col pane:
   - tmux split-window -h -l 4 'xmux bar'
   - Verify icons render with correct colors, update within ~500ms

3. Keyboard navigation when focused:
   - Focus bar pane (Ctrl-b arrow), j/k selects service, Enter opens tmux popup with less log view

4. Full integration:
   - Add watch + bar to a real tmuxinator YAML
   - tmuxinator start <project>
   - Verify sidebar on right, services populate, alert state triggers on error output

5. Existing session picker (xmux with no args) is unchanged

