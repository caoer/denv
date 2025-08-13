package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Rule struct {
	Action string `yaml:"action"`
	Range  []int  `yaml:"range,omitempty"`
	Base   string `yaml:"base,omitempty"`
	OnlyIf []string `yaml:"only_if,omitempty"`
}

type PatternRule struct {
	Pattern string `yaml:"pattern"`
	Rule    Rule   `yaml:"rule"`
}

type Config struct {
	Projects map[string]string `yaml:"projects"`
	Patterns []PatternRule     `yaml:"patterns"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default config file if it doesn't exist
			cfg := defaultConfig()
			if saveErr := SaveConfig(path, cfg); saveErr != nil {
				// Return default config even if save fails
				return cfg, nil
			}
			return cfg, nil
		}
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Merge with defaults if patterns is empty
	if len(cfg.Patterns) == 0 {
		cfg.Patterns = defaultPatterns()
	}
	
	// Initialize projects map if nil
	if cfg.Projects == nil {
		cfg.Projects = make(map[string]string)
	}

	return &cfg, nil
}

func SaveConfig(path string, cfg *Config) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func defaultConfig() *Config {
	return &Config{
		Projects: make(map[string]string),
		Patterns: defaultPatterns(),
	}
}

func defaultPatterns() []PatternRule {
	// Pattern order matters - first match wins in ApplyRules
	// System patterns must be listed before generic patterns
	var patterns []PatternRule
	
	// System PATH variables that should not be modified
	systemPaths := []string{
		"CARGO_HOME",
		"RUSTUP_HOME",
		"PNPM_HOME",
		"NIX_PATH",
		"NIX_USER_PROFILE_DIR",
		"BROWSERS_PROFILE_PATH",
		"SOLANA_HOME",
		"KITTY_INSTALLATION_DIR",
		"ZSH_CACHE_DIR",
		"MINIO_HOME",
		"DOT_PATH",
		"FORGIT_INSTALL_DIR",
		"__MISE_ORIG_PATH",
		"TMUX_PLUGIN_MANAGER_PATH",
		"GOPATH",
		"GOROOT",
		"NVM_DIR",
		"RBENV_ROOT",
		"PYENV_ROOT",
		"SDKMAN_DIR",
		"HOMEBREW_PREFIX",
		"HOMEBREW_CELLAR",
		"HOMEBREW_REPOSITORY",
	}
	
	// Add keep rules for system paths (these come first)
	for _, path := range systemPaths {
		patterns = append(patterns, PatternRule{
			Pattern: path,
			Rule: Rule{
				Action: "keep",
			},
		})
	}
	
	// Add generic patterns after system-specific ones
	patterns = append(patterns, PatternRule{
		Pattern: "*_PORT|PORT",
		Rule: Rule{
			Action: "random_port",
			Range:  []int{30000, 39999},
		},
	})
	patterns = append(patterns, PatternRule{
		Pattern: "*_ROOT|*_DIR|*_PATH|*_HOME",
		Rule: Rule{
			Action: "isolate",
			Base:   "${DENV_ENV}",
		},
	})
	patterns = append(patterns, PatternRule{
		Pattern: "*_URL|*_URI|*_ENDPOINT|DATABASE_URL|REDIS_URL",
		Rule: Rule{
			Action: "rewrite_ports",
		},
	})
	patterns = append(patterns, PatternRule{
		Pattern: "*_KEY|*_TOKEN|*_SECRET|*_PASSWORD|*_CREDENTIAL",
		Rule: Rule{
			Action: "keep",
		},
	})
	patterns = append(patterns, PatternRule{
		Pattern: "*_HOST|*_HOSTNAME",
		Rule: Rule{
			Action: "keep",
			OnlyIf: []string{"localhost", "127.0.0.1", "0.0.0.0"},
		},
	})
	
	return patterns
}