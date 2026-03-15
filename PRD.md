# sshm — Project Plan

## What is sshm?

A terminal-based SSH connection manager written in Go. Store your servers once, connect by alias forever. No more digging through notes or `.ssh/config` trying to remember which key goes where.

---

## Tech Stack

| Choice                                            | Why                                                                                |
| ------------------------------------------------- | ---------------------------------------------------------------------------------- |
| **Go**                                            | Single static binary, no runtime deps, fast startup, easy cross-compilation        |
| **Cobra** (`spf13/cobra`)                         | Industry-standard CLI framework — subcommands, flags, help text, shell completions |
| **fatih/color**                                   | Colored terminal output (success/error feedback)                                   |
| **Bubble Tea** (`charmbracelet/bubbletea`)        | Interactive TUI framework — keyboard navigation, filtering, selection              |
| **Bubbles** (`charmbracelet/bubbles/list`)        | Pre-built list component with fuzzy search, arrow-key nav, and item selection      |
| **Lip Gloss** (`charmbracelet/lipgloss`)          | Terminal styling/layout (used by Bubbles internally, also useful for custom views) |
| **JSON file** (`~/.config/sshm/connections.json`) | Simple, human-readable, easy to back up or version-control                         |

---

## Commands (v1 Scope)

### `sshm add <alias>`

Save a new connection.

```
sshm add prod-api --host 10.0.1.5 --user deploy --key ~/.ssh/prod_rsa --port 22 --group production
```

- `--host` / `-H` (required) — server IP or hostname
- `--user` / `-u` (required) — SSH username
- `--port` / `-p` — SSH port, defaults to 22
- `--key` / `-k` — path to private key (validated on add)
- `--group` / `-g` — optional label for grouping

### `sshm connect <alias>` (shorthand: `c`)

Open an SSH session. Uses `syscall.Exec` to replace the process with a real SSH session (not a subprocess).

```
sshm connect prod-api
sshm c prod              # partial match — auto-connects if unique
sshm connect prod --dry  # prints the ssh command without executing
```

- If no exact match, does a substring search
- Single match → auto-connect with a warning
- Multiple matches → lists them and asks you to be specific

### `sshm list` (shorthand: `ls`)

Interactive TUI — browse, filter, and connect from one screen.

```
sshm list
sshm list --group production
```

Opens a full-screen interactive list powered by Bubble Tea:

```
┌─ sshm ─────────────────────────────────────────┐
│                                                 │
│  Filter: prod_                                  │
│                                                 │
│  > prod-api        deploy@10.0.1.5   production │
│    prod-db         admin@10.0.2.10   production │
│    prod-worker     deploy@10.0.3.1   production │
│                                                 │
│  ↑/↓ navigate  / filter  enter connect  q quit  │
└─────────────────────────────────────────────────┘
```

**Keyboard controls:**

- `↑` / `↓` or `j` / `k` — navigate items
- `/` — start typing to filter (fuzzy match on alias, host, user, group)
- `Enter` — connect to the highlighted server immediately
- `q` / `Esc` — quit without connecting

**Why this matters:** This is the primary workflow. You type `sshm ls`, scan the list visually, arrow to your server, hit Enter. One command, zero memorization.

### `sshm show <alias>`

Full detail view for one connection, including the raw SSH command it would run.

### `sshm edit <alias>`

Partial update — only the flags you pass get changed, everything else stays.

```
sshm edit prod-api --host 10.0.1.99
sshm edit prod-api --rename api-server
```

### `sshm remove <alias>` (shorthand: `rm`)

Delete with confirmation prompt. Pass `-y` to skip.

```
sshm remove old-server
sshm rm old-server -y
```

### `sshm version`

Prints the version string.

---

## Project Structure

```
sshm/
├── main.go                     # Entry point — just calls cmd.Execute()
├── go.mod
├── Makefile
├── README.md
├── .gitignore
├── cmd/
│   ├── root.go                 # Cobra root command + subcommand registration
│   ├── add.go                  # sshm add
│   ├── connect.go              # sshm connect
│   ├── list.go                 # sshm list (launches TUI)
│   ├── show.go                 # sshm show
│   ├── edit.go                 # sshm edit
│   └── remove.go               # sshm remove
└── internal/
    ├── config/
    │   └── config.go           # Load/Save/Add/Remove/Edit/Search on JSON store
    ├── ssh/
    │   └── ssh.go              # Build ssh args + syscall.Exec to connect
    └── tui/
        ├── list.go             # Bubble Tea model for the interactive server list
        └── styles.go           # Lip Gloss styles (colors, borders, layout)
```

---

## Data Model

Each connection is a flat struct stored as JSON:

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

1. **syscall.Exec over os/exec** — The connect command replaces the sshm process with ssh entirely. This means the terminal session behaves exactly like a normal SSH session (signals, TTY, etc. all work correctly). No wrapper overhead.

2. **JSON over TOML/YAML** — Go has JSON in the standard library. No extra dependency, no parser quirks. The file is small enough that readability differences are negligible.

3. **Alias as primary key** — Case-insensitive uniqueness on alias. No numeric IDs to remember.

4. **Partial matching on connect** — Typing `sshm c prod` should just work if there's only one match. Reduces friction for the most common operation.

5. **Cobra for CLI** — Gets you `--help`, flag parsing, subcommand routing, and shell completion generation for free.

6. **Bubble Tea for interactive list** — The `charmbracelet` stack is the de facto standard for Go TUIs (used by GitHub CLI, Charm tools, etc.). The `bubbles/list` component gives you keyboard nav, fuzzy filtering, and custom item rendering out of the box. When the user presses Enter, the TUI exits cleanly and returns the selected alias, which is then passed to `ssh.Connect()` via `syscall.Exec`.

---

## Build & Install

```bash
# Development
go mod tidy
make build        # → ./sshm binary

# Install to PATH
sudo mv sshm /usr/local/bin/

# Cross-compile all platforms
make release      # → dist/sshm-linux-amd64, sshm-darwin-arm64, etc.
```

---

## Implementation Order

| Phase     | What                                                                                    | Estimated Time |
| --------- | --------------------------------------------------------------------------------------- | -------------- |
| 1         | Scaffold: `go mod init`, Cobra root, Makefile                                           | 15 min         |
| 2         | `internal/config` — Load, Save, Add, Get, Remove                                        | 30 min         |
| 3         | `cmd/add` + basic `cmd/list` (non-interactive, for testing)                             | 20 min         |
| 4         | `internal/tui` — Bubble Tea list model, Lip Gloss styles, fuzzy filter, Enter-to-select | 45 min         |
| 5         | Wire `cmd/list` → TUI → `ssh.Connect()` (select a server, hit Enter, you're in)         | 20 min         |
| 6         | `internal/ssh` + `cmd/connect` — direct connect by alias                                | 20 min         |
| 7         | `cmd/show`, `cmd/edit`, `cmd/remove` — remaining CRUD                                   | 30 min         |
| 8         | Polish: error messages, edge cases, README                                              | 20 min         |
| **Total** |                                                                                         | **~3 hours**   |

---

## Future Ideas (post-v1)

- **Import from `~/.ssh/config`** — parse existing hosts automatically
- **SSH config export** — generate an `~/.ssh/config` block from sshm data
- **Tags instead of single group** — more flexible filtering
- **`sshm copy <alias> <local> <remote>`** — SCP wrapper
- **`sshm tunnel <alias> --local 8080 --remote 3000`** — port forwarding shortcut
- **Encryption at rest** — optional passphrase on the JSON file
- **Shell completions** — Cobra can auto-generate bash/zsh/fish completions
