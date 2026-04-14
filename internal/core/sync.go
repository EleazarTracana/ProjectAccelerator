package core

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gentleman-programming/project-accelerator/internal/config"
)

// SyncAction records what happened to one file during a sync pass.
type SyncAction struct {
	RelPath string // Path relative to the .cursor/rules destination dir.
	Action  string // "updated", "added", "skipped", "override_preserved"
	Reason  string // Human-readable explanation.
}

// SyncResult summarises the outcome of a Sync call.
type SyncResult struct {
	Actions   []SyncAction
	Updated   int
	Added     int
	Skipped   int
	Overrides int
	Errors    []string
}

// Sync synchronises Cursor rules from the PA source into a registered project.
//
// For every source rule file (universal + stack-specific) it compares the
// source checksum against what was recorded at scaffold/last-sync time and
// the current on-disk state of the target file, then decides one of:
//
//   - "added"              – file not present in target; copy it now.
//   - "updated"            – file exists and matches the recorded checksum;
//     safe to overwrite with the new version.
//   - "override_preserved" – file exists but differs from the recorded
//     checksum; the user has modified it locally; leave it alone.
//   - "skipped"            – file exists but has no recorded checksum; treat
//     conservatively as a potential override and leave it alone.
//
// After the sync the entry's Checksums and LastSync are updated in the
// registry and saved.
func Sync(paHome string, entry *config.ProjectEntry, manifest *config.Manifest) (*SyncResult, error) {
	// Validate all stacks exist in the manifest before doing any work.
	for _, stackKey := range entry.Stacks {
		if _, ok := manifest.Stacks[stackKey]; !ok {
			return nil, fmt.Errorf("stack %q not found in manifest", stackKey)
		}
	}

	result := &SyncResult{}
	newChecksums := make(map[string]string, len(entry.Checksums))

	// Copy existing checksums so unrelated files are preserved.
	for k, v := range entry.Checksums {
		newChecksums[k] = v
	}

	// Collect source directories: universal first, then each stack's rules.
	type sourceDir struct {
		dir string // Absolute path to the source rule directory.
	}

	var sourceDirs []sourceDir
	if manifest.UniversalRulesDir != "" {
		sourceDirs = append(sourceDirs, sourceDir{filepath.Join(paHome, manifest.UniversalRulesDir)})
	}
	for _, stackKey := range entry.Stacks {
		stack := manifest.Stacks[stackKey]
		if stack.RulesDir != "" {
			sourceDirs = append(sourceDirs, sourceDir{filepath.Join(paHome, stack.RulesDir)})
		}
	}

	rulesDestDir := filepath.Join(entry.Path, ".cursor", "rules")

	for _, src := range sourceDirs {
		if _, err := os.Stat(src.dir); os.IsNotExist(err) {
			// Source directory absent — skip silently.
			continue
		}

		err := filepath.WalkDir(src.dir, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				return nil
			}

			// Use just the filename as the relative key inside .cursor/rules.
			filename := d.Name()
			registryKey := filepath.Join(".cursor", "rules", filename)

			// 1. Source checksum.
			srcSum, err := FileChecksum(path)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("checksumming source %q: %v", path, err))
				return nil // Non-fatal: skip this file.
			}

			dstPath := filepath.Join(rulesDestDir, filename)

			// 2. Does the target exist?
			if _, err := os.Stat(dstPath); os.IsNotExist(err) {
				// File not present → add it.
				if copyErr := copyFile(path, dstPath); copyErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("adding %q: %v", filename, copyErr))
					return nil
				}
				newChecksums[registryKey] = srcSum
				result.Actions = append(result.Actions, SyncAction{
					RelPath: filename,
					Action:  "added",
					Reason:  "file did not exist in target",
				})
				result.Added++
				return nil
			}

			// 3. Target exists — compute its current checksum.
			dstSum, err := FileChecksum(dstPath)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("checksumming target %q: %v", dstPath, err))
				return nil
			}

			recordedSum, hasRecord := entry.Checksums[registryKey]

			switch {
			case !hasRecord:
				// No recorded checksum → unknown provenance; leave it alone.
				result.Actions = append(result.Actions, SyncAction{
					RelPath: filename,
					Action:  "skipped",
					Reason:  "no recorded checksum; treating as potential local override",
				})
				result.Skipped++

			case dstSum != recordedSum:
				// Target differs from what we last wrote → user edited it.
				result.Actions = append(result.Actions, SyncAction{
					RelPath: filename,
					Action:  "override_preserved",
					Reason:  fmt.Sprintf("local file (checksum %s) differs from recorded checksum %s", dstSum, recordedSum),
				})
				result.Overrides++

			default:
				// Target checksum matches the recorded value → safe to update.
				if copyErr := copyFile(path, dstPath); copyErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("updating %q: %v", filename, copyErr))
					return nil
				}
				newChecksums[registryKey] = srcSum
				result.Actions = append(result.Actions, SyncAction{
					RelPath: filename,
					Action:  "updated",
					Reason:  fmt.Sprintf("source checksum changed from %s to %s", recordedSum, srcSum),
				})
				result.Updated++
			}

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walking source directory %q: %w", src.dir, err)
		}
	}

	// Persist the updated checksums and timestamp back to the registry.
	registry, err := config.LoadRegistry()
	if err != nil {
		return result, fmt.Errorf("loading registry for checksum update: %w", err)
	}

	registry.UpdateChecksums(entry.Path, newChecksums)

	if err := registry.Save(); err != nil {
		return result, fmt.Errorf("saving registry after sync: %w", err)
	}

	return result, nil
}
