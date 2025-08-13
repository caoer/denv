package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
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
	"github.com/zitao/denv/internal/shell"
)

func Enter(envName string) error {
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
	if err := createProjectSymlinks(cwd, envPath, projectPath); err != nil {
		// Log error but don't fail (symlinks are optional convenience)
		fmt.Fprintf(os.Stderr, "Warning: failed to create symlinks: %v\n", err)
	}

	// Load or create runtime
	runtime, _ := environment.LoadRuntime(envPath)
	if runtime == nil {
		runtime = environment.NewRuntime(projectName, envName)
	}

	// Setup port manager
	pm := ports.NewPortManager(envPath)
	
	// Get common ports
	commonPorts := []int{3000, 3001, 3002, 5432, 6379, 8080, 8081}
	for _, port := range commonPorts {
		mappedPort := pm.GetPort(port)
		runtime.Ports[port] = mappedPort
	}

	// Create session (skip in test mode)
	var sessionHandle *session.SessionHandle
	if os.Getenv("DENV_TEST_MODE") != "1" {
		sessionHandle = session.CreateSession(envPath, "")
		runtime.Sessions[sessionHandle.ID] = environment.Session{
			ID:      sessionHandle.ID,
			PID:     sessionHandle.PID,
			Started: time.Now(),
		}
	} else {
		// Create a dummy session for test mode
		sessionHandle = &session.SessionHandle{
			ID:  "test-session",
			PID: os.Getpid(),
		}
	}

	// Save runtime
	environment.SaveRuntime(envPath, runtime)

	// Prepare environment variables
	env := make(map[string]string)
	
	// Core denv variables
	env["DENV_HOME"] = paths.DenvHome()
	env["DENV_ENV"] = envPath
	env["DENV_PROJECT"] = projectPath
	env["DENV_ENV_NAME"] = envName
	env["DENV_PROJECT_NAME"] = projectName
	env["DENV_SESSION"] = sessionHandle.ID

	// Port mappings
	for orig, mapped := range runtime.Ports {
		env[fmt.Sprintf("PORT_%d", orig)] = strconv.Itoa(mapped)
		env[fmt.Sprintf("ORIGINAL_PORT_%d", orig)] = strconv.Itoa(orig)
	}

	// Apply override rules
	currentEnv := os.Environ()
	envMap := make(map[string]string)
	for _, e := range currentEnv {
		if kv := splitEnv(e); len(kv) == 2 {
			envMap[kv[0]] = kv[1]
		}
	}
	
	overridden, overrides := override.ApplyRules(envMap, cfg, runtime.Ports, envPath)
	for k, v := range overridden {
		env[k] = v
	}
	
	// Store overrides in runtime for persistence
	runtime.Overrides = overrides
	environment.SaveRuntime(envPath, runtime)

	// Check for test mode
	if os.Getenv("DENV_TEST_MODE") == "1" {
		allEnvPorts := getAllProjectEnvironmentPorts(projectName, envName)
		printEnterMessage(envName, projectName, runtime.Ports, overrides, allEnvPorts)
		return nil
	}

	// Detect shell type
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/bash"
	}
	shellType, _ := shell.DetectShell(shellPath)
	
	// Generate shell-specific wrapper script
	wrapperScript := shell.GenerateShellWrapper(shellType, env)
	
	// Write wrapper to temp file
	tmpFile, err := os.CreateTemp("", "denv-wrapper-*.sh")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	
	tmpFile.WriteString(wrapperScript)
	tmpFile.Close()

	// Print entry message with all project environments
	allEnvPorts := getAllProjectEnvironmentPorts(projectName, envName)
	printEnterMessage(envName, projectName, runtime.Ports, overrides, allEnvPorts)

	// Get shell-specific command
	shellArgs := shell.GetShellCommand(shellType, tmpFile.Name())
	
	// Start new shell with appropriate method
	cmd := exec.Command(shellArgs[0], shellArgs[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func splitEnv(env string) []string {
	if idx := strings.Index(env, "="); idx >= 0 {
		return []string{env[:idx], env[idx+1:]}
	}
	return []string{env}
}

func createProjectSymlinks(projectDir, envPath, projectPath string) error {
	// Create .denv directory in project
	denvDir := filepath.Join(projectDir, ".denv")
	if err := os.MkdirAll(denvDir, 0755); err != nil {
		return fmt.Errorf("failed to create .denv directory: %w", err)
	}

	// Create or update current symlink (points to environment)
	currentLink := filepath.Join(denvDir, "current")
	
	// Remove existing symlink if it exists
	os.Remove(currentLink)
	
	// Create new symlink
	if err := os.Symlink(envPath, currentLink); err != nil {
		return fmt.Errorf("failed to create current symlink: %w", err)
	}

	// Create or update project symlink (points to shared project directory)
	projectLink := filepath.Join(denvDir, "project")
	
	// Remove existing symlink if it exists
	os.Remove(projectLink)
	
	// Create new symlink
	if err := os.Symlink(projectPath, projectLink); err != nil {
		return fmt.Errorf("failed to create project symlink: %w", err)
	}

	// Don't modify .gitignore - let the user decide whether to ignore .denv/
	// They can add it manually if they want

	return nil
}

// getAllProjectEnvironmentPorts returns a map of mapped port -> environment name for all environments in the project
func getAllProjectEnvironmentPorts(projectName, currentEnv string) map[int]string {
	portOwners := make(map[int]string)
	
	// Get all environments for this project
	envPattern := filepath.Join(paths.DenvHome(), projectName+"-*")
	matches, err := filepath.Glob(envPattern)
	if err != nil {
		return portOwners
	}
	
	for _, envPath := range matches {
		// Extract environment name from path
		envDir := filepath.Base(envPath)
		envName := strings.TrimPrefix(envDir, projectName+"-")
		
		// Load runtime for this environment
		runtime, err := environment.LoadRuntime(envPath)
		if err != nil || runtime == nil {
			continue
		}
		
		// Record which environment owns each mapped port
		for _, mappedPort := range runtime.Ports {
			portOwners[mappedPort] = envName
		}
	}
	
	return portOwners
}

func printEnterMessage(envName, projectName string, ports map[int]int, overrides map[string]environment.Override, allEnvPorts map[int]string) {
	fmt.Printf("\n🚀 Entering '%s' environment for %s\n", envName, projectName)
	fmt.Println(strings.Repeat("─", 50))
	
	// Environment variable modifications section
	if len(overrides) > 0 {
		fmt.Println("\n🔧 Environment Variable Modifications:")
		
		// Group by rule type for better organization
		portVars := []string{}
		urlRewrites := []string{}
		isolatedPaths := []string{}
		
		// Sort variable names for consistent display
		var varNames []string
		for key := range overrides {
			varNames = append(varNames, key)
		}
		sort.Strings(varNames)
		
		for _, key := range varNames {
			override := overrides[key]
			
			// Format the entry showing: VAR_NAME: original → new
			var entry string
			switch override.Rule {
			case "random_port":
				entry = fmt.Sprintf("   %s: %s → %s", key, override.Original, override.Current)
				portVars = append(portVars, entry)
			case "rewrite_ports":
				if override.Original != override.Current {
					// Show abbreviated URLs for readability
					orig := truncateValue(override.Original, 50)
					curr := truncateValue(override.Current, 50)
					entry = fmt.Sprintf("   %s:\n      %s\n      → %s", key, orig, curr)
					urlRewrites = append(urlRewrites, entry)
				}
			case "isolate":
				entry = fmt.Sprintf("   %s:\n      %s\n      → %s", key, override.Original, override.Current)
				isolatedPaths = append(isolatedPaths, entry)
			}
		}
		
		// Display each category
		if len(portVars) > 0 {
			fmt.Println("\n   [Port Variables]")
			for _, entry := range portVars {
				fmt.Println(entry)
			}
		}
		
		if len(urlRewrites) > 0 {
			fmt.Println("\n   [URL/Endpoint Rewrites]")
			for _, entry := range urlRewrites {
				fmt.Println(entry)
			}
		}
		
		if len(isolatedPaths) > 0 {
			fmt.Println("\n   [Isolated Paths]")
			for _, entry := range isolatedPaths {
				fmt.Println(entry)
			}
		}
	}
	
	// Also show the core port mappings table for quick reference
	if len(ports) > 0 {
		fmt.Println("\n📍 Port Mapping Summary:")
		// Sort ports for consistent display
		var portList []int
		for orig := range ports {
			portList = append(portList, orig)
		}
		sort.Ints(portList)
		for _, orig := range portList {
			fmt.Printf("   %d → %d\n", orig, ports[orig])
		}
	}
	
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("✨ Environment ready! Type 'exit' to leave.")
}

func truncateValue(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

