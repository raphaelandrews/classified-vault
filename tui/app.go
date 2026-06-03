package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"classified-vault/tui/client"
	"classified-vault/tui/screens"
)

type Model struct {
	screen screens.Screen

	apiClient *client.APIClient

	loginModel        screens.LoginModel
	dashModel         screens.DashboardModel
	docListModel      screens.DocumentListModel
	docViewModel      screens.DocumentViewModel
	docCreateModel    screens.DocCreateModel
	accessDeniedModel screens.AccessDeniedModel
	usersModel        screens.UsersModel
	auditModel        screens.AuditLogModel

	current tea.Model
	width   int
	height  int
}

func New(serverURL string) *Model {
	api := client.New(serverURL)
	m := &Model{
		screen:    screens.ScreenLogin,
		apiClient: api,
	}
	m.loginModel = screens.NewLoginModel(api)
	m.current = m.loginModel
	return m
}

func (m *Model) Init() tea.Cmd {
	return m.current.Init()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case screens.LoginMsg:
		m.dashModel = screens.NewDashboardModel(m.apiClient, msg.User)
		m.screen = screens.ScreenDashboard
		m.current = m.dashModel
		return m, m.current.Init()

	case screens.NavigateMsg:
		return m.handleNavigate(msg)

	case screens.DocSelectedMsg:
		m.docViewModel = screens.NewDocumentViewModel(msg.Doc, m.apiClient.User)
		m.screen = screens.ScreenDocView
		m.current = m.docViewModel
		return m, m.current.Init()

	case screens.DocAccessDeniedMsg:
		m.accessDeniedModel = screens.NewAccessDeniedModel(msg)
		m.screen = screens.ScreenAccessDenied
		m.current = m.accessDeniedModel
		return m, m.current.Init()

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
		m.loginModel = screens.NewLoginModel(m.apiClient)
		m.screen = screens.ScreenLogin
		m.current = m.loginModel

	case screens.ScreenDashboard:
		m.dashModel = screens.NewDashboardModel(m.apiClient, m.apiClient.User)
		m.screen = screens.ScreenDashboard
		m.current = m.dashModel

	case screens.ScreenDocList:
		m.docListModel = screens.NewDocumentListModel(m.apiClient, m.apiClient.User)
		m.screen = screens.ScreenDocList
		m.current = m.docListModel

	case screens.ScreenDocCreate:
		m.docCreateModel = screens.NewDocCreateModel(m.apiClient, m.apiClient.User)
		m.screen = screens.ScreenDocCreate
		m.current = m.docCreateModel

	case screens.ScreenUsers:
		m.usersModel = screens.NewUsersModel(m.apiClient)
		m.screen = screens.ScreenUsers
		m.current = m.usersModel

	case screens.ScreenAudit:
		m.auditModel = screens.NewAuditLogModel(m.apiClient)
		m.screen = screens.ScreenAudit
		m.current = m.auditModel
	}

	return m, m.current.Init()
}

func (m *Model) View() string {
	return m.current.View()
}
