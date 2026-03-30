package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/saadh393/sshm/internal/config"
	"github.com/saadh393/sshm/internal/scp"
)

// CopyResult is returned after the copy form TUI exits.
type CopyResult struct {
	Direction  scp.Direction
	LocalPath  string
	RemotePath string
	Confirmed  bool
}

// copyField indices
const (
	copyFieldDirection = iota
	copyFieldLocal
	copyFieldRemote
	numCopyFields
)

type copyFormModel struct {
	conn      config.Connection
	direction scp.Direction // Upload or Download
	inputs    [2]textinput.Model
	focused   int // 0=direction toggle, 1=local, 2=remote
	confirmed bool
	quit      bool
}

func newCopyFormModel(conn config.Connection) copyFormModel {
	local := textinput.New()
	local.Placeholder = "~/files/document.txt  (local path)"
	local.CharLimit = 300

	remote := textinput.New()
	remote.Placeholder = "/home/ubuntu/document.txt  (remote path)"
	remote.CharLimit = 300

	local.Focus()

	return copyFormModel{
		conn:      conn,
		direction: scp.Upload,
		inputs:    [2]textinput.Model{local, remote},
		focused:   copyFieldDirection,
	}
}

func (m copyFormModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m copyFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit

		case "ctrl+s":
			if m.canConfirm() {
				m.confirmed = true
				return m, tea.Quit
			}

		case "enter":
			if m.focused == copyFieldDirection {
				// Enter on direction row moves to local field
				m.focused = copyFieldLocal
				m.inputs[0].Focus()
				return m, textinput.Blink
			}
			if m.focused == copyFieldRemote && m.canConfirm() {
				m.confirmed = true
				return m, tea.Quit
			}
			// move forward
			return m.nextField()

		case "tab", "down":
			return m.nextField()

		case "shift+tab", "up":
			return m.prevField()

		case "left", "right":
			// Toggle direction only when focused on the direction row
			if m.focused == copyFieldDirection {
				if m.direction == scp.Upload {
					m.direction = scp.Download
				} else {
					m.direction = scp.Upload
				}
				return m, nil
			}
		}
	}

	// Route key events to the active text input (offset by 1 for direction row)
	if m.focused >= copyFieldLocal {
		idx := m.focused - copyFieldLocal
		var cmd tea.Cmd
		m.inputs[idx], cmd = m.inputs[idx].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m copyFormModel) nextField() (tea.Model, tea.Cmd) {
	if m.focused < copyFieldLocal {
		m.inputs[0].Focus()
	} else if m.focused == copyFieldLocal {
		m.inputs[0].Blur()
		m.inputs[1].Focus()
	}
	m.focused = min(m.focused+1, copyFieldRemote)
	return m, textinput.Blink
}

func (m copyFormModel) prevField() (tea.Model, tea.Cmd) {
	if m.focused == copyFieldLocal {
		m.inputs[0].Blur()
	} else if m.focused == copyFieldRemote {
		m.inputs[1].Blur()
		m.inputs[0].Focus()
	}
	m.focused = max(m.focused-1, copyFieldDirection)
	return m, textinput.Blink
}

func (m copyFormModel) canConfirm() bool {
	return strings.TrimSpace(m.inputs[0].Value()) != "" &&
		strings.TrimSpace(m.inputs[1].Value()) != ""
}

var (
	copyFormBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7C3AED")).
			Padding(1, 2).
			Margin(1, 2)

	copyFormTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#7C3AED")).
			Padding(0, 1).
			MarginBottom(1)

	copyLabelActive = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#06B6D4")).
			Width(18)

	copyLabelInactive = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				Width(18)

	directionActiveStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#10B981")).
				Background(lipgloss.Color("#1F2937")).
				Padding(0, 1)

	directionInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6B7280")).
				Padding(0, 1)

	copyFormHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginTop(1)

	copyPreviewStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#9CA3AF")).
				MarginTop(1).
				Italic(true)
)

func (m copyFormModel) View() string {
	if m.confirmed || m.quit {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(copyFormTitle.Render(fmt.Sprintf("Copy — %s  (%s@%s)", m.conn.Alias, m.conn.User, m.conn.Host)))
	sb.WriteString("\n")

	// Direction toggle row
	dirLabel := "Direction"
	if m.focused == copyFieldDirection {
		sb.WriteString(copyLabelActive.Render(dirLabel))
	} else {
		sb.WriteString(copyLabelInactive.Render(dirLabel))
	}
	sb.WriteString("  ")

	uploadStyle := directionInactiveStyle
	downloadStyle := directionInactiveStyle
	if m.direction == scp.Upload {
		uploadStyle = directionActiveStyle
	} else {
		downloadStyle = directionActiveStyle
	}
	sb.WriteString(uploadStyle.Render("↑ Upload  (local → remote)"))
	sb.WriteString("   ")
	sb.WriteString(downloadStyle.Render("↓ Download  (remote → local)"))
	sb.WriteString("\n")

	// Local path
	localLabel := "Local path"
	if m.focused == copyFieldLocal {
		sb.WriteString(copyLabelActive.Render(localLabel))
	} else {
		sb.WriteString(copyLabelInactive.Render(localLabel))
	}
	sb.WriteString("  ")
	sb.WriteString(m.inputs[0].View())
	sb.WriteString("\n")

	// Remote path
	remoteLabel := "Remote path"
	if m.focused == copyFieldRemote {
		sb.WriteString(copyLabelActive.Render(remoteLabel))
	} else {
		sb.WriteString(copyLabelInactive.Render(remoteLabel))
	}
	sb.WriteString("  ")
	sb.WriteString(m.inputs[1].View())
	sb.WriteString("\n")

	// Preview of the scp command (if both paths filled)
	local := strings.TrimSpace(m.inputs[0].Value())
	remote := strings.TrimSpace(m.inputs[1].Value())
	if local != "" && remote != "" {
		preview := buildPreview(m.conn, local, remote, m.direction)
		sb.WriteString(copyPreviewStyle.Render("Preview: " + preview))
		sb.WriteString("\n")
	}

	helpText := "←/→ toggle direction  tab/↑↓ navigate  ctrl+s confirm  enter next/confirm  esc cancel"
	if !m.canConfirm() {
		helpText = "fill both paths to confirm  |  " + helpText
	}
	sb.WriteString(copyFormHelp.Render(helpText))

	return copyFormBox.Render(sb.String())
}

func buildPreview(c config.Connection, local, remote string, dir scp.Direction) string {
	args := []string{"scp"}
	if c.Port != 0 && c.Port != 22 {
		args = append(args, fmt.Sprintf("-P %d", c.Port))
	}
	if c.KeyPath != "" {
		args = append(args, "-i", c.KeyPath)
	}
	remoteTarget := fmt.Sprintf("%s@%s:%s", c.User, c.Host, remote)
	if dir == scp.Upload {
		args = append(args, local, remoteTarget)
	} else {
		args = append(args, remoteTarget, local)
	}
	return strings.Join(args, " ")
}

func (m copyFormModel) result() CopyResult {
	if !m.confirmed {
		return CopyResult{Confirmed: false}
	}
	return CopyResult{
		Direction:  m.direction,
		LocalPath:  strings.TrimSpace(m.inputs[0].Value()),
		RemotePath: strings.TrimSpace(m.inputs[1].Value()),
		Confirmed:  true,
	}
}

// RunCopyForm opens the interactive copy form for the given connection.
// Returns CopyResult with Confirmed=true if the user confirmed the transfer.
func RunCopyForm(conn config.Connection) CopyResult {
	m := newCopyFormModel(conn)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return CopyResult{Confirmed: false}
	}
	fm, ok := finalModel.(copyFormModel)
	if !ok {
		return CopyResult{Confirmed: false}
	}
	return fm.result()
}

