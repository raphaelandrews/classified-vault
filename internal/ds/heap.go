package ds

type MaxHeap struct {
	items []heapItem
}

type heapItem struct {
	docID string
	score int
}

func NewMaxHeap() *MaxHeap {
	return &MaxHeap{
		items: make([]heapItem, 0),
	}
}

func (h *MaxHeap) Insert(docID string, score int) {
	h.items = append(h.items, heapItem{docID: docID, score: score})
	h.siftUp(len(h.items) - 1)
}

func (h *MaxHeap) ExtractMax() (string, int, bool) {
	if len(h.items) == 0 {
		return "", 0, false
	}

	max := h.items[0]
	last := len(h.items) - 1
	h.items[0] = h.items[last]
	h.items = h.items[:last]

	if len(h.items) > 0 {
		h.siftDown(0)
	}

	return max.docID, max.score, true
}

func (h *MaxHeap) Peek() (string, int, bool) {
	if len(h.items) == 0 {
		return "", 0, false
	}
	return h.items[0].docID, h.items[0].score, true
}

func (h *MaxHeap) Size() int {
	return len(h.items)
}

func (h *MaxHeap) TopN(n int) []struct {
	DocID string
	Score int
} {
	if n <= 0 {
		return nil
	}

	temp := make([]heapItem, len(h.items))
	copy(temp, h.items)
	backup := h.items
	h.items = temp

	var result []struct {
		DocID string
		Score int
	}
	for i := 0; i < n && h.Size() > 0; i++ {
		id, score, ok := h.ExtractMax()
		if !ok {
			break
		}
		result = append(result, struct {
			DocID string
			Score int
		}{id, score})
	}

	h.items = backup
	return result
}

func (h *MaxHeap) Clear() {
	h.items = h.items[:0]
}

func (h *MaxHeap) siftUp(idx int) {
	for idx > 0 {
		parent := (idx - 1) / 2
		if h.items[idx].score <= h.items[parent].score {
			break
		}
		h.items[idx], h.items[parent] = h.items[parent], h.items[idx]
		idx = parent
	}
}

func (h *MaxHeap) siftDown(idx int) {
	size := len(h.items)
	for {
		largest := idx
		left := 2*idx + 1
		right := 2*idx + 2

		if left < size && h.items[left].score > h.items[largest].score {
			largest = left
		}
		if right < size && h.items[right].score > h.items[largest].score {
			largest = right
		}

		if largest == idx {
			break
		}

		h.items[idx], h.items[largest] = h.items[largest], h.items[idx]
		idx = largest
	}
}
