package screens

import tea "github.com/charmbracelet/bubbletea"

type Screen int

const (
	ScreenLogin Screen = iota
	ScreenDashboard
	ScreenDocList
	ScreenDocView
	ScreenDocCreate
	ScreenDocEdit
	ScreenAccessDenied
	ScreenUsers
	ScreenAudit
	ScreenPasswordChange
)

type NavigateMsg struct {
	Screen Screen
	Data   interface{}
}

type ConfirmPromptMsg struct {
	Message string
	OnYes   func() tea.Msg
}
