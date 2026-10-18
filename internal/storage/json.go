package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/batt0s/tamagotchi.ssh/internal/game"
)

type JSONStorage struct {
	dir string
	mu  sync.Mutex
}

func NewJSONStorage(dir string) (*JSONStorage, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("couldn't create the Storage Directory: %w", err)
	}
	return &JSONStorage{dir: dir}, nil
}

func (s *JSONStorage) Load(fp string) (*game.Pet, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path(fp))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("read error: %w", err)
	}
	var p game.Pet
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}
	return &p, nil
}

func (s *JSONStorage) Save(fp string, p *game.Pet) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("save marshall error: %w", err)
	}

	path := s.path(fp)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("tmp write error: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename error: %w", err)
	}
	return nil
}

func (s *JSONStorage) LoadOrCreate(fp string, now time.Time) (*game.Pet, error) {
	p, err := s.Load(fp)
	if err == nil || !errors.Is(err, ErrNotFound) {
		return p, err
	}
	p = game.New(now)
	if err := s.Save(fp, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *JSONStorage) path(fp string) string {
	h := sha256.Sum256([]byte(fp))
	return filepath.Join(s.dir, hex.EncodeToString(h[:])+".json")
}
