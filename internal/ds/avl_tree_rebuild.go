package ds

func (t *AVLTree) RebuildIndex(docs []struct {
	ID             string
	Classification int
}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.Root = nil
	for _, doc := range docs {
		t.Root = insert(t.Root, doc.Classification, doc.ID)
	}
}
