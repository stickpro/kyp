package styles

import "github.com/charmbracelet/lipgloss"

// AdaptiveColor: Light — для светлого фона, Dark — для тёмного.
var (
	ActiveStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("99")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 2)

	InactiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "238", Dark: "246"}).
			Padding(0, 2)

	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))

	// Подсказки (tab: next • esc: back …)
	HintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "241", Dark: "243"})

	// Лейблы полей (Username:, Password: …)
	LabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "238", Dark: "246"})

	// Значения полей
	ValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "232", Dark: "255"})

	ErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	OkStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
)
