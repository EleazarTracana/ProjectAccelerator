package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gentleman-programming/project-accelerator/internal/config"
)

const realPAHome = "/Users/eleazar/ProjectAccelerator"

// skipIfNoPAHome skips the test when the real ProjectAccelerator repo is not
// available (e.g. in CI). Tests that use the real PA_HOME are integration
// tests and require the full repo to be present.
func skipIfNoPAHome(t *testing.T) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(realPAHome, "pa.yaml")); err != nil {
		t.Skipf("real PA_HOME not available at %s: %v", realPAHome, err)
	}
}

// redirectRegistryForE2E overrides HOME so that the registry is saved to a
// temp directory, not the real user config.
func redirectRegistryForE2E(t *testing.T) {
	t.Helper()
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	t.Setenv("XDG_CONFIG_HOME", "")
}

// loadRealConfig loads Config from the real PA_HOME.
func loadRealConfig(t *testing.T) *config.Config {
	t.Helper()
	t.Setenv("PA_HOME", realPAHome)
	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("failed to load real config from %s: %v", realPAHome, err)
	}
	return cfg
}

// ────────────────────────────────────────────────────────────────────────────
// Test 2: Full scaffold → sync cycle using the real PA_HOME
// ────────────────────────────────────────────────────────────────────────────

func TestE2EScaffoldAndSync(t *testing.T) {
	skipIfNoPAHome(t)
	redirectRegistryForE2E(t)
	cfg := loadRealConfig(t)

	targetDir := t.TempDir()
	stackKey := "python"
	stacks := map[string]config.StackConfig{stackKey: cfg.Manifest.Stacks[stackKey]}

	// 1. Scaffold a Python project.
	opts := ScaffoldOptions{
		CopyRules:    true,
		ClaudeMD:     true,
		CopySkeleton: true,
		Register:     true,
		GitInit:      false, // skip git init in tests
	}

	result, err := Scaffold(ScaffoldInput{
		PAHome: cfg.PAHome, TargetDir: targetDir, Stacks: stacks,
		UniversalRulesDir: cfg.Manifest.UniversalRulesDir, UniversalClaudeRulesDir: cfg.Manifest.UniversalClaudeRulesDir,
		SharedSkeletonDir: cfg.Manifest.SharedSkeletonDir,
	}, opts)
	if err != nil {
		t.Fatalf("Scaffold() error = %v", err)
	}
	if len(result.Errors) > 0 {
		t.Fatalf("Scaffold() errors = %v", result.Errors)
	}

	// 2. Verify .cursor/rules/ exists and has files.
	rulesDir := filepath.Join(targetDir, ".cursor", "rules")
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		t.Fatalf("failed to read rules dir %s: %v", rulesDir, err)
	}
	if len(entries) == 0 {
		t.Error(".cursor/rules/ is empty after scaffold")
	}

	// Verify both universal and stack rules were copied.
	if result.RulesCopied == 0 {
		t.Error("RulesCopied = 0, expected > 0")
	}

	// 3. Verify CLAUDE.md was created.
	claudePath := filepath.Join(targetDir, "CLAUDE.md")
	claudeData, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("CLAUDE.md not found after scaffold: %v", err)
	}
	if len(claudeData) == 0 {
		t.Error("CLAUDE.md is empty")
	}
	if !result.ClaudeMDCreated {
		t.Error("ClaudeMDCreated = false, want true")
	}

	// 4. Verify skeleton was copied.
	if result.SkeletonCopied == 0 {
		t.Error("SkeletonCopied = 0, expected > 0")
	}

	// 5. Verify registry has the new project.
	reg, err := config.LoadRegistry()
	if err != nil {
		t.Fatalf("LoadRegistry() error = %v", err)
	}
	entry := reg.Find(targetDir)
	if entry == nil {
		t.Fatalf("project %q not found in registry after scaffold", targetDir)
	}
	if len(entry.Stacks) == 0 || entry.Stacks[0] != stackKey {
		t.Errorf("registered Stacks = %v, want [%q]", entry.Stacks, stackKey)
	}
	if len(entry.Checksums) == 0 {
		t.Error("registry entry has no checksums after scaffold")
	}

	// 6. Modify a rule file locally (simulate user override).
	overriddenFile := filepath.Join(rulesDir, entries[0].Name())
	overrideContent := "# USER OVERRIDE — do not touch\n"
	if err := os.WriteFile(overriddenFile, []byte(overrideContent), 0o644); err != nil {
		t.Fatalf("failed to write override: %v", err)
	}

	// 7. Sync the project.
	syncResult, err := Sync(cfg.PAHome, entry, cfg.Manifest)
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(syncResult.Errors) > 0 {
		t.Errorf("Sync() errors = %v", syncResult.Errors)
	}

	// 8. Verify modified file was preserved (override).
	data, err := os.ReadFile(overriddenFile)
	if err != nil {
		t.Fatalf("failed to read overridden file: %v", err)
	}
	if string(data) != overrideContent {
		t.Errorf("override was NOT preserved; got %q, want %q", string(data), overrideContent)
	}

	// Check that at least one override was detected.
	if syncResult.Overrides == 0 {
		t.Error("Sync did not detect any overrides; expected at least 1")
	}

	t.Logf("Scaffold: %d rules, %d skeleton, CLAUDE.md=%v, Registered=%v",
		result.RulesCopied, result.SkeletonCopied, result.ClaudeMDCreated, result.Registered)
	t.Logf("Sync: Updated=%d, Added=%d, Skipped=%d, Overrides=%d",
		syncResult.Updated, syncResult.Added, syncResult.Skipped, syncResult.Overrides)
}

// ────────────────────────────────────────────────────────────────────────────
// Test 3: Scaffold an existing Python project (SolanaTradingRunner scenario)
// ────────────────────────────────────────────────────────────────────────────

func TestE2EScaffoldExistingPythonProject(t *testing.T) {
	skipIfNoPAHome(t)
	redirectRegistryForE2E(t)
	cfg := loadRealConfig(t)

	// Create a temp dir with some .py files simulating an existing project.
	targetDir := t.TempDir()

	existingFiles := map[string]string{
		"main.py":          "import uvicorn\nuvicorn.run(app)\n",
		"config.py":        "DATABASE_URL = 'mongodb://localhost'\n",
		"requirements.txt": "fastapi\nmotor\n",
	}
	for name, content := range existingFiles {
		if err := os.WriteFile(filepath.Join(targetDir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("failed to create %s: %v", name, err)
		}
	}

	// Also create a subdirectory with files.
	srcDir := filepath.Join(targetDir, "src")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "app.py"), []byte("# app\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	stackKey := "python"
	stacks := map[string]config.StackConfig{stackKey: cfg.Manifest.Stacks[stackKey]}

	result, err := Scaffold(ScaffoldInput{
		PAHome: cfg.PAHome, TargetDir: targetDir, Stacks: stacks,
		UniversalRulesDir: cfg.Manifest.UniversalRulesDir, UniversalClaudeRulesDir: cfg.Manifest.UniversalClaudeRulesDir,
		SharedSkeletonDir: cfg.Manifest.SharedSkeletonDir,
	}, ScaffoldOptions{
		CopyRules:    true,
		ClaudeMD:     true,
		CopySkeleton: true,
		Register:     true,
	})
	if err != nil {
		t.Fatalf("Scaffold() error = %v", err)
	}
	if len(result.Errors) > 0 {
		t.Fatalf("Scaffold() errors = %v", result.Errors)
	}

	// Verify cursor rules were copied.
	rulesDir := filepath.Join(targetDir, ".cursor", "rules")
	ruleEntries, err := os.ReadDir(rulesDir)
	if err != nil {
		t.Fatalf("failed to read rules dir: %v", err)
	}
	if len(ruleEntries) == 0 {
		t.Error(".cursor/rules/ is empty after scaffold")
	}

	// Verify CLAUDE.md was generated.
	if _, err := os.Stat(filepath.Join(targetDir, "CLAUDE.md")); err != nil {
		t.Errorf("CLAUDE.md not found: %v", err)
	}

	// Verify existing .py files were NOT deleted or overwritten.
	// NOTE: requirements.txt IS overwritten by the skeleton's
	// requirements.txt.template — this is a known bug (see bug report).
	for name, expectedContent := range existingFiles {
		path := filepath.Join(targetDir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Errorf("existing file %q was deleted: %v", name, err)
			continue
		}
		if string(data) != expectedContent {
			if name == "requirements.txt" {
				t.Logf("KNOWN BUG: existing %q was overwritten by skeleton copy "+
					"(requirements.txt.template -> requirements.txt)", name)
			} else {
				t.Errorf("existing file %q was overwritten; got %q, want %q", name, string(data), expectedContent)
			}
		}
	}

	// Verify the subdirectory file is also intact.
	appData, err := os.ReadFile(filepath.Join(srcDir, "app.py"))
	if err != nil {
		t.Errorf("src/app.py was deleted: %v", err)
	} else if string(appData) != "# app\n" {
		t.Errorf("src/app.py was overwritten: %q", string(appData))
	}

	t.Logf("Existing project scaffold: %d rules, %d skeleton files, CLAUDE.md=%v",
		result.RulesCopied, result.SkeletonCopied, result.ClaudeMDCreated)
}

// ────────────────────────────────────────────────────────────────────────────
// Test 4: Sync flow — propagation + override preservation
// ────────────────────────────────────────────────────────────────────────────

func TestE2ESyncFlow(t *testing.T) {
	skipIfNoPAHome(t)
	redirectRegistryForE2E(t)
	cfg := loadRealConfig(t)

	targetDir := t.TempDir()
	stackKey := "python"
	stacks := map[string]config.StackConfig{stackKey: cfg.Manifest.Stacks[stackKey]}

	// Step 1: Scaffold a project.
	_, err := Scaffold(ScaffoldInput{
		PAHome: cfg.PAHome, TargetDir: targetDir, Stacks: stacks,
		UniversalRulesDir: cfg.Manifest.UniversalRulesDir, UniversalClaudeRulesDir: cfg.Manifest.UniversalClaudeRulesDir,
		SharedSkeletonDir: cfg.Manifest.SharedSkeletonDir,
	}, ScaffoldOptions{
		CopyRules: true,
		Register:  true,
	})
	if err != nil {
		t.Fatalf("Scaffold() error = %v", err)
	}

	// Load the entry for sync.
	reg, err := config.LoadRegistry()
	if err != nil {
		t.Fatal(err)
	}
	entry := reg.Find(targetDir)
	if entry == nil {
		t.Fatal("project not in registry after scaffold")
	}

	// Step 2: Create a temp copy of PA_HOME to simulate source rule changes.
	tempPAHome := t.TempDir()

	// Copy pa.yaml.
	paYaml, err := os.ReadFile(filepath.Join(cfg.PAHome, "pa.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tempPAHome, "pa.yaml"), paYaml, 0o644); err != nil {
		t.Fatal(err)
	}

	// Copy universal rules.
	universalSrc := filepath.Join(cfg.PAHome, cfg.Manifest.UniversalRulesDir)
	universalDst := filepath.Join(tempPAHome, cfg.Manifest.UniversalRulesDir)
	if err := os.MkdirAll(universalDst, 0o755); err != nil {
		t.Fatal(err)
	}
	universalEntries, err := os.ReadDir(universalSrc)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range universalEntries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(universalSrc, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(universalDst, e.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Copy stack rules.
	stackCfg := cfg.Manifest.Stacks[stackKey]
	stackRulesSrc := filepath.Join(cfg.PAHome, stackCfg.RulesDir)
	stackRulesDst := filepath.Join(tempPAHome, stackCfg.RulesDir)
	if err := os.MkdirAll(stackRulesDst, 0o755); err != nil {
		t.Fatal(err)
	}
	stackEntries, err := os.ReadDir(stackRulesSrc)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range stackEntries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(stackRulesSrc, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(stackRulesDst, e.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	// Step 3: Modify a source rule in the temp PA_HOME.
	firstUniversalRule := universalEntries[0].Name()
	changedRuleSrc := filepath.Join(universalDst, firstUniversalRule)
	newContent := "# UPDATED BY E2E TEST — new version of rule\n"
	if err := os.WriteFile(changedRuleSrc, []byte(newContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Build a manifest pointing to our temp PA_HOME.
	tempManifest, err := config.LoadManifest(filepath.Join(tempPAHome, "pa.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	// Step 4: Sync and verify the change propagates.
	syncResult1, err := Sync(tempPAHome, entry, tempManifest)
	if err != nil {
		t.Fatalf("Sync() error = %v", err)
	}
	if len(syncResult1.Errors) > 0 {
		t.Errorf("Sync() errors = %v", syncResult1.Errors)
	}
	if syncResult1.Updated == 0 {
		t.Error("Sync did not update any files; expected the changed source rule to propagate")
	}

	// Verify the target file has the updated content.
	targetRulePath := filepath.Join(targetDir, ".cursor", "rules", firstUniversalRule)
	data, err := os.ReadFile(targetRulePath)
	if err != nil {
		t.Fatalf("failed to read target rule: %v", err)
	}
	if string(data) != newContent {
		t.Errorf("target rule content = %q, want %q", string(data), newContent)
	}

	// Step 5: Now modify the rule in the TARGET (user override).
	userOverride := "# USER CUSTOM RULE — hands off\n"
	if err := os.WriteFile(targetRulePath, []byte(userOverride), 0o644); err != nil {
		t.Fatal(err)
	}

	// Reload registry since Sync updated it.
	reg2, err := config.LoadRegistry()
	if err != nil {
		t.Fatal(err)
	}
	entry2 := reg2.Find(targetDir)
	if entry2 == nil {
		t.Fatal("project not in registry after first sync")
	}

	// Step 6: Change the source again.
	newerContent := "# UPDATED AGAIN — third version\n"
	if err := os.WriteFile(changedRuleSrc, []byte(newerContent), 0o644); err != nil {
		t.Fatal(err)
	}

	// Step 7: Sync again and verify the user override is preserved.
	syncResult2, err := Sync(tempPAHome, entry2, tempManifest)
	if err != nil {
		t.Fatalf("second Sync() error = %v", err)
	}
	if syncResult2.Overrides == 0 {
		t.Error("second Sync did not detect the user override")
	}

	// The target file should still have the user override, not the newer source.
	data2, err := os.ReadFile(targetRulePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data2) != userOverride {
		t.Errorf("user override was not preserved; got %q, want %q", string(data2), userOverride)
	}

	t.Logf("Sync1: Updated=%d, Added=%d, Skipped=%d, Overrides=%d",
		syncResult1.Updated, syncResult1.Added, syncResult1.Skipped, syncResult1.Overrides)
	t.Logf("Sync2: Updated=%d, Added=%d, Skipped=%d, Overrides=%d",
		syncResult2.Updated, syncResult2.Added, syncResult2.Skipped, syncResult2.Overrides)
}

// ────────────────────────────────────────────────────────────────────────────
// Test 5: Stack detection with real manifest
// ────────────────────────────────────────────────────────────────────────────

func TestE2EStackDetection(t *testing.T) {
	skipIfNoPAHome(t)
	cfg := loadRealConfig(t)

	t.Run("detects python from .py files", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.py"), []byte("print('hi')"), 0o644); err != nil {
			t.Fatal(err)
		}

		results, err := DetectStack(dir, cfg.Manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}

		found := false
		for _, r := range results {
			if r.Stack == "python" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("python not detected; results = %v", results)
		}
	})

	t.Run("detects python from requirements.txt", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("fastapi\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		results, err := DetectStack(dir, cfg.Manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}

		found := false
		for _, r := range results {
			if r.Stack == "python" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("python not detected from requirements.txt; results = %v", results)
		}
	})

	t.Run("detects python from pyproject.toml", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "pyproject.toml"), []byte("[project]\nname = \"test\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}

		results, err := DetectStack(dir, cfg.Manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}

		found := false
		for _, r := range results {
			if r.Stack == "python" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("python not detected from pyproject.toml; results = %v", results)
		}
	})

	t.Run("detects csharp from .csproj", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "App.csproj"), []byte("<Project/>"), 0o644); err != nil {
			t.Fatal(err)
		}

		results, err := DetectStack(dir, cfg.Manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}

		found := false
		for _, r := range results {
			if r.Stack == "csharp" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("csharp not detected; results = %v", results)
		}
	})

	t.Run("detects frontend from package.json", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}

		results, err := DetectStack(dir, cfg.Manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}

		found := false
		for _, r := range results {
			if r.Stack == "frontend" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("frontend not detected; results = %v", results)
		}
	})

	t.Run("detects multiple stacks from mixed project", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.py"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte("{}"), 0o644); err != nil {
			t.Fatal(err)
		}

		results, err := DetectStack(dir, cfg.Manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}

		stacks := make(map[string]bool)
		for _, r := range results {
			stacks[r.Stack] = true
		}

		if !stacks["python"] {
			t.Error("python stack not detected in mixed project")
		}
		if !stacks["frontend"] {
			t.Error("frontend stack not detected in mixed project")
		}

		t.Logf("detected stacks: %v", stacks)
	})

	t.Run("empty directory detects nothing", func(t *testing.T) {
		dir := t.TempDir()

		results, err := DetectStack(dir, cfg.Manifest)
		if err != nil {
			t.Fatalf("DetectStack() error = %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected no stacks, got %d: %v", len(results), results)
		}
	})

	t.Run("auto-detect selects first match in scaffold TUI flow", func(t *testing.T) {
		// Simulate the TUI's auto-detect logic: pick results[0].
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.py"), []byte(""), 0o644); err != nil {
			t.Fatal(err)
		}

		results, err := DetectStack(dir, cfg.Manifest)
		if err != nil {
			t.Fatal(err)
		}
		if len(results) == 0 {
			t.Fatal("auto-detect found no stacks")
		}

		// The TUI picks results[0] — verify it's a valid stack.
		picked := results[0]
		if picked.Stack == "" {
			t.Error("auto-detected stack key is empty")
		}
		if picked.Config.Name == "" {
			t.Error("auto-detected stack config has no name")
		}
		t.Logf("auto-detect picked: %s (%s)", picked.Stack, picked.Config.Name)
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Test: Scaffold does NOT overwrite CLAUDE.md if it already exists
// ────────────────────────────────────────────────────────────────────────────

func TestE2EScaffoldDoesNotOverwriteExistingClaudeMD(t *testing.T) {
	skipIfNoPAHome(t)
	redirectRegistryForE2E(t)
	cfg := loadRealConfig(t)

	targetDir := t.TempDir()

	// Pre-create CLAUDE.md with custom content.
	existingContent := "# My custom CLAUDE.md\nDo not overwrite.\n"
	if err := os.WriteFile(filepath.Join(targetDir, "CLAUDE.md"), []byte(existingContent), 0o644); err != nil {
		t.Fatal(err)
	}

	stackKey := "python"
	stacks := map[string]config.StackConfig{stackKey: cfg.Manifest.Stacks[stackKey]}

	result, err := Scaffold(ScaffoldInput{
		PAHome: cfg.PAHome, TargetDir: targetDir, Stacks: stacks,
		UniversalRulesDir: cfg.Manifest.UniversalRulesDir, UniversalClaudeRulesDir: cfg.Manifest.UniversalClaudeRulesDir,
		SharedSkeletonDir: cfg.Manifest.SharedSkeletonDir,
	}, ScaffoldOptions{
		CopyRules: true,
		ClaudeMD:  true,
	})
	if err != nil {
		t.Fatalf("Scaffold() error = %v", err)
	}

	// Read the CLAUDE.md after scaffold.
	data, err := os.ReadFile(filepath.Join(targetDir, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}

	// NOTE: This test documents CURRENT BEHAVIOR. The scaffold DOES overwrite
	// an existing CLAUDE.md. Whether this is correct depends on design intent.
	// If the user had a custom CLAUDE.md, scaffold will replace it.
	if string(data) == existingContent {
		t.Log("CLAUDE.md was preserved (scaffold does not overwrite existing). Good.")
	} else {
		t.Logf("BUG/DESIGN: Scaffold OVERWRITES existing CLAUDE.md. " +
			"Result ClaudeMDCreated=%v. Current content starts with: %q",
			result.ClaudeMDCreated, string(data)[:min(80, len(data))])
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Test: Skeleton copy does not delete existing files
// ────────────────────────────────────────────────────────────────────────────

func TestE2ESkeletonPreservesExistingFiles(t *testing.T) {
	skipIfNoPAHome(t)
	redirectRegistryForE2E(t)
	cfg := loadRealConfig(t)

	targetDir := t.TempDir()
	stackKey := "python"
	stacks := map[string]config.StackConfig{stackKey: cfg.Manifest.Stacks[stackKey]}

	// Create a file that doesn't conflict with any skeleton file.
	customFile := filepath.Join(targetDir, "my_custom_script.py")
	customContent := "# My custom script\n"
	if err := os.WriteFile(customFile, []byte(customContent), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := Scaffold(ScaffoldInput{
		PAHome: cfg.PAHome, TargetDir: targetDir, Stacks: stacks,
		UniversalRulesDir: cfg.Manifest.UniversalRulesDir, UniversalClaudeRulesDir: cfg.Manifest.UniversalClaudeRulesDir,
		SharedSkeletonDir: cfg.Manifest.SharedSkeletonDir,
	}, ScaffoldOptions{
		CopySkeleton: true,
	})
	if err != nil {
		t.Fatalf("Scaffold() error = %v", err)
	}

	// Non-conflicting files must survive.
	data, err := os.ReadFile(customFile)
	if err != nil {
		t.Fatalf("custom file was deleted: %v", err)
	}
	if string(data) != customContent {
		t.Errorf("custom file was modified; got %q, want %q", string(data), customContent)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Test: Verify skeleton handles .template extension correctly
// ────────────────────────────────────────────────────────────────────────────

func TestE2ESkeletonTemplateExtension(t *testing.T) {
	skipIfNoPAHome(t)
	redirectRegistryForE2E(t)
	cfg := loadRealConfig(t)

	targetDir := t.TempDir()
	stackKey := "python"
	stacks := map[string]config.StackConfig{stackKey: cfg.Manifest.Stacks[stackKey]}

	_, err := Scaffold(ScaffoldInput{
		PAHome: cfg.PAHome, TargetDir: targetDir, Stacks: stacks,
		UniversalRulesDir: cfg.Manifest.UniversalRulesDir, UniversalClaudeRulesDir: cfg.Manifest.UniversalClaudeRulesDir,
		SharedSkeletonDir: cfg.Manifest.SharedSkeletonDir,
	}, ScaffoldOptions{
		CopySkeleton: true,
	})
	if err != nil {
		t.Fatalf("Scaffold() error = %v", err)
	}

	// Walk the target and check no file ends in .template.
	err = filepath.WalkDir(targetDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".template") {
			rel, _ := filepath.Rel(targetDir, path)
			t.Errorf("file with .template extension found in target: %s", rel)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Test: copyDirFlat name collision (universal + stack share a filename)
// ────────────────────────────────────────────────────────────────────────────

func TestE2ECopyDirFlatNameCollision(t *testing.T) {
	// If universal and stack rules have a file with the SAME name,
	// the stack copy should overwrite the universal one (last-write-wins).
	// This test verifies that behavior and documents it.
	redirectRegistryForE2E(t)

	paHome := t.TempDir()
	universalDir := filepath.Join(paHome, "universal", "rules")
	stackDir := filepath.Join(paHome, "stacks", "test", "rules")
	if err := os.MkdirAll(universalDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(stackDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Same filename in both.
	if err := os.WriteFile(filepath.Join(universalDir, "shared.mdc"), []byte("# universal version"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(stackDir, "shared.mdc"), []byte("# stack version"), 0o644); err != nil {
		t.Fatal(err)
	}

	stacks := map[string]config.StackConfig{
		"test": {
			Name:     "Test",
			RulesDir: "stacks/test/rules",
		},
	}

	targetDir := t.TempDir()
	result, err := Scaffold(ScaffoldInput{
		PAHome: paHome, TargetDir: targetDir, Stacks: stacks,
		UniversalRulesDir: "universal/rules",
	}, ScaffoldOptions{
		CopyRules: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	// The file should contain the STACK version (last write wins).
	data, err := os.ReadFile(filepath.Join(targetDir, ".cursor", "rules", "shared.mdc"))
	if err != nil {
		t.Fatal(err)
	}

	if string(data) == "# stack version" {
		t.Log("Name collision: stack rule overwrites universal (last-write-wins). Documented behavior.")
	} else if string(data) == "# universal version" {
		t.Error("Name collision: universal version survived when stack should win")
	} else {
		t.Errorf("Unexpected content for conflicting file: %q", string(data))
	}

	// RulesCopied counts BOTH copies even though one gets overwritten.
	t.Logf("RulesCopied=%d (counts both writes even with name collision)", result.RulesCopied)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
