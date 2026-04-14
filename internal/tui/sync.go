package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gentleman-programming/project-accelerator/internal/config"
	"github.com/gentleman-programming/project-accelerator/internal/core"
)

type syncStep int

const (
	syncStepPickProject syncStep = iota
	syncStepPreview
	syncStepExecuting
	syncStepResult
)

// syncModel drives the multi-step sync flow.
type syncModel struct {
	cfg *config.Config

	step   syncStep
	cursor int
	err    error

	// syncStepPickProject state
	registry *config.Registry

	// syncStepPreview — selected project.
	selectedEntry *config.ProjectEntry

	// syncStepExecuting state
	spinner spinner.Model

	// syncStepResult state
	result *core.SyncResult
}

func newSyncModel(cfg *config.Config) syncModel {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(colorTeal)

	return syncModel{
		cfg:     cfg,
		step:    syncStepPickProject,
		spinner: sp,
	}
}

func (m syncModel) Init() tea.Cmd {
	return m.loadRegistry()
}

func (m syncModel) loadRegistry() tea.Cmd {
	return func() tea.Msg {
		reg, err := config.LoadRegistry()
		if err != nil {
			return registryErrorMsg{err: err}
		}
		return registryLoadedMsg{registry: reg}
	}
}

func (m syncModel) Update(msg tea.Msg) (syncModel, tea.Cmd) {
	// Handle registry loading across all steps.
	switch msg := msg.(type) {
	case registryLoadedMsg:
		m.registry = msg.registry
		m.err = nil
		return m, nil
	case registryErrorMsg:
		m.err = msg.err
		return m, nil
	}

	switch m.step {
	case syncStepPickProject:
		return m.updatePickProject(msg)
	case syncStepPreview:
		return m.updatePreview(msg)
	case syncStepExecuting:
		return m.updateExecuting(msg)
	case syncStepResult:
		return m.updateResult(msg)
	}
	return m, nil
}

func (m syncModel) View() string {
	var b strings.Builder

	header := headerStyle.Render("Sync a Project")
	b.WriteString(header)
	b.WriteString("\n")

	switch m.step {
	case syncStepPickProject:
		b.WriteString(m.viewPickProject())
	case syncStepPreview:
		b.WriteString(m.viewPreview())
	case syncStepExecuting:
		b.WriteString(m.viewExecuting())
	case syncStepResult:
		b.WriteString(m.viewResult())
	}

	return b.String()
}

// ── Step 1: Pick Project ────────────────────────────────────────────────────

func (m syncModel) updatePickProject(msg tea.Msg) (syncModel, tea.Cmd) {
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
			if m.registry != nil && m.cursor < len(m.registry.Projects)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Enter):
			if m.registry == nil || len(m.registry.Projects) == 0 {
				return m, nil
			}
			m.selectedEntry = &m.registry.Projects[m.cursor]
			m.step = syncStepPreview
			m.cursor = 0
			return m, nil
		}
	}
	return m, nil
}

func (m syncModel) viewPickProject() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("Select a registered project to sync"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error loading registry: %s", m.err)))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc: back to menu"))
		return b.String()
	}

	if m.registry == nil {
		b.WriteString(dimStyle.Render("Loading registry..."))
		return b.String()
	}

	if len(m.registry.Projects) == 0 {
		b.WriteString(dimStyle.Render("No registered projects found."))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Use Scaffold first to register a project."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc: back to menu"))
		return b.String()
	}

	for i, p := range m.registry.Projects {
		label := fmt.Sprintf("%s  [%s]", p.Path, strings.Join(p.Stacks, ", "))
		if !p.LastSync.IsZero() {
			label += dimStyle.Render(fmt.Sprintf("  (synced: %s)", p.LastSync.Format("2006-01-02")))
		}
		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render("  " + label))
		} else {
			b.WriteString(normalItemStyle.Render("  " + label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("up/down: navigate  enter: select  esc: back"))

	return b.String()
}

// ── Step 2: Preview ─────────────────────────────────────────────────────────

func (m syncModel) updatePreview(msg tea.Msg) (syncModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back):
			m.step = syncStepPickProject
			return m, nil
		case key.Matches(msg, keys.Enter):
			m.step = syncStepExecuting
			return m, tea.Batch(m.spinner.Tick, m.executeSync())
		}
	}
	return m, nil
}

func (m syncModel) viewPreview() string {
	var b strings.Builder

	b.WriteString(subtitleStyle.Render("Sync preview"))
	b.WriteString("\n\n")

	e := m.selectedEntry
	b.WriteString(fmt.Sprintf("  %s %s\n", accentStyle.Render("Project:"), e.Path))
	b.WriteString(fmt.Sprintf("  %s %s\n", accentStyle.Render("Stacks:"), strings.Join(e.Stacks, ", ")))

	checksumCount := len(e.Checksums)
	b.WriteString(fmt.Sprintf("  %s %d\n", accentStyle.Render("Tracked files:"), checksumCount))

	if !e.LastSync.IsZero() {
		b.WriteString(fmt.Sprintf("  %s %s\n", accentStyle.Render("Last sync:"), e.LastSync.Format("2006-01-02 15:04")))
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("This will sync cursor rules from ProjectAccelerator to the target project."))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("Local overrides will be preserved."))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("enter: run sync  esc: back"))

	return b.String()
}

// ── Step 3: Executing ───────────────────────────────────────────────────────

func (m syncModel) executeSync() tea.Cmd {
	return func() tea.Msg {
		result, err := core.Sync(
			m.cfg.PAHome,
			m.selectedEntry,
			m.cfg.Manifest,
		)
		if err != nil {
			return syncErrorMsg{err: err}
		}
		return syncDoneMsg{result: result}
	}
}

func (m syncModel) updateExecuting(msg tea.Msg) (syncModel, tea.Cmd) {
	switch msg := msg.(type) {
	case syncDoneMsg:
		m.result = msg.result
		m.step = syncStepResult
		m.err = nil
		return m, nil
	case syncErrorMsg:
		m.err = msg.err
		m.step = syncStepResult
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m syncModel) viewExecuting() string {
	return sectionStyle.Render(
		m.spinner.View() + accentStyle.Render(" Syncing project..."),
	)
}

// ── Step 4: Result ──────────────────────────────────────────────────────────

func (m syncModel) updateResult(msg tea.Msg) (syncModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Enter), key.Matches(msg, keys.Back):
			return m, func() tea.Msg { return backToMenuMsg{} }
		}
	}
	return m, nil
}

func (m syncModel) viewResult() string {
	var b strings.Builder

	if m.err != nil {
		b.WriteString(errorStyle.Render("Sync failed"))
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render(m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("enter/esc: back to menu"))
		return b.String()
	}

	b.WriteString(successStyle.Render("Sync complete!"))
	b.WriteString("\n\n")

	r := m.result
	b.WriteString(fmt.Sprintf("  %s %d\n", accentStyle.Render("Updated:"), r.Updated))
	b.WriteString(fmt.Sprintf("  %s %d\n", accentStyle.Render("Added:"), r.Added))
	b.WriteString(fmt.Sprintf("  %s %d\n", accentStyle.Render("Skipped:"), r.Skipped))
	b.WriteString(fmt.Sprintf("  %s %d\n", accentStyle.Render("Overrides preserved:"), r.Overrides))

	if len(r.Actions) > 0 {
		b.WriteString("\n")
		b.WriteString(accentStyle.Render("Details:"))
		b.WriteString("\n")
		for _, a := range r.Actions {
			var icon string
			switch a.Action {
			case "added":
				icon = successStyle.Render("+")
			case "updated":
				icon = accentStyle.Render("~")
			case "override_preserved":
				icon = warningStyle.Render("!")
			case "skipped":
				icon = dimStyle.Render("-")
			default:
				icon = " "
			}
			b.WriteString(fmt.Sprintf("  %s %s  %s\n", icon, a.RelPath, dimStyle.Render(a.Reason)))
		}
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
