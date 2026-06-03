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
	title       string
	department     string
	userCle     domain.ClearanceLevel
	requiredCle domain.ClearanceLevel
	width       int
	height      int
}

func NewAccessDeniedModel(msg DocAccessDeniedMsg) AccessDeniedModel {
	return AccessDeniedModel{
		title:       msg.Title,
		department:     msg.Department,
		userCle:     msg.UserCle,
		requiredCle: msg.RequiredCle,
	}
}

func (m *AccessDeniedModel) Init() tea.Cmd { return nil }

func (m *AccessDeniedModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch strings.ToUpper(msg.String()) {
		case "H", "Q", "ENTER":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocList} }
		case "CTRL+C":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *AccessDeniedModel) View() string {
	var sb strings.Builder

	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(styles.Error).Bold(true).Render("          SEALED") + "\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(styles.Error).Bold(true).Render("    ACCESS DENIED") + "\n\n")
	sb.WriteString(styles.DocMeta.Render("    Insufficient tier or wrong department.") + "\n\n")
	sb.WriteString(fmt.Sprintf("    Your tier:        %s\n", styles.ClearanceBadge(m.userCle.String())))
	sb.WriteString(fmt.Sprintf("    Required tier:    %s\n", styles.ClearanceBadge(m.requiredCle.String())))
	sb.WriteString(fmt.Sprintf("    Scroll department:   %s\n", styles.DepartmentBadge(m.department)))
	sb.WriteString(fmt.Sprintf("    Scroll:           %s\n", styles.DocMeta.Render(m.title)))
	sb.WriteString("\n")
	sb.WriteString(styles.DocMeta.Render("    This attempt has been recorded in the Town Ledger.") + "\n\n")
	sb.WriteString(styles.DocMeta.Render("        [h] Back"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.BorderStyle.Render(sb.String()),
	)
}
