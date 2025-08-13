package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateWrapper(t *testing.T) {
	env := map[string]string{
		"DENV_ENV":          "/home/user/.denv/myproject-test",
		"DENV_PROJECT":      "/home/user/.denv/myproject",
		"DENV_SESSION":      "abc123",
		"DENV_ENV_NAME":     "test",
		"DENV_PROJECT_NAME": "myproject",
		"PORT_3000":         "33000",
		"DATABASE_URL":      "postgres://localhost:35432/db",
	}

	script := GenerateWrapper(env)

	// Test: Should contain signal traps
	assert.Contains(t, script, "trap")
	assert.Contains(t, script, "EXIT")
	assert.Contains(t, script, "SIGTERM")
	assert.Contains(t, script, "SIGINT")

	// Test: Should source hooks
	assert.Contains(t, script, "on-enter.sh")
	assert.Contains(t, script, "on-exit.sh")

	// Test: Should export environment variables
	assert.Contains(t, script, "export DENV_ENV=")
	assert.Contains(t, script, "export PORT_3000=")
}

func TestGenerateExportScript(t *testing.T) {
	env := map[string]string{
		"VAR1": "value1",
		"VAR2": "value with spaces",
		"VAR3": "value'with\"quotes",
	}

	script := GenerateExportScript(env)

	// Test: Should export all variables
	assert.Contains(t, script, "export VAR1=")
	assert.Contains(t, script, "export VAR2=")
	assert.Contains(t, script, "export VAR3=")

	// Test: Should properly quote values
	assert.Contains(t, script, `"value with spaces"`)
}

func TestEscapeShellValue(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with spaces", "with spaces"},
		{`with"quotes`, `with\"quotes`},
		{"with$var", `with\$var`},
		{"with`cmd`", "with\\`cmd\\`"},
	}

	for _, tt := range tests {
		result := escapeShellValue(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}

func TestCleanupScript(t *testing.T) {
	script := GenerateCleanupScript("/tmp/denv/sessions/abc123.lock", "/tmp/denv/project/hooks/on-exit.sh")

	// Test: Should remove lock file
	assert.Contains(t, script, "rm -f")
	assert.Contains(t, script, "abc123.lock")

	// Test: Should run exit hook if exists
	assert.Contains(t, script, "on-exit.sh")
}
