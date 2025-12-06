package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

const ConfigFile = "beans.toml"

// Config holds the beans configuration.
type Config struct {
	Statuses StatusConfig `toml:"statuses"`
}

// StatusConfig defines available statuses and the default.
type StatusConfig struct {
	Available []string `toml:"available"`
	Default   string   `toml:"default"`
}

// Default returns a Config with default values.
func Default() *Config {
	return &Config{
		Statuses: StatusConfig{
			Available: []string{"open", "in-progress", "done"},
			Default:   "open",
		},
	}
}

// Load reads configuration from the given .beans directory.
// Returns default config if the file doesn't exist.
func Load(root string) (*Config, error) {
	path := filepath.Join(root, ConfigFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Default(), nil
		}
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Save writes the configuration to the given .beans directory.
func (c *Config) Save(root string) error {
	path := filepath.Join(root, ConfigFile)

	data, err := toml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// IsValidStatus returns true if the status is in the available list.
func (c *Config) IsValidStatus(status string) bool {
	for _, s := range c.Statuses.Available {
		if s == status {
			return true
		}
	}
	return false
}

// StatusList returns a comma-separated list of available statuses.
func (c *Config) StatusList() string {
	if len(c.Statuses.Available) == 0 {
		return ""
	}

	result := c.Statuses.Available[0]
	for i := 1; i < len(c.Statuses.Available); i++ {
		result += ", " + c.Statuses.Available[i]
	}
	return result
}
