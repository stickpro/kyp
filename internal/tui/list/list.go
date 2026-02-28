package list

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stickpro/kyp/internal/vault"
)

type Item struct {
	entry vault.EntryDTO
}

func (i Item) Title() string { return i.entry.Title }

func (i Item) Description() string {
	var parts []string
	if i.entry.Username != nil {
		parts = append(parts, *i.entry.Username)
	}
	if i.entry.URL != nil {
		parts = append(parts, *i.entry.URL)
	}
	if len(parts) == 0 {
		return "—"
	}
	return fmt.Sprintf("%s", joinParts(parts))
}

func (i Item) FilterValue() string {
	s := i.entry.Title
	if i.entry.Username != nil {
		s += " " + *i.entry.Username
	}
	return s
}

func (i Item) Entry() vault.EntryDTO { return i.entry }

type EntriesLoadedMsg struct {
	Items []list.Item
}

type EntrySelectedMsg struct {
	Entry vault.EntryDTO
}

type NewEntryMsg struct{}

type ErrMsg struct {
	Err error
}

func ReloadCmd(v *vault.Vault) tea.Cmd {
	return loadEntries(v)
}

type Model struct {
	list  list.Model
	vault *vault.Vault
	err   error
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func New(v *vault.Vault, width, height int) Model {
	l := list.New(nil, list.NewDefaultDelegate(), width, height)
	l.Title = "Keep Your Passwords"
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("99")).
		Padding(0, 1)

	return Model{
		list:  l,
		vault: v,
	}
}

func (m *Model) Init() tea.Cmd {
	return loadEntries(m.vault)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(Item); ok {
				e := i.Entry()
				return m, func() tea.Msg { return EntrySelectedMsg{Entry: e} }
			}
		case "n":
			return m, func() tea.Msg { return NewEntryMsg{} }
		}

	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)

	case EntriesLoadedMsg:
		m.list.SetItems(msg.Items)
		return m, nil

	case ErrMsg:
		m.err = msg.Err
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	if m.err != nil {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.err.Error())
	}
	return docStyle.Render(m.list.View())
}

func (m *Model) SelectedEntry() *vault.EntryDTO {
	if i, ok := m.list.SelectedItem().(Item); ok {
		e := i.Entry()
		return &e
	}
	return nil
}

func loadEntries(v *vault.Vault) tea.Cmd {
	return func() tea.Msg {
		entries, err := v.ListEntries(context.Background(), 100, 0)
		if err != nil {
			return ErrMsg{Err: err}
		}
		items := make([]list.Item, len(entries))
		for i, e := range entries {
			items[i] = Item{entry: e}
		}
		return EntriesLoadedMsg{Items: items}
	}
}

func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += " • "
		}
		result += p
	}
	return result
}
