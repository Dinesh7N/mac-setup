package config

type Category struct {
	Key         string
	Name        string
	Description string
	Required    bool
	Selectable  bool
}

func Categories() []Category {
	return []Category{
		{Key: "installed", Name: "Already Installed", Description: "Packages already present on the system", Required: false, Selectable: true},
		{Key: "core", Name: "Core", Description: "Required system prerequisites", Required: true, Selectable: false},
		{Key: "shell_cli", Name: "Shell & CLI", Description: "Essential shell and CLI tools", Required: true, Selectable: false},
		{Key: "terminals", Name: "Terminals", Description: "Terminal emulators", Required: false, Selectable: true},
		{Key: "editors", Name: "Editors", Description: "Code editors and IDEs", Required: false, Selectable: true},
		{Key: "browsers", Name: "Browsers", Description: "Web browsers", Required: false, Selectable: true},
		{Key: "productivity", Name: "Productivity", Description: "Productivity tools", Required: false, Selectable: true},
		{Key: "dev_env", Name: "Dev Environment", Description: "Local dev services & containers", Required: false, Selectable: true},
		{Key: "python", Name: "Python Tooling", Description: "Python development tools", Required: false, Selectable: true},
		{Key: "go", Name: "Go Tooling", Description: "Go development tools", Required: false, Selectable: true},
		{Key: "nodejs", Name: "Node.js / TypeScript", Description: "JS/TS development tools", Required: false, Selectable: true},
		{Key: "devops", Name: "DevOps", Description: "Infrastructure and DevOps tools", Required: false, Selectable: true},
		{Key: "ai", Name: "AI Tools", Description: "AI assistants and tools", Required: false, Selectable: true},
		{Key: "communication", Name: "Communication", Description: "Communication tools", Required: false, Selectable: true},
		{Key: "optional", Name: "Optional Apps", Description: "Optional applications", Required: false, Selectable: true},
		{Key: "fonts", Name: "Fonts", Description: "Developer fonts", Required: false, Selectable: false},
	}
}
