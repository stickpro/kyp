package unlock

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stickpro/kyp/internal/models"
	"github.com/stickpro/kyp/internal/tui/styles"
	"github.com/stickpro/kyp/internal/vault"
)

type step int

const (
	stepSelectVault step = iota
	stepPassword
)

type Model struct {
	input    textinput.Model
	vault    *vault.Vault
	vaults   []*models.VaultMetum
	selected int
	step     step
	err      error
	width    int
	height   int
}

type VaultUnlockedMsg struct {
	Vault *vault.Vault
}

type ErrMsg struct {
	Err error
}

func New(v *vault.Vault, vaults []*models.VaultMetum) Model {
	ti := textinput.New()
	ti.Placeholder = "Enter master password"
	ti.EchoMode = textinput.EchoPassword
	ti.Width = 32

	m := Model{
		input:  ti,
		vault:  v,
		vaults: vaults,
	}

	if len(vaults) == 1 {
		m.selected = 0
		m.step = stepPassword
		m.input.Focus()
	} else {
		m.step = stepSelectVault
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	if m.step == stepPassword {
		return textinput.Blink
	}
	return nil
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

		case tea.KeyUp:
			if m.step == stepSelectVault && m.selected > 0 {
				m.selected--
			}

		case tea.KeyDown:
			if m.step == stepSelectVault && m.selected < len(m.vaults)-1 {
				m.selected++
			}

		case tea.KeyEnter:
			if m.step == stepSelectVault {
				m.step = stepPassword
				m.input.Focus()
				return m, textinput.Blink
			}
			return m, openVault(m.vault, m.input.Value(), m.vaults[m.selected].Name)

		case tea.KeyEsc:
			if m.step == stepPassword && len(m.vaults) > 1 {
				m.step = stepSelectVault
				m.input.Blur()
				m.input.Reset()
				m.err = nil
			}

		default:
			if m.step == stepPassword {
				m.err = nil
			}
		}

	case VaultUnlockedMsg:
		return m, tea.Quit

	case ErrMsg:
		m.err = fmt.Errorf("invalid master password")
		return m, nil
	}

	if m.step == stepPassword {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	title := styles.TitleStyle.Render("Keep Your Passwords")

	var body string
	var hint string

	if m.step == stepSelectVault {
		hint = styles.HintStyle.Render("↑/↓: select • enter: confirm • ctrl+c: quit")

		var rows []string
		rows = append(rows, styles.LabelStyle.Render("Select vault:"))
		for i, v := range m.vaults {
			if i == m.selected {
				rows = append(rows, styles.ActiveStyle.Render("> "+v.Name))
			} else {
				rows = append(rows, styles.InactiveStyle.Render("  "+v.Name))
			}
		}

		body = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(0, 1).
			Width(36).
			Render(strings.Join(rows, "\n"))
	} else {
		backHint := ""
		if len(m.vaults) > 1 {
			backHint = " • esc: back"
		}
		hint = styles.HintStyle.Render("enter: unlock • ctrl+c: quit" + backHint)

		vaultLabel := styles.LabelStyle.Render("Vault: ") + styles.ValueStyle.Render(m.vaults[m.selected].Name)

		body = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("99")).
			Padding(0, 1).
			Width(36).
			Render(vaultLabel + "\n" + m.input.View())
	}

	var errStr string
	if m.err != nil {
		errStr = "\n" + styles.ErrStyle.Render(m.err.Error())
	}

	content := strings.Join([]string{
		title,
		"",
		body,
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

func openVault(v *vault.Vault, password, name string) tea.Cmd {
	return func() tea.Msg {
		err := v.Open(context.Background(), password, name)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return VaultUnlockedMsg{Vault: v}
	}
}
