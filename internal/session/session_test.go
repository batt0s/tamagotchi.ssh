package session

import "testing"

func TestManagerTryAcquire(t *testing.T) {
	m := NewManager()
	if !m.TryAcquire("a") {
		t.Fatal("ilk acquire başarısız")
	}
	if m.TryAcquire("a") {
		t.Fatal("ikinci acquire reddedilmeliydi")
	}
	if !m.TryAcquire("b") {
		t.Fatal("farklı fp acquire başarısız")
	}
	if m.Len() != 2 {
		t.Errorf("Len = %d, want 2", m.Len())
	}

	m.Release("a")
	if !m.TryAcquire("a") {
		t.Fatal("release sonrası acquire başarısız")
	}
}
