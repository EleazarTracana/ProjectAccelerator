package tui

import "github.com/charmbracelet/lipgloss"

// Color palette.
var (
	colorPurple = lipgloss.Color("#7C3AED")
	colorGray   = lipgloss.Color("#6B7280")
	colorGreen  = lipgloss.Color("#10B981")
	colorYellow = lipgloss.Color("#F59E0B")
	colorRed    = lipgloss.Color("#EF4444")
	colorTeal   = lipgloss.Color("#14B8A6")
	colorWhite  = lipgloss.Color("#FFFFFF")
	colorDim    = lipgloss.Color("#9CA3AF")
)

// Text styles.
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPurple).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(colorGray).
			Italic(true)

	successStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	warningStyle = lipgloss.NewStyle().
			Foreground(colorYellow)

	errorStyle = lipgloss.NewStyle().
			Foreground(colorRed)

	accentStyle = lipgloss.NewStyle().
			Foreground(colorTeal)

	dimStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)

// List item styles.
var (
	selectedItemStyle = lipgloss.NewStyle().
				Background(colorTeal).
				Foreground(colorWhite).
				PaddingLeft(2).
				PaddingRight(2)

	normalItemStyle = lipgloss.NewStyle().
			PaddingLeft(2).
			PaddingRight(2)
)

// Layout styles.
var (
	appStyle = lipgloss.NewStyle().
			Padding(1, 2)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPurple).
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorPurple).
			MarginBottom(1).
			PaddingBottom(1)

	sectionStyle = lipgloss.NewStyle().
			MarginTop(1).
			MarginBottom(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(colorDim).
			MarginTop(1)
)

// Checklist styles.
var (
	checkedStyle = lipgloss.NewStyle().
			Foreground(colorGreen)

	uncheckedStyle = lipgloss.NewStyle().
			Foreground(colorDim)
)
