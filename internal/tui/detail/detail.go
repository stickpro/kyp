package detail

import (
	"fmt"
	"strings"
	"time"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/stickpro/kyp/internal/totp"
	tuistyles "github.com/stickpro/kyp/internal/tui/styles"
	"github.com/stickpro/kyp/internal/vault"
)

type BackMsg struct{}
type EditMsg struct {
	Entry vault.EntryDTO
}

type tickMsg time.Time
type copiedMsg string

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func clearCopied() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return copiedMsg("")
	})
}

type Model struct {
	entry        vault.EntryDTO
	showPassword bool
	now          time.Time
	copied       string
}

func New(entry vault.EntryDTO) Model {
	return Model{
		entry: entry,
		now:   time.Now(),
	}
}

func (m *Model) Init() tea.Cmd {
	if m.entry.TOTPSecret != nil {
		return tick()
	}
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q", "esc", "backspace":
			return m, func() tea.Msg { return BackMsg{} }
		case "e":
			return m, func() tea.Msg { return EditMsg{Entry: m.entry} }
		case "p", " ":
			m.showPassword = !m.showPassword
		case "u":
			return m, m.copyField("Username", m.entry.Username)
		case "c":
			return m, m.copyField("Password", m.entry.Password)
		case "t":
			if m.entry.TOTPSecret != nil {
				return m, m.copyTOTP()
			}
		}

	case copiedMsg:
		m.copied = string(msg)

	case tickMsg:
		m.now = time.Time(msg)
		if m.entry.TOTPSecret != nil {
			return m, tick()
		}
	}
	return m, nil
}

func (m *Model) copyField(name string, val *string) tea.Cmd {
	if val == nil || *val == "" {
		return nil
	}
	v := *val
	return func() tea.Msg {
		_ = clipboard.WriteAll(v)
		return copiedMsg(name)
	}
}

func (m *Model) copyTOTP() tea.Cmd {
	secret := *m.entry.TOTPSecret
	digits := int(m.entry.TOTPDigits)
	period := int(m.entry.TOTPPeriod)
	if digits == 0 {
		digits = 6
	}
	if period == 0 {
		period = 30
	}
	now := m.now
	return func() tea.Msg {
		code, err := totp.Code(secret, now, digits, period)
		if err != nil {
			return nil
		}
		_ = clipboard.WriteAll(code)
		return copiedMsg("TOTP")
	}
}

func (m *Model) View() string {
	var b strings.Builder

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")).
		Background(lipgloss.Color("99")).
		Padding(0, 1).
		Render(m.entry.Title)

	b.WriteString(title + "\n\n")

	b.WriteString(copyableField("Username", strVal(m.entry.Username, "—"), "u", m.copied == "Username"))
	b.WriteString(copyableField("Password", passwordVal(m.entry.Password, m.showPassword), "c", m.copied == "Password"))
	b.WriteString(field("URL", strVal(m.entry.URL, "—")))
	b.WriteString(field("Notes", strVal(m.entry.Notes, "—")))

	if m.entry.TOTPSecret != nil {
		b.WriteString("\n")
		b.WriteString(m.totpView())
	}

	b.WriteString("\n")

	hints := []string{
		"u: copy login",
		"c: copy password",
	}
	if m.entry.TOTPSecret != nil {
		hints = append(hints, "t: copy TOTP")
	}
	hints = append(hints, "p: show/hide password", "e: edit", "esc: back")

	hint := tuistyles.HintStyle.Render(strings.Join(hints, " • "))
	b.WriteString(hint)

	if m.copied != "" {
		notice := tuistyles.OkStyle.Render(fmt.Sprintf("  ✓ %s copied", m.copied))
		b.WriteString(notice)
	}

	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}

func (m *Model) totpView() string {
	secret := *m.entry.TOTPSecret
	digits := int(m.entry.TOTPDigits)
	period := int(m.entry.TOTPPeriod)
	if digits == 0 {
		digits = 6
	}
	if period == 0 {
		period = 30
	}

	code, err := totp.Code(secret, m.now, digits, period)
	remaining := totp.Remaining(period, m.now)

	var codeStr string
	if err != nil {
		codeStr = tuistyles.ErrStyle.Render("invalid secret")
	} else {
		color := "82"
		if remaining <= 5 {
			color = "196"
		} else if remaining <= 10 {
			color = "214"
		}
		codeStr = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(color)).Render(formatCode(code))
	}

	timer := tuistyles.HintStyle.Render(fmt.Sprintf("(%ds)", remaining))

	issuer := ""
	if m.entry.TOTPIssuer != nil {
		issuer = " · " + *m.entry.TOTPIssuer
	}

	copied := ""
	if m.copied == "TOTP" {
		copied = " " + lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("[t: copied]")
	} else {
		copied = " " + tuistyles.HintStyle.Render("[t: copy]")
	}

	label := labelStyle.Render("TOTP" + issuer + ":")
	return label + " " + codeStr + " " + timer + copied + "\n"
}

func formatCode(code string) string {
	if len(code) == 6 {
		return code[:3] + " " + code[3:]
	}
	return code
}

var (
	labelStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "238", Dark: "246"}).Width(16)
	valueStyle  = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "232", Dark: "255"})
	copiedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	copyStyle   = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "241", Dark: "243"})
)

func field(label, value string) string {
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value) + "\n"
}

func copyableField(label, value, key string, wasCopied bool) string {
	hint := copyStyle.Render(fmt.Sprintf("[%s: copy]", key))
	if wasCopied {
		hint = copiedStyle.Render(fmt.Sprintf("[%s: copied]", key))
	}
	return labelStyle.Render(label+":") + " " + valueStyle.Render(value) + "  " + hint + "\n"
}

func strVal(s *string, fallback string) string {
	if s == nil || *s == "" {
		return fallback
	}
	return *s
}

func passwordVal(s *string, show bool) string {
	if s == nil || *s == "" {
		return "—"
	}
	if show {
		return *s
	}
	return strings.Repeat("•", len(*s))
}
