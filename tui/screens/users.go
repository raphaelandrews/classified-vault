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
