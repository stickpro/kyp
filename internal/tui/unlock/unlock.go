package unlock

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stickpro/kyp/internal/tui/styles"
	"github.com/stickpro/kyp/internal/vault"
)

type Model struct {
	input  textinput.Model
	vault  *vault.Vault
	err    error
	width  int
	height int
}

type VaultUnlockedMsg struct {
	Vault *vault.Vault
}

type ErrMsg struct {
	Err error
}

func New(v *vault.Vault) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter master password"
	ti.EchoMode = textinput.EchoPassword
	ti.Width = 32
	ti.Focus()
	return Model{
		input: ti,
		vault: v,
	}
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			return m, openVault(m.vault, m.input.Value())
		default:
			m.err = nil
		}

	case VaultUnlockedMsg:
		return m, tea.Quit

	case ErrMsg:
		m.err = fmt.Errorf("invalid master password")
		return m, nil
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	title := styles.TitleStyle.Render("Keep Your Passwords")
	hint := styles.HintStyle.Render("enter: unlock • ctrl+c: quit")

	var errStr string
	if m.err != nil {
		errStr = "\n" + styles.ErrStyle.Render(m.err.Error())
	}

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")).
		Padding(0, 1).
		Width(36).
		Render(m.input.View())

	content := strings.Join([]string{
		title,
		"",
		inputBox,
		hint,
	}, "\n") + errStr

	if m.width == 0 || m.height == 0 {
		return content
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func openVault(v *vault.Vault, password string) tea.Cmd {
	return func() tea.Msg {
		err := v.Open(context.Background(), password, "default")
		if err != nil {
			return ErrMsg{Err: err}
		}
		return VaultUnlockedMsg{Vault: v}
	}
}
