package styles

import "github.com/charmbracelet/lipgloss"

var (
	ActiveStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("99")).
			Foreground(lipgloss.Color("255")).
			Padding(0, 2)

	InactiveStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("246")).
			Padding(0, 2)

	TitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	HintStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("243"))
	LabelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	ErrStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	OkStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
)
