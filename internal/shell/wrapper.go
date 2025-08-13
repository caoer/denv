package shell

import (
	"fmt"
	"sort"
	"strings"
)

func GenerateWrapper(env map[string]string) string {
	var script strings.Builder

	// Add shebang and set options
	script.WriteString("#!/bin/bash\n")
	script.WriteString("set -e\n\n")

	// Export environment variables
	script.WriteString("# Set environment variables\n")
	script.WriteString(GenerateExportScript(env))
	script.WriteString("\n")

	// Define cleanup function
	script.WriteString("# Cleanup function\n")
	script.WriteString("cleanup() {\n")
	script.WriteString("    # Run exit hook if exists\n")
	script.WriteString(`    if [ -f "$DENV_PROJECT/hooks/on-exit.sh" ]; then`)
	script.WriteString("\n")
	script.WriteString(`        source "$DENV_PROJECT/hooks/on-exit.sh"`)
	script.WriteString("\n")
	script.WriteString("    fi\n")
	script.WriteString(`    # Remove session lock`)
	script.WriteString("\n")
	script.WriteString(`    rm -f "$DENV_ENV/sessions/$DENV_SESSION.lock"`)
	script.WriteString("\n")
	script.WriteString("}\n\n")

	// Setup signal handlers
	script.WriteString("# Setup signal handlers\n")
	script.WriteString("trap cleanup EXIT\n")
	script.WriteString("trap 'cleanup; exit' SIGINT SIGTERM\n\n")

	// Run enter hook
	script.WriteString("# Run enter hook if exists\n")
	script.WriteString(`if [ -f "$DENV_PROJECT/hooks/on-enter.sh" ]; then`)
	script.WriteString("\n")
	script.WriteString(`    source "$DENV_PROJECT/hooks/on-enter.sh"`)
	script.WriteString("\n")
	script.WriteString("fi\n\n")

	// Start shell with colored prompt
	script.WriteString("# Start shell with colored prompt\n")
	// Extract environment name from the env map
	envName := env["DENV_ENV_NAME"]
	if envName == "" {
		envName = "denv"
	}
	// Detect shell type from SHELL env var for proper prompt generation
	shellPath := env["SHELL"]
	if shellPath == "" {
		shellPath = "/bin/bash"
	}
	shellType, _ := DetectShell(shellPath)
	promptCmd := GenerateColoredPrompt(envName, shellType)
	script.WriteString(promptCmd)
	script.WriteString("\n")
	
	// Export the appropriate variable based on shell type
	if shellType == Zsh {
		script.WriteString("export PROMPT\n")
	} else {
		script.WriteString("export PS1\n")
	}

	return script.String()
}

func GenerateExportScript(env map[string]string) string {
	var script strings.Builder

	// Sort keys for consistent output
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := env[key]
		script.WriteString(fmt.Sprintf("export %s=\"%s\"\n", key, escapeShellValue(value)))
	}

	return script.String()
}

func GenerateCleanupScript(lockFile, exitHook string) string {
	var script strings.Builder

	script.WriteString("#!/bin/bash\n")
	script.WriteString("# Cleanup script\n\n")

	if exitHook != "" {
		script.WriteString(fmt.Sprintf("if [ -f \"%s\" ]; then\n", exitHook))
		script.WriteString(fmt.Sprintf("    source \"%s\"\n", exitHook))
		script.WriteString("fi\n\n")
	}

	if lockFile != "" {
		script.WriteString(fmt.Sprintf("rm -f \"%s\"\n", lockFile))
	}

	return script.String()
}

func escapeShellValue(value string) string {
	// Escape special characters for shell
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	value = strings.ReplaceAll(value, "$", `\$`)
	value = strings.ReplaceAll(value, "`", "\\`")
	return value
}

