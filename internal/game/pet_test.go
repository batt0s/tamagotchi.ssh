package game

import (
	"errors"
	"testing"
	"time"
)

var t0 = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

func testBalance() *Balance {
	return &Balance{
		Decay: Decay{
			SatietyPerHourAwake:    3.0,
			SatietyPerHourSleeping: 1.5,
			HappinessPerHour:       2.0,
			EnergyPerHourAwake:     2.0,
			EnergyPerHourSleeping:  -15.0,
		},
		Thresholds: Thresholds{SickAfterZeroHours: 6, DeathAfterSickHours: 24},
		Modifiers:  Modifiers{SickHappinessMultiplier: 2.0},
		Feed:       FeedAction{Cooldown: Duration(30 * time.Minute), Gain: 30, MaxSatiety: 90},
		Play:       PlayAction{Cooldown: Duration(15 * time.Minute), EnergyCost: 15, SatietyCost: 5, HappinessGain: 25, MinEnergy: 20},
	}
}

func TestApplyDecay(t *testing.T) {
	b := testBalance()
	p := &Pet{Satiety: 60, Happiness: 80, Energy: 100, LastCalc: t0}

	p.Apply(t0.Add(2*time.Hour), b)

	if p.Satiety != 54 { // 60 - 3*2
		t.Errorf("satiety = %d, want 54", p.Satiety)
	}
	if p.Happiness != 76 { // 80 - 2*2
		t.Errorf("happiness = %d, want 76", p.Happiness)
	}
	if p.Energy != 96 { // 100 - 2*2
		t.Errorf("energy = %d, want 96", p.Energy)
	}
}

func TestFeedCooldown(t *testing.T) {
	b := testBalance()
	p := &Pet{Satiety: 40, Happiness: 50, Energy: 100, LastCalc: t0}
	p.Apply(t0, b)

	if err := p.Feed(t0, b); err != nil {
		t.Fatalf("ilk feed: %v", err)
	}
	if err := p.Feed(t0.Add(5*time.Minute), b); !errors.Is(err, ErrCooldown) {
		t.Errorf("cooldown sırasında err = %v, want ErrCooldown", err)
	}
	if err := p.Feed(t0.Add(31*time.Minute), b); err != nil {
		t.Errorf("cooldown sonrası feed: %v", err)
	}
}

func TestFeedRejectsWhenFull(t *testing.T) {
	b := testBalance()
	p := &Pet{Satiety: 95, Happiness: 50, Energy: 100, LastCalc: t0}
	if err := p.Feed(t0, b); !errors.Is(err, ErrTooFull) {
		t.Errorf("tokken feed err = %v, want ErrTooFull", err)
	}
}

func TestPlayRequiresEnergy(t *testing.T) {
	b := testBalance()
	p := &Pet{Satiety: 50, Happiness: 50, Energy: 10, LastCalc: t0}
	if err := p.Play(t0, b); !errors.Is(err, ErrTooTired) {
		t.Errorf("enerji yetersizken err = %v, want ErrTooTired", err)
	}
}

func TestSickAfterSixHoursAtZero(t *testing.T) {
	b := testBalance()
	p := &Pet{Satiety: 3, Happiness: 100, Energy: 100, LastCalc: t0}
	// Satiety 3/3 = 1h'de sıfıra iner, sonra 7h daha 0'da kalır = 7h zero
	p.Apply(t0.Add(8*time.Hour), b)

	if !p.Sick {
		t.Fatalf("8h sonra hasta olmalıydı; ZeroHours=%.2f", p.ZeroHours)
	}
	if p.SickHours <= 0 {
		t.Errorf("SickHours = %.2f, > 0 olmalı", p.SickHours)
	}
}

func TestDeathAfterTwentyFourHoursSick(t *testing.T) {
	b := testBalance()
	p := &Pet{Satiety: 0, Happiness: 0, Energy: 100, LastCalc: t0}
	// 6h eşik + 24h ölüm = 30h; biraz marj bırak
	p.Apply(t0.Add(40*time.Hour), b)

	if !p.Dead {
		t.Fatalf("40h sonra ölmüş olmalıydı; SickHours=%.2f", p.SickHours)
	}
}

func TestSleepRegeneratesEnergy(t *testing.T) {
	b := testBalance()
	p := &Pet{Satiety: 80, Happiness: 80, Energy: 20, LastCalc: t0}
	if err := p.Sleep(t0, b); err != nil {
		t.Fatalf("sleep: %v", err)
	}
	p.Apply(t0.Add(2*time.Hour), b)
	// 2h * 15 = +30 → 50
	if p.Energy != 50 {
		t.Errorf("energy = %d, want 50", p.Energy)
	}
	if err := p.Feed(t0.Add(2*time.Hour), b); !errors.Is(err, ErrSleeping) {
		t.Errorf("uykuda feed err = %v, want ErrSleeping", err)
	}
}

func TestRecoveryCuresSick(t *testing.T) {
	b := testBalance()
	// 0/0'da 10h geçir → hasta olmalı, sonra besleyip oynayarak iyileştir
	p := &Pet{Satiety: 0, Happiness: 0, Energy: 100, LastCalc: t0}
	p.Apply(t0.Add(10*time.Hour), b)
	if !p.Sick {
		t.Fatal("önce hasta olmalı")
	}

	// Besle + oyna → her iki stat > 0 → iyileşme
	t1 := t0.Add(10 * time.Hour)
	if err := p.Feed(t1, b); err != nil {
		t.Fatalf("feed: %v", err)
	}
	if err := p.Play(t1, b); err != nil {
		t.Fatalf("play: %v", err)
	}
	if p.Sick {
		t.Errorf("besle+oyna sonrası sick=false olmalı")
	}
	if p.ZeroHours != 0 || p.SickHours != 0 {
		t.Errorf("sayaçlar sıfırlanmalı: ZeroHours=%.2f SickHours=%.2f", p.ZeroHours, p.SickHours)
	}
}
