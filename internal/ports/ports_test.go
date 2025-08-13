package ports

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindFreePort(t *testing.T) {
	// Test: Should find a free port
	port := FindFreePort(30000, 40000)
	assert.Greater(t, port, 30000)
	assert.Less(t, port, 40000)

	// Test: Port should actually be free
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	assert.NoError(t, err)
	ln.Close()
}

func TestPortPersistence(t *testing.T) {
	tmpDir := t.TempDir()
	pm := NewPortManager(tmpDir)

	// Test: Assign and persist
	port1 := pm.GetPort(3000)
	assert.Greater(t, port1, 30000)

	// Test: Same port on reload
	pm2 := NewPortManager(tmpDir)
	port2 := pm2.GetPort(3000)
	assert.Equal(t, port1, port2)
}

func TestPortConflict(t *testing.T) {
	// Start a server on a port
	ln, _ := net.Listen("tcp", ":0") // Use :0 to get any free port
	defer ln.Close()
	
	// Get the actual port that was assigned
	addr := ln.Addr().(*net.TCPAddr)
	busyPort := addr.Port

	// Test: Should detect port is in use
	assert.False(t, IsPortAvailable(busyPort))

	// Test: Should skip busy port
	pm := NewPortManager(t.TempDir())
	pm.SetRange(busyPort, busyPort+10) // Narrow range starting with busy port
	port := pm.GetPort(3000)
	assert.NotEqual(t, busyPort, port)
	assert.Greater(t, port, busyPort)
}

func TestIsPortAvailable(t *testing.T) {
	// Test: High port should be available
	assert.True(t, IsPortAvailable(39999))
	
	// Test: Common ports might be in use (skip if they are)
	if IsPortAvailable(80) {
		t.Skip("Port 80 is available, skipping test")
	}
}