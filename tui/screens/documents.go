package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"classified-vault/internal/domain"
	"classified-vault/tui/client"
	"classified-vault/tui/styles"
)

type DocumentListModel struct {
	docs      []client.CatalogEntry
	cursor    int
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
	Faction     string
	UserCle     domain.ClearanceLevel
	RequiredCle domain.ClearanceLevel
}

func NewDocumentListModel(api *client.APIClient, user *domain.User) DocumentListModel {
	return DocumentListModel{
		apiClient: api,
		user:      user,
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

func (m *DocumentListModel) canAccess(entry client.CatalogEntry) bool {
	docTier := domain.ClearanceLevel(entry.Classification)
	docFaction := domain.Faction(entry.Faction)

	if docTier == domain.TierPublic {
		return true
	}
	if m.user.Faction == docFaction && m.user.Clearance >= docTier {
		return true
	}
	if m.user.Faction == domain.FactionMayorsOffice && m.user.Clearance >= domain.TierArcane {
		return true
	}
	if m.user.Faction == domain.FactionWizardsTower && containsArcane(entry.Tags) {
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

func (m *DocumentListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case []client.CatalogEntry:
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
			} else {
				m.cursor = len(m.docs) - 1
			}
		case "DOWN", "J":
			if m.cursor < len(m.docs)-1 {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "ENTER", "L":
			if len(m.docs) > 0 && m.cursor < len(m.docs) {
				entry := m.docs[m.cursor]
				docTier := domain.ClearanceLevel(entry.Classification)
				if m.canAccess(entry) {
					return m, func() tea.Msg { return DocSelectedMsg{DocID: entry.ID} }
				}
				return m, func() tea.Msg {
					return DocAccessDeniedMsg{
						Title:       entry.Title,
						Faction:     entry.Faction,
						UserCle:     m.user.Clearance,
						RequiredCle: docTier,
					}
				}
			}
		case "A":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocCreate} }
		case "H", "Q":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDashboard} }
		case "CTRL+C":
			return m, tea.Quit
		case "R":
			return m, m.loadDocs
		}
	}

	return m, nil
}

func (m *DocumentListModel) View() string {
	var sb strings.Builder

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary).
		Render(fmt.Sprintf("SCROLLS — %s  %s", styles.ClearanceBadge(m.user.Clearance.String()), styles.FactionBadge(string(m.user.Faction))))
	sb.WriteString(header + "\n\n")

	if m.err != "" {
		sb.WriteString(styles.ErrorStyle.Render(m.err) + "\n\n")
	}

	if len(m.docs) == 0 {
		sb.WriteString(styles.DocMeta.Render("  No scrolls found.\n"))
	} else {
		t := table.New().
			Border(lipgloss.NormalBorder()).
			BorderStyle(lipgloss.NewStyle().Foreground(styles.BorderCol)).
			Width(m.width-8).
			StyleFunc(func(row, col int) lipgloss.Style {
				base := lipgloss.NewStyle().Padding(0, 1)
				switch {
				case row == table.HeaderRow:
					return base.Foreground(styles.Foreground).Bold(true)
				case row == m.cursor:
					return base.Foreground(styles.DarkText).Background(styles.Selected).Bold(true)
				case row%2 == 0:
					return base.Foreground(styles.RowEven)
				default:
					return base.Foreground(styles.RowOdd)
				}
			}).
			Headers("", "TITLE", "TIER", "FACTION", "FOLDER", "DATE")

		for i, entry := range m.docs {
			marker := fmt.Sprintf("%d", i+1)
			title := entry.Title
			tier := styles.ClearanceBadge(domain.ClearanceLevel(entry.Classification).String())
			faction := styles.FactionBadge(entry.Faction)
			folder := ""
			if entry.Folder != "" {
				folder = styles.DocMeta.Render("▸ " + entry.Folder)
			}
			date := styles.DocMeta.Render(entry.CreatedAt[:10])

			if !m.canAccess(entry) {
				marker = "🔒"
				title = styles.DocMeta.Render(title)
				tier = styles.DocMeta.Render("[ SEALED        ]")
			}
			if i == m.cursor {
				marker = "▶"
			}

			t.Row(marker, title, tier, faction, folder, date)
		}

		sb.WriteString(t.Render())
	}

	content := styles.BorderStyle.Render(sb.String())
	main := lipgloss.Place(m.width, m.height-1, lipgloss.Center, lipgloss.Top, content)
	footer := styles.StatusBarStyle.Width(m.width).Render("[j/k] Move  [l] Open  [a] New  [h] Back  [r] Refresh")

	return main + "\n" + footer
}
