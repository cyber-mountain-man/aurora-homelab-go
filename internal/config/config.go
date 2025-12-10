package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/cyber-mountain-man/aurora-homelab-go/internal/models"
)

// Config is the top-level configuration structure.
type Config struct {
	Services []models.Service `yaml:"services"`
}

// Load reads a YAML config file from the given path and returns a Config.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal yaml: %w", err)
	}

	return &cfg, nil
}
