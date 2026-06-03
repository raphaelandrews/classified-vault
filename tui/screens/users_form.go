package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

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
			m.numBuf = ""
		case 3:
			m.addStep = 4
			m.numBuf = ""
		case 4:
			role := addRoles[m.addRole]
			department := addDepts[m.addDept]
			return m, func() tea.Msg {
				_, err := m.apiClient.CreateUser(m.addUser, m.addEmail, m.addPass, role, department)
				if err != nil {
					return err
				}
				return fmt.Errorf("villager registered: %s", m.addUser)
			}
		}
		m.err = ""
	case "backspace":
		switch m.addStep {
		case 0:
			if len(m.addUser) > 0 {
				m.addUser = m.addUser[:len(m.addUser)-1]
			} else {
				m.adding = false
			}
		case 1:
			if len(m.addEmail) > 0 {
				m.addEmail = m.addEmail[:len(m.addEmail)-1]
			} else {
				m.addStep = 0
			}
		case 2:
			if len(m.addPass) > 0 {
				m.addPass = m.addPass[:len(m.addPass)-1]
			} else {
				m.addStep = 1
			}
		case 3:
			m.addStep = 2
		case 4:
			m.addStep = 3
		}
	default:
		k := msg.String()
		if m.addStep == 3 {
			m.numBuf += k
			handleNumKey(&m.addRole, m.numBuf, len(addRoles))
		} else if m.addStep == 4 {
			m.numBuf += k
			handleNumKey(&m.addDept, m.numBuf, len(addDepts))
		} else if len(msg.Runes) == 1 {
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

func handleNumKey(selected *int, key string, count int) {
	for i := 0; i < count; i++ {
		if key == fmt.Sprintf("%d", i+1) {
			*selected = i
		}
	}
}

func (m *UsersModel) viewAddForm() string {
	var sb strings.Builder
	sb.WriteString(styles.DocTitle.Render("★ Register New Villager") + "\n\n")

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
		for i, r := range addRoles {
			marker := " "
			if i == m.addRole {
				marker = "▶"
			}
			sb.WriteString(fmt.Sprintf("  %s %s  [%d]\n", marker, r, i+1))
		}
	case 4:
		sb.WriteString(styles.DocPrompt.Render("Username: ") + styles.DocMeta.Render(m.addUser) + "\n")
		sb.WriteString(styles.DocPrompt.Render("Email: ") + styles.DocMeta.Render(m.addEmail) + "\n")
		sb.WriteString(styles.DocPrompt.Render("Role: ") + string(addRoles[m.addRole]) + "\n\n")
		sb.WriteString(styles.DocPrompt.Render("Department:\n"))
		for i, f := range addDepts {
			marker := " "
			if i == m.addDept {
				marker = "▶"
			}
			sb.WriteString(fmt.Sprintf("  %s %s  [%d]\n", marker, f, i+1))
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
