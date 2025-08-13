package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/zitao/denv/internal/config"
	"github.com/zitao/denv/internal/environment"
	"github.com/zitao/denv/internal/override"
	"github.com/zitao/denv/internal/paths"
	"github.com/zitao/denv/internal/ports"
	"github.com/zitao/denv/internal/project"
	"github.com/zitao/denv/internal/session"
)

// PrepareEnvResponse contains all the information needed by the bash wrapper
type PrepareEnvResponse struct {
	EnvPath     string            `json:"env_path"`
	ProjectPath string            `json:"project_path"`
	ProjectName string            `json:"project_name"`
	EnvName     string            `json:"env_name"`
	SessionID   string            `json:"session_id"`
	Ports       map[string]string `json:"ports"`
	Overrides   map[string]string `json:"overrides"`
}

// PrepareEnv prepares environment data for the bash wrapper
func PrepareEnv(envName string) error {
	if envName == "" {
		envName = "default"
	}

	// Detect project
	cwd, _ := os.Getwd()
	projectName, err := project.DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	// Check for project override
	cfg, _ := config.LoadConfig(filepath.Join(paths.DenvHome(), "config.yaml"))
	projectName = project.DetectProjectWithConfig(cwd, cfg)

	// Create environment path
	envPath := paths.EnvironmentPath(projectName, envName)
	os.MkdirAll(envPath, 0755)

	// Create project path
	projectPath := paths.ProjectPath(projectName)
	os.MkdirAll(projectPath, 0755)
	os.MkdirAll(filepath.Join(projectPath, "hooks"), 0755)

	// Create .denv symlinks in project directory
	createProjectSymlinks(cwd, envPath, projectPath)

	// Load or create runtime
	runtime, _ := environment.LoadRuntime(envPath)
	if runtime == nil {
		runtime = environment.NewRuntime(projectName, envName)
	}

	// Setup port manager and allocate ports
	pm := ports.NewPortManager(envPath)
	commonPorts := []int{3000, 3001, 3002, 5432, 6379, 8080, 8081}
	portMappings := make(map[string]string)
	
	for _, port := range commonPorts {
		mappedPort := pm.GetPort(port)
		runtime.Ports[port] = mappedPort
		portMappings[strconv.Itoa(port)] = strconv.Itoa(mappedPort)
	}

	// Create session
	sessionHandle := session.CreateSession(envPath, "")
	runtime.Sessions[sessionHandle.ID] = environment.Session{
		ID:      sessionHandle.ID,
		PID:     sessionHandle.PID,
		Started: time.Now(),
	}

	// Save runtime
	environment.SaveRuntime(envPath, runtime)

	// Prepare overrides
	currentEnv := os.Environ()
	envMap := make(map[string]string)
	for _, e := range currentEnv {
		if kv := splitEnv(e); len(kv) == 2 {
			envMap[kv[0]] = kv[1]
		}
	}
	overrides, _ := override.ApplyRules(envMap, cfg, runtime.Ports, envPath)

	// Create response
	response := PrepareEnvResponse{
		EnvPath:     envPath,
		ProjectPath: projectPath,
		ProjectName: projectName,
		EnvName:     envName,
		SessionID:   sessionHandle.ID,
		Ports:       portMappings,
		Overrides:   overrides,
	}

	// Output as JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

// GetEnvOverrides returns environment variable overrides as shell export commands
func GetEnvOverrides(envName string) error {
	if envName == "" {
		envName = "default"
	}

	// Detect project
	cwd, _ := os.Getwd()
	projectName, err := project.DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	// Load config
	cfg, _ := config.LoadConfig(filepath.Join(paths.DenvHome(), "config.yaml"))
	projectName = project.DetectProjectWithConfig(cwd, cfg)

	// Load runtime
	envPath := paths.EnvironmentPath(projectName, envName)
	runtime, err := environment.LoadRuntime(envPath)
	if err != nil {
		return err
	}

	// Get current environment
	currentEnv := os.Environ()
	envMap := make(map[string]string)
	for _, e := range currentEnv {
		if kv := splitEnv(e); len(kv) == 2 {
			envMap[kv[0]] = kv[1]
		}
	}

	// Apply overrides
	overrides, _ := override.ApplyRules(envMap, cfg, runtime.Ports, envPath)

	// Output as shell export commands
	for key, value := range overrides {
		fmt.Printf("export %s=\"%s\"\n", key, escapeShellValue(value))
	}

	return nil
}

// CleanupSession removes a session lock
func CleanupSession(sessionID string) error {
	// Find the session file across all environments
	homeDir := paths.DenvHome()
	
	// Walk through all environment directories
	err := filepath.Walk(homeDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}

		// Look for session directories
		if info.IsDir() && info.Name() == "sessions" {
			lockFile := filepath.Join(path, sessionID+".lock")
			if _, err := os.Stat(lockFile); err == nil {
				// Found the lock file, remove it
				os.Remove(lockFile)
				
				// Also update the runtime to remove the session
				envPath := filepath.Dir(path)
				if runtime, err := environment.LoadRuntime(envPath); err == nil {
					delete(runtime.Sessions, sessionID)
					environment.SaveRuntime(envPath, runtime)
				}
			}
		}
		
		return nil
	})

	return err
}

// Helper function for escaping shell values
func escapeShellValue(value string) string {
	// Escape special characters for shell
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "$", `\$`)
	value = strings.ReplaceAll(value, "`", "\\`")
	return value
}