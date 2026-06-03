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

type DocCreateModel struct {
	apiClient *client.APIClient
	user      *domain.User
	width     int
	height    int

	title          string
	content        string
	classification int
	department        string
	tags           string

	step   int
	err    string
	done   bool
	result string
}

const tierCount = 6

var tierKeys = []string{"1", "2", "3", "4", "5", "6"}

func NewDocCreateModel(api *client.APIClient, user *domain.User) DocCreateModel {
	return DocCreateModel{
		apiClient:      api,
		user:           user,
		classification: 0,
		department:        string(user.Department),
	}
}

func (m *DocCreateModel) Init() tea.Cmd { return nil }

func (m *DocCreateModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		if m.done {
			switch strings.ToUpper(msg.String()) {
			case "Q", "ENTER":
				cmds := []tea.Cmd{
					func() tea.Msg { return NavigateMsg{Screen: ScreenDocList} },
				}
				return m, tea.Batch(cmds...)
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDashboard} }
		case "enter":
			switch m.step {
			case 0:
				if m.title == "" {
					m.err = "Title is required"
					return m, nil
				}
				m.step = 1
				m.err = ""
			case 1:
				if m.content == "" {
					m.err = "Content is required"
					return m, nil
				}
				m.step = 2
				m.err = ""
			case 2:
				m.step = 3
				m.err = ""
			case 3:
				m.err = ""
				var tags []string
				if strings.TrimSpace(m.tags) != "" {
					for _, t := range strings.Split(m.tags, ",") {
						tags = append(tags, strings.TrimSpace(t))
					}
				}
				if tags == nil {
					tags = []string{}
				}

				return m, func() tea.Msg {
					doc, err := m.apiClient.CreateDocument(m.title, m.content, domain.ClearanceLevel(m.classification), domain.Department(m.department), tags)
					if err != nil {
						return err
					}
					return doc
				}
			}
		case "backspace":
			switch m.step {
			case 0:
				if len(m.title) > 0 {
					m.title = m.title[:len(m.title)-1]
				}
			case 1:
				if len(m.content) > 0 {
					m.content = m.content[:len(m.content)-1]
				} else {
					m.step = 0
				}
			case 2:
				m.step = 1
			case 3:
				if len(m.tags) > 0 {
					m.tags = m.tags[:len(m.tags)-1]
				} else {
					m.step = 2
				}
			}
		default:
			key := msg.String()
			if m.step == 2 {
				for i, k := range tierKeys {
					if key == k {
						m.classification = i
					}
				}
			} else if len(msg.Runes) == 1 {
				switch m.step {
				case 0:
					if len(m.title) < 60 {
						m.title += msg.String()
					}
				case 1:
					if len(m.content) < 500 {
						m.content += msg.String()
					}
				case 3:
					if len(m.tags) < 100 {
						m.tags += msg.String()
					}
				}
			}
		}

	case *domain.Document:
		m.done = true
		m.result = styles.SuccessStyle.Render(fmt.Sprintf("Scroll scribed: %s", msg.Title))
		return m, nil

	case error:
		m.err = msg.Error()
		return m, nil
	}

	return m, nil
}

func (m *DocCreateModel) View() string {
	var sb strings.Builder

	if m.done {
		sb.WriteString(styles.DocTitle.Render("★ Scroll Scribed") + "\n\n")
		sb.WriteString(m.result + "\n\n")
		sb.WriteString(styles.DocMeta.Render("[Q] Back to Scrolls"))
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			styles.BorderStyle.Render(sb.String()),
		)
	}

	sb.WriteString(styles.DocTitle.Render("★ Scribe New Scroll") + "\n\n")

	switch m.step {
	case 0:
		sb.WriteString(styles.DocPrompt.Render("Title: ") + m.title + "_")
	case 1:
		sb.WriteString(styles.DocPrompt.Render("Title: ") + styles.DocMeta.Render(m.title) + "\n\n")
		sb.WriteString(styles.DocPrompt.Render("Content: ") + m.content + "_")
	case 2:
		sb.WriteString(styles.DocPrompt.Render("Title: ") + styles.DocMeta.Render(m.title) + "\n")
		sb.WriteString(styles.DocPrompt.Render("Content: ") + styles.DocMeta.Render(truncate(m.content, 40)) + "\n\n")
		sb.WriteString(styles.DocPrompt.Render("Access Tier:\n"))
		tiers := []domain.ClearanceLevel{domain.TierPublic, domain.TierCouncil, domain.TierGuild, domain.TierCorporate, domain.TierArcane, domain.TierJunimo}
		for i, tier := range tiers {
			marker := " "
			if i == m.classification {
				marker = "▶"
			}
			sb.WriteString(fmt.Sprintf("  %s %s  %s\n", marker, styles.ClearanceBadge(tier.String()), fmt.Sprintf("[%d]", i+1)))
		}
	case 3:
		sb.WriteString(styles.DocPrompt.Render("Title: ") + styles.DocMeta.Render(m.title) + "\n")
		sb.WriteString(styles.DocPrompt.Render("Content: ") + styles.DocMeta.Render(truncate(m.content, 40)) + "\n")
		sb.WriteString(styles.DocPrompt.Render("Tier: ") + styles.ClearanceBadge(domain.ClearanceLevel(m.classification).String()) + "\n\n")
		sb.WriteString(styles.DocPrompt.Render("Tags (comma-separated): ") + m.tags + "_")
	}

	if m.err != "" {
		sb.WriteString("\n" + styles.ErrorStyle.Render(m.err))
	}
	sb.WriteString(styles.DocMeta.Render("\n\n[Enter] Continue  [Backspace] Back  [Esc] Cancel"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		styles.BorderStyle.Render(sb.String()),
	)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
