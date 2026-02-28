package form

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stickpro/kyp/internal/crypto"
	"github.com/stickpro/kyp/internal/tui/styles"
	"github.com/stickpro/kyp/internal/vault"
)

const (
	fieldTitle = iota
	fieldUsername
	fieldPassword
	fieldURL
	fieldNotes
	fieldTOTPSecret
	fieldTOTPIssuer
	btnSave
	btnCancel
	fieldsCount
)

type EntrySavedMsg struct{}
type BackMsg struct{}
type ErrMsg struct{ Err error }

type Model struct {
	inputs  [7]textinput.Model
	focused int
	vault   *vault.Vault
	editID  *string
	err     error
}

func New(v *vault.Vault, entry *vault.EntryDTO) Model {
	iTitle := textinput.New()
	iTitle.Placeholder = "Title"
	iTitle.Focus()

	iUsername := textinput.New()
	iUsername.Placeholder = "Username"

	iPassword := textinput.New()
	iPassword.Placeholder = "Password"
	iPassword.EchoMode = textinput.EchoPassword

	iURL := textinput.New()
	iURL.Placeholder = "URL"

	iNotes := textinput.New()
	iNotes.Placeholder = "Notes"

	iTOTPSecret := textinput.New()
	iTOTPSecret.Placeholder = "TOTP secret (optional)"

	iTOTPIssuer := textinput.New()
	iTOTPIssuer.Placeholder = "TOTP issuer (optional)"

	m := Model{
		inputs:  [7]textinput.Model{iTitle, iUsername, iPassword, iURL, iNotes, iTOTPSecret, iTOTPIssuer},
		focused: fieldTitle,
		vault:   v,
	}

	if entry != nil {
		m.editID = entry.ID
		m.inputs[fieldTitle].SetValue(entry.Title)
		if entry.Username != nil {
			m.inputs[fieldUsername].SetValue(*entry.Username)
		}
		if entry.Password != nil {
			m.inputs[fieldPassword].SetValue(*entry.Password)
		}
		if entry.URL != nil {
			m.inputs[fieldURL].SetValue(*entry.URL)
		}
		if entry.Notes != nil {
			m.inputs[fieldNotes].SetValue(*entry.Notes)
		}
		if entry.TOTPSecret != nil {
			m.inputs[fieldTOTPSecret].SetValue(*entry.TOTPSecret)
		}
		if entry.TOTPIssuer != nil {
			m.inputs[fieldTOTPIssuer].SetValue(*entry.TOTPIssuer)
		}
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlS:
			return m, m.submit()

		case tea.KeyCtrlG:
			if m.focused == fieldPassword {
				pwd, err := crypto.GeneratePassword(20, true, true, true)
				if err == nil {
					m.inputs[fieldPassword].SetValue(pwd)
				}
			}

		case tea.KeyCtrlP:
			if m.inputs[fieldPassword].EchoMode == textinput.EchoPassword {
				m.inputs[fieldPassword].EchoMode = textinput.EchoNormal
			} else {
				m.inputs[fieldPassword].EchoMode = textinput.EchoPassword
			}

		case tea.KeyEsc:
			return m, func() tea.Msg { return BackMsg{} }

		case tea.KeyTab, tea.KeyDown:
			m.setFocus((m.focused + 1) % fieldsCount)

		case tea.KeyShiftTab, tea.KeyUp:
			m.setFocus((m.focused - 1 + fieldsCount) % fieldsCount)

		case tea.KeyEnter:
			switch m.focused {
			case btnCancel:
				return m, func() tea.Msg { return BackMsg{} }
			case btnSave:
				return m, m.submit()
			default:
				m.setFocus(m.focused + 1)
			}

		default:
			m.err = nil
		}

	case ErrMsg:
		m.err = msg.Err
		return m, nil

	case EntrySavedMsg:
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
	isEdit := m.editID != nil
	titleText := "New Entry"
	if isEdit {
		titleText = "Edit Entry"
	}

	title := styles.TitleStyle.Render(titleText)
	hint := styles.HintStyle.Render("tab: next • ctrl+s: save • ctrl+g: generate • ctrl+p: show/hide password • esc: back")

	var errStr string
	if m.err != nil {
		errStr = styles.ErrStyle.Render(m.err.Error())
	}

	saveBtn := styles.InactiveStyle.Render("Save")
	cancelBtn := styles.InactiveStyle.Render("Cancel")
	if m.focused == btnSave {
		saveBtn = styles.ActiveStyle.Render("Save")
	}
	if m.focused == btnCancel {
		cancelBtn = styles.ActiveStyle.Render("Cancel")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Top, saveBtn, "  ", cancelBtn)

	sep := styles.HintStyle.Render("─── TOTP (optional) ───")

	fields := m.inputs[fieldTitle].View() + "\n" +
		m.inputs[fieldUsername].View() + "\n" +
		m.inputs[fieldPassword].View() + "\n" +
		m.inputs[fieldURL].View() + "\n" +
		m.inputs[fieldNotes].View() + "\n\n" +
		" " + sep + "\n" +
		m.inputs[fieldTOTPSecret].View() + "\n" +
		m.inputs[fieldTOTPIssuer].View()

	return fmt.Sprintf("%s\n\n %s\n\n %s\n\n %s\n\n %s",
		title,
		fields,
		buttons,
		hint,
		errStr,
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
	title := m.inputs[fieldTitle].Value()
	if title == "" {
		return func() tea.Msg {
			return ErrMsg{Err: fmt.Errorf("title is required")}
		}
	}

	dto := vault.EntryDTO{
		Title:    title,
		Username: strPtr(m.inputs[fieldUsername].Value()),
		Password: strPtr(m.inputs[fieldPassword].Value()),
		URL:      strPtr(m.inputs[fieldURL].Value()),
		Notes:    strPtr(m.inputs[fieldNotes].Value()),
		// TOTP defaults
		TOTPAlgorithm: "SHA1",
		TOTPDigits:    6,
		TOTPPeriod:    30,
	}

	if s := m.inputs[fieldTOTPSecret].Value(); s != "" {
		dto.TOTPSecret = &s
	}
	if s := m.inputs[fieldTOTPIssuer].Value(); s != "" {
		dto.TOTPIssuer = &s
	}

	if m.editID != nil {
		id := *m.editID
		return func() tea.Msg {
			if err := m.vault.UpdateEntry(context.Background(), id, dto); err != nil {
				return ErrMsg{Err: err}
			}
			return EntrySavedMsg{}
		}
	}

	return func() tea.Msg {
		if err := m.vault.CreateEntry(context.Background(), dto); err != nil {
			return ErrMsg{Err: err}
		}
		return EntrySavedMsg{}
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
