package main

import (
	"fmt"
	"sort"
	"strings"

	"macsetup/internal/config"
)

type row struct {
	name        string
	typ         string
	category    string
	subCategory string
	description string
	link        string
}

func main() {
	var rows []row
	for _, pkg := range config.AllPackages() {
		switch pkg.Type {
		case config.TypeFormula:
			rows = append(rows, toRow(pkg, "formula", "https://formulae.brew.sh/formula/"+pkg.Name))
		case config.TypeCask:
			rows = append(rows, toRow(pkg, "cask", "https://formulae.brew.sh/cask/"+pkg.Name))
		}
	}

	sort.Slice(rows, func(i, j int) bool { return strings.Compare(rows[i].name, rows[j].name) < 0 })

	fmt.Println("# Packages")
	fmt.Println()
	fmt.Println("Reference list of Homebrew packages used by this repo.")
	fmt.Println()
	fmt.Println("| Package | Type | Category | Subcategory | Description | Link |")
	fmt.Println("|---|---|---|---|---|---|")
	for _, r := range rows {
		desc := strings.ReplaceAll(r.description, "|", `\|`)
		fmt.Printf("| `%s` | `%s` | `%s` | `%s` | %s | %s |\n", r.name, r.typ, r.category, r.subCategory, desc, r.link)
	}
	fmt.Println()
	fmt.Println("Generated from `internal/config/packages.go`.")
}

func toRow(pkg config.Package, typ, link string) row {
	category := pkg.Category
	if pkg.Tap != "" {
		if category == "" {
			category = "tap: " + pkg.Tap
		} else {
			category = fmt.Sprintf("%s (tap: %s)", category, pkg.Tap)
		}
	}
	return row{
		name:        pkg.Name,
		typ:         typ,
		category:    category,
		subCategory: pkg.SubCategory,
		description: pkg.Description,
		link:        link,
	}
}
