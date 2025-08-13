package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProjectSymlinks(t *testing.T) {
	// Test creating project-specific symlinks
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "myproject")
	envPath := filepath.Join(tmpDir, ".denv", "myproject-dev")
	projectPath := filepath.Join(tmpDir, ".denv", "myproject")

	// Create directories
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.MkdirAll(envPath, 0755))
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	// Create symlinks
	err := createProjectSymlinks(projectDir, envPath, projectPath, "myproject", "dev")
	require.NoError(t, err)

	// Check that .denv directory was created
	denvDir := filepath.Join(projectDir, ".denv")
	info, err := os.Stat(denvDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	// Check environment-specific symlink with project name prefix
	envLink := filepath.Join(denvDir, "*myproject-dev")
	target, err := os.Readlink(envLink)
	require.NoError(t, err)
	assert.Equal(t, envPath, target)

	// Check project symlink with project name
	projectLink := filepath.Join(denvDir, "myproject")
	target, err = os.Readlink(projectLink)
	require.NoError(t, err)
	assert.Equal(t, projectPath, target)
}

func TestReplaceExistingSymlinks(t *testing.T) {
	// Test that existing symlinks are replaced, not kept
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "project")
	oldEnvPath := filepath.Join(tmpDir, ".denv", "project-old")
	newEnvPath := filepath.Join(tmpDir, ".denv", "project-new")
	projectPath := filepath.Join(tmpDir, ".denv", "project")

	// Create directories
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.MkdirAll(oldEnvPath, 0755))
	require.NoError(t, os.MkdirAll(newEnvPath, 0755))
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	// Create initial symlinks
	denvDir := filepath.Join(projectDir, ".denv")
	require.NoError(t, os.MkdirAll(denvDir, 0755))
	
	// Old style symlinks (to be removed)
	oldCurrentLink := filepath.Join(denvDir, "current")
	require.NoError(t, os.Symlink(oldEnvPath, oldCurrentLink))
	
	oldProjectLink := filepath.Join(denvDir, "project")
	require.NoError(t, os.Symlink(projectPath, oldProjectLink))

	// Create new symlinks with new naming convention
	err := createProjectSymlinks(projectDir, newEnvPath, projectPath, "project", "new")
	require.NoError(t, err)

	// Old "current" symlink should be removed
	_, err = os.Stat(oldCurrentLink)
	assert.True(t, os.IsNotExist(err), "Old 'current' symlink should be removed")

	// Project symlink should be replaced (not removed) since project name is "project"
	// The new symlink has the same name but points to the correct target
	_, err = os.Stat(oldProjectLink)
	assert.NoError(t, err, "Project symlink should exist")

	// New environment symlink should exist with star prefix
	newEnvLink := filepath.Join(denvDir, "*project-new")
	target, err := os.Readlink(newEnvLink)
	require.NoError(t, err)
	assert.Equal(t, newEnvPath, target)

	// New project symlink should exist with project name
	newProjectLink := filepath.Join(denvDir, "project")
	target, err = os.Readlink(newProjectLink)
	require.NoError(t, err)
	assert.Equal(t, projectPath, target)
}

func TestMultipleEnvironmentSymlinks(t *testing.T) {
	// Test that multiple environment symlinks can coexist
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "webapp")
	devEnvPath := filepath.Join(tmpDir, ".denv", "webapp-dev")
	stagingEnvPath := filepath.Join(tmpDir, ".denv", "webapp-staging")
	projectPath := filepath.Join(tmpDir, ".denv", "webapp")

	// Create directories
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.MkdirAll(devEnvPath, 0755))
	require.NoError(t, os.MkdirAll(stagingEnvPath, 0755))
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	// Create dev environment symlinks
	err := createProjectSymlinks(projectDir, devEnvPath, projectPath, "webapp", "dev")
	require.NoError(t, err)

	// Create staging environment symlinks
	err = createProjectSymlinks(projectDir, stagingEnvPath, projectPath, "webapp", "staging")
	require.NoError(t, err)

	denvDir := filepath.Join(projectDir, ".denv")

	// Both environment symlinks should exist
	devLink := filepath.Join(denvDir, "*webapp-dev")
	target, err := os.Readlink(devLink)
	require.NoError(t, err)
	assert.Equal(t, devEnvPath, target)

	stagingLink := filepath.Join(denvDir, "*webapp-staging")
	target, err = os.Readlink(stagingLink)
	require.NoError(t, err)
	assert.Equal(t, stagingEnvPath, target)

	// Project symlink should still exist
	projectLink := filepath.Join(denvDir, "webapp")
	target, err = os.Readlink(projectLink)
	require.NoError(t, err)
	assert.Equal(t, projectPath, target)
}

func TestSymlinkNamingConvention(t *testing.T) {
	// Test the naming convention for symlinks
	tests := []struct {
		projectName string
		envName     string
		expectedEnv string
		expectedProj string
	}{
		{
			projectName: "myapp",
			envName:     "dev",
			expectedEnv: "*myapp-dev",
			expectedProj: "myapp",
		},
		{
			projectName: "backend-api",
			envName:     "production",
			expectedEnv: "*backend-api-production",
			expectedProj: "backend-api",
		},
		{
			projectName: "test_project",
			envName:     "feature-123",
			expectedEnv: "*test_project-feature-123",
			expectedProj: "test_project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.projectName+"-"+tt.envName, func(t *testing.T) {
			tmpDir := t.TempDir()
			projectDir := filepath.Join(tmpDir, "project")
			envPath := filepath.Join(tmpDir, ".denv", tt.projectName+"-"+tt.envName)
			projectPath := filepath.Join(tmpDir, ".denv", tt.projectName)

			// Create directories
			require.NoError(t, os.MkdirAll(projectDir, 0755))
			require.NoError(t, os.MkdirAll(envPath, 0755))
			require.NoError(t, os.MkdirAll(projectPath, 0755))

			// Create symlinks
			err := createProjectSymlinks(projectDir, envPath, projectPath, tt.projectName, tt.envName)
			require.NoError(t, err)

			denvDir := filepath.Join(projectDir, ".denv")

			// Check environment symlink name
			envLink := filepath.Join(denvDir, tt.expectedEnv)
			_, err = os.Stat(envLink)
			require.NoError(t, err, "Environment symlink %s should exist", tt.expectedEnv)

			// Check project symlink name
			projLink := filepath.Join(denvDir, tt.expectedProj)
			_, err = os.Stat(projLink)
			require.NoError(t, err, "Project symlink %s should exist", tt.expectedProj)
		})
	}
}

func TestCleanupOldSymlinks(t *testing.T) {
	// Test that old symlink naming convention is cleaned up
	tmpDir := t.TempDir()
	projectDir := filepath.Join(tmpDir, "app")
	envPath := filepath.Join(tmpDir, ".denv", "app-test")
	projectPath := filepath.Join(tmpDir, ".denv", "app")

	// Create directories
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.MkdirAll(envPath, 0755))
	require.NoError(t, os.MkdirAll(projectPath, 0755))

	denvDir := filepath.Join(projectDir, ".denv")
	require.NoError(t, os.MkdirAll(denvDir, 0755))

	// Create old-style symlinks that should be removed
	oldLinks := []string{"current", "project"}
	for _, link := range oldLinks {
		linkPath := filepath.Join(denvDir, link)
		require.NoError(t, os.Symlink("/tmp/dummy", linkPath))
	}

	// Create some other files that should be preserved
	regularFile := filepath.Join(denvDir, "config.yaml")
	require.NoError(t, os.WriteFile(regularFile, []byte("test"), 0644))

	// Create new symlinks
	err := createProjectSymlinks(projectDir, envPath, projectPath, "app", "test")
	require.NoError(t, err)

	// Old symlinks should be removed
	for _, link := range oldLinks {
		linkPath := filepath.Join(denvDir, link)
		_, err := os.Stat(linkPath)
		assert.True(t, os.IsNotExist(err), "Old symlink %s should be removed", link)
	}

	// Regular files should be preserved
	_, err = os.Stat(regularFile)
	require.NoError(t, err, "Regular files should be preserved")

	// New symlinks should exist
	assert.FileExists(t, filepath.Join(denvDir, "*app-test"))
	assert.FileExists(t, filepath.Join(denvDir, "app"))
}