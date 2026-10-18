// Session manager
package session

import "sync"

type Manager struct {
	mu     sync.Mutex
	active map[string]struct{}
}

func NewManager() *Manager {
	return &Manager{active: make(map[string]struct{})}
}

// TryAcquire locks the fingerprint
func (m *Manager) TryAcquire(fp string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.active[fp]; ok {
		return false
	}
	m.active[fp] = struct{}{}
	return true
}

func (m *Manager) Release(fp string) {
	m.mu.Lock()
	delete(m.active, fp)
	m.mu.Unlock()
}

func (m *Manager) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.active)
}
