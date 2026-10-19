package tui

import (
	"strings"
	"testing"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
)

const testPetArt = `
   .---.
  ( ^_^ )
  /|   |\
   |___|
`

func TestRenderPetArtKeepsFixedHeight(t *testing.T) {
	for _, offset := range []int{-2, 0, 1, 2, 3, 10} {
		got := renderPetArt(testPetArt, offset, petFrameHeight)
		if lines := strings.Count(got, "\n") + 1; lines != petFrameHeight {
			t.Errorf("offset %d: got %d lines, want %d", offset, lines, petFrameHeight)
		}
	}
}

func TestRenderPetArtRaisesPet(t *testing.T) {
	grounded := renderPetArt(testPetArt, 0, petFrameHeight)
	raised := renderPetArt(testPetArt, 2, petFrameHeight)

	if !strings.HasPrefix(grounded, "\n\n\n") {
		t.Errorf("grounded pet should have top padding: %q", grounded)
	}
	if !strings.HasSuffix(raised, "\n\n") {
		t.Errorf("raised pet should have bottom padding: %q", raised)
	}
}

func TestShiftPetArtKeepsFixedWidth(t *testing.T) {
	width := -1
	for _, offset := range []int{-5, -2, 0, 2, 5} {
		for _, line := range strings.Split(shiftPetArt(testPetArt, offset), "\n") {
			if line == "" {
				continue
			}
			if width == -1 {
				width = len(line)
			}
			if len(line) != width {
				t.Fatalf("offset %d: line width %d, want %d", offset, len(line), width)
			}
		}
	}
}

func TestShiftPetArtMovesHorizontally(t *testing.T) {
	leading := func(art string) int {
		for _, line := range strings.Split(art, "\n") {
			if strings.Contains(line, ".---.") {
				return len(line) - len(strings.TrimLeft(line, " "))
			}
		}
		return -1
	}

	left := leading(shiftPetArt(testPetArt, -1))
	center := leading(shiftPetArt(testPetArt, 0))
	right := leading(shiftPetArt(testPetArt, 1))
	if left != center-1 || right != center+1 {
		t.Fatalf("leading spaces left=%d center=%d right=%d", left, center, right)
	}
}

func TestBounceStartsSingleAnimationLoop(t *testing.T) {
	m := New(Deps{Pet: &game.Pet{}, Balance: &game.Balance{}}, 80, 24)

	if cmd := m.bounce(55); cmd == nil || !m.animating {
		t.Fatal("bounce should start the animation loop")
	}
	if cmd := m.bounce(10); cmd != nil {
		t.Fatal("bounce should not start a second animation loop")
	}
}

func TestAnimationSettles(t *testing.T) {
	for name, trigger := range map[string]func(*Model){
		"bounce": func(m *Model) { m.bounce(90) },
		"shake":  func(m *Model) { m.shake(50) },
	} {
		t.Run(name, func(t *testing.T) {
			m := New(Deps{Pet: &game.Pet{}, Balance: &game.Balance{}}, 80, 24)
			trigger(&m)

			frames := 0
			for {
				cmd := m.updateAnimation()
				frames++
				if cmd == nil {
					break
				}
				if frames > 200 {
					t.Fatal("animation did not settle")
				}
			}

			if m.animating || m.petY != 0 || m.petV != 0 || m.petX != 0 || m.petXV != 0 {
				t.Fatalf("animation not reset: animating=%v y=%v v=%v x=%v xv=%v",
					m.animating, m.petY, m.petV, m.petX, m.petXV)
			}
			if frames < 10 {
				t.Fatalf("animation settled too quickly: %d frames", frames)
			}
		})
	}
}
