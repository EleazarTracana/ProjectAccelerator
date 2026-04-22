package tui

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gentleman-programming/project-accelerator/internal/config"
	"github.com/gentleman-programming/project-accelerator/internal/core"
)

type scaffoldStep int

const (
	stepPickStack scaffoldStep = iota
	stepEnterPath
	stepOptions
	stepExecuting
	stepResult
)

// checklistItem represents a toggleable option in the scaffold checklist.
type checklistItem struct {
	label   string
	checked bool
}

// stackChecklistItem represents a toggleable stack entry in step 1.
type stackChecklistItem struct {
	key     string
	cfg     config.StackConfig
	checked bool
}

// scaffoldModel drives the multi-step scaffold wizard.
type scaffoldModel struct {
	cfg *config.Config

	step   scaffoldStep
	cursor int
	err    error

	// stepPickStack state — checklist of available stacks.
	stackItems    []stackChecklistItem // one entry per stack in manifest
	autoDetectMsg string               // non-empty after auto-detect ran

	// stepEnterPath state
	pathInput textinput.Model

	// stepOptions state
	options []checklistItem

	// stepExecuting state
	spinner spinner.Model

	// stepResult state
	result *core.ScaffoldResult

	// selected values (built just before execution)
	selectedStacks map[string]config.StackConfig
}

func newScaffoldModel(cfg *config.Config) scaffoldModel {
	// Build sorted stack key list.
	stackKeys := make([]string, 0, len(cfg.Manifest.Stacks))
	for k := range cfg.Manifest.Stacks {
		stackKeys = append(stackKeys, k)
	}
	sort.Strings(stackKeys)

	stackItems := make([]stackChecklistItem, 0, len(stackKeys))
	for _, k := range stackKeys {
		stackItems = append(stackItems, stackChecklistItem{
			key:     k,
			cfg:     cfg.Manifest.Stacks[k],
			checked: false,
		})
	}

	// Text input for path.
	ti := textinput.New()
	ti.Placeholder = "/path/to/your/project"
	ti.Prompt = accentStyle.Render("> ")
	cwd, err := os.Getwd()
	if err == nil {
		ti.SetValue(cwd)
	}
	ti.CharLimit = 256

	// Spinner for executing step.
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(colorTeal)

	// Default options — all checked.
	opts := []checklistItem{
		{label: "Initialize git repository", checked: true},
		{label: "Copy rules (Cursor + Claude)", checked: true},
		{label: "Generate CLAUDE.md from template", checked: true},
		{label: "Copy skeleton structure", checked: true},
		{label: "Register for future sync", checked: true},
	}

	return scaffoldModel{
		cfg:        cfg,
		step:       stepPickStack,
		stackItems: stackItems,
		pathInput:  ti,
		options:    opts,
		spinner:    sp,
	}
}

func (m scaffoldModel) Init() tea.Cmd {
	return nil
}

func (m scaffoldModel) Update(msg tea.Msg) (scaffoldModel, tea.Cmd) {
	switch m.step {
	case stepPickStack:
		return m.updatePickStack(msg)
	case stepEnterPath:
		return m.updateEnterPath(msg)
	case stepOptions:
		return m.updateOptions(msg)
	case stepExecuting:
		return m.updateExecuting(msg)
	case stepResult:
		return m.updateResult(msg)
	}
	return m, nil
}

func (m scaffoldModel) View() string {
	var b strings.Builder

	header := headerStyle.Render("Scaffold a New Project")
	b.WriteString(header)
	b.WriteString("\n")

	switch m.step {
	case stepPickStack:
		b.WriteString(m.viewPickStack())
	case stepEnterPath:
		b.WriteString(m.viewEnterPath())
	case stepOptions:
		b.WriteString(m.viewOptions())
	case stepExecuting:
		b.WriteString(m.viewExecuting())
	case stepResult:
		b.WriteString(m.viewResult())
	}

	return b.String()
}

// ── Step 1: Pick Stacks (checklist) ─────────────────────────────────────────

// checkedStackCount returns how many stacks are currently toggled on.
func (m scaffoldModel) checkedStackCount() int {
	n := 0
	for _, item := range m.stackItems {
		if item.checked {
			n++
		}
	}
	return n
}

func (m scaffoldModel) updatePickStack(msg tea.Msg) (scaffoldModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			return m, func() tea.Msg { return backToMenuMsg{} }

		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}

		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.stackItems)-1 {
				m.cursor++
			}

		case key.Matches(msg, keys.Toggle):
			// Toggle the stack under the cursor.
			m.stackItems[m.cursor].checked = !m.stackItems[m.cursor].checked
			m.err = nil

		case key.Matches(msg, keys.AutoDetect):
			// Run auto-detect immediately and toggle on all detected stacks.
			return m, m.runAutoDetect()

		case key.Matches(msg, keys.Enter):
			if m.checkedStackCount() == 0 {
				m.err = fmt.Errorf("select at least one stack (space: toggle, a: auto-detect)")
				return m, nil
			}
			m.err = nil
			m.step = stepEnterPath
			m.cursor = 0
			return m, m.pathInput.Focus()
		}
	case autoDetectDoneMsg:
		// Mark detected stacks as checked; leave others as-is.
		detected := make(map[string]bool, len(msg.detected))
		for _, k := range msg.detected {
			detected[k] = true
		}
		for i := range m.stackItems {
			if detected[m.stackItems[i].key] {
				m.stackItems[i].checked = true
			}
		}
		// Build human-readable message.
		if len(msg.detected) > 0 {
			m.autoDetectMsg = "Auto-detected: " + strings.Join(msg.detected, ", ")
		} else {
			m.autoDetectMsg = "No stacks detected — select manually"
		}
		m.err = nil
		return m, nil
	case autoDetectErrMsg:
		m.err = msg.err
		return m, nil
	}
	return m, nil
}

// autoDetectDoneMsg carries the list of detected stack keys.
type autoDetectDoneMsg struct {
	detected []string
}

// autoDetectErrMsg carries an auto-detect error.
type autoDetectErrMsg struct {
	err error
}

// runAutoDetect uses the current path input value (or cwd as fallback) to
// detect which stacks apply, then returns an autoDetectDoneMsg.
func (m scaffoldModel) runAutoDetect() tea.Cmd {
	return func() tea.Msg {
		targetDir := m.pathInput.Value()
		if strings.TrimSpace(targetDir) == "" {
			cwd, err := os.Getwd()
			if err != nil {
				return autoDetectErrMsg{err: fmt.Errorf("auto-detect: cannot resolve working directory: %w", err)}
			}
			targetDir = cwd
		}
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			return autoDetectErrMsg{err: fmt.Errorf("directory %q does not exist; create it or select stacks manually", targetDir)}
		}
		results, err := core.DetectStack(targetDir, m.cfg.Manifest)
		if err != nil {
			return autoDetectErrMsg{err: fmt.Errorf("auto-detect failed: %w", err)}
		}
		keys := make([]string, 0, len(results))
		for _, r := range results {
			keys = append(keys, r.Stack)
		}
		return autoDetectDoneMsg{detected: keys}
	}
}

func (m scaffoldModel) viewPickStack() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("Step 1/3 — Choose stacks (space: toggle, a: auto-detect)"))
	b.WriteString("\n\n")

	// Auto-detect status message.
	if m.autoDetectMsg != "" {
		b.WriteString(accentStyle.Render(m.autoDetectMsg))
		b.WriteString("\n\n")
	}

	for i, item := range m.stackItems {
		checkbox := uncheckedStyle.Render("[ ]")
		if item.checked {
			checkbox = checkedStyle.Render("[x]")
		}
		label := fmt.Sprintf(" %s (%s)", item.cfg.Name, item.key)
		line := checkbox + label
		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render(line))
		} else {
			b.WriteString(normalItemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("up/down: navigate  space: toggle  a: auto-detect  enter: continue  esc: back"))

	return b.String()
}

// ── Step 2: Enter Path ──────────────────────────────────────────────────────

func (m scaffoldModel) updateEnterPath(msg tea.Msg) (scaffoldModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.step = stepPickStack
			m.pathInput.Blur()
			return m, nil
		case key.Matches(msg, keys.Enter):
			path := m.pathInput.Value()
			if strings.TrimSpace(path) == "" {
				m.err = fmt.Errorf("path cannot be empty")
				return m, nil
			}
			m.err = nil
			m.step = stepOptions
			m.cursor = 0
			m.pathInput.Blur()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.pathInput, cmd = m.pathInput.Update(msg)
	return m, cmd
}

func (m scaffoldModel) viewEnterPath() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("Step 2/3 — Target directory"))
	b.WriteString("\n\n")

	b.WriteString(m.pathInput.View())
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter: confirm  esc: back"))

	return b.String()
}

// ── Step 3: Options Checklist ───────────────────────────────────────────────

func (m scaffoldModel) updateOptions(msg tea.Msg) (scaffoldModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.step = stepEnterPath
			return m, m.pathInput.Focus()
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Toggle):
			m.options[m.cursor].checked = !m.options[m.cursor].checked
		case key.Matches(msg, keys.Enter):
			m.step = stepExecuting
			return m, tea.Batch(m.spinner.Tick, m.executeScaffold())
		}
	}
	return m, nil
}

func (m scaffoldModel) viewOptions() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("Step 3/3 — Options"))
	b.WriteString("\n\n")

	// Summarise selected stacks.
	stackLabels := make([]string, 0)
	for _, item := range m.stackItems {
		if item.checked {
			stackLabels = append(stackLabels, item.key)
		}
	}
	stackSummary := strings.Join(stackLabels, ", ")
	b.WriteString(dimStyle.Render(fmt.Sprintf("Stacks: %s  Path: %s", stackSummary, m.pathInput.Value())))
	b.WriteString("\n\n")

	for i, opt := range m.options {
		checkbox := uncheckedStyle.Render("[ ]")
		if opt.checked {
			checkbox = checkedStyle.Render("[x]")
		}

		label := fmt.Sprintf(" %s", opt.label)
		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render(checkbox + label))
		} else {
			b.WriteString(normalItemStyle.Render(checkbox + label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("up/down: navigate  space: toggle  enter: scaffold  esc: back"))

	return b.String()
}

// ── Step 4: Executing ───────────────────────────────────────────────────────

func (m scaffoldModel) executeScaffold() tea.Cmd {
	return func() tea.Msg {
		targetDir := m.pathInput.Value()

		// Build selectedStacks from checked items.
		selectedStacks := make(map[string]config.StackConfig)
		for _, item := range m.stackItems {
			if item.checked {
				selectedStacks[item.key] = item.cfg
			}
		}

		// Validate — should not be reachable, but guard anyway.
		if len(selectedStacks) == 0 {
			return scaffoldErrorMsg{err: fmt.Errorf("no stacks selected")}
		}

		opts := core.ScaffoldOptions{
			GitInit:      m.options[0].checked,
			CopyRules:    m.options[1].checked,
			ClaudeMD:     m.options[2].checked,
			CopySkeleton: m.options[3].checked,
			Register:     m.options[4].checked,
		}

		result, err := core.Scaffold(
			core.ScaffoldInput{
				PAHome:                  m.cfg.PAHome,
				TargetDir:               targetDir,
				Stacks:                  selectedStacks,
				UniversalRulesDir:       m.cfg.Manifest.UniversalRulesDir,
				UniversalClaudeRulesDir: m.cfg.Manifest.UniversalClaudeRulesDir,
				SharedSkeletonDir:       m.cfg.Manifest.SharedSkeletonDir,
			},
			opts,
		)
		if err != nil {
			return scaffoldErrorMsg{err: err}
		}

		return scaffoldDoneMsg{result: result}
	}
}

func (m scaffoldModel) updateExecuting(msg tea.Msg) (scaffoldModel, tea.Cmd) {
	switch msg := msg.(type) {
	case scaffoldDoneMsg:
		m.result = msg.result
		m.step = stepResult
		m.err = nil
		return m, nil
	case scaffoldErrorMsg:
		m.err = msg.err
		m.step = stepResult
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m scaffoldModel) viewExecuting() string {
	return sectionStyle.Render(
		m.spinner.View() + accentStyle.Render(" Scaffolding project..."),
	)
}

// ── Step 5: Result ──────────────────────────────────────────────────────────

func (m scaffoldModel) updateResult(msg tea.Msg) (scaffoldModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Enter), key.Matches(msg, keys.Back):
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
	}
	return m, nil
}

func (m scaffoldModel) viewResult() string {
	var b strings.Builder

	if m.err != nil {
		b.WriteString(errorStyle.Render("Scaffold failed"))
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("enter/esc: back to menu"))
		return b.String()
	}

	b.WriteString(successStyle.Render("Scaffold complete!"))
	b.WriteString("\n\n")

	r := m.result
	b.WriteString(fmt.Sprintf("  %s %d\n", accentStyle.Render("Cursor rules copied:"), r.RulesCopied))
	if r.ClaudeRulesCopied > 0 {
		b.WriteString(fmt.Sprintf("  %s %d\n", accentStyle.Render("Claude rules copied:"), r.ClaudeRulesCopied))
	}
	b.WriteString(fmt.Sprintf("  %s %d\n", accentStyle.Render("Skeleton files:"), r.SkeletonCopied))

	if r.ClaudeMDCreated {
		b.WriteString(fmt.Sprintf("  %s %s\n", accentStyle.Render("CLAUDE.md:"), successStyle.Render("created")))
	}
	if r.GitInitDone {
		b.WriteString(fmt.Sprintf("  %s %s\n", accentStyle.Render("Git init:"), successStyle.Render("done")))
	}
	if r.Registered {
		b.WriteString(fmt.Sprintf("  %s %s\n", accentStyle.Render("Registered:"), successStyle.Render("yes")))
	}

	if len(r.Errors) > 0 {
		b.WriteString("\n")
		b.WriteString(warningStyle.Render("Warnings:"))
		b.WriteString("\n")
		for _, e := range r.Errors {
			b.WriteString(fmt.Sprintf("  %s %s\n", warningStyle.Render("!"), e))
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("enter/esc: back to menu"))

	return b.String()
}
