package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const manifestFileName = "pa.yaml"

// Config holds the resolved runtime configuration.
type Config struct {
	PAHome   string    // Resolved absolute path to the ProjectAccelerator root.
	Manifest *Manifest // Parsed pa.yaml.
}

// Load resolves PA_HOME and loads the manifest.
//
// Resolution order:
//  1. PA_HOME environment variable
//  2. Walk up from the current working directory looking for pa.yaml
//  3. Walk up from fallbackDir looking for pa.yaml
//
// Returns an error if pa.yaml cannot be found or parsed.
func Load(fallbackDir string) (*Config, error) {
	home, err := resolveHome(fallbackDir)
	if err != nil {
		return nil, fmt.Errorf("locating pa.yaml: %w", err)
	}

	manifestPath := filepath.Join(home, manifestFileName)
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	return &Config{
		PAHome:   home,
		Manifest: manifest,
	}, nil
}

// resolveHome determines the ProjectAccelerator root directory.
func resolveHome(fallbackDir string) (string, error) {
	// 1. Explicit env override.
	if envHome := os.Getenv("PA_HOME"); envHome != "" {
		abs, err := filepath.Abs(envHome)
		if err != nil {
			return "", fmt.Errorf("resolving PA_HOME %q: %w", envHome, err)
		}
		return abs, nil
	}

	// 2. Walk up from CWD.
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting working directory: %w", err)
	}

	if home, err := findPAHome(cwd); err == nil {
		return home, nil
	}

	// 3. Walk up from fallbackDir (e.g. the binary's location).
	if fallbackDir != "" {
		abs, err := filepath.Abs(fallbackDir)
		if err != nil {
			return "", fmt.Errorf("resolving fallback directory %q: %w", fallbackDir, err)
		}
		if home, err := findPAHome(abs); err == nil {
			return home, nil
		}
	}

	return "", errors.New("pa.yaml not found; set PA_HOME or run from within the ProjectAccelerator tree")
}

// findPAHome walks up from startDir looking for pa.yaml.
// Returns the directory containing pa.yaml, or an error if not found.
func findPAHome(startDir string) (string, error) {
	dir := startDir
	for {
		candidate := filepath.Join(dir, manifestFileName)
		if _, err := os.Stat(candidate); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding pa.yaml.
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("pa.yaml not found starting from %q", startDir)
}
