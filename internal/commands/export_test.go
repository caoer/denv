package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/paths"
	"github.com/caoer/denv/internal/testutil"
)

func TestExportCommand(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "exporttest")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/exporttest.git")

	os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)

	// Create an environment with some state
	envPath := paths.EnvironmentPath("exporttest", "test")
	os.MkdirAll(envPath, 0755)
	
	runtime := &environment.Runtime{
		Project:     "exporttest",
		Environment: "test",
		Ports: map[int]int{
			3000: 33000,
			5432: 35432,
		},
	}
	environment.SaveRuntime(envPath, runtime)

	// Test: Export should output environment variables
	var output bytes.Buffer
	err := Export("test", &output)
	assert.NoError(t, err)

	result := output.String()
	assert.Contains(t, result, "export DENV_ENV=")
	assert.Contains(t, result, "export DENV_PROJECT=")
	assert.Contains(t, result, "export PORT_3000=\"33000\"")
	assert.Contains(t, result, "export PORT_5432=\"35432\"")
}

func TestExportForDirenv(t *testing.T) {
	// Setup similar to above
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "direnvtest")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/direnvtest.git")

	os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)

	// Create environment
	envPath := paths.EnvironmentPath("direnvtest", "default")
	os.MkdirAll(envPath, 0755)
	
	runtime := &environment.Runtime{
		Project:     "direnvtest",
		Environment: "default",
		Ports: map[int]int{
			3000: 33000,
		},
	}
	environment.SaveRuntime(envPath, runtime)

	// Test: Export without environment name should use default
	var output bytes.Buffer
	err := Export("", &output)
	assert.NoError(t, err)

	result := output.String()
	lines := strings.Split(result, "\n")
	
	// Should be valid shell export statements
	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, "#") {
			assert.True(t, strings.HasPrefix(line, "export "),
				"Line should start with 'export ': %s", line)
		}
	}
}