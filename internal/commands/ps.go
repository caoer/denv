package commands

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/zitao/denv/internal/color"
	"github.com/zitao/denv/internal/environment"
	"github.com/zitao/denv/internal/paths"
	"github.com/zitao/denv/internal/project"
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
		portVars := []struct {
			name     string
			override environment.Override
			port     int
		}{}
		urlRewrites := []struct {
			name     string
			override environment.Override
		}{}
		isolatedPaths := []struct {
			name     string
			override environment.Override
		}{}
		
		// Sort variable names for consistent display
		var varNames []string
		for key := range runtime.Overrides {
			varNames = append(varNames, key)
		}
		sort.Strings(varNames)
		
		// Categorize each override and extract port info where applicable
		for _, key := range varNames {
			override := runtime.Overrides[key]
			
			switch override.Rule {
			case "random_port":
				// Extract port number from the value
				port := 0
				if override.Original != "" {
					port, _ = strconv.Atoi(override.Original)
				}
				portVars = append(portVars, struct {
					name     string
					override environment.Override
					port     int
				}{key, override, port})
				
			case "rewrite_ports":
				if override.Original != override.Current {
					urlRewrites = append(urlRewrites, struct {
						name     string
						override environment.Override
					}{key, override})
				}
				
			case "isolate":
				isolatedPaths = append(isolatedPaths, struct {
					name     string
						override environment.Override
				}{key, override})
			}
		}
		
		// Display port variables with improved visual design
		if len(portVars) > 0 {
			// Convert to PortMapping format for the new display
			var portMappings []color.PortMapping
			for _, pv := range portVars {
				origPort, _ := strconv.Atoi(pv.override.Original)
				currPort, _ := strconv.Atoi(pv.override.Current)
				portMappings = append(portMappings, color.PortMapping{
					Name:     pv.name,
					Original: origPort,
					Mapped:   currPort,
				})
			}
			
			// Use the new card display format
			fmt.Print(color.FormatPortCard(portMappings))
		}
		
		// Display URL rewrites with new card format
		if len(urlRewrites) > 0 {
			var urlRewriteList []color.URLRewrite
			for _, ur := range urlRewrites {
				// Colorize ports in URLs
				orig := ur.override.Original
				curr := ur.override.Current
				
				for origPort, newPort := range runtime.Ports {
					origPortStr := fmt.Sprintf(":%d", origPort)
					newPortStr := fmt.Sprintf(":%d", newPort)
					if strings.Contains(orig, origPortStr) {
						orig = color.ColorizePortInURL(orig, origPort)
					}
					if strings.Contains(curr, newPortStr) {
						curr = color.ColorizePortInURL(curr, newPort)
					}
				}
				
				urlRewriteList = append(urlRewriteList, color.URLRewrite{
					Name:     ur.name,
					Original: orig,
					Current:  curr,
				})
			}
			fmt.Print("\n")
			fmt.Print(color.FormatURLCard(urlRewriteList))
		}
		
		// Display isolated paths with new card format
		if len(isolatedPaths) > 0 {
			var pathList []color.IsolatedPath
			for _, ip := range isolatedPaths {
				// Apply path shortening to isolated paths
				origShort := paths.ShortenPath(ip.override.Original, 0)
				currShort := paths.ShortenPath(ip.override.Current, 0)
				
				pathList = append(pathList, color.IsolatedPath{
					Name:     ip.name,
					Original: origShort,
					Current:  currShort,
				})
			}
			fmt.Print("\n")
			fmt.Print(color.FormatIsolatedPathCard(pathList))
		}
	}
}

// truncateValue and sessionExists are defined in enter.go and list.go respectively