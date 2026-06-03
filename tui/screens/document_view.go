package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/internal/domain"
	"classified-vault/tui/styles"
)

type DocumentViewModel struct {
	doc    *domain.Document
	user   *domain.User
	width  int
	height int
}

func NewDocumentViewModel(doc *domain.Document, user *domain.User) DocumentViewModel {
	return DocumentViewModel{
		doc:  doc,
		user: user,
	}
}

func (m *DocumentViewModel) Init() tea.Cmd { return nil }

func (m *DocumentViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch strings.ToUpper(msg.String()) {
		case "H", "Q":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocList} }
		case "CTRL+C":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *DocumentViewModel) View() string {
	header := fmt.Sprintf("📄 %s", styles.DocTitle.Render(m.doc.Title))

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	sb.WriteString(styles.DocMeta.Render(
		fmt.Sprintf("Classification:  %s", styles.ClearanceBadge(m.doc.Classification.String())),
	) + "\n")
	sb.WriteString(styles.DocMeta.Render(
		fmt.Sprintf("Created:        %s", m.doc.CreatedAt.Format("2006-01-02 15:04")),
	) + "\n")
	sb.WriteString(styles.DocMeta.Render(
		fmt.Sprintf("Author:         %s", m.doc.CreatedBy),
	) + "\n")
	if len(m.doc.Tags) > 0 {
		sb.WriteString(styles.DocMeta.Render(
			fmt.Sprintf("Tags:           %s", strings.Join(m.doc.Tags, ", ")),
		) + "\n")
	}
	sb.WriteString("\n" + lipgloss.NewStyle().
		Width(60).
		MaxHeight(20).
		Render(m.doc.Content))

	content := styles.BorderStyle.Render(sb.String())
	main := lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Center, content)
	footer := styles.StatusBarStyle.Width(m.width).Render("[h] Back  [q] Quit")

	return main + "\n" + footer
}
