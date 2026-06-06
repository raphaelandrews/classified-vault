package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/internal/domain"
	"classified-vault/tui/client"
	"classified-vault/tui/styles"
)

type DocumentListModel struct {
	docs      []client.CatalogEntry
	filtered  []client.CatalogEntry
	cursor    int
	searching bool
	search    textinput.Model
	paginator paginator.Model
	err       string
	apiClient *client.APIClient
	user      *domain.User
	width     int
	height    int
}

type DocSelectedMsg struct {
	DocID string
}

type DocAccessDeniedMsg struct {
	Title       string
	Department  string
	UserCle     domain.ClearanceLevel
	RequiredCle domain.ClearanceLevel
}

type EditDocMsg struct {
	DocID      string
	Title      string
	Content    string
	Department string
	Classif    domain.ClearanceLevel
	Tags       []string
}

func NewDocumentListModel(api *client.APIClient, user *domain.User) DocumentListModel {
	s := textinput.New()
	s.Placeholder = "Search scrolls..."
	s.Width = 40
	s.Prompt = "/ "

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 20
	p.ActiveDot = lipgloss.NewStyle().Foreground(styles.Primary).Render("●")
	p.InactiveDot = lipgloss.NewStyle().Foreground(styles.Dimmed).Render("○")

	return DocumentListModel{
		apiClient: api,
		user:      user,
		search:    s,
		paginator: p,
		cursor:    -1,
	}
}

func (m *DocumentListModel) Init() tea.Cmd {
	return m.loadDocs
}

func (m *DocumentListModel) loadDocs() tea.Msg {
	entries, err := m.apiClient.ListCatalog()
	if err != nil {
		return fmt.Errorf("failed to load catalog: %w", err)
	}
	return entries
}

func (m *DocumentListModel) contentSearch() tea.Msg {
	if m.search.Value() == "" {
		return nil
	}
	entries, err := m.apiClient.SearchDocuments(m.search.Value())
	if err != nil {
		return fmt.Errorf("search failed: %w", err)
	}
	return entries
}

func (m *DocumentListModel) canAccess(entry client.CatalogEntry) bool {
	docTier := domain.ClearanceLevel(entry.Classification)
	docDept := domain.Department(entry.Department)

	if docTier == domain.TierPublic {
		return true
	}
	if m.user.Department == docDept && m.user.Clearance >= docTier {
		return true
	}
	if m.user.Department == domain.DepartmentMayorsOffice && m.user.Clearance >= domain.TierArcane {
		return true
	}
	if m.user.Department == domain.DepartmentWizardsTower && containsArcane(entry.Tags) {
		return true
	}
	return false
}

func containsArcane(tags []string) bool {
	for _, t := range tags {
		if t == "arcane" {
			return true
		}
	}
	return false
}

func (m *DocumentListModel) applyFilter() {
	if m.search.Value() == "" {
		m.filtered = m.docs
	} else {
		query := strings.ToLower(m.search.Value())
		var filtered []client.CatalogEntry
		for _, doc := range m.docs {
			if strings.Contains(strings.ToLower(doc.Title), query) ||
				strings.Contains(strings.ToLower(doc.Department), query) ||
				strings.Contains(strings.ToLower(doc.Folder), query) {
				filtered = append(filtered, doc)
			}
		}
		m.filtered = filtered
	}
	m.paginator.SetTotalPages(len(m.filtered))
	m.paginator.Page = 0
	m.cursor = -1
}

func (m *DocumentListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc":
				m.searching = false
				m.search.SetValue("")
				m.search.Placeholder = "Search scrolls..."
				m.search.Blur()
				m.applyFilter()
				return m, nil
			case "enter":
				m.searching = false
				m.search.Blur()
				if m.search.Placeholder == "FTS5 content search..." {
					return m, m.contentSearch
				}
				m.applyFilter()
				return m, nil
			}
		}
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		m.applyFilter()
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case []client.CatalogEntry:
		m.docs = msg
		m.applyFilter()
		m.err = ""
		return m, nil

	case error:
		m.err = msg.Error()
		return m, nil

	case tea.KeyMsg:
		switch strings.ToUpper(msg.String()) {
		case "/":
			m.searching = true
			m.search.SetValue("")
			m.search.Focus()
			return m, nil
		case "ctrl+f":
			m.searching = true
			m.search.SetValue("")
			m.search.Placeholder = "FTS5 content search..."
			m.search.Focus()
			return m, nil
		case "UP", "K":
			start, end := m.paginator.GetSliceBounds(len(m.filtered))
			pageLen := end - start
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = pageLen - 1
			}
		case "DOWN", "J":
			start, end := m.paginator.GetSliceBounds(len(m.filtered))
			pageLen := end - start
			if m.cursor < pageLen-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "LEFT", "H":
			if m.paginator.Page > 0 {
				m.paginator.PrevPage()
				m.cursor = -1
			}
		case "RIGHT", "L":
			if m.paginator.Page < m.paginator.TotalPages-1 {
				m.paginator.NextPage()
				m.cursor = -1
			}
		case "ENTER":
			idx := m.actualIndex()
			if idx >= 0 && idx < len(m.filtered) {
				entry := m.filtered[idx]
				docTier := domain.ClearanceLevel(entry.Classification)
				if m.canAccess(entry) {
					return m, func() tea.Msg { return DocSelectedMsg{DocID: entry.ID} }
				}
				return m, func() tea.Msg {
					return DocAccessDeniedMsg{
						Title:       entry.Title,
						Department:  entry.Department,
						UserCle:     m.user.Clearance,
						RequiredCle: docTier,
					}
				}
			}
		case "E":
			idx := m.actualIndex()
			if idx >= 0 && idx < len(m.filtered) {
				entry := m.filtered[idx]
				if !m.canAccess(entry) {
					return m, nil
				}
				return m, func() tea.Msg {
					doc, err := m.apiClient.GetDocument(entry.ID)
					if err != nil {
						return err
					}
					tags := doc.Tags
					if tags == nil {
						tags = []string{}
					}
					return EditDocMsg{
						DocID:      doc.ID,
						Title:      doc.Title,
						Content:    doc.Content,
						Department: string(doc.Department),
						Classif:    doc.Classification,
						Tags:       tags,
					}
				}
			}
		case "A":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocCreate} }
		case "D":
			idx := m.actualIndex()
			if idx >= 0 && idx < len(m.filtered) {
				entry := m.filtered[idx]
				id := entry.ID
				title := entry.Title
				return m, func() tea.Msg {
					return ConfirmPromptMsg{
						Message: fmt.Sprintf("Destroy scroll \"%s\"?\nThis action cannot be undone.", title),
						OnYes: func() tea.Msg {
							err := m.apiClient.DeleteDocument(id)
							if err != nil {
								return err
							}
							return NavigateMsg{Screen: ScreenDocList}
						},
					}
				}
			}
		case "Q":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDashboard} }
		case "CTRL+C":
			return m, tea.Quit
		case "R":
			return m, m.loadDocs
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *DocumentListModel) View() string {
	var sb strings.Builder

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary).
		Render(fmt.Sprintf("SCROLLS — %s  %s", styles.ClearanceBadge(m.user.Clearance.String()), styles.DepartmentBadge(string(m.user.Department))))
	sb.WriteString(header + "\n")

	if m.searching {
		hint := "[/] Title search (type + enter to filter)"
		if m.search.Placeholder == "FTS5 content search..." {
			hint = "[ctrl+f] FTS5 content search (type + enter to search)"
		}
		sb.WriteString("\n" + m.search.View() + "\n")
		sb.WriteString(styles.DocMeta.Render(hint) + "\n")
	} else if m.search.Value() != "" {
		sb.WriteString(styles.DocMeta.Render(fmt.Sprintf("\nFilter: \"%s\"  [/] Change  [esc] Clear", m.search.Value())) + "\n")
	}
	sb.WriteString("\n")

	if m.err != "" {
		sb.WriteString(styles.ErrorStyle.Render(m.err) + "\n\n")
	}

	if len(m.filtered) == 0 {
		sb.WriteString(styles.DocMeta.Render("  No scrolls found.\n"))
	} else {
		start, end := m.paginator.GetSliceBounds(len(m.filtered))
		page := m.filtered[start:end]

		for i, entry := range page {
			marker := "  "
			if i == m.cursor {
				marker = "▶ "
			}

			selected := i == m.cursor

			if !m.canAccess(entry) {
				card := lipgloss.NewStyle().
					Border(lipgloss.NormalBorder(), false, false, false, true).
					BorderForeground(styles.BorderCol).
					Padding(0, 1).
					PaddingBottom(1)
				if selected {
					card = card.BorderForeground(styles.Accent)
				}
				titleLine := marker
				if selected {
					titleLine += styles.DocTitle.Render("████  SEALED  ████")
				} else {
					titleLine += styles.DocMeta.Render("████  SEALED  ████")
				}
				body := titleLine + "\n" +
					"   " + styles.DocMeta.Render(entry.Department)
				sb.WriteString(card.Render(body) + "\n")
				continue
			}

			card := lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(styles.BorderCol).
				Padding(0, 1).
				PaddingBottom(1)

			if selected {
				card = card.BorderForeground(styles.Accent)
			}

			tierBadge := styles.ClearanceBadge(domain.ClearanceLevel(entry.Classification).String())
			deptBadge := styles.DepartmentBadge(entry.Department)
			date := styles.DocMeta.Render(entry.CreatedAt[:10])

			meta := tierBadge + "  ·  " + deptBadge
			if entry.Folder != "" {
				meta += "  ·  " + styles.DocMeta.Render("▸ "+entry.Folder)
			}

			titleLine := marker
			if selected {
				titleLine += styles.DocTitle.Render(truncate(entry.Title, 50))
			} else {
				titleLine += truncate(entry.Title, 50)
			}

			body := titleLine + "\n" +
				"   " + meta + "  ·  " + date

			sb.WriteString(card.Render(body) + "\n")
		}

		if m.paginator.TotalPages > 1 {
			sb.WriteString("\n  " + m.paginator.View())
		}
	}

	content := styles.BorderStyle.Render(sb.String())
	footer := styles.StatusBarStyle.Width(m.width).Render("[/] Search  [ctrl+f] FTS5  [j/k] Move  [h/l] Page  [enter] Open  [e] Edit  [a] New  [d] Destroy  [q] Back  [r] Refresh")

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}

func (m *DocumentListModel) actualIndex() int {
	if m.cursor < 0 {
		return -1
	}
	start, _ := m.paginator.GetSliceBounds(len(m.filtered))
	return start + m.cursor
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
