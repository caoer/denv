package ports

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortManagerRespectsExistingMappings(t *testing.T) {
	// Test that port manager doesn't overwrite existing port mappings
	tmpDir := t.TempDir()

	// Create initial port mappings file
	initialMappings := map[int]int{
		3000: 30123,
		8080: 38080,
		5432: 35432,
	}
	
	portFile := filepath.Join(tmpDir, "ports.json")
	data, _ := json.MarshalIndent(initialMappings, "", "  ")
	err := os.WriteFile(portFile, data, 0644)
	require.NoError(t, err)

	// Create port manager - should load existing mappings
	pm := NewPortManager(tmpDir)

	// Request the same ports - should get the existing mappings
	assert.Equal(t, 30123, pm.GetPort(3000))
	assert.Equal(t, 38080, pm.GetPort(8080))
	assert.Equal(t, 35432, pm.GetPort(5432))

	// Request a new port - should get a new mapping
	newPort := pm.GetPort(9000)
	assert.NotEqual(t, 0, newPort)
	assert.NotEqual(t, 9000, newPort) // Should be mapped to different port
	assert.GreaterOrEqual(t, newPort, 30000)
	assert.LessOrEqual(t, newPort, 39999)
}

func TestPortManagerPersistsMappings(t *testing.T) {
	// Test that port mappings persist across port manager instances
	tmpDir := t.TempDir()

	// First instance - create mappings
	pm1 := NewPortManager(tmpDir)
	port3000 := pm1.GetPort(3000)
	port8080 := pm1.GetPort(8080)

	// Verify ports are in the expected range
	assert.GreaterOrEqual(t, port3000, 30000)
	assert.LessOrEqual(t, port3000, 39999)
	assert.GreaterOrEqual(t, port8080, 30000)
	assert.LessOrEqual(t, port8080, 39999)

	// Second instance - should load the same mappings
	pm2 := NewPortManager(tmpDir)
	assert.Equal(t, port3000, pm2.GetPort(3000))
	assert.Equal(t, port8080, pm2.GetPort(8080))
}

func TestPortManagerWithExistingRuntime(t *testing.T) {
	// Test that port manager can be initialized with existing runtime ports
	tmpDir := t.TempDir()

	// Simulate existing runtime with port mappings
	existingPorts := map[int]int{
		3000: 31111,
		4000: 34444,
	}

	// Option 1: Create ports.json directly
	portFile := filepath.Join(tmpDir, "ports.json")
	data, _ := json.MarshalIndent(existingPorts, "", "  ")
	err := os.WriteFile(portFile, data, 0644)
	require.NoError(t, err)

	// Create port manager
	pm := NewPortManager(tmpDir)

	// Should respect existing mappings
	assert.Equal(t, 31111, pm.GetPort(3000))
	assert.Equal(t, 34444, pm.GetPort(4000))
}

func TestPortManagerInitializeWithRuntimePorts(t *testing.T) {
	// Test initializing port manager with pre-defined port mappings
	tmpDir := t.TempDir()

	pm := NewPortManager(tmpDir)
	
	// Simulate loading existing runtime ports
	runtimePorts := map[int]int{
		3000: 32222,
		5000: 35555,
		8080: 38888,
	}

	// Initialize port manager with runtime ports
	pm.InitializeWithPorts(runtimePorts)

	// Should use the initialized ports
	assert.Equal(t, 32222, pm.GetPort(3000))
	assert.Equal(t, 35555, pm.GetPort(5000))
	assert.Equal(t, 38888, pm.GetPort(8080))

	// New port should get a fresh mapping
	newPort := pm.GetPort(9090)
	assert.NotEqual(t, 0, newPort)
	assert.NotEqual(t, 9090, newPort)
}

func TestPortManagerDoesNotRandomizeExistingPorts(t *testing.T) {
	// Ensure that once a port is mapped, it doesn't get randomized on subsequent calls
	tmpDir := t.TempDir()

	pm := NewPortManager(tmpDir)
	
	// Get initial mapping
	firstMapping := pm.GetPort(3000)
	
	// Call multiple times - should always get the same mapping
	for i := 0; i < 10; i++ {
		assert.Equal(t, firstMapping, pm.GetPort(3000), "Port mapping should remain consistent")
	}

	// Create new port manager instance
	pm2 := NewPortManager(tmpDir)
	
	// Should still get the same mapping
	assert.Equal(t, firstMapping, pm2.GetPort(3000), "Port mapping should persist across instances")
}

func TestPortManagerHandlesPortConflicts(t *testing.T) {
	// Test that port manager handles conflicts when a mapped port is no longer available
	tmpDir := t.TempDir()

	// Create initial mapping where port is available
	initialMappings := map[int]int{
		3000: 30123,
	}
	
	portFile := filepath.Join(tmpDir, "ports.json")
	data, _ := json.MarshalIndent(initialMappings, "", "  ")
	err := os.WriteFile(portFile, data, 0644)
	require.NoError(t, err)

	pm := NewPortManager(tmpDir)

	// Mock the case where port 30123 is no longer available
	// Since we can't easily block a port in test, we'll test the logic
	// by verifying that if IsPortAvailable returns false, a new port is allocated
	
	// For now, just verify the existing mapping is returned when port is available
	mappedPort := pm.GetPort(3000)
	if IsPortAvailable(30123) {
		assert.Equal(t, 30123, mappedPort)
	} else {
		// If port is not available, should get a different port
		assert.NotEqual(t, 30123, mappedPort)
		assert.GreaterOrEqual(t, mappedPort, 30000)
		assert.LessOrEqual(t, mappedPort, 39999)
	}
}