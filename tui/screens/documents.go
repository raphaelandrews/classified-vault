package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/internal/domain"
	"classified-vault/tui/client"
	"classified-vault/tui/styles"
)

type DocumentListModel struct {
	docs      []*domain.Document
	cursor    int
	err       string
	apiClient *client.APIClient
	user      *domain.User
	width     int
	height    int
}

type DocSelectedMsg struct {
	Doc *domain.Document
}

type DocAccessDeniedMsg struct {
	Doc         *domain.Document
	UserCle     domain.ClearanceLevel
	RequiredCle domain.ClearanceLevel
}

func NewDocumentListModel(api *client.APIClient, user *domain.User) DocumentListModel {
	return DocumentListModel{
		apiClient: api,
		user:      user,
	}
}

func (m DocumentListModel) Init() tea.Cmd {
	return m.loadDocs
}

func (m DocumentListModel) loadDocs() tea.Msg {
	docs, err := m.apiClient.ListDocuments()
	if err != nil {
		return fmt.Errorf("failed to load documents: %w", err)
	}
	return docs
}

func (m DocumentListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case []*domain.Document:
		m.docs = msg
		m.err = ""
		return m, nil
	case error:
		m.err = msg.Error()
		return m, nil
	case tea.KeyMsg:
		switch strings.ToUpper(msg.String()) {
		case "UP", "K":
			if m.cursor > 0 {
				m.cursor--
			}
		case "DOWN", "J":
			if m.cursor < len(m.docs)-1 {
				m.cursor++
			}
		case "ENTER":
			if len(m.docs) > 0 {
				doc := m.docs[m.cursor]
				if m.user.Clearance >= doc.Classification {
					return m, func() tea.Msg { return DocSelectedMsg{Doc: doc} }
				}
				return m, func() tea.Msg {
					return DocAccessDeniedMsg{
						Doc:         doc,
						UserCle:     m.user.Clearance,
						RequiredCle: doc.Classification,
					}
				}
			}
		case "N":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocCreate} }
		case "Q":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDashboard} }
		case "CTRL+C", "ESC":
			return m, tea.Quit
		case "R":
			return m, m.loadDocs
		}
	}
	return m, nil
}

func (m DocumentListModel) View() string {
	header := fmt.Sprintf("📁 DOCUMENTS — clearance: %s",
		styles.ClearanceBadge(m.user.Clearance.String()))

	var sb strings.Builder
	sb.WriteString(header + "\n\n")

	if m.err != "" {
		sb.WriteString(styles.ErrorStyle.Render(m.err) + "\n\n")
	}

	if len(m.docs) == 0 {
		sb.WriteString(styles.DocMeta.Render("  No documents accessible at your clearance level.\n"))
	} else {
		for i, doc := range m.docs {
			cursor := " "
			if i == m.cursor {
				cursor = "▶"
			}
			title := doc.Title
			if m.user.Clearance < doc.Classification {
				title = "[BLOCKED] " + title
				title = styles.DocMeta.Render(title)
			}
			line := fmt.Sprintf("%s %-50s %s %s",
				cursor, title,
				styles.ClearanceBadge(doc.Classification.String()),
				styles.DocMeta.Render(doc.CreatedAt.Format("2006-01-02")),
			)
			if i == m.cursor {
				line = styles.SelectedStyle.Render(line)
			}
			sb.WriteString(line + "\n")
		}
	}

	sb.WriteString(styles.DocMeta.Render("\n[Enter] Open  [N] New  [R] Refresh  [Q] Back"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.BorderStyle.Render(sb.String()),
	)
}
