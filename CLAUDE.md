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
- `internal/ssh` — builds `ssh` args from a `Connection` and uses `syscall.Exec` to replace the current process with the SSH binary
- `internal/tui` — Bubble Tea interactive list (fuzzy-filterable) used by `sshm list` and `sshm connect` when no alias is given
- `cmd/` — one file per Cobra subcommand (add, connect, edit, list, remove, show); each loads config, mutates it, saves it, then delegates to `internal/ssh` or `internal/tui`

**Key design decisions:**
- `Connect` uses `syscall.Exec` (not `exec.Command`) so the SSH process fully replaces `sshm` — signals, TTY, and exit codes behave as if you typed `ssh` directly
- `connect` resolves aliases: exact match → single substring match (with warning) → ambiguous error → TUI picker
- Module path in `go.mod` is `github.com/saadh393/sshm`; the `Makefile` has a stale `MODULE := github.com/sadh/sshm` that affects `-ldflags` version injection but not functionality
