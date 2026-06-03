package ds_test

import (
	"testing"

	"classified-vault/internal/ds"
)

func TestAVLTreeInsertAndQuery(t *testing.T) {
	tree := ds.NewAVLTree()

	tree.Insert(0, "doc1")
	tree.Insert(0, "doc2")
	tree.Insert(2, "doc3")
	tree.Insert(4, "doc4")

	result := tree.QueryUpTo(0)
	if len(result) != 2 {
		t.Errorf("expected 2 docs at level 0, got %d", len(result))
	}

	result = tree.QueryUpTo(2)
	if len(result) != 3 {
		t.Errorf("expected 3 docs at level ≤2, got %d", len(result))
	}

	result = tree.QueryUpTo(4)
	if len(result) != 4 {
		t.Errorf("expected 4 docs at level ≤4, got %d", len(result))
	}
}

func TestAVLTreeRemove(t *testing.T) {
	tree := ds.NewAVLTree()

	tree.Insert(2, "doc_a")
	tree.Insert(2, "doc_b")
	tree.Insert(3, "doc_c")

	if len(tree.QueryUpTo(4)) != 3 {
		t.Error("expected 3 docs initially")
	}

	tree.Remove(2, "doc_a")
	if len(tree.QueryUpTo(4)) != 2 {
		t.Error("expected 2 docs after removing doc_a")
	}

	tree.Remove(2, "doc_b")
	if len(tree.QueryUpTo(4)) != 1 {
		t.Error("expected 1 doc after removing doc_b")
	}

	result := tree.QueryUpTo(2)
	if len(result) != 0 {
		t.Errorf("level 2 node should be empty after removing all docs from it, got %d docs", len(result))
	}
}

func TestAVLTreeRebalance(t *testing.T) {
	tree := ds.NewAVLTree()

	for i := 0; i <= 4; i++ {
		tree.Insert(i, "test")
	}

	result := tree.QueryUpTo(4)
	if len(result) != 5 {
		t.Errorf("expected 5 docs, got %d", len(result))
	}
}
