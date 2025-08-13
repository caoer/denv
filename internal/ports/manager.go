package ports

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"sync"
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

// InitializeWithPorts sets the port mappings from an existing runtime
// This ensures that existing port mappings are respected
func (pm *PortManager) InitializeWithPorts(ports map[int]int) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Merge the provided ports with existing mappings
	// Existing mappings take precedence to maintain consistency
	for orig, mapped := range ports {
		if _, exists := pm.mappings[orig]; !exists {
			pm.mappings[orig] = mapped
		}
	}
	
	pm.save()
}

func (pm *PortManager) load() {
	portFile := filepath.Join(pm.dir, "ports.json")
	data, err := os.ReadFile(portFile)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, &pm.mappings); err != nil {
		return
	}
}

func (pm *PortManager) save() {
	portFile := filepath.Join(pm.dir, "ports.json")
	data, err := json.MarshalIndent(pm.mappings, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(portFile, data, 0644)
}

func FindFreePort(minPort, maxPort int) int {
	// rand.Seed is deprecated in Go 1.20+, auto-seeded by default
	
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