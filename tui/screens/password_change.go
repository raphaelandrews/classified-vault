package screens

import (
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/tui/client"
	"classified-vault/tui/styles"
)

type PasswordChangeModel struct {
	currentInput textinput.Model
	newInput     textinput.Model
	confirmInput textinput.Model
	spinner      spinner.Model

	apiClient *client.APIClient
	width     int
	height    int

	focusIdx int
	err      string
	success  string
	loading  bool
}

func NewPasswordChangeModel(api *client.APIClient) PasswordChangeModel {
	ci := textinput.New()
	ci.Placeholder = "Current password"
	ci.EchoMode = textinput.EchoPassword
	ci.EchoCharacter = '•'
	ci.CharLimit = 64
	ci.Width = 30
	ci.PromptStyle = lipgloss.NewStyle().Foreground(styles.Dimmed)
	ci.TextStyle = lipgloss.NewStyle().Foreground(styles.Foreground)
	ci.Focus()

	ni := textinput.New()
	ni.Placeholder = "New password"
	ni.EchoMode = textinput.EchoPassword
	ni.EchoCharacter = '•'
	ni.CharLimit = 64
	ni.Width = 30
	ni.PromptStyle = lipgloss.NewStyle().Foreground(styles.Dimmed)
	ni.TextStyle = lipgloss.NewStyle().Foreground(styles.Foreground)

	cfi := textinput.New()
	cfi.Placeholder = "Confirm new password"
	cfi.EchoMode = textinput.EchoPassword
	cfi.EchoCharacter = '•'
	cfi.CharLimit = 64
	cfi.Width = 30
	cfi.PromptStyle = lipgloss.NewStyle().Foreground(styles.Dimmed)
	cfi.TextStyle = lipgloss.NewStyle().Foreground(styles.Foreground)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.Accent)

	return PasswordChangeModel{
		currentInput: ci,
		newInput:     ni,
		confirmInput: cfi,
		spinner:      sp,
		apiClient:    api,
	}
}

func (m *PasswordChangeModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *PasswordChangeModel) focusField(idx int) {
	m.focusIdx = idx
	m.currentInput.Blur()
	m.newInput.Blur()
	m.confirmInput.Blur()
	switch idx {
	case 0:
		m.currentInput.Focus()
	case 1:
		m.newInput.Focus()
	case 2:
		m.confirmInput.Focus()
	}
}

func (m *PasswordChangeModel) doChange() tea.Msg {
	err := m.apiClient.ChangePassword(m.currentInput.Value(), m.newInput.Value())
	if err != nil {
		return err
	}
	return "password changed"
}

func (m *PasswordChangeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case string:
		m.err = ""
		m.loading = false
		m.success = msg
		return m, nil

	case error:
		m.err = msg.Error()
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc", "q":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDashboard} }
		case "tab":
			m.focusField((m.focusIdx + 1) % 3)
			return m, nil
		case "shift+tab":
			m.focusField((m.focusIdx + 2) % 3)
			return m, nil
		case "enter":
			if m.focusIdx < 2 {
				m.focusField(m.focusIdx + 1)
				return m, nil
			}
			if m.currentInput.Value() == "" || m.newInput.Value() == "" || m.confirmInput.Value() == "" {
				m.err = "All fields are required"
				return m, nil
			}
			if m.newInput.Value() != m.confirmInput.Value() {
				m.err = "New passwords do not match"
				return m, nil
			}
			if string(m.newInput.Value()) == string(m.currentInput.Value()) {
				m.err = "New password must differ from current"
				return m, nil
			}
			m.err = ""
			m.success = ""
			m.loading = true
			return m, m.doChange
		}
	}

	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0)
	if m.currentInput.Focused() {
		m.currentInput, cmd = m.currentInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.newInput.Focused() {
		m.newInput, cmd = m.newInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.confirmInput.Focused() {
		m.confirmInput, cmd = m.confirmInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m *PasswordChangeModel) View() string {
	var body strings.Builder
	body.WriteString(styles.DocTitle.Render("CHANGE PASSWORD") + "\n\n")

	inputStyle := lipgloss.NewStyle().Width(30)
	rowStyle := lipgloss.NewStyle().Width(46)

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

	body.WriteString(rowStyle.Render(marker0+styles.DocPrompt.Render("Current:")+" "+inputStyle.Render(m.currentInput.View())) + "\n")
	body.WriteString(rowStyle.Render(marker1+styles.DocPrompt.Render("New:    ")+" "+inputStyle.Render(m.newInput.View())) + "\n")
	body.WriteString(rowStyle.Render(marker2+styles.DocPrompt.Render("Confirm:")+" "+inputStyle.Render(m.confirmInput.View())) + "\n")
	body.WriteString("\n")
	body.WriteString(styles.DocMeta.Render("  8+ chars, letters + numbers/symbols") + "\n\n")

	if m.success != "" {
		body.WriteString(styles.SuccessStyle.Render("*** "+m.success+" ***") + "\n\n")
	} else if m.err != "" {
		body.WriteString(styles.ErrorStyle.Render(m.err) + "\n\n")
	} else if m.loading {
		body.WriteString(m.spinner.View() + " Changing password...\n\n")
	}

	body.WriteString(styles.DocMeta.Render("[Tab] Next field  [Enter] Submit  [Esc/Q] Dashboard"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(lipgloss.Center,
			body.String(),
		),
	)
}
