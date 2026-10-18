// Package tui is for Bubbletea
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
	"github.com/batt0s/tamagotchi.ssh/internal/storage"
	"github.com/charmbracelet/bubbles/progress"
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
	OfflineTime time.Duration
}

type Model struct {
	deps        Deps
	width       int
	height      int
	quitting    bool
	lastSavedAt time.Time
	saveErr     error

	showReport bool
	statusMsg  string
	statusErr  bool

	pSat, pHap, pEne progress.Model
}

func New(d Deps, width, height int) Model {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	showReport := d.OfflineTime > 1*time.Minute
	return Model{
		deps:       d,
		width:      width,
		height:     height,
		showReport: showReport,
		pSat:       progress.New(progress.WithDefaultGradient(), progress.WithWidth(30)),
		pHap:       progress.New(progress.WithDefaultGradient(), progress.WithWidth(30)),
		pEne:       progress.New(progress.WithDefaultGradient(), progress.WithWidth(30)),
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
		w := m.width - 25
		if w < 10 {
			w = 10
		}
		m.pSat.Width = w
		m.pHap.Width = w
		m.pEne.Width = w
		return m, nil
	case saveTickMsg:
		m.persist()
		return m, saveTick()
	case refreshMsg:
		m.deps.Pet.Apply(time.Now(), m.deps.Balance)
		return m, tickEverySecond()
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if m.showReport {
		if key == "q" || key == "ctrl+c" {
			m.persist()
			m.quitting = true
			return m, tea.Quit
		}
		m.showReport = false
		return m, nil
	}

	switch key {
	case "q", "esc", "ctrl+c":
		m.persist()
		m.quitting = true
		return m, tea.Quit
	case "f", "F":
		m.doAction(m.deps.Pet.Feed(time.Now(), m.deps.Balance), "You feed. +30 Satiety!")
	case "p", "P":
		m.doAction(m.deps.Pet.Play(time.Now(), m.deps.Balance), "You play. +25 Happiness!")
	case "s", "S":
		if m.deps.Pet.Sleeping {
			m.doAction(m.deps.Pet.Wake(time.Now(), m.deps.Balance), "Wake up!")
		} else {
			m.doAction(m.deps.Pet.Sleep(time.Now(), m.deps.Balance), "Sleeping...")
		}
	case "r", "R":
		if m.deps.Pet.Dead {
			newPet := game.New(time.Now())
			m.deps.Pet = newPet
			m.deps.Storage.Save(m.deps.Fingerprint, newPet)
			m.setStatus("A new egg! Reborn.", false)
		}
	}
	return m, nil
}

func (m *Model) doAction(err error, successMsg string) {
	if err != nil {
		m.setStatus(err.Error(), true)
	} else {
		m.setStatus(successMsg, false)
	}
}

func (m *Model) setStatus(msg string, isErr bool) {
	m.statusMsg = msg
	m.statusErr = isErr
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
	if m.showReport {
		return m.viewReport()
	}
	return m.viewGame()
}

func (m *Model) viewReport() string {
	hours := m.deps.OfflineTime.Hours()

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205")).
		Render("🐣 While you were away...")

	text := fmt.Sprintf("%.1f hours passed.", hours)
	if hours > 24 {
		text = fmt.Sprintf("%.1f days passed.", hours/24)
	}

	p := m.deps.Pet
	durum := "Looks well"
	stil := lipgloss.NewStyle().Foreground(lipgloss.Color("42")) // yeşil

	if p.Dead {
		durum = "Died from hunger..."
		stil = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // kırmızı
	} else if p.Sick {
		durum = "Fell ill, it needs medication!"
		stil = lipgloss.NewStyle().Foreground(lipgloss.Color("208")) // turuncu
	} else if p.Satiety < 30 || p.Happiness < 30 {
		durum = "Very hungry and sad..."
		stil = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	}

	body := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		text,
		"",
		getPetAscii(p),
		"",
		stil.Render(durum),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Press any key to continue..."),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}

func (m *Model) viewGame() string {
	p := m.deps.Pet

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		Render("🐣 SSH Tamagotchi @ battos.dev")

	petVisual := lipgloss.NewStyle().
		Foreground(lipgloss.Color("212")).
		Render(getPetAscii(p))

	statStyle := lipgloss.NewStyle().Width(12)
	satStr := statStyle.Render("Satiety") + m.pSat.ViewAs(float64(p.Satiety)/100)
	hapStr := statStyle.Render("Happiness") + m.pHap.ViewAs(float64(p.Happiness)/100)
	eneStr := statStyle.Render("Energy") + m.pEne.ViewAs(float64(p.Energy)/100)

	stats := lipgloss.JoinVertical(lipgloss.Left, satStr, hapStr, eneStr)

	var actions []string
	if p.Dead {
		actions = append(actions, lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("[R] Revive"))
	} else if p.Sleeping {
		actions = append(actions, "[S] Wake up!")
	} else {
		if p.Satiety >= m.deps.Balance.Feed.MaxSatiety {
			actions = append(actions, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[F] Feed (Full)"))
		} else {
			actions = append(actions, "[F] Feed")
		}

		if p.Energy < m.deps.Balance.Play.MinEnergy {
			actions = append(actions, lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("[P] Play (Tired)"))
		} else {
			actions = append(actions, "[P] Play")
		}

		actions = append(actions, "[S] Sleep")
	}

	actionBar := lipgloss.JoinHorizontal(lipgloss.Left, actions...)

	statusLine := " "
	if m.statusMsg != "" {
		c := lipgloss.Color("42")
		if m.statusErr {
			c = lipgloss.Color("196")
		}
		statusLine = lipgloss.NewStyle().Foreground(c).Render(m.statusMsg)
	}

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

	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render(fmt.Sprintf("Last save: %s | Help: %s", saveInfo, help))

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		petVisual,
		"",
		stats,
		"",
		actionBar,
		"",
		statusLine,
		"",
		footer,
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
