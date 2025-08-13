package commands

import (
	"fmt"
	"os"

	"github.com/zitao/denv/internal/environment"
	"github.com/zitao/denv/internal/paths"
	"github.com/zitao/denv/internal/project"
	"github.com/zitao/denv/internal/session"
)

func Clean(envName string) error {
	if envName == "" {
		return fmt.Errorf("environment name required")
	}

	// Detect current project
	cwd, _ := os.Getwd()
	projectName, err := project.DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	envPath := paths.EnvironmentPath(projectName, envName)

	// Check if environment exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist", envName)
	}

	// Check for active sessions
	runtime, _ := environment.LoadRuntime(envPath)
	if runtime != nil && len(runtime.Sessions) > 0 {
		activeSessions := 0
		for _, sess := range runtime.Sessions {
			if session.ProcessExists(sess.PID) {
				activeSessions++
			}
		}
		if activeSessions > 0 {
			return fmt.Errorf("cannot clean environment with %d active session(s)", activeSessions)
		}
	}

	// Remove environment directory
	if err := os.RemoveAll(envPath); err != nil {
		return fmt.Errorf("failed to remove environment: %w", err)
	}

	fmt.Printf("Removed environment '%s' for project %s\n", envName, projectName)
	return nil
}