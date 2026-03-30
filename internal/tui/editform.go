package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/saadh393/sshm/internal/config"
)

// EditResult is returned after the edit form TUI exits.
type EditResult struct {
	Conn  *config.Connection
	Saved bool
}

const (
	fieldAlias = iota
	fieldHost
	fieldUser
	fieldPort
	fieldKey
	fieldGroup
	numFields
)

var fieldLabels = [numFields]string{
	"Alias (rename)",
	"Host",
	"User",
	"Port",
	"Key path",
	"Group",
}

type formModel struct {
	inputs  [numFields]textinput.Model
	focused int
	saved   bool
	quit    bool
	orig    config.Connection
}

func newFormModel(conn config.Connection) formModel {
	var inputs [numFields]textinput.Model
	for i := range inputs {
		t := textinput.New()
		t.CharLimit = 200
		inputs[i] = t
	}

	inputs[fieldAlias].Placeholder = "alias"
	inputs[fieldAlias].SetValue(conn.Alias)

	inputs[fieldHost].Placeholder = "hostname or IP"
	inputs[fieldHost].SetValue(conn.Host)

	inputs[fieldUser].Placeholder = "username"
	inputs[fieldUser].SetValue(conn.User)

	inputs[fieldPort].Placeholder = "22"
	if conn.Port != 0 {
		inputs[fieldPort].SetValue(strconv.Itoa(conn.Port))
	}

	inputs[fieldKey].Placeholder = "~/.ssh/id_rsa"
	inputs[fieldKey].SetValue(conn.KeyPath)

	inputs[fieldGroup].Placeholder = "production"
	inputs[fieldGroup].SetValue(conn.Group)

	inputs[0].Focus()

	return formModel{
		inputs:  inputs,
		focused: 0,
		orig:    conn,
	}
}

func (m formModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m formModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit

		case "enter":
			if m.focused == numFields-1 {
				m.saved = true
				return m, tea.Quit
			}
			m.inputs[m.focused].Blur()
			m.focused++
			m.inputs[m.focused].Focus()
			return m, textinput.Blink

		case "tab", "down":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused + 1) % numFields
			m.inputs[m.focused].Focus()
			return m, textinput.Blink

		case "shift+tab", "up":
			m.inputs[m.focused].Blur()
			m.focused = (m.focused - 1 + numFields) % numFields
			m.inputs[m.focused].Focus()
			return m, textinput.Blink

		case "ctrl+s":
			m.saved = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
	return m, cmd
}

var (
	formLabelActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#06B6D4")).
			Width(18)

	formLabelInactive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				Width(18)

	formBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7C3AED")).
		Padding(1, 2).
		Margin(1, 2)

	formTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1).
			MarginBottom(1)

	formHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1)
)

func (m formModel) View() string {
	if m.saved || m.quit {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(formTitle.Render(fmt.Sprintf("Edit — %s", m.orig.Alias)))
	sb.WriteString("\n")

	for i := range m.inputs {
		var label string
		if i == m.focused {
			label = formLabelActive.Render(fieldLabels[i])
		} else {
			label = formLabelInactive.Render(fieldLabels[i])
		}
		sb.WriteString(fmt.Sprintf("%s  %s\n", label, m.inputs[i].View()))
	}

	sb.WriteString(formHelp.Render("tab/↑↓ navigate  ctrl+s save  enter next/save  esc cancel"))

	return formBox.Render(sb.String())
}

func (m formModel) result() EditResult {
	if !m.saved {
		return EditResult{Saved: false}
	}

	conn := m.orig

	if v := strings.TrimSpace(m.inputs[fieldAlias].Value()); v != "" {
		conn.Alias = v
	}
	if v := strings.TrimSpace(m.inputs[fieldHost].Value()); v != "" {
		conn.Host = v
	}
	if v := strings.TrimSpace(m.inputs[fieldUser].Value()); v != "" {
		conn.User = v
	}
	if v := strings.TrimSpace(m.inputs[fieldPort].Value()); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			conn.Port = p
		}
	}
	conn.KeyPath = strings.TrimSpace(m.inputs[fieldKey].Value())
	conn.Group = strings.TrimSpace(m.inputs[fieldGroup].Value())

	return EditResult{Conn: &conn, Saved: true}
}

// RunEditForm opens the interactive edit form pre-filled with conn's values.
// Returns EditResult with Saved=true and the updated Connection if the user saved.
func RunEditForm(conn config.Connection) EditResult {
	m := newFormModel(conn)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return EditResult{Saved: false}
	}
	fm, ok := finalModel.(formModel)
	if !ok {
		return EditResult{Saved: false}
	}
	return fm.result()
}
