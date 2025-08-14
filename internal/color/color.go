package color

import (
	"fmt"
	"hash/fnv"
	"strings"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	
	// Regular colors
	Red     = "\033[31m"
	Green   = "\033[32m"
	Yellow  = "\033[33m"
	Blue    = "\033[34m"
	Magenta = "\033[35m"
	Cyan    = "\033[36m"
	
	// Bright colors
	BrightRed     = "\033[91m"
	BrightGreen   = "\033[92m"
	BrightYellow  = "\033[93m"
	BrightBlue    = "\033[94m"
	BrightMagenta = "\033[95m"
	BrightCyan    = "\033[96m"
)

// Color palette for ports - using distinct, pleasant colors
var portColors = []string{
	Cyan,
	BrightGreen,
	BrightYellow,
	BrightBlue,
	BrightMagenta,
	BrightCyan,
	Green,
	Yellow,
	Blue,
	Magenta,
}

// PortColor returns a consistent color for a given port number
// The same port will always get the same color
func PortColor(port int) string {
	// Use FNV-1a hash for consistent distribution
	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%d", port)))
	hash := h.Sum32()
	
	// Map hash to color index (safe modulo operation)
	// #nosec G115 -- len(portColors) is a small constant, no overflow risk
	colorIndex := hash % uint32(len(portColors))
	return portColors[colorIndex]
}

// ColorizePort returns a colored string for a port number
func ColorizePort(port int) string {
	color := PortColor(port)
	return fmt.Sprintf("%s%d%s", color, port, Reset)
}

// ColorizePorts returns a colored string for port mapping (original → mapped)
func ColorizePorts(original, mapped int) string {
	origColor := PortColor(original)
	mappedColor := PortColor(mapped)
	return fmt.Sprintf("%s%d%s → %s%d%s", origColor, original, Reset, mappedColor, mapped, Reset)
}

// ColorizePortWithAlignment returns aligned and colored port mapping
func ColorizePortWithAlignment(original, mapped int, origWidth int) string {
	origColor := PortColor(original)
	mappedColor := PortColor(mapped)
	// The width calculation needs to account for the actual number width, not the color codes
	origStr := fmt.Sprintf("%d", original)
	padding := origWidth - len(origStr)
	spaces := ""
	for i := 0; i < padding; i++ {
		spaces += " "
	}
	return fmt.Sprintf("%s%s%s%s → %s%d%s", spaces, origColor, origStr, Reset, mappedColor, mapped, Reset)
}

// ColorizePortInURL colorizes port numbers within a URL string
func ColorizePortInURL(url string, port int) string {
	portStr := fmt.Sprintf(":%d", port)
	coloredPort := fmt.Sprintf(":%s%d%s", PortColor(port), port, Reset)
	return strings.ReplaceAll(url, portStr, coloredPort)
}

// PortMapping represents a port variable mapping
type PortMapping struct {
	Name     string
	Original int
	Mapped   int
}

// FormatPortDisplay formats port mappings with colors and alignment
func FormatPortDisplay(ports []PortMapping) string {
	if len(ports) == 0 {
		return ""
	}
	
	// Find max lengths for alignment
	maxNameLen := 0
	for _, pm := range ports {
		if len(pm.Name) > maxNameLen {
			maxNameLen = len(pm.Name)
		}
	}
	
	var result strings.Builder
	for _, pm := range ports {
		// Format: NAME      : 3000 → 33000 with colors
		origColor := PortColor(pm.Original)
		mappedColor := PortColor(pm.Mapped)
		result.WriteString(fmt.Sprintf("   %-*s : %s%s%d%s → %s%s%d%s\n", 
			maxNameLen, pm.Name,
			Bold, origColor, pm.Original, Reset,
			Bold, mappedColor, pm.Mapped, Reset))
	}
	
	return result.String()
}

// FormatPortCard creates a visually appealing card display for ports
func FormatPortCard(ports []PortMapping) string {
	if len(ports) == 0 {
		return ""
	}
	
	var result strings.Builder
	
	// Find max lengths for proper alignment
	maxNameLen := 0
	maxOrigLen := 0
	maxMappedLen := 0
	
	for _, pm := range ports {
		if len(pm.Name) > maxNameLen {
			maxNameLen = len(pm.Name)
		}
		origStr := fmt.Sprintf("%d", pm.Original)
		if len(origStr) > maxOrigLen {
			maxOrigLen = len(origStr)
		}
		mappedStr := fmt.Sprintf("%d", pm.Mapped)
		if len(mappedStr) > maxMappedLen {
			maxMappedLen = len(mappedStr)
		}
	}
	
	// Calculate total width: " NAME: ORIGPORT → MAPPEDPORT " + padding
	// Add space for padding: 1 (start) + maxNameLen + 2 (: ) + maxOrigLen + 3 ( → ) + maxMappedLen + 1 (end)
	contentWidth := 1 + maxNameLen + 2 + maxOrigLen + 3 + maxMappedLen + 1
	boxWidth := contentWidth
	if boxWidth < 40 {
		boxWidth = 40 // Minimum width
	}
	
	// Top border
	result.WriteString("   ┌")
	for i := 0; i < boxWidth; i++ {
		result.WriteString("─")
	}
	result.WriteString("┐\n")
	
	// Title
	title := "🔌 Port Mappings"
	titleLen := 16 // Actual display length without emoji width issues
	titlePadding := (boxWidth - titleLen) / 2
	result.WriteString("   │")
	for i := 0; i < titlePadding; i++ {
		result.WriteString(" ")
	}
	result.WriteString(Bold + Cyan + title + Reset)
	for i := 0; i < boxWidth - titlePadding - titleLen; i++ {
		result.WriteString(" ")
	}
	result.WriteString("│\n")
	
	// Separator
	result.WriteString("   ├")
	for i := 0; i < boxWidth; i++ {
		result.WriteString("─")
	}
	result.WriteString("┤\n")
	
	// Port mappings with right-aligned ports and vertically aligned arrows
	for _, pm := range ports {
		origColor := PortColor(pm.Original)
		mappedColor := PortColor(pm.Mapped)
		
		// Build the line with proper alignment
		result.WriteString("   │ ")
		
		// Variable name, left-aligned
		result.WriteString(pm.Name)
		result.WriteString(":")
		
		// Padding after name to align the original port
		nameSpace := maxNameLen - len(pm.Name) + 1
		for i := 0; i < nameSpace; i++ {
			result.WriteString(" ")
		}
		
		// Original port, right-aligned
		origStr := fmt.Sprintf("%d", pm.Original)
		origPadding := maxOrigLen - len(origStr)
		for i := 0; i < origPadding; i++ {
			result.WriteString(" ")
		}
		result.WriteString(origColor)
		result.WriteString(origStr)
		result.WriteString(Reset)
		
		// Arrow
		result.WriteString(" → ")
		
		// Mapped port, right-aligned
		mappedStr := fmt.Sprintf("%d", pm.Mapped)
		mappedPadding := maxMappedLen - len(mappedStr)
		for i := 0; i < mappedPadding; i++ {
			result.WriteString(" ")
		}
		result.WriteString(mappedColor)
		result.WriteString(mappedStr)
		result.WriteString(Reset)
		
		// Padding to box edge
		currentLen := 1 + len(pm.Name) + 1 + nameSpace + origPadding + len(origStr) + 3 + mappedPadding + len(mappedStr) + 1
		endPadding := boxWidth - currentLen
		for i := 0; i < endPadding; i++ {
			result.WriteString(" ")
		}
		
		result.WriteString(" │\n")
	}
	
	// Bottom border
	result.WriteString("   └")
	for i := 0; i < boxWidth; i++ {
		result.WriteString("─")
	}
	result.WriteString("┘\n")
	
	return result.String()
}

// ColorizePortBadge creates a badge-style display for a port mapping
func ColorizePortBadge(original, mapped int) string {
	origColor := PortColor(original)
	mappedColor := PortColor(mapped)
	
	return fmt.Sprintf("%s%s %d %s → %s%s %d %s", 
		Bold, origColor, original, Reset,
		Bold, mappedColor, mapped, Reset)
}

// URLRewrite represents a URL rewrite mapping
type URLRewrite struct {
	Name     string
	Original string
	Current  string
}

// FormatURLCard creates a visually appealing card display for URL rewrites
func FormatURLCard(rewrites []URLRewrite) string {
	if len(rewrites) == 0 {
		return ""
	}
	
	var result strings.Builder
	
	// Find max lengths for alignment
	maxNameLen := 0
	for _, ur := range rewrites {
		if len(ur.Name) > maxNameLen {
			maxNameLen = len(ur.Name)
		}
	}
	
	// Calculate box width
	boxWidth := 60 // Default width for URLs
	
	// Top border
	result.WriteString("   ┌")
	for i := 0; i < boxWidth; i++ {
		result.WriteString("─")
	}
	result.WriteString("┐\n")
	
	// Title
	title := "🔗 URL/Connection String Rewrites"
	titleLen := 33 // Actual display length
	titlePadding := (boxWidth - titleLen) / 2
	result.WriteString("   │")
	for i := 0; i < titlePadding; i++ {
		result.WriteString(" ")
	}
	result.WriteString(Bold + Yellow + title + Reset)
	for i := 0; i < boxWidth - titlePadding - titleLen; i++ {
		result.WriteString(" ")
	}
	result.WriteString("│\n")
	
	// Separator
	result.WriteString("   ├")
	for i := 0; i < boxWidth; i++ {
		result.WriteString("─")
	}
	result.WriteString("┤\n")
	
	// URL rewrites
	for _, ur := range rewrites {
		// Variable name line
		result.WriteString("   │ ")
		result.WriteString(Bold + ur.Name + Reset + ":")
		namePadding := boxWidth - len(ur.Name) - 2
		for i := 0; i < namePadding; i++ {
			result.WriteString(" ")
		}
		result.WriteString("│\n")
		
		// Original URL (truncated if needed)
		origTrunc := truncateValue(ur.Original, boxWidth - 6)
		result.WriteString("   │   ")
		result.WriteString(origTrunc)
		origPadding := boxWidth - len(origTrunc) - 3
		for i := 0; i < origPadding; i++ {
			result.WriteString(" ")
		}
		result.WriteString("│\n")
		
		// Arrow
		result.WriteString("   │   → ")
		currTrunc := truncateValue(ur.Current, boxWidth - 8)
		result.WriteString(Green + currTrunc + Reset)
		currPadding := boxWidth - len(currTrunc) - 5
		for i := 0; i < currPadding; i++ {
			result.WriteString(" ")
		}
		result.WriteString("│\n")
	}
	
	// Bottom border
	result.WriteString("   └")
	for i := 0; i < boxWidth; i++ {
		result.WriteString("─")
	}
	result.WriteString("┘\n")
	
	return result.String()
}

// IsolatedPath represents an isolated path mapping
type IsolatedPath struct {
	Name     string
	Original string
	Current  string
}

// FormatIsolatedPathCard creates a visually appealing card display for isolated paths
func FormatIsolatedPathCard(paths []IsolatedPath) string {
	if len(paths) == 0 {
		return ""
	}
	
	var result strings.Builder
	
	// Calculate box width
	boxWidth := 60 // Default width for paths
	
	// Top border
	result.WriteString("   ┌")
	for i := 0; i < boxWidth; i++ {
		result.WriteString("─")
	}
	result.WriteString("┐\n")
	
	// Title
	title := "📁 Isolated Paths"
	titleLen := 17 // Actual display length
	titlePadding := (boxWidth - titleLen) / 2
	result.WriteString("   │")
	for i := 0; i < titlePadding; i++ {
		result.WriteString(" ")
	}
	result.WriteString(Bold + Blue + title + Reset)
	for i := 0; i < boxWidth - titlePadding - titleLen; i++ {
		result.WriteString(" ")
	}
	result.WriteString("│\n")
	
	// Separator
	result.WriteString("   ├")
	for i := 0; i < boxWidth; i++ {
		result.WriteString("─")
	}
	result.WriteString("┤\n")
	
	// Isolated paths
	for _, ip := range paths {
		// Variable name line
		result.WriteString("   │ ")
		result.WriteString(Bold + ip.Name + Reset + ":")
		namePadding := boxWidth - len(ip.Name) - 2
		for i := 0; i < namePadding; i++ {
			result.WriteString(" ")
		}
		result.WriteString("│\n")
		
		// Original path
		result.WriteString("   │   ")
		result.WriteString(ip.Original)
		origPadding := boxWidth - len(ip.Original) - 3
		for i := 0; i < origPadding; i++ {
			result.WriteString(" ")
		}
		result.WriteString("│\n")
		
		// Arrow and new path
		result.WriteString("   │   → ")
		result.WriteString(Cyan + ip.Current + Reset)
		currPadding := boxWidth - len(ip.Current) - 5
		for i := 0; i < currPadding; i++ {
			result.WriteString(" ")
		}
		result.WriteString("│\n")
	}
	
	// Bottom border
	result.WriteString("   └")
	for i := 0; i < boxWidth; i++ {
		result.WriteString("─")
	}
	result.WriteString("┘\n")
	
	return result.String()
}

// truncateValue truncates a string to fit within a certain width
func truncateValue(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}