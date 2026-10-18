package tui

import (
	"testing"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
)

func Test_getPetAscii(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		p    *game.Pet
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPetAscii(tt.p)
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("getPetAscii() = %v, want %v", got, tt.want)
			}
		})
	}
}
