package ports

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type PortManager struct {
	dir      string
	minPort  int
	maxPort  int
	mu       sync.Mutex
	mappings map[int]int
}

func NewPortManager(dir string) *PortManager {
	pm := &PortManager{
		dir:      dir,
		minPort:  30000,
		maxPort:  39999,
		mappings: make(map[int]int),
	}
	pm.load()
	return pm
}

func (pm *PortManager) SetRange(min, max int) {
	pm.minPort = min
	pm.maxPort = max
}

func (pm *PortManager) GetPort(originalPort int) int {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if we already have a mapping
	if mapped, ok := pm.mappings[originalPort]; ok {
		// Verify it's still available
		if IsPortAvailable(mapped) {
			return mapped
		}
	}

	// Find a new free port
	newPort := FindFreePort(pm.minPort, pm.maxPort)
	pm.mappings[originalPort] = newPort
	pm.save()
	return newPort
}

func (pm *PortManager) load() {
	portFile := filepath.Join(pm.dir, "ports.json")
	data, err := os.ReadFile(portFile)
	if err != nil {
		return
	}
	json.Unmarshal(data, &pm.mappings)
}

func (pm *PortManager) save() {
	portFile := filepath.Join(pm.dir, "ports.json")
	data, _ := json.MarshalIndent(pm.mappings, "", "  ")
	os.WriteFile(portFile, data, 0644)
}

func FindFreePort(minPort, maxPort int) int {
	rand.Seed(time.Now().UnixNano())
	
	// Try random ports in range
	for i := 0; i < 1000; i++ {
		port := minPort + rand.Intn(maxPort-minPort)
		if IsPortAvailable(port) {
			return port
		}
	}
	
	// Fallback: scan sequentially
	for port := minPort; port <= maxPort; port++ {
		if IsPortAvailable(port) {
			return port
		}
	}
	
	return 0
}

func IsPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}