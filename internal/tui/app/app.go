package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stickpro/kyp/internal/tui/create"
	"github.com/stickpro/kyp/internal/tui/detail"
	"github.com/stickpro/kyp/internal/tui/form"
	listscreen "github.com/stickpro/kyp/internal/tui/list"
	"github.com/stickpro/kyp/internal/tui/unlock"
	"github.com/stickpro/kyp/internal/vault"
)

type screen int

const (
	screenCreate screen = iota
	screenUnlock
	screenList
	screenDetail
	screenForm
)

type Model struct {
	screen screen
	vault  *vault.Vault
	create *create.Model
	unlock *unlock.Model
	list   *listscreen.Model
	detail *detail.Model
	form   *form.Model
	width  int
	height int
}

func New(v *vault.Vault) Model {
	if v.IsInitialized(context.Background()) {
		u := unlock.New(v)
		return Model{
			screen: screenUnlock,
			vault:  v,
			unlock: &u,
		}
	}

	c := create.New(v)
	return Model{
		screen: screenCreate,
		vault:  v,
		create: &c,
	}
}

func (m Model) Init() tea.Cmd {
	switch m.screen {
	case screenCreate:
		return m.create.Init()
	case screenUnlock:
		return m.unlock.Init()
	case screenList:
		return m.list.Init()
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case create.VaultCreatedMsg:
		l := listscreen.New(m.vault, m.width, m.height)
		m.list = &l
		m.screen = screenList
		return m, m.list.Init()

	case unlock.VaultUnlockedMsg:
		l := listscreen.New(m.vault, m.width, m.height)
		m.list = &l
		m.screen = screenList
		return m, m.list.Init()

	case listscreen.EntrySelectedMsg:
		d := detail.New(msg.Entry)
		m.detail = &d
		m.screen = screenDetail
		return m, m.detail.Init()

	case listscreen.NewEntryMsg:
		f := form.New(m.vault, nil)
		m.form = &f
		m.screen = screenForm
		return m, m.form.Init()

	case detail.BackMsg:
		m.screen = screenList
		return m, listscreen.ReloadCmd(m.vault)

	case detail.EditMsg:
		f := form.New(m.vault, &msg.Entry)
		m.form = &f
		m.screen = screenForm
		return m, m.form.Init()

	case form.BackMsg:
		if m.detail != nil {
			m.screen = screenDetail
		} else {
			m.screen = screenList
		}
		return m, nil

	case form.EntrySavedMsg:
		m.detail = nil
		m.screen = screenList
		return m, listscreen.ReloadCmd(m.vault)
	}

	switch m.screen {
	case screenCreate:
		_, cmd := m.create.Update(msg)
		return m, cmd
	case screenUnlock:
		_, cmd := m.unlock.Update(msg)
		return m, cmd
	case screenList:
		_, cmd := m.list.Update(msg)
		return m, cmd
	case screenDetail:
		_, cmd := m.detail.Update(msg)
		return m, cmd
	case screenForm:
		_, cmd := m.form.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) View() string {
	switch m.screen {
	case screenCreate:
		return m.create.View()
	case screenUnlock:
		return m.unlock.View()
	case screenList:
		return m.list.View()
	case screenDetail:
		return m.detail.View()
	case screenForm:
		return m.form.View()
	}
	return ""
}
