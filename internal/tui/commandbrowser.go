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
	quitting  bool
	noCommand bool
}

func newCommandBrowserModel(conn config.Connection) commandBrowserModel {
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
	l.Title = fmt.Sprintf("Commands — %s  [enter] run  [a] add  [/] filter  [q] back", conn.Alias)
	l.Styles.Title = TitleStyle
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)

	return commandBrowserModel{
		conn:      conn,
		list:      l,
		noCommand: len(items) == 0,
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
	if m.quitting || m.adding || m.selected != nil {
		return ""
	}
	if m.noCommand {
		view := m.list.View()
		emptyHint := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")).
			Italic(true).
			Render("No saved commands yet. Press 'a' to add one.")
		return AppStyle.Render(strings.TrimSpace(view) + "\n\n" + emptyHint)
	}
	return AppStyle.Render(m.list.View())
}

type CommandBrowserResult struct {
	Name    string
	Command string
	AddNew  bool
	Quit    bool
}

func RunCommandBrowser(conn config.Connection) CommandBrowserResult {
	m := newCommandBrowserModel(conn)
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
	if fm.selected != nil {
		return CommandBrowserResult{Name: fm.selected.name, Command: fm.selected.command}
	}
	return CommandBrowserResult{Quit: true}
}
