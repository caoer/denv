package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/testutil"
)

func TestEnterWithDifferentShells(t *testing.T) {
	tests := []struct {
		shell string
	}{
		{"/bin/bash"},
		{"/bin/zsh"},
		{"/usr/bin/fish"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			// Setup
			tmpDir := t.TempDir()
			tmpProject := filepath.Join(t.TempDir(), "shelltest")
			_ = os.MkdirAll(tmpProject, 0755)

			testutil.RunCmd(t, tmpProject, "git", "init")
			testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/shelltest.git")

			_ = os.Chdir(tmpProject)
			os.Setenv("DENV_HOME", tmpDir)
			os.Setenv("DENV_TEST_MODE", "1")
			os.Setenv("SHELL", tt.shell)

			// Test: Should work with different shells
			err := Enter("test")
			assert.NoError(t, err)

			// Verify environment was created
			envPath := filepath.Join(tmpDir, "shelltest-test")
			assert.DirExists(t, envPath)

			// Verify runtime
			runtime, err := environment.LoadRuntime(envPath)
			assert.NoError(t, err)
			assert.NotNil(t, runtime)
		})
	}
}

func TestEnterPreventNestedEnvironments(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "nestedtest")
	_ = os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/nestedtest.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")
	
	// Set DENV_ENV_NAME to simulate being inside an environment
	os.Setenv("DENV_ENV_NAME", "existing-env")
	defer os.Unsetenv("DENV_ENV_NAME")

	// Test: Should fail when trying to enter a new environment
	err := Enter("test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already in")
	assert.Contains(t, err.Error(), "existing-env")
	
	// Verify no new environment was created
	envPath := filepath.Join(tmpDir, "nestedtest-test")
	assert.NoDirExists(t, envPath)
}