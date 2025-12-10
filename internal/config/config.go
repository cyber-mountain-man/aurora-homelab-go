package config

import "github.com/cyber-mountain-man/aurora-homelab-go/internal/models"

// Config is the top-level configuration structure.
type Config struct {
	Services []models.Service `yaml:"services"`
}

// Load will read config.yaml in a future milestone.
// For now it just returns an empty config so the app compiles.
func Load(path string) (*Config, error) {
	return &Config{}, nil
}
