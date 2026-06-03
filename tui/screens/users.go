package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

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

func (m *UsersModel) Init() tea.Cmd {
	return m.loadUsers
}

func (m *UsersModel) loadUsers() tea.Msg {
	users, err := m.apiClient.ListUsers()
	if err != nil {
		return fmt.Errorf("failed to load users: %w", err)
	}
	return users
}

func (m *UsersModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.adding {
			return m.updateAddForm(msg)
		}
		if m.err != "" && strings.Contains(m.err, "user created:") {
			m.err = ""
			m.adding = false
			m.addUser = ""
			m.addEmail = ""
			m.addPass = ""
			return m, m.loadUsers
		}

		switch strings.ToUpper(msg.String()) {
		case "UP", "K":
			if m.cursor > 0 {
				m.cursor--
			} else if len(m.users) > 0 {
				m.cursor = len(m.users) - 1
			}
		case "DOWN", "J":
			if m.cursor < len(m.users)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "A":
			m.adding = true
			m.addStep = 0
			m.addUser = ""
			m.addEmail = ""
			m.addPass = ""
			m.addRole = 3
			return m, nil
		case "D":
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
		case "H", "Q":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDashboard} }
		case "CTRL+C":
			return m, tea.Quit
		}

		if m.adding {
			return m.updateAddForm(msg)
		}

		switch strings.ToUpper(msg.String()) {
		case "UP", "K":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = len(m.users) - 1
			}
		case "DOWN", "J":
			if m.cursor < len(m.users)-1 {
				m.cursor++
			} else {
				m.cursor = 0
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

func (m *UsersModel) View() string {
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

	if len(m.users) == 0 {
		sb.WriteString(styles.DocMeta.Render("  No users found.\n"))
	} else {
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(styles.BorderCol)).
			Width(m.width-8).
			StyleFunc(func(row, col int) lipgloss.Style {
				base := lipgloss.NewStyle().Padding(0, 1)
				switch {
				case row == table.HeaderRow:
					return base.Foreground(styles.Foreground).Bold(true)
				case row == m.cursor:
					return base.Foreground(styles.DarkText).Background(styles.Selected).Bold(true)
				case row%2 == 0:
					return base.Foreground(styles.RowEven)
				default:
					return base.Foreground(styles.RowOdd)
				}
			}).
			Headers("", "USERNAME", "ROLE", "CLEARANCE", "STATUS")

		for i, u := range m.users {
			marker := fmt.Sprintf("%d", i+1)
			if i == m.cursor {
				marker = "▶"
			}
			status := ""
			if !u.Active {
				status = "INACTIVE"
			}
			t.Row(marker, u.Username, string(u.Role), styles.ClearanceBadge(u.Clearance.String()), status)
		}

		sb.WriteString(t.Render())
	}

	content := styles.BorderStyle.Render(sb.String())
	main := lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, content)
	footer := styles.StatusBarStyle.Width(m.width).Render("[j/k] Move  [a] Add  [d] Delete  [r] Refresh  [h] Back")

	return main + "\n" + footer
}
