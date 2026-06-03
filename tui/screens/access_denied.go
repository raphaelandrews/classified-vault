package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/internal/domain"
	"classified-vault/tui/styles"
)

type AccessDeniedModel struct {
	doc         *domain.Document
	userCle     domain.ClearanceLevel
	requiredCle domain.ClearanceLevel
	width       int
	height      int
}

func NewAccessDeniedModel(msg DocAccessDeniedMsg) AccessDeniedModel {
	return AccessDeniedModel{
		doc:         msg.Doc,
		userCle:     msg.UserCle,
		requiredCle: msg.RequiredCle,
	}
}

func (m AccessDeniedModel) Init() tea.Cmd { return nil }

func (m AccessDeniedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch strings.ToUpper(msg.String()) {
		case "Q", "ENTER", "ESC", "BACKSPACE":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocList} }
		case "CTRL+C":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m AccessDeniedModel) View() string {
	var sb strings.Builder

	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true).Render("          🚫") + "\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true).Render("    ACCESS DENIED") + "\n\n")
	sb.WriteString(styles.DocMeta.Render("    Insufficient clearance level.") + "\n\n")
	sb.WriteString(fmt.Sprintf("    Your clearance:     %s\n", styles.ClearanceBadge(m.userCle.String())))
	sb.WriteString(fmt.Sprintf("    Required:           %s\n", styles.ClearanceBadge(m.requiredCle.String())))
	sb.WriteString(fmt.Sprintf("    Document:           %s\n", styles.DocMeta.Render(m.doc.Title)))
	sb.WriteString("\n")
	sb.WriteString(styles.DocMeta.Render("    This attempt has been logged.") + "\n\n")
	sb.WriteString(styles.DocMeta.Render("        [Q] Back"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.BorderStyle.Render(sb.String()),
	)
}
