package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Setup: Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yaml := `projects:
  /path/to/project: custom-name
patterns:
  - pattern: "*_PORT|PORT"
    rule:
      action: random_port
      range: [30000, 39999]
`
	os.WriteFile(configPath, []byte(yaml), 0644)

	// Test: Load config
	cfg, err := LoadConfig(configPath)
	assert.NoError(t, err)
	assert.Equal(t, "custom-name", cfg.Projects["/path/to/project"])
	
	// Find the pattern in the slice
	found := false
	for _, pr := range cfg.Patterns {
		if pr.Pattern == "*_PORT|PORT" {
			assert.Equal(t, "random_port", pr.Rule.Action)
			assert.Equal(t, []int{30000, 39999}, pr.Rule.Range)
			found = true
			break
		}
	}
	assert.True(t, found, "Pattern *_PORT|PORT should exist")
}

func TestDefaultConfig(t *testing.T) {
	// Test: Returns defaults when no config exists
	cfg, err := LoadConfig("/nonexistent/path")
	assert.NoError(t, err)
	assert.NotNil(t, cfg.Patterns)
	
	// Check that default patterns are present
	found := false
	for _, pr := range cfg.Patterns {
		if pr.Pattern == "*_PORT|PORT" {
			found = true
			break
		}
	}
	assert.True(t, found, "Default patterns should include *_PORT|PORT")
}