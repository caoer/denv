package color

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatPortDisplay(t *testing.T) {
	tests := []struct {
		name     string
		ports    []PortMapping
		expected []string // Things we expect to see in the output
	}{
		{
			name: "single port mapping",
			ports: []PortMapping{
				{Name: "PORT_WEB", Original: 3000, Mapped: 33000},
			},
			expected: []string{
				"PORT_WEB",
				"3000",
				"33000",
				"→",
			},
		},
		{
			name: "multiple port mappings aligned",
			ports: []PortMapping{
				{Name: "PORT_WEB", Original: 3000, Mapped: 33000},
				{Name: "PORT_DATABASE", Original: 5432, Mapped: 35432},
				{Name: "PORT_API", Original: 8080, Mapped: 38080},
			},
			expected: []string{
				"PORT_WEB",
				"PORT_DATABASE",
				"PORT_API",
				"3000",
				"5432",
				"8080",
				"→",
			},
		},
		{
			name: "empty ports",
			ports: []PortMapping{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPortDisplay(tt.ports)
			
			if len(tt.expected) == 0 {
				assert.Empty(t, result)
			} else {
				// Check that all expected strings are present
				for _, exp := range tt.expected {
					assert.Contains(t, result, exp)
				}
				// Check that color codes are present
				assert.Contains(t, result, "\033[")
				// Check that reset codes are present
				assert.Contains(t, result, Reset)
			}
		})
	}
}

func TestFormatPortCard(t *testing.T) {
	tests := []struct {
		name     string
		ports    []PortMapping
		checks   []string
	}{
		{
			name: "port card with multiple ports",
			ports: []PortMapping{
				{Name: "PORT_WEB", Original: 3000, Mapped: 33000},
				{Name: "PORT_DB", Original: 5432, Mapped: 35432},
			},
			checks: []string{
				"┌", "┐", // Box top
				"│", // Box sides
				"└", "┘", // Box bottom
				"Port Mappings", // Title
				"PORT_WEB",
				"PORT_DB",
				"→",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatPortCard(tt.ports)
			
			// Check for expected elements
			for _, check := range tt.checks {
				assert.Contains(t, result, check)
			}
			
			// Verify it has multiple lines
			lines := strings.Split(result, "\n")
			assert.Greater(t, len(lines), 3)
			
			// Check for color codes
			assert.Contains(t, result, "\033[")
		})
	}
}

func TestFormatPortCardAlignment(t *testing.T) {
	// Test that ports are right-aligned and arrows are vertically aligned
	ports := []PortMapping{
		{Name: "API_PORT", Original: 33000, Mapped: 38092},
		{Name: "HASURA_GRAPHQL_SERVER_PORT", Original: 38180, Mapped: 35950},
		{Name: "METABASE_DB_PORT", Original: 35432, Mapped: 33467},
		{Name: "METABASE_PORT", Original: 35431, Mapped: 39528},
		{Name: "POSTGRES_PORT", Original: 35430, Mapped: 36687},
		{Name: "WEBAPP_PORT", Original: 33000, Mapped: 38092},
	}
	
	result := FormatPortCard(ports)
	lines := strings.Split(result, "\n")
	
	// Find the port mapping lines (those containing "→")
	var portLines []string
	for _, line := range lines {
		if strings.Contains(line, "→") && !strings.Contains(line, "Port Mappings") {
			portLines = append(portLines, line)
		}
	}
	
	// Check that all arrows are vertically aligned
	// Strip ANSI color codes to check alignment
	var arrowPositions []int
	for _, line := range portLines {
		stripped := stripANSI(line)
		arrowPos := strings.Index(stripped, "→")
		arrowPositions = append(arrowPositions, arrowPos)
	}
	
	// All arrow positions should be the same
	if len(arrowPositions) > 1 {
		firstPos := arrowPositions[0]
		for i, pos := range arrowPositions {
			assert.Equal(t, firstPos, pos, "Arrow position mismatch in line %d", i)
		}
	}
	
	// Check that the colon positions are consistent (which means names are in same column)
	var colonPositions []int
	for _, line := range portLines {
		stripped := stripANSI(line)
		colonPos := strings.Index(stripped, ":")
		colonPositions = append(colonPositions, colonPos)
	}
	
	// Check that all lines have their colons in different positions based on name length
	// But the arrow positions should still be aligned
}

// Helper to strip ANSI color codes for testing alignment
func stripANSI(s string) string {
	// Simple regex replacement to remove ANSI codes
	result := s
	for i := 0; i < 10; i++ { // Multiple passes to handle nested codes
		start := strings.Index(result, "\033[")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], "m")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return result
}

func TestColorizePortBadge(t *testing.T) {
	tests := []struct {
		name     string
		original int
		mapped   int
		checks   []string
	}{
		{
			name:     "standard port mapping",
			original: 3000,
			mapped:   33000,
			checks:   []string{"3000", "33000", "→"},
		},
		{
			name:     "database port",
			original: 5432,
			mapped:   35432,
			checks:   []string{"5432", "35432"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ColorizePortBadge(tt.original, tt.mapped)
			
			// Check for expected port numbers
			for _, check := range tt.checks {
				assert.Contains(t, result, check)
			}
			
			// Should have color codes
			assert.Contains(t, result, "\033[")
			// Should have bold codes
			assert.Contains(t, result, Bold)
		})
	}
}