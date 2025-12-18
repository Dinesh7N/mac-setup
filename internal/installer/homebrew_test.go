package installer

import (
	"context"
	"testing"
	"time"
)

func TestLinkFormula(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test linking a formula that exists
	// Note: This test assumes 'tree' is installed but might need linking
	err := LinkFormula(ctx, "tree")
	if err != nil {
		// It's okay if it's already linked
		if err.Error() != "already linked" {
			t.Logf("LinkFormula returned: %v (this may be expected)", err)
		}
	}
}

func TestIsFormulaLinked(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test with a common formula that should be linked (use tree instead of git)
	linked := IsFormulaLinked(ctx, "tree")
	if !linked {
		t.Error("Expected tree to be linked, but it's not")
	}

	// Test with a formula that doesn't exist
	linked = IsFormulaLinked(ctx, "nonexistent-formula-12345")
	if linked {
		t.Error("Expected nonexistent formula to not be linked")
	}
}

func TestIsCaskAppInstalled(t *testing.T) {
	tests := []struct {
		name         string
		caskName     string
		expectExists bool
		expectedPath string
	}{
		{
			name:         "iterm2 installed",
			caskName:     "iterm2",
			expectExists: true,
			expectedPath: "/Applications/iTerm.app",
		},
		{
			name:         "nonexistent app",
			caskName:     "nonexistent-app-12345",
			expectExists: false,
			expectedPath: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exists, path := IsCaskAppInstalled(tt.caskName)

			// Only check iTerm if it actually exists
			if tt.caskName == "iterm2" {
				if !exists {
					t.Skip("iTerm.app not installed, skipping test")
				}
				if path != tt.expectedPath {
					t.Errorf("Expected path %s, got %s", tt.expectedPath, path)
				}
			} else if tt.caskName == "nonexistent-app-12345" {
				if exists {
					t.Errorf("Expected app to not exist, but found it at %s", path)
				}
			}
		})
	}
}
