package themes

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name       string
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
}

var All = []Theme{
	GruvboxDark,
	CatppuccinMocha,
	Nord,
	TokyoNight,
	StardewWarm,
	VaultClassic,
}

var GruvboxDark = Theme{
	Name:       "Gruvbox Dark",
	Primary:    lipgloss.Color("#a9b665"),
	Accent:     lipgloss.Color("#d8a657"),
	Success:    lipgloss.Color("#a9b665"),
	Error:      lipgloss.Color("#ea6962"),
	Warning:    lipgloss.Color("#d8a657"),
	Selected:   lipgloss.Color("#e78a4e"),
	Foreground: lipgloss.Color("#d4be98"),
	Dimmed:     lipgloss.Color("#928374"),
	BorderCol:  lipgloss.Color("#504945"),
	RowEven:    lipgloss.Color("#504945"),
	RowOdd:     lipgloss.Color("#3c3836"),
	DarkText:   lipgloss.Color("#1d2021"),
	Bg:         lipgloss.Color("#1d2021"),
}

var CatppuccinMocha = Theme{
	Name:       "Catppuccin Mocha",
	Primary:    lipgloss.Color("#a6e3a1"),
	Accent:     lipgloss.Color("#89b4fa"),
	Success:    lipgloss.Color("#a6e3a1"),
	Error:      lipgloss.Color("#f38ba8"),
	Warning:    lipgloss.Color("#f9e2af"),
	Selected:   lipgloss.Color("#f5c2e7"),
	Foreground: lipgloss.Color("#cdd6f4"),
	Dimmed:     lipgloss.Color("#6c7086"),
	BorderCol:  lipgloss.Color("#45475a"),
	RowEven:    lipgloss.Color("#45475a"),
	RowOdd:     lipgloss.Color("#313244"),
	DarkText:   lipgloss.Color("#1e1e2e"),
	Bg:         lipgloss.Color("#1e1e2e"),
}

var Nord = Theme{
	Name:       "Nord",
	Primary:    lipgloss.Color("#88c0d0"),
	Accent:     lipgloss.Color("#81a1c1"),
	Success:    lipgloss.Color("#a3be8c"),
	Error:      lipgloss.Color("#bf616a"),
	Warning:    lipgloss.Color("#ebcb8b"),
	Selected:   lipgloss.Color("#d08770"),
	Foreground: lipgloss.Color("#d8dee9"),
	Dimmed:     lipgloss.Color("#4c566a"),
	BorderCol:  lipgloss.Color("#434c5e"),
	RowEven:    lipgloss.Color("#434c5e"),
	RowOdd:     lipgloss.Color("#3b4252"),
	DarkText:   lipgloss.Color("#2e3440"),
	Bg:         lipgloss.Color("#2e3440"),
}

var TokyoNight = Theme{
	Name:       "Tokyo Night",
	Primary:    lipgloss.Color("#7aa2f7"),
	Accent:     lipgloss.Color("#bb9af7"),
	Success:    lipgloss.Color("#9ece6a"),
	Error:      lipgloss.Color("#f7768e"),
	Warning:    lipgloss.Color("#e0af68"),
	Selected:   lipgloss.Color("#ff9e64"),
	Foreground: lipgloss.Color("#c0caf5"),
	Dimmed:     lipgloss.Color("#565f89"),
	BorderCol:  lipgloss.Color("#3b4261"),
	RowEven:    lipgloss.Color("#3b4261"),
	RowOdd:     lipgloss.Color("#292e42"),
	DarkText:   lipgloss.Color("#1a1b26"),
	Bg:         lipgloss.Color("#1a1b26"),
}

var StardewWarm = Theme{
	Name:       "Stardew Warm",
	Primary:    lipgloss.Color("#6b8c42"),
	Accent:     lipgloss.Color("#e6a817"),
	Success:    lipgloss.Color("#6b8c42"),
	Error:      lipgloss.Color("#c45c3a"),
	Warning:    lipgloss.Color("#e6a817"),
	Selected:   lipgloss.Color("#f4c542"),
	Foreground: lipgloss.Color("#e8d5b7"),
	Dimmed:     lipgloss.Color("#8a7560"),
	BorderCol:  lipgloss.Color("#5a4a3a"),
	RowEven:    lipgloss.Color("#5a4a3a"),
	RowOdd:     lipgloss.Color("#3d3025"),
	DarkText:   lipgloss.Color("#1e1810"),
	Bg:         lipgloss.Color("#1e1810"),
}

var VaultClassic = Theme{
	Name:       "Vault Classic",
	Primary:    lipgloss.Color("#A550DF"),
	Accent:     lipgloss.Color("#C084FC"),
	Success:    lipgloss.Color("#34D399"),
	Error:      lipgloss.Color("#FB7185"),
	Warning:    lipgloss.Color("#FBBF24"),
	Selected:   lipgloss.Color("#FDE68A"),
	Foreground: lipgloss.Color("#F1F5F9"),
	Dimmed:     lipgloss.Color("#94A3B8"),
	BorderCol:  lipgloss.Color("#334155"),
	RowEven:    lipgloss.Color("#1E293B"),
	RowOdd:     lipgloss.Color("#0F172A"),
	DarkText:   lipgloss.Color("#1F1C23"),
	Bg:         lipgloss.Color("#0F172A"),
}
