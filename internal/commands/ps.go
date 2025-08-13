package commands

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/zitao/denv/internal/environment"
)

// Ps shows the current environment status and modifications
func Ps() error {
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

	// Print status
	fmt.Printf("\n🚀 Current Environment: %s\n", envName)
	fmt.Printf("📦 Project: %s\n", projectName)
	fmt.Printf("🔑 Session: %s\n", sessionID)
	fmt.Println(strings.Repeat("─", 50))

	// Show environment variable modifications if any
	if runtime != nil && len(runtime.Overrides) > 0 {
		fmt.Println("\n🔧 Active Environment Variable Modifications:")
		
		// Group by rule type
		portVars := []string{}
		urlRewrites := []string{}
		isolatedPaths := []string{}
		
		// Sort variable names for consistent display
		var varNames []string
		for key := range runtime.Overrides {
			varNames = append(varNames, key)
		}
		sort.Strings(varNames)
		
		for _, key := range varNames {
			override := runtime.Overrides[key]
			
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

	// Show port mappings
	if runtime != nil && len(runtime.Ports) > 0 {
		fmt.Println("\n📍 Port Mapping Summary:")
		// Sort ports for consistent display
		var portList []int
		for orig := range runtime.Ports {
			portList = append(portList, orig)
		}
		sort.Ints(portList)
		for _, orig := range portList {
			fmt.Printf("   %d → %d\n", orig, runtime.Ports[orig])
		}
	}

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

	// Show environment paths
	fmt.Println("\n📂 Environment Paths:")
	fmt.Printf("   Environment: %s\n", envPath)
	fmt.Printf("   Project:     %s\n", os.Getenv("DENV_PROJECT"))
	
	fmt.Println("\n" + strings.Repeat("─", 50))
	fmt.Println("Type 'exit' to leave this environment")
	
	return nil
}

// sessionExists is defined in list.go

// Use the truncateValue function from enter.go