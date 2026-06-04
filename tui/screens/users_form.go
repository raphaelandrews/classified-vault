package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/tui/styles"
)

func (m *UsersModel) focusAddField(idx int) {
	m.addFocus = idx
	m.addUserInput.Blur()
	m.addEmailInput.Blur()
	m.addPassInput.Blur()
	switch idx {
	case 0:
		m.addUserInput.Focus()
	case 1:
		m.addEmailInput.Focus()
	case 2:
		m.addPassInput.Focus()
	}
}

func (m *UsersModel) updateAddForm(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.adding = false
			m.err = ""
			return m, nil
		case "tab":
			m.focusAddField((m.addFocus + 1) % 5)
			return m, nil
		case "shift+tab":
			m.focusAddField((m.addFocus + 4) % 5)
			return m, nil
		case "up", "k":
			if m.addFocus == 3 {
				if m.addRoleSel > 0 {
					m.addRoleSel--
				}
			} else if m.addFocus == 4 {
				if m.addDeptSel > 0 {
					m.addDeptSel--
				}
			}
		case "down", "j":
			if m.addFocus == 3 {
				if m.addRoleSel < len(addRoles)-1 {
					m.addRoleSel++
				}
			} else if m.addFocus == 4 {
				if m.addDeptSel < len(addDepts)-1 {
					m.addDeptSel++
				}
			}
		case "enter":
			if m.addFocus < 4 {
				m.focusAddField(m.addFocus + 1)
				m.err = ""
			} else {
				m.err = ""
				role := addRoles[m.addRoleSel]
				department := addDepts[m.addDeptSel]
				return m, func() tea.Msg {
					_, err := m.apiClient.CreateUser(m.addUserInput.Value(), m.addEmailInput.Value(), m.addPassInput.Value(), role, department)
					if err != nil {
						return err
					}
					return fmt.Errorf("villager registered: %s", m.addUserInput.Value())
				}
			}
		case "backspace":
			if m.addFocus > 0 && m.addFocus <= 4 {
				if m.addFocus == 2 && m.addPassInput.Value() == "" {
					m.focusAddField(1)
				} else if m.addFocus == 1 && m.addEmailInput.Value() == "" {
					m.focusAddField(0)
				} else if m.addFocus == 4 {
					m.focusAddField(2)
				}
			}
		}
	}

	if m.addFocus == 0 {
		var cmd tea.Cmd
		m.addUserInput, cmd = m.addUserInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.addFocus == 1 {
		var cmd tea.Cmd
		m.addEmailInput, cmd = m.addEmailInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.addFocus == 2 {
		var cmd tea.Cmd
		m.addPassInput, cmd = m.addPassInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *UsersModel) viewAddForm() string {
	var sb strings.Builder
	sb.WriteString(styles.DocTitle.Render("★ Register New Villager") + "\n\n")

	userLabel := styles.DocPrompt.Render("Username")
	if m.addFocus == 0 {
		userLabel = styles.DocPrompt.Render("▶ Username")
	}
	sb.WriteString(userLabel + "\n" + m.addUserInput.View() + "\n\n")

	emailLabel := styles.DocPrompt.Render("Email")
	if m.addFocus == 1 {
		emailLabel = styles.DocPrompt.Render("▶ Email")
	}
	sb.WriteString(emailLabel + "\n" + m.addEmailInput.View() + "\n\n")

	passLabel := styles.DocPrompt.Render("Password")
	if m.addFocus == 2 {
		passLabel = styles.DocPrompt.Render("▶ Password")
	}
	sb.WriteString(passLabel + "\n" + m.addPassInput.View() + "\n\n")

	roleLabel := styles.DocPrompt.Render("Role")
	if m.addFocus == 3 {
		roleLabel = styles.DocPrompt.Render("▶ Role")
	}
	sb.WriteString(roleLabel + "\n")
	for i, r := range addRoles {
		marker := " "
		if i == m.addRoleSel {
			marker = "▶"
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", marker, r))
	}
	sb.WriteString("\n")

	deptLabel := styles.DocPrompt.Render("Department")
	if m.addFocus == 4 {
		deptLabel = styles.DocPrompt.Render("▶ Department")
	}
	sb.WriteString(deptLabel + "\n")
	for i, d := range addDepts {
		marker := " "
		if i == m.addDeptSel {
			marker = "▶"
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", marker, d))
	}

	if m.err != "" && !strings.Contains(m.err, "villager registered:") {
		sb.WriteString("\n" + styles.ErrorStyle.Render(m.err))
	}
	sb.WriteString(styles.DocMeta.Render("\n\n[Tab] Next  [↑/↓] Select  [Enter] Register  [Esc] Cancel"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Left, lipgloss.Top,
		styles.BorderStyle.Render(sb.String()),
	)
}
