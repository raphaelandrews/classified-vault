package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/internal/domain"
	"classified-vault/tui/client"
	"classified-vault/tui/styles"
)

type DashboardModel struct {
	user      *domain.User
	apiClient *client.APIClient
	docCount  int
	stats     *client.StatsResponse
	width     int
	height    int
}

func NewDashboardModel(api *client.APIClient, user *domain.User) DashboardModel {
	return DashboardModel{
		user:      user,
		apiClient: api,
	}
}

func (m *DashboardModel) Init() tea.Cmd {
	return m.loadStats
}

func (m *DashboardModel) loadStats() tea.Msg {
	stats, err := m.apiClient.GetStats()
	if err != nil {
		return nil
	}
	return stats
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case *client.StatsResponse:
		m.stats = msg
		if msg != nil {
			m.docCount = msg.TotalScrolls
		}
		return m, nil
	case tea.KeyMsg:
		switch strings.ToUpper(msg.String()) {
		case "D":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocList} }
		case "A":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocCreate} }
		case "U":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenUsers} }
		case "L":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenAudit} }
		case "R":
			return m, m.loadStats
		case "P":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenPasswordChange} }
		case "Q", "H":
			m.apiClient.Logout()
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenLogin} }
		case "CTRL+C":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *DashboardModel) View() string {
	header := fmt.Sprintf("★ PELICAN TOWN ARCHIVES    %s  %s  %s",
		styles.SuccessStyle.Render(m.user.RoleName+" "+m.user.Username),
		styles.DepartmentBadge(string(m.user.Department)),
		styles.ClearanceBadge(m.user.Clearance.String()),
	)

	var sb strings.Builder
	sb.WriteString(header + "\n\n")

	sb.WriteString(styles.BorderStyle.Render(
		styles.DocTitle.Render(fmt.Sprintf("SCROLLS (%d accessible)", m.docCount))+"\n"+
			styles.DocPrompt.Render("[D]")+" Browse Scrolls\n"+
			styles.DocPrompt.Render("[A]")+" Scribe New Scroll",
	) + "\n\n")

	if m.user.Role == domain.RoleMayor {
		sb.WriteString(styles.BorderStyle.Render(
			styles.DocTitle.Render("★ MAYOR'S ADMINISTRATION")+"\n"+
				styles.DocPrompt.Render("[U]")+" Manage Villagers\n"+
				styles.DocPrompt.Render("[L]")+" Town Ledger",
		) + "\n\n")
	} else if m.user.Role == domain.RoleKeeper {
		sb.WriteString(styles.BorderStyle.Render(
			styles.DocTitle.Render("★ DIRECTOR")+"\n"+
				styles.DocPrompt.Render("[A]")+" Scribe New Scrolls",
		) + "\n\n")
	}

	if m.stats != nil {
		sb.WriteString(m.renderStats())
	}

	main := lipgloss.Place(m.width, m.height-1, lipgloss.Left, lipgloss.Top, sb.String())
	footer := styles.StatusBarStyle.Width(m.width).Render(fmt.Sprintf("[d] Scrolls  [a] New  [u] Villagers  [l] Ledger  [p] Password  [r] Refresh  [ctrl+t] Theme: %s  [q] Sign Out", styles.CurrentTheme.Name))

	return main + "\n" + footer
}

func (m *DashboardModel) renderStats() string {
	var sb strings.Builder
	sb.WriteString(styles.BorderStyle.Render(
		styles.DocTitle.Render("★ TOWN STATISTICS"),
	) + "\n\n")

	sb.WriteString(styles.DocMeta.Render(
		fmt.Sprintf("Total Scrolls: %d  |  Villagers: %d  |  Scribbled This Month: %d",
			m.stats.TotalScrolls, m.stats.TotalVillagers, m.stats.CreatedThisMonth),
	) + "\n\n")

	if m.stats.MostActive != "" {
		sb.WriteString(styles.DocMeta.Render(
			fmt.Sprintf("Most Active Scribe: %s (%d scrolls)",
				m.stats.MostActive, m.stats.MostActiveCount),
		) + "\n\n")
	}

	sb.WriteString(styles.DocTitle.Render("Scrolls per Tier") + "\n")
	sb.WriteString(m.renderBarChart(m.stats.TierCounts, []string{
		"JUNIMO SCRIPT", "ARCANE SEALED", "VAULT SEALED",
		"COUNCIL SEALED", "GUILD SEALED", "TOWN NOTICE",
	}) + "\n")

	sb.WriteString(styles.DocTitle.Render("Scrolls per Department") + "\n")
	topDepts := topN(m.stats.DepartmentCounts, 5)
	for _, dept := range topDepts {
		sb.WriteString(styles.DocMeta.Render(
			fmt.Sprintf("  %s: %d", dept.name, dept.count),
		) + "\n")
	}

	return sb.String()
}

func (m *DashboardModel) renderBarChart(counts map[string]int, order []string) string {
	if len(counts) == 0 {
		return styles.DocMeta.Render("  (none)") + "\n"
	}

	maxVal := 0
	for _, v := range counts {
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == 0 {
		maxVal = 1
	}

	maxBar := 30

	var sb strings.Builder
	for _, label := range order {
		count := counts[label]
		barLen := count * maxBar / maxVal
		bar := strings.Repeat("█", barLen)
		if barLen == 0 && count > 0 {
			bar = "▏"
		}
		badge := styles.ClearanceBadge(label)
		sb.WriteString(fmt.Sprintf("  %s  %s %d\n", bar, badge, count))
	}
	return sb.String()
}

type deptEntry struct {
	name  string
	count int
}

func topN(counts map[string]int, n int) []deptEntry {
	entries := make([]deptEntry, 0, len(counts))
	for k, v := range counts {
		entries = append(entries, deptEntry{k, v})
	}
	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].count > entries[i].count {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	if len(entries) > n {
		entries = entries[:n]
	}
	return entries
}
