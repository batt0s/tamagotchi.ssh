package tui

import (
	"github.com/batt0s/tamagotchi.ssh/internal/game"
	"github.com/charmbracelet/lipgloss"
)

type palette struct {
	title  lipgloss.Color
	accent lipgloss.Color
	border lipgloss.Color
	text   lipgloss.Color
	muted  lipgloss.Color

	healthy  lipgloss.Color
	sick     lipgloss.Color
	hungry   lipgloss.Color
	sleeping lipgloss.Color
	dead     lipgloss.Color

	success lipgloss.Color
	warning lipgloss.Color
	error   lipgloss.Color

	satietyA, satietyB     string
	happinessA, happinessB string
	energyA, energyB       string
	empty                  string
}

func newPalette(renderer *lipgloss.Renderer) palette {
	if renderer.HasDarkBackground() {
		return palette{
			title:  lipgloss.Color("205"),
			accent: lipgloss.Color("212"),
			border: lipgloss.Color("63"),
			text:   lipgloss.Color("252"),
			muted:  lipgloss.Color("241"),

			healthy:  lipgloss.Color("42"),
			sick:     lipgloss.Color("196"),
			hungry:   lipgloss.Color("208"),
			sleeping: lipgloss.Color("111"),
			dead:     lipgloss.Color("196"),

			success: lipgloss.Color("42"),
			warning: lipgloss.Color("208"),
			error:   lipgloss.Color("196"),

			satietyA: "#F5A97F", satietyB: "#ED8796",
			happinessA: "#C6A0F6", happinessB: "#F5BDE6",
			energyA: "#8BD5CA", energyB: "#7DC4E4",
			empty: "#4C4F69",
		}
	}

	return palette{
		title:  lipgloss.Color("126"),
		accent: lipgloss.Color("161"),
		border: lipgloss.Color("97"),
		text:   lipgloss.Color("238"),
		muted:  lipgloss.Color("244"),

		healthy:  lipgloss.Color("28"),
		sick:     lipgloss.Color("160"),
		hungry:   lipgloss.Color("166"),
		sleeping: lipgloss.Color("25"),
		dead:     lipgloss.Color("160"),

		success: lipgloss.Color("28"),
		warning: lipgloss.Color("166"),
		error:   lipgloss.Color("160"),

		satietyA: "#C35A00", satietyB: "#A61E4D",
		happinessA: "#7C3AED", happinessB: "#BE185D",
		energyA: "#0E7490", energyB: "#1D4ED8",
		empty: "#BCC0CC",
	}
}

func (p palette) petColor(pet *game.Pet) lipgloss.Color {
	switch {
	case pet.Dead:
		return p.dead
	case pet.Sick:
		return p.sick
	case pet.Sleeping:
		return p.sleeping
	case pet.Satiety < 20 || pet.Happiness < 20:
		return p.hungry
	default:
		return p.healthy
	}
}
