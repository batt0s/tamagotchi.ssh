package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
)

func testModel() Model {
	now := time.Now()
	return New(Deps{
		Fingerprint: "SHA256:0123456789abcdef",
		Pet: &game.Pet{
			Satiety:   80,
			Happiness: 80,
			Energy:    80,
			LastCalc:  now,
		},
		Balance: &game.Balance{
			Feed: game.FeedAction{
				Cooldown:   game.Duration(30 * time.Minute),
				Gain:       30,
				MaxSatiety: 90,
			},
			Play: game.PlayAction{
				Cooldown:      game.Duration(15 * time.Minute),
				EnergyCost:    15,
				SatietyCost:   5,
				HappinessGain: 25,
				MinEnergy:     20,
			},
		},
	}, 80, 24)
}

func TestFormatCooldown(t *testing.T) {
	tests := []struct {
		in   time.Duration
		want string
	}{
		{0, ""},
		{5 * time.Second, "5s"},
		{5 * time.Minute, "5m"},
		{2 * time.Hour, "2h"},
	}
	for _, tt := range tests {
		if got := formatCooldown(tt.in); got != tt.want {
			t.Errorf("formatCooldown(%v) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestActionBarShowsCooldowns(t *testing.T) {
	now := time.Now()
	m := testModel()
	m.deps.Pet.LastFed = now.Add(-time.Minute)
	m.deps.Pet.LastPlayed = now.Add(-time.Minute)

	bar := m.renderActions(m.deps.Pet, now)
	if !strings.Contains(bar, "Feed (") {
		t.Errorf("action bar missing feed cooldown: %q", bar)
	}
	if !strings.Contains(bar, "Play (") {
		t.Errorf("action bar missing play cooldown: %q", bar)
	}
}

func TestStatusExpires(t *testing.T) {
	m := testModel()
	m.setStatus("You feed.", false)
	m.statusUntil = time.Now().Add(-time.Second)

	updated, _ := m.Update(refreshMsg{})
	got := updated.(Model)
	if got.statusMsg != "" {
		t.Fatalf("status should have expired, got %q", got.statusMsg)
	}
}

func TestGameViewFitsTerminal(t *testing.T) {
	m := testModel()
	view := m.viewGame()
	if lines := strings.Count(view, "\n") + 1; lines > 24 {
		t.Fatalf("view is %d lines, want <= 24", lines)
	}
}

func TestReportViewFitsTerminal(t *testing.T) {
	m := testModel()
	m.deps.OfflineTime = 2 * time.Hour
	m.showReport = true
	view := m.viewReport()
	if lines := strings.Count(view, "\n") + 1; lines > 24 {
		t.Fatalf("report view is %d lines, want <= 24", lines)
	}
}
