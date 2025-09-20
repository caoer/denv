package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/caoer/denv/internal/environment"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionExists(t *testing.T) {
	t.Run("should detect current process as existing", func(t *testing.T) {
		// Get current process PID which definitely exists
		currentPID := os.Getpid()
		
		// This should return true for the current process
		exists := sessionExists(currentPID)
		assert.True(t, exists, "Current process should be detected as existing")
	})
	
	t.Run("should detect non-existent process as not existing", func(t *testing.T) {
		// Use an impossible PID that won't exist
		nonExistentPID := 999999999
		
		// This should return false for a non-existent process
		exists := sessionExists(nonExistentPID)
		assert.False(t, exists, "Non-existent process should be detected as not existing")
	})
	
	t.Run("should detect parent process as existing", func(t *testing.T) {
		// Get parent process PID which should exist
		parentPID := os.Getppid()
		
		// This should return true for the parent process
		exists := sessionExists(parentPID)
		assert.True(t, exists, "Parent process should be detected as existing")
	})
}

func TestListPlainFormat(t *testing.T) {
	// Create a temporary denv home for testing
	tempDir := t.TempDir()
	origHome := os.Getenv("DENV_HOME")
	t.Cleanup(func() {
		if origHome != "" {
			os.Setenv("DENV_HOME", origHome)
		} else {
			os.Unsetenv("DENV_HOME")
		}
	})
	os.Setenv("DENV_HOME", tempDir)

	t.Run("plain format outputs tab-separated values", func(t *testing.T) {
		// Create test environment directories with runtime files
		projectEnvs := []struct {
			project string
			env     string
			active  bool
			ports   int
		}{
			{"myproject", "dev", false, 3},
			{"myproject", "staging", true, 5},
			{"otherproj", "prod", false, 0},
		}

		for _, pe := range projectEnvs {
			envDir := filepath.Join(tempDir, pe.project+"-"+pe.env)
			require.NoError(t, os.MkdirAll(envDir, 0755))

			// Create runtime with test data
			runtime := &environment.Runtime{
				Created:     time.Now(),
				Project:     pe.project,
				Environment: pe.env,
				Ports:       make(map[int]int),
				Overrides:   make(map[string]environment.Override),
				Sessions:    make(map[string]environment.Session),
			}

			// Add ports if specified
			for i := 0; i < pe.ports; i++ {
				runtime.Ports[3000+i] = 30000 + i
			}

			// Add active session if specified
			if pe.active {
				runtime.Sessions["test-session"] = environment.Session{
					ID:      "test-session",
					PID:     os.Getpid(), // Use current PID so it's detected as active
					Started: time.Now(),
				}
			}

			require.NoError(t, environment.SaveRuntime(envDir, runtime))
		}

		// Capture output
		var buf bytes.Buffer
		err := ListPlain(&buf)
		require.NoError(t, err)

		output := buf.String()
		lines := strings.Split(strings.TrimSpace(output), "\n")

		// Check we have expected number of lines
		assert.Len(t, lines, 3)

		// Parse and validate each line
		expected := map[string]struct {
			project  string
			env      string
			status   string
			sessions string
			ports    string
		}{
			"myproject\tdev": {"myproject", "dev", "inactive", "0", "3"},
			"myproject\tstaging": {"myproject", "staging", "active", "1", "5"},
			"otherproj\tprod": {"otherproj", "prod", "inactive", "0", "0"},
		}

		for _, line := range lines {
			fields := strings.Split(line, "\t")
			assert.Len(t, fields, 5, "Each line should have 5 tab-separated fields")

			key := fields[0] + "\t" + fields[1]
			exp, ok := expected[key]
			assert.True(t, ok, "Unexpected environment: %s", key)

			assert.Equal(t, exp.project, fields[0], "Project name mismatch")
			assert.Equal(t, exp.env, fields[1], "Environment name mismatch")
			assert.Equal(t, exp.status, fields[2], "Status mismatch")
			assert.Equal(t, exp.sessions, fields[3], "Sessions count mismatch")
			assert.Equal(t, exp.ports, fields[4], "Ports count mismatch")
		}
	})

	t.Run("plain format with no environments", func(t *testing.T) {
		// Clear the temp directory
		entries, _ := os.ReadDir(tempDir)
		for _, entry := range entries {
			os.RemoveAll(filepath.Join(tempDir, entry.Name()))
		}

		var buf bytes.Buffer
		err := ListPlain(&buf)
		require.NoError(t, err)

		output := strings.TrimSpace(buf.String())
		assert.Empty(t, output, "Plain format should output nothing when no environments exist")
	})
}