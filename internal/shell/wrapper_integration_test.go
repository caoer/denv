package shell

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrapperWithColoredPrompt(t *testing.T) {
	tests := []struct {
		name       string
		shellPath  string
		envName    string
		projectName string
		checks     func(t *testing.T, wrapper string)
	}{
		{
			name:      "bash wrapper includes colored prompt",
			shellPath: "/bin/bash",
			envName:   "test-env",
			projectName: "myproject",
			checks: func(t *testing.T, wrapper string) {
				// Should have the colored prompt command
				assert.Contains(t, wrapper, "PS1=", "should set PS1")
				assert.Contains(t, wrapper, "\033[", "should contain ANSI color code")
				assert.Contains(t, wrapper, "(test-env)", "should contain environment name")
				assert.Contains(t, wrapper, "\033[0m", "should reset color")
				assert.Contains(t, wrapper, "export PS1", "should export PS1")
			},
		},
		{
			name:      "zsh wrapper includes colored prompt",
			shellPath: "/usr/bin/zsh",
			envName:   "staging",
			projectName: "api",
			checks: func(t *testing.T, wrapper string) {
				// Zsh uses PROMPT variable and %F{n} color syntax
				assert.Contains(t, wrapper, "PROMPT=", "should set PROMPT")
				assert.Contains(t, wrapper, "%F{", "should contain zsh color syntax")
				assert.Contains(t, wrapper, "(staging)", "should contain environment name")
				assert.Contains(t, wrapper, "%f", "should reset color with %f")
				assert.Contains(t, wrapper, "export PROMPT", "should export PROMPT")
			},
		},
		{
			name:      "fish wrapper includes colored prompt function",
			shellPath: "/usr/local/bin/fish",
			envName:   "production",
			projectName: "web",
			checks: func(t *testing.T, wrapper string) {
				// Fish uses function and set_color
				assert.Contains(t, wrapper, "function fish_prompt", "should define fish_prompt function")
				assert.Contains(t, wrapper, "set_color", "should use set_color command")
				assert.Contains(t, wrapper, "(production)", "should contain environment name")
				assert.Contains(t, wrapper, "set_color normal", "should reset color")
				assert.Contains(t, wrapper, "__fish_default_prompt", "should call default prompt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare environment variables
			env := map[string]string{
				"DENV_ENV_NAME":     tt.envName,
				"DENV_PROJECT_NAME": tt.projectName,
				"DENV_ENV":          "/tmp/denv/" + tt.projectName + "-" + tt.envName,
				"DENV_PROJECT":      "/tmp/denv/" + tt.projectName,
				"DENV_SESSION":      "test-session",
				"SHELL":             tt.shellPath,
			}
			
			// Detect shell type
			shellType, _ := DetectShell(tt.shellPath)
			
			// Generate wrapper
			wrapper := GenerateShellWrapper(shellType, env)
			
			// Verify wrapper is not empty
			require.NotEmpty(t, wrapper, "wrapper should not be empty")
			
			// Run specific checks
			tt.checks(t, wrapper)
		})
	}
}

func TestColorConsistencyInWrapper(t *testing.T) {
	// Same environment should get same color in wrapper
	env1 := map[string]string{
		"DENV_ENV_NAME":     "myproject-dev",
		"DENV_PROJECT_NAME": "myproject",
		"DENV_ENV":          "/tmp/denv/myproject-dev",
		"DENV_PROJECT":      "/tmp/denv/myproject",
		"DENV_SESSION":      "session1",
		"SHELL":             "/bin/bash",
	}
	
	env2 := map[string]string{
		"DENV_ENV_NAME":     "myproject-dev",
		"DENV_PROJECT_NAME": "myproject",
		"DENV_ENV":          "/tmp/denv/myproject-dev",
		"DENV_PROJECT":      "/tmp/denv/myproject",
		"DENV_SESSION":      "session2",
		"SHELL":             "/bin/bash",
	}
	
	wrapper1 := GenerateWrapper(env1)
	wrapper2 := GenerateWrapper(env2)
	
	// Extract PS1 lines
	ps1Line1 := extractPS1Line(wrapper1)
	ps1Line2 := extractPS1Line(wrapper2)
	
	assert.Equal(t, ps1Line1, ps1Line2, "same environment should produce same colored prompt")
}

func TestPromptPreservesOriginalPS1(t *testing.T) {
	env := map[string]string{
		"DENV_ENV_NAME": "test",
		"SHELL":         "/bin/bash",
	}
	
	wrapper := GenerateWrapper(env)
	
	// Check that original PS1 is preserved
	assert.Contains(t, wrapper, "$PS1", "should reference original PS1")
	
	// Should append to PS1, not replace it
	ps1Line := extractPS1Line(wrapper)
	assert.Contains(t, ps1Line, "$PS1\"", "should append to existing PS1")
}

// Helper function to extract PS1 line from wrapper
func extractPS1Line(wrapper string) string {
	lines := strings.Split(wrapper, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PS1=") {
			return line
		}
	}
	return ""
}