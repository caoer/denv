package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/caoer/denv/internal/config"
	"github.com/caoer/denv/internal/paths"
	"github.com/caoer/denv/internal/project"
)

// Project manages project name overrides
func Project(action string, w io.Writer) error {
	// Detect current project
	cwd, _ := os.Getwd()
	projectName, err := project.DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	configPath := filepath.Join(paths.DenvHome(), "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	switch {
	case action == "":
		// Show current project name
		if override, ok := cfg.Projects[cwd]; ok {
			fmt.Fprintf(w, "Project: %s (overridden from %s)\n", override, projectName)
		} else {
			fmt.Fprintf(w, "Project: %s\n", projectName)
		}

	case strings.HasPrefix(action, "rename "):
		// Rename project
		newName := strings.TrimPrefix(action, "rename ")
		newName = strings.TrimSpace(newName)
		
		if newName == "" {
			return fmt.Errorf("new name required")
		}

		if cfg.Projects == nil {
			cfg.Projects = make(map[string]string)
		}
		cfg.Projects[cwd] = newName

		if err := config.SaveConfig(configPath, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Fprintf(w, "Project renamed to: %s\n", newName)
		fmt.Fprintf(w, "Override saved to %s\n", configPath)

	case action == "unset":
		// Remove project override
		if cfg.Projects != nil {
			delete(cfg.Projects, cwd)
			
			if err := config.SaveConfig(configPath, cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}
		}

		fmt.Fprintf(w, "Project override removed\n")
		fmt.Fprintf(w, "Will use detected name: %s\n", projectName)

	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}