package override

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/caoer/denv/internal/config"
)

func TestSystemPathsNotIsolated(t *testing.T) {
	// Create config with default patterns
	cfg := &config.Config{
		Projects: make(map[string]string),
		Patterns: []config.PatternRule{
			// System paths (should keep)
			{Pattern: "CARGO_HOME", Rule: config.Rule{Action: "keep"}},
			{Pattern: "PNPM_HOME", Rule: config.Rule{Action: "keep"}},
			{Pattern: "NIX_PATH", Rule: config.Rule{Action: "keep"}},
			{Pattern: "GOPATH", Rule: config.Rule{Action: "keep"}},
			// Generic patterns (should isolate)
			{Pattern: "*_HOME", Rule: config.Rule{Action: "isolate", Base: "${DENV_ENV}"}},
			{Pattern: "*_PATH", Rule: config.Rule{Action: "isolate", Base: "${DENV_ENV}"}},
		},
	}
	
	// Test environment variables
	env := map[string]string{
		"CARGO_HOME":   "/home/user/.cargo",
		"PNPM_HOME":    "/home/user/.pnpm",
		"NIX_PATH":     "/nix/var/nix/profiles",
		"GOPATH":       "/home/user/go",
		"MY_APP_HOME":  "/home/user/myapp",  // Should be isolated
		"MY_APP_PATH":  "/home/user/mypath", // Should be isolated
	}
	
	ports := map[int]int{}
	envPath := "/home/user/.denv/project-default"
	
	result, overrides := ApplyRules(env, cfg, ports, envPath)
	
	// System paths should be unchanged
	assert.Equal(t, "/home/user/.cargo", result["CARGO_HOME"])
	assert.Equal(t, "/home/user/.pnpm", result["PNPM_HOME"])
	assert.Equal(t, "/nix/var/nix/profiles", result["NIX_PATH"])
	assert.Equal(t, "/home/user/go", result["GOPATH"])
	
	// These should not be in overrides (no change)
	assert.NotContains(t, overrides, "CARGO_HOME")
	assert.NotContains(t, overrides, "PNPM_HOME")
	assert.NotContains(t, overrides, "NIX_PATH")
	assert.NotContains(t, overrides, "GOPATH")
	
	// Generic paths should be isolated
	assert.NotEqual(t, "/home/user/myapp", result["MY_APP_HOME"])
	assert.Contains(t, result["MY_APP_HOME"], envPath)
	assert.NotEqual(t, "/home/user/mypath", result["MY_APP_PATH"])
	assert.Contains(t, result["MY_APP_PATH"], envPath)
	
	// These should be in overrides (changed)
	assert.Contains(t, overrides, "MY_APP_HOME")
	assert.Contains(t, overrides, "MY_APP_PATH")
	assert.Equal(t, "isolate", overrides["MY_APP_HOME"].Rule)
	assert.Equal(t, "isolate", overrides["MY_APP_PATH"].Rule)
}

func TestPatternMatchingOrder(t *testing.T) {
	// Test that first matching pattern wins
	cfg := &config.Config{
		Projects: make(map[string]string),
		Patterns: []config.PatternRule{
			// Specific pattern first
			{Pattern: "CARGO_HOME", Rule: config.Rule{Action: "keep"}},
			// Generic pattern after
			{Pattern: "*_HOME", Rule: config.Rule{Action: "isolate", Base: "${DENV_ENV}"}},
		},
	}
	
	env := map[string]string{
		"CARGO_HOME": "/home/user/.cargo",
	}
	
	ports := map[int]int{}
	envPath := "/home/user/.denv/project-default"
	
	result, overrides := ApplyRules(env, cfg, ports, envPath)
	
	// Should use first matching pattern (keep)
	assert.Equal(t, "/home/user/.cargo", result["CARGO_HOME"])
	assert.NotContains(t, overrides, "CARGO_HOME")
}