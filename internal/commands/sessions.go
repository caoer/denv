package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/paths"
	"github.com/caoer/denv/internal/project"
	"github.com/caoer/denv/internal/session"
)

func Sessions(cleanup, kill bool) error {
	// Detect current project
	cwd, _ := os.Getwd()
	projectName, err := project.DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	// Get all environments for this project
	home := paths.DenvHome()
	entries, _ := os.ReadDir(home)
	prefix := projectName + "-"

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			envPath := paths.EnvironmentPath(projectName, strings.TrimPrefix(entry.Name(), prefix))
			
			if cleanup {
				cleaned := session.CleanupOrphaned(envPath)
				if cleaned > 0 {
					fmt.Printf("Cleaned %d orphaned session(s) in %s\n", cleaned, entry.Name())
				}
			} else if kill {
				runtime, _ := environment.LoadRuntime(envPath)
				if runtime != nil && len(runtime.Sessions) > 0 {
					for id, sess := range runtime.Sessions {
						if session.ProcessExists(sess.PID) {
							proc, _ := os.FindProcess(sess.PID)
							proc.Signal(os.Interrupt)
							fmt.Printf("Sent SIGTERM to session %s (PID %d)\n", id, sess.PID)
						}
					}
				}
			} else {
				// List sessions
				runtime, _ := environment.LoadRuntime(envPath)
				if runtime != nil && len(runtime.Sessions) > 0 {
					fmt.Printf("\nActive sessions in %s:\n", entry.Name())
					for id, sess := range runtime.Sessions {
						status := "active"
						if !session.ProcessExists(sess.PID) {
							status = "orphaned"
						}
						fmt.Printf("  - Session %s (PID %d) - %s\n", id, sess.PID, status)
					}
				}
			}
		}
	}

	return nil
}