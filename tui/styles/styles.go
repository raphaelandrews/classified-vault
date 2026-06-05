package styles

import (
	"github.com/charmbracelet/lipgloss"

	"classified-vault/tui/themes"
)

var CurrentTheme *themes.Theme

var (
	Primary    lipgloss.Color
	Accent     lipgloss.Color
	Success    lipgloss.Color
	Error      lipgloss.Color
	Warning    lipgloss.Color
	Selected   lipgloss.Color
	Foreground lipgloss.Color
	Dimmed     lipgloss.Color
	BorderCol  lipgloss.Color
	RowEven    lipgloss.Color
	RowOdd     lipgloss.Color
	DarkText   lipgloss.Color
	Bg         lipgloss.Color
)

var (
	DocTitle           lipgloss.Style
	DocMeta            lipgloss.Style
	DocPrompt          lipgloss.Style
	DocViewTitle       lipgloss.Style
	HeaderStyle        lipgloss.Style
	BorderStyle        lipgloss.Style
	SelectedStyle      lipgloss.Style
	ErrorStyle         lipgloss.Style
	SuccessStyle       lipgloss.Style
	TitleStyle         lipgloss.Style
	StatusBarStyle     lipgloss.Style
	ConfirmBoxStyle    lipgloss.Style
	ConfirmTitleStyle  lipgloss.Style
	ConfirmPromptStyle lipgloss.Style
)

func SetTheme(t *themes.Theme) {
	CurrentTheme = t
	Primary = t.Primary
	Accent = t.Accent
	Success = t.Success
	Error = t.Error
	Warning = t.Warning
	Selected = t.Selected
	Foreground = t.Foreground
	Dimmed = t.Dimmed
	BorderCol = t.BorderCol
	RowEven = t.RowEven
	RowOdd = t.RowOdd
	DarkText = t.DarkText
	Bg = t.Bg

	DocTitle = lipgloss.NewStyle().Bold(true).Foreground(Foreground)
	DocMeta = lipgloss.NewStyle().Foreground(Dimmed)
	DocPrompt = lipgloss.NewStyle().Foreground(Accent).Bold(true)
	DocViewTitle = lipgloss.NewStyle().Bold(true).Foreground(DarkText).Background(Primary).Padding(0, 2)

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

	ConfirmBoxStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(1, 3)

	ConfirmTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(Warning)
	ConfirmPromptStyle = lipgloss.NewStyle().Foreground(Dimmed)
}

func init() {
	SetTheme(&themes.GruvboxDark)
}

func ClearanceBadge(level string) string {
	switch level {
	case "TOWN NOTICE":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")).Render("[ TOWN NOTICE    ]")
	case "GUILD SEALED":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a9b665")).Render("[ GUILD SEALED   ]")
	case "COUNCIL SEALED":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d8a657")).Render("[ COUNCIL SEALED ]")
	case "VAULT SEALED":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#e78a4e")).Render("[ VAULT SEALED   ]")
	case "ARCANE SEALED":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d3869b")).Render("[ ARCANE SEALED  ]")
	case "JUNIMO SCRIPT":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#ea6962")).Render("[ JUNIMO SCRIPT  ]")
	default:
		return lipgloss.NewStyle().Foreground(Dimmed).Render("[ UNKNOWN        ]")
	}
}

func DepartmentBadge(department string) string {
	switch department {
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
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#83a598")).Render("[ Mr. Qi's Off.  ]")
	case "Pier & Docks":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#7daea3")).Render("[ Pier & Docks   ]")
	case "Roving Trader":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d8a657")).Render("[ Roving Trader  ]")
	default:
		return lipgloss.NewStyle().Foreground(Dimmed).Render("[ " + department + " ]")
	}
}

func SuccessBadge(text string) string {
	return SuccessStyle.Render("[OK] " + text)
}

func ErrorBadge(text string) string {
	return ErrorStyle.Render("[FAIL] " + text)
}
