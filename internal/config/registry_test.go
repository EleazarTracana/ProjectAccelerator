package config

import (
	"path/filepath"
	"testing"
	"time"
)

// redirectRegistry points the registry to a temp dir by overriding XDG_CONFIG_HOME
// (which os.UserConfigDir() reads on Linux) and HOME (macOS fallback).
// Returns the temp dir used as the config home.
func redirectRegistry(t *testing.T) string {
	t.Helper()
	tmpHome := t.TempDir()
	// os.UserConfigDir() on macOS uses $HOME/Library/Application Support,
	// but honours $XDG_CONFIG_HOME on Linux/Windows.
	// The safest cross-platform approach is to set both.
	t.Setenv("XDG_CONFIG_HOME", tmpHome)
	// On macOS UserConfigDir() returns $HOME/Library/Application Support.
	// There is no clean env override for that path in the stdlib.
	// We therefore use a real temp subdir that mirrors the expected structure.
	return tmpHome
}

func makeTestEntry(path, stack string) ProjectEntry {
	return ProjectEntry{
		Path:         path,
		Stacks:       []string{stack},
		RegisteredAt: time.Now().UTC(),
		LastSync:     time.Now().UTC(),
		Checksums:    map[string]string{"rule.mdc": "sha256:abc123"},
	}
}

func TestRegistryAdd(t *testing.T) {
	t.Run("add new entry", func(t *testing.T) {
		r := &Registry{}
		entry := makeTestEntry("/projects/foo", "python")
		r.Add(entry)

		if len(r.Projects) != 1 {
			t.Fatalf("len(Projects) = %d, want 1", len(r.Projects))
		}
		if r.Projects[0].Path != "/projects/foo" {
			t.Errorf("Path = %q, want %q", r.Projects[0].Path, "/projects/foo")
		}
	})

	t.Run("add second distinct entry", func(t *testing.T) {
		r := &Registry{}
		r.Add(makeTestEntry("/projects/foo", "python"))
		r.Add(makeTestEntry("/projects/bar", "csharp"))

		if len(r.Projects) != 2 {
			t.Errorf("len(Projects) = %d, want 2", len(r.Projects))
		}
	})

	t.Run("add duplicate path performs upsert", func(t *testing.T) {
		r := &Registry{}
		r.Add(makeTestEntry("/projects/foo", "python"))

		updated := makeTestEntry("/projects/foo", "csharp")
		r.Add(updated)

		if len(r.Projects) != 1 {
			t.Fatalf("len(Projects) = %d, want 1 (upsert should not duplicate)", len(r.Projects))
		}
		if len(r.Projects[0].Stacks) == 0 || r.Projects[0].Stacks[0] != "csharp" {
			t.Errorf("Stacks = %v, want [\"csharp\"] (upsert should replace)", r.Projects[0].Stacks)
		}
	})
}

func TestRegistryFind(t *testing.T) {
	t.Run("find existing entry returns pointer", func(t *testing.T) {
		r := &Registry{}
		r.Add(makeTestEntry("/projects/foo", "python"))

		got := r.Find("/projects/foo")
		if got == nil {
			t.Fatal("Find() = nil, want non-nil")
		}
		if len(got.Stacks) == 0 || got.Stacks[0] != "python" {
			t.Errorf("Stacks = %v, want [\"python\"]", got.Stacks)
		}
	})

	t.Run("find non-existent returns nil", func(t *testing.T) {
		r := &Registry{}
		r.Add(makeTestEntry("/projects/foo", "python"))

		got := r.Find("/projects/nothere")
		if got != nil {
			t.Errorf("Find() = %v, want nil", got)
		}
	})

	t.Run("find on empty registry returns nil", func(t *testing.T) {
		r := &Registry{}
		if got := r.Find("/any/path"); got != nil {
			t.Errorf("Find() = %v, want nil", got)
		}
	})
}

func TestRegistryRemove(t *testing.T) {
	t.Run("remove existing entry returns true", func(t *testing.T) {
		r := &Registry{}
		r.Add(makeTestEntry("/projects/foo", "python"))
		r.Add(makeTestEntry("/projects/bar", "csharp"))

		removed := r.Remove("/projects/foo")
		if !removed {
			t.Error("Remove() = false, want true")
		}
		if len(r.Projects) != 1 {
			t.Errorf("len(Projects) = %d, want 1", len(r.Projects))
		}
		if r.Projects[0].Path != "/projects/bar" {
			t.Errorf("remaining Path = %q, want %q", r.Projects[0].Path, "/projects/bar")
		}
	})

	t.Run("remove non-existent returns false", func(t *testing.T) {
		r := &Registry{}
		r.Add(makeTestEntry("/projects/foo", "python"))

		removed := r.Remove("/projects/nothere")
		if removed {
			t.Error("Remove() = true, want false")
		}
		if len(r.Projects) != 1 {
			t.Error("Remove() should not have altered the registry")
		}
	})

	t.Run("remove from empty registry returns false", func(t *testing.T) {
		r := &Registry{}
		if r.Remove("/any/path") {
			t.Error("Remove() = true on empty registry, want false")
		}
	})
}

func TestRegistryUpdateChecksums(t *testing.T) {
	t.Run("updates checksums and sets LastSync", func(t *testing.T) {
		r := &Registry{}
		entry := makeTestEntry("/projects/foo", "python")
		entry.LastSync = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		r.Add(entry)

		before := time.Now().UTC()
		newChecksums := map[string]string{
			".cursor/rules/new.mdc": "sha256:deadbeef",
		}
		r.UpdateChecksums("/projects/foo", newChecksums)

		got := r.Find("/projects/foo")
		if got == nil {
			t.Fatal("entry disappeared after UpdateChecksums")
		}
		if got.Checksums[".cursor/rules/new.mdc"] != "sha256:deadbeef" {
			t.Errorf("Checksums not updated, got %v", got.Checksums)
		}
		if !got.LastSync.After(before) && !got.LastSync.Equal(before) {
			t.Errorf("LastSync %v should be >= %v", got.LastSync, before)
		}
	})

	t.Run("UpdateChecksums on non-existent path is a no-op", func(t *testing.T) {
		r := &Registry{}
		// Should not panic.
		r.UpdateChecksums("/projects/notregistered", map[string]string{"x": "y"})
		if len(r.Projects) != 0 {
			t.Error("UpdateChecksums on missing path should not add an entry")
		}
	})
}

func TestRegistrySaveAndLoad(t *testing.T) {
	// Redirect registry file to a temp directory.
	// On macOS, os.UserConfigDir() returns $HOME/Library/Application Support.
	// We override HOME so the stdlib resolves to our temp dir.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	// Also clear XDG_CONFIG_HOME so Linux paths don't interfere.
	t.Setenv("XDG_CONFIG_HOME", "")

	t.Run("save and load roundtrip", func(t *testing.T) {
		r := &Registry{}
		entry := makeTestEntry(filepath.Join(tmpHome, "projects", "myapp"), "python")
		entry.Checksums = map[string]string{
			".cursor/rules/global.mdc": "sha256:cafebabe",
			"CLAUDE.md":                "sha256:deadcode",
		}
		r.Add(entry)

		if err := r.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		loaded, err := LoadRegistry()
		if err != nil {
			t.Fatalf("LoadRegistry() error = %v", err)
		}

		if len(loaded.Projects) != 1 {
			t.Fatalf("loaded len(Projects) = %d, want 1", len(loaded.Projects))
		}

		got := loaded.Find(entry.Path)
		if got == nil {
			t.Fatal("Find() returned nil after roundtrip")
		}
		if len(got.Stacks) == 0 || got.Stacks[0] != "python" {
			t.Errorf("Stacks = %v, want [\"python\"]", got.Stacks)
		}
		if len(got.Checksums) != 2 {
			t.Errorf("len(Checksums) = %d, want 2", len(got.Checksums))
		}
		if got.Checksums[".cursor/rules/global.mdc"] != "sha256:cafebabe" {
			t.Errorf("Checksums[global.mdc] = %q, want %q", got.Checksums[".cursor/rules/global.mdc"], "sha256:cafebabe")
		}
	})

	t.Run("LoadRegistry returns empty registry when file does not exist", func(t *testing.T) {
		// Use a fresh temp HOME where no registry has been saved.
		freshHome := t.TempDir()
		t.Setenv("HOME", freshHome)
		t.Setenv("XDG_CONFIG_HOME", "")

		r, err := LoadRegistry()
		if err != nil {
			t.Fatalf("LoadRegistry() error = %v", err)
		}
		if r == nil {
			t.Fatal("LoadRegistry() returned nil registry")
		}
		if len(r.Projects) != 0 {
			t.Errorf("len(Projects) = %d, want 0 for empty registry", len(r.Projects))
		}
	})
}
