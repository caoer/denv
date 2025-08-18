package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPortCard(t *testing.T) {
	ports := []PortMapping{
		{Name: "API_PORT", Original: 3000, Mapped: 30001},
		{Name: "DB_PORT", Original: 5432, Mapped: 35432},
	}

	result := RenderPortCard(ports)

	// Should contain the port mappings
	assert.Contains(t, result, "Port Mappings")
	assert.Contains(t, result, "API_PORT")
	assert.Contains(t, result, "3000")
	assert.Contains(t, result, "30001")
	assert.Contains(t, result, "DB_PORT")
	assert.Contains(t, result, "5432")
	assert.Contains(t, result, "35432")
	
	// Should have proper box structure
	assert.True(t, strings.Contains(result, "┌") || strings.Contains(result, "╭"))
	assert.True(t, strings.Contains(result, "└") || strings.Contains(result, "╰"))
}

func TestURLCard(t *testing.T) {
	urls := []URLRewrite{
		{
			Name:     "DATABASE_URL",
			Original: "postgres://localhost:5432/mydb",
			Current:  "postgres://localhost:35432/mydb",
		},
	}

	result := RenderURLCard(urls)

	// Should contain the URL rewrites
	assert.Contains(t, result, "URL/Connection String Rewrites")
	assert.Contains(t, result, "DATABASE_URL")
	assert.Contains(t, result, "postgres://localhost:5432/mydb")
	assert.Contains(t, result, "postgres://localhost:35432/mydb")
}

func TestIsolatedPathCard(t *testing.T) {
	paths := []IsolatedPath{
		{
			Name:     "NODE_MODULES",
			Original: "~/project/node_modules",
			Current:  "~/.denv/project-dev/node_modules",
		},
	}

	result := RenderIsolatedPathCard(paths)

	// Should contain the path isolation info
	assert.Contains(t, result, "Isolated Paths")
	assert.Contains(t, result, "NODE_MODULES")
	assert.Contains(t, result, "~/project/node_modules")
	assert.Contains(t, result, "~/.denv/project-dev/node_modules")
}

func TestEmptyCards(t *testing.T) {
	// Empty inputs should return empty strings
	assert.Equal(t, "", RenderPortCard(nil))
	assert.Equal(t, "", RenderPortCard([]PortMapping{}))
	
	assert.Equal(t, "", RenderURLCard(nil))
	assert.Equal(t, "", RenderURLCard([]URLRewrite{}))
	
	assert.Equal(t, "", RenderIsolatedPathCard(nil))
	assert.Equal(t, "", RenderIsolatedPathCard([]IsolatedPath{}))
}

func TestPortColorConsistency(t *testing.T) {
	// Same port should always get same styling
	port1 := StylePort(3000)
	port2 := StylePort(3000)
	assert.Equal(t, port1, port2)
	
	// Different ports should get different styling (in most cases)
	port3 := StylePort(3001)
	// We can't guarantee they're different due to hash collisions, 
	// but we can verify the function returns something
	assert.NotEmpty(t, port3)
}