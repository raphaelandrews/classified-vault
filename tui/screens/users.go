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

type UsersModel struct {
	apiClient *client.APIClient
	users     []*domain.User
	cursor    int
	err       string
	width     int
	height    int

	adding   bool
	addUser  string
	addEmail string
	addPass  string
	addRole  int
	addStep  int
	addDone  string
}

func NewUsersModel(api *client.APIClient) UsersModel {
	return UsersModel{
		apiClient: api,
		addRole:   3,
	}
}

func (m UsersModel) Init() tea.Cmd {
	return m.loadUsers
}

func (m UsersModel) loadUsers() tea.Msg {
	users, err := m.apiClient.ListUsers()
	if err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}
	return users
}

func (m UsersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case []*domain.User:
		m.users = msg
		m.err = ""
		return m, nil

	case error:
		m.err = msg.Error()
		m.adding = false
		return m, nil

	case tea.KeyMsg:
		if m.addDone != "" {
			m.addDone = ""
			m.adding = false
			m.addUser = ""
			m.addEmail = ""
			m.addPass = ""
			return m, m.loadUsers
		}

		if m.adding {
			return m.updateAddForm(msg)
		}

		switch strings.ToUpper(msg.String()) {
		case "UP", "K":
			if m.cursor > 0 {
				m.cursor--
			}
		case "DOWN", "J":
			if m.cursor < len(m.users)-1 {
				m.cursor++
			}
		case "N":
			m.adding = true
			m.addStep = 0
			return m, nil
		case "DELETE":
			if m.cursor < len(m.users) {
				user := m.users[m.cursor]
				return m, func() tea.Msg {
					err := m.apiClient.DeleteUser(user.ID)
					if err != nil {
						return err
					}
					return NavigateMsg{Screen: ScreenUsers}
				}
			}
		case "R":
			return m, m.loadUsers
		case "Q":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDashboard} }
		case "CTRL+C":
			return m, tea.Quit
		}
	case NavigateMsg:
		return m, m.loadUsers
	}
	return m, nil
}

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
			if m.addPass == "" {
				m.addPass = ""
			}
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

func (m UsersModel) View() string {
	var sb strings.Builder

	if m.addDone != "" {
		sb.WriteString(styles.SuccessStyle.Render(m.addDone) + "\n\n")
	}

	if m.adding {
		return m.viewAddForm()
	}

	sb.WriteString(styles.DocTitle.Render("👥 User Management") + "\n\n")

	if m.err != "" {
		sb.WriteString(styles.ErrorStyle.Render(m.err) + "\n")
	}

	for i, u := range m.users {
		cursor := " "
		if i == m.cursor {
			cursor = "▶"
		}
		active := ""
		if !u.Active {
			active = styles.DocMeta.Render(" [INACTIVE]")
		}
		line := fmt.Sprintf("%s %-20s %-12s %s%s",
			cursor, u.Username, u.Role, styles.ClearanceBadge(u.Clearance.String()), active)
		if i == m.cursor {
			line = styles.SelectedStyle.Render(line)
		}
		sb.WriteString(line + "\n")
	}

	sb.WriteString(styles.DocMeta.Render("\n[N] Add User  [Del] Delete  [R] Refresh  [Q] Back"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.BorderStyle.Render(sb.String()),
	)
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
