package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
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
	spinner   spinner.Model
	err       string
	loading   bool
	apiClient *client.APIClient
	width     int
	height    int

	mode     int
	deptIdx  int
	focusIdx int
}

func NewLoginModel(api *client.APIClient) LoginModel {
	u := textinput.New()
	u.Placeholder = "lewis"
	u.Focus()
	u.CharLimit = 32
	u.Width = 30
	u.PromptStyle = lipgloss.NewStyle().Foreground(styles.Dimmed)
	u.TextStyle = lipgloss.NewStyle().Foreground(styles.Foreground)

	p := textinput.New()
	p.Placeholder = "••••••••"
	p.EchoMode = textinput.EchoPassword
	p.EchoCharacter = '•'
	p.CharLimit = 64
	p.Width = 30
	p.PromptStyle = lipgloss.NewStyle().Foreground(styles.Dimmed)
	p.TextStyle = lipgloss.NewStyle().Foreground(styles.Foreground)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Accent)

	return LoginModel{
		username:  u,
		password:  p,
		spinner:   s,
		apiClient: api,
	}
}

func (m *LoginModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *LoginModel) focusField(idx int) {
	m.focusIdx = idx
	m.username.Blur()
	m.password.Blur()
	switch idx {
	case 0:
		m.username.Focus()
	case 1:
		m.password.Focus()
	}
}

func (m *LoginModel) doRegister() tea.Msg {
	dept := ""
	if m.deptIdx > 0 {
		dept = string(registerDepts[m.deptIdx-1])
	}
	user, err := m.apiClient.Register(m.username.Value(), m.password.Value(), dept)
	if err != nil {
		return err
	}
	return user
}

func (m *LoginModel) doLogin() tea.Msg {
	resp, err := m.apiClient.Login(m.username.Value(), m.password.Value())
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	return LoginMsg{User: &resp.User, Token: resp.Token}
}

var registerDepts = []domain.Department{
	domain.DepartmentMayorsOffice,
	domain.DepartmentWizardsTower,
	domain.DepartmentJojaCorp,
	domain.DepartmentAdventurersGuild,
	domain.DepartmentHarveysClinic,
	domain.DepartmentCommunityCenter,
	domain.DepartmentCarpentersShop,
	domain.DepartmentMuseum,
	domain.DepartmentBulletinBoard,
	domain.DepartmentQisOffice,
	domain.DepartmentPierDocks,
}

func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case *domain.User:
		m.err = ""
		m.loading = false
		m.mode = 0
		m.deptIdx = 0
		m.focusIdx = 0
		m.focusField(0)
		m.username.SetValue("")
		m.password.SetValue("")
		return m, nil

	case LoginMsg:
		return m, func() tea.Msg { return msg }

	case error:
		m.err = msg.Error()
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "tab":
			if m.mode == 0 {
				if m.username.Focused() {
					m.focusField(1)
				} else {
					m.focusField(0)
				}
			} else {
				m.focusField((m.focusIdx + 1) % 3)
			}
			return m, nil
		case "enter":
			if m.mode == 0 {
				if m.username.Value() == "" || m.password.Value() == "" {
					m.err = "Username and password required"
					return m, nil
				}
				m.err = ""
				m.loading = true
				return m, m.doLogin
			} else if m.mode == 1 {
				if m.focusIdx < 2 {
					m.focusField(m.focusIdx + 1)
					m.err = ""
				} else {
					if m.username.Value() == "" || m.password.Value() == "" {
						m.err = "Username and password required"
						return m, nil
					}
					m.err = ""
					m.loading = true
					return m, m.doRegister
				}
			}
		case "s":
			if m.mode == 0 {
				m.mode = 1
				m.focusIdx = 0
				m.focusField(0)
				m.deptIdx = 0
				m.err = ""
				m.username.Reset()
				m.password.Reset()
			} else {
				m.mode = 0
				m.focusIdx = 0
				m.focusField(0)
				m.err = ""
				m.username.Reset()
				m.password.Reset()
			}
			return m, nil
		case "up", "k":
			if m.mode == 1 && m.focusIdx == 2 {
				if m.deptIdx > 0 {
					m.deptIdx--
				}
			} else if m.mode == 0 {
				m.focusField(0)
			}
		case "down", "j":
			if m.mode == 1 && m.focusIdx == 2 {
				if m.deptIdx < len(registerDepts) {
					m.deptIdx++
				}
			} else if m.mode == 0 {
				m.focusField(1)
			}
		}
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
	var body strings.Builder

	emoji := styles.TitleStyle.Render("🦩")
	title := styles.TitleStyle.Render("PELICAN TOWN ARCHIVES")
	subtitle := styles.DocMeta.Render("Secure Record Management System")

	inputStyle := lipgloss.NewStyle().Width(30)
	rowStyle := lipgloss.NewStyle().Width(44)

	if m.mode == 0 {
		body.WriteString(rowStyle.Render("  "+styles.DocPrompt.Render("Username:")+" "+inputStyle.Render(m.username.View())) + "\n")
		body.WriteString(rowStyle.Render("  "+styles.DocPrompt.Render("Password:")+" "+inputStyle.Render(m.password.View())) + "\n")
		body.WriteString("\n")

		if m.err != "" {
			body.WriteString(styles.ErrorStyle.Render(m.err) + "\n")
		} else if m.loading {
			body.WriteString(m.spinner.View() + " Authenticating...\n")
		} else {
			body.WriteString("\n")
		}

		body.WriteString(styles.DocMeta.Render("\n  [Tab] Switch field  [Enter] Sign In  [S] Sign Up  [Esc] Quit"))
	} else {
		marker0, marker1, marker2 := "  ", "  ", "  "
		if m.focusIdx == 0 {
			marker0 = "▶ "
		}
		if m.focusIdx == 1 {
			marker1 = "▶ "
		}
		if m.focusIdx == 2 {
			marker2 = "▶ "
		}

		body.WriteString(rowStyle.Render(marker0+styles.DocPrompt.Render("Username:")+" "+inputStyle.Render(m.username.View())) + "\n")
		body.WriteString(rowStyle.Render(marker1+styles.DocPrompt.Render("Password:")+" "+inputStyle.Render(m.password.View())) + "\n")
		body.WriteString("\n")
		body.WriteString(rowStyle.Render(marker2+styles.DocPrompt.Render("Department (optional):")) + "\n")

		for i := -1; i < len(registerDepts); i++ {
			sel := false
			label := "  No department"
			if i >= 0 {
				sel = m.deptIdx == i+1
				label = "  " + string(registerDepts[i])
			} else {
				sel = m.deptIdx == 0
			}
			mark := "  "
			if sel {
				mark = "▶ "
			}
			body.WriteString(rowStyle.Render(mark+label) + "\n")
		}
		body.WriteString("\n")

		if m.err != "" {
			body.WriteString(styles.ErrorStyle.Render(m.err) + "\n")
		} else if m.loading {
			body.WriteString(m.spinner.View() + " Registering...\n")
		} else {
			body.WriteString("\n")
		}

		body.WriteString(styles.DocMeta.Render("\n  [Tab] Next  [↕] Select  [Enter] Register  [S] Sign In  [Esc] Quit"))
	}

	header := lipgloss.JoinVertical(lipgloss.Center, emoji, title, subtitle)

	modeTitle := ""
	if m.mode == 0 {
		modeTitle = styles.DocTitle.Render("\n Sign In \n")
	} else {
		modeTitle = styles.DocTitle.Render("\n Sign Up \n")
	}
	header = lipgloss.JoinVertical(lipgloss.Center, header, modeTitle)

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			header,
			body.String(),
		),
	)
}
