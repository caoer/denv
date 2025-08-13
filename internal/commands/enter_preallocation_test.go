package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/caoer/denv/internal/environment"
	"github.com/caoer/denv/internal/testutil"
)

func TestEnterCommand_NoPortPreallocation(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "noprealloc")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/noprealloc.git")

	os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")
	defer os.Unsetenv("DENV_TEST_MODE")

	// Set only specific port environment variables
	os.Setenv("PORT", "3000")
	os.Setenv("DB_PORT", "5432")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("DB_PORT")

	// Enter environment
	err := Enter("test-no-prealloc")
	require.NoError(t, err)

	// Load runtime to check ports
	envPath := filepath.Join(tmpDir, "noprealloc-test-no-prealloc")
	runtime, err := environment.LoadRuntime(envPath)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	// Test: Only ports that have associated environment variables should be allocated
	// Note: May include other *_PORT variables from the test environment
	assert.LessOrEqual(t, len(runtime.Ports), 4, "Should have at most 4 ports (PORT, DB_PORT, and possibly test env ports)")
	
	// Check that only the used ports are allocated
	_, has3000 := runtime.Ports[3000]
	assert.True(t, has3000, "Port 3000 should be allocated (from PORT env var)")
	
	_, has5432 := runtime.Ports[5432]
	assert.True(t, has5432, "Port 5432 should be allocated (from DB_PORT env var)")
	
	// Check that unused ports are NOT allocated
	_, has3001 := runtime.Ports[3001]
	assert.False(t, has3001, "Port 3001 should NOT be allocated (no env var uses it)")
	
	_, has3002 := runtime.Ports[3002]
	assert.False(t, has3002, "Port 3002 should NOT be allocated (no env var uses it)")
	
	_, has8080 := runtime.Ports[8080]
	assert.False(t, has8080, "Port 8080 should NOT be allocated (no env var uses it)")
	
	_, has8081 := runtime.Ports[8081]
	assert.False(t, has8081, "Port 8081 should NOT be allocated (no env var uses it)")
	
	_, has6379 := runtime.Ports[6379]
	assert.False(t, has6379, "Port 6379 should NOT be allocated (no env var uses it)")
}

func TestEnterCommand_AllocatesPortsFromURLs(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	tmpProject := filepath.Join(t.TempDir(), "urlports")
	os.MkdirAll(tmpProject, 0755)

	testutil.RunCmd(t, tmpProject, "git", "init")
	testutil.RunCmd(t, tmpProject, "git", "remote", "add", "origin", "https://github.com/user/urlports.git")

	os.Chdir(tmpProject)
	os.Setenv("DENV_HOME", tmpDir)
	os.Setenv("DENV_TEST_MODE", "1")
	defer os.Unsetenv("DENV_TEST_MODE")

	// Set environment variables with URLs containing ports
	os.Setenv("DATABASE_URL", "postgres://localhost:5432/mydb")
	os.Setenv("REDIS_URL", "redis://localhost:6379")
	os.Setenv("API_URL", "http://localhost:8080/api")
	defer os.Unsetenv("DATABASE_URL")
	defer os.Unsetenv("REDIS_URL")
	defer os.Unsetenv("API_URL")

	// Enter environment
	err := Enter("test-url-ports")
	require.NoError(t, err)

	// Load runtime to check ports
	envPath := filepath.Join(tmpDir, "urlports-test-url-ports")
	runtime, err := environment.LoadRuntime(envPath)
	require.NoError(t, err)
	require.NotNil(t, runtime)

	// Test: Only ports referenced in URLs should be allocated (plus any test env ports)
	assert.GreaterOrEqual(t, len(runtime.Ports), 3, "Should have at least 3 ports allocated (from URLs)")
	assert.LessOrEqual(t, len(runtime.Ports), 5, "Should have at most 5 ports (URLs plus test env ports)")
	
	// Check that only the used ports are allocated
	_, has5432 := runtime.Ports[5432]
	assert.True(t, has5432, "Port 5432 should be allocated (from DATABASE_URL)")
	
	_, has6379 := runtime.Ports[6379]
	assert.True(t, has6379, "Port 6379 should be allocated (from REDIS_URL)")
	
	_, has8080 := runtime.Ports[8080]
	assert.True(t, has8080, "Port 8080 should be allocated (from API_URL)")
	
	// Check that unused ports are NOT allocated
	_, has3000 := runtime.Ports[3000]
	assert.False(t, has3000, "Port 3000 should NOT be allocated (not referenced)")
	
	_, has3001 := runtime.Ports[3001]
	assert.False(t, has3001, "Port 3001 should NOT be allocated (not referenced)")
}