package tui

import "github.com/batt0s/tamagotchi.ssh/internal/game"

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
