package core

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// expectedSHA256 computes the expected "sha256:<hex>" string for known content.
func expectedSHA256(content []byte) string {
	h := sha256.Sum256(content)
	return fmt.Sprintf("sha256:%x", h)
}

func TestFileChecksum(t *testing.T) {
	t.Run("known content produces expected SHA256", func(t *testing.T) {
		dir := t.TempDir()
		content := []byte("hello, world\n")
		path := filepath.Join(dir, "test.txt")
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatal(err)
		}

		want := expectedSHA256(content)
		got, err := FileChecksum(path)
		if err != nil {
			t.Fatalf("FileChecksum() error = %v", err)
		}
		if got != want {
			t.Errorf("FileChecksum() = %q, want %q", got, want)
		}
	})

	t.Run("empty file produces expected SHA256", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "empty.txt")
		if err := os.WriteFile(path, []byte{}, 0o644); err != nil {
			t.Fatal(err)
		}

		want := expectedSHA256([]byte{})
		got, err := FileChecksum(path)
		if err != nil {
			t.Fatalf("FileChecksum() error = %v", err)
		}
		if got != want {
			t.Errorf("FileChecksum() = %q, want %q", got, want)
		}
	})

	t.Run("non-existent file returns error", func(t *testing.T) {
		_, err := FileChecksum(filepath.Join(t.TempDir(), "no_such_file.txt"))
		if err == nil {
			t.Error("FileChecksum() error = nil, want an error for missing file")
		}
	})

	t.Run("different content produces different checksum", func(t *testing.T) {
		dir := t.TempDir()

		path1 := filepath.Join(dir, "a.txt")
		path2 := filepath.Join(dir, "b.txt")
		if err := os.WriteFile(path1, []byte("content A"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path2, []byte("content B"), 0o644); err != nil {
			t.Fatal(err)
		}

		sum1, err := FileChecksum(path1)
		if err != nil {
			t.Fatal(err)
		}
		sum2, err := FileChecksum(path2)
		if err != nil {
			t.Fatal(err)
		}
		if sum1 == sum2 {
			t.Error("different content should produce different checksums")
		}
	})

	t.Run("checksum has sha256: prefix", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "file.txt")
		if err := os.WriteFile(path, []byte("data"), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := FileChecksum(path)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) < 7 || got[:7] != "sha256:" {
			t.Errorf("FileChecksum() = %q, want prefix \"sha256:\"", got)
		}
	})
}

func TestDirChecksums(t *testing.T) {
	t.Run("directory with files produces correct map", func(t *testing.T) {
		dir := t.TempDir()

		files := map[string][]byte{
			"rule.mdc":   []byte("rule content"),
			"another.md": []byte("another file"),
		}
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(dir, name), content, 0o644); err != nil {
				t.Fatal(err)
			}
		}

		got, err := DirChecksums(dir)
		if err != nil {
			t.Fatalf("DirChecksums() error = %v", err)
		}

		if len(got) != len(files) {
			t.Errorf("len(result) = %d, want %d", len(got), len(files))
		}

		for name, content := range files {
			want := expectedSHA256(content)
			if got[name] != want {
				t.Errorf("checksum for %q = %q, want %q", name, got[name], want)
			}
		}
	})

	t.Run("empty directory produces empty map", func(t *testing.T) {
		dir := t.TempDir()

		got, err := DirChecksums(dir)
		if err != nil {
			t.Fatalf("DirChecksums() error = %v", err)
		}
		if len(got) != 0 {
			t.Errorf("len(result) = %d, want 0 for empty directory", len(got))
		}
	})

	t.Run("nested files use relative paths as keys", func(t *testing.T) {
		dir := t.TempDir()
		subDir := filepath.Join(dir, "sub")
		if err := os.MkdirAll(subDir, 0o755); err != nil {
			t.Fatal(err)
		}

		rootContent := []byte("root file")
		subContent := []byte("sub file")
		if err := os.WriteFile(filepath.Join(dir, "root.txt"), rootContent, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(subDir, "nested.txt"), subContent, 0o644); err != nil {
			t.Fatal(err)
		}

		got, err := DirChecksums(dir)
		if err != nil {
			t.Fatalf("DirChecksums() error = %v", err)
		}

		if len(got) != 2 {
			t.Errorf("len(result) = %d, want 2", len(got))
		}

		rootKey := "root.txt"
		subKey := filepath.Join("sub", "nested.txt")

		if got[rootKey] != expectedSHA256(rootContent) {
			t.Errorf("root file checksum mismatch")
		}
		if got[subKey] != expectedSHA256(subContent) {
			t.Errorf("nested file checksum mismatch, key %q not found or wrong value", subKey)
		}
	})
}
