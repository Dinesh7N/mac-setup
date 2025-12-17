package installer

import (
	"context"
	"fmt"

	"macsetup/internal/utils"
)

type MiseRuntime struct {
	Name    string
	Version string
}

var defaultRuntimes = []MiseRuntime{
	{Name: "node", Version: "latest"},
	{Name: "python", Version: "latest"},
	{Name: "go", Version: "latest"},
}

func SetupMise(ctx context.Context) error {
	for _, rt := range defaultRuntimes {
		_, err := utils.Run(ctx, 0, "mise", "use", "--global", fmt.Sprintf("%s@%s", rt.Name, rt.Version))
		if err != nil {
			return fmt.Errorf("failed to install %s: %w", rt.Name, err)
		}
	}
	return nil
}
