# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build        # Build binary (./sshm)
make install      # Build and install to /usr/local/bin
make test         # Run all tests
make tidy         # go mod tidy
make release      # Cross-compile for all platforms into dist/
make clean        # Remove binary and dist/
go test ./...                    # Run all tests
go test ./internal/config/...    # Run tests for a specific package
```

## Architecture

`sshm` is a CLI SSH connection manager built with Cobra + Bubble Tea.

**Data flow:**
- Connections are stored as JSON in `~/.config/sshm/connections.json`
- `internal/config` — the data layer: `Connection` struct, Load/Save/Add/Remove/Update/FindExact/FindSubstring
- `internal/ssh` — builds `ssh` args from a `Connection` and uses `syscall.Exec` to replace the current process with the SSH binary; also `ParseCommand` to parse raw SSH commands into a `Connection`
- `internal/tui` — three Bubble Tea UIs:
  - `list.go` + `picker.go` — fuzzy-filterable connection list; `Run()` connects after selection, `RunPicker(title)` returns the selected connection for callers that decide what to do next
  - `editform.go` — interactive edit form with `textinput` bubbles, pre-filled with current values; `RunEditForm(conn)` returns an `EditResult`
  - `styles.go` — shared Lip Gloss styles and colour palette
- `cmd/` — one file per Cobra subcommand (add, connect, edit, list, remove, show); each loads config, mutates it, saves it, then delegates to `internal/ssh` or `internal/tui`

**Key design decisions:**
- `Connect` uses `syscall.Exec` (not `exec.Command`) so the SSH process fully replaces `sshm` — signals, TTY, and exit codes behave as if you typed `ssh` directly
- `connect` resolves aliases: exact match → single substring match (with warning) → ambiguous error → TUI picker
- `sshm` with no arguments opens the TUI list directly (root `RunE`); `sshm -h` shows help
- `edit` and `remove` are now alias-optional: omitting the alias opens a TUI picker first
- `edit` without flags opens `RunEditForm` (interactive); with flags it applies changes non-interactively
- Module path in `go.mod` is `github.com/saadh393/sshm`; the `Makefile` has a stale `MODULE := github.com/sadh/sshm` that affects `-ldflags` version injection but not functionality

## File Map

```
sshm/
├── main.go
├── cmd/
│   ├── root.go       # Root command — no-arg RunE shows TUI list
│   ├── add.go        # Interactive wizard (prompt-based)
│   ├── connect.go    # Direct connect by alias; doConnect() shared helper
│   ├── edit.go       # Optional alias; TUI picker → edit form or flag-based
│   ├── list.go       # Launches TUI, selects → doConnect
│   ├── remove.go     # Optional alias; TUI picker → confirmation
│   └── show.go       # Plain-text detail view
└── internal/
    ├── config/
    │   └── config.go       # Connection struct, JSON load/save, CRUD helpers
    ├── ssh/
    │   ├── ssh.go          # BuildArgs, CommandString, Connect (syscall.Exec)
    │   └── parse.go        # ParseCommand — parses raw ssh … strings
    └── tui/
        ├── list.go         # Bubble Tea list model; Run() and NewModel(title)
        ├── picker.go       # RunPicker(conns, title) — selection only, no connect
        ├── editform.go     # RunEditForm(conn) — textinput form, returns EditResult
        └── styles.go       # Shared Lip Gloss styles
```

## Adding a New Subcommand

1. Create `cmd/<name>.go` with a `var <name>Cmd` and register it in `init()` via `rootCmd.AddCommand`.
2. If the command needs a TUI picker, call `tui.RunPicker(conns, "Title...")` — it returns a `tui.Result`.
3. If the command needs SSH execution, call `doConnect(conn, dry)` (defined in `cmd/connect.go`).
4. Keep business logic in `internal/`; `cmd/` files should only wire flags → internal calls.

## Roadmap (open for contribution)

See `README.md` → **Roadmap & Open Contributions** for the full list with difficulty labels.

Priority next features:
- `sshm exec <alias> <command>` — run a one-off remote command
- `sshm tunnel <alias> --local --remote` — port-forwarding shortcut
- `sshm import` — bulk import from `~/.ssh/config`
- `sshm ping <alias>` — reachability check
