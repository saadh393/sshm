# sshm

<p align="center">
  <img src="https://img.shields.io/github/v/release/saadh393/sshm?style=flat-square&color=7C3AED&label=latest" alt="Latest Release">
  <img src="https://img.shields.io/github/actions/workflow/status/saadh393/sshm/release.yml?style=flat-square&color=10B981&label=build" alt="Build Status">
  <img src="https://img.shields.io/github/go-mod/go-version/saadh393/sshm?style=flat-square&color=00ADD8&label=go" alt="Go Version">
  <img src="https://img.shields.io/github/license/saadh393/sshm?style=flat-square&color=F59E0B" alt="License">
  <img src="https://img.shields.io/github/downloads/saadh393/sshm/total?style=flat-square&color=06B6D4&label=downloads" alt="Downloads">
  <img src="https://img.shields.io/github/issues/saadh393/sshm?style=flat-square&color=EF4444&label=issues" alt="Open Issues">
  <img src="https://img.shields.io/github/contributors/saadh393/sshm?style=flat-square&color=8B5CF6&label=contributors" alt="Contributors">
</p>

<p align="center">
  A fast, terminal-based SSH connection manager — store your servers once, connect by alias forever.
</p>

```
┌─ sshm ─────────────────────────────────────────────┐
│                                                     │
│  Filter: prod_                                      │
│                                                     │
│  > prod-api      deploy@10.0.1.5     production     │
│    prod-db       admin@10.0.2.10     production     │
│    prod-worker   deploy@10.0.3.1     production     │
│                                                     │
│  ↑/↓ navigate   / filter   enter connect   q quit  │
└─────────────────────────────────────────────────────┘
```

<p align="center">
  No config files to hand-craft. No flags to memorize.<br>
  Type <code>sshm</code>, arrow to your server, press Enter.
</p>

---

## Install

**macOS / Linux — one-liner**

```bash
curl -fsSL https://raw.githubusercontent.com/saadh393/sshm/main/install.sh | bash
```

**Homebrew** *(coming soon)*

```bash
brew install saadh393/tap/sshm
```

**Download binary directly**

Grab the binary for your platform from the [Releases](https://github.com/saadh393/sshm/releases) page.

| Platform | Binary |
|---|---|
| macOS (Apple Silicon) | `sshm_*_darwin_arm64.tar.gz` |
| macOS (Intel) | `sshm_*_darwin_amd64.tar.gz` |
| Linux (x86_64) | `sshm_*_linux_amd64.tar.gz` |
| Linux (ARM64) | `sshm_*_linux_arm64.tar.gz` |
| Windows (x86_64) | `sshm_*_windows_amd64.zip` |

Extract and move the binary to somewhere in your `$PATH`:

```bash
tar -xzf sshm_*_darwin_arm64.tar.gz
sudo mv sshm /usr/local/bin/
```

**Build from source** *(requires Go 1.21+)*

```bash
git clone https://github.com/saadh393/sshm
cd sshm
sudo make install
```

**Shell completions** *(optional but recommended)*

```bash
# Zsh
mkdir -p ~/.zsh/completions
sshm completion zsh > ~/.zsh/completions/_sshm
# Add to ~/.zshrc: fpath=(~/.zsh/completions $fpath) && autoload -Uz compinit && compinit

# Bash
sshm completion bash > /etc/bash_completion.d/sshm

# Fish
sshm completion fish > ~/.config/fish/completions/sshm.fish
```

---

## Usage

### Browse and connect — the main workflow

```bash
sshm
```

Running `sshm` with no arguments opens the interactive TUI immediately. Arrow to your server, press Enter.

```bash
sshm list       # same thing, explicit
sshm ls         # alias
```

| Key | Action |
|---|---|
| `↑` `↓` or `j` `k` | Navigate |
| `/` | Filter (fuzzy match on alias, host, user, group) |
| `Enter` | Connect |
| `q` `Esc` | Quit |

### Add a connection

```bash
sshm add
```

Paste your SSH command and answer a few short prompts — sshm parses the user, host, port, and key automatically.

```
  Add New SSH Connection
  ─────────────────────────────────────

  Alias (name for this connection): prod-api

  Paste your SSH command (e.g. ssh -i ~/.ssh/id_ed25519 ubuntu@1.2.3.4)
  SSH command: ssh -i ~/.ssh/id_ed25519 ubuntu@18.136.130.144

  Parsed:
    Host : 18.136.130.144
    User : ubuntu
    Key  : ~/.ssh/id_ed25519

  Port [22]:
  Group / tag (optional): production

  ─────────────────────────────────────
  Preview
  ─────────────────────────────────────
    Alias   : prod-api
    Host    : 18.136.130.144
    User    : ubuntu
    Port    : 22
    Key     : ~/.ssh/id_ed25519
    Group   : production

    Command : ssh -i ~/.ssh/id_ed25519 ubuntu@18.136.130.144
  ─────────────────────────────────────

  Save this connection? [Y/n]:
  ✓ Connection "prod-api" saved.
```

### Connect directly by alias

```bash
sshm connect prod-api      # exact match
sshm c prod                # partial match — auto-connects if unique
sshm connect prod --dry    # print the ssh command without running it
```

### Edit a connection

```bash
sshm edit                  # TUI picker → interactive edit form
sshm edit prod-api         # jump straight to the edit form for prod-api
sshm edit prod-api --host 10.0.1.99          # flag-based, no form
sshm edit prod-api --port 2222 --key ~/.ssh/new_key
sshm edit prod-api --rename api-server
```

Without flags, `sshm edit` opens an interactive form pre-filled with the current values:

```
┌─ Edit — prod-api ──────────────────────────────────┐
│                                                     │
│  Alias (rename)    prod-api                         │
│  Host            ▌ 18.136.130.144                   │
│  User              ubuntu                           │
│  Port              22                               │
│  Key path          ~/.ssh/id_ed25519                │
│  Group             production                       │
│                                                     │
│  tab/↑↓ navigate  ctrl+s save  enter next  esc quit │
└─────────────────────────────────────────────────────┘
```

### Remove a connection

```bash
sshm remove                # TUI picker → confirmation prompt
sshm rm prod-api           # jump straight to confirmation for prod-api
sshm rm prod-api -y        # skip confirmation
```

### Show connection details

```bash
sshm show prod-api
```

```
Connection Details
  Alias        prod-api
  Host         18.136.130.144
  User         ubuntu
  Port         22
  Key          ~/.ssh/id_ed25519
  Group        production

SSH Command:   ssh -i ~/.ssh/id_ed25519 ubuntu@18.136.130.144
```

### Save and run per-connection remote commands

```bash
sshm command add prod-api restart-nginx "sudo systemctl restart nginx"
sshm command add prod-api tail-nginx "tail -f /var/log/nginx/access.log"
sshm command list prod-api
sshm command run prod-api restart-nginx
sshm command run prod-api tail-nginx --dry
sshm command update prod-api tail-nginx "tail -n 200 /var/log/nginx/access.log"
sshm command delete prod-api restart-nginx
```

### All commands

| Command | Alias | Description |
|---|---|---|
| `sshm` | | Open interactive TUI — browse and connect |
| `sshm add` | | Interactive wizard to add a connection |
| `sshm list` | `ls` | Same as `sshm` — browse and connect via TUI |
| `sshm connect <alias>` | `c` | Connect directly by alias (partial match supported) |
| `sshm show <alias>` | | Show full connection details |
| `sshm command` | `cmd` | Manage saved remote commands per connection |
| `sshm edit [alias]` | | Edit a connection — TUI picker if no alias given |
| `sshm remove [alias]` | `rm` | Remove a connection — TUI picker if no alias given |
| `sshm version` | | Print version |
| `sshm completion <shell>` | | Generate shell completion script |
| `sshm -h` | | Show available commands |

---

## How it works

- Connections are stored as plain JSON at `~/.config/sshm/connections.json`
- `sshm connect` uses `syscall.Exec` to replace itself with the SSH process — the session is fully native, not a subprocess
- File permissions: `0600` (file), `0700` (directory)

```json
[
  {
    "alias": "prod-api",
    "host": "18.136.130.144",
    "user": "ubuntu",
    "port": 22,
    "key_path": "~/.ssh/id_ed25519",
    "group": "production",
    "commands": {
      "restart-nginx": "sudo systemctl restart nginx",
      "tail-nginx": "tail -f /var/log/nginx/access.log"
    }
  }
]
```

The file is yours — back it up, sync it across machines, or version-control it however you like.

---

## Building from source

```bash
make build          # build ./sshm
sudo make install   # install to /usr/local/bin
make release        # cross-compile and package release archives → dist/
make test           # run tests
sudo make uninstall # remove from /usr/local/bin
```

---

## Contributing

Contributions are welcome and appreciated! Here's how to get involved:

**Report a bug or request a feature**

Open an [issue](https://github.com/saadh393/sshm/issues) and describe what you found or what you'd like to see. Please check existing issues first to avoid duplicates.

**Submit a pull request**

1. Fork the repository
2. Create a branch — `git checkout -b feat/my-feature` or `fix/my-bug`
3. Make your changes and add tests if relevant
4. Run `make test` to make sure everything passes
5. Open a pull request with a clear description of the change

---

## Roadmap & Open Contributions

The items below are planned features open for community contribution. Each is self-contained and labelled by difficulty. Pick one, open an issue to claim it, and send a PR.

### Easy — good first issues

| Feature | Description |
|---|---|
| `sshm ping <alias>` | Run `ssh -o ConnectTimeout=3 ... exit` to check host reachability. Print latency. |
| `sshm duplicate <alias> <new-alias>` | Clone a connection under a new name for easy editing. |
| Last-connected timestamp | Record `last_connected` in the JSON on every `connect`. Show in `sshm show`. |
| `sshm search <query>` | Non-interactive grep across alias/host/user/group, prints a table. Good for scripting. |

### Medium — weekend projects

| Feature | Description |
|---|---|
| `sshm exec <alias> <command>` | Run a one-off command on a remote host without an interactive session. e.g. `sshm exec prod-api "df -h"` |
| `sshm tunnel <alias> --local 8080 --remote 3000` | Port-forwarding shortcut. Runs `ssh -L 8080:localhost:3000 ...` using stored connection data. |
| `sshm copy <alias> <local> <remote>` | SCP wrapper. Use stored connection to transfer files. Support both directions. |
| `sshm import` from `~/.ssh/config` | Parse existing SSH config and bulk-add connections to sshm. Fast onboarding path. |
| Tags (replace single group with multiple) | Change `group string` → `tags []string` in data model. TUI filter already supports it. |
| Health indicator in TUI list | Show a colored dot (green/red) per host in `sshm list`. Run reachability checks in parallel before rendering. |

### Harder — larger scope

| Feature | Description |
|---|---|
| `sshm export` to `~/.ssh/config` | Generate a valid `~/.ssh/config` block from sshm data. Useful for VS Code Remote, Ansible, etc. |
| `sshm batch <command> --group <group>` | Run a command across all connections in a group. Output prefixed by alias. |
| Encryption at rest | AES-GCM encrypt `connections.json` with a passphrase. Prompt on load/save. |
| Homebrew tap | Package and publish a Homebrew formula at `saadh393/homebrew-tap`. |
| `sshm backup` / `sshm restore` | Export/import `connections.json` for syncing between machines. |

If you're unsure whether your idea fits, open a discussion first — happy to talk it through.

---

## License

MIT © [saadh393](https://github.com/saadh393)
