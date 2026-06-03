package styles

import "github.com/charmbracelet/lipgloss"

var (
	Primary    = lipgloss.Color("#a9b665")
	Accent     = lipgloss.Color("#d8a657")
	Success    = lipgloss.Color("#a9b665")
	Error      = lipgloss.Color("#ea6962")
	Warning    = lipgloss.Color("#d8a657")
	Selected   = lipgloss.Color("#e78a4e")
	Foreground = lipgloss.Color("#d4be98")
	Dimmed     = lipgloss.Color("#928374")
	BorderCol  = lipgloss.Color("#504945")
	RowEven    = lipgloss.Color("#504945")
	RowOdd     = lipgloss.Color("#3c3836")
	DarkText   = lipgloss.Color("#1d2021")
)

var (
	DocTitle  = lipgloss.NewStyle().Bold(true).Foreground(Foreground)
	DocMeta   = lipgloss.NewStyle().Foreground(Dimmed)
	DocPrompt = lipgloss.NewStyle().Foreground(Accent).Bold(true)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(DarkText).
			Background(Primary).
			Padding(0, 1)

	BorderStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(BorderCol).
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
	case "PUBLIC NOTICE":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")).Render("[ PUBLIC NOTICE  ]")
	case "COUNCIL EYES ONLY":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a9b665")).Render("[ COUNCIL EYES   ]")
	case "GUILD BUSINESS":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d8a657")).Render("[ GUILD BUSINESS ]")
	case "CORPORATE ACCESS":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#e78a4e")).Render("[ CORPORATE ACC  ]")
	case "ARCANE KNOWLEDGE":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d3869b")).Render("[ ARCANE KNOWLEDGE ]")
	case "JUNIMO SCRIPT":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ea6962")).Render("[ JUNIMO SCRIPT  ]")
	default:
		return lipgloss.NewStyle().Foreground(Dimmed).Render("[ UNKNOWN        ]")
	}
}

func FactionBadge(faction string) string {
	switch faction {
	case "Mayor's Office":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ea6962")).Render("[ Mayor's Office ]")
	case "Wizard's Tower":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d3869b")).Render("[ Wizard's Tower ]")
	case "Joja Corp":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#e78a4e")).Render("[ Joja Corp      ]")
	case "Adventurer's Guild":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d8a657")).Render("[ Adventurer's G ]")
	case "Harvey's Clinic":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#83a598")).Render("[ Harvey's Clinic ]")
	case "Community Center":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a9b665")).Render("[ Community Ctr  ]")
	case "Carpenter's Shop":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#7daea3")).Render("[ Carpenter's    ]")
	case "Museum":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")).Render("[ Museum         ]")
	case "Bulletin Board":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#bdae93")).Render("[ Bulletin Board ]")
	case "Mr. Qi's Office":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#83a598")).Render("[ Mr. Qi's Off. ]")
	case "Pier & Docks":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#7daea3")).Render("[ Pier & Docks  ]")
	default:
		return lipgloss.NewStyle().Foreground(Dimmed).Render("[ " + faction + " ]")
	}
}

func SuccessBadge(text string) string {
	return SuccessStyle.Render("[OK] " + text)
}

func ErrorBadge(text string) string {
	return ErrorStyle.Render("[FAIL] " + text)
}
