package main_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zitao/denv/internal/commands"
	"github.com/zitao/denv/internal/environment"
	"github.com/zitao/denv/internal/paths"
)

func TestDenvHomeEndToEnd(t *testing.T) {
	t.Run("DENV_HOME environment variable affects all commands", func(t *testing.T) {
		// Create a temporary directory for custom DENV_HOME
		tmpDir := t.TempDir()
		customHome := filepath.Join(tmpDir, "custom_denv")
		
		// Save original DENV_HOME
		originalHome := os.Getenv("DENV_HOME")
		defer os.Setenv("DENV_HOME", originalHome)
		
		// Set custom DENV_HOME
		os.Setenv("DENV_HOME", customHome)
		
		// Create the custom home directory
		require.NoError(t, os.MkdirAll(customHome, 0755))
		
		// Test 1: Verify paths use custom home
		assert.Equal(t, customHome, paths.DenvHome())
		
		// Test 2: List command uses custom home
		envs, err := commands.ListEnvironments()
		require.NoError(t, err)
		assert.Empty(t, envs, "Should have no environments in custom home")
		
		// Test 3: Create an environment in custom home
		runtime := environment.NewRuntime("test-project", "test-env")
		runtime.Created = time.Now()
		envPath := paths.EnvironmentPath("test-project", "test-env")
		require.NoError(t, os.MkdirAll(envPath, 0755))
		require.NoError(t, environment.SaveRuntime(envPath, runtime))
		
		// Test 4: Verify environment was created in custom location
		assert.FileExists(t, filepath.Join(envPath, "runtime.json"))
		assert.DirExists(t, customHome)
		assert.Contains(t, envPath, customHome)
		
		// Test 5: List should now find the environment
		envs, err = commands.ListEnvironments()
		require.NoError(t, err)
		assert.Len(t, envs, 1, "Should have one environment in custom home")
		
		// Verify no files in default location
		defaultHome := filepath.Join(os.Getenv("HOME"), ".denv")
		defaultEnvPath := filepath.Join(defaultHome, "test-project-test-env")
		assert.NoFileExists(t, filepath.Join(defaultEnvPath, "runtime.json"))
	})
	
	t.Run("DENV_HOME unset uses default location", func(t *testing.T) {
		// Save original DENV_HOME
		originalHome := os.Getenv("DENV_HOME")
		defer os.Setenv("DENV_HOME", originalHome)
		
		// Unset DENV_HOME
		os.Unsetenv("DENV_HOME")
		
		// Should use default ~/.denv
		expected := filepath.Join(os.Getenv("HOME"), ".denv")
		assert.Equal(t, expected, paths.DenvHome())
		
		// List should work without error
		_, err := commands.ListEnvironments()
		// May error if ~/.denv doesn't exist, but that's ok
		// We're just testing that it doesn't panic without DENV_HOME
		if err != nil {
			assert.Contains(t, err.Error(), ".denv")
		}
	})
}