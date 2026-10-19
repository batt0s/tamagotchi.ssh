package tui

import (
	"strings"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
)

func getPetAscii(p *game.Pet) string {
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

	const margin = 2
	width += margin * 2
	if offset < -margin {
		offset = -margin
	}
	if offset > margin {
		offset = margin
	}

	padLeft := margin + offset
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
