package config

import "testing"

func TestAllPackagesHaveCategory(t *testing.T) {
	categories := make(map[string]bool)
	for _, c := range Categories() {
		categories[c.Key] = true
	}

	for _, pkg := range AllPackages() {
		if !categories[pkg.Category] {
			t.Fatalf("package %q has unknown category: %q", pkg.Name, pkg.Category)
		}
	}
}

func TestNoDuplicatePackages(t *testing.T) {
	seen := make(map[string]bool)
	for _, pkg := range AllPackages() {
		key := string(pkg.Type) + ":" + pkg.Name
		if seen[key] {
			t.Fatalf("duplicate package: %s", key)
		}
		seen[key] = true
	}
}
