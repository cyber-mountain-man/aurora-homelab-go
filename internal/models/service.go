package models

// Service represents an application or endpoint in your homelab.
// Soon this will be populated from config.yaml.
type Service struct {
	Name        string `yaml:"name"`
	URL         string `yaml:"url"`
	Icon        string `yaml:"icon,omitempty"`
	Category    string `yaml:"category,omitempty"`
	Description string `yaml:"description,omitempty"`
}
