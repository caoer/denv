package commands

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zitao/denv/internal/environment"
)

func TestEnterRespectsExistingRuntimePorts(t *testing.T) {
	// Test that enter command respects existing runtime.json port mappings
	tmpDir := t.TempDir()
	os.Setenv("DENV_HOME", tmpDir)
	defer os.Unsetenv("DENV_HOME")
	os.Setenv("DENV_TEST_MODE", "1")
	defer os.Unsetenv("DENV_TEST_MODE")

	// Create a test project directory
	projectDir := filepath.Join(tmpDir, "testproject")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	
	// Initialize git repo
	gitDir := filepath.Join(projectDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))
	
	// Create git config with remote
	gitConfig := `[remote "origin"]
	url = https://github.com/test/testproject.git`
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "config"), []byte(gitConfig), 0644))

	// Change to project directory
	oldCwd, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(oldCwd)

	// Set environment variables that use ports
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/db")
	os.Setenv("API_PORT", "3000")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("API_PORT")

	// First enter - should create runtime with new port mappings
	err := Enter("test")
	require.NoError(t, err)

	// Load the runtime to get the initial port mappings
	envPath := filepath.Join(tmpDir, "testproject-test")
	runtime1, err := environment.LoadRuntime(envPath)
	require.NoError(t, err)
	require.NotNil(t, runtime1)

	// Store initial port mappings
	initialPort3000 := runtime1.Ports[3000]
	initialPort5432 := runtime1.Ports[5432]

	// Verify ports were allocated
	assert.NotEqual(t, 0, initialPort3000)
	assert.NotEqual(t, 0, initialPort5432)
	assert.GreaterOrEqual(t, initialPort3000, 30000)
	assert.LessOrEqual(t, initialPort3000, 39999)
	assert.GreaterOrEqual(t, initialPort5432, 30000)
	assert.LessOrEqual(t, initialPort5432, 39999)

	// Simulate another user/session entering the same environment
	// Clear the session to simulate a new entry
	runtime1.Sessions = make(map[string]environment.Session)
	environment.SaveRuntime(envPath, runtime1)

	// Second enter - should respect existing port mappings
	err = Enter("test")
	require.NoError(t, err)

	// Load runtime again
	runtime2, err := environment.LoadRuntime(envPath)
	require.NoError(t, err)
	require.NotNil(t, runtime2)

	// Verify ports are unchanged
	assert.Equal(t, initialPort3000, runtime2.Ports[3000])
	assert.Equal(t, initialPort5432, runtime2.Ports[5432])
}

func TestEnterAddsNewPortsWithoutLosingExisting(t *testing.T) {
	// Test that new ports can be added without losing existing mappings
	tmpDir := t.TempDir()
	os.Setenv("DENV_HOME", tmpDir)
	defer os.Unsetenv("DENV_HOME")
	os.Setenv("DENV_TEST_MODE", "1")
	defer os.Unsetenv("DENV_TEST_MODE")

	// Create a test project directory
	projectDir := filepath.Join(tmpDir, "webapp")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	
	// Initialize git repo
	gitDir := filepath.Join(projectDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))
	
	// Create git config with remote
	gitConfig := `[remote "origin"]
	url = https://github.com/test/webapp.git`
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "config"), []byte(gitConfig), 0644))

	// Change to project directory
	oldCwd, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(oldCwd)

	// First enter with initial ports
	os.Setenv("API_PORT", "3000")
	defer os.Unsetenv("API_PORT")

	err := Enter("dev")
	require.NoError(t, err)

	// Load runtime
	envPath := filepath.Join(tmpDir, "webapp-dev")
	runtime1, err := environment.LoadRuntime(envPath)
	require.NoError(t, err)

	initialPort3000 := runtime1.Ports[3000]
	assert.NotEqual(t, 0, initialPort3000)

	// Clear session for next entry
	runtime1.Sessions = make(map[string]environment.Session)
	environment.SaveRuntime(envPath, runtime1)

	// Second enter with additional port
	os.Setenv("DB_PORT", "5432")
	defer os.Unsetenv("DB_PORT")

	err = Enter("dev")
	require.NoError(t, err)

	// Load runtime again
	runtime2, err := environment.LoadRuntime(envPath)
	require.NoError(t, err)

	// Original port should be unchanged
	assert.Equal(t, initialPort3000, runtime2.Ports[3000])
	
	// New port should be added
	assert.NotEqual(t, 0, runtime2.Ports[5432])
	assert.GreaterOrEqual(t, runtime2.Ports[5432], 30000)
	assert.LessOrEqual(t, runtime2.Ports[5432], 39999)
}

func TestPortsJsonAndRuntimeJsonConsistency(t *testing.T) {
	// Test that ports.json and runtime.json remain consistent
	tmpDir := t.TempDir()
	os.Setenv("DENV_HOME", tmpDir)
	defer os.Unsetenv("DENV_HOME")
	os.Setenv("DENV_TEST_MODE", "1")
	defer os.Unsetenv("DENV_TEST_MODE")

	// Create a test project directory
	projectDir := filepath.Join(tmpDir, "service")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	
	// Initialize git repo
	gitDir := filepath.Join(projectDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))
	
	// Create git config with remote
	gitConfig := `[remote "origin"]
	url = https://github.com/test/service.git`
	require.NoError(t, os.WriteFile(filepath.Join(gitDir, "config"), []byte(gitConfig), 0644))

	// Change to project directory
	oldCwd, _ := os.Getwd()
	os.Chdir(projectDir)
	defer os.Chdir(oldCwd)

	// Set environment variables
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("PORT")

	// Enter environment
	err := Enter("prod")
	require.NoError(t, err)

	// Load runtime
	envPath := filepath.Join(tmpDir, "service-prod")
	runtime, err := environment.LoadRuntime(envPath)
	require.NoError(t, err)

	// Load ports.json
	portFile := filepath.Join(envPath, "ports.json")
	data, err := os.ReadFile(portFile)
	require.NoError(t, err)

	var portMappings map[int]int
	err = json.Unmarshal(data, &portMappings)
	require.NoError(t, err)

	// Verify consistency between runtime.json and ports.json
	assert.Equal(t, runtime.Ports[8080], portMappings[8080])
}