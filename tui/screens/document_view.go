package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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
	contentWidth := m.width - 12
	if contentWidth < 40 {
		contentWidth = 40
	}

	header := styles.DocViewTitle.Render("*", m.doc.Title)

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	sb.WriteString(styles.DocMeta.Render(
		fmt.Sprintf("Tier:       %s", styles.ClearanceBadge(m.doc.Classification.String())),
	) + "\n")
	sb.WriteString(styles.DocMeta.Render(
		fmt.Sprintf("Department: %s", styles.DepartmentBadge(string(m.doc.Department))),
	) + "\n")
	if m.doc.Folder != "" {
		sb.WriteString(styles.DocMeta.Render(
			fmt.Sprintf("Folder:     ▸ %s", m.doc.Folder),
		) + "\n")
	}
	sb.WriteString(styles.DocMeta.Render(
		fmt.Sprintf("Scribed:    %s", m.doc.CreatedAt.Format("2006-01-02 15:04")),
	) + "\n")
	sb.WriteString(styles.DocMeta.Render(
		fmt.Sprintf("Author:     %s", m.doc.CreatedBy),
	) + "\n")
	if len(m.doc.Tags) > 0 {
		sb.WriteString(styles.DocMeta.Render(
			fmt.Sprintf("Tags:       %s", strings.Join(m.doc.Tags, ", ")),
		) + "\n")
	}
	if len(m.doc.ReferenceIDs) > 0 {
		sb.WriteString(styles.DocMeta.Render(
			fmt.Sprintf("Refs:       %s", strings.Join(m.doc.ReferenceIDs, ", ")),
		) + "\n")
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(contentWidth),
	)
	if err != nil {
		sb.WriteString("\n" + m.doc.Content)
	} else {
		md, err := r.Render(m.doc.Content)
		if err != nil {
			sb.WriteString("\n" + m.doc.Content)
		} else {
			sb.WriteString("\n" + md)
		}
	}

	content := styles.BorderStyle.Render(sb.String())
	main := lipgloss.Place(m.width, m.height-1, lipgloss.Left, lipgloss.Top, content)
	footer := styles.StatusBarStyle.Width(m.width).Render("[h] Back  [q] Quit")

	return main + "\n" + footer
}
