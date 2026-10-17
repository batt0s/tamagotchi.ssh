package game

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Balance struct {
	Decay      Decay      `yaml:"decay"`
	Thresholds Thresholds `yaml:"thresholds"`
	Modifiers  Modifiers  `yaml:"modifiers"`
	Feed       FeedAction `yaml:"feed"`
	Play       PlayAction `yaml:"play"`
}

type Decay struct {
	SatietyPerHourAwake    float64 `yaml:"satiety_per_hour_awake"`
	SatietyPerHourSleeping float64 `yaml:"satiety_per_hour_sleeping"`
	HappinessPerHour       float64 `yaml:"happiness_per_hour"`
	EnergyPerHourAwake     float64 `yaml:"energy_per_hour_awake"`
	EnergyPerHourSleeping  float64 `yaml:"energy_per_hour_sleeping"`
}

type Thresholds struct {
	SickAfterZeroHours  float64 `yaml:"sick_after_zero_hours"`
	DeathAfterSickHours float64 `yaml:"death_after_sick_hours"`
}

type Modifiers struct {
	SickHappinessMultiplier float64 `yaml:"sick_happiness_multipler"`
}

type FeedAction struct {
	Cooldown   Duration `yaml:"cooldown"`
	Gain       int      `yaml:"gain"`
	MaxSatiety int      `yaml:"max_satiety"`
}

type PlayAction struct {
	Cooldown      Duration `yaml:"cooldown"`
	EnergyCost    int      `yaml:"energy_cost"`
	SatietyCost   int      `yaml:"satiety_cost"`
	HappinessGain int      `yaml:"happiness_gain"`
	MinEnergy     int      `yaml:"min_energy"`
}

type Duration time.Duration

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	var s string
	if err := value.Decode(&s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return fmt.Errorf("Couldn't unmarshal duration %q: %w", s, err)
	}
	*d = Duration(parsed)
	return nil
}

func (d Duration) Std() time.Duration { return time.Duration(d) }

func LoadBalance(path string) (*Balance, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Couldn't load Balance: %w", err)
	}
	var b Balance
	if err := yaml.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("Couldn't unmarshal Balance: %w", err)
	}
	return &b, nil
}
