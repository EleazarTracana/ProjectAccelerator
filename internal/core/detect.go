package core

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gentleman-programming/project-accelerator/internal/config"
)

// DetectResult holds the outcome of matching a single stack against a directory.
type DetectResult struct {
	Stack   string             // Stack key from the manifest (e.g. "python").
	Config  config.StackConfig // The matched stack configuration.
	Matches []string           // Detect patterns that produced a match.
}

// DetectStack inspects targetDir against every stack defined in manifest.
// A stack matches when at least one of its detect patterns resolves to an
// existing file or when a glob pattern matches at least one file in targetDir.
//
// Supported pattern forms:
//   - Exact filename  "pyproject.toml" → targetDir/pyproject.toml must exist.
//   - Glob            "*.py"           → filepath.Glob(targetDir/*.py) must be non-empty.
//
// Returns all matching stacks; the slice is empty when nothing is detected.
func DetectStack(targetDir string, manifest *config.Manifest) ([]DetectResult, error) {
	if _, err := os.Stat(targetDir); err != nil {
		return nil, fmt.Errorf("accessing target directory %q: %w", targetDir, err)
	}

	// Build a flat list of the files that sit directly inside targetDir so we
	// can check glob patterns without a full recursive walk for each stack.
	rootEntries, err := rootFileNames(targetDir)
	if err != nil {
		return nil, fmt.Errorf("listing %q: %w", targetDir, err)
	}

	var results []DetectResult

	for key, stack := range manifest.Stacks {
		var matches []string

		for _, pattern := range stack.Detect {
			if matched, err := patternMatches(targetDir, pattern, rootEntries); err != nil {
				return nil, fmt.Errorf("evaluating detect pattern %q for stack %q: %w", pattern, key, err)
			} else if matched {
				matches = append(matches, pattern)
			}
		}

		if len(matches) > 0 {
			results = append(results, DetectResult{
				Stack:   key,
				Config:  stack,
				Matches: matches,
			})
		}
	}

	return results, nil
}

// rootFileNames returns the names (not full paths) of all non-directory entries
// directly inside dir. Hidden files are included; subdirectories are excluded.
func rootFileNames(dir string) (map[string]struct{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	names := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			names[e.Name()] = struct{}{}
		}
	}
	return names, nil
}

// patternMatches returns true when pattern matches something in targetDir.
// It distinguishes between plain filenames and glob expressions.
func patternMatches(targetDir, pattern string, rootEntries map[string]struct{}) (bool, error) {
	// If the pattern contains a glob metacharacter, use filepath.Glob.
	if containsGlobMeta(pattern) {
		full := filepath.Join(targetDir, pattern)
		matched, err := filepath.Glob(full)
		if err != nil {
			return false, fmt.Errorf("glob %q: %w", full, err)
		}
		return len(matched) > 0, nil
	}

	// Plain filename — check the precomputed root entries map first (O(1)),
	// then fall back to os.Stat for patterns that include a sub-path.
	if _, ok := rootEntries[pattern]; ok {
		return true, nil
	}

	// Allow patterns that specify a relative sub-path (e.g. "src/main.py").
	full := filepath.Join(targetDir, pattern)
	if _, err := os.Stat(full); err == nil {
		return true, nil
	}

	return false, nil
}

// containsGlobMeta reports whether s contains any filepath.Match metacharacter.
func containsGlobMeta(s string) bool {
	for _, r := range s {
		switch r {
		case '*', '?', '[', ']':
			return true
		}
	}
	return false
}
