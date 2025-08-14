package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/testutil"
)

func TestRm_BasicFunctionality(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "testproject")
	os.MkdirAll(tmpProject, 0755)

	// Initialize git repo
	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/testproject.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")

	// Create an environment first
	err := Enter("test-env")
	assert.NoError(t, err)

	envPath := filepath.Join(tmpDir, "testproject-test-env")
	assert.DirExists(t, envPath)

	// Test: Remove environment
	err = Rm("test-env", false) // false = not --all flag
	assert.NoError(t, err)
	assert.NoDirExists(t, envPath)
}

func TestRm_RequiresEnvironmentName(t *testing.T) {
	// Test that Rm requires an environment name when not using --all
	err := Rm("", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment name required")
}

func TestRm_NonExistentEnvironment(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "testproject")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/testproject.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)

	// Test: Remove non-existent environment
	err := Rm("nonexistent", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestRm_WithActiveSessions(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "testproject")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/testproject.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)

	// Create environment
	err := Enter("test-env")
	assert.NoError(t, err)

	envPath := filepath.Join(tmpDir, "testproject-test-env")
	
	// Simulate active session by adding a fake session with current PID
	runtime, err := environment.LoadRuntime(envPath)
	assert.NoError(t, err)
	
	runtime.Sessions["fake-session"] = environment.Session{
		ID:  "fake-session",
		PID: os.Getpid(), // Current process PID (will be active)
	}
	err = environment.SaveRuntime(envPath, runtime)
	assert.NoError(t, err)

	// Test: Should fail to remove environment with active sessions
	err = Rm("test-env", false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "active session")
	assert.DirExists(t, envPath) // Should still exist
}

func TestRm_AllFlag_RemovesOnlyInactiveEnvironments(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "testproject")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/testproject.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")

	// Create multiple environments
	err := Enter("inactive1")
	assert.NoError(t, err)
	
	err = Enter("inactive2")
	assert.NoError(t, err)
	
	err = Enter("active")
	assert.NoError(t, err)

	// Make one environment have an active session
	activeEnvPath := filepath.Join(tmpDir, "testproject-active")
	runtime, err := environment.LoadRuntime(activeEnvPath)
	assert.NoError(t, err)
	
	runtime.Sessions["active-session"] = environment.Session{
		ID:  "active-session",
		PID: os.Getpid(), // Current process PID (will be active)
	}
	err = environment.SaveRuntime(activeEnvPath, runtime)
	assert.NoError(t, err)

	// Test: Remove all inactive environments
	err = Rm("", true) // true = --all flag
	assert.NoError(t, err)

	// Verify only inactive environments were removed
	assert.NoDirExists(t, filepath.Join(tmpDir, "testproject-inactive1"))
	assert.NoDirExists(t, filepath.Join(tmpDir, "testproject-inactive2"))
	assert.DirExists(t, filepath.Join(tmpDir, "testproject-active")) // Should still exist
}

func TestRm_AllFlag_NoEnvironments(t *testing.T) {
	// Setup empty environment
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "testproject")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/testproject.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)

	// Test: Remove all when no environments exist
	err := Rm("", true) // true = --all flag
	assert.NoError(t, err) // Should not error
}

func TestRm_AllFlag_AllEnvironmentsActive(t *testing.T) {
	// Setup test environment
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "testproject")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/testproject.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")

	// Create environments and make them all active
	envNames := []string{"active1", "active2"}
	for _, name := range envNames {
		err := Enter(name)
		assert.NoError(t, err)

		envPath := filepath.Join(tmpDir, "testproject-"+name)
		runtime, err := environment.LoadRuntime(envPath)
		assert.NoError(t, err)
		
		runtime.Sessions["session-"+name] = environment.Session{
			ID:  "session-" + name,
			PID: os.Getpid(),
		}
		err = environment.SaveRuntime(envPath, runtime)
		assert.NoError(t, err)
	}

	// Test: Remove all when all environments are active
	err := Rm("", true) // true = --all flag
	assert.NoError(t, err)

	// Verify all environments still exist
	for _, name := range envNames {
		assert.DirExists(t, filepath.Join(tmpDir, "testproject-"+name))
	}
}