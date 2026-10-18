// Package storage is for session matched pet storage
package storage

import (
	"errors"
	"time"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
)

var ErrNotFound = errors.New("couldn't found Pet")

type Storage interface {
	Load(fp string) (*game.Pet, error)
	Save(fp string, p *game.Pet) error
	LoadOrCreate(fp string, now time.Time) (*game.Pet, error)
}
