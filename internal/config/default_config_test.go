package config

import (
	"os"
	"path/filepath"
	"strings"
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
	var firstSystemPathIndex = -1
	var genericPathIndex = -1
	for i, pr := range cfg.Patterns {
		// Check for any system paths pattern
		if containsSystemPath(pr.Pattern) && firstSystemPathIndex == -1 {
			firstSystemPathIndex = i
		}
		if pr.Pattern == "*_ROOT | *_DIR | *_PATH | *_HOME" {
			genericPathIndex = i
		}
	}
	
	assert.NotEqual(t, -1, firstSystemPathIndex, "Should have system path patterns")
	assert.NotEqual(t, -1, genericPathIndex, "Should have generic pattern")
	assert.Less(t, firstSystemPathIndex, genericPathIndex, 
		"System-specific patterns should come before generic patterns")
}

func containsSystemPath(pattern string) bool {
	// Check if the pattern contains any of the system paths
	systemPaths := []string{"CARGO_HOME", "RUSTUP_HOME", "PNPM_HOME", "NIX_PATH", "DBT_PROFILES_DIR", "GHOSTTY_RESOURCES_DIR", "TMP_DIR"}
	for _, path := range systemPaths {
		if strings.Contains(pattern, path) {
			return true
		}
	}
	return false
}

func TestSystemPathsAreKept(t *testing.T) {
	cfg := defaultConfig()
	
	// List of system paths that should have "keep" action
	systemPaths := []string{
		"CARGO_HOME",
		"RUSTUP_HOME",
		"PNPM_HOME",
		"DIRENV_DIR",
		"NIX_PATH",
		"NIX_USER_PROFILE_DIR",
		"GOPATH",
		"GOROOT",
	}
	
	// Check that each system path is in at least one pattern with "keep" action
	for _, path := range systemPaths {
		found := false
		for _, pr := range cfg.Patterns {
			if strings.Contains(pr.Pattern, path) {
				assert.Equal(t, "keep", pr.Rule.Action,
					"%s should have 'keep' action", path)
				found = true
				break
			}
		}
		assert.True(t, found, "%s should be in a pattern", path)
	}
}

func TestPatternOrder(t *testing.T) {
	cfg := defaultConfig()
	
	// Check that patterns are in the expected order
	// System-specific patterns should come first
	var foundSystemPattern, foundGenericPattern bool
	for _, pr := range cfg.Patterns {
		// Check if we found the consolidated system-specific pattern
		if containsSystemPath(pr.Pattern) {
			foundSystemPattern = true
		}
		
		// Check if we found a generic pattern
		if pr.Pattern == "*_ROOT | *_DIR | *_PATH | *_HOME" {
			foundGenericPattern = true
			// At this point, we should have already seen system patterns
			assert.True(t, foundSystemPattern,
				"System patterns should appear before generic patterns")
		}
	}
	
	assert.True(t, foundSystemPattern, "Should have system patterns")
	assert.True(t, foundGenericPattern, "Should have generic patterns")
}

func TestApplicationDirectoriesAreKept(t *testing.T) {
	cfg := defaultConfig()

	// List of application-specific directories that should have "keep" action
	// We test a representative subset to ensure the functionality works
	applicationDirs := []string{
		"DBT_PROFILES_DIR",
		"TMP_DIR",
		"GHOSTTY_RESOURCES_DIR",
		"FOUNDRY_DIR",
	}

	// Check that each application directory is in at least one pattern with "keep" action
	for _, dir := range applicationDirs {
		found := false
		for _, pr := range cfg.Patterns {
			if strings.Contains(pr.Pattern, dir) {
				assert.Equal(t, "keep", pr.Rule.Action,
					"%s should have 'keep' action", dir)
				found = true
				break
			}
		}
		assert.True(t, found, "%s should be in a pattern", dir)
	}
}

func TestGroupedSystemPathPatterns(t *testing.T) {
	cfg := defaultConfig()
	
	// Count how many "keep" patterns with OR syntax we have at the beginning
	keepPatternCount := 0
	for _, pr := range cfg.Patterns {
		if pr.Rule.Action == "keep" && strings.Contains(pr.Pattern, "|") {
			keepPatternCount++
			// These should all be system patterns
			if !strings.Contains(pr.Pattern, "*") {
				// Non-wildcard keep patterns are system paths
				assert.True(t, len(strings.Split(pr.Pattern, "|")) > 1,
					"System patterns should group multiple items")
			}
		}
		// Stop counting when we hit generic patterns
		if strings.HasPrefix(pr.Pattern, "*") && pr.Rule.Action != "keep" {
			break
		}
	}
	
	// We should have multiple grouped patterns (6 groups of system paths)
	assert.Equal(t, 6, keepPatternCount,
		"Should have exactly 6 grouped system path patterns")
	
	// Verify all expected system paths are present across all patterns
	expectedPaths := []string{
		"CARGO_HOME", "RUSTUP_HOME", "PNPM_HOME", "NIX_PATH",
		"NIX_USER_PROFILE_DIR", "BROWSERS_PROFILE_PATH", "SOLANA_HOME",
		"KITTY_INSTALLATION_DIR", "ZSH_CACHE_DIR", "MINIO_HOME",
		"DOT_PATH", "FORGIT_INSTALL_DIR", "__MISE_ORIG_PATH",
		"TMUX_PLUGIN_MANAGER_PATH", "GOPATH", "GOROOT", "NVM_DIR",
		"RBENV_ROOT", "PYENV_ROOT", "SDKMAN_DIR", "HOMEBREW_PREFIX",
		"HOMEBREW_CELLAR", "HOMEBREW_REPOSITORY",
		// New application and development tool directories
		"DBT_PROFILES_DIR", "DBT_PROJECT_DIR", "CCC_SESSIONS_DIR",
		"TMP_DIR", "PLAYGROUND_DIR", "GHOSTTY_RESOURCES_DIR",
		"GHOSTTY_BIN_DIR", "FOUNDRY_DIR",
	}
	
	// Check each path is in at least one pattern
	for _, path := range expectedPaths {
		found := false
		for _, pr := range cfg.Patterns {
			if strings.Contains(pr.Pattern, path) {
				found = true
				break
			}
		}
		assert.True(t, found, "Path %s should be in at least one pattern", path)
	}
	
	// Verify patterns are using OR syntax efficiently
	for _, pr := range cfg.Patterns {
		if containsSystemPath(pr.Pattern) {
			// Each grouped pattern should have multiple items
			items := strings.Split(pr.Pattern, "|")
			assert.Greater(t, len(items), 1, 
				"System path patterns should group multiple paths with OR syntax")
			
			// But not too many (for readability)
			assert.LessOrEqual(t, len(items), 10, 
				"Patterns should not have too many items for readability")
		}
	}
}