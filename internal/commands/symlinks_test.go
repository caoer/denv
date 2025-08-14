package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/caoer/denv/internal/paths"
	"github.com/caoer/denv/internal/testutil"
)

func TestCreateSymlinks(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "symlinktest")
	_ = os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/symlinktest.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")

	// Enter environment (should create symlinks)
	err := Enter("test")
	assert.NoError(t, err)

	// Test: .denv directory should exist
	denvDir := filepath.Join(tmpProject, ".denv")
	assert.DirExists(t, denvDir)

	// Test: environment symlink should exist with star prefix
	envLink := filepath.Join(denvDir, "*symlinktest-test")
	assert.FileExists(t, envLink)
	
	// Check symlink target
	target, err := os.Readlink(envLink)
	assert.NoError(t, err)
	expectedEnvPath := paths.EnvironmentPath("symlinktest", "test")
	assert.Equal(t, expectedEnvPath, target)

	// Test: project symlink should exist with project name
	projectLink := filepath.Join(denvDir, "symlinktest")
	assert.FileExists(t, projectLink)
	
	target2, err := os.Readlink(projectLink)
	assert.NoError(t, err)
	expectedProjectPath := paths.ProjectPath("symlinktest")
	assert.Equal(t, expectedProjectPath, target2)
}

func TestUpdateSymlinksOnEnvironmentChange(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "symlinkswitchtest")
	_ = os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/symlinkswitchtest.git")

	_ = os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")

	// Enter first environment
	err := Enter("dev")
	assert.NoError(t, err)

	// Check environment symlink points to dev
	envLink := filepath.Join(tmpProject, ".denv", "*symlinkswitchtest-dev")
	target1, _ := os.Readlink(envLink)
	assert.Contains(t, target1, "symlinkswitchtest-dev")

	// Enter different environment
	err = Enter("prod")
	assert.NoError(t, err)

	// Check environment symlink now points to prod
	envLink2 := filepath.Join(tmpProject, ".denv", "*symlinkswitchtest-prod")
	target2, _ := os.Readlink(envLink2)
	assert.Contains(t, target2, "symlinkswitchtest-prod")
}

// TestSymlinksInGitignore is intentionally commented out because
// the current implementation doesn't modify .gitignore files.
// This is a design choice to let users decide whether to ignore .denv/
// func TestSymlinksInGitignore(t *testing.T) {
// 	// The implementation intentionally doesn't modify .gitignore
// 	// See enter.go:204 - "Don't modify .gitignore - let the user decide whether to ignore .denv/"
// }