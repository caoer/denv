package shell

import (
	"fmt"
	"path/filepath"
	"strings"
)

type ShellType int

const (
	Bash ShellType = iota
	Zsh
	Fish
	Sh
)

func (s ShellType) String() string {
	switch s {
	case Bash:
		return "bash"
	case Zsh:
		return "zsh"
	case Fish:
		return "fish"
	case Sh:
		return "sh"
	default:
		return "unknown"
	}
}

// DetectShell detects the shell type from the shell path
func DetectShell(shellPath string) (ShellType, string) {
	if shellPath == "" {
		return Bash, "bash" // Default to bash
	}

	base := filepath.Base(shellPath)

	switch {
	case strings.Contains(base, "zsh"):
		return Zsh, "zsh"
	case strings.Contains(base, "fish"):
		return Fish, "fish"
	case strings.Contains(base, "bash"):
		return Bash, "bash"
	case base == "sh":
		return Sh, "sh"
	default:
		return Bash, "bash" // Default to bash for unknown shells
	}
}

// GetShellCommand returns the command and arguments to start a shell with the environment
func GetShellCommand(shellType ShellType, envScript string) []string {
	switch shellType {
	case Bash:
		// Bash supports --init-file to source a script on startup
		return []string{"bash", "--init-file", envScript}

	case Zsh:
		// Zsh doesn't have --init-file, so we source and exec
		return []string{"zsh", "-c", fmt.Sprintf("source %s && exec zsh", envScript)}

	case Fish:
		// Fish uses different syntax for sourcing
		return []string{"fish", "-c", fmt.Sprintf("source %s; and exec fish", envScript)}

	case Sh:
		// Plain sh uses . instead of source
		return []string{"sh", "-c", fmt.Sprintf(". %s && exec sh", envScript)}

	default:
		// Default to bash behavior
		return []string{"bash", "--init-file", envScript}
	}
}

// GenerateShellWrapper generates a shell-specific wrapper script
func GenerateShellWrapper(shellType ShellType, env map[string]string) string {
	switch shellType {
	case Fish:
		return generateFishWrapper(env)
	default:
		// Bash, Zsh, and Sh use similar syntax
		return GenerateWrapper(env)
	}
}

func generateFishWrapper(env map[string]string) string {
	var script strings.Builder

	// Fish uses different syntax
	script.WriteString("#!/usr/bin/env fish\n\n")

	// Export environment variables (Fish syntax)
	script.WriteString("# Set environment variables\n")
	for key, value := range env {
		script.WriteString(fmt.Sprintf("set -x %s \"%s\"\n", key, escapeFishValue(value)))
	}
	script.WriteString("\n")

	// Define cleanup function (Fish syntax)
	script.WriteString("# Cleanup function\n")
	script.WriteString("function cleanup\n")
	script.WriteString("    # Run exit hook if exists\n")
	script.WriteString("    if test -f \"$DENV_PROJECT/hooks/on-exit.sh\"\n")
	script.WriteString("        source \"$DENV_PROJECT/hooks/on-exit.sh\"\n")
	script.WriteString("    end\n")
	script.WriteString("    # Remove session lock\n")
	script.WriteString("    rm -f \"$DENV_ENV/sessions/$DENV_SESSION.lock\"\n")
	script.WriteString("end\n\n")

	// Setup signal handlers (Fish syntax)
	script.WriteString("# Setup signal handlers\n")
	script.WriteString("trap cleanup EXIT\n")
	script.WriteString("trap cleanup SIGINT\n")
	script.WriteString("trap cleanup SIGTERM\n\n")

	// Run enter hook
	script.WriteString("# Run enter hook if exists\n")
	script.WriteString("if test -f \"$DENV_PROJECT/hooks/on-enter.sh\"\n")
	script.WriteString("    source \"$DENV_PROJECT/hooks/on-enter.sh\"\n")
	script.WriteString("end\n\n")

	// Modify prompt with color (Fish syntax)
	script.WriteString("# Modify prompt with color\n")
	envName := env["DENV_ENV_NAME"]
	if envName == "" {
		envName = "denv"
	}
	promptFunc := GenerateColoredPrompt(envName, Fish)
	script.WriteString(promptFunc)
	script.WriteString("\n")

	return script.String()
}

func escapeFishValue(value string) string {
	// Fish has different escaping rules
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, `$`, `\$`)
	return value
}

