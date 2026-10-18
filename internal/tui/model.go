// Package tui is for Bubbletea
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
	"github.com/batt0s/tamagotchi.ssh/internal/storage"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const saveInterval = 15 * time.Second

type saveTickMsg struct{}
type refreshMsg struct{}

type Deps struct {
	Fingerprint string
	Pet         *game.Pet
	Balance     *game.Balance
	Storage     storage.Storage
}

type Model struct {
	deps        Deps
	width       int
	height      int
	quitting    bool
	lastSavedAt time.Time
	saveErr     error
}

func New(d Deps, width, height int) Model {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	return Model{
		deps:   d,
		width:  width,
		height: height,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(saveTick(), tickEverySecond())
}

func saveTick() tea.Cmd {
	return tea.Tick(saveInterval, func(time.Time) tea.Msg { return saveTickMsg{} })
}

func tickEverySecond() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return refreshMsg{} })
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case saveTickMsg:
		m.persist()
		return m, saveTick()
	case refreshMsg:
		m.deps.Pet.Apply(time.Now(), m.deps.Balance)
		return m, tickEverySecond()
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.persist()
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) persist() {
	if m.deps.Pet == nil || m.deps.Storage == nil {
		return
	}
	m.deps.Pet.Apply(time.Now(), m.deps.Balance)
	if err := m.deps.Storage.Save(m.deps.Fingerprint, m.deps.Pet); err != nil {
		m.saveErr = err
		return
	}
	m.saveErr = nil
	m.lastSavedAt = time.Now()
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	p := m.deps.Pet

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("🐣 SSH Tamagotchi @ battos.dev")

	status := fmt.Sprintf("Satiety %d · Happiness %d · Energy %d",
		p.Satiety, p.Happiness, p.Energy)

	saveInfo := "-"
	if !m.lastSavedAt.IsZero() {
		saveInfo = fmt.Sprintf("%s seconds ago",
			time.Since(m.lastSavedAt).Round(time.Second))
	}
	if m.saveErr != nil {
		saveInfo = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).
			Render("ERROR: " + m.saveErr.Error())
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("q / esc / ctrl+c → quit")

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		label("Identity")+"  "+shorten(m.deps.Fingerprint),
		label("Status")+"    "+status,
		"                    "+statusLine(p),
		"",
		label("Last Saved")+" "+saveInfo,
		"",
		help,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func label(s string) string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Width(11).Render(s)
}

func shorten(fp string) string {
	s := strings.TrimPrefix(fp, "SHA256:")
	if len(s) > 12 {
		s = s[:12] + "..."
	}
	return "SHA256:" + s
}

func statusLine(p *game.Pet) string {
	switch {
	case p.Dead:
		return "Dead"
	case p.Sick:
		return "Sick"
	case p.Sleeping:
		return "Sleeping"
	default:
		return "Awake, healthy"
	}
}

type MessageModel struct {
	text     string
	width    int
	height   int
	quitting bool
}

func NewMessage(text string, width, height int) *MessageModel {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	return &MessageModel{text: text, width: width, height: height}
}

func (m *MessageModel) Init() tea.Cmd { return nil }

func (m *MessageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c", "enter":
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *MessageModel) View() string {
	if m.quitting {
		return ""
	}
	body := lipgloss.JoinVertical(lipgloss.Center,
		m.text,
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).
			Render("enter / q  →  exit"),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}
