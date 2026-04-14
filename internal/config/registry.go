package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

const registryFileName = "projects.yaml"

// ProjectEntry represents a registered project in the registry.
type ProjectEntry struct {
	Path         string            `yaml:"path"`
	Stacks       []string          `yaml:"stacks"`
	RegisteredAt time.Time         `yaml:"registered_at"`
	LastSync     time.Time         `yaml:"last_sync"`
	Checksums    map[string]string `yaml:"checksums"` // relative path → "sha256:hex"
}

// Registry holds all registered projects.
type Registry struct {
	Projects []ProjectEntry `yaml:"projects"`
}

// RegistryPath returns the absolute path to ~/.config/pa/projects.yaml.
func RegistryPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving user config directory: %w", err)
	}
	return filepath.Join(configDir, "pa", registryFileName), nil
}

// LoadRegistry loads the registry from disk.
// Returns an empty registry if the file does not exist yet.
func LoadRegistry() (*Registry, error) {
	path, err := RegistryPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &Registry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading registry %q: %w", path, err)
	}

	var r Registry
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parsing registry %q: %w", path, err)
	}

	return &r, nil
}

// Save writes the registry to disk, creating parent directories if needed.
func (r *Registry) Save() error {
	path, err := RegistryPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating registry directory: %w", err)
	}

	data, err := yaml.Marshal(r)
	if err != nil {
		return fmt.Errorf("marshalling registry: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing registry %q: %w", path, err)
	}

	return nil
}

// Add adds a new project entry or replaces an existing one with the same path.
func (r *Registry) Add(entry ProjectEntry) {
	for i, p := range r.Projects {
		if p.Path == entry.Path {
			r.Projects[i] = entry
			return
		}
	}
	r.Projects = append(r.Projects, entry)
}

// Find returns a pointer to the entry with the given path, or nil if not found.
// The returned pointer is valid until the next structural modification of the registry.
func (r *Registry) Find(path string) *ProjectEntry {
	for i := range r.Projects {
		if r.Projects[i].Path == path {
			return &r.Projects[i]
		}
	}
	return nil
}

// Remove removes the entry with the given path.
// Returns true if an entry was removed, false if it was not found.
func (r *Registry) Remove(path string) bool {
	for i, p := range r.Projects {
		if p.Path == path {
			r.Projects = append(r.Projects[:i], r.Projects[i+1:]...)
			return true
		}
	}
	return false
}

// UpdateChecksums updates the checksums and last_sync timestamp for the project
// at the given path. Does nothing if the project is not registered.
func (r *Registry) UpdateChecksums(path string, checksums map[string]string) {
	entry := r.Find(path)
	if entry == nil {
		return
	}
	entry.Checksums = checksums
	entry.LastSync = time.Now().UTC()
}
