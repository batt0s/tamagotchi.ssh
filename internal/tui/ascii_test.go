package tui

import (
	"strings"
	"testing"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
)

func TestGetPetAsciiMood(t *testing.T) {
	tests := []struct {
		name string
		pet  *game.Pet
		face string
	}{
		{name: "dead", pet: &game.Pet{Dead: true}, face: "RIP"},
		{name: "sick", pet: &game.Pet{Sick: true}, face: "( x_x )"},
		{name: "sleeping", pet: &game.Pet{Sleeping: true}, face: "( -_- ) zZ"},
		{name: "hungry", pet: &game.Pet{Satiety: 10, Happiness: 50}, face: "( T_T )"},
		{name: "sad", pet: &game.Pet{Satiety: 50, Happiness: 10}, face: "( T_T )"},
		{name: "healthy", pet: &game.Pet{Satiety: 80, Happiness: 80}, face: "( ^_^ )"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPetAscii(tt.pet); !strings.Contains(got, tt.face) {
				t.Errorf("getPetAscii() = %q, want face %q", got, tt.face)
			}
		})
	}
}

func TestGetPetAsciiActionFaces(t *testing.T) {
	pet := &game.Pet{Satiety: 80, Happiness: 80}

	feed := getPetAsciiFor(pet, petActionFeed)
	if !strings.Contains(feed, "( o_o )") {
		t.Errorf("feed face = %q, want chewing face", feed)
	}

	play := getPetAsciiFor(pet, petActionPlay)
	if !strings.Contains(play, "( ^o^ )") || !strings.Contains(play, `\|   |/`) {
		t.Errorf("play face = %q, want happy jumping pose", play)
	}
}

func TestPetAsciiPriority(t *testing.T) {
	pet := &game.Pet{Dead: true, Sick: true, Sleeping: true}

	if got := getPetAsciiFor(pet, petActionPlay); !strings.Contains(got, "RIP") {
		t.Errorf("dead pet should win over actions: %q", got)
	}
}
