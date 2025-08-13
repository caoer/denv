package environment

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveLoadRuntime(t *testing.T) {
	tmpDir := t.TempDir()
	runtime := &Runtime{
		Created:     time.Now(),
		Project:     "myproject",
		Environment: "default",
		Ports: map[int]int{
			3000: 33000,
			5432: 35432,
		},
		Overrides: map[string]Override{
			"DATABASE_URL": {
				Original: "postgres://localhost:5432/db",
				Current:  "postgres://localhost:35432/db",
				Rule:     "rewrite_ports",
			},
		},
		Sessions: map[string]Session{},
	}

	// Test: Save and load
	err := SaveRuntime(tmpDir, runtime)
	assert.NoError(t, err)

	loaded, err := LoadRuntime(tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, runtime.Project, loaded.Project)
	assert.Equal(t, runtime.Environment, loaded.Environment)
	assert.Equal(t, runtime.Ports[3000], loaded.Ports[3000])
	assert.Equal(t, runtime.Overrides["DATABASE_URL"].Current, loaded.Overrides["DATABASE_URL"].Current)
}

func TestLoadRuntimeNotExists(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Test: Should return nil when file doesn't exist
	runtime, err := LoadRuntime(tmpDir)
	assert.NoError(t, err)
	assert.Nil(t, runtime)
}

func TestRuntimePersistence(t *testing.T) {
	// Test that runtime.json persists port configurations across sessions
	tmpDir := t.TempDir()

	// Create initial runtime with port mappings
	runtime1 := &Runtime{
		Created:     time.Now(),
		Project:     "test-project",
		Environment: "dev",
		Ports: map[int]int{
			3000: 31234,
			8080: 38765,
		},
		Overrides: make(map[string]Override),
		Sessions:  make(map[string]Session),
	}

	// Save runtime
	err := SaveRuntime(tmpDir, runtime1)
	require.NoError(t, err)

	// Load runtime - should get the same port mappings
	runtime2, err := LoadRuntime(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, runtime2)

	// Verify ports are preserved
	assert.Equal(t, runtime1.Ports, runtime2.Ports)
	assert.Equal(t, 31234, runtime2.Ports[3000])
	assert.Equal(t, 38765, runtime2.Ports[8080])
}

func TestRuntimePortsNotOverwritten(t *testing.T) {
	// Test that existing port mappings in runtime.json are not overwritten
	tmpDir := t.TempDir()

	// Create initial runtime with specific port mappings
	existingRuntime := &Runtime{
		Created:     time.Now().Add(-24 * time.Hour), // Created yesterday
		Project:     "shared-project",
		Environment: "staging",
		Ports: map[int]int{
			3000: 30001, // Specific mapping
			5432: 35432, // Database port
		},
		Overrides: map[string]Override{
			"DB_HOST": {
				Original: "localhost",
				Current:  "db.staging",
				Rule:     "staging-db",
			},
		},
		Sessions: make(map[string]Session),
	}

	// Save the existing runtime
	err := SaveRuntime(tmpDir, existingRuntime)
	require.NoError(t, err)

	// Simulate a new session loading the runtime
	loadedRuntime, err := LoadRuntime(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, loadedRuntime)

	// The loaded runtime should have the same port mappings
	assert.Equal(t, existingRuntime.Ports[3000], loadedRuntime.Ports[3000])
	assert.Equal(t, existingRuntime.Ports[5432], loadedRuntime.Ports[5432])
	assert.Equal(t, 30001, loadedRuntime.Ports[3000])
	assert.Equal(t, 35432, loadedRuntime.Ports[5432])

	// Existing overrides should also be preserved
	assert.Equal(t, existingRuntime.Overrides["DB_HOST"], loadedRuntime.Overrides["DB_HOST"])
}

func TestRuntimeSessionManagement(t *testing.T) {
	// Test that sessions can be added without affecting port mappings
	tmpDir := t.TempDir()

	// Create runtime with existing data
	runtime := &Runtime{
		Created:     time.Now(),
		Project:     "multi-user-project",
		Environment: "prod",
		Ports: map[int]int{
			8080: 38080,
			9090: 39090,
		},
		Overrides: make(map[string]Override),
		Sessions:  make(map[string]Session),
	}

	// Save initial runtime
	err := SaveRuntime(tmpDir, runtime)
	require.NoError(t, err)

	// Load runtime and add a new session
	loaded, err := LoadRuntime(tmpDir)
	require.NoError(t, err)

	// Add new session
	loaded.Sessions["session-123"] = Session{
		ID:      "session-123",
		PID:     12345,
		Started: time.Now(),
		TTY:     "/dev/pts/0",
	}

	// Save with new session
	err = SaveRuntime(tmpDir, loaded)
	require.NoError(t, err)

	// Load again and verify ports are unchanged
	final, err := LoadRuntime(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, runtime.Ports, final.Ports)
	assert.Equal(t, 38080, final.Ports[8080])
	assert.Equal(t, 39090, final.Ports[9090])
	assert.Contains(t, final.Sessions, "session-123")
}

func TestRuntimeFilePermissions(t *testing.T) {
	// Test that runtime.json is created with proper permissions
	tmpDir := t.TempDir()

	runtime := NewRuntime("test", "dev")
	runtime.Ports[3000] = 33000

	err := SaveRuntime(tmpDir, runtime)
	require.NoError(t, err)

	// Check file permissions
	runtimePath := filepath.Join(tmpDir, "runtime.json")
	info, err := os.Stat(runtimePath)
	require.NoError(t, err)

	// Should be readable by all users (for multi-user scenarios)
	mode := info.Mode()
	assert.Equal(t, os.FileMode(0644), mode.Perm())
}

func TestRuntimePortMerging(t *testing.T) {
	// Test that new ports can be added without losing existing ones
	tmpDir := t.TempDir()

	// Create initial runtime with some ports
	runtime := &Runtime{
		Created:     time.Now(),
		Project:     "expandable-project",
		Environment: "dev",
		Ports: map[int]int{
			3000: 33000,
		},
		Overrides: make(map[string]Override),
		Sessions:  make(map[string]Session),
	}

	err := SaveRuntime(tmpDir, runtime)
	require.NoError(t, err)

	// Load and add more ports
	loaded, err := LoadRuntime(tmpDir)
	require.NoError(t, err)

	// Add new port mapping (simulating a new service being added)
	loaded.Ports[4000] = 34000

	err = SaveRuntime(tmpDir, loaded)
	require.NoError(t, err)

	// Load again and verify both ports exist
	final, err := LoadRuntime(tmpDir)
	require.NoError(t, err)

	assert.Equal(t, 33000, final.Ports[3000])
	assert.Equal(t, 34000, final.Ports[4000])
	assert.Len(t, final.Ports, 2)
}