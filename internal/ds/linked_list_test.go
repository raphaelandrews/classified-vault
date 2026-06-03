package ds_test

import (
	"testing"

	"classified-vault/internal/ds"
)

func TestLinkedListAppend(t *testing.T) {
	l := ds.NewLinkedList[int]()
	if l.Size() != 0 {
		t.Error("expected size 0")
	}

	l.Append(1)
	l.Append(2)
	l.Append(3)

	if l.Size() != 3 {
		t.Errorf("expected size 3, got %d", l.Size())
	}
}

func TestLinkedListLastN(t *testing.T) {
	l := ds.NewLinkedList[int]()
	for i := 1; i <= 10; i++ {
		l.Append(i)
	}

	result := l.LastN(5)
	if len(result) != 5 {
		t.Errorf("expected 5, got %d", len(result))
	}
	expected := []int{6, 7, 8, 9, 10}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("position %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

func TestLinkedListLastNMoreThanSize(t *testing.T) {
	l := ds.NewLinkedList[int]()
	l.Append(1)
	l.Append(2)

	result := l.LastN(10)
	if len(result) != 2 {
		t.Errorf("expected 2, got %d", len(result))
	}
}

func TestLinkedListEmpty(t *testing.T) {
	l := ds.NewLinkedList[string]()
	result := l.LastN(5)
	if len(result) != 0 {
		t.Error("expected empty slice")
	}
}
