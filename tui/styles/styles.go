package styles

import "github.com/charmbracelet/lipgloss"

var (
	Primary    = lipgloss.Color("#A550DF")
	Accent     = lipgloss.Color("#C084FC")
	Success    = lipgloss.Color("#34D399")
	Error      = lipgloss.Color("#FB7185")
	Warning    = lipgloss.Color("#FBBF24")
	Selected   = lipgloss.Color("#FDE68A")
	Foreground = lipgloss.Color("#F1F5F9")
	Dimmed     = lipgloss.Color("#94A3B8")
	BorderCol  = lipgloss.Color("#334155")
	RowEven    = lipgloss.Color("#E2E8F0")
	RowOdd     = lipgloss.Color("#CBD5E1")
	DarkText   = lipgloss.Color("#1F1C23")
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
			Foreground(DarkText).
			Background(Selected).
			Bold(true)

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
			Padding(0, 1)
)

func ClearanceBadge(level string) string {
	switch level {
	case "PUBLIC":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render("[ PUBLIC     ]")
	case "RESTRICTED":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#60A5FA")).Render("[ RESTRICTED ]")
	case "CONFIDENTIAL":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Render("[ CONFIDENTIAL]")
	case "SECRET":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#FB923C")).Render("[ SECRET     ]")
	case "TOP SECRET":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#F87171")).Render("[ TOP SECRET ]")
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
