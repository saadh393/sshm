package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/saadh393/sshm/internal/config"
)

type CommandFormResult struct {
	Name    string
	Command string
	Saved   bool
}

type commandFormModel struct {
	conn    config.Connection
	name    textinput.Model
	command textinput.Model
	focused int
	saved   bool
	quit    bool
}

func newCommandFormModel(conn config.Connection) commandFormModel {
	return newCommandFormModelWithDefaults(conn, "", "")
}

func newCommandFormModelWithDefaults(conn config.Connection, defaultName, defaultCommand string) commandFormModel {
	defaultName = strings.TrimSpace(defaultName)
	defaultCommand = strings.TrimSpace(defaultCommand)

	name := textinput.New()
	name.Placeholder = "restart-nginx"
	name.CharLimit = 100
	name.SetValue(defaultName)

	command := textinput.New()
	command.Placeholder = "sudo systemctl restart nginx"
	command.CharLimit = 500
	command.SetValue(defaultCommand)

	if strings.TrimSpace(defaultName) == "" {
		name.Focus()
	} else {
		command.Focus()
	}

	return commandFormModel{
		conn:    conn,
		name:    name,
		command: command,
		focused: func() int {
			if strings.TrimSpace(defaultName) == "" {
				return 0
			}
			return 1
		}(),
	}
}

func (m commandFormModel) Init() tea.Cmd { return textinput.Blink }

func (m commandFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		case "tab", "down", "enter":
			if m.focused == 0 {
				m.focused = 1
				m.name.Blur()
				m.command.Focus()
				return m, textinput.Blink
			}
			if m.canSave() {
				m.saved = true
				return m, tea.Quit
			}
		case "shift+tab", "up":
			if m.focused == 1 {
				m.focused = 0
				m.command.Blur()
				m.name.Focus()
				return m, textinput.Blink
			}
		case "ctrl+s":
			if m.canSave() {
				m.saved = true
				return m, tea.Quit
			}
		}
	}

	var cmd tea.Cmd
	if m.focused == 0 {
		m.name, cmd = m.name.Update(msg)
	} else {
		m.command, cmd = m.command.Update(msg)
	}
	return m, cmd
}

func (m commandFormModel) canSave() bool {
	return strings.TrimSpace(m.name.Value()) != "" && strings.TrimSpace(m.command.Value()) != ""
}

func (m commandFormModel) View() string {
	if m.quit || m.saved {
		return ""
	}

	label := func(active bool, text string) string {
		if active {
			return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#06B6D4")).Width(18).Render(text)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Width(18).Render(text)
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#7C3AED")).
		Padding(0, 1).
		MarginBottom(1).
		Render(fmt.Sprintf("Command — %s", m.conn.Alias))

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		MarginTop(1).
		Render("tab/↑↓ navigate  ctrl+s save  enter next/save  esc cancel")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Margin(1, 2)

	var sb strings.Builder
	sb.WriteString(title)
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s  %s\n", label(m.focused == 0, "Command name"), m.name.View()))
	sb.WriteString(fmt.Sprintf("%s  %s\n", label(m.focused == 1, "Remote command"), m.command.View()))
	sb.WriteString(help)
	return box.Render(sb.String())
}

func (m commandFormModel) result() CommandFormResult {
	if !m.saved {
		return CommandFormResult{Saved: false}
	}
	return CommandFormResult{
		Name:    strings.TrimSpace(m.name.Value()),
		Command: strings.TrimSpace(m.command.Value()),
		Saved:   true,
	}
}

func RunCommandForm(conn config.Connection) CommandFormResult {
	m := newCommandFormModel(conn)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return CommandFormResult{Saved: false}
	}
	fm, ok := finalModel.(commandFormModel)
	if !ok {
		return CommandFormResult{Saved: false}
	}
	return fm.result()
}

func RunUpdateCommandForm(conn config.Connection, name, command string) CommandFormResult {
	m := newCommandFormModelWithDefaults(conn, name, command)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return CommandFormResult{Saved: false}
	}
	fm, ok := finalModel.(commandFormModel)
	if !ok {
		return CommandFormResult{Saved: false}
	}
	return fm.result()
}
