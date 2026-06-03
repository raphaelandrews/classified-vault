package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/internal/domain"
	"classified-vault/tui/styles"
)

func (m *UsersModel) updateAddForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.adding = false
		m.err = ""
		return m, nil
	case "enter":
		switch m.addStep {
		case 0:
			if m.addUser == "" {
				m.err = "Username required"
				return m, nil
			}
			m.addStep = 1
		case 1:
			m.addStep = 2
		case 2:
			m.addStep = 3
		case 3:
			roles := []domain.Role{domain.RoleIntern, domain.RoleViewer, domain.RoleAnalyst, domain.RoleAdmin}
			role := roles[m.addRole]
			return m, func() tea.Msg {
				_, err := m.apiClient.CreateUser(m.addUser, m.addEmail, m.addPass, role)
				if err != nil {
					return err
				}
				return fmt.Errorf("user created: %s", m.addUser)
			}
		}
		m.err = ""
	case "backspace":
		if m.addStep > 0 {
			m.addStep--
		} else {
			m.adding = false
		}
	case "1", "2", "3", "4":
		if m.addStep == 3 {
			for r := 0; r < 4; r++ {
				if msg.String() == fmt.Sprintf("%d", r+1) {
					m.addRole = r
				}
			}
		}
	default:
		if len(msg.Runes) == 1 {
			switch m.addStep {
			case 0:
				m.addUser += msg.String()
			case 1:
				m.addEmail += msg.String()
			case 2:
				m.addPass += msg.String()
			}
		}
	}
	return m, nil
}

func (m UsersModel) viewAddForm() string {
	var sb strings.Builder
	sb.WriteString(styles.DocTitle.Render("➕ Add User") + "\n\n")

	switch m.addStep {
	case 0:
		sb.WriteString(styles.DocPrompt.Render("Username: ") + m.addUser + "_")
	case 1:
		sb.WriteString(styles.DocPrompt.Render("Username: ") + styles.DocMeta.Render(m.addUser) + "\n\n")
		sb.WriteString(styles.DocPrompt.Render("Email: ") + m.addEmail + "_")
	case 2:
		sb.WriteString(styles.DocPrompt.Render("Username: ") + styles.DocMeta.Render(m.addUser) + "\n")
		sb.WriteString(styles.DocPrompt.Render("Email: ") + styles.DocMeta.Render(m.addEmail) + "\n\n")
		sb.WriteString(styles.DocPrompt.Render("Password: ") + strings.Repeat("•", len(m.addPass)) + "_")
	case 3:
		sb.WriteString(styles.DocPrompt.Render("Username: ") + styles.DocMeta.Render(m.addUser) + "\n")
		sb.WriteString(styles.DocPrompt.Render("Email: ") + styles.DocMeta.Render(m.addEmail) + "\n\n")
		sb.WriteString(styles.DocPrompt.Render("Role:\n"))
		roles := []domain.Role{domain.RoleIntern, domain.RoleViewer, domain.RoleAnalyst, domain.RoleAdmin}
		for i, r := range roles {
			marker := " "
			if i == m.addRole {
				marker = "▶"
			}
			sb.WriteString(fmt.Sprintf("  %s %s  [%d]\n", marker, r, i+1))
		}
	}

	if m.err != "" {
		sb.WriteString("\n" + styles.ErrorStyle.Render(m.err))
	}
	sb.WriteString(styles.DocMeta.Render("\n\n[Enter] Continue  [Backspace] Back  [Esc] Cancel"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.BorderStyle.Render(sb.String()),
	)
}
