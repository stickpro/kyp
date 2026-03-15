package form

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
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
	inputs    [6]textinput.Model
	notesArea textarea.Model
	focused   int
	vault     *vault.Vault
	editID    *string
	err       error
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

	iTOTPSecret := textinput.New()
	iTOTPSecret.Placeholder = "TOTP secret (optional)"

	iTOTPIssuer := textinput.New()
	iTOTPIssuer.Placeholder = "TOTP issuer (optional)"

	ta := textarea.New()
	ta.Placeholder = "Notes"
	ta.SetHeight(3)
	ta.SetWidth(40)
	ta.ShowLineNumbers = false

	m := Model{
		inputs:    [6]textinput.Model{iTitle, iUsername, iPassword, iURL, iTOTPSecret, iTOTPIssuer},
		notesArea: ta,
		focused:   fieldTitle,
		vault:     v,
	}

	if entry != nil {
		m.editID = entry.ID
		m.inputs[0].SetValue(entry.Title)
		if entry.Username != nil {
			m.inputs[1].SetValue(*entry.Username)
		}
		if entry.Password != nil {
			m.inputs[2].SetValue(*entry.Password)
		}
		if entry.URL != nil {
			m.inputs[3].SetValue(*entry.URL)
		}
		if entry.Notes != nil {
			m.notesArea.SetValue(*entry.Notes)
		}
		if entry.TOTPSecret != nil {
			m.inputs[4].SetValue(*entry.TOTPSecret)
		}
		if entry.TOTPIssuer != nil {
			m.inputs[5].SetValue(*entry.TOTPIssuer)
		}
	}

	return m
}

func (m *Model) Init() tea.Cmd {
	return textinput.Blink
}

func inputIdx(focused int) int {
	if focused < fieldNotes {
		return focused
	}
	return focused - 1
}

func (m *Model) setFocus(idx int) tea.Cmd {
	if m.focused == fieldNotes {
		m.notesArea.Blur()
	} else if m.focused < fieldNotes || (m.focused > fieldNotes && m.focused < btnSave) {
		m.inputs[inputIdx(m.focused)].Blur()
	}

	m.focused = idx

	if m.focused == fieldNotes {
		return m.notesArea.Focus()
	} else if m.focused < fieldNotes || (m.focused > fieldNotes && m.focused < btnSave) {
		m.inputs[inputIdx(m.focused)].Focus()
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlS:
			cmd := m.submit()
			return m, cmd

		case tea.KeyCtrlG:
			if m.focused == fieldPassword {
				pwd, err := crypto.GeneratePassword(20, true, true, true)
				if err == nil {
					m.inputs[inputIdx(fieldPassword)].SetValue(pwd)
				}
			}

		case tea.KeyCtrlP:
			if m.focused == fieldPassword {
				idx := inputIdx(fieldPassword)
				if m.inputs[idx].EchoMode == textinput.EchoPassword {
					m.inputs[idx].EchoMode = textinput.EchoNormal
				} else {
					m.inputs[idx].EchoMode = textinput.EchoPassword
				}
			}

		case tea.KeyEsc:
			return m, func() tea.Msg { return BackMsg{} }

		case tea.KeyTab:
			cmd := m.setFocus((m.focused + 1) % fieldsCount)
			return m, cmd

		case tea.KeyShiftTab:
			cmd := m.setFocus((m.focused - 1 + fieldsCount) % fieldsCount)
			return m, cmd

		case tea.KeyDown:
			if m.focused != fieldNotes {
				cmd := m.setFocus((m.focused + 1) % fieldsCount)
				return m, cmd
			}

		case tea.KeyUp:
			if m.focused != fieldNotes {
				cmd := m.setFocus((m.focused - 1 + fieldsCount) % fieldsCount)
				return m, cmd
			}

		case tea.KeyEnter:
			switch m.focused {
			case btnCancel:
				return m, func() tea.Msg { return BackMsg{} }
			case btnSave:
				cmd := m.submit()
				return m, cmd
			case fieldNotes:
				// textarea handles Enter itself (insert newline)
			default:
				cmd := m.setFocus(m.focused + 1)
				return m, cmd
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

	if m.focused == fieldNotes {
		var cmd tea.Cmd
		m.notesArea, cmd = m.notesArea.Update(msg)
		return m, cmd
	}

	if m.focused < fieldsCount-2 { // not buttons
		idx := inputIdx(m.focused)
		if idx >= 0 && idx < len(m.inputs) {
			var cmd tea.Cmd
			m.inputs[idx], cmd = m.inputs[idx].Update(msg)
			return m, cmd
		}
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

	fields := m.inputs[0].View() + "\n" +
		m.inputs[1].View() + "\n" +
		m.inputs[2].View() + "\n" +
		m.inputs[3].View() + "\n" +
		m.notesArea.View() + "\n\n" +
		" " + sep + "\n" +
		m.inputs[4].View() + "\n" +
		m.inputs[5].View()

	return styles.DocStyle.Render(fmt.Sprintf("%s\n\n %s\n\n %s\n\n %s\n\n %s",
		title,
		fields,
		buttons,
		hint,
		errStr,
	))
}

func (m *Model) submit() tea.Cmd {
	title := m.inputs[inputIdx(fieldTitle)].Value()
	if title == "" {
		return func() tea.Msg {
			return ErrMsg{Err: fmt.Errorf("title is required")}
		}
	}

	dto := vault.EntryDTO{
		Title:         title,
		Username:      strPtr(m.inputs[inputIdx(fieldUsername)].Value()),
		Password:      strPtr(m.inputs[inputIdx(fieldPassword)].Value()),
		URL:           strPtr(m.inputs[inputIdx(fieldURL)].Value()),
		Notes:         strPtr(m.notesArea.Value()),
		TOTPAlgorithm: "SHA1",
		TOTPDigits:    6,
		TOTPPeriod:    30,
	}

	if s := m.inputs[inputIdx(fieldTOTPSecret)].Value(); s != "" {
		dto.TOTPSecret = &s
	}
	if s := m.inputs[inputIdx(fieldTOTPIssuer)].Value(); s != "" {
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
