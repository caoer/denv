package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zitao/denv/internal/environment"
	"github.com/zitao/denv/internal/testutil"
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
			os.MkdirAll(tmpProject, 0755)

			testutil.RunCmd(t, tmpProject, "git", "init")
			testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/shelltest.git")

			os.Chdir(tmpProject)
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