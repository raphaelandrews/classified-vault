package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"classified-vault/internal/domain"
	"classified-vault/tui/client"
	"classified-vault/tui/styles"
)

type AuditLogModel struct {
	apiClient *client.APIClient
	logs      []*domain.AuditLog
	paginator paginator.Model
	err       string
	width     int
	height    int
}

func NewAuditLogModel(api *client.APIClient) AuditLogModel {
	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 15
	p.ActiveDot = lipgloss.NewStyle().Foreground(styles.Primary).Render("●")
	p.InactiveDot = lipgloss.NewStyle().Foreground(styles.Dimmed).Render("○")

	return AuditLogModel{
		apiClient: api,
		paginator: p,
	}
}

func (m *AuditLogModel) Init() tea.Cmd {
	return m.loadLogs
}

func (m *AuditLogModel) loadLogs() tea.Msg {
	logs, err := m.apiClient.ListAuditLogs()
	if err != nil {
		return fmt.Errorf("failed to load town ledger: %w", err)
	}
	return logs
}

func (m *AuditLogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case []*domain.AuditLog:
		m.logs = msg
		m.paginator.SetTotalPages(len(m.logs))
		m.err = ""
		return m, nil
	case error:
		m.err = msg.Error()
		return m, nil
	case tea.KeyMsg:
		switch strings.ToUpper(msg.String()) {
		case "LEFT", "H":
			m.paginator.PrevPage()
		case "RIGHT", "L":
			m.paginator.NextPage()
		case "R":
			return m, m.loadLogs
		case "Q":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDashboard} }
		case "CTRL+C":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AuditLogModel) View() string {
	var sb strings.Builder
	sb.WriteString(styles.DocTitle.Render("★ Town Ledger") + "\n\n")

	if m.err != "" {
		sb.WriteString(styles.ErrorStyle.Render(m.err) + "\n")
	}

	if len(m.logs) == 0 {
		sb.WriteString(styles.DocMeta.Render("  No ledger entries.\n"))
	} else {
		start := m.paginator.Page * m.paginator.PerPage
		end := min(start+m.paginator.PerPage, len(m.logs))
		page := m.logs[start:end]

		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(styles.BorderCol)).
			Width(m.width-8).
			StyleFunc(func(row, col int) lipgloss.Style {
				base := lipgloss.NewStyle().Padding(0, 1)
				if row == table.HeaderRow {
					return base.Foreground(styles.Foreground).Bold(true)
				}
				if row%2 == 0 {
					return base.Foreground(styles.Foreground).Background(styles.RowEven)
				}
				return base.Foreground(styles.Foreground).Background(styles.RowOdd)
			}).
			Headers("", "TIME", "ACTION", "USER", "RESOURCE")

		for _, entry := range page {
			status := "★"
			if !entry.Success {
				status = "✗"
			}
			time := entry.Timestamp.Format("15:04:05")
			details := entry.Resource
			if len(entry.Details) > 0 {
				details += " (" + entry.Details + ")"
			}
			t.Row(status, time, string(entry.Action), entry.Username, details)
		}

		sb.WriteString(t.Render())

		if m.paginator.TotalPages > 1 {
			sb.WriteString("\n" + m.paginator.View())
		}
	}

	content := styles.BorderStyle.Render(sb.String())
	footer := styles.StatusBarStyle.Width(m.width).Render("[←/→] Page  [r] Refresh  [q] Back")

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}
