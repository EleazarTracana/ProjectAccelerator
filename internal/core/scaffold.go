package core

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/gentleman-programming/project-accelerator/internal/config"
)

// ScaffoldOptions controls which scaffold steps are executed.
type ScaffoldOptions struct {
	GitInit      bool
	CopyRules    bool
	ClaudeMD     bool
	CopySkeleton bool
	Register     bool
}

// ScaffoldResult reports what the scaffold operation did.
type ScaffoldResult struct {
	RulesCopied     int
	SkeletonCopied  int
	ClaudeMDCreated bool
	GitInitDone     bool
	Registered      bool
	Errors          []string
}

// Scaffold executes the scaffold process for a target directory.
//
//   - paHome            – resolved absolute path to the ProjectAccelerator root.
//   - targetDir         – destination directory (created if absent).
//   - stacks            – map of stackKey → StackConfig for all selected stacks.
//   - universalRulesDir – manifest-level universal rules dir (relative to paHome).
//   - opts              – which steps to perform.
func Scaffold(paHome, targetDir string, stacks map[string]config.StackConfig, universalRulesDir string, opts ScaffoldOptions) (*ScaffoldResult, error) {
	result := &ScaffoldResult{}

	// Ensure the target directory exists.
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, fmt.Errorf("creating target directory %q: %w", targetDir, err)
	}

	// Track checksums for files we copy so they can be stored in the registry.
	copiedChecksums := make(map[string]string)

	// ── 1. Copy rules ────────────────────────────────────────────────────────
	if opts.CopyRules {
		rulesDestDir := filepath.Join(targetDir, ".cursor", "rules")
		if err := os.MkdirAll(rulesDestDir, 0o755); err != nil {
			return nil, fmt.Errorf("creating rules destination directory: %w", err)
		}

		// Universal rules (copied once, regardless of how many stacks are selected).
		if universalRulesDir != "" {
			uSrc := filepath.Join(paHome, universalRulesDir)
			n, checksums, err := copyDirFlat(uSrc, rulesDestDir)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("copying universal rules: %v", err))
			} else {
				result.RulesCopied += n
				for rel, sum := range checksums {
					copiedChecksums[filepath.Join(".cursor", "rules", rel)] = sum
				}
			}
		}

		// Stack-specific rules for each selected stack.
		for stackKey, stack := range stacks {
			if stack.RulesDir != "" {
				sSrc := filepath.Join(paHome, stack.RulesDir)
				n, checksums, err := copyDirFlat(sSrc, rulesDestDir)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("copying stack rules for %q: %v", stackKey, err))
				} else {
					result.RulesCopied += n
					for rel, sum := range checksums {
						copiedChecksums[filepath.Join(".cursor", "rules", rel)] = sum
					}
				}
			}
		}
	}

	// ── 2. Copy skeleton ─────────────────────────────────────────────────────
	if opts.CopySkeleton {
		for stackKey, stack := range stacks {
			if stack.SkeletonDir != "" {
				skelSrc := filepath.Join(paHome, stack.SkeletonDir)
				n, checksums, err := copySkeleton(skelSrc, targetDir)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("copying skeleton for %q: %v", stackKey, err))
				} else {
					result.SkeletonCopied += n
					for rel, sum := range checksums {
						copiedChecksums[rel] = sum
					}
				}
			}
		}
	}

	// ── 3. Render CLAUDE.md ──────────────────────────────────────────────────
	if opts.ClaudeMD {
		claudeDst := filepath.Join(targetDir, "CLAUDE.md")
		// Don't overwrite an existing CLAUDE.md.
		if _, err := os.Stat(claudeDst); os.IsNotExist(err) {
			created, err := renderMultiClaudeMD(paHome, targetDir, stacks)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("rendering CLAUDE.md: %v", err))
			} else {
				result.ClaudeMDCreated = created
				if created {
					sum, err := FileChecksum(claudeDst)
					if err == nil {
						copiedChecksums["CLAUDE.md"] = sum
					}
				}
			}
		}
	}

	// ── 4. git init ──────────────────────────────────────────────────────────
	if opts.GitInit {
		cmd := exec.Command("git", "init", targetDir)
		cmd.Dir = targetDir
		if out, err := cmd.CombinedOutput(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("git init: %v — %s", err, strings.TrimSpace(string(out))))
		} else {
			result.GitInitDone = true
		}
	}

	// ── 5. Register project ──────────────────────────────────────────────────
	if opts.Register {
		stackKeys := make([]string, 0, len(stacks))
		for k := range stacks {
			stackKeys = append(stackKeys, k)
		}
		if err := registerProject(targetDir, stackKeys, copiedChecksums); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("registering project: %v", err))
		} else {
			result.Registered = true
		}
	}

	return result, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

// copyDirFlat copies all regular files from srcDir into dstDir (flat — no
// subdirectory structure is preserved). Returns the number of files copied
// and a map of filename → checksum.
func copyDirFlat(srcDir, dstDir string) (int, map[string]string, error) {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		// Nothing to copy — treat as a no-op rather than an error.
		return 0, nil, nil
	}

	count := 0
	checksums := make(map[string]string)

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		dstPath := filepath.Join(dstDir, d.Name())
		if copyErr := copyFile(path, dstPath); copyErr != nil {
			return copyErr
		}

		sum, hashErr := FileChecksum(dstPath)
		if hashErr != nil {
			return hashErr
		}

		checksums[d.Name()] = sum
		count++
		return nil
	})
	if err != nil {
		return 0, nil, err
	}

	return count, checksums, nil
}

// copySkeleton copies srcDir into dstDir recursively, preserving directory
// structure. Files ending in ".template" are copied with that suffix stripped.
// Hidden directories (name starting with ".") are skipped, except that hidden
// files inside visible directories are copied normally.
// Returns the number of files copied and a map of relative path → checksum.
func copySkeleton(srcDir, dstDir string) (int, map[string]string, error) {
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return 0, nil, nil
	}

	count := 0
	checksums := make(map[string]string)

	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, relErr := filepath.Rel(srcDir, path)
		if relErr != nil {
			return relErr
		}

		// Skip hidden directories (e.g. .git) but not the root itself.
		if d.IsDir() {
			if rel != "." && strings.HasPrefix(d.Name(), ".") {
				return fs.SkipDir
			}
			return nil
		}

		// Determine destination name: strip .template suffix if present.
		destName := d.Name()
		if strings.HasSuffix(destName, ".template") {
			destName = strings.TrimSuffix(destName, ".template")
		}

		destRel := filepath.Join(filepath.Dir(rel), destName)
		dstPath := filepath.Join(dstDir, destRel)

		if mkErr := os.MkdirAll(filepath.Dir(dstPath), 0o755); mkErr != nil {
			return mkErr
		}

		// Skip files that already exist in the target (don't overwrite user's files).
		if _, err := os.Stat(dstPath); err == nil {
			return nil
		}

		if copyErr := copyFile(path, dstPath); copyErr != nil {
			return copyErr
		}

		sum, hashErr := FileChecksum(dstPath)
		if hashErr != nil {
			return hashErr
		}

		checksums[destRel] = sum
		count++
		return nil
	})
	if err != nil {
		return 0, nil, err
	}

	return count, checksums, nil
}

// claudeTemplateVars holds the values injected into a CLAUDE.md template.
type claudeTemplateVars struct {
	ProjectName string
	Stack       string
	Date        string
}

// renderTemplateSection reads a Go text/template file and renders it with
// project vars, returning the rendered content as a string.
// Returns empty string (no error) if the template file does not exist.
func renderTemplateSection(tmplPath, projectName, stackKey string) (string, error) {
	raw, err := os.ReadFile(tmplPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("reading CLAUDE.md template %q: %w", tmplPath, err)
	}

	tmpl, err := template.New("claude").Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("parsing CLAUDE.md template %q: %w", tmplPath, err)
	}

	vars := claudeTemplateVars{
		ProjectName: projectName,
		Stack:       stackKey,
		Date:        time.Now().Format("2006-01-02"),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("rendering CLAUDE.md template %q: %w", tmplPath, err)
	}

	return buf.String(), nil
}

// renderMultiClaudeMD renders CLAUDE.md by concatenating the rendered template
// sections from all stacks that define a claude_template. Each section is
// prefixed with a "## {StackName} Stack" header, and the first line of each
// rendered template is stripped if it starts with "# " (the per-template
// project header). A single "# {ProjectName}" header is written at the top.
//
// Returns true if the file was written (at least one template contributed
// content), false if no stack had a claude_template.
func renderMultiClaudeMD(paHome, targetDir string, stacks map[string]config.StackConfig) (bool, error) {
	projectName := filepath.Base(targetDir)
	date := time.Now().Format("2006-01-02")

	var combined strings.Builder
	anyContent := false

	for stackKey, stack := range stacks {
		if stack.ClaudeTemplate == "" {
			continue
		}
		tmplPath := filepath.Join(paHome, stack.ClaudeTemplate)
		rendered, err := renderTemplateSection(tmplPath, projectName, stackKey)
		if err != nil {
			return false, err
		}
		if rendered == "" {
			continue
		}

		// Strip the first line if it starts with "# " (per-template project header).
		lines := strings.SplitN(rendered, "\n", 2)
		body := rendered
		if len(lines) > 0 && strings.HasPrefix(lines[0], "# ") {
			if len(lines) > 1 {
				body = strings.TrimLeft(lines[1], "\n")
			} else {
				body = ""
			}
		}

		if body == "" {
			continue
		}

		// Write stack section header.
		stackName := stack.Name
		if stackName == "" {
			stackName = stackKey
		}
		combined.WriteString(fmt.Sprintf("## %s Stack\n\n", stackName))
		combined.WriteString(body)
		if !strings.HasSuffix(body, "\n") {
			combined.WriteString("\n")
		}
		combined.WriteString("\n")
		anyContent = true
	}

	if !anyContent {
		return false, nil
	}

	// Build the final file: single top-level header + all stack sections.
	var out strings.Builder
	out.WriteString(fmt.Sprintf("# %s\n\nCreated: %s\n\n", projectName, date))
	out.WriteString(combined.String())

	destPath := filepath.Join(targetDir, "CLAUDE.md")
	if err := os.WriteFile(destPath, []byte(out.String()), 0o644); err != nil {
		return false, fmt.Errorf("writing CLAUDE.md: %w", err)
	}

	return true, nil
}

// copyFile copies the file at src to dst, creating dst if necessary.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening source file %q: %w", src, err)
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("creating directory for %q: %w", dst, err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating destination file %q: %w", dst, err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copying %q → %q: %w", src, dst, err)
	}

	return out.Sync()
}

// registerProject persists a new ProjectEntry via the config registry API.
func registerProject(projectPath string, stackKeys []string, checksums map[string]string) error {
	registry, err := config.LoadRegistry()
	if err != nil {
		return fmt.Errorf("loading registry: %w", err)
	}

	now := time.Now().UTC()

	entry := config.ProjectEntry{
		Path:         projectPath,
		Stacks:       stackKeys,
		RegisteredAt: now,
		LastSync:     now,
		Checksums:    checksums,
	}

	registry.Add(entry)

	if err := registry.Save(); err != nil {
		return fmt.Errorf("saving registry: %w", err)
	}

	return nil
}
