package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gentleman-programming/project-accelerator/internal/config"
)

// setupPAHome creates a fake PA_HOME directory structure:
//   - universal/rules/  with given universal rule files
//   - stacks/<stack>/rules/ with given stack rule files
//   - stacks/<stack>/skeleton/ with given skeleton files
//   - stacks/<stack>/CLAUDE.md.template
func setupPAHome(t *testing.T, universalFiles, stackRuleFiles, skeletonFiles map[string]string, claudeTemplate string) (paHome string, stacks map[string]config.StackConfig, universalRulesDir string) {
	t.Helper()
	paHome = t.TempDir()

	universalRulesDir = "universal/rules"
	universalDir := filepath.Join(paHome, universalRulesDir)
	if err := os.MkdirAll(universalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range universalFiles {
		if err := os.WriteFile(filepath.Join(universalDir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	stackRulesDir := "stacks/mystack/rules"
	stackRulesPath := filepath.Join(paHome, stackRulesDir)
	if err := os.MkdirAll(stackRulesPath, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range stackRuleFiles {
		if err := os.WriteFile(filepath.Join(stackRulesPath, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	skeletonDir := "stacks/mystack/skeleton"
	skeletonPath := filepath.Join(paHome, skeletonDir)
	if err := os.MkdirAll(skeletonPath, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, content := range skeletonFiles {
		if err := os.WriteFile(filepath.Join(skeletonPath, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	tmplDir := "stacks/mystack"
	if claudeTemplate != "" {
		if err := os.WriteFile(filepath.Join(paHome, tmplDir, "CLAUDE.md.template"), []byte(claudeTemplate), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	stacks = map[string]config.StackConfig{
		"mystack": {
			Name:           "MyStack",
			RulesDir:       stackRulesDir,
			SkeletonDir:    skeletonDir,
			ClaudeTemplate: filepath.Join(tmplDir, "CLAUDE.md.template"),
		},
	}

	return paHome, stacks, universalRulesDir
}

// redirectRegistryForScaffold overrides HOME so that scaffold's registerProject
// saves the registry into a temp dir, not the real user home.
func redirectRegistryForScaffold(t *testing.T) {
	t.Helper()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")
}

func TestScaffold(t *testing.T) {
	redirectRegistryForScaffold(t)

	universalFiles := map[string]string{
		"global.mdc": "# global rule",
	}
	stackRuleFiles := map[string]string{
		"stack-specific.mdc": "# stack rule",
	}
	skeletonFiles := map[string]string{
		"main.py.template": "# main",
		"README.md":        "# readme",
	}
	claudeTemplate := "# {{.ProjectName}} ({{.Stack}})\nCreated: {{.Date}}\n"

	paHome, stacks, universalRulesDir := setupPAHome(t,
		universalFiles,
		stackRuleFiles,
		skeletonFiles,
		claudeTemplate,
	)

	t.Run("copies universal rules to .cursor/rules/", func(t *testing.T) {
		targetDir := t.TempDir()

		result, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{
			CopyRules: true,
		})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("Scaffold() errors = %v", result.Errors)
		}

		dst := filepath.Join(targetDir, ".cursor", "rules", "global.mdc")
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			t.Errorf("universal rule file not found at %s", dst)
		}
	})

	t.Run("copies stack rules to .cursor/rules/", func(t *testing.T) {
		targetDir := t.TempDir()

		result, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{
			CopyRules: true,
		})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("Scaffold() errors = %v", result.Errors)
		}

		dst := filepath.Join(targetDir, ".cursor", "rules", "stack-specific.mdc")
		if _, err := os.Stat(dst); os.IsNotExist(err) {
			t.Errorf("stack rule file not found at %s", dst)
		}

		if result.RulesCopied != 2 {
			t.Errorf("RulesCopied = %d, want 2 (1 universal + 1 stack)", result.RulesCopied)
		}
	})

	t.Run("renders CLAUDE.md from template", func(t *testing.T) {
		targetDir := t.TempDir()

		result, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{
			ClaudeMD: true,
		})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("Scaffold() errors = %v", result.Errors)
		}
		if !result.ClaudeMDCreated {
			t.Error("ClaudeMDCreated = false, want true")
		}

		claudePath := filepath.Join(targetDir, "CLAUDE.md")
		data, err := os.ReadFile(claudePath)
		if err != nil {
			t.Fatalf("CLAUDE.md not found: %v", err)
		}

		projectName := filepath.Base(targetDir)
		content := string(data)
		if len(content) == 0 {
			t.Error("CLAUDE.md is empty")
		}
		// The project name should appear in the rendered output (top-level header).
		if !strings.Contains(content, "# "+projectName) {
			t.Errorf("CLAUDE.md content = %q, expected to contain top-level header %q", content, "# "+projectName)
		}
		// The stack section header should appear.
		if !strings.Contains(content, "## MyStack Stack") {
			t.Errorf("CLAUDE.md content = %q, expected to contain stack section %q", content, "## MyStack Stack")
		}
	})

	t.Run("copies skeleton stripping .template extension", func(t *testing.T) {
		targetDir := t.TempDir()

		result, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{
			CopySkeleton: true,
		})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("Scaffold() errors = %v", result.Errors)
		}

		// main.py.template → main.py
		mainPy := filepath.Join(targetDir, "main.py")
		if _, err := os.Stat(mainPy); os.IsNotExist(err) {
			t.Error("main.py not found; .template extension should have been stripped")
		}
		// .template source file should NOT appear in target.
		mainPyTemplate := filepath.Join(targetDir, "main.py.template")
		if _, err := os.Stat(mainPyTemplate); err == nil {
			t.Error("main.py.template should not exist in target")
		}

		// README.md (no .template suffix) → README.md
		readme := filepath.Join(targetDir, "README.md")
		if _, err := os.Stat(readme); os.IsNotExist(err) {
			t.Error("README.md not found")
		}

		if result.SkeletonCopied != 2 {
			t.Errorf("SkeletonCopied = %d, want 2", result.SkeletonCopied)
		}
	})

	t.Run("registers project in registry when Register=true", func(t *testing.T) {
		targetDir := t.TempDir()

		result, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{
			CopyRules: true,
			Register:  true,
		})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Fatalf("Scaffold() errors = %v", result.Errors)
		}
		if !result.Registered {
			t.Error("Registered = false, want true")
		}

		reg, err := config.LoadRegistry()
		if err != nil {
			t.Fatalf("LoadRegistry() error = %v", err)
		}
		entry := reg.Find(targetDir)
		if entry == nil {
			t.Fatalf("project %q not found in registry after scaffold", targetDir)
		}
		if len(entry.Stacks) == 0 || entry.Stacks[0] != "mystack" {
			t.Errorf("registered Stacks = %v, want [\"mystack\"]", entry.Stacks)
		}
	})

	t.Run("partial options — only rules, no skeleton", func(t *testing.T) {
		targetDir := t.TempDir()

		result, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{
			CopyRules:    true,
			CopySkeleton: false,
			ClaudeMD:     false,
			Register:     false,
		})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}

		if result.SkeletonCopied != 0 {
			t.Errorf("SkeletonCopied = %d, want 0 when CopySkeleton=false", result.SkeletonCopied)
		}
		if result.ClaudeMDCreated {
			t.Error("ClaudeMDCreated = true, want false when ClaudeMD=false")
		}
		if result.Registered {
			t.Error("Registered = true, want false when Register=false")
		}
		if result.RulesCopied == 0 {
			t.Error("RulesCopied = 0, want > 0 when CopyRules=true")
		}
	})

	t.Run("partial options — no options does not fail", func(t *testing.T) {
		targetDir := t.TempDir()

		result, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}
		if len(result.Errors) > 0 {
			t.Errorf("Scaffold() unexpected errors = %v", result.Errors)
		}
	})

	t.Run("creates target directory if it does not exist", func(t *testing.T) {
		base := t.TempDir()
		targetDir := filepath.Join(base, "new", "project")

		_, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			t.Errorf("target directory %q was not created", targetDir)
		}
	})
}

func TestScaffoldChecksums(t *testing.T) {
	redirectRegistryForScaffold(t)

	universalFiles := map[string]string{
		"global.mdc": "rule content",
	}
	paHome, stacks, universalRulesDir := setupPAHome(t, universalFiles, nil, nil, "")

	t.Run("registry entry has checksums for copied rules", func(t *testing.T) {
		targetDir := t.TempDir()

		_, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{
			CopyRules: true,
			Register:  true,
		})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}

		reg, err := config.LoadRegistry()
		if err != nil {
			t.Fatalf("LoadRegistry() error = %v", err)
		}
		entry := reg.Find(targetDir)
		if entry == nil {
			t.Fatal("project not in registry")
		}

		key := filepath.Join(".cursor", "rules", "global.mdc")
		sum, ok := entry.Checksums[key]
		if !ok {
			t.Errorf("checksum for %q not found in registry entry", key)
		}
		if len(sum) < 7 || sum[:7] != "sha256:" {
			t.Errorf("checksum %q does not have expected sha256: prefix", sum)
		}
	})

	t.Run("RegisteredAt is set to a recent time", func(t *testing.T) {
		targetDir := t.TempDir()
		before := time.Now().UTC().Add(-time.Second)

		_, err := Scaffold(paHome, targetDir, stacks, universalRulesDir, ScaffoldOptions{
			Register: true,
		})
		if err != nil {
			t.Fatalf("Scaffold() error = %v", err)
		}

		reg, err := config.LoadRegistry()
		if err != nil {
			t.Fatal(err)
		}
		entry := reg.Find(targetDir)
		if entry == nil {
			t.Fatal("project not in registry")
		}
		if entry.RegisteredAt.Before(before) {
			t.Errorf("RegisteredAt %v should be after %v", entry.RegisteredAt, before)
		}
	})
}
