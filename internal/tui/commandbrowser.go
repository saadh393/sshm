package tui

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/saadh393/sshm/internal/config"
)

type commandItem struct {
	name    string
	command string
}

func (i commandItem) Title() string       { return i.name }
func (i commandItem) Description() string { return i.command }
func (i commandItem) FilterValue() string { return i.name + " " + i.command }

type commandDelegate struct{}

func (d commandDelegate) Height() int                             { return 2 }
func (d commandDelegate) Spacing() int                            { return 0 }
func (d commandDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d commandDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(commandItem)
	if !ok {
		return
	}

	isSelected := index == m.Index()
	name := AliasStyle.Render(item.name)
	cmd := HostStyle.UnsetWidth().Render(item.command)
	row := lipgloss.JoinHorizontal(lipgloss.Top, name, cmd)

	if isSelected {
		row = SelectedItemStyle.Render("> ") + row
	} else {
		row = "  " + row
	}
	fmt.Fprintln(w, row)
}

type commandBrowserModel struct {
	conn      config.Connection
	list      list.Model
	selected  *commandItem
	adding    bool
	updating  bool
	deleting  bool
	quitting  bool
	noCommand bool
	status    string
}

func newCommandBrowserModel(conn config.Connection, status string) commandBrowserModel {
	names := make([]string, 0, len(conn.Commands))
	for name := range conn.Commands {
		names = append(names, name)
	}
	sort.Strings(names)

	items := make([]list.Item, 0, len(names))
	for _, n := range names {
		items = append(items, commandItem{name: n, command: conn.Commands[n]})
	}

	l := list.New(items, commandDelegate{}, 80, 24)
	l.Title = fmt.Sprintf("Commands — %s  [enter] run  [a] add  [u] update  [d] delete  [/] filter  [q] back", conn.Alias)
	l.Styles.Title = TitleStyle
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)

	return commandBrowserModel{
		conn:      conn,
		list:      l,
		noCommand: len(items) == 0,
		status:    status,
	}
}

func (m commandBrowserModel) Init() tea.Cmd { return nil }

func (m commandBrowserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := AppStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}
		switch msg.String() {
		case "enter":
			if item, ok := m.list.SelectedItem().(commandItem); ok {
				chosen := item
				m.selected = &chosen
				return m, tea.Quit
			}
		case "a":
			m.adding = true
			return m, tea.Quit
		case "u":
			if item, ok := m.list.SelectedItem().(commandItem); ok {
				chosen := item
				m.selected = &chosen
				m.updating = true
				return m, tea.Quit
			}
			m.status = "No command selected. Use ↑/↓ to choose one, then press 'u' to update."
		case "d":
			if item, ok := m.list.SelectedItem().(commandItem); ok {
				chosen := item
				m.selected = &chosen
				m.deleting = true
				return m, tea.Quit
			}
			m.status = "No command selected. Use ↑/↓ to choose one, then press 'd' to delete."
		case "q", "esc":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m commandBrowserModel) View() string {
	if m.quitting || m.adding || m.updating || m.deleting || m.selected != nil {
		return ""
	}
	statusView := ""
	if strings.TrimSpace(m.status) != "" {
		statusView = "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true).
			Render(m.status)
	}
	if m.noCommand {
		view := m.list.View()
		emptyHint := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Italic(true).
			Render("No saved commands yet. Press 'a' to add one.")
		return AppStyle.Render(strings.TrimSpace(view) + "\n\n" + emptyHint + statusView)
	}
	return AppStyle.Render(m.list.View() + statusView)
}

type CommandBrowserResult struct {
	Name    string
	Command string
	AddNew  bool
	Update  bool
	Delete  bool
	Quit    bool
}

func RunCommandBrowser(conn config.Connection, status string) CommandBrowserResult {
	m := newCommandBrowserModel(conn, status)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return CommandBrowserResult{Quit: true}
	}
	fm, ok := finalModel.(commandBrowserModel)
	if !ok || fm.quitting {
		return CommandBrowserResult{Quit: true}
	}
	if fm.adding {
		return CommandBrowserResult{AddNew: true}
	}
	if fm.updating && fm.selected != nil {
		return CommandBrowserResult{Name: fm.selected.name, Command: fm.selected.command, Update: true}
	}
	if fm.deleting && fm.selected != nil {
		return CommandBrowserResult{Name: fm.selected.name, Command: fm.selected.command, Delete: true}
	}
	if fm.selected != nil {
		return CommandBrowserResult{Name: fm.selected.name, Command: fm.selected.command}
	}
	return CommandBrowserResult{Quit: true}
}
