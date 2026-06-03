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

type AuditLogModel struct {
	apiClient *client.APIClient
	logs      []*domain.AuditLog
	err       string
	width     int
	height    int
}

func NewAuditLogModel(api *client.APIClient) AuditLogModel {
	return AuditLogModel{
		apiClient: api,
	}
}

func (m AuditLogModel) Init() tea.Cmd {
	return m.loadLogs
}

func (m AuditLogModel) loadLogs() tea.Msg {
	logs, err := m.apiClient.ListAuditLogs()
	if err != nil {
		return fmt.Errorf("failed to load audit logs: %w", err)
	}
	return logs
}

func (m AuditLogModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case []*domain.AuditLog:
		m.logs = msg
		m.err = ""
		return m, nil
	case error:
		m.err = msg.Error()
		return m, nil
	case tea.KeyMsg:
		switch strings.ToUpper(msg.String()) {
		case "R":
			return m, m.loadLogs
		case "Q",
			"ESC":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDashboard} }
		case "CTRL+C":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AuditLogModel) View() string {
	var sb strings.Builder
	sb.WriteString(styles.DocTitle.Render("📋 Audit Log") + "\n\n")

	if m.err != "" {
		sb.WriteString(styles.ErrorStyle.Render(m.err) + "\n")
	}

	for _, entry := range m.logs {
		status := "✅"
		if !entry.Success {
			status = "🚫"
		}
		time := entry.Timestamp.Format("15:04:05")
		line := fmt.Sprintf("%s %s  %-18s %-15s -> %-20s",
			status,
			styles.DocMeta.Render(time),
			styles.DocPrompt.Render(string(entry.Action)),
			entry.Username,
			entry.Resource,
		)
		if len(entry.Details) > 0 {
			line += " " + styles.DocMeta.Render("("+entry.Details+")")
		}
		sb.WriteString(line + "\n")
	}

	if len(m.logs) == 0 {
		sb.WriteString(styles.DocMeta.Render("  No audit entries.\n"))
	}

	sb.WriteString(styles.DocMeta.Render("\n[R] Refresh  [Q] Back"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.BorderStyle.Render(sb.String()),
	)
}
