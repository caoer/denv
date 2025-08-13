package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/caoer/denv/internal/config"
	"github.com/caoer/denv/internal/paths"
	"github.com/caoer/denv/internal/testutil"
)

func TestProjectShow(t *testing.T) {
	// Setup
	tmpProject := filepath.Join(t.TempDir(), "projectshow")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/projectshow.git")

	os.Chdir(tmpProject)

	// Test: Show current project name
	var output bytes.Buffer
	err := Project("", &output)
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "projectshow")
}

func TestProjectRename(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "projectrename")
	os.MkdirAll(tmpProject, 0755)
	os.Setenv("DENV_HOME", tmpDir)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/projectrename.git")

	os.Chdir(tmpProject)

	// Test: Rename project
	var output bytes.Buffer
	err := Project("rename custom-name", &output)
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "renamed")
	assert.Contains(t, output.String(), "custom-name")

	// Verify config was updated
	configPath := filepath.Join(paths.DenvHome(), "config.yaml")
	cfg, err := config.LoadConfig(configPath)
	assert.NoError(t, err)
	
	// Debug: print what we have
	t.Logf("Config projects: %v", cfg.Projects)
	t.Logf("Looking for path: %s", tmpProject)
	
	// The project saves based on cwd, not tmpProject
	found := false
	for _, name := range cfg.Projects {
		if name == "custom-name" {
			found = true
			break
		}
	}
	assert.True(t, found, "custom-name should be in config")
}

func TestProjectUnset(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "projectunset")
	os.MkdirAll(tmpProject, 0755)
	os.Setenv("DENV_HOME", tmpDir)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/projectunset.git")

	os.Chdir(tmpProject)

	// First set a custom name using current working directory (not tmpProject)
	configPath := filepath.Join(paths.DenvHome(), "config.yaml")
	defaultCfg, _ := config.LoadConfig(configPath)
	
	// Get actual current working directory
	cwd, _ := os.Getwd()
	
	cfg := &config.Config{
		Projects: map[string]string{
			cwd: "custom-override",
		},
		Patterns: defaultCfg.Patterns,
	}
	_ = config.SaveConfig(configPath, cfg)

	// Test: Unset project override
	var output bytes.Buffer
	err := Project("unset", &output)
	assert.NoError(t, err)
	assert.Contains(t, output.String(), "removed")

	// Verify config was updated
	cfg2, err := config.LoadConfig(configPath)
	assert.NoError(t, err)
	
	// Should not have the override anymore
	found := false
	for _, name := range cfg2.Projects {
		if name == "custom-override" {
			found = true
			break
		}
	}
	assert.False(t, found, "custom-override should be removed from config")
}