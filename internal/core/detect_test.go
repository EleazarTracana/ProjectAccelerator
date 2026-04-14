package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gentleman-programming/project-accelerator/internal/config"
)

// buildManifest is a helper that creates a Manifest with the given stacks.
func buildManifest(stacks map[string]config.StackConfig) *config.Manifest {
	return &config.Manifest{
		Version: 1,
		Stacks:  stacks,
	}
}

func TestDetectStack(t *testing.T) {
	t.Run("python project detects python stack", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.py"), []byte("print('hi')"), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := buildManifest(map[string]config.StackConfig{
			"python": {
				Name:   "Python",
				Detect: []string{"*.py"},
			},
		})

		results, err := DetectStack(dir, manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("len(results) = %d, want 1", len(results))
		}
		if results[0].Stack != "python" {
			t.Errorf("Stack = %q, want %q", results[0].Stack, "python")
		}
		if len(results[0].Matches) == 0 {
			t.Error("Matches should not be empty")
		}
	})

	t.Run("csharp project detects csharp stack", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "MyApp.csproj"), []byte("<Project/>"), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := buildManifest(map[string]config.StackConfig{
			"csharp": {
				Name:   "C#",
				Detect: []string{"*.csproj"},
			},
		})

		results, err := DetectStack(dir, manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("len(results) = %d, want 1", len(results))
		}
		if results[0].Stack != "csharp" {
			t.Errorf("Stack = %q, want %q", results[0].Stack, "csharp")
		}
	})

	t.Run("frontend project detects frontend stack via exact filename", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := buildManifest(map[string]config.StackConfig{
			"frontend": {
				Name:   "Frontend",
				Detect: []string{"package.json"},
			},
		})

		results, err := DetectStack(dir, manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("len(results) = %d, want 1", len(results))
		}
		if results[0].Stack != "frontend" {
			t.Errorf("Stack = %q, want %q", results[0].Stack, "frontend")
		}
	})

	t.Run("empty directory detects nothing", func(t *testing.T) {
		dir := t.TempDir()

		manifest := buildManifest(map[string]config.StackConfig{
			"python": {
				Name:   "Python",
				Detect: []string{"*.py", "pyproject.toml"},
			},
			"csharp": {
				Name:   "C#",
				Detect: []string{"*.csproj"},
			},
		})

		results, err := DetectStack(dir, manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}
		if len(results) != 0 {
			t.Errorf("len(results) = %d, want 0 for empty directory", len(results))
		}
	})

	t.Run("multiple stacks can match simultaneously", func(t *testing.T) {
		dir := t.TempDir()
		// Create files that match both Python and Frontend stacks.
		if err := os.WriteFile(filepath.Join(dir, "script.py"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := buildManifest(map[string]config.StackConfig{
			"python": {
				Name:   "Python",
				Detect: []string{"*.py"},
			},
			"frontend": {
				Name:   "Frontend",
				Detect: []string{"package.json"},
			},
		})

		results, err := DetectStack(dir, manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}
		if len(results) != 2 {
			t.Errorf("len(results) = %d, want 2 (both stacks should match)", len(results))
		}

		found := map[string]bool{}
		for _, r := range results {
			found[r.Stack] = true
		}
		if !found["python"] {
			t.Error("python stack should have been detected")
		}
		if !found["frontend"] {
			t.Error("frontend stack should have been detected")
		}
	})

	t.Run("no match when detect patterns do not match", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := buildManifest(map[string]config.StackConfig{
			"python": {
				Name:   "Python",
				Detect: []string{"*.py"},
			},
		})

		results, err := DetectStack(dir, manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}
		if len(results) != 0 {
			t.Errorf("len(results) = %d, want 0", len(results))
		}
	})

	t.Run("non-existent directory returns error", func(t *testing.T) {
		manifest := buildManifest(map[string]config.StackConfig{
			"python": {Name: "Python", Detect: []string{"*.py"}},
		})

		_, err := DetectStack(filepath.Join(t.TempDir(), "does_not_exist"), manifest)
		if err == nil {
			t.Error("DetectStack() error = nil, want error for non-existent directory")
		}
	})

	t.Run("python stack matches via exact filename pyproject.toml", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[tool.poetry]"), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := buildManifest(map[string]config.StackConfig{
			"python": {
				Name:   "Python",
				Detect: []string{"pyproject.toml"},
			},
		})

		results, err := DetectStack(dir, manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}
		if len(results) != 1 || results[0].Stack != "python" {
			t.Errorf("expected python detection, got %v", results)
		}
	})

	t.Run("glob pattern in subdirectory is not matched by root-only detection", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "src")
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatal(err)
		}
		// File is in a subdirectory — glob *.py only looks at root.
		if err := os.WriteFile(filepath.Join(subDir, "main.py"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		manifest := buildManifest(map[string]config.StackConfig{
			"python": {
				Name:   "Python",
				Detect: []string{"*.py"},
			},
		})

		results, err := DetectStack(dir, manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}
		// *.py glob uses filepath.Glob(dir/*.py) — subdirectory files are NOT matched.
		if len(results) != 0 {
			t.Errorf("len(results) = %d, want 0 (*.py should not match files in subdirectories)", len(results))
		}
	})
}
