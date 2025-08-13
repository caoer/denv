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
	h.Write([]byte(fmt.Sprintf("%d", port)))
	hash := h.Sum32()
	
	// Map hash to color index
	colorIndex := int(hash % uint32(len(portColors)))
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