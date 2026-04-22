package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifest(t *testing.T) {
	tests := []struct {
		name    string
		content string
		setup   func(t *testing.T) string // returns path to manifest file
		wantErr bool
		check   func(t *testing.T, m *Manifest)
	}{
		{
			name: "valid pa.yaml parses correctly",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "pa.yaml")
				content := `version: 1
universal_rules_dir: .cursor/rules/universal
stacks:
  python:
    name: Python
    detect:
      - "*.py"
      - pyproject.toml
    rules_dir: stacks/python/rules
    skeleton_dir: stacks/python/skeleton
    claude_template: stacks/python/CLAUDE.md.template
  csharp:
    name: C#
    detect:
      - "*.csproj"
    rules_dir: stacks/csharp/rules
`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: false,
			check: func(t *testing.T, m *Manifest) {
				if m.Version != 1 {
					t.Errorf("Version = %d, want 1", m.Version)
				}
				if m.UniversalRulesDir != ".cursor/rules/universal" {
					t.Errorf("UniversalRulesDir = %q, want %q", m.UniversalRulesDir, ".cursor/rules/universal")
				}
				if len(m.Stacks) != 2 {
					t.Errorf("len(Stacks) = %d, want 2", len(m.Stacks))
				}
				python, ok := m.Stacks["python"]
				if !ok {
					t.Fatal("python stack not found")
				}
				if python.Name != "Python" {
					t.Errorf("python.Name = %q, want %q", python.Name, "Python")
				}
				if len(python.Detect) != 2 {
					t.Errorf("python.Detect len = %d, want 2", len(python.Detect))
				}
			},
		},
		{
			name: "invalid version returns error",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "pa.yaml")
				content := `version: 2
stacks:
  python:
    name: Python
`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: true,
		},
		{
			name: "version zero returns error",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "pa.yaml")
				content := `version: 0
stacks:
  python:
    name: Python
`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: true,
		},
		{
			name: "empty stacks returns error",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "pa.yaml")
				content := `version: 1
stacks: {}
`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: true,
		},
		{
			name: "no stacks key returns error",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "pa.yaml")
				content := `version: 1
`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: true,
		},
		{
			name: "new fields parse correctly when present",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "pa.yaml")
				content := `version: 1
universal_rules_dir: .cursor/rules
universal_claude_rules_dir: .claude/rules
shared_skeleton_dir: shared
stacks:
  python:
    name: Python
    detect:
      - "*.py"
`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: false,
			check: func(t *testing.T, m *Manifest) {
				if m.UniversalClaudeRulesDir != ".claude/rules" {
					t.Errorf("UniversalClaudeRulesDir = %q, want %q", m.UniversalClaudeRulesDir, ".claude/rules")
				}
				if m.SharedSkeletonDir != "shared" {
					t.Errorf("SharedSkeletonDir = %q, want %q", m.SharedSkeletonDir, "shared")
				}
				if m.UniversalRulesDir != ".cursor/rules" {
					t.Errorf("UniversalRulesDir = %q, want %q", m.UniversalRulesDir, ".cursor/rules")
				}
			},
		},
		{
			name: "new fields default to empty when absent",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "pa.yaml")
				content := `version: 1
universal_rules_dir: .cursor/rules
stacks:
  python:
    name: Python
`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: false,
			check: func(t *testing.T, m *Manifest) {
				if m.UniversalClaudeRulesDir != "" {
					t.Errorf("UniversalClaudeRulesDir = %q, want empty", m.UniversalClaudeRulesDir)
				}
				if m.SharedSkeletonDir != "" {
					t.Errorf("SharedSkeletonDir = %q, want empty", m.SharedSkeletonDir)
				}
				if m.UniversalRulesDir != ".cursor/rules" {
					t.Errorf("existing UniversalRulesDir = %q, want %q", m.UniversalRulesDir, ".cursor/rules")
				}
			},
		},
		{
			name: "non-existent file returns error",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does_not_exist.yaml")
			},
			wantErr: true,
		},
		{
			name: "invalid YAML returns error",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				path := filepath.Join(dir, "pa.yaml")
				content := `version: 1
stacks:
  python:
    - this is not valid
    detect: [bad
`
				if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
					t.Fatal(err)
				}
				return path
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)

			m, err := LoadManifest(path)

			if (err != nil) != tt.wantErr {
				t.Errorf("LoadManifest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, m)
			}
		})
	}
}
