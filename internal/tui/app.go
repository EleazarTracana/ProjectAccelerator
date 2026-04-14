package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gentleman-programming/project-accelerator/internal/config"
)

type screen int

const (
	screenMenu screen = iota
	screenScaffold
	screenSync
	screenList
)

type menuItem struct {
	icon  string
	label string
}

var menuItems = []menuItem{
	{icon: ">>", label: "Scaffold a project"},
	{icon: "<>", label: "Sync a project"},
	{icon: "[]", label: "List registered projects"},
	{icon: "--", label: "Quit"},
}

// appModel is the root Bubble Tea model for the PA TUI.
type appModel struct {
	screen   screen
	cursor   int
	loading  bool
	err      error
	quitting bool

	cfg *config.Config

	// Loading spinner.
	spinner spinner.Model

	// Sub-models.
	scaffold scaffoldModel
	sync     syncModel

	// List view.
	registry *config.Registry
}

// NewApp creates a new root app model.
func NewApp() appModel {
	sp := spinner.New(spinner.WithSpinner(spinner.Dot))
	sp.Style = lipgloss.NewStyle().Foreground(colorTeal)

	return appModel{
		screen:  screenMenu,
		loading: true,
		spinner: sp,
	}
}

func (m appModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, loadConfig)
}

func loadConfig() tea.Msg {
	cfg, err := config.Load("")
	if err != nil {
		return configErrorMsg{err: err}
	}
	return configLoadedMsg{cfg: cfg}
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Global quit handling — ctrl+c works from any screen.
	if msg, ok := msg.(tea.KeyMsg); ok {
		if msg.Type == tea.KeyCtrlC {
			m.quitting = true
			return m, tea.Quit
		}
		// 'q' only quits from the main menu (not text inputs).
		if key.Matches(msg, keys.Quit) && m.screen == screenMenu && !m.loading {
			m.quitting = true
			return m, tea.Quit
		}
	}

	// Handle config loading.
	switch msg := msg.(type) {
	case configLoadedMsg:
		m.cfg = msg.cfg
		m.loading = false
		m.err = nil
		return m, nil
	case configErrorMsg:
		m.err = msg.err
		m.loading = false
		return m, nil
	case backToMenuMsg:
		m.screen = screenMenu
		m.cursor = 0
		return m, nil
	}

	// Loading spinner.
	if m.loading {
		if msg, ok := msg.(spinner.TickMsg); ok {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	// Delegate to active screen.
	switch m.screen {
	case screenMenu:
		return m.updateMenu(msg)
	case screenScaffold:
		return m.updateScaffold(msg)
	case screenSync:
		return m.updateSync(msg)
	case screenList:
		return m.updateList(msg)
	}

	return m, nil
}

func (m appModel) View() string {
	if m.quitting {
		return ""
	}

	var content string

	if m.loading {
		content = m.viewLoading()
	} else if m.err != nil && m.screen == screenMenu {
		content = m.viewError()
	} else {
		switch m.screen {
		case screenMenu:
			content = m.viewMenu()
		case screenScaffold:
			content = m.scaffold.View()
		case screenSync:
			content = m.sync.View()
		case screenList:
			content = m.viewList()
		}
	}

	return appStyle.Render(content)
}

// ── Loading ─────────────────────────────────────────────────────────────────

func (m appModel) viewLoading() string {
	return m.spinner.View() + accentStyle.Render(" Loading configuration...")
}

func (m appModel) viewError() string {
	var b strings.Builder
	b.WriteString(errorStyle.Render("Failed to load configuration"))
	b.WriteString("\n\n")
	b.WriteString(errorStyle.Render(m.err.Error()))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("q: quit"))
	return b.String()
}

// ── Main Menu ───────────────────────────────────────────────────────────────

func (m appModel) updateMenu(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, keys.Down):
			if m.cursor < len(menuItems)-1 {
				m.cursor++
			}
		case key.Matches(msg, keys.Enter):
			return m.selectMenuItem()
		}
	}
	return m, nil
}

func (m appModel) selectMenuItem() (tea.Model, tea.Cmd) {
	switch m.cursor {
	case 0: // Scaffold
		m.screen = screenScaffold
		m.scaffold = newScaffoldModel(m.cfg)
		return m, m.scaffold.Init()
	case 1: // Sync
		m.screen = screenSync
		m.sync = newSyncModel(m.cfg)
		return m, m.sync.Init()
	case 2: // List
		m.screen = screenList
		return m, m.loadRegistryForList()
	case 3: // Quit
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m appModel) viewMenu() string {
	var b strings.Builder

	title := titleStyle.Render("ProjectAccelerator")
	b.WriteString(title)
	b.WriteString("\n")
	b.WriteString(subtitleStyle.Render("Bootstrap and maintain project standards"))
	b.WriteString("\n\n")

	for i, item := range menuItems {
		label := fmt.Sprintf("%s  %s", item.icon, item.label)
		if i == m.cursor {
			b.WriteString(selectedItemStyle.Render(label))
		} else {
			b.WriteString(normalItemStyle.Render(label))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("up/down: navigate  enter: select  q: quit"))

	return b.String()
}

// ── Scaffold Delegation ─────────────────────────────────────────────────────

func (m appModel) updateScaffold(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for backToMenuMsg first.
	if _, ok := msg.(backToMenuMsg); ok {
		m.screen = screenMenu
		m.cursor = 0
		return m, nil
	}

	var cmd tea.Cmd
	m.scaffold, cmd = m.scaffold.Update(msg)
	return m, cmd
}

// ── Sync Delegation ─────────────────────────────────────────────────────────

func (m appModel) updateSync(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Check for backToMenuMsg first.
	if _, ok := msg.(backToMenuMsg); ok {
		m.screen = screenMenu
		m.cursor = 0
		return m, nil
	}

	var cmd tea.Cmd
	m.sync, cmd = m.sync.Update(msg)
	return m, cmd
}

// ── List View ───────────────────────────────────────────────────────────────

func (m appModel) loadRegistryForList() tea.Cmd {
	return func() tea.Msg {
		reg, err := config.LoadRegistry()
		if err != nil {
			return registryErrorMsg{err: err}
		}
		return registryLoadedMsg{registry: reg}
	}
}

func (m appModel) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case registryLoadedMsg:
		m.registry = msg.registry
		m.err = nil
		return m, nil
	case registryErrorMsg:
		m.err = msg.err
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Back), key.Matches(msg, keys.Quit):
			m.screen = screenMenu
			m.cursor = 0
			m.registry = nil
			m.err = nil
			return m, nil
		}
	}
	return m, nil
}

func (m appModel) viewList() string {
	var b strings.Builder

	header := headerStyle.Render("Registered Projects")
	b.WriteString(header)
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("Error loading registry: %s", m.err)))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc/q: back to menu"))
		return b.String()
	}

	if m.registry == nil {
		b.WriteString(dimStyle.Render("Loading..."))
		return b.String()
	}

	if len(m.registry.Projects) == 0 {
		b.WriteString(dimStyle.Render("No registered projects."))
		b.WriteString("\n")
		b.WriteString(dimStyle.Render("Use Scaffold to register a project."))
		b.WriteString("\n\n")
		b.WriteString(helpStyle.Render("esc/q: back to menu"))
		return b.String()
	}

	for i, p := range m.registry.Projects {
		num := dimStyle.Render(fmt.Sprintf("%d.", i+1))
		stack := accentStyle.Render(fmt.Sprintf("[%s]", strings.Join(p.Stacks, ", ")))
		path := p.Path

		var syncInfo string
		if !p.LastSync.IsZero() {
			syncInfo = dimStyle.Render(fmt.Sprintf("  synced: %s", p.LastSync.Format("2006-01-02")))
		}

		b.WriteString(fmt.Sprintf("  %s %s %s%s\n", num, stack, path, syncInfo))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("esc/q: back to menu"))

	return b.String()
}
