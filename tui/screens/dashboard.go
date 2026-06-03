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
	header := fmt.Sprintf("🔒 CLASSIFIED VAULT    User: %s  [%s] [%s]",
		styles.SuccessStyle.Render(m.user.Username),
		styles.ClearanceBadge(m.user.Clearance.String()),
		styles.DocMeta.Render(string(m.user.Role)),
	)

	var sb strings.Builder
	sb.WriteString(header + "\n\n")

	sb.WriteString(styles.BorderStyle.Render(
		styles.DocTitle.Render(fmt.Sprintf("📁 DOCUMENTS (%d accessible)", m.docCount))+"\n"+
			styles.DocPrompt.Render("[D]")+" List Documents\n"+
			styles.DocPrompt.Render("[A]")+" New Document",
	) + "\n\n")

	if m.user.Role == domain.RoleAdmin {
		sb.WriteString(styles.BorderStyle.Render(
			styles.DocTitle.Render("⚙ ADMINISTRATION")+"\n"+
				styles.DocPrompt.Render("[U]")+" Manage Users\n"+
				styles.DocPrompt.Render("[A]")+" Audit Log",
		) + "\n\n")
	} else if m.user.Role == domain.RoleAnalyst {
		sb.WriteString(styles.BorderStyle.Render(
			styles.DocTitle.Render("⚙ ANALYST")+"\n"+
				styles.DocPrompt.Render("[A]")+" Create Documents",
		) + "\n\n")
	}

	main := lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, sb.String())
	footer := styles.StatusBarStyle.Width(m.width).Render("[d] Documents  [a] New Doc  [u] Users  [l] Audit  [q] Logout")

	return main + "\n" + footer
}
