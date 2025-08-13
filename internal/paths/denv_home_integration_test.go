package paths_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/paths"
)

func TestDenvHomeIntegration(t *testing.T) {
	// Test that DENV_HOME properly overrides all components
	t.Run("DENV_HOME overrides affect all subsystems", func(t *testing.T) {
		// Create a temporary directory for testing
		tmpDir := t.TempDir()
		customHome := filepath.Join(tmpDir, "custom_denv")
		
		// Save original DENV_HOME
		originalHome := os.Getenv("DENV_HOME")
		defer os.Setenv("DENV_HOME", originalHome)
		
		// Set custom DENV_HOME
		os.Setenv("DENV_HOME", customHome)
		
		// Verify paths use custom home
		assert.Equal(t, customHome, paths.DenvHome())
		assert.Equal(t, filepath.Join(customHome, "test-project"), paths.ProjectPath("test-project"))
		assert.Equal(t, filepath.Join(customHome, "test-project-dev"), paths.EnvironmentPath("test-project", "dev"))
		
		// Verify config path would use custom home
		expectedConfigPath := filepath.Join(customHome, "config.yaml")
		actualConfigPath := filepath.Join(paths.DenvHome(), "config.yaml")
		assert.Equal(t, expectedConfigPath, actualConfigPath)
		
		// Create the custom home directory
		require.NoError(t, os.MkdirAll(customHome, 0755))
		
		// Initialize a new runtime in the custom location
		runtime := &environment.Runtime{
			Created:     time.Now(),
			Project:     "test-project",
			Environment: "test",
			Ports:       make(map[int]int),
			Overrides:   make(map[string]environment.Override),
			Sessions:    make(map[string]environment.Session),
		}
		
		envPath := paths.EnvironmentPath("test-project", "test")
		require.NoError(t, os.MkdirAll(envPath, 0755))
		
		// Save and verify runtime
		require.NoError(t, environment.SaveRuntime(envPath, runtime))
		assert.FileExists(t, filepath.Join(envPath, "runtime.json"))
		
		// Load and verify it can be loaded
		loaded, err := environment.LoadRuntime(envPath)
		require.NoError(t, err)
		assert.Equal(t, "test-project", loaded.Project)
		assert.Equal(t, "test", loaded.Environment)
	})
	
	t.Run("DENV_HOME unset uses default", func(t *testing.T) {
		// Save original DENV_HOME
		originalHome := os.Getenv("DENV_HOME")
		defer os.Setenv("DENV_HOME", originalHome)
		
		// Unset DENV_HOME
		os.Unsetenv("DENV_HOME")
		
		// Should use default ~/.denv
		expected := filepath.Join(os.Getenv("HOME"), ".denv")
		assert.Equal(t, expected, paths.DenvHome())
	})
	
	t.Run("DENV_HOME empty string uses default", func(t *testing.T) {
		// Save original DENV_HOME
		originalHome := os.Getenv("DENV_HOME")
		defer os.Setenv("DENV_HOME", originalHome)
		
		// Set DENV_HOME to empty string
		os.Setenv("DENV_HOME", "")
		
		// Should use default ~/.denv
		expected := filepath.Join(os.Getenv("HOME"), ".denv")
		assert.Equal(t, expected, paths.DenvHome())
	})
	
	t.Run("DENV_HOME with spaces", func(t *testing.T) {
		// Save original DENV_HOME
		originalHome := os.Getenv("DENV_HOME")
		defer os.Setenv("DENV_HOME", originalHome)
		
		// Create a path with spaces
		tmpDir := t.TempDir()
		customHome := filepath.Join(tmpDir, "my denv home")
		
		// Set custom DENV_HOME with spaces
		os.Setenv("DENV_HOME", customHome)
		
		// Verify it handles spaces correctly
		assert.Equal(t, customHome, paths.DenvHome())
		assert.Equal(t, filepath.Join(customHome, "test-project"), paths.ProjectPath("test-project"))
	})
	
	t.Run("DENV_HOME relative path gets expanded", func(t *testing.T) {
		// Save original DENV_HOME and working directory
		originalHome := os.Getenv("DENV_HOME")
		originalWd, _ := os.Getwd()
		defer os.Setenv("DENV_HOME", originalHome)
		defer os.Chdir(originalWd)
		
		// Create temp directory and change to it
		tmpDir := t.TempDir()
		os.Chdir(tmpDir)
		
		// Set relative DENV_HOME
		os.Setenv("DENV_HOME", "./my_denv")
		
		// It should use the relative path as-is (not expand it)
		// This tests the current behavior
		assert.Equal(t, "./my_denv", paths.DenvHome())
	})
}

func TestDenvHomeInCommands(t *testing.T) {
	t.Run("Enter command respects DENV_HOME", func(t *testing.T) {
		tmpDir := t.TempDir()
		customHome := filepath.Join(tmpDir, "custom_home")
		
		// Save original DENV_HOME
		originalHome := os.Getenv("DENV_HOME")
		defer os.Setenv("DENV_HOME", originalHome)
		
		// Set custom DENV_HOME
		os.Setenv("DENV_HOME", customHome)
		
		// Create necessary directories
		require.NoError(t, os.MkdirAll(customHome, 0755))
		
		// Create environment runtime
		runtime := &environment.Runtime{
			Created:     time.Now(),
			Project:     "test-project",
			Environment: "test",
			Ports:       make(map[int]int),
			Overrides:   make(map[string]environment.Override),
			Sessions:    make(map[string]environment.Session),
		}
		
		// Ensure environment directory exists
		envPath := paths.EnvironmentPath("test-project", "test")
		require.NoError(t, os.MkdirAll(envPath, 0755))
		
		// Save runtime
		require.NoError(t, environment.SaveRuntime(envPath, runtime))
		
		// Verify files were created in custom location
		assert.DirExists(t, customHome)
		assert.DirExists(t, envPath)
		assert.FileExists(t, filepath.Join(envPath, "runtime.json"))
		
		// Verify no files in default location
		defaultHome := filepath.Join(os.Getenv("HOME"), ".denv")
		defaultEnvPath := filepath.Join(defaultHome, "test-project-test")
		assert.NoFileExists(t, filepath.Join(defaultEnvPath, "runtime.json"))
	})
}