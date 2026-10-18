package game

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var (
	ErrDead      = errors.New("It's Dead")
	ErrSleeping  = errors.New("I'm Sleeping")
	ErrAwake     = errors.New("I'm Awake")
	ErrTooFull   = errors.New("I'm Full")
	ErrTooHungry = errors.New("I'm Hungry")
	ErrTooTired  = errors.New("I'm Tired")
	ErrCooldown  = errors.New("Not Ready Yet")
)

type Pet struct {
	Satiety   int `json:"satiety"`
	Happiness int `json:"happiness"`
	Energy    int `json:"energy"`

	Sleeping bool `json:"sleeping"`
	Sick     bool `json:"sick"`
	Dead     bool `json:"dead"`

	ZeroHours float64 `json:"zero_hours"`
	SickHours float64 `json:"sick_hours"`

	CreatedAt  time.Time `json:"created_at"`
	LastCalc   time.Time `json:"last_calc"`
	LastFed    time.Time `json:"last_fed"`
	LastPlayed time.Time `json:"last_played"`
}

func New(now time.Time) *Pet {
	return &Pet{
		Satiety:   80,
		Happiness: 80,
		Energy:    80,
		CreatedAt: now,
		LastCalc:  now,
	}
}

// Aplly changes from LastCalc to now
func (p *Pet) Apply(now time.Time, b *Balance) {
	if p.Dead {
		p.LastCalc = now
		return
	}
	if p.LastCalc.IsZero() || p.LastCalc.After(now) {
		p.LastCalc = now
		return
	}
	elapsed := now.Sub(p.LastCalc)
	if elapsed <= 0 {
		p.LastCalc = now
		return
	}
	hours := elapsed.Hours()
	wasSick := p.Sick

	satRate := b.Decay.SatietyPerHourAwake
	energyRate := b.Decay.EnergyPerHourAwake
	if p.Sleeping {
		satRate = b.Decay.SatietyPerHourSleeping
		energyRate = b.Decay.EnergyPerHourSleeping
	}
	hapRate := b.Decay.HappinessPerHour
	if wasSick {
		hapRate *= b.Modifiers.SickHappinessMultiplier
	}

	// Zero hours: not full OR not happy
	setZeroIn := min(timeUntilZero(float64(p.Satiety), satRate), hours)
	hapZeroIn := min(timeUntilZero(float64(p.Happiness), hapRate), hours)
	firstZero := math.Min(setZeroIn, hapZeroIn)
	if firstZero < hours {
		p.ZeroHours += hours - firstZero
	}

	// Apply decay
	p.Satiety = clampF(float64(p.Satiety)-satRate*hours, 0, 100)
	p.Happiness = clampF(float64(p.Happiness)-hapRate*hours, 0, 100)
	p.Energy = clampF(float64(p.Energy)-energyRate*hours, 0, 100)

	// If both stats are positive, reset timers
	if p.Satiety > 0 && p.Happiness > 0 {
		p.ZeroHours = 0
		if p.Sick {
			p.Sick = false
			p.SickHours = 0
		}
	}

	// Is Sick?
	if !p.Sick && p.ZeroHours >= b.Thresholds.SickAfterZeroHours {
		p.Sick = true
	}

	// Sick Hours
	switch {
	case wasSick && p.Sick:
		p.SickHours += hours
	case !wasSick && p.Sick:
		p.SickHours = p.ZeroHours - b.Thresholds.SickAfterZeroHours
	}

	// Death (noooooo)
	if p.Sick && p.SickHours >= b.Thresholds.DeathAfterSickHours {
		p.Dead = true
	}

	p.LastCalc = now
}

func (p *Pet) Feed(now time.Time, b *Balance) error {
	p.Apply(now, b)
	if p.Dead {
		return ErrDead
	}
	if p.Sleeping {
		return ErrSleeping
	}
	if p.Satiety >= b.Feed.MaxSatiety {
		return ErrTooFull
	}
	if !p.LastFed.IsZero() {
		if remaining := b.Feed.Cooldown.Std() - now.Sub(p.LastFed); remaining > 0 {
			return fmt.Errorf("%w (%s seconds left)", ErrCooldown, remaining.Round(time.Second))
		}
	}
	p.Satiety = min(p.Satiety+b.Feed.Gain, 100)
	p.LastFed = now
	p.recoverIfHealthy()
	return nil
}

func (p *Pet) Play(now time.Time, b *Balance) error {
	p.Apply(now, b)
	if p.Dead {
		return ErrDead
	}
	if p.Sleeping {
		return ErrSleeping
	}
	if p.Energy < b.Play.MinEnergy {
		return fmt.Errorf("%w (%d/%d)", ErrTooTired, p.Energy, b.Play.MinEnergy)
	}
	if !p.LastPlayed.IsZero() {
		if remaining := b.Play.Cooldown.Std() - now.Sub(p.LastPlayed); remaining > 0 {
			return fmt.Errorf("%w (%s seconds left)", ErrCooldown, remaining.Round(time.Second))
		}
	}
	p.Energy = max(0, p.Energy-b.Play.EnergyCost)
	p.Satiety = max(0, p.Satiety-b.Play.SatietyCost)
	p.Happiness = max(p.Happiness+b.Play.HappinessGain, 100)
	p.LastPlayed = now
	p.recoverIfHealthy()
	return nil
}

func (p *Pet) Sleep(now time.Time, b *Balance) error {
	p.Apply(now, b)
	if p.Dead {
		return ErrDead
	}
	if p.Sleeping {
		return ErrSleeping
	}
	p.Sleeping = true
	return nil
}

func (p *Pet) Wake(now time.Time, b *Balance) error {
	p.Apply(now, b)
	if p.Dead {
		return ErrDead
	}
	p.Sleeping = false
	return nil
}

func timeUntilZero(current, ratePerHour float64) float64 {
	if current <= 0 {
		return 0
	}
	if ratePerHour <= 0 {
		return math.Inf(1)
	}
	return current / ratePerHour
}

func clampF(v, lo, hi float64) int {
	return int(math.Max(lo, math.Min(hi, math.Round(v))))
}

func (p *Pet) recoverIfHealthy() {
	if p.Satiety > 0 && p.Happiness > 0 {
		p.ZeroHours = 0
		if p.Sick {
			p.Sick = false
			p.SickHours = 0
		}
	}
}
