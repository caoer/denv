package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/paths"
	"github.com/caoer/denv/internal/project"
	"github.com/caoer/denv/internal/ui"
)

// Ps shows the current environment status and modifications or a specific environment if provided
func Ps(targetEnv string) error {
	// If a specific environment is requested, show that environment's info
	if targetEnv != "" {
		return showSpecificEnvironment(targetEnv)
	}

	// Otherwise show the current environment (if any)
	return showCurrentEnvironment()
}

func showCurrentEnvironment() error {
	// Check if we're in a denv environment
	envPath := os.Getenv("DENV_ENV")
	if envPath == "" {
		fmt.Println("Not in a denv environment")
		fmt.Println("\nUse 'denv enter [name]' to enter an environment")
		return nil
	}

	// Get environment info from env vars
	envName := os.Getenv("DENV_ENV_NAME")
	projectName := os.Getenv("DENV_PROJECT_NAME")
	sessionID := os.Getenv("DENV_SESSION")

	// Load runtime to get detailed info
	runtime, err := environment.LoadRuntime(envPath)
	if err != nil {
		return fmt.Errorf("failed to load runtime: %w", err)
	}

	// Print header
	fmt.Printf("\n🚀 Current Environment: %s\n", envName)
	fmt.Printf("📦 Project: %s\n", projectName)
	fmt.Printf("🔑 Session: %s\n", sessionID)
	fmt.Println(strings.Repeat("─", 50))

	// Show the environment details
	showEnvironmentDetails(runtime)

	// Show active sessions in this environment
	if runtime != nil && len(runtime.Sessions) > 0 {
		fmt.Println("\n👥 Active Sessions:")
		for id, session := range runtime.Sessions {
			status := "active"
			if !sessionExists(session.PID) {
				status = "orphaned"
			}
			fmt.Printf("   %s (PID: %d) - %s\n", id, session.PID, status)
		}
	}

	// Show environment paths with shortened display
	fmt.Println("\n📂 Environment Paths:")
	fmt.Printf("   Environment: %s\n", paths.ShortenPath(envPath, 0))
	fmt.Printf("   Project:     %s\n", paths.ShortenPath(os.Getenv("DENV_PROJECT"), 0))
	
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("Type 'exit' to leave this environment")
	
	return nil
}

func showSpecificEnvironment(envName string) error {
	// Detect project for the current directory
	cwd, _ := os.Getwd()
	projectName, err := project.DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	// Build the environment path
	envPath := paths.EnvironmentPath(projectName, envName)
	
	// Check if the environment exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("environment '%s' does not exist for project '%s'", envName, projectName)
	}

	// Load runtime for the specific environment
	runtime, err := environment.LoadRuntime(envPath)
	if err != nil {
		return fmt.Errorf("failed to load runtime for environment '%s': %w", envName, err)
	}

	// Print header
	fmt.Printf("\n📋 Environment: %s\n", envName)
	fmt.Printf("📦 Project: %s\n", projectName)
	fmt.Println(strings.Repeat("─", 50))

	// Show the environment details
	showEnvironmentDetails(runtime)

	// Show sessions if any
	if runtime != nil && len(runtime.Sessions) > 0 {
		fmt.Println("\n👥 Sessions in this environment:")
		activeCount := 0
		for id, session := range runtime.Sessions {
			status := "orphaned"
			if sessionExists(session.PID) {
				status = "active"
				activeCount++
			}
			fmt.Printf("   %s (PID: %d) - %s\n", id, session.PID, status)
		}
		if activeCount > 0 {
			fmt.Printf("\n⚠️  This environment has %d active session(s)\n", activeCount)
		}
	}

	// Show environment paths with shortened display
	fmt.Println("\n📂 Environment Paths:")
	fmt.Printf("   Environment: %s\n", paths.ShortenPath(envPath, 0))
	fmt.Printf("   Project:     %s\n", paths.ShortenPath(paths.ProjectPath(projectName), 0))
	
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Printf("Use 'denv enter %s' to enter this environment\n", envName)
	
	return nil
}

func showEnvironmentDetails(runtime *environment.Runtime) {
	if runtime == nil {
		return
	}

	// First, show environment variable modifications organized by type
	if len(runtime.Overrides) > 0 {
		fmt.Println("\n🔧 Environment Variable Modifications:")
		
		// Categorize overrides
		var portMappings []ui.PortMapping
		var urlRewrites []ui.URLRewrite
		var isolatedPaths []ui.IsolatedPath
		
		// Process each override
		for key, override := range runtime.Overrides {
			switch override.Rule {
			case "random_port":
				origPort, _ := strconv.Atoi(override.Original)
				currPort, _ := strconv.Atoi(override.Current)
				portMappings = append(portMappings, ui.PortMapping{
					Name:     key,
					Original: origPort,
					Mapped:   currPort,
				})
				
			case "rewrite_ports":
				if override.Original != override.Current {
					// Colorize ports in URLs
					orig := override.Original
					curr := override.Current
					
					for origPort, newPort := range runtime.Ports {
						orig = ui.ColorizePortInURL(orig, origPort)
						curr = ui.ColorizePortInURL(curr, newPort)
					}
					
					urlRewrites = append(urlRewrites, ui.URLRewrite{
						Name:     key,
						Original: orig,
						Current:  curr,
					})
				}
				
			case "isolate":
				// Apply path shortening to isolated paths
				origShort := paths.ShortenPath(override.Original, 0)
				currShort := paths.ShortenPath(override.Current, 0)
				
				isolatedPaths = append(isolatedPaths, ui.IsolatedPath{
					Name:     key,
					Original: origShort,
					Current:  currShort,
				})
			}
		}
		
		// Display using the new UI cards
		if len(portMappings) > 0 {
			fmt.Print(ui.RenderPortCard(portMappings))
		}
		
		if len(urlRewrites) > 0 {
			fmt.Print("\n")
			fmt.Print(ui.RenderURLCard(urlRewrites))
		}
		
		if len(isolatedPaths) > 0 {
			fmt.Print("\n")
			fmt.Print(ui.RenderIsolatedPathCard(isolatedPaths))
		}
	}
}

// truncateValue and sessionExists are defined in enter.go and list.go respectively