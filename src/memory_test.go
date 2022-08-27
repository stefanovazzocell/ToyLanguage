package interpreter

import (
	"bytes"
	"testing"
)

/*
* Tests
**/

func TestMemory(t *testing.T) {
	m := NewMemory()
	// Check setup
	if m == nil {
		t.Fatal("Failed to setup memory: null pointer")
	}
	bts := m.Bytes()
	if len(bts) != 0 {
		t.Fatalf("Got %+d instead of [] for the inital memory", bts)
	}

	// Check boundaries
	if m.Prev() == nil {
		t.Fatal("[0] Didn't stop at boundary")
	}
	if m.p != 0 {
		t.Fatalf("[0] Incorrectly set pointer to %d", m.p)
	}
	m.p = MemSize - 1
	if m.Next() == nil {
		t.Fatal("[1] Didn't stop at boundary")
	}
	if m.p != MemSize-1 {
		t.Fatalf("[1] Incorrectly set pointer to %d", m.p)
	}

	// Set and Get operations
	m.Reset()
	if m.Get() != 0 {
		t.Fatalf("Got %d instead of blank for read after reset", m.Get())
	}
	m.Set(10)
	if m.Get() != 10 {
		t.Fatalf("Got %d instead of 10", m.Get())
	}
	m.Set(255)
	if m.Get() != 255 {
		t.Fatalf("Got %d instead of 255", m.Get())
	}
	m.Set(0)
	if m.Get() != 0 {
		t.Fatalf("Got %d instead of 0", m.Get())
	}

	// Write and move
	if m.Next() != nil {
		t.Fatal("Failed to move")
	}
	m.Set(10)
	if m.Next() != nil {
		t.Fatal("Failed to move")
	}
	m.Set(3)
	m.Decr()
	if m.Prev() != nil {
		t.Fatal("Failed to move")
	}
	m.Set(0)
	m.Incr()
	bts = m.Bytes()
	if !bytes.Equal(bts, []byte{0, 1, 2}) {
		t.Fatalf("Got %+d instead of [0, 1, 2]", bts)
	}
}

/*
* Benchmarks
**/

func BenchmarkMemory(b *testing.B) {
	b.Run("Set", func(b *testing.B) {
		m := NewMemory()
		b.ResetTimer()
		for i := b.N - 1; i >= 0; i-- {
			m.Set(byte(i))
		}
	})
	b.Run("Get", func(b *testing.B) {
		m := NewMemory()
		b.ResetTimer()
		for i := b.N - 1; i >= 0; i-- {
			m.Get()
		}
	})
	b.Run("Next", func(b *testing.B) {
		m := NewMemory()
		b.ResetTimer()
		for i := b.N - 1; i >= 0; i-- {
			if i%(MemSize+1) == 0 {
				m.p = 0
			}
			m.Next()
		}
	})
	b.Run("Prev", func(b *testing.B) {
		m := NewMemory()
		b.ResetTimer()
		for i := b.N - 1; i >= 0; i-- {
			m.Prev()
			if i%(MemSize+1) == 0 {
				m.p = 0
			}
		}
	})
}
