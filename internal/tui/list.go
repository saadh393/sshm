package tui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/saadh393/sshm/internal/config"
)

// ConnectMsg is sent when the user presses Enter on an item.
type ConnectMsg struct {
	Conn config.Connection
}

// QuitMsg is sent when the user quits without selecting.
type QuitMsg struct{}

// connItem wraps a config.Connection to implement list.Item.
type connItem struct {
	conn config.Connection
}

func (i connItem) Title() string {
	return i.conn.Alias
}

func (i connItem) Description() string {
	host := fmt.Sprintf("%s@%s", i.conn.User, i.conn.Host)
	if i.conn.Port != 0 && i.conn.Port != 22 {
		host += fmt.Sprintf(":%d", i.conn.Port)
	}
	group := i.conn.Group
	if group == "" {
		group = "—"
	}
	return fmt.Sprintf("%-28s %s", host, group)
}

func (i connItem) FilterValue() string {
	return strings.Join([]string{
		i.conn.Alias,
		i.conn.Host,
		i.conn.User,
		i.conn.Group,
	}, " ")
}

// itemDelegate renders each row.
type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 2 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(connItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()

	aliasStr := AliasStyle.Render(item.conn.Alias)
	hostStr := fmt.Sprintf("%s@%s", item.conn.User, item.conn.Host)
	if item.conn.Port != 0 && item.conn.Port != 22 {
		hostStr += fmt.Sprintf(":%d", item.conn.Port)
	}
	hostRendered := HostStyle.Render(hostStr)

	group := item.conn.Group
	if group == "" {
		group = "—"
	}
	groupRendered := GroupStyle.Render(group)

	row := lipgloss.JoinHorizontal(lipgloss.Top, aliasStr, hostRendered, groupRendered)

	if isSelected {
		row = SelectedItemStyle.Render("> ") + row
	} else {
		row = "  " + row
	}

	fmt.Fprintln(w, row)
}

// Model is the top-level Bubble Tea model for the connection list.
type Model struct {
	list                  list.Model
	chosen                *config.Connection
	quitting              bool
	openCommands          bool
	enableCommandShortcut bool
	allConns              []config.Connection
}

// NewModel constructs a Model from a slice of connections.
func NewModel(conns []config.Connection, width, height int, title string, enableCommandShortcut bool) Model {
	items := make([]list.Item, len(conns))
	for i, c := range conns {
		items[i] = connItem{conn: c}
	}

	delegate := itemDelegate{}

	l := list.New(items, delegate, width, height)
	l.Title = title
	l.Styles.Title = TitleStyle
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)

	return Model{
		list:                  l,
		enableCommandShortcut: enableCommandShortcut,
		allConns:              conns,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := AppStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case tea.KeyMsg:
		// Don't intercept keys while filtering
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(connItem); ok {
				m.chosen = &item.conn
				return m, tea.Quit
			}
		case "c":
			if m.enableCommandShortcut {
				if item, ok := m.list.SelectedItem().(connItem); ok {
					m.chosen = &item.conn
					m.openCommands = true
					return m, tea.Quit
				}
			}
		case "q", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if m.quitting || m.chosen != nil {
		return ""
	}
	return AppStyle.Render(m.list.View())
}

// Result holds the outcome after the TUI exits.
type Result struct {
	Conn         *config.Connection
	OpenCommands bool
	Quit         bool
}

// Run launches the TUI and blocks until the user selects a connection or quits.
func Run(conns []config.Connection) Result {
	if len(conns) == 0 {
		return Result{Quit: true}
	}

	m := NewModel(conns, 80, 24, "SSH Connections  [enter] connect  [c] commands  [/] filter  [q] quit", true)
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
		return Result{Conn: fm.chosen, OpenCommands: fm.openCommands}
	}
	return Result{Quit: true}
}
