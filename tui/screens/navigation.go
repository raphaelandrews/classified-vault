package screens

type Screen int

const (
	ScreenLogin Screen = iota
	ScreenDashboard
	ScreenDocList
	ScreenDocView
	ScreenDocCreate
	ScreenAccessDenied
	ScreenUsers
	ScreenAudit
)

type NavigateMsg struct {
	Screen Screen
	Data   interface{}
}
