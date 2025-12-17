package config

func DefaultSelection() map[string]bool {
	selected := make(map[string]bool)
	for _, pkg := range AllPackages() {
		if pkg.Required || pkg.Default {
			selected[pkgKey(pkg)] = true
		}
	}
	return selected
}

func pkgKey(pkg Package) string {
	if pkg.Type == TypeSystem {
		return pkg.Name
	}
	return pkg.Name
}
