package commands

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"

	"github.com/zitao/denv/internal/environment"
	"github.com/zitao/denv/internal/paths"
	"github.com/zitao/denv/internal/project"
)

// Export outputs environment variables for direnv integration
func Export(envName string, w io.Writer) error {
	if envName == "" {
		envName = "default"
	}

	// Detect current project
	cwd, _ := os.Getwd()
	projectName, err := project.DetectProject(cwd)
	if err != nil {
		return fmt.Errorf("failed to detect project: %w", err)
	}

	// Load runtime
	envPath := paths.EnvironmentPath(projectName, envName)
	runtime, err := environment.LoadRuntime(envPath)
	if err != nil {
		return fmt.Errorf("failed to load environment: %w", err)
	}
	if runtime == nil {
		return fmt.Errorf("environment '%s' does not exist for project %s", envName, projectName)
	}

	// Output environment variables
	fmt.Fprintf(w, "# denv environment: %s/%s\n", projectName, envName)
	
	// Core variables
	fmt.Fprintf(w, "export DENV_HOME=\"%s\"\n", paths.DenvHome())
	fmt.Fprintf(w, "export DENV_ENV=\"%s\"\n", envPath)
	fmt.Fprintf(w, "export DENV_PROJECT=\"%s\"\n", paths.ProjectPath(projectName))
	fmt.Fprintf(w, "export DENV_ENV_NAME=\"%s\"\n", envName)
	fmt.Fprintf(w, "export DENV_PROJECT_NAME=\"%s\"\n", projectName)

	// Port mappings (sorted for consistency)
	var ports []int
	for p := range runtime.Ports {
		ports = append(ports, p)
	}
	sort.Ints(ports)

	for _, orig := range ports {
		mapped := runtime.Ports[orig]
		fmt.Fprintf(w, "export PORT_%d=\"%d\"\n", orig, mapped)
		fmt.Fprintf(w, "export ORIGINAL_PORT_%d=\"%d\"\n", orig, orig)
	}

	// Apply overrides if any
	if len(runtime.Overrides) > 0 {
		fmt.Fprintln(w, "\n# Variable overrides")
		
		// Sort keys for consistency
		var keys []string
		for k := range runtime.Overrides {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		
		for _, key := range keys {
			override := runtime.Overrides[key]
			fmt.Fprintf(w, "export %s=\"%s\"\n", key, escapeForShell(override.Current))
		}
	}

	return nil
}

func escapeForShell(value string) string {
	// Basic escaping for shell export
	return strconv.Quote(value)[1:len(strconv.Quote(value))-1]
}