package override

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zitao/denv/internal/config"
)

func TestPatternMatch(t *testing.T) {
	tests := []struct {
		pattern string
		key     string
		match   bool
	}{
		{"*_PORT", "DB_PORT", true},
		{"*_PORT", "PORT_DB", false},
		{"*_PORT|PORT", "PORT", true},
		{"DATABASE_URL", "DATABASE_URL", true},
		{"*_URL", "DATABASE_URL", true},
		{"*_KEY|*_TOKEN", "API_KEY", true},
		{"*_KEY|*_TOKEN", "AUTH_TOKEN", true},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			assert.Equal(t, tt.match, MatchesPattern(tt.pattern, tt.key))
		})
	}
}

func TestRewriteURLPorts(t *testing.T) {
	ports := map[int]int{
		5432: 35432,
		3000: 33000,
		6379: 36379,
	}

	tests := []struct {
		input    string
		expected string
	}{
		{
			"postgres://localhost:5432/db",
			"postgres://localhost:35432/db",
		},
		{
			"http://127.0.0.1:3000/api",
			"http://127.0.0.1:33000/api",
		},
		{
			"redis://localhost:6379",
			"redis://localhost:36379",
		},
		{
			"redis://external.com:6379",
			"redis://external.com:6379", // Don't change external
		},
		{
			"http://0.0.0.0:3000",
			"http://0.0.0.0:33000",
		},
	}

	for _, tt := range tests {
		result := RewriteURL(tt.input, ports)
		assert.Equal(t, tt.expected, result)
	}
}

func TestApplyRules(t *testing.T) {
	cfg := &config.Config{
		Patterns: []config.PatternRule{
			{Pattern: "*_PORT", Rule: config.Rule{Action: "random_port"}},
			{Pattern: "*_URL", Rule: config.Rule{Action: "rewrite_ports"}},
			{Pattern: "*_KEY", Rule: config.Rule{Action: "keep"}},
			{Pattern: "*_PATH", Rule: config.Rule{Action: "isolate", Base: "/tmp/denv"}},
		},
	}

	env := map[string]string{
		"DB_PORT":      "5432",
		"DATABASE_URL": "postgres://localhost:5432/db",
		"API_KEY":      "secret123",
		"DATA_PATH":    "/var/data",
	}

	ports := map[int]int{5432: 35432}
	result, overrides := ApplyRules(env, cfg, ports, "/tmp/denv")

	assert.Equal(t, "35432", result["DB_PORT"])
	assert.Contains(t, result["DATABASE_URL"], "35432")
	assert.Equal(t, "secret123", result["API_KEY"]) // Unchanged
	assert.Equal(t, "/tmp/denv/data", result["DATA_PATH"])
	
	// Test that overrides are tracked correctly
	assert.NotNil(t, overrides)
	assert.Contains(t, overrides, "DB_PORT")
	assert.Contains(t, overrides, "DATABASE_URL")
	assert.Contains(t, overrides, "DATA_PATH")
	assert.NotContains(t, overrides, "API_KEY") // Should not be in overrides since it wasn't changed
}