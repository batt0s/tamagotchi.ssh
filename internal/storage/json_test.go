package storage_test

import (
	"errors"
	"testing"
	"time"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
	"github.com/batt0s/tamagotchi.ssh/internal/storage"
)

func TestJSONStorageRoundTrip(t *testing.T) {
	s, err := storage.NewJSONStorage(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}

	p := &game.Pet{Satiety: 42, Happiness: 77, Energy: 88}
	if err := s.Save("SHA256:abc", p); err != nil {
		t.Fatal(err)
	}

	got, err := s.Load("SHA256:abc")
	if err != nil {
		t.Fatal(err)
	}
	if got.Satiety != 42 || got.Happiness != 77 || got.Energy != 88 {
		t.Errorf("got %+v", got)
	}
}

func TestJSONStorageNotFound(t *testing.T) {
	s, _ := storage.NewJSONStorage(t.TempDir())
	_, err := s.Load("SHA256:none")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestJSONStorageLoadOrCreate(t *testing.T) {
	s, _ := storage.NewJSONStorage(t.TempDir())
	now := time.Now()

	p1, err := s.LoadOrCreate("SHA256:x", now)
	if err != nil {
		t.Fatal(err)
	}
	if p1.Satiety != 80 {
		t.Errorf("yeni pet satiety = %d, want 80", p1.Satiety)
	}

	p1.Satiety = 30
	if err := s.Save("SHA256:x", p1); err != nil {
		t.Fatal(err)
	}

	p2, err := s.LoadOrCreate("SHA256:x", now)
	if err != nil {
		t.Fatal(err)
	}
	if p2.Satiety != 30 {
		t.Errorf("mevcut pet satiety = %d, want 30", p2.Satiety)
	}
}

func TestJSONStorageSeparateFiles(t *testing.T) {
	s, _ := storage.NewJSONStorage(t.TempDir())
	_ = s.Save("SHA256:a", &game.Pet{Satiety: 10})
	_ = s.Save("SHA256:b", &game.Pet{Satiety: 20})

	a, _ := s.Load("SHA256:a")
	b, _ := s.Load("SHA256:b")
	if a.Satiety != 10 || b.Satiety != 20 {
		t.Errorf("izolasyon bozuk: a=%d b=%d", a.Satiety, b.Satiety)
	}
}
