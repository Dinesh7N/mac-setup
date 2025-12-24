# Mac Setup CLI - Implementation Specification

> A TUI-based macOS setup tool for team onboarding. Built with Go, Bubble Tea, and Cobra.

## Table of Contents

1. [Overview](#overview)
2. [Idempotency](#idempotency-running-on-existing-machines)
3. [Target Environment](#target-environment)
4. [Tech Stack](#tech-stack)
5. [Project Structure](#project-structure)
6. [Package Taxonomy](#package-taxonomy)
7. [TUI Design](#tui-design)
8. [Data Structures](#data-structures)
9. [Installation Flow](#installation-flow)
10. [Component Specifications](#component-specifications)
11. [Embedded Config Templates](#embedded-config-templates)
12. [Bootstrap Script](#bootstrap-script)
13. [Build & Release](#build--release)
14. [Error Handling](#error-handling)
15. [Testing Strategy](#testing-strategy)

---

## Overview

### Purpose

Replace a monolithic bash script with an interactive Go CLI that helps new team members set up their MacBooks quickly. The tool provides a checkbox-based interface to select software categories and individual packages.

### Key Features

- Interactive TUI with category/package selection
- Parallel formula installation for speed
- Progress tracking with real-time feedback
- Error resilience (continue on failure, report at end)
- Dotfile management with automatic backups
- Single-command bootstrap via curl
- **Idempotent** - Safe to run on already-configured machines
- **Initial Scan** - Detects and groups already installed packages separately

---

## Idempotency (Running on Existing Machines)

The tool must be safe to run on machines that already have some or all tools installed. This is critical because:

1. Team members may run it to add new tools later
2. Someone may re-run after a partial failure
3. Allows "topping up" an existing setup

### Behavior Rules

| Component | If Already Exists | Action |
|-----------|-------------------|--------|
| **Homebrew** | Installed | Skip install, run `brew update` |
| **Xcode CLI Tools** | Installed | Skip entirely |
| **Brew Formulas** | Package installed | Skip (no reinstall), unless selected explicitly |
| **Brew Casks** | App installed | Skip (no reinstall), unless selected explicitly |
| **Oh My Zsh** | `~/.oh-my-zsh` exists | Skip install |
| **Zsh Plugins** | Plugin dir exists | Skip clone |
| **Neovim Config** | `~/.config/nvim` exists | **Skip** (don't overwrite user's config) |
| **TPM (tmux plugins)** | Dir exists | Skip clone |
| **Dotfiles** (.zshrc, etc.) | File exists | **Backup first**, then overwrite |
| **Config Directories** | Dir exists | No-op (mkdir -p is safe) |
| **Mise Runtimes** | Version installed | Skip |

### Implementation Details

```go
// internal/installer/homebrew.go

func InstallPackage(pkg Package) InstallResult {
    // Check if already installed BEFORE attempting install
    if IsPackageInstalled(pkg) {
        return InstallResult{
            Package: pkg,
            Status:  StatusSkipped,
            Message: "Already installed",
        }
    }
    
    // Proceed with installation
    err := brewInstall(pkg)
    if err != nil {
        return InstallResult{
            Package: pkg,
            Status:  StatusFailed,
            Error:   err.Error(),
        }
    }
    
    return InstallResult{
        Package: pkg,
        Status:  StatusInstalled,
    }
}
```

### TUI Display for Existing Installs

In the progress screen, show different icons for each status:

```
┌─────────────────────────────────────────────────────────────────┐
│  Installing packages...                                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ✓ Homebrew (already installed)                                 │
│  ✓ ripgrep (already installed)                                  │
│  ✓ fzf (already installed)                                      │
│  + jq (installed)                                               │
│  ⠋ Installing starship...                                       │
│  ○ httpie                                                       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Legend:**
- `✓` (dim) = Already installed, skipped
- `+` (green) = Newly installed
- `✗` (red) = Failed
- `⠋` (spinner) = In progress
- `○` = Pending

### Summary Screen with Skip Count

```
┌─────────────────────────────────────────────────────────────────┐
│  ✓ Setup Complete!                                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  Newly Installed:  12 packages                                  │
│  Already Present:  35 packages (skipped)                        │
│  Failed:           0 packages                                   │
│                                                                 │
│  Dotfiles:                                                      │
│    ✓ ~/.zshrc (backed up existing → ~/.zshrc.bak.20240115)      │
│    ✓ ~/.config/starship/starship.toml (created)                 │
│    ○ ~/.config/nvim (skipped - already exists)                  │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Special Cases

#### Neovim Config
**Do NOT overwrite** if `~/.config/nvim` exists. Users may have their own config.
- If missing: Clone kickstart.nvim
- If exists: Skip and note in summary

#### Dotfiles (.zshrc, starship.toml, etc.)
**Always backup and overwrite** - these are managed by this tool.
- Create timestamped backup: `~/.zshrc.bak.20240115_143022`
- Write new config
- Show backup path in summary

#### User Prompt Option (Future Enhancement)
Consider adding a `--force` flag or interactive prompt:
```
~/.config/nvim already exists. 
[s]kip  [b]ackup and replace  [q]uit
```

---

## Target Environment

| Requirement | Value |
|-------------|-------|
| OS | macOS only |
| Architecture | Apple Silicon (arm64) only |
| Shell | Zsh (macOS default) |
| Privileges | Requires sudo for some installations |

---

## Tech Stack

| Component | Library | Version | Purpose |
|-----------|---------|---------|---------|
| CLI Framework | `github.com/spf13/cobra` | v1.8+ | Command structure |
| TUI Framework | `github.com/charmbracelet/bubbletea` | v0.25+ | Interactive UI |
| TUI Components | `github.com/charmbracelet/bubbles` | v0.18+ | List, spinner, progress |
| Styling | `github.com/charmbracelet/lipgloss` | v0.10+ | Colors, borders, layout |
| Embedded Files | `embed` (stdlib) | - | Bundle config templates |

---

## Project Structure

```
mac-setup/
├── main.go                          # Entry point
├── go.mod
├── go.sum
├── Taskfile.yaml                    # Build commands (replaces Makefile)
├── .goreleaser.yaml                 # Release automation
├── install.sh                       # Bootstrap script
├── README.md                        # User documentation
├── IMPLEMENTATION_SPEC.md           # This file
│
├── cmd/
│   └── root.go                      # Cobra root command
│
├── internal/
│   ├── config/
│   │   ├── packages.go              # All package definitions
│   │   └── categories.go            # Category groupings & defaults
│   │
│   ├── installer/
│   │   ├── homebrew.go              # Homebrew operations
│   │   ├── xcode.go                 # Xcode CLI tools
│   │   ├── mise.go                  # Mise + language runtimes
│   │   ├── ohmyzsh.go               # Oh My Zsh + plugins
│   │   ├── dotfiles.go              # Config writer with backup
│   │   ├── git.go                   # Git clone operations
│   │   ├── scan.go                  # Installed package scanning
│   │   ├── verify.go                # Post-install verification
│   │   └── directories.go           # Create config directories
│   │
│   ├── tui/
│   │   ├── app.go                   # Main app, state machine
│   │   ├── styles.go                # Lip Gloss styles
│   │   ├── welcome.go               # Welcome screen
│   │   ├── selection.go             # Package selection screen
│   │   ├── progress.go              # Installation progress
│   │   └── summary.go               # Final summary
│   │
│   ├── constants/
│   │   └── constants.go             # Global constants
│   │
│   └── utils/
│       ├── exec.go                  # Command execution
│       ├── filesystem.go            # File operations
│       ├── retry.go                 # Retry mechanisms
│       └── system.go                # OS checks, sudo
│
└── configs/                         # Embedded via go:embed
    ├── zshrc.tmpl
    ├── starship.toml
    ├── alacritty.toml
    ├── ghostty.conf
    ├── tmux.conf
    └── zellij.kdl                   # Optional: for zellij users
```

---

## Package Taxonomy

### Category Overview

```
┌────────────────────────────────────────────────────────────────────┐
│                         PACKAGE CATEGORIES                          │
├────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ALREADY INSTALLED (Detected automatically)                         │
│  └── List of packages found on system (optional re-selection)       │
│                                                                     │
│  CORE (Required - No opt-out, auto-installed first)                 │
│  └── Homebrew, Xcode CLI Tools                                      │
│                                                                     │
│  SHELL & CLI (Required - No opt-out)                                │
│  ├── Shell: Oh My Zsh + plugins, Starship, Zsh                      │
│  └── Utils: ripgrep, fzf, jq, httpie, autojump, tree, htop,         │
│             gh, telnet, ca-certificates, mise, neovim, tmux         │
│                                                                     │
│  TERMINALS (Default: ON, can customize)                             │
│  ├── Ghostty (default ON)                                           │
│  ├── iTerm2 (default OFF)                                           │
│  └── Zellij (default OFF) - Modern tmux alternative                 │
│                                                                     │
│  EDITORS (Default: ON, can customize)                               │
│  ├── VS Code (default ON)                                           │
│  ├── Zed (default OFF)                                              │
│  ├── Sublime Text (default OFF)                                     │
│  └── JetBrains Toolbox (default OFF)                                │
│                                                                     │
│  BROWSERS (Default: ON, can customize)                              │
│  ├── Brave (default ON)                                             │
│  ├── Google Chrome (default OFF)                                    │
│  └── Firefox (default OFF)                                          │
│                                                                     │
│  PRODUCTIVITY (Default: ON, can customize)                          │
│  ├── Raycast (default ON)                                           │
│  └── Rectangle (default ON)                                         │
│                                                                     │
│  DEV ENVIRONMENT (Default: OFF, selectable)                         │
│  ├── OrbStack (default OFF)                                         │
│  ├── PostgreSQL (default OFF)                                       │
│  └── Redis (default OFF)                                            │
│                                                                     │
│  PYTHON TOOLING (Default: OFF, selectable)                          │
│  ├── poetry                                                         │
│  ├── ruff                                                           │
│  └── ty                                                             │
│                                                                     │
│  GO TOOLING (Default: OFF, selectable)                              │
│  ├── gofumpt                                                        │
│  ├── golangci-lint                                                  │
│  ├── golang-migrate                                                 │
│  └── go-task                                                        │
│                                                                     │
│  NODE.JS / TYPESCRIPT TOOLING (Default: OFF, selectable)            │
│  ├── biome                                                          │
│  ├── bun                                                            │
│  ├── deno                                                           │
│  ├── eslint                                                         │
│  └── prettier                                                       │
│                                                                     │
│  DEVOPS (Default: OFF, selectable as group)                         │
│  ├── IaC: opentofu, terraform, tfenv, ansible                       │
│  ├── Cloud: awscli                                                  │
│  └── Security: gitleaks, checkov, trivy, tfsec, opa                 │
│                                                                     │
│  AI TOOLS (Default: OFF, selectable)                                │
│  ├── CLI: gemini-cli, claude-code, codex                            │
│  └── Apps: ChatGPT, Claude                                          │
│                                                                     │
│  OPTIONAL APPS (Default: OFF, selectable)                           │
│  ├── 1Password + 1Password CLI                                      │
│  └── Spotify                                                        │
│                                                                     │
└────────────────────────────────────────────────────────────────────┘
```

### Complete Package List

#### Core (No opt-out)

| Package | Type | Tap | Notes |
|---------|------|-----|-------|
| Xcode CLI Tools | system | - | Via `xcode-select --install` |
| Homebrew | system | - | Via official install script |

#### Shell & CLI (No opt-out)

| Package | Type | Tap | Notes |
|---------|------|-----|-------|
| Zsh | formula | - | Shell (required) |
| Oh My Zsh | script | - | Via curl installer |
| zsh-autosuggestions | git | - | Oh My Zsh plugin |
| zsh-autocomplete | git | - | Oh My Zsh plugin |
| zsh-syntax-highlighting | git | - | Oh My Zsh plugin |
| zsh-completions | git | - | Oh My Zsh plugin |
| starship | formula | - | Prompt |
| ripgrep | formula | - | Fast grep |
| fzf | formula | - | Fuzzy finder |
| jq | formula | - | JSON processor |
| httpie | formula | - | HTTP client |
| autojump | formula | - | Directory jumper |
| tree | formula | - | Directory tree |
| htop | formula | - | Process viewer |
| gh | formula | - | GitHub CLI |
| telnet | formula | - | Network utility |
| ca-certificates | formula | - | Root certs |
| mise | formula | - | Runtime version manager |
| neovim | formula | - | Editor (aliased to vim) |
| tmux | formula | - | Terminal multiplexer |

#### Terminals

| Package | Type | Default | Notes |
|---------|------|---------|-------|
| ghostty | cask | ON | Modern GPU-accelerated terminal emulator |
| iterm2 | cask | OFF | Feature-rich, classic choice |
| zellij | formula | OFF | Modern tmux alternative |

#### Editors

| Package | Type | Default |
|---------|------|---------|
| visual-studio-code | cask | ON |
| zed | cask | OFF |
| sublime-text | cask | OFF |
| jetbrains-toolbox | cask | OFF |

#### Browsers

| Package | Type | Default |
|---------|------|---------|
| brave-browser | cask | ON |
| google-chrome | cask | OFF |
| firefox | cask | OFF |

#### Productivity

| Package | Type | Default |
|---------|------|---------|
| raycast | cask | ON |
| rectangle | cask | ON |

#### Dev Environment

| Package | Type | Default |
|---------|------|---------|
| orbstack | cask | OFF |
| postgresql@14 | formula | OFF |
| redis | formula | OFF |

#### Python Tooling

| Package | Type | Default |
|---------|------|---------|
| poetry | formula | OFF |
| ruff | formula | OFF |
| ty | formula | OFF |

#### Go Tooling

| Package | Type | Default |
|---------|------|---------|
| gofumpt | formula | OFF |
| golangci-lint | formula | OFF |
| golang-migrate | formula | OFF |
| go-task | formula | OFF |

#### Node.js / TypeScript Tooling

| Package | Type | Tap | Default |
|---------|------|-----|---------|
| biome | formula | - | OFF |
| bun | formula | oven-sh/bun | OFF |
| deno | formula | - | OFF |
| eslint | formula | - | OFF |
| prettier | formula | - | OFF |

#### DevOps

| Package | Type | Tap | Default |
|---------|------|-----|---------|
| opentofu | formula | - | OFF |
| terraform | formula | hashicorp/tap | OFF |
| tfenv | formula | - | OFF |
| ansible | formula | - | OFF |
| awscli | formula | - | OFF |
| gitleaks | formula | - | OFF |
| checkov | formula | - | OFF |
| trivy | formula | - | OFF |
| tfsec | formula | - | OFF |
| opa | formula | - | OFF |

#### AI Tools

| Package | Type | Default |
|---------|------|---------|
| gemini-cli | formula | OFF |
| claude-code | cask | OFF |
| codex | cask | OFF |
| chatgpt | cask | OFF |
| claude | cask | OFF |

#### Optional Apps

| Package | Type | Default |
|---------|------|---------|
| 1password | cask | OFF |
| 1password-cli | cask | OFF |
| spotify | cask | OFF |

---

## TUI Design

### Screen Flow

```
┌─────────────┐
│   Welcome   │
│   Screen    │
└──────┬──────┘
       │ [Enter]
       ▼
┌─────────────┐
│  Scanning   │ ◄── New Phase: Detect installed packages
│   System    │
└──────┬──────┘
       │ (Complete)
       ▼
┌─────────────┐
│  Package    │
│  Selection  │ ◄── Collapsible categories, One Dark theme
└──────┬──────┘
       │ [Enter] (Confirm)
       ▼
┌─────────────┐
│  Xcode      │
│  Install    │
└──────┬──────┘
       │ (Complete)
       ▼
┌─────────────┐
│  Progress   │
│   Screen    │
└──────┬──────┘
       │ (Complete)
       ▼
┌─────────────┐
│   Summary   │
│   Screen    │
└─────────────┘
```

### Screen 2: Package Selection (Updated)

```
┌─────────────────────────────────────────────────────────────────┐
│  Select packages to install                                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  [+] ALREADY INSTALLED (12)                                     │
│                                                                 │
│  [+] CORE (2) (required)                                        │
│                                                                 │
│  [-] TERMINALS                                      [2 selected]│
│      [x] Ghostty                                                │
│      [ ] iTerm2                                                 │
│      [ ] Zellij                                                 │
│                                                                 │
│  [+] EDITORS                                        [1 selected]│
│                                                                 │
│  [+] BROWSERS                                       [1 selected]│
│                                                                 │
│  ──────────────────────────────────────────────────────────────│
│  [↑↓] Navigate  [Space] Toggle  [a] Select section  [Enter] Go  │
│  [n] Deselect section  [q] Quit                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Key Interaction Changes:**
- **Already Installed**: Items found on the system are moved to this category by default.
- **Selection**: Unselected "Installed" items show `[✓]`. Selected "Installed" items (forced reinstall) move back to their original category with `[x] (reinstall)` status.
- **Collapsible**: Categories start collapsed (except typically the first focused one, or all collapsed). Toggle expansion with `Space` on the header.
- **Theme**: One Dark Pro Flat colors (Blue/Green/Red/Grey).

---

## Data Structures

### Category Definition

```go
// internal/config/categories.go

func Categories() []Category {
    return []Category{
        {Key: "installed", Name: "Already Installed", Description: "Packages already present on the system", Required: false, Selectable: true},
        // ... other categories
    }
}
```

---

## Installation Flow

### Phase 0: Scanning

```go
// internal/installer/scan.go

func ScanInstalledPackages(ctx context.Context, packages []config.Package) (map[string]bool, error) {
    // 1. Bulk scan Homebrew formulas and casks
    // 2. Map installed items
    // 3. Return map[pkgName]bool
}
```

### Phase 1: Pre-flight Checks (Updated)

Includes dependency validation (`git`, `curl`, `sudo`).

### Phase 5: Post-Install Setup (Updated)

- Zsh Plugins now installed to `~/.oh-my-zsh/custom/plugins`.
- Includes `zsh-completions`.

---

## Component Specifications

### Oh My Zsh Installer

```go
// internal/installer/ohmyzsh.go

var zshPlugins = map[string]string{
    "zsh-autosuggestions":     "https://github.com/zsh-users/zsh-autosuggestions.git",
    "zsh-autocomplete":        "https://github.com/marlonrichert/zsh-autocomplete.git",
    "zsh-syntax-highlighting": "https://github.com/zsh-users/zsh-syntax-highlighting.git",
    "zsh-completions":         "https://github.com/zsh-users/zsh-completions.git",
}

// Installs to: ~/.oh-my-zsh/custom/plugins/
```

---

## Bootstrap Script

### install.sh

```bash
#!/bin/bash
set -e
# ... (standard checks)
# Download binary from GitHub Releases
# Execute binary
```

---

## Build & Release

### Taskfile (Replaces Makefile)

```yaml
version: '3'

vars:
  BINARY_NAME: macsetup

tasks:
  default:
    cmds:
      - task: build

  build:
    desc: Build the binary
    cmds:
      - go build -ldflags "-s -w -X main.version={{.VERSION}}" -o bin/{{.BINARY_NAME}} .

  run:
    cmds:
      - go run . -- {{.CLI_ARGS}}

  test:
    cmds:
      - go test ./...
```

---