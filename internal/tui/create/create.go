package create

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stickpro/kyp/internal/tui/styles"
	"github.com/stickpro/kyp/internal/vault"
)

const (
	fieldName = iota
	fieldPassword
	fieldConfirm
	btnCreate
	btnCancel
	fieldsCount
)

type VaultCreatedMsg struct {
	Vault *vault.Vault
}

type ErrMsg struct {
	Err error
}

type Model struct {
	inputs  [3]textinput.Model
	focused int
	vault   *vault.Vault
	err     error
	width   int
	height  int
}

func New(v *vault.Vault) Model {
	iName := textinput.New()
	iName.Placeholder = "Vault name"
	iName.SetValue("default")

	iPassword := textinput.New()
	iPassword.Placeholder = "Master password"
	iPassword.EchoMode = textinput.EchoPassword
	iPassword.Focus()

	iConfirm := textinput.New()
	iConfirm.Placeholder = "Confirm master password"
	iConfirm.EchoMode = textinput.EchoPassword

	return Model{
		inputs:  [3]textinput.Model{iName, iPassword, iConfirm},
		focused: fieldName,
		vault:   v,
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

		case tea.KeyTab, tea.KeyDown:
			m.setFocus((m.focused + 1) % fieldsCount)

		case tea.KeyShiftTab, tea.KeyUp:
			m.setFocus((m.focused - 1 + fieldsCount) % fieldsCount)

		case tea.KeyEnter:
			switch m.focused {
			case btnCancel:
				return m, tea.Quit
			case btnCreate:
				return m, m.submit()
			default:
				m.setFocus(m.focused + 1)
			}
		default:
			m.err = nil
		}

	case VaultCreatedMsg:
		return m, nil

	case ErrMsg:
		m.err = msg.Err
		return m, nil
	}

	if m.focused < len(m.inputs) {
		var cmd tea.Cmd
		m.inputs[m.focused], cmd = m.inputs[m.focused].Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *Model) View() string {
	title := styles.TitleStyle.Render("Keep Your Passwords — Create Vault")
	hint := styles.HintStyle.Render("tab: next field • enter: confirm • ctrl+c: quit")

	var errStr string
	if m.err != nil {
		errStr = styles.ErrStyle.Render(m.err.Error())
	}

	createBtn := styles.InactiveStyle.Render("Create")
	cancelBtn := styles.InactiveStyle.Render("Cancel")
	if m.focused == btnCreate {
		createBtn = styles.ActiveStyle.Render("Create")
	}
	if m.focused == btnCancel {
		cancelBtn = styles.ActiveStyle.Render("Cancel")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, createBtn, "  ", cancelBtn)

	content := fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s",
		title,
		m.inputs[fieldName].View()+"\n"+m.inputs[fieldPassword].View()+"\n"+m.inputs[fieldConfirm].View(),
		buttons,
		hint,
		errStr,
	)

	if m.width == 0 || m.height == 0 {
		return content
	}

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		content,
	)
}

func (m *Model) setFocus(idx int) {
	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Blur()
	}
	m.focused = idx
	if m.focused < len(m.inputs) {
		m.inputs[m.focused].Focus()
	}
}

func (m *Model) submit() tea.Cmd {
	password := m.inputs[fieldPassword].Value()
	confirm := m.inputs[fieldConfirm].Value()

	if password != confirm {
		return func() tea.Msg {
			return ErrMsg{Err: fmt.Errorf("passwords do not match")}
		}
	}

	name := m.inputs[fieldName].Value()
	if name == "" {
		name = "default"
	}

	return func() tea.Msg {
		err := m.vault.Create(context.Background(), password, name)
		if err != nil {
			return ErrMsg{Err: err}
		}
		return VaultCreatedMsg{Vault: m.vault}
	}
}
