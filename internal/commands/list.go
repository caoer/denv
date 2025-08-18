package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/paths"
	"github.com/caoer/denv/internal/ui"
)

// EnvironmentInfo represents basic environment information
type EnvironmentInfo struct {
	Project     string
	Environment string
	Sessions    int
	Ports       int
}

// ListEnvironments returns a list of all environments without printing
func ListEnvironments() ([]EnvironmentInfo, error) {
	denvHome := paths.DenvHome()
	
	// Find all environment directories
	entries, err := os.ReadDir(denvHome)
	if err != nil {
		return nil, fmt.Errorf("failed to read denv home: %w", err)
	}

	var environments []EnvironmentInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		// Parse project-environment pattern
		parts := strings.SplitN(name, "-", 2)
		if len(parts) == 2 {
			project := parts[0]
			env := parts[1]
			
			envPath := filepath.Join(denvHome, name)
			runtime, _ := environment.LoadRuntime(envPath)
			
			sessionCount := 0
			portCount := 0
			if runtime != nil {
				// Count active sessions
				for _, session := range runtime.Sessions {
					if sessionExists(session.PID) {
						sessionCount++
					}
				}
				portCount = len(runtime.Ports)
			}
			
			environments = append(environments, EnvironmentInfo{
				Project:     project,
				Environment: env,
				Sessions:    sessionCount,
				Ports:       portCount,
			})
		}
	}

	return environments, nil
}

// List shows all environments across all projects (previously ps -a)
func List() error {
	denvHome := paths.DenvHome()
	
	// Find all environment directories
	entries, err := os.ReadDir(denvHome)
	if err != nil {
		return fmt.Errorf("failed to read denv home: %w", err)
	}

	// Group environments by project
	projectEnvs := make(map[string][]ui.EnvInfo)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		
		name := entry.Name()
		// Parse project-environment pattern
		parts := strings.SplitN(name, "-", 2)
		if len(parts) == 2 {
			project := parts[0]
			envName := parts[1]
			
			// Load runtime to get session and port info
			envPath := filepath.Join(denvHome, name)
			runtime, _ := environment.LoadRuntime(envPath)
			
			sessionCount := 0
			portCount := 0
			active := false
			
			if runtime != nil {
				// Count active sessions
				for _, session := range runtime.Sessions {
					if sessionExists(session.PID) {
						sessionCount++
						active = true
					}
				}
				portCount = len(runtime.Ports)
			}
			
			envInfo := ui.EnvInfo{
				Name:     envName,
				Active:   active,
				Sessions: sessionCount,
				Ports:    portCount,
			}
			
			projectEnvs[project] = append(projectEnvs[project], envInfo)
		} else if name != "" && !strings.HasPrefix(name, ".") {
			// Standalone project directory
			projectEnvs[name] = []ui.EnvInfo{}
		}
	}

	if len(projectEnvs) == 0 {
		fmt.Println("No denv environments found")
		return nil
	}

	// Use the new UI renderer
	fmt.Println(ui.RenderEnvironmentList("All denv Environments", projectEnvs))
	return nil
}

func sessionExists(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Try to send signal 0 (doesn't actually send a signal, just checks if process exists)
	err = process.Signal(syscall.Signal(0))
	return err == nil
}