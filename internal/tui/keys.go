package tui

import "github.com/charmbracelet/bubbles/key"

// keyMap defines the keybindings used across the TUI.
type keyMap struct {
	Quit       key.Binding
	Back       key.Binding
	Enter      key.Binding
	Up         key.Binding
	Down       key.Binding
	Toggle     key.Binding
	Tab        key.Binding
	AutoDetect key.Binding
}

var keys = keyMap{
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q/ctrl+c", "quit"),
	),
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Up: key.NewBinding(
		key.WithKeys("k", "up"),
		key.WithHelp("up/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("j", "down"),
		key.WithHelp("down/j", "down"),
	),
	Toggle: key.NewBinding(
		key.WithKeys(" "),
		key.WithHelp("space", "toggle"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "cycle"),
	),
	AutoDetect: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "auto-detect"),
	),
}
