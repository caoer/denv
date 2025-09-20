package commands

import (
	"fmt"
	"path/filepath"

	"github.com/caoer/denv/internal/config"
	"github.com/caoer/denv/internal/paths"
)

// ConfigUpdate updates the config file with new default patterns while preserving projects
func ConfigUpdate() error {
	configPath := filepath.Join(paths.DenvHome(), "config.yaml")

	// Load existing config
	existingCfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load existing config: %w", err)
	}

	// Create new config with default patterns
	newCfg := &config.Config{
		Projects: existingCfg.Projects, // Preserve existing project overrides
		Patterns: config.GetDefaultPatterns(), // Use new default patterns
	}

	// Save the updated config
	if err := config.SaveConfig(configPath, newCfg); err != nil {
		return fmt.Errorf("failed to save updated config: %w", err)
	}

	fmt.Printf("✅ Config updated with new default patterns at %s\n", configPath)
	fmt.Println("\nNew ignored directories added:")
	fmt.Println("  • DBT_PROFILES_DIR, DBT_PROJECT_DIR")
	fmt.Println("  • CCC_SESSIONS_DIR, TMP_DIR, PLAYGROUND_DIR")
	fmt.Println("  • GHOSTTY_RESOURCES_DIR, GHOSTTY_BIN_DIR")
	fmt.Println("  • FOUNDRY_DIR, CLAUDE_BASH_MAINTAIN_PROJECT_WORKING_DIR")
	fmt.Println("\nYour project overrides have been preserved.")

	return nil
}