package override

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/caoer/denv/internal/config"
)

func TestMatchesPatternWithOR(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		key      string
		expected bool
	}{
		{
			name:     "single pattern match",
			pattern:  "CARGO_HOME",
			key:      "CARGO_HOME",
			expected: true,
		},
		{
			name:     "single pattern no match",
			pattern:  "CARGO_HOME",
			key:      "RUSTUP_HOME",
			expected: false,
		},
		{
			name:     "OR pattern first match",
			pattern:  "CARGO_HOME | RUSTUP_HOME",
			key:      "CARGO_HOME",
			expected: true,
		},
		{
			name:     "OR pattern second match",
			pattern:  "CARGO_HOME | RUSTUP_HOME",
			key:      "RUSTUP_HOME",
			expected: true,
		},
		{
			name:     "OR pattern no match",
			pattern:  "CARGO_HOME | RUSTUP_HOME",
			key:      "GOPATH",
			expected: false,
		},
		{
			name:     "complex OR pattern",
			pattern:  "CARGO_HOME | RUSTUP_HOME | PNPM_HOME | NIX_PATH",
			key:      "PNPM_HOME",
			expected: true,
		},
		{
			name:     "wildcard with OR",
			pattern:  "*_PORT | PORT",
			key:      "API_PORT",
			expected: true,
		},
		{
			name:     "wildcard with OR second match",
			pattern:  "*_PORT | PORT",
			key:      "PORT",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MatchesPattern(tt.pattern, tt.key)
			assert.Equal(t, tt.expected, result,
				"Pattern %s matching %s should be %v", tt.pattern, tt.key, tt.expected)
		})
	}
}

func TestApplyRulesWithGroupedPatterns(t *testing.T) {
	// Create a config with grouped system paths patterns
	cfg := &config.Config{
		Patterns: []config.PatternRule{
			{
				Pattern: "CARGO_HOME | RUSTUP_HOME | GOPATH | GOROOT | NVM_DIR",
				Rule: config.Rule{
					Action: "keep",
				},
			},
			{
				Pattern: "HOMEBREW_PREFIX | HOMEBREW_CELLAR",
				Rule: config.Rule{
					Action: "keep",
				},
			},
			{
				Pattern: "*_PORT | PORT",
				Rule: config.Rule{
					Action: "random_port",
					Range:  []int{30000, 39999},
				},
			},
		},
	}

	// Test environment variables
	env := map[string]string{
		"CARGO_HOME": "/home/user/.cargo",
		"RUSTUP_HOME": "/home/user/.rustup",
		"GOPATH": "/home/user/go",
		"API_PORT": "3000",
		"OTHER_VAR": "value",
	}

	// Port mappings
	ports := map[int]int{
		3000: 30001,
	}

	result, overrides := ApplyRules(env, cfg, ports, "/tmp/test")

	// System paths should be kept unchanged
	assert.Equal(t, "/home/user/.cargo", result["CARGO_HOME"])
	assert.Equal(t, "/home/user/.rustup", result["RUSTUP_HOME"])
	assert.Equal(t, "/home/user/go", result["GOPATH"])
	
	// Port should be remapped
	assert.Equal(t, "30001", result["API_PORT"])
	
	// Other variables should pass through
	assert.Equal(t, "value", result["OTHER_VAR"])
	
	// Check overrides
	assert.Empty(t, overrides["CARGO_HOME"], "CARGO_HOME should not have override")
	assert.Empty(t, overrides["RUSTUP_HOME"], "RUSTUP_HOME should not have override")
	assert.Empty(t, overrides["GOPATH"], "GOPATH should not have override")
	assert.NotEmpty(t, overrides["API_PORT"], "API_PORT should have override")
}

func TestApplicationDirectoriesNotRemapped(t *testing.T) {
	// Create a config with application-specific directories in system paths
	cfg := &config.Config{
		Patterns: []config.PatternRule{
			{
				Pattern: "DBT_PROFILES_DIR | TMP_DIR | GHOSTTY_RESOURCES_DIR | FOUNDRY_DIR",
				Rule: config.Rule{
					Action: "keep",
				},
			},
			{
				Pattern: "*_DIR",
				Rule: config.Rule{
					Action: "isolate",
					Base:   "${DENV_ENV}",
				},
			},
		},
	}

	// Test environment variables including application directories
	env := map[string]string{
		"DBT_PROFILES_DIR": "/Users/test/dbt",
		"TMP_DIR": "/var/tmp",
		"GHOSTTY_RESOURCES_DIR": "/Applications/Ghostty.app/Resources",
		"USER_DIR": "/Users/test/userdata",
	}

	// Port mappings (none needed for this test)
	ports := map[int]int{}

	result, overrides := ApplyRules(env, cfg, ports, "/tmp/test-env")

	// Application-specific directories should NOT be remapped (kept as-is)
	assert.Equal(t, "/Users/test/dbt", result["DBT_PROFILES_DIR"], "DBT_PROFILES_DIR should not be remapped")
	assert.Empty(t, overrides["DBT_PROFILES_DIR"], "DBT_PROFILES_DIR should not have override record")

	assert.Equal(t, "/var/tmp", result["TMP_DIR"], "TMP_DIR should not be remapped")
	assert.Empty(t, overrides["TMP_DIR"], "TMP_DIR should not have override record")

	assert.Equal(t, "/Applications/Ghostty.app/Resources", result["GHOSTTY_RESOURCES_DIR"], "GHOSTTY_RESOURCES_DIR should not be remapped")
	assert.Empty(t, overrides["GHOSTTY_RESOURCES_DIR"], "GHOSTTY_RESOURCES_DIR should not have override record")

	// USER_DIR should be remapped (matches *_DIR but not in system paths)
	assert.Equal(t, "/tmp/test-env/userdata", result["USER_DIR"], "USER_DIR should be remapped")
	assert.NotEmpty(t, overrides["USER_DIR"], "USER_DIR should have override record")
}

func TestDenvHomeNotRemapped(t *testing.T) {
	// Create a config with DENV_HOME in system paths
	cfg := &config.Config{
		Patterns: []config.PatternRule{
			{
				Pattern: "DENV_HOME | CARGO_HOME | RUSTUP_HOME",
				Rule: config.Rule{
					Action: "keep",
				},
			},
			{
				Pattern: "*_HOME",
				Rule: config.Rule{
					Action: "isolate",
					Base:   "${DENV_ENV}",
				},
			},
		},
	}

	// Test environment variables including DENV_HOME
	env := map[string]string{
		"DENV_HOME": "/Users/test/.denv",
		"USER_HOME": "/Users/test/user_data",
		"CARGO_HOME": "/Users/test/.cargo",
	}

	// Port mappings (none needed for this test)
	ports := map[int]int{}

	result, overrides := ApplyRules(env, cfg, ports, "/tmp/test-env")

	// DENV_HOME should NOT be remapped (kept as-is)
	assert.Equal(t, "/Users/test/.denv", result["DENV_HOME"], "DENV_HOME should not be remapped")
	assert.Empty(t, overrides["DENV_HOME"], "DENV_HOME should not have override record")
	
	// CARGO_HOME should also not be remapped (in same pattern)
	assert.Equal(t, "/Users/test/.cargo", result["CARGO_HOME"], "CARGO_HOME should not be remapped")
	assert.Empty(t, overrides["CARGO_HOME"], "CARGO_HOME should not have override record")
	
	// USER_HOME should be remapped (matches *_HOME but not in system paths)
	assert.Equal(t, "/tmp/test-env/user_data", result["USER_HOME"], "USER_HOME should be remapped")
	assert.NotEmpty(t, overrides["USER_HOME"], "USER_HOME should have override record")
}