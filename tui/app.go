package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"classified-vault/tui/client"
	"classified-vault/tui/screens"
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
	lm := screens.NewLoginModel(api)
	m.loginModel = &lm
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
		dm := screens.NewDashboardModel(m.apiClient, msg.User)
		m.dashModel = &dm
		m.screen = screens.ScreenDashboard
		m.current = m.dashModel
		return m, m.current.Init()

	case screens.NavigateMsg:
		return m.handleNavigate(msg)

	case screens.DocSelectedMsg:
		doc, err := m.apiClient.GetDocument(msg.DocID)
		if err != nil {
			return m, nil
		}
		dvm := screens.NewDocumentViewModel(doc, m.apiClient.User)
		m.docViewModel = &dvm
		m.screen = screens.ScreenDocView
		m.current = m.docViewModel
		return m, m.current.Init()

	case screens.DocAccessDeniedMsg:
		adm := screens.NewAccessDeniedModel(msg)
		m.accessDeniedModel = &adm
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
	return m.current.View()
}
