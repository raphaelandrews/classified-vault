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
	return func() tea.Msg {
		docs, err := m.apiClient.ListDocuments()
		if err != nil {
			return nil
		}
		return len(docs)
	}
}

func (m *DashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case int:
		m.docCount = msg
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

	main := lipgloss.Place(m.width, m.height-1, lipgloss.Left, lipgloss.Top, sb.String())
	footer := styles.StatusBarStyle.Width(m.width).Render(fmt.Sprintf("[d] Scrolls  [a] New  [u] Villagers  [l] Ledger  [ctrl+t] Theme: %s  [q] Sign Out", styles.CurrentTheme.Name))

	return main + "\n" + footer
}
