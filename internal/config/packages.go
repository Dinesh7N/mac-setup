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
		{Name: "zsh", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Zsh shell"},
		{Name: "starship", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Prompt"},
		{Name: "ripgrep", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Fast grep"},
		{Name: "fzf", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Fuzzy finder"},
		{Name: "jq", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "JSON processor"},
		{Name: "httpie", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "HTTP client"},
		{Name: "autojump", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Directory jumper"},
		{Name: "tree", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Directory tree"},
		{Name: "htop", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Process viewer"},
		{Name: "gh", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "GitHub CLI"},
		{Name: "telnet", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Network utility"},
		{Name: "ca-certificates", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Root certificates"},
		{Name: "mise", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Runtime manager"},
		{Name: "neovim", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Editor"},
		{Name: "tmux", Type: TypeFormula, Category: "shell_cli", Required: true, Default: true, Description: "Terminal multiplexer"},

		// Terminals
		{Name: "iterm2", Type: TypeCask, Category: "terminals", Default: true, Description: "Feature-rich terminal emulator"},
		{Name: "ghostty", Type: TypeCask, Category: "terminals", Default: false, Description: "Modern terminal emulator"},
		{Name: "zellij", Type: TypeFormula, Category: "terminals", Default: false, Description: "Modern tmux alternative"},

		// Editors
		{Name: "visual-studio-code", Type: TypeCask, Category: "editors", Default: true, Description: "VS Code"},
		{Name: "zed", Type: TypeCask, Category: "editors", Default: false, Description: "Zed"},
		{Name: "sublime-text", Type: TypeCask, Category: "editors", Default: false, Description: "Sublime Text"},
		{Name: "jetbrains-toolbox", Type: TypeCask, Category: "editors", Default: false, Description: "JetBrains Toolbox"},

		// Browsers
		{Name: "brave-browser", Type: TypeCask, Category: "browsers", Default: true, Description: "Brave"},
		{Name: "google-chrome", Type: TypeCask, Category: "browsers", Default: false, Description: "Google Chrome"},
		{Name: "firefox", Type: TypeCask, Category: "browsers", Default: false, Description: "Firefox"},

		// Productivity
		{Name: "raycast", Type: TypeCask, Category: "productivity", Default: true, Description: "Launcher"},
		{Name: "rectangle", Type: TypeCask, Category: "productivity", Default: true, Description: "Window manager"},

		// Dev env
		{Name: "orbstack", Type: TypeCask, Category: "dev_env", Default: false, Description: "Container runtime"},
		{Name: "postgresql@14", Type: TypeFormula, Category: "dev_env", Default: false, Description: "PostgreSQL"},
		{Name: "redis", Type: TypeFormula, Category: "dev_env", Default: false, Description: "Redis"},

		// Python
		{Name: "poetry", Type: TypeFormula, Category: "python", Default: false, Description: "Dependency management"},
		{Name: "ruff", Type: TypeFormula, Category: "python", Default: false, Description: "Linter"},
		{Name: "ty", Type: TypeFormula, Category: "python", Default: false, Description: "Type checker"},

		// Go
		{Name: "gofumpt", Type: TypeFormula, Category: "go", Default: false, Description: "Formatter"},
		{Name: "golangci-lint", Type: TypeFormula, Category: "go", Default: false, Description: "Linter"},
		{Name: "golang-migrate", Type: TypeFormula, Category: "go", Default: false, Description: "Database migrations"},
		{Name: "go-task", Type: TypeFormula, Category: "go", Default: false, Description: "Task runner"},

		// Node.js / TypeScript
		{Name: "biome", Type: TypeFormula, Category: "nodejs", Default: false, Description: "Formatter / linter"},
		{Name: "bun", Type: TypeFormula, Category: "nodejs", Default: false, Description: "bun runtime", Tap: "oven-sh/bun"},
		{Name: "deno", Type: TypeFormula, Category: "nodejs", Default: false, Description: "Deno runtime"},
		{Name: "eslint", Type: TypeFormula, Category: "nodejs", Default: false, Description: "ESLint"},
		{Name: "prettier", Type: TypeFormula, Category: "nodejs", Default: false, Description: "Prettier"},

		// DevOps
		{Name: "hashicorp/tap", Type: TypeTap, Category: "devops", Default: false, Description: "Homebrew tap for HashiCorp tools", Tap: "hashicorp/tap"},
		{Name: "opentofu", Type: TypeFormula, Category: "devops", Default: false, Description: "IaC tool"},
		{Name: "terraform", Type: TypeFormula, Category: "devops", Default: false, Description: "Terraform", Tap: "hashicorp/tap"},
		{Name: "tfenv", Type: TypeFormula, Category: "devops", Default: false, Description: "Terraform version manager"},
		{Name: "ansible", Type: TypeFormula, Category: "devops", Default: false, Description: "Automation"},
		{Name: "awscli", Type: TypeFormula, Category: "devops", Default: false, Description: "AWS CLI"},
		{Name: "gitleaks", Type: TypeFormula, Category: "devops", Default: false, Description: "Secrets scanner"},
		{Name: "checkov", Type: TypeFormula, Category: "devops", Default: false, Description: "IaC scanner"},
		{Name: "trivy", Type: TypeFormula, Category: "devops", Default: false, Description: "Vulnerability scanner"},
		{Name: "tfsec", Type: TypeFormula, Category: "devops", Default: false, Description: "Terraform scanner"},
		{Name: "opa", Type: TypeFormula, Category: "devops", Default: false, Description: "Open Policy Agent"},

		// AI
		{Name: "gemini-cli", Type: TypeFormula, Category: "ai", Default: false, Description: "Gemini CLI"},
		{Name: "claude-code", Type: TypeCask, Category: "ai", Default: false, Description: "Claude Code"},
		{Name: "codex", Type: TypeCask, Category: "ai", Default: false, Description: "Codex"},
		{Name: "chatgpt", Type: TypeCask, Category: "ai", Default: false, Description: "ChatGPT"},
		{Name: "claude", Type: TypeCask, Category: "ai", Default: false, Description: "Claude"},

		// Communication
		{Name: "microsoft-teams", Type: TypeCask, Category: "communication", Default: false, Description: "Microsoft Teams"},

		// Optional
		{Name: "1password", Type: TypeCask, Category: "optional", Default: false, Description: "1Password"},
		{Name: "1password-cli", Type: TypeCask, Category: "optional", Default: false, Description: "1Password CLI"},
		{Name: "spotify", Type: TypeCask, Category: "optional", Default: false, Description: "Spotify"},

		// Fonts (default ON, no individual selection in the TUI)
		{Name: "font-sf-mono", Type: TypeCask, Category: "fonts", Default: true, Description: "SF Mono font"},
		{Name: "font-sf-pro", Type: TypeCask, Category: "fonts", Default: true, Description: "SF Pro font"},
	}
}
