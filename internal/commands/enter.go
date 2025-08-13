package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/caoer/denv/internal/color"
	"github.com/caoer/denv/internal/config"
	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/override"
	"github.com/caoer/denv/internal/paths"
	"github.com/caoer/denv/internal/ports"
	"github.com/caoer/denv/internal/project"
	"github.com/caoer/denv/internal/session"
	"github.com/caoer/denv/internal/shell"
)

func Enter(envName string) error {
	// Check if we're already in a denv environment
	if existingEnv := os.Getenv("DENV_ENV_NAME"); existingEnv != "" {
		return fmt.Errorf("already in denv environment '%s'. Please exit the current environment before entering a new one", existingEnv)
	}
	
	if envName == "" {
		envName = "default"
	}

	// Detect project
	cwd, _ := os.Getwd()
	_, err := project.DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	// Check for project override
	cfg, _ := config.LoadConfig(filepath.Join(paths.DenvHome(), "config.yaml"))
	projectName := project.DetectProjectWithConfig(cwd, cfg)

	// Create environment path
	envPath := paths.EnvironmentPath(projectName, envName)
	os.MkdirAll(envPath, 0755)

	// Create project path
	projectPath := paths.ProjectPath(projectName)
	os.MkdirAll(projectPath, 0755)
	os.MkdirAll(filepath.Join(projectPath, "hooks"), 0755)

	// Create .denv symlinks in project directory
	if err := createProjectSymlinks(cwd, envPath, projectPath, projectName, envName); err != nil {
		// Log error but don't fail (symlinks are optional convenience)
		fmt.Fprintf(os.Stderr, "Warning: failed to create symlinks: %v\n", err)
	}

	// Load or create runtime
	runtime, _ := environment.LoadRuntime(envPath)
	if runtime == nil {
		runtime = environment.NewRuntime(projectName, envName)
	}

	// Setup port manager and initialize with existing runtime ports
	pm := ports.NewPortManager(envPath)
	
	// Initialize port manager with existing runtime ports to respect them
	if len(runtime.Ports) > 0 {
		pm.InitializeWithPorts(runtime.Ports)
	}
	
	// Collect ports that are actually used by environment variables
	usedPorts := collectUsedPorts(os.Environ(), cfg)
	for port := range usedPorts {
		// Check if we already have a mapping in runtime
		if existingPort, exists := runtime.Ports[port]; exists {
			// Use existing mapping
			runtime.Ports[port] = existingPort
		} else {
			// Get new mapping from port manager
			mappedPort := pm.GetPort(port)
			runtime.Ports[port] = mappedPort
		}
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
	_ = environment.SaveRuntime(envPath, runtime)

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
	
	_, _ = tmpFile.WriteString(wrapperScript)
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

	// Run the shell and wait for it to exit
	err = cmd.Run()
	
	// Clean up the session after shell exits
	cleanupSession(envPath, sessionHandle)
	
	return err
}

// cleanupSession removes the session from runtime and releases the lock
func cleanupSession(envPath string, sessionHandle *session.SessionHandle) {
	if sessionHandle == nil {
		return
	}
	
	// Release the session lock
	sessionHandle.Release()
	
	// Remove session from runtime.json
	runtime, err := environment.LoadRuntime(envPath)
	if err != nil || runtime == nil {
		return
	}
	
	// Remove this session from the runtime
	delete(runtime.Sessions, sessionHandle.ID)
	
	// Save the updated runtime
	environment.SaveRuntime(envPath, runtime)
	
	// Remove the lock file
	lockPath := filepath.Join(envPath, "sessions", sessionHandle.ID+".lock")
	os.Remove(lockPath)
	
	// Clean up .denv symlinks if no more sessions
	if len(runtime.Sessions) == 0 {
		cleanupProjectSymlinks(runtime.Project, runtime.Environment)
	}
}

// cleanupProjectSymlinks removes the symlinks in .denv directory
func cleanupProjectSymlinks(projectName, envName string) {
	cwd, _ := os.Getwd()
	denvDir := filepath.Join(cwd, ".denv")
	
	// Remove environment-specific symlink
	envLinkName := fmt.Sprintf("*%s-%s", projectName, envName)
	envLink := filepath.Join(denvDir, envLinkName)
	os.Remove(envLink)
	
	// Check if there are any other active environments for this project
	pattern := filepath.Join(denvDir, fmt.Sprintf("*%s-*", projectName))
	matches, _ := filepath.Glob(pattern)
	
	// If no other environments, remove the project symlink too
	if len(matches) == 0 {
		projectLink := filepath.Join(denvDir, projectName)
		os.Remove(projectLink)
	}
}

func splitEnv(env string) []string {
	if idx := strings.Index(env, "="); idx >= 0 {
		return []string{env[:idx], env[idx+1:]}
	}
	return []string{env}
}

// collectUsedPorts analyzes environment variables to find which ports are actually referenced
func collectUsedPorts(environ []string, cfg *config.Config) map[int]bool {
	ports := make(map[int]bool)
	envMap := make(map[string]string)
	
	// Parse environment into map
	for _, e := range environ {
		if kv := splitEnv(e); len(kv) == 2 {
			envMap[kv[0]] = kv[1]
		}
	}
	
	// Check each environment variable against patterns
	for key, value := range envMap {
		for _, pr := range cfg.Patterns {
			if override.MatchesPattern(pr.Pattern, key) {
				switch pr.Rule.Action {
				case "random_port":
					// This is a port variable, extract the port number
					if port, err := strconv.Atoi(value); err == nil {
						ports[port] = true
					}
				case "rewrite_ports":
					// Extract ports from URLs
					extractedPorts := extractPortsFromURL(value)
					for _, p := range extractedPorts {
						ports[p] = true
					}
				}
				break // First matching pattern wins
			}
		}
	}
	
	return ports
}

// extractPortsFromURL finds port numbers in URLs
func extractPortsFromURL(url string) []int {
	var ports []int
	// Simple regex to find :PORT patterns
	pattern := regexp.MustCompile(`:([0-9]+)`)
	matches := pattern.FindAllStringSubmatch(url, -1)
	for _, match := range matches {
		if len(match) > 1 {
			if port, err := strconv.Atoi(match[1]); err == nil {
				ports = append(ports, port)
			}
		}
	}
	return ports
}

func createProjectSymlinks(projectDir, envPath, projectPath, projectName, envName string) error {
	// Create .denv directory in project
	denvDir := filepath.Join(projectDir, ".denv")
	if err := os.MkdirAll(denvDir, 0755); err != nil {
		return fmt.Errorf("failed to create .denv directory: %w", err)
	}

	// Clean up old symlink naming convention
	oldLinks := []string{"current", "project"}
	for _, link := range oldLinks {
		linkPath := filepath.Join(denvDir, link)
		// Remove if it's a symlink (preserve regular files)
		if info, err := os.Lstat(linkPath); err == nil && info.Mode()&os.ModeSymlink != 0 {
			os.Remove(linkPath)
		}
	}

	// Create environment-specific symlink with star prefix
	// Format: *projectname-environment
	envLinkName := fmt.Sprintf("*%s-%s", projectName, envName)
	envLink := filepath.Join(denvDir, envLinkName)
	
	// Remove existing symlink if it exists
	os.Remove(envLink)
	
	// Create new symlink
	if err := os.Symlink(envPath, envLink); err != nil {
		return fmt.Errorf("failed to create environment symlink: %w", err)
	}

	// Create project symlink with just the project name
	projectLink := filepath.Join(denvDir, projectName)
	
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
					orig := override.Original
					curr := override.Current
					if len(orig) > 50 {
						orig = orig[:47] + "..."
					}
					if len(curr) > 50 {
						curr = curr[:47] + "..."
					}
					
					// Colorize ports in URLs
					for origPort, newPort := range ports {
						origPortStr := fmt.Sprintf(":%d", origPort)
						newPortStr := fmt.Sprintf(":%d", newPort)
						if strings.Contains(override.Original, origPortStr) {
							orig = color.ColorizePortInURL(orig, origPort)
						}
						if strings.Contains(override.Current, newPortStr) {
							curr = color.ColorizePortInURL(curr, newPort)
						}
					}
					
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
			// Convert to PortMapping format for the new display
			var portMappings []color.PortMapping
			for key, override := range overrides {
				if override.Rule == "random_port" {
					origPort, _ := strconv.Atoi(override.Original)
					currPort, _ := strconv.Atoi(override.Current)
					portMappings = append(portMappings, color.PortMapping{
						Name:     key,
						Original: origPort,
						Mapped:   currPort,
					})
				}
			}
			
			// Sort port mappings by name for consistent display
			sort.Slice(portMappings, func(i, j int) bool {
				return portMappings[i].Name < portMappings[j].Name
			})
			
			// Use the new card display format
			fmt.Print(color.FormatPortCard(portMappings))
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
	
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("✨ Environment ready! Type 'exit' to leave.")
}


