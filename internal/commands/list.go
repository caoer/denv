package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/zitao/denv/internal/environment"
	"github.com/zitao/denv/internal/paths"
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
	projectEnvs := make(map[string][]string)
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
			projectEnvs[project] = append(projectEnvs[project], env)
		} else if name != "" && !strings.HasPrefix(name, ".") {
			// Standalone project directory
			projectEnvs[name] = []string{}
		}
	}

	if len(projectEnvs) == 0 {
		fmt.Println("No denv environments found")
		return nil
	}

	fmt.Println("\n🌍 All denv Environments:")
	fmt.Println(strings.Repeat("─", 50))

	// Sort projects for consistent display
	var projects []string
	for project := range projectEnvs {
		projects = append(projects, project)
	}
	sort.Strings(projects)

	for _, project := range projects {
		envs := projectEnvs[project]
		fmt.Printf("\n📦 %s\n", project)
		
		if len(envs) == 0 {
			fmt.Println("   (shared project directory only)")
			continue
		}
		
		// Sort environments
		sort.Strings(envs)
		for _, env := range envs {
			envPath := filepath.Join(denvHome, project+"-"+env)
			runtime, _ := environment.LoadRuntime(envPath)
			
			status := "inactive"
			sessionCount := 0
			if runtime != nil && len(runtime.Sessions) > 0 {
				// Count active sessions
				for _, session := range runtime.Sessions {
					if sessionExists(session.PID) {
						sessionCount++
					}
				}
				if sessionCount > 0 {
					status = fmt.Sprintf("%d active session(s)", sessionCount)
				}
			}
			
			portInfo := ""
			if runtime != nil && len(runtime.Ports) > 0 {
				portInfo = fmt.Sprintf(" [%d ports mapped]", len(runtime.Ports))
			}
			
			fmt.Printf("   • %s: %s%s\n", env, status, portInfo)
		}
	}
	
	fmt.Println("\n" + strings.Repeat("─", 50))
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