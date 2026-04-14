package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gentleman-programming/project-accelerator/internal/config"
)

// syncTestSetup creates a minimal environment for Sync tests:
//   - A PA_HOME with one universal rules file and one stack rule file.
//   - A target project directory with .cursor/rules/ already populated.
//   - A registry entry for the project with checksums reflecting what was
//     originally scaffold-copied.
//
// It returns the paHome, targetDir, the registry entry pointer (already added
// to a saved registry), and the manifest.
func syncTestSetup(t *testing.T, universalContent, stackContent string) (
	paHome, targetDir string,
	entry config.ProjectEntry,
	manifest *config.Manifest,
) {
	t.Helper()

	// Redirect registry to temp home.
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")

	paHome = t.TempDir()

	// Create source rule files.
	universalRulesDir := "universal/rules"
	universalSrc := filepath.Join(paHome, universalRulesDir)
	if err := os.MkdirAll(universalSrc, 0o755); err != nil {
		t.Fatal(err)
	}
	universalFile := filepath.Join(universalSrc, "global.mdc")
	if err := os.WriteFile(universalFile, []byte(universalContent), 0o644); err != nil {
		t.Fatal(err)
	}

	stackRulesDir := "stacks/mystack/rules"
	stackSrc := filepath.Join(paHome, stackRulesDir)
	if err := os.MkdirAll(stackSrc, 0o755); err != nil {
		t.Fatal(err)
	}
	stackFile := filepath.Join(stackSrc, "stack.mdc")
	if err := os.WriteFile(stackFile, []byte(stackContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Create target project with .cursor/rules/.
	targetDir = t.TempDir()
	rulesDir := filepath.Join(targetDir, ".cursor", "rules")
	if err := os.MkdirAll(rulesDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Copy source files to target (as if scaffold had run).
	uDst := filepath.Join(rulesDir, "global.mdc")
	if err := os.WriteFile(uDst, []byte(universalContent), 0o644); err != nil {
		t.Fatal(err)
	}
	sDst := filepath.Join(rulesDir, "stack.mdc")
	if err := os.WriteFile(sDst, []byte(stackContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Compute checksums as they were at scaffold time.
	uSum, err := FileChecksum(uDst)
	if err != nil {
		t.Fatal(err)
	}
	sSum, err := FileChecksum(sDst)
	if err != nil {
		t.Fatal(err)
	}

	// Build registry entry.
	entry = config.ProjectEntry{
		Path:         targetDir,
		Stacks:       []string{"mystack"},
		RegisteredAt: time.Now().UTC(),
		LastSync:     time.Now().UTC(),
		Checksums: map[string]string{
			filepath.Join(".cursor", "rules", "global.mdc"): uSum,
			filepath.Join(".cursor", "rules", "stack.mdc"):  sSum,
		},
	}

	// Save registry.
	reg := &config.Registry{}
	reg.Add(entry)
	if err := reg.Save(); err != nil {
		t.Fatal(err)
	}

	// Build manifest.
	manifest = &config.Manifest{
		Version:           1,
		UniversalRulesDir: universalRulesDir,
		Stacks: map[string]config.StackConfig{
			"mystack": {
				Name:     "MyStack",
				RulesDir: stackRulesDir,
			},
		},
	}

	return paHome, targetDir, entry, manifest
}

func TestSync(t *testing.T) {
	t.Run("adds new files not present in target", func(t *testing.T) {
		paHome, targetDir, entry, manifest := syncTestSetup(t, "# universal", "# stack")

		// Add a new rule file to the universal source that did NOT exist at scaffold time.
		newFileSrc := filepath.Join(paHome, "universal", "rules", "new-rule.mdc")
		if err := os.WriteFile(newFileSrc, []byte("# new rule"), 0o644); err != nil {
			t.Fatal(err)
		}

		result, err := Sync(paHome, &entry, manifest)
		if err != nil {
			t.Fatalf("Sync() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("Sync() errors = %v", result.Errors)
		}
		if result.Added != 1 {
			t.Errorf("Added = %d, want 1", result.Added)
		}

		newFileDst := filepath.Join(targetDir, ".cursor", "rules", "new-rule.mdc")
		if _, err := os.Stat(newFileDst); os.IsNotExist(err) {
			t.Error("new-rule.mdc should have been added to target")
		}
	})

	t.Run("updates files when target matches recorded checksum", func(t *testing.T) {
		// Use a manifest that only has universal rules (no stack rules) so exactly
		// one source file is involved, making the Updated count deterministic.
		tmpHome := t.TempDir()
		t.Setenv("HOME", tmpHome)
		t.Setenv("XDG_CONFIG_HOME", "")

		localPAHome := t.TempDir()
		universalRulesDir := "universal/rules"
		universalSrc := filepath.Join(localPAHome, universalRulesDir)
		if err := os.MkdirAll(universalSrc, 0o755); err != nil {
			t.Fatal(err)
		}
		initialContent := "# universal v1"
		uFile := filepath.Join(universalSrc, "global.mdc")
		if err := os.WriteFile(uFile, []byte(initialContent), 0o644); err != nil {
			t.Fatal(err)
		}

		// Target project directory with the same content (as if scaffold ran).
		targetDir := t.TempDir()
		rulesDir := filepath.Join(targetDir, ".cursor", "rules")
		if err := os.MkdirAll(rulesDir, 0o755); err != nil {
			t.Fatal(err)
		}
		dstPath := filepath.Join(rulesDir, "global.mdc")
		if err := os.WriteFile(dstPath, []byte(initialContent), 0o644); err != nil {
			t.Fatal(err)
		}

		uSum, err := FileChecksum(dstPath)
		if err != nil {
			t.Fatal(err)
		}

		registryKey := filepath.Join(".cursor", "rules", "global.mdc")
		entry := config.ProjectEntry{
			Path:         targetDir,
			Stacks:       []string{"mystack"},
			RegisteredAt: time.Now().UTC(),
			LastSync:     time.Now().UTC(),
			Checksums:    map[string]string{registryKey: uSum},
		}
		reg := &config.Registry{}
		reg.Add(entry)
		if err := reg.Save(); err != nil {
			t.Fatal(err)
		}

		localManifest := &config.Manifest{
			Version:           1,
			UniversalRulesDir: universalRulesDir,
			Stacks: map[string]config.StackConfig{
				"mystack": {Name: "MyStack"},
			},
		}

		// Now update the source file to a new version.
		updatedContent := "# universal v2 — updated"
		if err := os.WriteFile(uFile, []byte(updatedContent), 0o644); err != nil {
			t.Fatal(err)
		}

		result, err := Sync(localPAHome, &entry, localManifest)
		if err != nil {
			t.Fatalf("Sync() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("Sync() errors = %v", result.Errors)
		}
		if result.Updated != 1 {
			t.Errorf("Updated = %d, want 1", result.Updated)
		}

		data, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != updatedContent {
			t.Errorf("file content = %q, want %q", string(data), updatedContent)
		}
	})

	t.Run("preserves overrides when target differs from recorded checksum", func(t *testing.T) {
		paHome, targetDir, entry, manifest := syncTestSetup(t, "# universal", "# stack")

		// User modifies the local global.mdc — this is an override.
		dstPath := filepath.Join(targetDir, ".cursor", "rules", "global.mdc")
		userContent := "# USER OVERRIDE — do not touch"
		if err := os.WriteFile(dstPath, []byte(userContent), 0o644); err != nil {
			t.Fatal(err)
		}

		// PA_HOME also has a new version of global.mdc.
		newSourceContent := "# universal v2"
		universalSrc := filepath.Join(paHome, "universal", "rules", "global.mdc")
		if err := os.WriteFile(universalSrc, []byte(newSourceContent), 0o644); err != nil {
			t.Fatal(err)
		}

		result, err := Sync(paHome, &entry, manifest)
		if err != nil {
			t.Fatalf("Sync() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("Sync() errors = %v", result.Errors)
		}
		if result.Overrides != 1 {
			t.Errorf("Overrides = %d, want 1", result.Overrides)
		}

		// The local override must NOT have been replaced.
		data, err := os.ReadFile(dstPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != userContent {
			t.Errorf("override was overwritten; content = %q, want %q", string(data), userContent)
		}
	})

	t.Run("skips files with no recorded checksum", func(t *testing.T) {
		paHome, targetDir, entry, manifest := syncTestSetup(t, "# universal", "# stack")

		// Place a file in the target rules dir that has no registry record (unknown provenance).
		unknownPath := filepath.Join(targetDir, ".cursor", "rules", "unknown.mdc")
		if err := os.WriteFile(unknownPath, []byte("# unknown"), 0o644); err != nil {
			t.Fatal(err)
		}
		// Also put the same file in the PA_HOME source so Sync will encounter it.
		unknownSrc := filepath.Join(paHome, "universal", "rules", "unknown.mdc")
		if err := os.WriteFile(unknownSrc, []byte("# unknown from source"), 0o644); err != nil {
			t.Fatal(err)
		}
		// Do NOT add a checksum for unknown.mdc to the entry.

		result, err := Sync(paHome, &entry, manifest)
		if err != nil {
			t.Fatalf("Sync() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("Sync() errors = %v", result.Errors)
		}
		if result.Skipped != 1 {
			t.Errorf("Skipped = %d, want 1", result.Skipped)
		}

		// The file should remain unchanged (not overwritten from source).
		data, err := os.ReadFile(unknownPath)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "# unknown" {
			t.Errorf("skipped file was modified; content = %q, want %q", string(data), "# unknown")
		}
	})

	t.Run("stack not found in manifest returns error", func(t *testing.T) {
		_, targetDir, entry, _ := syncTestSetup(t, "# universal", "# stack")
		emptyManifest := &config.Manifest{
			Version: 1,
			Stacks:  map[string]config.StackConfig{},
		}
		paHome := t.TempDir()

		_, err := Sync(paHome, &entry, emptyManifest)
		if err == nil {
			t.Errorf("Sync() error = nil, want error for unknown stacks %v in %q", entry.Stacks, targetDir)
		}
	})

	t.Run("sync updates registry checksums and LastSync", func(t *testing.T) {
		paHome, targetDir, entry, manifest := syncTestSetup(t, "# universal", "# stack")

		oldLastSync := entry.LastSync

		// Update source so Sync has something to update.
		universalSrc := filepath.Join(paHome, "universal", "rules", "global.mdc")
		if err := os.WriteFile(universalSrc, []byte("# universal v2"), 0o644); err != nil {
			t.Fatal(err)
		}

		_, err := Sync(paHome, &entry, manifest)
		if err != nil {
			t.Fatalf("Sync() error = %v", err)
		}

		// Re-load registry and check the entry was updated.
		reg, err := config.LoadRegistry()
		if err != nil {
			t.Fatal(err)
		}
		got := reg.Find(targetDir)
		if got == nil {
			t.Fatal("entry not found in registry after sync")
		}

		if !got.LastSync.After(oldLastSync) {
			t.Errorf("LastSync %v should be after %v", got.LastSync, oldLastSync)
		}

		key := filepath.Join(".cursor", "rules", "global.mdc")
		if _, ok := got.Checksums[key]; !ok {
			t.Errorf("checksum for %q not in registry after sync", key)
		}
	})
}
