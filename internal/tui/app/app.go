package app

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stickpro/kyp/internal/tui/create"
	"github.com/stickpro/kyp/internal/tui/detail"
	"github.com/stickpro/kyp/internal/tui/form"
	listscreen "github.com/stickpro/kyp/internal/tui/list"
	"github.com/stickpro/kyp/internal/tui/unlock"
	"github.com/stickpro/kyp/internal/vault"
)

const lockTickInterval = 30 * time.Second

type lockTickMsg struct{}

type screen int

const (
	screenCreate screen = iota
	screenUnlock
	screenList
	screenDetail
	screenForm
)

type Model struct {
	screen       screen
	vault        *vault.Vault
	create       *create.Model
	unlock       *unlock.Model
	list         *listscreen.Model
	detail       *detail.Model
	form         *form.Model
	width        int
	height       int
	lockTimeout  time.Duration
	lastActivity time.Time
}

func New(v *vault.Vault, lockTimeout time.Duration) Model {
	ctx := context.Background()
	vaults, err := v.ListVaults(ctx)
	if err == nil && len(vaults) > 0 {
		u := unlock.New(v, vaults)
		return Model{
			screen:      screenUnlock,
			vault:       v,
			unlock:      &u,
			lockTimeout: lockTimeout,
		}
	}

	c := create.New(v)
	return Model{
		screen:      screenCreate,
		vault:       v,
		create:      &c,
		lockTimeout: lockTimeout,
	}
}

func lockTick() tea.Cmd {
	return tea.Tick(lockTickInterval, func(time.Time) tea.Msg {
		return lockTickMsg{}
	})
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

	case tea.KeyMsg:
		if m.isLocked() {
			m.lastActivity = time.Now()
		}

	case lockTickMsg:
		if m.isLocked() && m.lockTimeout > 0 && time.Since(m.lastActivity) >= m.lockTimeout {
			return m.lock()
		}
		return m, lockTick()

	case create.VaultCreatedMsg:
		l := listscreen.New(m.vault, m.width, m.height)
		m.list = &l
		m.screen = screenList
		m.lastActivity = time.Now()
		w, h := m.width, m.height
		return m, tea.Batch(m.list.Init(), lockTick(), func() tea.Msg {
			return tea.WindowSizeMsg{Width: w, Height: h}
		})

	case unlock.VaultUnlockedMsg:
		l := listscreen.New(m.vault, m.width, m.height)
		m.list = &l
		m.screen = screenList
		m.lastActivity = time.Now()
		w, h := m.width, m.height
		return m, tea.Batch(m.list.Init(), lockTick(), func() tea.Msg {
			return tea.WindowSizeMsg{Width: w, Height: h}
		})

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

	case detail.DeleteMsg:
		id := msg.ID
		m.detail = nil
		m.screen = screenList
		return m, func() tea.Msg {
			_ = m.vault.DeleteEntry(context.Background(), id)
			return listscreen.ReloadCmd(m.vault)()
		}

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
			return m, m.detail.Init()
		}
		m.screen = screenList
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

// isLocked returns true when the vault is open and should be monitored for inactivity.
func (m Model) isLocked() bool {
	return m.screen == screenList || m.screen == screenDetail || m.screen == screenForm
}

// lock closes the vault and returns to the unlock screen.
func (m Model) lock() (tea.Model, tea.Cmd) {
	m.vault.Close()
	ctx := context.Background()
	vaults, _ := m.vault.ListVaults(ctx)
	u := unlock.New(m.vault, vaults)
	m.unlock = &u
	m.screen = screenUnlock
	m.list = nil
	m.detail = nil
	m.form = nil
	w, h := m.width, m.height
	return m, tea.Batch(m.unlock.Init(), func() tea.Msg {
		return tea.WindowSizeMsg{Width: w, Height: h}
	})
}
