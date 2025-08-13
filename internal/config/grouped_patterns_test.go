package config

import (
	"testing"
	"strings"
	
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupedPatternsReadability(t *testing.T) {
	cfg := defaultConfig()
	
	// Check that no single pattern is too long for readability (with spaces it's longer)
	for _, pr := range cfg.Patterns {
		assert.LessOrEqual(t, len(pr.Pattern), 140,
			"Pattern should not exceed 140 characters for readability: %s", pr.Pattern)
	}
}

func TestGroupedPatternsEfficiency(t *testing.T) {
	cfg := defaultConfig()
	
	// Count total patterns
	totalPatterns := len(cfg.Patterns)
	
	// With grouped patterns, we should have around 9 patterns total
	assert.LessOrEqual(t, totalPatterns, 10,
		"Should have 10 or fewer total patterns with grouping")
	
	// Count how many individual items would exist without grouping
	totalItems := 0
	for _, pr := range cfg.Patterns {
		if strings.Contains(pr.Pattern, "|") {
			totalItems += len(strings.Split(pr.Pattern, "|"))
		} else {
			totalItems++
		}
	}
	
	// We should have consolidated many items into fewer patterns
	assert.Greater(t, totalItems, totalPatterns*2,
		"Grouping should reduce pattern count significantly")
}

func TestAllSystemPathsCovered(t *testing.T) {
	cfg := defaultConfig()
	
	// Comprehensive list of all system paths that should be kept
	allSystemPaths := []string{
		// Core denv system
		"DENV_HOME",
		// Programming languages
		"CARGO_HOME", "RUSTUP_HOME", "GOPATH", "GOROOT", 
		"NVM_DIR", "RBENV_ROOT", "PYENV_ROOT", "PNPM_HOME", "SDKMAN_DIR",
		// Package managers
		"HOMEBREW_PREFIX", "HOMEBREW_CELLAR", "HOMEBREW_REPOSITORY",
		"NIX_PATH", "NIX_USER_PROFILE_DIR",
		// Applications
		"SOLANA_HOME", "KITTY_INSTALLATION_DIR", "MINIO_HOME",
		"TMUX_PLUGIN_MANAGER_PATH", "BROWSERS_PROFILE_PATH",
		// Shell/tools
		"ZSH_CACHE_DIR", "DOT_PATH", "FORGIT_INSTALL_DIR", "__MISE_ORIG_PATH",
	}
	
	// Verify each system path is covered
	for _, path := range allSystemPaths {
		found := false
		for _, pr := range cfg.Patterns {
			if strings.Contains(pr.Pattern, path) {
				require.Equal(t, "keep", pr.Rule.Action,
					"%s should have 'keep' action", path)
				found = true
				break
			}
		}
		require.True(t, found, "System path %s should be in a pattern", path)
	}
}