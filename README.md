# mac-setup

> An interactive CLI tool to bootstrap and configure macOS (Apple Silicon) for development.

![License](https://img.shields.io/github/license/Dinesh7N/mac-setup)
![Go Version](https://img.shields.io/github/go-mod/go-version/Dinesh7N/mac-setup)

A TUI-based onboarding tool for setting up fresh MacBooks. It installs Homebrew, development tools (Zsh, Neovim, Tmux, Git), applications (Terminals, Editors, Browsers), and manages dotfiles idempotently.

## Features

*   **Interactive TUI**: Select exactly what you want to install.
*   **Idempotent**: Safe to run multiple times; detects installed apps and backups existing configs.
*   **Smart Detection**: Identifies already installed packages and groups them separately.
*   **Parallel Installation**: Fast installation using concurrent workers.
*   **Dotfile Management**: Automatically configures Zsh, Starship, Neovim (Kickstart), and Tmux/Zellij.

## Quick Start (One-Liner)

To set up a fresh Mac, simply run this command in your terminal:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Dinesh7N/mac-setup/main/install.sh)"
```

This script will:
1.  Check for Apple Silicon (M1/M2/M3).
2.  Download the latest release binary.
3.  Launch the setup tool.

## Installation Categories

*   **Core**: Homebrew, Xcode CLI Tools (Required).
*   **Shell**: Zsh, Oh My Zsh, Starship, plugins (autosuggestions, syntax-highlighting).
*   **Terminals**: Ghostty, iTerm2, Zellij.
*   **Editors**: VS Code, Zed, Sublime Text, Neovim.
*   **Dev Tools**: Docker (OrbStack), Postgres, Redis, Python/Go/Node.js tooling.
*   **Apps**: Browsers, Productivity tools, Fonts.

## Manual Usage

### Build from Source

Requirements: Go 1.21+

```bash
# Clone the repo
git clone https://github.com/Dinesh7N/mac-setup.git
cd mac-setup

# Install task runner (optional, or use go build)
go install github.com/go-task/task/v3/cmd/task@latest

# Build
task build

# Run
`./bin/macsetup` or `task run`
```

### CLI Options

```bash
# Run in headless mode (installs default packages without TUI)
./bin/macsetup --headless

# Dry run (simulate actions without changes)
./bin/macsetup --dry-run

# Increase verbosity
./bin/macsetup --verbose
```

## Contributing

Pull requests are welcome! Please ensure you run tests before submitting.

```bash
task test
```

## License

MIT