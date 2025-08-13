package project

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zitao/denv/internal/config"
	"github.com/zitao/denv/internal/testutil"
)

func TestDetectGitProject(t *testing.T) {
	// Setup: Create temp git repo
	tmpDir := t.TempDir()
	testutil.RunCmd(t, tmpDir, "git", "init")
	testutil.RunCmd(t, tmpDir, "git", "remote", "add", "origin", "https://github.com/user/myproject.git")

	// Test: Should detect project name from git remote
	name, err := DetectProject(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "myproject", name)
}

func TestDetectGitWorktree(t *testing.T) {
	// Setup: Create main repo and worktree
	mainDir := filepath.Join(t.TempDir(), "main")
	worktreeDir := filepath.Join(t.TempDir(), "worktree")

	testutil.RunCmd(t, "", "mkdir", "-p", mainDir)
	testutil.RunCmd(t, mainDir, "git", "init")
	testutil.RunCmd(t, mainDir, "git", "remote", "add", "origin", "https://github.com/user/myproject.git")
	testutil.RunCmd(t, mainDir, "git", "commit", "--allow-empty", "-m", "init")
	testutil.RunCmd(t, mainDir, "git", "worktree", "add", worktreeDir)

	// Test: Both should detect same project
	mainName, _ := DetectProject(mainDir)
	worktreeName, _ := DetectProject(worktreeDir)
	assert.Equal(t, mainName, worktreeName)
	assert.Equal(t, "myproject", mainName)
}

func TestDetectFolderName(t *testing.T) {
	// Test: Should use folder name when no git
	tmpDir := filepath.Join(t.TempDir(), "testfolder")
	testutil.RunCmd(t, "", "mkdir", "-p", tmpDir)

	name, err := DetectProject(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, "testfolder", name)
}

func TestDetectWithOverride(t *testing.T) {
	// Setup: Create config with override
	cfg := &config.Config{
		Projects: map[string]string{
			"/my/path": "custom-project",
		},
	}

	// Test: Should use override
	name := DetectProjectWithConfig("/my/path", cfg)
	assert.Equal(t, "custom-project", name)
}