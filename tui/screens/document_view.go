package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/internal/domain"
	"classified-vault/tui/styles"
)

type DocumentViewModel struct {
	doc      *domain.Document
	user     *domain.User
	width    int
	height   int
	viewport viewport.Model
	ready    bool
}

func NewDocumentViewModel(doc *domain.Document, user *domain.User) DocumentViewModel {
	return DocumentViewModel{
		doc:  doc,
		user: user,
	}
}

func (m *DocumentViewModel) Init() tea.Cmd { return nil }

func (m *DocumentViewModel) buildContent() string {
	contentWidth := m.width - 12
	if contentWidth < 40 {
		contentWidth = 40
	}

	header := styles.DocViewTitle.Render("★ " + m.doc.Title)

	var sb strings.Builder
	sb.WriteString(header + "\n\n")
	sb.WriteString(styles.DocMeta.Render(
		fmt.Sprintf("Status:     %s", statusBadge(m.doc.Status)),
	) + "\n")
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

	if !m.doc.VerifyIntegrity() {
		sb.WriteString("\n")
		sb.WriteString(styles.ErrorStyle.Render("⚠ TAMPERING DETECTED — content hash mismatch") + "\n")
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

	return styles.BorderStyle.Render(sb.String())
}

func (m *DocumentViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		if !m.ready {
			m.viewport = viewport.New(m.width, m.height-1)
			m.viewport.SetContent(m.buildContent())
			m.ready = true
		} else {
			m.viewport.Width = m.width
			m.viewport.Height = m.height - 1
			m.viewport.SetContent(m.buildContent())
		}
		return m, nil

	case tea.KeyMsg:
		switch strings.ToUpper(msg.String()) {
		case "H", "Q":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocList} }
		case "CTRL+C":
			return m, tea.Quit
		}
	}

	m.viewport, cmd = m.viewport.Update(msg)
	return m, cmd
}

func (m *DocumentViewModel) View() string {
	footer := styles.StatusBarStyle.Width(m.width).Render("[↑/↓] Scroll  [h] Back  [q] Quit  [?] Help")
	return m.viewport.View() + "\n" + footer
}

func statusBadge(s domain.DocumentStatus) string {
	switch s {
	case domain.StatusDraft:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")).Render("[ DRAFT    ]")
	case domain.StatusReview:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#d8a657")).Render("[ REVIEW   ]")
	case domain.StatusFrozen:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#83a598")).Render("[ FROZEN   ]")
	case domain.StatusArchived:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#928374")).Render("[ ARCHIVED ]")
	case domain.StatusPublic:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#a9b665")).Render("[ PUBLIC   ]")
	default:
		return lipgloss.NewStyle().Foreground(styles.Dimmed).Render("[ " + string(s) + " ]")
	}
}
