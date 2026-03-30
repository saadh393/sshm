package tui

import (
	"github.com/saadh393/sshm/internal/config"
	tea "github.com/charmbracelet/bubbletea"
)

// RunPicker opens the interactive list TUI with a custom title and returns
// the selected connection (or quit). It does NOT connect — callers decide
// what to do with the result.
func RunPicker(conns []config.Connection, title string) Result {
	if len(conns) == 0 {
		return Result{Quit: true}
	}

	m := NewModel(conns, 80, 24, title)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return Result{Quit: true}
	}

	fm, ok := finalModel.(Model)
	if !ok || fm.quitting {
		return Result{Quit: true}
	}
	if fm.chosen != nil {
		return Result{Conn: fm.chosen}
	}
	return Result{Quit: true}
}
