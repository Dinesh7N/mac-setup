package installer

import (
	"macsetup/internal/config"
	"macsetup/internal/utils"
)

func classifyInstallError(pkg config.Package, err error) string {
	if err == nil {
		return ""
	}
	ie := utils.ClassifyError(pkg.Name, err, err.Error())
	return ie.Error()
}
