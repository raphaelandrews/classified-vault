package tui

import (
	"fmt"
	"strings"

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
	docEditModel      *screens.DocEditModel
	accessDeniedModel *screens.AccessDeniedModel
	usersModel        *screens.UsersModel
	auditModel        *screens.AuditLogModel

	current  tea.Model
	width    int
	height   int
	themeIdx int

	confirm  *screens.ConfirmPromptMsg
	showHelp bool
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

func (m *Model) helpView() string {
	var sb strings.Builder
	sb.WriteString(styles.DocTitle.Render("HELP") + "\n\n")

	switch m.screen {
	case screens.ScreenLogin:
		sb.WriteString("[Tab] Switch field\n[Enter] Sign In\n[Esc] Quit\n[Ctrl+T] Cycle theme")
	case screens.ScreenDashboard:
		sb.WriteString("[d] Browse Scrolls\n[a] Scribe New\n[u] Manage Villagers\n[l] Town Ledger\n[q] Sign Out\n[Ctrl+T] Cycle theme\n[?] Close help")
	case screens.ScreenDocList:
		sb.WriteString("[/] Search\n[j/k] Move cursor\n[←/→] Change page\n[enter] Open scroll\n[e] Edit scroll\n[a] New scroll\n[d] Delete scroll\n[q] Back\n[r] Refresh\n[?] Close help")
	case screens.ScreenDocView:
		sb.WriteString("[↑/↓] Scroll\n[h/q] Back to list\n[?] Close help")
	case screens.ScreenDocCreate:
		sb.WriteString("[Tab] Next field\n[Shift+Tab] Previous field\n[Enter] Continue/Save\n[Backspace] Go back\n[Esc] Cancel\n[?] Close help")
	case screens.ScreenDocEdit:
		sb.WriteString("[Tab] Next field\n[Shift+Tab] Previous field\n[Enter] Save changes\n[Backspace] Go back\n[Esc] Cancel\n[?] Close help")
	case screens.ScreenUsers:
		sb.WriteString("[j/k] Move cursor\n[a] Register villager\n[d] Dismiss villager\n[r] Refresh\n[q] Back\n[?] Close help")
	case screens.ScreenAudit:
		sb.WriteString("[←/→] Page\n[r] Refresh\n[q] Back\n[?] Close help")
	default:
		sb.WriteString("[?] Close help")
	}

	return styles.BorderStyle.Render(sb.String())
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.showHelp {
		if key, ok := msg.(tea.KeyMsg); ok {
			if key.String() == "?" || key.String() == "esc" || key.String() == "q" {
				m.showHelp = false
				return m, nil
			}
		}
		return m, nil
	}

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
		if msg.String() == "?" {
			m.showHelp = true
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

	case screens.EditDocMsg:
		dem := screens.NewDocEditModel(m.apiClient, m.apiClient.User, msg)
		m.docEditModel = &dem
		m.screen = screens.ScreenDocEdit
		m.current = m.docEditModel
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
	if m.showHelp {
		helpBox := m.helpView()
		return lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, helpBox)
	}

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
