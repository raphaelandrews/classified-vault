package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
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

	adding        bool
	addUserInput  textinput.Model
	addEmailInput textinput.Model
	addPassInput  textinput.Model
	addRoleSel    int
	addDeptSel    int
	addFocus      int
}

var addRoles = []domain.Role{domain.RoleAssociate, domain.RoleVillager, domain.RoleKeeper, domain.RoleMayor}
var addDepts = []domain.Department{
	domain.DepartmentMuseum,
	domain.DepartmentBulletinBoard,
	domain.DepartmentCommunityCenter,
	domain.DepartmentCarpentersShop,
	domain.DepartmentPierDocks,
	domain.DepartmentAdventurersGuild,
	domain.DepartmentHarveysClinic,
	domain.DepartmentJojaCorp,
	domain.DepartmentWizardsTower,
	domain.DepartmentQisOffice,
	domain.DepartmentMayorsOffice,
}

func NewUsersModel(api *client.APIClient) UsersModel {
	ui := textinput.New()
	ui.Placeholder = "Username"
	ui.CharLimit = 32
	ui.Width = 30
	ui.Prompt = ""

	ei := textinput.New()
	ei.Placeholder = "villager@pelican.town"
	ei.CharLimit = 64
	ei.Width = 30
	ei.Prompt = ""

	pi := textinput.New()
	pi.Placeholder = "Password"
	pi.EchoMode = textinput.EchoPassword
	pi.EchoCharacter = '•'
	pi.CharLimit = 64
	pi.Width = 30
	pi.Prompt = ""

	return UsersModel{
		apiClient:     api,
		addUserInput:  ui,
		addEmailInput: ei,
		addPassInput:  pi,
		addRoleSel:    1,
		addDeptSel:    0,
	}
}

func (m *UsersModel) Init() tea.Cmd {
	return m.loadUsers
}

func (m *UsersModel) loadUsers() tea.Msg {
	users, err := m.apiClient.ListUsers()
	if err != nil {
		return fmt.Errorf("failed to load villagers: %w", err)
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
		if m.err != "" && strings.Contains(m.err, "villager registered:") {
			m.err = ""
			m.adding = false
			m.addUserInput.Reset()
			m.addEmailInput.Reset()
			m.addPassInput.Reset()
			m.cursor = 0
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
			m.addFocus = 0
			m.addUserInput.Reset()
			m.addEmailInput.Reset()
			m.addPassInput.Reset()
			m.addRoleSel = 1
			m.addDeptSel = 0
			m.addUserInput.Focus()
			return m, nil
		case "D":
			if m.cursor < len(m.users) {
				user := m.users[m.cursor]
				id := user.ID
				name := user.Username
				return m, func() tea.Msg {
					return ConfirmPromptMsg{
						Message: fmt.Sprintf("Dismiss villager \"%s\"?\nThey will lose all access to the archives.", name),
						OnYes: func() tea.Msg {
							err := m.apiClient.DeleteUser(id)
							if err != nil {
								return err
							}
							return NavigateMsg{Screen: ScreenUsers}
						},
					}
				}
			}
		case "R":
			return m, m.loadUsers
		case "H", "Q":
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

	if m.adding {
		return m.viewAddForm()
	}

	sb.WriteString(styles.DocTitle.Render("★ Villagers of Pelican Town") + "\n\n")

	if m.err != "" {
		sb.WriteString(styles.ErrorStyle.Render(m.err) + "\n")
	}

	if len(m.users) == 0 {
		sb.WriteString(styles.DocMeta.Render("  No villagers found.\n"))
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
					return base.Foreground(styles.Foreground).Background(styles.RowEven)
				default:
					return base.Foreground(styles.Foreground).Background(styles.RowOdd)
				}
			}).
			Headers("", "USERNAME", "DEPT", "ROLE", "TIER", "STATUS")

		for i, u := range m.users {
			marker := fmt.Sprintf("%d", i+1)
			if i == m.cursor {
				marker = "▶"
			}
			status := ""
			if !u.Active {
				status = "INACTIVE"
			}
			t.Row(marker, u.Username, string(u.Department), string(u.Role), styles.ClearanceBadge(u.Clearance.String()), status)
		}

		sb.WriteString(t.Render())
	}

	content := styles.BorderStyle.Render(sb.String())
	footer := styles.StatusBarStyle.Width(m.width).Render("[j/k] Move  [a] Register  [d] Dismiss  [r] Refresh  [h] Back")

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}
