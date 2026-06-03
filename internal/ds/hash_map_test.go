package ds_test

import (
	"fmt"
	"testing"

	"classified-vault/internal/ds"
)

func TestHashMapSetGet(t *testing.T) {
	m := ds.NewHashMap[string](16)

	_, ok := m.Get("nonexistent")
	if ok {
		t.Error("expected false for missing key")
	}

	m.Set("hello", "world")
	val, ok := m.Get("hello")
	if !ok || val != "world" {
		t.Errorf("expected 'world', got '%s'", val)
	}
}

func TestHashMapOverwrite(t *testing.T) {
	m := ds.NewHashMap[int](8)
	m.Set("x", 1)
	m.Set("x", 2)
	val, _ := m.Get("x")
	if val != 2 {
		t.Errorf("expected 2, got %d", val)
	}
}

func TestHashMapDelete(t *testing.T) {
	m := ds.NewHashMap[string](8)
	m.Set("a", "first")
	m.Set("b", "second")

	if !m.Delete("a") {
		t.Error("expected true for delete existing key")
	}
	_, ok := m.Get("a")
	if ok {
		t.Error("key should be deleted")
	}

	if m.Delete("nonexistent") {
		t.Error("expected false for delete missing key")
	}

	_, ok = m.Get("b")
	if !ok {
		t.Error("key b should still exist")
	}
}

func TestHashMapCount(t *testing.T) {
	m := ds.NewHashMap[int](4)
	if m.Count() != 0 {
		t.Error("expected 0 initially")
	}
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	if m.Count() != 3 {
		t.Errorf("expected 3, got %d", m.Count())
	}
	m.Delete("a")
	if m.Count() != 2 {
		t.Errorf("expected 2, got %d", m.Count())
	}
}

func TestHashMapCollisions(t *testing.T) {
	m := ds.NewHashMap[string](2)
	keys := []string{"a", "b", "c", "d", "e"}
	for i, k := range keys {
		m.Set(k, fmt.Sprintf("v%d", i))
	}
	if m.Count() != len(keys) {
		t.Errorf("expected %d entries, got %d", len(keys), m.Count())
	}
	for i, k := range keys {
		val, ok := m.Get(k)
		if !ok {
			t.Errorf("key %s should exist", k)
		}
		if val != fmt.Sprintf("v%d", i) {
			t.Errorf("expected v%d, got %s", i, val)
		}
	}
}
