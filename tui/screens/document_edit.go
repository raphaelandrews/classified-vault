package screens

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"classified-vault/internal/domain"
	"classified-vault/tui/client"
	"classified-vault/tui/styles"
)

type DocEditModel struct {
	apiClient *client.APIClient
	user      *domain.User
	docID     string
	width     int
	height    int

	titleInput   textinput.Model
	contentInput textarea.Model
	tagsInput    textinput.Model
	spinner      spinner.Model

	classification int
	department     string

	focusIdx int
	err      string
	done     bool
	result   string
	loading  bool
}

func NewDocEditModel(api *client.APIClient, user *domain.User, msg EditDocMsg) DocEditModel {
	ti := textinput.New()
	ti.SetValue(msg.Title)
	ti.CharLimit = 80
	ti.Width = 50
	ti.Prompt = ""

	ca := textarea.New()
	ca.SetValue(msg.Content)
	ca.CharLimit = 5000
	ca.SetWidth(60)
	ca.SetHeight(10)
	ca.ShowLineNumbers = false

	ta := textinput.New()
	ta.SetValue(strings.Join(msg.Tags, ", "))
	ta.CharLimit = 200
	ta.Width = 50
	ta.Prompt = ""

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(styles.Accent)

	return DocEditModel{
		apiClient:      api,
		user:           user,
		docID:          msg.DocID,
		classification: int(msg.Classif),
		department:     msg.Department,
		titleInput:     ti,
		contentInput:   ca,
		tagsInput:      ta,
		spinner:        sp,
	}
}

func (m *DocEditModel) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m *DocEditModel) updateDoc() tea.Msg {
	var tags []string
	if strings.TrimSpace(m.tagsInput.Value()) != "" {
		for _, t := range strings.Split(m.tagsInput.Value(), ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
	}
	if tags == nil {
		tags = []string{}
	}

	doc, err := m.apiClient.UpdateDocument(
		m.docID,
		m.titleInput.Value(),
		m.contentInput.Value(),
		domain.ClearanceLevel(m.classification),
		domain.Department(m.department),
		tags,
	)
	if err != nil {
		return err
	}
	return doc
}

func (m *DocEditModel) focusField(idx int) {
	m.focusIdx = idx
	m.titleInput.Blur()
	m.contentInput.Blur()
	m.tagsInput.Blur()
	switch idx {
	case 0:
		m.titleInput.Focus()
	case 1:
		m.contentInput.Focus()
	case 2:
		m.tagsInput.Focus()
	}
}

func (m *DocEditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch strings.ToUpper(key.String()) {
			case "Q", "ENTER":
				return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocList} }
			}
		}
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case *domain.Document:
		m.done = true
		m.result = styles.SuccessStyle.Render(fmt.Sprintf("Scroll updated: %s", msg.Title))
		return m, nil

	case error:
		m.err = msg.Error()
		m.loading = false
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			return m, func() tea.Msg { return NavigateMsg{Screen: ScreenDocList} }
		case "tab":
			m.focusField((m.focusIdx + 1) % 4)
			return m, nil
		case "shift+tab":
			m.focusField((m.focusIdx + 3) % 4)
			return m, nil
		case "up", "k":
			if m.focusIdx == 3 && m.classification > 0 {
				m.classification--
			}
		case "down", "j":
			if m.focusIdx == 3 && m.classification < tierCount-1 {
				m.classification++
			}
		case "enter":
			if m.focusIdx < 2 {
				m.focusField(m.focusIdx + 1)
				m.err = ""
			} else if m.focusIdx == 2 {
				if m.titleInput.Value() == "" {
					m.err = "Title is required"
					return m, nil
				}
				if m.contentInput.Value() == "" {
					m.err = "Content is required"
					return m, nil
				}
				m.loading = true
				m.err = ""
				return m, m.updateDoc
			} else if m.focusIdx == 3 {
				if m.titleInput.Value() == "" {
					m.err = "Title is required"
					return m, nil
				}
				if m.contentInput.Value() == "" {
					m.err = "Content is required"
					return m, nil
				}
				m.loading = true
				m.err = ""
				return m, m.updateDoc
			}
		case "backspace":
			if m.focusIdx == 3 {
				m.focusIdx = 2
				m.focusField(2)
				return m, nil
			}
		}
	}

	if m.focusIdx == 0 {
		var cmd tea.Cmd
		m.titleInput, cmd = m.titleInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.focusIdx == 1 {
		var cmd tea.Cmd
		m.contentInput, cmd = m.contentInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.focusIdx == 2 {
		var cmd tea.Cmd
		m.tagsInput, cmd = m.tagsInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *DocEditModel) View() string {
	if m.done {
		var sb strings.Builder
		sb.WriteString(styles.DocTitle.Render("★ Scroll Updated") + "\n\n")
		sb.WriteString(m.result + "\n\n")
		sb.WriteString(styles.DocMeta.Render("[Q] Back to Scrolls"))
		return lipgloss.Place(
			m.width, m.height,
			lipgloss.Center, lipgloss.Center,
			styles.BorderStyle.Render(sb.String()),
		)
	}

	var sb strings.Builder
	sb.WriteString(styles.DocTitle.Render("★ Edit Scroll") + "\n\n")

	titleLabel := styles.DocPrompt.Render("Title")
	if m.focusIdx == 0 {
		titleLabel = styles.DocPrompt.Render("▶ Title")
	}
	sb.WriteString(titleLabel + "\n" + m.titleInput.View() + "\n\n")

	contentLabel := styles.DocPrompt.Render("Content")
	if m.focusIdx == 1 {
		contentLabel = styles.DocPrompt.Render("▶ Content")
	}
	sb.WriteString(contentLabel + "\n" + m.contentInput.View() + "\n\n")

	tierLabel := styles.DocPrompt.Render("Access Tier")
	if m.focusIdx == 3 {
		tierLabel = styles.DocPrompt.Render("▶ Access Tier")
	}
	sb.WriteString(tierLabel + "\n")
	tiers := []domain.ClearanceLevel{domain.TierPublic, domain.TierCouncil, domain.TierGuild, domain.TierCorporate, domain.TierArcane, domain.TierJunimo}
	for i, tier := range tiers {
		marker := " "
		if i == m.classification {
			marker = "▶"
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", marker, styles.ClearanceBadge(tier.String())))
	}
	sb.WriteString("\n")

	tagsLabel := styles.DocPrompt.Render("Tags (comma-separated)")
	if m.focusIdx == 2 {
		tagsLabel = styles.DocPrompt.Render("▶ Tags (comma-separated)")
	}
	sb.WriteString(tagsLabel + "\n" + m.tagsInput.View() + "\n")

	if m.err != "" {
		sb.WriteString("\n" + styles.ErrorStyle.Render(m.err))
	}

	if m.loading {
		sb.WriteString("\n" + m.spinner.View() + " Updating scroll...")
	}

	sb.WriteString(styles.DocMeta.Render("\n\n[Tab] Next  [↑/↓] Select tier  [Enter] Save  [Esc] Cancel"))

	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Left, lipgloss.Top,
		styles.BorderStyle.Render(sb.String()),
	)
}
