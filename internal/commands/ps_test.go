package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zitao/denv/internal/environment"
	"github.com/zitao/denv/internal/paths"
	"github.com/zitao/denv/internal/testutil"
)

func TestPsCommand_ShowsEnvironmentVariableNames(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "pstest")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/pstest.git")

	os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)

	// Create an environment with various types of overrides
	envPath := paths.EnvironmentPath("pstest", "test")
	os.MkdirAll(envPath, 0755)
	
	runtime := &environment.Runtime{
		Project:     "pstest",
		Environment: "test",
		Ports: map[int]int{
			3000: 33000,
			5432: 35432,
			8080: 38080,
		},
		Overrides: map[string]environment.Override{
			"PORT": {
				Original: "3000",
				Current:  "33000",
				Rule:     "random_port",
			},
			"DB_PORT": {
				Original: "5432",
				Current:  "35432",
				Rule:     "random_port",
			},
			"API_PORT": {
				Original: "8080",
				Current:  "38080",
				Rule:     "random_port",
			},
			"DATABASE_URL": {
				Original: "postgres://localhost:5432/mydb",
				Current:  "postgres://localhost:35432/mydb",
				Rule:     "rewrite_ports",
			},
			"API_URL": {
				Original: "http://localhost:8080/api",
				Current:  "http://localhost:38080/api",
				Rule:     "rewrite_ports",
			},
			"HOME": {
				Original: "/Users/test",
				Current:  filepath.Join(tmpDir, "pstest-test", "home"),
				Rule:     "isolate",
			},
			"CACHE_DIR": {
				Original: "/tmp/cache",
				Current:  filepath.Join(tmpDir, "pstest-test", "cache"),
				Rule:     "isolate",
			},
		},
		Sessions: map[string]environment.Session{},
	}
	environment.SaveRuntime(envPath, runtime)

	// Test: showSpecificEnvironment should display environment variable names
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := showSpecificEnvironment("test")
	
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	
	// Check that port variable names are shown
	assert.Contains(t, output, "PORT:", "Should show PORT variable name")
	assert.Contains(t, output, "DB_PORT:", "Should show DB_PORT variable name")
	assert.Contains(t, output, "API_PORT:", "Should show API_PORT variable name")
	assert.Contains(t, output, "3000 → 33000", "Should show port mapping")
	assert.Contains(t, output, "5432 → 35432", "Should show port mapping")
	assert.Contains(t, output, "8080 → 38080", "Should show port mapping")
	
	// Check that URL rewrite variable names are shown
	assert.Contains(t, output, "DATABASE_URL:", "Should show DATABASE_URL variable name")
	assert.Contains(t, output, "API_URL:", "Should show API_URL variable name")
	assert.Contains(t, output, "postgres://localhost:5432/mydb", "Should show original DATABASE_URL")
	assert.Contains(t, output, "postgres://localhost:35432/mydb", "Should show new DATABASE_URL")
	
	// Check that isolated path variable names are shown
	assert.Contains(t, output, "HOME:", "Should show HOME variable name")
	assert.Contains(t, output, "CACHE_DIR:", "Should show CACHE_DIR variable name")
	assert.Contains(t, output, "/Users/test", "Should show original HOME")
	assert.Contains(t, output, filepath.Join(tmpDir, "pstest-test", "home"), "Should show new HOME")
	assert.Contains(t, output, "/tmp/cache", "Should show original CACHE_DIR")
	assert.Contains(t, output, filepath.Join(tmpDir, "pstest-test", "cache"), "Should show new CACHE_DIR")
	
	// Check section headers are present
	assert.Contains(t, output, "[Port Variables]", "Should have Port Variables section")
	assert.Contains(t, output, "[URL/Connection String Rewrites]", "Should have URL Rewrites section")
	assert.Contains(t, output, "[Isolated Paths]", "Should have Isolated Paths section")
}

func TestPsCommand_PortMappingSummaryShowsVariableNames(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "pstest2")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/pstest2.git")

	os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)

	// Create an environment with port mappings
	envPath := paths.EnvironmentPath("pstest2", "test")
	os.MkdirAll(envPath, 0755)
	
	runtime := &environment.Runtime{
		Project:     "pstest2",
		Environment: "test",
		Ports: map[int]int{
			3000: 33000,
			5432: 35432,
			8080: 38080,
		},
		Overrides: map[string]environment.Override{
			"PORT": {
				Original: "3000",
				Current:  "33000",
				Rule:     "random_port",
			},
			"DB_PORT": {
				Original: "5432",
				Current:  "35432",
				Rule:     "random_port",
			},
			"API_PORT": {
				Original: "8080",
				Current:  "38080",
				Rule:     "random_port",
			},
		},
		Sessions: map[string]environment.Session{},
	}
	environment.SaveRuntime(envPath, runtime)

	// Test: Port mapping summary should show variable names
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := showSpecificEnvironment("test")
	
	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	assert.NoError(t, err)
	
	// Check that Port Mapping Summary section shows variable names next to ports
	lines := strings.Split(output, "\n")
	foundPortMapping := false
	for i, line := range lines {
		if strings.Contains(line, "Port Mapping Summary:") {
			foundPortMapping = true
			// Check the next few lines for port mappings with variable names
			for j := i + 1; j < len(lines) && j < i + 10; j++ {
				if strings.Contains(lines[j], "3000 → 33000") {
					assert.Contains(t, lines[j], "PORT", "Port 3000 mapping should show PORT variable")
				}
				if strings.Contains(lines[j], "5432 → 35432") {
					assert.Contains(t, lines[j], "DB_PORT", "Port 5432 mapping should show DB_PORT variable")
				}
				if strings.Contains(lines[j], "8080 → 38080") {
					assert.Contains(t, lines[j], "API_PORT", "Port 8080 mapping should show API_PORT variable")
				}
			}
			break
		}
	}
	assert.True(t, foundPortMapping, "Should have Port Mapping Summary section")
}