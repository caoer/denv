package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zitao/denv/internal/commands"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	
	switch command {
	case "enter":
		envName := ""
		if len(os.Args) > 2 {
			envName = os.Args[2]
		}
		if err := commands.Enter(envName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "ls", "list":
		if err := commands.List(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "clean":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: environment name required\n")
			os.Exit(1)
		}
		if err := commands.Clean(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "ps":
		if err := commands.Ps(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "sessions":
		// Parse flags
		fs := flag.NewFlagSet("sessions", flag.ExitOnError)
		cleanup := fs.Bool("cleanup", false, "Clean orphaned sessions")
		kill := fs.Bool("kill", false, "Terminate all sessions")
		fs.Parse(os.Args[2:])

		if err := commands.Sessions(*cleanup, *kill); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "export":
		envName := ""
		if len(os.Args) > 2 {
			envName = os.Args[2]
		}
		if err := commands.Export(envName, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "project":
		action := ""
		if len(os.Args) > 2 {
			// Join all remaining args as the action
			action = strings.Join(os.Args[2:], " ")
		}
		if err := commands.Project(action, os.Stdout); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	// Commands for bash wrapper integration
	case "prepare-env":
		envName := ""
		if len(os.Args) > 2 {
			envName = os.Args[2]
		}
		if err := commands.PrepareEnv(envName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "get-env-overrides":
		envName := ""
		if len(os.Args) > 2 {
			envName = os.Args[2]
		}
		if err := commands.GetEnvOverrides(envName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "cleanup-session":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: session ID required\n")
			os.Exit(1)
		}
		if err := commands.CleanupSession(os.Args[2]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`denv - Development Environment Manager

Usage:
  denv enter [name]      Enter environment (default: "default")
  denv ls                List all environments across all projects
  denv ps                Show current environment status
  denv clean <name>      Remove environment
  denv sessions          Show active sessions
  denv sessions --cleanup Clean orphaned sessions
  denv sessions --kill   Terminate all sessions
  denv export [name]     Export environment variables (for direnv)
  denv project           Show current project name
  denv project rename <name> Rename current project
  denv project unset     Remove project override
  denv help             Show this help

Environment Variables:
  DENV_HOME             Base directory for denv (default: ~/.denv)

Inside an environment:
  DENV_ENV              Current environment directory
  DENV_PROJECT          Shared project directory
  DENV_ENV_NAME         Environment name
  DENV_PROJECT_NAME     Project name
  DENV_SESSION          Session ID
  PORT_*                Remapped ports`)
}