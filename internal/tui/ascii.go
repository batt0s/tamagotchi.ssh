package tui

import (
	"strings"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
	"github.com/charmbracelet/lipgloss"
)

type petAction int

const (
	petActionNone petAction = iota
	petActionFeed
	petActionPlay
)

const petArtMargin = 2

func getPetAscii(p *game.Pet) string {
	return getPetAsciiFor(p, petActionNone)
}

func getPetAsciiFor(p *game.Pet, action petAction) string {
	if p.Dead {
		return `
   .---.
  /     \
 | RIP   |
 |       |
 |_______|
`
	}
	if p.Sick {
		return `
   .---.
  ( x_x )
  /|   |\
   |___|
`
	}
	if p.Sleeping {
		return `
   .---.
  ( -_- ) zZ
  /|   |\
   |___|
`
	}
	if action == petActionPlay {
		return `
   .---.
  ( ^o^ )
  \|   |/
   |___|
`
	}
	if action == petActionFeed {
		return `
   .---.
  ( o_o )
  /|   |\
   |___|
`
	}
	if p.Satiety < 20 || p.Happiness < 20 {
		return `
   .---.
  ( T_T )
  /|   |\
   |___|
`
	}
	return `
   .---.
  ( ^_^ )
  /|   |\
   |___|
`
}

// renderPetArt places the pet inside a fixed-height area. offset raises the
// pet within that area without changing the overall layout height.
func renderPetArt(art string, offset, height int) string {
	lines := strings.Split(strings.Trim(art, "\n"), "\n")
	if len(lines) >= height {
		return strings.Join(lines, "\n")
	}

	maxOffset := height - len(lines)
	if offset < 0 {
		offset = 0
	}
	if offset > maxOffset {
		offset = maxOffset
	}

	padTop := maxOffset - offset
	return strings.Repeat("\n", padTop) + strings.Join(lines, "\n") + strings.Repeat("\n", offset)
}

// shiftPetArt moves the pet horizontally inside a fixed-width area so a shake
// doesn't change the layout width.
func shiftPetArt(art string, offset int) string {
	lines := strings.Split(art, "\n")
	width := 0
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}

	width += petArtMargin * 2
	if offset < -petArtMargin {
		offset = -petArtMargin
	}
	if offset > petArtMargin {
		offset = petArtMargin
	}

	padLeft := petArtMargin + offset
	var b strings.Builder
	for i, line := range lines {
		if i > 0 {
			b.WriteByte('\n')
		}
		if line == "" {
			continue
		}
		b.WriteString(strings.Repeat(" ", padLeft))
		b.WriteString(line)
		b.WriteString(strings.Repeat(" ", width-padLeft-len(line)))
	}
	return b.String()
}

func petArtWidth(art string) int {
	width := 0
	for _, line := range strings.Split(strings.Trim(art, "\n"), "\n") {
		if len(line) > width {
			width = len(line)
		}
	}
	return width + petArtMargin*2
}

func groundArt(width, jump int) string {
	shadowWidth := 9 - jump
	if shadowWidth < 4 {
		shadowWidth = 4
	}
	return lipgloss.PlaceHorizontal(width, lipgloss.Center, strings.Repeat("─", shadowWidth))
}
