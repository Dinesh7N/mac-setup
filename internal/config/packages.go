package config

type PackageType string

const (
	TypeSystem  PackageType = "system"
	TypeTap     PackageType = "tap"
	TypeFormula PackageType = "formula"
	TypeCask    PackageType = "cask"
	TypeTask    PackageType = "task"
)

type Package struct {
	Name        string
	Type        PackageType
	Category    string
	SubCategory string
	Required    bool
	Default     bool
	Description string
	Tap         string
}

func AllPackages() []Package {
	return []Package{
		{Name: "Xcode CLI Tools", Type: TypeSystem, Category: "core", Required: true, Default: true, Description: "Command line developer tools"},
		{Name: "Homebrew", Type: TypeSystem, Category: "core", Required: true, Default: true, Description: "Package manager"},

		// Shell & CLI (required)
		{Name: "zsh", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "UNIX shell (command interpreter)"},
		{Name: "starship", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Cross-shell prompt for astronauts"},
		{Name: "grep", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "GNU grep, egrep and fgrep"},
		{Name: "ripgrep", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Search tool like grep and The Silver Searcher"},
		{Name: "fzf", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Command-line fuzzy finder written in Go"},
		{Name: "jq", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Lightweight and flexible command-line JSON processor"},
		{Name: "httpie", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "User-friendly cURL replacement (command-line HTTP client)"},
		{Name: "autojump", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Shell extension to jump to frequently used directories"},
		{Name: "tree", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Display directories as trees (with optional color/HTML output)"},
		{Name: "htop", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Improved top (interactive process viewer)"},
		{Name: "gh", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "GitHub command-line tool"},
		{Name: "telnet", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "User interface to the TELNET protocol"},
		{Name: "ca-certificates", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Mozilla CA certificate store"},
		{Name: "neovim", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Ambitious Vim-fork focused on extensibility and agility"},
		{Name: "tmux", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Terminal multiplexer"},
		{Name: "zellij", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Pluggable terminal workspace, with terminal multiplexer as the base feature"},

		// Terminals
		{Name: "iterm2", Type: TypeCask, Category: "terminals", Default: true, Description: "Terminal emulator as alternative to Apple's Terminal app"},
		{Name: "ghostty", Type: TypeCask, Category: "terminals", Default: false, Description: "Terminal emulator that uses platform-native UI and GPU acceleration"},

		// Editors
		{Name: "visual-studio-code", Type: TypeCask, Category: "editors", Default: false, Description: "Open-source code editor"},
		{Name: "zed", Type: TypeCask, Category: "editors", Default: false, Description: "Multiplayer code editor"},
		{Name: "sublime-text", Type: TypeCask, Category: "editors", Default: false, Description: "Text editor for code, markup and prose"},
		{Name: "jetbrains-toolbox", Type: TypeCask, Category: "editors", Default: false, Description: "JetBrains tools manager"},

		// Browsers
		{Name: "brave-browser", Type: TypeCask, Category: "browsers", Default: false, Description: "Web browser focusing on privacy"},
		{Name: "google-chrome", Type: TypeCask, Category: "browsers", Default: false, Description: "Web browser"},
		{Name: "firefox", Type: TypeCask, Category: "browsers", Default: false, Description: "Web browser"},

		// Productivity
		{Name: "raycast", Type: TypeCask, Category: "productivity", Default: true, Description: "Control your tools with a few keystrokes"},
		{Name: "rectangle", Type: TypeCask, Category: "productivity", Default: true, Description: "Move and resize windows using keyboard shortcuts or snap areas"},

		// Dev env
		{Name: "orbstack", Type: TypeCask, Category: "dev_env", Default: false, Description: "Replacement for Docker Desktop"},
		{Name: "postgresql@17", Type: TypeFormula, Category: "dev_env", Default: false, Description: "Object-relational database system"},
		{Name: "redis", Type: TypeFormula, Category: "dev_env", Default: false, Description: "Persistent key-value database, with built-in net interface"},
		{Name: "amazon-workspaces", Type: TypeCask, Category: "dev_env", Default: false, Description: "Cloud native persistent desktop virtualization"},
		{Name: "postman", Type: TypeCask, Category: "dev_env", Default: false, Description: "Collaboration platform for API development"},
		{Name: "bruno", Type: TypeCask, Category: "dev_env", Default: false, Description: "Open source IDE for exploring and testing APIs"},

		// Programming Utilities
		// General
		{Name: "mise", Type: TypeFormula, Category: "programming", SubCategory: "others", Required: true, Default: true, Description: "Polyglot runtime manager (asdf rust clone)"},

		// Python
		{Name: "poetry", Type: TypeFormula, Category: "programming", SubCategory: "python", Default: false, Description: "Python package management tool"},
		{Name: "ruff", Type: TypeFormula, Category: "programming", SubCategory: "python", Default: false, Description: "Extremely fast Python linter, written in Rust"},
		{Name: "ty", Type: TypeFormula, Category: "programming", SubCategory: "python", Default: false, Description: "Extremely fast Python type checker, written in Rust"},
		{Name: "black", Type: TypeFormula, Category: "programming", SubCategory: "python", Default: false, Description: "Python code formatter"},
		{Name: "pydantic", Type: TypeFormula, Category: "programming", SubCategory: "python", Default: false, Description: "Data validation using Python type hints"},

		// Go
		{Name: "gofumpt", Type: TypeFormula, Category: "programming", SubCategory: "go", Default: false, Description: "Stricter gofmt"},
		{Name: "golangci-lint", Type: TypeFormula, Category: "programming", SubCategory: "go", Default: false, Description: "Fast linters runner for Go"},
		{Name: "golang-migrate", Type: TypeFormula, Category: "programming", SubCategory: "go", Default: false, Description: "Database migrations CLI tool"},
		{Name: "go-task", Type: TypeFormula, Category: "programming", SubCategory: "go", Default: false, Description: "Task runner/build tool that aims to be simpler and easier to use"},

		// Node.js / TypeScript
		{Name: "biome", Type: TypeFormula, Category: "programming", SubCategory: "nodejs", Default: false, Description: "Toolchain of the web"},
		{Name: "bun", Type: TypeFormula, Category: "programming", SubCategory: "nodejs", Default: false, Description: "Incredibly fast JavaScript runtime, bundler, transpiler and package manager", Tap: "oven-sh/bun"},
		{Name: "deno", Type: TypeFormula, Category: "programming", SubCategory: "nodejs", Default: false, Description: "Secure runtime for JavaScript and TypeScript"},
		{Name: "eslint", Type: TypeFormula, Category: "programming", SubCategory: "nodejs", Default: false, Description: "AST-based pattern checker for JavaScript"},
		{Name: "prettier", Type: TypeFormula, Category: "programming", SubCategory: "nodejs", Default: false, Description: "Code formatter for JavaScript, CSS, JSON, GraphQL, Markdown, YAML"},

		// DevOps
		{Name: "opentofu", Type: TypeFormula, Category: "devops", Default: false, Description: "Drop-in replacement for Terraform. Infrastructure as Code Tool"},
		{Name: "terraform", Type: TypeFormula, Category: "devops", Default: false, Description: "Tool to build, change, and version infrastructure", Tap: "hashicorp/tap"},
		{Name: "tfenv", Type: TypeFormula, Category: "devops", Default: false, Description: "Terraform version manager inspired by rbenv"},
		{Name: "ansible", Type: TypeFormula, Category: "devops", Default: false, Description: "Automate deployment, configuration, and upgrading"},
		{Name: "awscli", Type: TypeFormula, Category: "devops", Default: false, Description: "Official Amazon AWS command-line interface"},
		{Name: "gitleaks", Type: TypeFormula, Category: "devops", Default: false, Description: "Audit git repos for secrets"},
		{Name: "checkov", Type: TypeFormula, Category: "devops", Default: false, Description: "Prevent cloud misconfigurations during build-time for IaC tools"},
		{Name: "trivy", Type: TypeFormula, Category: "devops", Default: false, Description: "Vulnerability scanner for container images, file systems, and Git repos"},
		{Name: "tfsec", Type: TypeFormula, Category: "devops", Default: false, Description: "Static analysis security scanner for your terraform code"},
		{Name: "opa", Type: TypeFormula, Category: "devops", Default: false, Description: "Open source, general-purpose policy engine"},

		// AI
		{Name: "gemini-cli", Type: TypeFormula, Category: "ai", Default: false, Description: "Interact with Google Gemini AI models from the command-line"},
		{Name: "claude-code", Type: TypeCask, Category: "ai", Default: false, Description: "Terminal-based AI coding assistant"},
		{Name: "codex", Type: TypeCask, Category: "ai", Default: false, Description: "OpenAI's coding agent that runs in your terminal"},
		{Name: "chatgpt", Type: TypeCask, Category: "ai", Default: false, Description: "OpenAI's official ChatGPT desktop app"},
		{Name: "claude", Type: TypeCask, Category: "ai", Default: false, Description: "Anthropic's official Claude AI desktop app"},

		// Optional
		{Name: "1password", Type: TypeCask, Category: "optional", Default: false, Description: "Password manager that keeps all passwords secure behind one password"},
		{Name: "1password-cli", Type: TypeCask, Category: "optional", Default: false, Description: "Command-line interface for 1Password"},
		{Name: "spotify", Type: TypeCask, Category: "optional", Default: false, Description: "Music streaming service"},
	}
}
