# sshm — Product Requirements Document

## What is sshm?

A terminal-based SSH connection manager written in Go. Store your servers once, connect by alias forever. No more digging through notes or `.ssh/config` trying to remember which key goes where.

---

## Tech Stack

| Choice                                            | Why                                                                                |
| ------------------------------------------------- | ---------------------------------------------------------------------------------- |
| **Go**                                            | Single static binary, no runtime deps, fast startup, easy cross-compilation        |
| **Cobra** (`spf13/cobra`)                         | Industry-standard CLI framework — subcommands, flags, help text, shell completions |
| **fatih/color**                                   | Colored terminal output (success/error feedback)                                   |
| **Bubble Tea** (`charmbracelet/bubbletea`)        | Interactive TUI framework — keyboard navigation, filtering, selection, forms       |
| **Bubbles** (`charmbracelet/bubbles`)             | Pre-built list + textinput components used by the TUI layer                        |
| **Lip Gloss** (`charmbracelet/lipgloss`)          | Terminal styling/layout — borders, colours, column widths                          |
| **JSON file** (`~/.config/sshm/connections.json`) | Simple, human-readable, easy to back up or version-control                         |

---

## Current State (v1 — implemented)

### Commands

| Command | Alias | Status | Notes |
|---|---|---|---|
| `sshm` (no args) | | ✅ Done | Opens TUI list directly |
| `sshm -h` | | ✅ Done | Shows help (Cobra built-in) |
| `sshm add` | | ✅ Done | Interactive prompt-based wizard |
| `sshm list` | `ls` | ✅ Done | Full-screen fuzzy-filterable TUI |
| `sshm connect <alias>` | `c` | ✅ Done | Exact + partial match; `--dry` flag |
| `sshm show <alias>` | | ✅ Done | Plain-text detail view |
| `sshm edit [alias]` | | ✅ Done | TUI picker if no alias; interactive form if no flags |
| `sshm remove [alias]` | `rm` | ✅ Done | TUI picker if no alias; confirmation prompt |
| `sshm version` | | ✅ Done | Prints version string |
| `sshm completion <shell>` | | ✅ Done | Cobra-generated shell completions |

### UX Highlights

- **`sshm` alone** opens the connection list — zero arguments needed for the primary workflow
- **`sshm edit`** without an alias opens a TUI picker, then a full interactive edit form (tab between fields, ctrl+s to save)
- **`sshm remove`** without an alias opens a TUI picker, then a confirmation prompt
- All TUI screens share the same keyboard model: `/` to filter, `↑↓` to navigate, `Enter` to select, `q`/`Esc` to quit

### Architecture

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

### Data Model

```json
[
  {
    "alias": "prod-api",
    "host": "10.0.1.5",
    "user": "deploy",
    "port": 22,
    "key_path": "~/.ssh/prod_rsa",
    "group": "production"
  }
]
```

Stored at: `~/.config/sshm/connections.json`
Permissions: `0600` (file), `0700` (directory)

---

## Key Design Decisions

1. **syscall.Exec over os/exec** — The connect command replaces the sshm process with ssh entirely. Terminal session behaves exactly like a native SSH session (signals, TTY, exit codes all work correctly).

2. **JSON over TOML/YAML** — Go has JSON in the standard library. No extra dependency, no parser quirks.

3. **Alias as primary key** — Case-insensitive uniqueness on alias. No numeric IDs to remember.

4. **Partial matching on connect** — `sshm c prod` auto-connects if there's only one match. Reduces friction for the most common operation.

5. **TUI picker as shared primitive** — `tui.RunPicker(conns, title)` is reused by `list`, `edit`, and `remove`. Title is configurable so context is always clear to the user.

6. **Root command opens TUI** — `sshm` alone is the entry point for the primary workflow. Help is behind `-h` which is the convention for power users who know the commands.

7. **Interactive edit form** — `tui.RunEditForm(conn)` pre-fills all fields. Flag-based edits still work for scripting. Both paths converge on the same save logic.

---

## Roadmap (open for contribution)

Features are grouped by scope. Community contributors should open an issue to claim a feature before starting work.

### Easy — good first issues

| Feature | Description | Key files to touch |
|---|---|---|
| `sshm ping <alias>` | SSH reachability check with latency | `cmd/ping.go`, `internal/ssh/ssh.go` |
| `sshm duplicate <alias> <new>` | Clone a connection under a new name | `cmd/duplicate.go`, `internal/config/config.go` |
| Last-connected timestamp | Record `last_connected` in JSON on connect; show in `sshm show` | `internal/config/config.go`, `cmd/show.go`, `internal/ssh/ssh.go` |
| `sshm search <query>` | Non-interactive table output for scripting | `cmd/search.go` |

### Medium — weekend projects

| Feature | Description | Key files to touch |
|---|---|---|
| `sshm exec <alias> <command>` | Run a one-off remote command without interactive session | `cmd/exec.go`, `internal/ssh/ssh.go` |
| `sshm tunnel <alias> --local 8080 --remote 3000` | Port-forwarding shortcut via `ssh -L` | `cmd/tunnel.go`, `internal/ssh/ssh.go` |
| `sshm copy <alias> <local> <remote>` | SCP wrapper using stored connection data | `cmd/copy.go`, `internal/ssh/scp.go` |
| `sshm import` | Bulk-import hosts from `~/.ssh/config` | `cmd/import.go`, `internal/ssh/parse.go` |
| Tags (replace single group) | `group string` → `tags []string` in data model; TUI filter already supports multi-field | `internal/config/config.go`, `internal/tui/list.go` |
| Health indicator in TUI list | Colored dot per host — parallel reachability checks before render | `internal/tui/list.go`, new `internal/ssh/ping.go` |

### Harder — larger scope

| Feature | Description | Key files to touch |
|---|---|---|
| `sshm export` to `~/.ssh/config` | Generate valid SSH config block from sshm data | `cmd/export.go` |
| `sshm batch <cmd> --group <g>` | Run a command across all connections in a group | `cmd/batch.go`, `internal/ssh/ssh.go` |
| Encryption at rest | AES-GCM encrypt `connections.json`; prompt on load/save | `internal/config/config.go`, new `internal/crypto/` |
| Homebrew tap | Package and publish formula at `saadh393/homebrew-tap` | Separate tap repository |
| `sshm backup` / `sshm restore` | Export/import `connections.json` for machine sync | `cmd/backup.go`, `cmd/restore.go` |

---

## Build & Install

```bash
# Development
go mod tidy
make build        # → ./sshm binary

# Install to PATH
sudo make install

# Cross-compile all platforms
make release      # → dist/sshm-linux-amd64, sshm-darwin-arm64, etc.
```
