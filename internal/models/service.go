package models

// Service represents an application or endpoint in your homelab.
// It now supports different health check types.
//
// type:
//   - "http" (default): uses URL
//   - "tcp": uses Host + Port
//
// For HTTP services:
//   - set Type: "http" (or leave empty to default to http)
//   - set URL
//
// For TCP services:
//   - set Type: "tcp"
//   - set Host and Port
type Service struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type,omitempty"` // "http" (default) or "tcp"
	URL         string `yaml:"url,omitempty"`  // used for HTTP
	Host        string `yaml:"host,omitempty"` // used for TCP
	Port        int    `yaml:"port,omitempty"` // used for TCP
	Icon        string `yaml:"icon,omitempty"`
	Category    string `yaml:"category,omitempty"`
	Description string `yaml:"description,omitempty"`
}
