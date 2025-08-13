package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/caoer/denv/internal/commands"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
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

	case "rm":
		// Parse flags for rm command
		fs := flag.NewFlagSet("rm", flag.ExitOnError)
		all := fs.Bool("all", false, "Remove all inactive environments")
		_ = fs.Parse(os.Args[2:])

		envName := ""
		if !*all && fs.NArg() < 1 {
			fmt.Fprintf(os.Stderr, "Error: environment name required (or use --all)\n")
			os.Exit(1)
		}
		if fs.NArg() > 0 {
			envName = fs.Arg(0)
		}
		
		if err := commands.Rm(envName, *all); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "ps":
		envName := ""
		if len(os.Args) > 2 {
			envName = os.Args[2]
		}
		if err := commands.Ps(envName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "sessions":
		// Parse flags
		fs := flag.NewFlagSet("sessions", flag.ExitOnError)
		cleanup := fs.Bool("cleanup", false, "Clean orphaned sessions")
		kill := fs.Bool("kill", false, "Terminate all sessions")
		_ = fs.Parse(os.Args[2:])

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
  denv rm <name>         Remove environment
  denv rm --all          Remove all inactive environments
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