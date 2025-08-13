package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfigCreation(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Test that LoadConfig creates default config when file doesn't exist
	cfg, err := LoadConfig(configPath)
	require.NoError(t, err)
	assert.NotNil(t, cfg)

	// Verify the file was created
	_, err = os.Stat(configPath)
	assert.NoError(t, err, "Config file should be created")

	// Verify default patterns are set
	assert.NotEmpty(t, cfg.Patterns)
	
	// Verify system PATH variables come before generic patterns
	var systemPathIndex, genericPathIndex int
	for i, pr := range cfg.Patterns {
		if pr.Pattern == "CARGO_HOME" {
			systemPathIndex = i
		}
		if pr.Pattern == "*_ROOT|*_DIR|*_PATH|*_HOME" {
			genericPathIndex = i
		}
	}
	
	assert.Less(t, systemPathIndex, genericPathIndex, 
		"System-specific patterns should come before generic patterns")
}

func TestSystemPathsAreKept(t *testing.T) {
	cfg := defaultConfig()
	
	// List of system paths that should have "keep" action
	systemPaths := []string{
		"CARGO_HOME",
		"RUSTUP_HOME",
		"PNPM_HOME",
		"NIX_PATH",
		"NIX_USER_PROFILE_DIR",
		"GOPATH",
		"GOROOT",
	}
	
	for _, path := range systemPaths {
		found := false
		for _, pr := range cfg.Patterns {
			if pr.Pattern == path {
				assert.Equal(t, "keep", pr.Rule.Action,
					"%s should have 'keep' action", path)
				found = true
				break
			}
		}
		assert.True(t, found, "%s pattern should exist", path)
	}
}

func TestPatternOrder(t *testing.T) {
	cfg := defaultConfig()
	
	// Check that patterns are in the expected order
	// System-specific patterns should come first
	var foundSystemPattern, foundGenericPattern bool
	for _, pr := range cfg.Patterns {
		// Check if we found a system-specific pattern
		if pr.Pattern == "CARGO_HOME" || pr.Pattern == "PNPM_HOME" {
			foundSystemPattern = true
		}
		
		// Check if we found a generic pattern
		if pr.Pattern == "*_ROOT|*_DIR|*_PATH|*_HOME" {
			foundGenericPattern = true
			// At this point, we should have already seen system patterns
			assert.True(t, foundSystemPattern,
				"System patterns should appear before generic patterns")
		}
	}
	
	assert.True(t, foundSystemPattern, "Should have system patterns")
	assert.True(t, foundGenericPattern, "Should have generic patterns")
}