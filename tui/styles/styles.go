package styles

import "github.com/charmbracelet/lipgloss"

var (
	Primary    = lipgloss.Color("#7C3AED")
	Accent     = lipgloss.Color("#A78BFA")
	Success    = lipgloss.Color("#10B981")
	Error      = lipgloss.Color("#EF4444")
	Warning    = lipgloss.Color("#F59E0B")
	Dimmed     = lipgloss.Color("#6B7280")
	Background = lipgloss.Color("#1F2937")
	Foreground = lipgloss.Color("#F9FAFB")
)

var (
	DocTitle  = lipgloss.NewStyle().Bold(true).Foreground(Foreground)
	DocMeta   = lipgloss.NewStyle().Foreground(Dimmed)
	DocPrompt = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Foreground).
			Background(Primary).
			Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(Primary).
			Padding(1, 2)

	SelectedStyle = lipgloss.NewStyle().
			Foreground(Foreground).
			Background(Primary)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success).
			Bold(true)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(Primary).
			Padding(0, 1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(Dimmed).
			Background(Background).
			Padding(0, 1)
)

func ClearanceBadge(level string) string {
	switch level {
	case "PUBLIC":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("[ PUBLIC     ]")
	case "RESTRICTED":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Render("[ RESTRICTED ]")
	case "CONFIDENTIAL":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render("[ CONFIDENTIAL]")
	case "SECRET":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F97316")).Render("[ SECRET     ]")
	case "TOP SECRET":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("[ TOP SECRET ]")
	default:
		return lipgloss.NewStyle().Foreground(Dimmed).Render("[ UNKNOWN    ]")
	}
}

func SuccessBadge(text string) string {
	return SuccessStyle.Render("[OK] " + text)
}

func ErrorBadge(text string) string {
	return ErrorStyle.Render("[FAIL] " + text)
}
