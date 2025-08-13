package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZshPromptHandling(t *testing.T) {
	tests := []struct {
		name      string
		envName   string
		checks    func(t *testing.T, prompt string)
	}{
		{
			name:    "zsh should use PROMPT variable",
			envName: "test-env",
			checks: func(t *testing.T, prompt string) {
				// Zsh should use PROMPT, not PS1
				assert.Contains(t, prompt, "PROMPT=", "should set PROMPT variable for zsh")
				
				// Should use zsh color syntax %F{} instead of ANSI codes
				assert.Contains(t, prompt, "%F{", "should use zsh color syntax")
				assert.Contains(t, prompt, "%f", "should reset color with %f")
				assert.Contains(t, prompt, "(test-env)", "should contain environment name")
				
				// Should preserve original prompt
				assert.Contains(t, prompt, "$PROMPT", "should preserve original PROMPT")
			},
		},
		{
			name:    "zsh prompt should handle 256 colors",
			envName: "colorful-env",
			checks: func(t *testing.T, prompt string) {
				// Should use 256 color syntax in zsh
				// %F{39} for color 39, etc.
				assert.Regexp(t, `%F\{[0-9]+\}`, prompt, "should use zsh 256 color syntax")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := GenerateColoredPrompt(tt.envName, Zsh)
			require.NotEmpty(t, prompt, "prompt should not be empty")
			tt.checks(t, prompt)
		})
	}
}

func TestZshWrapperIntegration(t *testing.T) {
	env := map[string]string{
		"DENV_ENV_NAME":     "zsh-test",
		"DENV_PROJECT_NAME": "myproject",
		"DENV_ENV":          "/tmp/denv/myproject-zsh-test",
		"DENV_PROJECT":      "/tmp/denv/myproject",
		"DENV_SESSION":      "test-session",
		"SHELL":             "/bin/zsh",
	}
	
	wrapper := GenerateWrapper(env)
	
	// Check that zsh wrapper uses PROMPT
	assert.Contains(t, wrapper, "PROMPT=", "zsh wrapper should set PROMPT")
	assert.Contains(t, wrapper, "export PROMPT", "zsh wrapper should export PROMPT")
	
	// Verify the prompt is properly formatted for zsh
	assert.Contains(t, wrapper, "%F{", "should use zsh color syntax")
	assert.Contains(t, wrapper, "%f", "should use zsh color reset")
}