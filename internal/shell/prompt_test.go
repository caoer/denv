package shell

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateColoredPrompt(t *testing.T) {
	tests := []struct {
		name      string
		envName   string
		shellType ShellType
		checks    func(t *testing.T, prompt string)
	}{
		{
			name:      "bash prompt with color",
			envName:   "myproject-dev",
			shellType: Bash,
			checks: func(t *testing.T, prompt string) {
				// Should contain ANSI color codes
				assert.Contains(t, prompt, "\033[", "prompt should contain ANSI color code")
				// Should contain the environment name
				assert.Contains(t, prompt, "myproject-dev", "prompt should contain environment name")
				// Should reset color at the end
				assert.Contains(t, prompt, "\033[0m", "prompt should reset color")
				// Should preserve original PS1
				assert.Contains(t, prompt, "$PS1", "prompt should preserve original PS1")
			},
		},
		{
			name:      "zsh prompt with color",
			envName:   "app-staging",
			shellType: Zsh,
			checks: func(t *testing.T, prompt string) {
				// Zsh uses %F{n} color syntax
				assert.Contains(t, prompt, "%F{", "prompt should contain zsh color syntax")
				assert.Contains(t, prompt, "app-staging", "prompt should contain environment name")
				assert.Contains(t, prompt, "%f", "prompt should reset color with %f")
				assert.Contains(t, prompt, "$PROMPT", "prompt should preserve original PROMPT")
			},
		},
		{
			name:      "fish prompt function",
			envName:   "service-prod",
			shellType: Fish,
			checks: func(t *testing.T, prompt string) {
				// Fish uses set_color command
				assert.Contains(t, prompt, "set_color", "fish prompt should use set_color")
				assert.Contains(t, prompt, "service-prod", "prompt should contain environment name")
				assert.Contains(t, prompt, "set_color normal", "prompt should reset color")
				// Fish uses functions instead of PS1 variable
				assert.Contains(t, prompt, "function fish_prompt", "should define fish_prompt function")
			},
		},
		{
			name:      "sh prompt with color",
			envName:   "simple-test",
			shellType: Sh,
			checks: func(t *testing.T, prompt string) {
				// Plain sh should also support basic ANSI codes
				assert.Contains(t, prompt, "\033[", "prompt should contain ANSI color code")
				assert.Contains(t, prompt, "simple-test", "prompt should contain environment name")
				assert.Contains(t, prompt, "\033[0m", "prompt should reset color")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := GenerateColoredPrompt(tt.envName, tt.shellType)
			require.NotEmpty(t, prompt, "prompt should not be empty")
			tt.checks(t, prompt)
		})
	}
}

func TestGetDarkModeColor(t *testing.T) {
	// Test that we get a valid color code
	color := GetDarkModeColor()

	// Should be an ANSI color code
	assert.True(t, strings.HasPrefix(color, "\033["), "color should start with ANSI escape sequence")
	assert.True(t, strings.HasSuffix(color, "m"), "color should end with 'm'")

	// Should be one of our predefined dark-mode friendly colors
	validColors := []string{
		"\033[38;5;39m",  // Bright blue
		"\033[38;5;46m",  // Bright green
		"\033[38;5;208m", // Orange
		"\033[38;5;99m",  // Purple
		"\033[38;5;87m",  // Cyan
		"\033[38;5;226m", // Yellow
		"\033[38;5;201m", // Magenta
		"\033[38;5;51m",  // Light blue
		"\033[38;5;118m", // Light green
		"\033[38;5;214m", // Light orange
	}

	found := false
	for _, valid := range validColors {
		if color == valid {
			found = true
			break
		}
	}
	assert.True(t, found, "color should be one of the predefined dark-mode friendly colors")
}

func TestColorConsistencyForEnvironment(t *testing.T) {
	// Same environment name should produce same color (deterministic)
	envName := "test-project"

	color1 := GetColorForEnvironment(envName)
	color2 := GetColorForEnvironment(envName)
	color3 := GetColorForEnvironment(envName)

	assert.Equal(t, color1, color2, "same environment should produce same color")
	assert.Equal(t, color2, color3, "color should be deterministic")

	// Different environments should (likely) produce different colors
	differentEnv := "another-project"
	differentColor := GetColorForEnvironment(differentEnv)

	// Note: There's a small chance they could be the same due to hash collision,
	// but it's unlikely with a reasonable number of colors
	_ = differentColor // We won't assert they're different due to possible collision
}

