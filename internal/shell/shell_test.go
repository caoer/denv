package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		shellPath    string
		expectedType ShellType
		expectedName string
	}{
		{"/bin/bash", Bash, "bash"},
		{"/usr/bin/bash", Bash, "bash"},
		{"/bin/zsh", Zsh, "zsh"},
		{"/usr/bin/zsh", Zsh, "zsh"},
		{"/usr/local/bin/fish", Fish, "fish"},
		{"/bin/sh", Sh, "sh"},
		{"", Bash, "bash"}, // Default to bash
	}

	for _, tt := range tests {
		t.Run(tt.shellPath, func(t *testing.T) {
			shellType, name := DetectShell(tt.shellPath)
			assert.Equal(t, tt.expectedType, shellType)
			assert.Equal(t, tt.expectedName, name)
		})
	}
}

func TestShellCommand(t *testing.T) {
	envScript := "/tmp/test-env.sh"

	tests := []struct {
		shellType ShellType
		expected  []string
	}{
		{
			Bash,
			[]string{"bash", "--init-file", envScript},
		},
		{
			Zsh,
			[]string{"zsh", "-c", "source /tmp/test-env.sh && exec zsh"},
		},
		{
			Fish,
			[]string{"fish", "-c", "source /tmp/test-env.sh; and exec fish"},
		},
		{
			Sh,
			[]string{"sh", "-c", ". /tmp/test-env.sh && exec sh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.shellType.String(), func(t *testing.T) {
			args := GetShellCommand(tt.shellType, envScript)
			assert.Equal(t, tt.expected, args)
		})
	}
}

func TestGenerateShellWrapper(t *testing.T) {
	env := map[string]string{
		"TEST_VAR":  "value",
		"PORT_3000": "33000",
	}

	tests := []struct {
		shellType ShellType
		contains  []string
	}{
		{
			Bash,
			[]string{"export TEST_VAR=", "trap", "cleanup"},
		},
		{
			Zsh,
			[]string{"export TEST_VAR=", "trap", "cleanup"},
		},
		{
			Fish,
			[]string{"set -x TEST_VAR", "function cleanup"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.shellType.String(), func(t *testing.T) {
			script := GenerateShellWrapper(tt.shellType, env)
			for _, expected := range tt.contains {
				assert.Contains(t, script, expected)
			}
		})
	}
}
