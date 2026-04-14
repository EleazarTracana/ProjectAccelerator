package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// StackConfig represents a stack entry in the manifest.
type StackConfig struct {
	Name           string   `yaml:"name"`
	Detect         []string `yaml:"detect"`
	RulesDir       string   `yaml:"rules_dir"`
	SkeletonDir    string   `yaml:"skeleton_dir"`
	ClaudeTemplate string   `yaml:"claude_template"`
}

// Manifest represents the pa.yaml file.
type Manifest struct {
	Version           int                    `yaml:"version"`
	Stacks            map[string]StackConfig `yaml:"stacks"`
	UniversalRulesDir string                 `yaml:"universal_rules_dir"`
}

// LoadManifest reads and parses pa.yaml from the given path.
// Returns an error if the file cannot be read, the YAML is invalid,
// the version is not 1, or the stacks map is empty.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading manifest %q: %w", path, err)
	}

	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing manifest %q: %w", path, err)
	}

	if m.Version != 1 {
		return nil, fmt.Errorf("unsupported manifest version %d (expected 1)", m.Version)
	}

	if len(m.Stacks) == 0 {
		return nil, fmt.Errorf("manifest %q defines no stacks", path)
	}

	return &m, nil
}
