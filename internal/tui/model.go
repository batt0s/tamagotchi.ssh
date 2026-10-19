// Package tui is for Bubbletea
package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
	"github.com/batt0s/tamagotchi.ssh/internal/storage"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/harmonica"
	"github.com/charmbracelet/lipgloss"
)

const (
	saveInterval   = 15 * time.Second
	frameInterval  = time.Second / 30
	petFrameHeight = 7
)

type saveTickMsg struct{}
type refreshMsg struct{}
type frameMsg struct{}

type Deps struct {
	Fingerprint string
	Pet         *game.Pet
	Balance     *game.Balance
	Storage     storage.Storage
	OfflineTime time.Duration
	Renderer    *lipgloss.Renderer
}

type Model struct {
	deps        Deps
	width       int
	height      int
	quitting    bool
	lastSavedAt time.Time
	saveErr     error

	showReport  bool
	statusMsg   string
	statusErr   bool
	statusUntil time.Time

	palette    palette
	actionKind petAction

	petSpring  harmonica.Spring
	petXSpring harmonica.Spring
	petY       float64
	petV       float64
	petX       float64
	petXV      float64
	animFrame  int
	animating  bool

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
	if d.Renderer == nil {
		d.Renderer = lipgloss.DefaultRenderer()
	}
	pal := newPalette(d.Renderer)
	return Model{
		deps:       d,
		width:      width,
		height:     height,
		showReport: showReport,
		palette:    pal,
		petSpring:  harmonica.NewSpring(frameInterval.Seconds(), 12, 0.35),
		petXSpring: harmonica.NewSpring(frameInterval.Seconds(), 25, 0.15),
		pSat:       newProgress(d.Renderer, pal, pal.satietyA, pal.satietyB),
		pHap:       newProgress(d.Renderer, pal, pal.happinessA, pal.happinessB),
		pEne:       newProgress(d.Renderer, pal, pal.energyA, pal.energyB),
	}
}

func newProgress(renderer *lipgloss.Renderer, pal palette, colorA, colorB string) progress.Model {
	p := progress.New(
		progress.WithGradient(colorA, colorB),
		progress.WithWidth(30),
		progress.WithColorProfile(renderer.ColorProfile()),
	)
	p.EmptyColor = pal.empty
	p.PercentageStyle = renderer.NewStyle().Foreground(pal.muted)
	return p
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

func tickFrame() tea.Cmd {
	return tea.Tick(frameInterval, func(time.Time) tea.Msg { return frameMsg{} })
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
		now := time.Now()
		m.deps.Pet.Apply(now, m.deps.Balance)
		if !m.statusUntil.IsZero() && !now.Before(m.statusUntil) {
			m.statusMsg = ""
			m.statusErr = false
			m.statusUntil = time.Time{}
		}
		return m, tickEverySecond()
	case frameMsg:
		return m, m.updateAnimation()
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

	var cmd tea.Cmd

	switch key {
	case "q", "esc", "ctrl+c":
		m.persist()
		m.quitting = true
		return m, tea.Quit
	case "f", "F":
		if m.doAction(m.deps.Pet.Feed(time.Now(), m.deps.Balance), "You feed. +30 Satiety!") {
			m.actionKind = petActionFeed
			cmd = m.shake(50)
		}
	case "p", "P":
		if m.doAction(m.deps.Pet.Play(time.Now(), m.deps.Balance), "You play. +25 Happiness!") {
			m.actionKind = petActionPlay
			cmd = m.bounce(90)
		}
	case "s", "S":
		if m.deps.Pet.Sleeping {
			if m.doAction(m.deps.Pet.Wake(time.Now(), m.deps.Balance), "Wake up!") {
				m.actionKind = petActionPlay
				cmd = m.bounce(35)
			}
		} else {
			m.doAction(m.deps.Pet.Sleep(time.Now(), m.deps.Balance), "Sleeping...")
			m.actionKind = petActionNone
			// Drop to the ground, but leave the current animation loop alone so
			// a pending frame can finish it without starting a second loop.
			m.petY, m.petV = 0, 0
		}
	case "r", "R":
		if m.deps.Pet.Dead {
			newPet := game.New(time.Now())
			m.deps.Pet = newPet
			m.deps.Storage.Save(m.deps.Fingerprint, newPet)
			m.setStatus("A new egg! Reborn.", false)
			m.actionKind = petActionPlay
			cmd = m.bounce(60)
		}
	}
	return m, cmd
}

func (m *Model) doAction(err error, successMsg string) bool {
	if err != nil {
		m.setStatus(err.Error(), true)
		return false
	}
	m.setStatus(successMsg, false)
	return true
}

// bounce adds an upward impulse to the pet and starts the animation loop.
func (m *Model) bounce(velocity float64) tea.Cmd {
	m.petV += velocity
	return m.startAnimation()
}

// shake adds a horizontal impulse to the pet and starts the animation loop.
func (m *Model) shake(velocity float64) tea.Cmd {
	m.petXV += velocity
	return m.startAnimation()
}

func (m *Model) startAnimation() tea.Cmd {
	if m.animating {
		return nil
	}
	m.animating = true
	m.animFrame = 0
	return tickFrame()
}

func (m *Model) updateAnimation() tea.Cmd {
	m.petY, m.petV = m.petSpring.Update(m.petY, m.petV, 0)
	m.petX, m.petXV = m.petXSpring.Update(m.petX, m.petXV, 0)
	m.animFrame++

	settledY := math.Abs(m.petY) < 0.02 && math.Abs(m.petV) < 0.02
	settledX := math.Abs(m.petX) < 0.02 && math.Abs(m.petXV) < 0.02
	if (settledY && settledX) || m.animFrame > 120 {
		m.stopAnimation()
		return nil
	}
	return tickFrame()
}

func (m *Model) stopAnimation() {
	m.petY, m.petV = 0, 0
	m.petX, m.petXV = 0, 0
	m.animFrame = 0
	m.animating = false
	m.actionKind = petActionNone
}

func (m *Model) setStatus(msg string, isErr bool) {
	m.statusMsg = msg
	m.statusErr = isErr
	if msg == "" {
		m.statusUntil = time.Time{}
		return
	}
	m.statusUntil = time.Now().Add(3 * time.Second)
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

	title := m.deps.Renderer.NewStyle().Bold(true).Foreground(m.palette.title).
		Render("🐣 While you were away...")

	text := fmt.Sprintf("%.1f hours passed.", hours)
	if hours > 24 {
		text = fmt.Sprintf("%.1f days passed.", hours/24)
	}

	p := m.deps.Pet
	mood := "Looks well"
	if p.Dead {
		mood = "Died from hunger..."
	} else if p.Sick {
		mood = "Fell ill, it needs medication!"
	} else if p.Satiety < 30 || p.Happiness < 30 {
		mood = "Very hungry and sad..."
	}

	petColor := m.palette.petColor(p)
	body := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		m.deps.Renderer.NewStyle().Foreground(m.palette.text).Render(text),
		"",
		m.deps.Renderer.NewStyle().Foreground(petColor).Render(getPetAscii(p)),
		"",
		m.deps.Renderer.NewStyle().Foreground(petColor).Render(mood),
		"",
		m.muted("Press any key to continue..."),
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.panel(body))
}

func (m *Model) viewGame() string {
	p := m.deps.Pet

	title := m.deps.Renderer.NewStyle().Bold(true).Foreground(m.palette.title).
		Render("🐣 SSH Tamagotchi @ battos.dev")
	subtitle := lipgloss.JoinHorizontal(lipgloss.Left,
		m.muted("Playing as "+shorten(m.deps.Fingerprint)),
		m.muted(" · "),
		m.deps.Renderer.NewStyle().Foreground(m.palette.petColor(p)).Render(statusLine(p)),
	)

	rawPet := getPetAsciiFor(p, m.actionKind)
	petArt := renderPetArt(rawPet, int(math.Round(m.petY)), petFrameHeight)
	petArt = shiftPetArt(petArt, int(math.Round(m.petX)))
	petVisual := m.deps.Renderer.NewStyle().Foreground(m.palette.petColor(p)).Render(petArt)
	ground := m.muted(groundArt(petArtWidth(rawPet), int(math.Round(m.petY))))

	satiety := m.label("Satiety") + m.pSat.ViewAs(float64(p.Satiety)/100)
	happy := m.label("Happiness") + m.pHap.ViewAs(float64(p.Happiness)/100)
	energy := m.label("Energy") + m.pEne.ViewAs(float64(p.Energy)/100)
	stats := lipgloss.JoinVertical(lipgloss.Left, satiety, "", happy, "", energy)

	statusText := " "
	if m.statusMsg != "" {
		color := m.palette.success
		if m.statusErr {
			color = m.palette.error
		}
		statusText = m.deps.Renderer.NewStyle().Foreground(color).Render(m.statusMsg)
	}

	saveInfo := "not saved yet"
	if !m.lastSavedAt.IsZero() {
		saveInfo = fmt.Sprintf("%s ago", time.Since(m.lastSavedAt).Round(time.Second))
	}
	footer := m.muted(fmt.Sprintf("Save: %s | q / esc / ctrl+c → quit", saveInfo))
	if m.saveErr != nil {
		footer = m.deps.Renderer.NewStyle().Foreground(m.palette.error).
			Render("Save error: " + m.saveErr.Error())
	}

	body := lipgloss.JoinVertical(lipgloss.Left,
		title,
		subtitle,
		"",
		petVisual,
		ground,
		"",
		stats,
		"",
		m.renderActions(p, time.Now()),
		"",
		statusText,
		footer,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, m.panel(body))
}

func (m *Model) renderActions(p *game.Pet, now time.Time) string {
	var actions []string

	switch {
	case p.Dead:
		actions = append(actions, m.action("R", "Revive", true))
	case p.Sleeping:
		actions = append(actions, m.action("S", "Wake up!", true))
	default:
		feedLeft := cooldownRemaining(now, p.LastFed, m.deps.Balance.Feed.Cooldown.Std())
		feedLabel := "Feed"
		feedEnabled := true
		switch {
		case p.Satiety >= m.deps.Balance.Feed.MaxSatiety:
			feedLabel = "Feed (Full)"
			feedEnabled = false
		case feedLeft > 0:
			feedLabel = "Feed (" + formatCooldown(feedLeft) + ")"
			feedEnabled = false
		}
		actions = append(actions, m.action("F", feedLabel, feedEnabled))

		playLeft := cooldownRemaining(now, p.LastPlayed, m.deps.Balance.Play.Cooldown.Std())
		playLabel := "Play"
		playEnabled := true
		switch {
		case p.Energy < m.deps.Balance.Play.MinEnergy:
			playLabel = "Play (Tired)"
			playEnabled = false
		case playLeft > 0:
			playLabel = "Play (" + formatCooldown(playLeft) + ")"
			playEnabled = false
		}
		actions = append(actions, m.action("P", playLabel, playEnabled))
		actions = append(actions, m.action("S", "Sleep", true))
	}

	return strings.Join(actions, "   ")
}

func (m *Model) action(key, label string, enabled bool) string {
	keyColor := m.palette.accent
	labelColor := m.palette.text
	if !enabled {
		keyColor = m.palette.muted
		labelColor = m.palette.muted
	}

	keyStyle := m.deps.Renderer.NewStyle().Bold(true).Foreground(keyColor)
	labelStyle := m.deps.Renderer.NewStyle().Foreground(labelColor)
	return keyStyle.Render("["+key+"]") + " " + labelStyle.Render(label)
}

func (m *Model) label(s string) string {
	return m.deps.Renderer.NewStyle().Foreground(m.palette.muted).Width(12).Render(s)
}

func (m *Model) muted(s string) string {
	return m.deps.Renderer.NewStyle().Foreground(m.palette.muted).Render(s)
}

func (m *Model) panel(body string) string {
	return m.deps.Renderer.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.palette.border).
		Padding(0, 2).
		Render(body)
}

func cooldownRemaining(now, last time.Time, cooldown time.Duration) time.Duration {
	if last.IsZero() || cooldown <= 0 {
		return 0
	}
	remaining := cooldown - now.Sub(last)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func formatCooldown(remaining time.Duration) string {
	if remaining <= 0 {
		return ""
	}
	remaining = remaining.Round(time.Second)
	switch {
	case remaining < time.Minute:
		return fmt.Sprintf("%ds", int(remaining.Seconds()))
	case remaining < time.Hour:
		return fmt.Sprintf("%dm", int(math.Ceil(remaining.Minutes())))
	default:
		return fmt.Sprintf("%dh", int(math.Ceil(remaining.Hours())))
	}
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
	renderer *lipgloss.Renderer
}

func NewMessage(text string, width, height int, renderer *lipgloss.Renderer) *MessageModel {
	if width <= 0 {
		width = 80
	}
	if height <= 0 {
		height = 24
	}
	if renderer == nil {
		renderer = lipgloss.DefaultRenderer()
	}
	return &MessageModel{text: text, width: width, height: height, renderer: renderer}
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
		m.renderer.NewStyle().Foreground(lipgloss.Color("241")).
			Render("enter / q  →  exit"),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}
