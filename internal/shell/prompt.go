package shell

import (
	crypto_rand "crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math/big"
)

// Dark mode friendly colors (256-color ANSI codes)
var darkModeColors = []string{
	"\033[38;5;39m",  // Bright blue
	"\033[38;5;46m",  // Bright green
	"\033[38;5;208m", // Orange
	"\033[38;5;99m",  // Purple
	"\033[38;5;87m",  // Cyan
	"\033[38;5;226m", // Yellow
	"\033[38;5;201m", // Magenta
	"\033[38;5;51m",  // Light blue
	"\033[38;5;118m", // Light green
	"\033[38;5;214m", // Light orange
}

const colorReset = "\033[0m"

// GenerateColoredPrompt generates a shell-specific prompt with colored environment name
func GenerateColoredPrompt(envName string, shellType ShellType) string {
	color := GetColorForEnvironment(envName)

	switch shellType {
	case Fish:
		return generateFishColoredPrompt(envName, color)
	case Zsh:
		return generateZshColoredPrompt(envName, color)
	default:
		// Bash and Sh use ANSI color syntax with PS1
		return generateBashStylePrompt(envName, color)
	}
}

// GetDarkModeColor returns a random color suitable for dark terminals
func GetDarkModeColor() string {
	n, err := crypto_rand.Int(crypto_rand.Reader, big.NewInt(int64(len(darkModeColors))))
	if err != nil {
		// Fallback to first color on error
		return darkModeColors[0]
	}
	return darkModeColors[n.Int64()]
}

// GetColorForEnvironment returns a consistent color for a given environment name
func GetColorForEnvironment(envName string) string {
	// Use SHA256 hash to get a deterministic but pseudo-random index
	hash := sha256.Sum256([]byte(envName))
	// Use first 4 bytes of hash as seed
	seed := binary.BigEndian.Uint32(hash[:4])
	// Safe modulo operation to prevent overflow
	index := seed % uint32(len(darkModeColors))
	return darkModeColors[index]
}

func generateBashStylePrompt(envName, color string) string {
	// Format: colored (envName) followed by original PS1
	return fmt.Sprintf(`PS1="%s(%s)%s $PS1"`, color, envName, colorReset)
}

func generateZshColoredPrompt(envName, color string) string {
	// Zsh uses PROMPT variable and has its own color syntax
	// Convert ANSI color to zsh %F{n} syntax
	zshColor := mapAnsiToZshColor(color)
	return fmt.Sprintf(`PROMPT="%%F{%s}(%s)%%f $PROMPT"`, zshColor, envName)
}

func generateFishColoredPrompt(envName, color string) string {
	// Fish uses set_color command instead of ANSI codes directly
	// Map ANSI codes to fish color names
	fishColor := mapAnsiToFishColor(color)

	return fmt.Sprintf(`function fish_prompt
    set_color %s
    echo -n "(%s) "
    set_color normal
    __fish_default_prompt
end`, fishColor, envName)
}

// mapAnsiToZshColor maps ANSI color codes to zsh color numbers
func mapAnsiToZshColor(ansiColor string) string {
	// Extract the color number from ANSI code like "\033[38;5;39m"
	// and return just the number for zsh's %F{n} syntax
	colorMap := map[string]string{
		"\033[38;5;39m":  "39",   // Bright blue
		"\033[38;5;46m":  "46",   // Bright green
		"\033[38;5;208m": "208",  // Orange
		"\033[38;5;99m":  "99",   // Purple
		"\033[38;5;87m":  "87",   // Cyan
		"\033[38;5;226m": "226",  // Yellow
		"\033[38;5;201m": "201",  // Magenta
		"\033[38;5;51m":  "51",   // Light blue
		"\033[38;5;118m": "118",  // Light green
		"\033[38;5;214m": "214",  // Light orange
	}
	
	if zshColor, ok := colorMap[ansiColor]; ok {
		return zshColor
	}
	// Default to bright blue if color not found
	return "39"
}

// mapAnsiToFishColor maps ANSI color codes to fish color names
func mapAnsiToFishColor(ansiColor string) string {
	// Map our ANSI 256 colors to fish color names or hex values
	colorMap := map[string]string{
		"\033[38;5;39m":  "blue --bold",
		"\033[38;5;46m":  "green --bold",
		"\033[38;5;208m": "brred", // orange-ish
		"\033[38;5;99m":  "magenta",
		"\033[38;5;87m":  "cyan --bold",
		"\033[38;5;226m": "yellow --bold",
		"\033[38;5;201m": "brmagenta",
		"\033[38;5;51m":  "brcyan",
		"\033[38;5;118m": "brgreen",
		"\033[38;5;214m": "bryellow",
	}

	if fishColor, ok := colorMap[ansiColor]; ok {
		return fishColor
	}
	// Default to bright blue if color not found
	return "blue --bold"
}

