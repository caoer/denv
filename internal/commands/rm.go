package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/paths"
	"github.com/caoer/denv/internal/project"
	"github.com/caoer/denv/internal/session"
)

func Rm(envName string, all bool) error {
	if all {
		return rmAll()
	}
	
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

func rmAll() error {
	denvHome := paths.DenvHome()
	
	// Find all environment directories
	entries, err := os.ReadDir(denvHome)
	if err != nil {
		return fmt.Errorf("failed to read denv home: %w", err)
	}

	removedCount := 0
	var removedEnvs []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		// Parse project-environment pattern
		parts := strings.SplitN(name, "-", 2)
		if len(parts) != 2 {
			// Skip non-environment directories (like standalone project dirs)
			continue
		}
		
		project := parts[0]
		env := parts[1]
		envPath := filepath.Join(denvHome, name)
		
		// Check if environment has active sessions
		runtime, _ := environment.LoadRuntime(envPath)
		hasActiveSessions := false
		
		if runtime != nil && len(runtime.Sessions) > 0 {
			for _, sess := range runtime.Sessions {
				if session.ProcessExists(sess.PID) {
					hasActiveSessions = true
					break
				}
			}
		}
		
		// Only remove if no active sessions
		if !hasActiveSessions {
			if err := os.RemoveAll(envPath); err != nil {
				return fmt.Errorf("failed to remove environment %s: %w", name, err)
			}
			removedCount++
			removedEnvs = append(removedEnvs, fmt.Sprintf("%s:%s", project, env))
		}
	}

	if removedCount == 0 {
		fmt.Println("No inactive environments found to remove")
	} else {
		fmt.Printf("Removed %d inactive environment(s):\n", removedCount)
		for _, env := range removedEnvs {
			fmt.Printf("  • %s\n", env)
		}
	}
	
	return nil
}