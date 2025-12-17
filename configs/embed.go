package configs

import "embed"

// FS provides embedded configuration templates and files.
//
//go:embed *.tmpl *.toml *.conf *.kdl
var FS embed.FS
