package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colour palette
	colorPrimary  = lipgloss.Color("#7C3AED") // purple
	colorSecondary = lipgloss.Color("#06B6D4") // cyan
	colorMuted    = lipgloss.Color("#6B7280") // grey
	colorSuccess  = lipgloss.Color("#10B981") // green
	colorWarning  = lipgloss.Color("#F59E0B") // amber

	// Title bar
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(colorPrimary).
			Padding(0, 1)

	// Selected item highlight
	SelectedItemStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary)

	// Normal item
	NormalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5E7EB"))

	// Alias column
	AliasStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorSecondary).
			Width(20)

	// user@host column
	HostStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB")).
			Width(28)

	// Group column
	GroupStyle = lipgloss.NewStyle().
			Foreground(colorSuccess).
			Width(16)

	// Status / help bar
	HelpStyle = lipgloss.NewStyle().
			Foreground(colorMuted).
			Padding(0, 1)

	// Filter prompt
	FilterPromptStyle = lipgloss.NewStyle().
				Foreground(colorWarning).
				Bold(true)

	// Outer container
	AppStyle = lipgloss.NewStyle().
			Margin(1, 2)
)
