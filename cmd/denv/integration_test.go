package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/caoer/denv/internal/commands"
	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/testutil"
)

func TestFullWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	// Setup test environment
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "testproject")
	_ = os.MkdirAll(tmpProject, 0755)

	// Initialize git repo
	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/testproject.git")

	// Test: Enter environment
	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")

	err := commands.Enter("test-env")
	assert.NoError(t, err)

	// Test: Environment should exist
	envPath := filepath.Join(tmpDir, "testproject-test-env")
	assert.DirExists(t, envPath)

	// Test: Runtime should be saved with ports
	runtime, err := environment.LoadRuntime(envPath)
	assert.NoError(t, err)
	assert.NotNil(t, runtime)
	assert.NotEmpty(t, runtime.Ports)
	assert.Equal(t, "testproject", runtime.Project)
	assert.Equal(t, "test-env", runtime.Environment)

	// Test: List environments
	err = commands.List()
	assert.NoError(t, err)

	// Test: Remove environment (should succeed in test mode as no real session exists)
	err = commands.Rm("test-env", false)
	assert.NoError(t, err)
	assert.NoDirExists(t, envPath)
}

func TestMultipleEnvironments(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "multitest")
	_ = os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/multitest.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")
	os.Setenv("PORT", "3000") // Add a port env variable to test port allocation

	// Create multiple environments
	err := commands.Enter("dev")
	assert.NoError(t, err)

	err = commands.Enter("staging")
	assert.NoError(t, err)

	err = commands.Enter("prod")
	assert.NoError(t, err)

	// Verify all exist
	assert.DirExists(t, filepath.Join(tmpDir, "multitest-dev"))
	assert.DirExists(t, filepath.Join(tmpDir, "multitest-staging"))
	assert.DirExists(t, filepath.Join(tmpDir, "multitest-prod"))

	// Load and verify different ports
	devRuntime, _ := environment.LoadRuntime(filepath.Join(tmpDir, "multitest-dev"))
	stagingRuntime, _ := environment.LoadRuntime(filepath.Join(tmpDir, "multitest-staging"))

	// Ports should be different between environments
	assert.NotEqual(t, devRuntime.Ports[3000], stagingRuntime.Ports[3000])
}