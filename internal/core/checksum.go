package core

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// FileChecksum computes the SHA256 hash of the file at path.
// Returns the digest as "sha256:<hex>".
func FileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("opening file for checksum %q: %w", path, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("hashing file %q: %w", path, err)
	}

	return fmt.Sprintf("sha256:%x", h.Sum(nil)), nil
}

// DirChecksums computes SHA256 checksums for all regular files under dir,
// walking recursively. Returns a map of path relative to dir → checksum.
// Symlinks are followed; directories themselves are not included.
func DirChecksums(dir string) (map[string]string, error) {
	results := make(map[string]string)

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("walking %q: %w", path, err)
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			// Skip symlinks to avoid surprises.
			return nil
		}

		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return fmt.Errorf("computing relative path for %q: %w", path, err)
		}

		sum, err := FileChecksum(path)
		if err != nil {
			return err
		}

		results[rel] = sum
		return nil
	})
	if err != nil {
		return nil, err
	}

	return results, nil
}
