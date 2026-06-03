package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/tui/client"
	"classified-vault/tui/screens"
	"classified-vault/tui/styles"
	"classified-vault/tui/themes"
)

type Model struct {
	screen screens.Screen

	apiClient *client.APIClient

	loginModel        *screens.LoginModel
	dashModel         *screens.DashboardModel
	docListModel      *screens.DocumentListModel
	docViewModel      *screens.DocumentViewModel
	docCreateModel    *screens.DocCreateModel
	accessDeniedModel *screens.AccessDeniedModel
	usersModel        *screens.UsersModel
	auditModel        *screens.AuditLogModel

	current  tea.Model
	width    int
	height   int
	themeIdx int

	confirm *screens.ConfirmPromptMsg
}

func New(serverURL string) *Model {
	api := client.New(serverURL)
	m := &Model{
		screen:    screens.ScreenLogin,
		apiClient: api,
	}
	lm := screens.NewLoginModel(api)
	m.loginModel = &lm
	m.current = m.loginModel
	return m
}

func (m *Model) resizeCmd() tea.Cmd {
	return func() tea.Msg { return tea.WindowSizeMsg{Width: m.width, Height: m.height} }
}

func (m *Model) Init() tea.Cmd {
	return m.current.Init()
}

func (m *Model) cycleTheme() {
	m.themeIdx = (m.themeIdx + 1) % len(themes.All)
	styles.SetTheme(&themes.All[m.themeIdx])
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		if m.confirm != nil {
			switch msg.String() {
			case "y", "Y", "enter":
				action := m.confirm
				m.confirm = nil
				return m, func() tea.Msg { return action.OnYes() }
			case "n", "N", "esc", "h", "q":
				m.confirm = nil
				return m, nil
			}
			return m, nil
		}
		if msg.String() == "ctrl+t" {
			m.cycleTheme()
			return m, nil
		}

	case screens.LoginMsg:
		dm := screens.NewDashboardModel(m.apiClient, msg.User)
		m.dashModel = &dm
		m.screen = screens.ScreenDashboard
		m.current = m.dashModel
		return m, tea.Batch(m.current.Init(), m.resizeCmd())

	case screens.NavigateMsg:
		next, cmd := m.handleNavigate(msg)
		return next, tea.Batch(cmd, m.resizeCmd())

	case screens.DocSelectedMsg:
		doc, err := m.apiClient.GetDocument(msg.DocID)
		if err != nil {
			return m, nil
		}
		dvm := screens.NewDocumentViewModel(doc, m.apiClient.User)
		m.docViewModel = &dvm
		m.screen = screens.ScreenDocView
		m.current = m.docViewModel
		return m, tea.Batch(m.current.Init(), m.resizeCmd())

	case screens.DocAccessDeniedMsg:
		adm := screens.NewAccessDeniedModel(msg)
		m.accessDeniedModel = &adm
		m.screen = screens.ScreenAccessDenied
		m.current = m.accessDeniedModel
		return m, tea.Batch(m.current.Init(), m.resizeCmd())

	case screens.ConfirmPromptMsg:
		m.confirm = &msg
		return m, nil

	case tea.QuitMsg:
		return m, tea.Quit
	}

	var cmd tea.Cmd
	updated, cmd := m.current.Update(msg)
	m.current = updated
	return m, cmd
}

func (m *Model) handleNavigate(msg screens.NavigateMsg) (tea.Model, tea.Cmd) {
	switch msg.Screen {
	case screens.ScreenLogin:
		lm := screens.NewLoginModel(m.apiClient)
		m.loginModel = &lm
		m.screen = screens.ScreenLogin
		m.current = m.loginModel

	case screens.ScreenDashboard:
		dm := screens.NewDashboardModel(m.apiClient, m.apiClient.User)
		m.dashModel = &dm
		m.screen = screens.ScreenDashboard
		m.current = m.dashModel

	case screens.ScreenDocList:
		dlm := screens.NewDocumentListModel(m.apiClient, m.apiClient.User)
		m.docListModel = &dlm
		m.screen = screens.ScreenDocList
		m.current = m.docListModel

	case screens.ScreenDocCreate:
		dcm := screens.NewDocCreateModel(m.apiClient, m.apiClient.User)
		m.docCreateModel = &dcm
		m.screen = screens.ScreenDocCreate
		m.current = m.docCreateModel

	case screens.ScreenUsers:
		um := screens.NewUsersModel(m.apiClient)
		m.usersModel = &um
		m.screen = screens.ScreenUsers
		m.current = m.usersModel

	case screens.ScreenAudit:
		am := screens.NewAuditLogModel(m.apiClient)
		m.auditModel = &am
		m.screen = screens.ScreenAudit
		m.current = m.auditModel
	}

	return m, m.current.Init()
}

func (m *Model) View() string {
	view := m.current.View()

	if m.confirm != nil {
		themeHint := ""
		if m.screen == screens.ScreenLogin || m.screen == screens.ScreenDashboard {
			themeHint = fmt.Sprintf(" │ %s", styles.CurrentTheme.Name)
		}

		overlay := lipgloss.NewStyle().
			Width(m.width).
			Height(m.height - 1).
			Background(styles.Bg)

		box := styles.ConfirmBoxStyle.Render(
			styles.ConfirmTitleStyle.Render("⚠  Confirm") + "\n\n" +
				m.confirm.Message + "\n\n" +
				styles.ConfirmPromptStyle.Render("[y] Yes  [n] No") +
				themeHint,
		)

		return overlay.Render(lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, box))
	}

	return view
}
