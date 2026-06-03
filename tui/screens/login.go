package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/internal/domain"
	"classified-vault/tui/client"
	"classified-vault/tui/styles"
)

type LoginMsg struct {
	User  *domain.User
	Token string
}

type LoginModel struct {
	username  textinput.Model
	password  textinput.Model
	err       string
	loading   bool
	apiClient *client.APIClient
	width     int
	height    int
}

func NewLoginModel(api *client.APIClient) LoginModel {
	u := textinput.New()
	u.Placeholder = "lewis"
	u.Focus()
	u.CharLimit = 32
	u.Width = 30

	p := textinput.New()
	p.Placeholder = "••••••••"
	p.EchoMode = textinput.EchoPassword
	p.EchoCharacter = '•'
	p.CharLimit = 64
	p.Width = 30

	return LoginModel{
		username:  u,
		password:  p,
		apiClient: api,
	}
}

func (m *LoginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.username.Value() == "" || m.password.Value() == "" {
				m.err = "Username and password required"
				return m, nil
			}
			m.err = ""
			m.loading = true
			return m, func() tea.Msg {
				resp, err := m.apiClient.Login(m.username.Value(), m.password.Value())
				if err != nil {
					return fmt.Errorf("login failed: %w", err)
				}
				return LoginMsg{User: &resp.User, Token: resp.Token}
			}
		case "tab", "up", "down":
			if m.username.Focused() {
				m.username.Blur()
				m.password.Focus()
			} else {
				m.password.Blur()
				m.username.Focus()
			}
			return m, nil
		}
	case error:
		m.err = msg.Error()
		m.loading = false
		return m, nil
	}

	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)
	if m.username.Focused() {
		m.username, cmd = m.username.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.password.Focused() {
		m.password, cmd = m.password.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *LoginModel) View() string {
	title := styles.TitleStyle.Render("★ PELICAN TOWN ARCHIVES")
	subtitle := styles.DocMeta.Render("Mayor's Office · Secure Record System")

	var body strings.Builder
	body.WriteString(fmt.Sprintf("\n  %s %s\n", styles.DocPrompt.Render("Username:"), m.username.View()))
	body.WriteString(fmt.Sprintf("\n  %s %s\n", styles.DocPrompt.Render("Password:"), m.password.View()))
	body.WriteString("\n")

	if m.err != "" {
		body.WriteString(styles.ErrorStyle.Render("  "+m.err) + "\n")
	}
	if m.loading {
		body.WriteString(styles.DocMeta.Render("  Authenticating...") + "\n")
	}

	body.WriteString(styles.DocMeta.Render("\n  [Enter] Sign In  [Tab] Switch  [Esc] Quit"))

	content := lipgloss.JoinVertical(lipgloss.Center,
		title,
		subtitle,
		body.String(),
	)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.BorderStyle.Render(content),
	)
}
