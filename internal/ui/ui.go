package ui

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Data structures matching the test expectations
type PortMapping struct {
	Name     string
	Original int
	Mapped   int
}

type URLRewrite struct {
	Name     string
	Original string
	Current  string
}

type IsolatedPath struct {
	Name     string
	Original string
	Current  string
}

// Define consistent styles using lipgloss
var (
	// Color palette for ports - using lipgloss color system
	portColors = []lipgloss.Color{
		lipgloss.Color("14"),  // Cyan
		lipgloss.Color("10"),  // Bright Green
		lipgloss.Color("11"),  // Bright Yellow
		lipgloss.Color("12"),  // Bright Blue
		lipgloss.Color("13"),  // Bright Magenta
		lipgloss.Color("6"),   // Cyan
		lipgloss.Color("2"),   // Green
		lipgloss.Color("3"),   // Yellow
		lipgloss.Color("4"),   // Blue
		lipgloss.Color("5"),   // Magenta
	}

	// Card container style - left aligned content with minimal padding
	cardStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(0, 2).
		Align(lipgloss.Left)

	// Title styles
	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center)

	portTitleStyle = titleStyle.Copy().
		Foreground(lipgloss.Color("14"))

	urlTitleStyle = titleStyle.Copy().
		Foreground(lipgloss.Color("11"))

	pathTitleStyle = titleStyle.Copy().
		Foreground(lipgloss.Color("12"))

	// Content styles
	labelStyle = lipgloss.NewStyle().
		Bold(true)

	arrowStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))
)

// getPortColor returns a consistent color for a given port number
func getPortColor(port int) lipgloss.Color {
	h := fnv.New32a()
	h.Write([]byte(fmt.Sprintf("%d", port)))
	hash := h.Sum32()
	colorIndex := hash % uint32(len(portColors))
	return portColors[colorIndex]
}

// StylePort returns a styled string for a port number
func StylePort(port int) string {
	style := lipgloss.NewStyle().Foreground(getPortColor(port))
	return style.Render(fmt.Sprintf("%d", port))
}

// RenderPortCard creates a card display for port mappings
func RenderPortCard(ports []PortMapping) string {
	if len(ports) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, portTitleStyle.Render("🔌 Port Mappings"))

	// Find max lengths for alignment
	maxNameLen := 0
	for _, p := range ports {
		if len(p.Name) > maxNameLen {
			maxNameLen = len(p.Name)
		}
	}

	// Render each port mapping with left alignment
	for _, p := range ports {
		origStyle := lipgloss.NewStyle().Foreground(getPortColor(p.Original))
		mappedStyle := lipgloss.NewStyle().Foreground(getPortColor(p.Mapped))
		
		line := fmt.Sprintf("%-*s : %s → %s",
			maxNameLen,
			p.Name,
			origStyle.Render(fmt.Sprintf("%d", p.Original)),
			mappedStyle.Render(fmt.Sprintf("%d", p.Mapped)))
		
		lines = append(lines, line)
	}

	content := strings.Join(lines, "\n")
	return cardStyle.Render(content)
}

// RenderURLCard creates a card display for URL rewrites
func RenderURLCard(urls []URLRewrite) string {
	if len(urls) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, urlTitleStyle.Render("🔗 URL/Connection String Rewrites"))

	for _, u := range urls {
		lines = append(lines, labelStyle.Render(u.Name+":"))
		lines = append(lines, fmt.Sprintf("  %s", u.Original))
		lines = append(lines, fmt.Sprintf("  %s %s", 
			arrowStyle.Render("→"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(u.Current)))
		if len(urls) > 1 {
			lines = append(lines, "") // Only add spacing between multiple items
		}
	}

	// Remove last empty line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	content := strings.Join(lines, "\n")
	return cardStyle.Render(content)
}

// RenderIsolatedPathCard creates a card display for isolated paths
func RenderIsolatedPathCard(paths []IsolatedPath) string {
	if len(paths) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, pathTitleStyle.Render("📁 Isolated Paths"))

	for _, p := range paths {
		lines = append(lines, labelStyle.Render(p.Name+":"))
		lines = append(lines, fmt.Sprintf("  %s", p.Original))
		lines = append(lines, fmt.Sprintf("  %s %s",
			arrowStyle.Render("→"),
			lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Render(p.Current)))
		if len(paths) > 1 {
			lines = append(lines, "") // Only add spacing between multiple items
		}
	}

	// Remove last empty line if present
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	content := strings.Join(lines, "\n")
	return cardStyle.Render(content)
}

// ColorizePortInURL colorizes port numbers within a URL string
func ColorizePortInURL(url string, port int) string {
	portStr := fmt.Sprintf(":%d", port)
	style := lipgloss.NewStyle().Foreground(getPortColor(port))
	coloredPort := fmt.Sprintf(":%s", style.Render(fmt.Sprintf("%d", port)))
	return strings.ReplaceAll(url, portStr, coloredPort)
}

// RenderEnvironmentList creates a formatted list of environments
func RenderEnvironmentList(title string, environments map[string][]EnvInfo) string {
	if len(environments) == 0 {
		return "No denv environments found"
	}

	var output strings.Builder
	
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("14"))
	
	output.WriteString("\n")
	output.WriteString(headerStyle.Render("🌍 " + title))
	output.WriteString("\n")
	output.WriteString(strings.Repeat("─", 50))
	output.WriteString("\n")

	// Sort projects for consistent display
	var projects []string
	for project := range environments {
		projects = append(projects, project)
	}
	// Simple sort
	for i := 0; i < len(projects); i++ {
		for j := i + 1; j < len(projects); j++ {
			if projects[i] > projects[j] {
				projects[i], projects[j] = projects[j], projects[i]
			}
		}
	}

	projectStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("12"))
	
	envStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10"))
	
	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241"))

	for _, project := range projects {
		envs := environments[project]
		output.WriteString("\n")
		output.WriteString(projectStyle.Render("📦 " + project))
		output.WriteString("\n")
		
		if len(envs) == 0 {
			output.WriteString("   ")
			output.WriteString(inactiveStyle.Render("(shared project directory only)"))
			output.WriteString("\n")
			continue
		}
		
		for _, env := range envs {
			output.WriteString("   • ")
			if env.Active {
				output.WriteString(envStyle.Render(env.Name))
				output.WriteString(": ")
				output.WriteString(fmt.Sprintf("%d active session(s)", env.Sessions))
			} else {
				output.WriteString(env.Name)
				output.WriteString(": ")
				output.WriteString(inactiveStyle.Render("inactive"))
			}
			
			if env.Ports > 0 {
				portsInfo := fmt.Sprintf(" [%d ports mapped]", env.Ports)
				output.WriteString(lipgloss.NewStyle().
					Foreground(lipgloss.Color("14")).
					Render(portsInfo))
			}
			output.WriteString("\n")
		}
	}
	
	output.WriteString("\n")
	output.WriteString(strings.Repeat("─", 50))
	return output.String()
}

// EnvInfo represents environment information for list display
type EnvInfo struct {
	Name     string
	Active   bool
	Sessions int
	Ports    int
}