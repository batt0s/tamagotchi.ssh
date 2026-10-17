package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	fingerprint string
	width       int
	height      int
	quitting    bool
}

func New(fingerprint string, width, height int) Model {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	return Model{
		fingerprint: fingerprint,
		width:       width,
		height:      height,
	}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("🐣 SSH Tamagotchi @ battos.dev")

	fpLine := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Identity: " + m.fingerprint)

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("q / esc / ctrl+c -> quit")

	body := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		fpLine,
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func shorten(fp string) string {
	s := strings.TrimPrefix(fp, "SHA256:")
	if len(s) > 12 {
		s = s[:12] + "..."
	}
	return "SHA256:" + s
}
