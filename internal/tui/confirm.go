package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ConfirmResult struct {
	Confirmed bool
}

type confirmModel struct {
	title     string
	message   string
	confirmed bool
	quit      bool
}

func newConfirmModel(title, message string) confirmModel {
	return confirmModel{title: title, message: message}
}

func (m confirmModel) Init() tea.Cmd { return nil }

func (m confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y", "enter":
			m.confirmed = true
			return m, tea.Quit
		case "n", "N", "q", "esc", "ctrl+c":
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m confirmModel) View() string {
	if m.confirmed || m.quit {
		return ""
	}
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#EF4444")).
		Padding(0, 1).
		MarginBottom(1).
		Render(m.title)

	msg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F3F4F6")).
		Render(m.message)

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		MarginTop(1).
		Render("Press y/Enter to confirm, n/Esc to cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#EF4444")).
		Padding(1, 2).
		Margin(1, 2)

	return box.Render(fmt.Sprintf("%s\n%s\n%s", title, msg, help))
}

func RunConfirm(title, message string) ConfirmResult {
	m := newConfirmModel(title, message)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return ConfirmResult{Confirmed: false}
	}
	fm, ok := finalModel.(confirmModel)
	if !ok {
		return ConfirmResult{Confirmed: false}
	}
	return ConfirmResult{Confirmed: fm.confirmed}
}
