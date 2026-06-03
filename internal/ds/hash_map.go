package ds

import "sync"

type Entry[V any] struct {
	Key   string
	Value V
	Next  *Entry[V]
}

type HashMap[V any] struct {
	buckets []*Entry[V]
	size    int
	count   int
	mu      sync.RWMutex
}

func NewHashMap[V any](size int) *HashMap[V] {
	return &HashMap[V]{
		buckets: make([]*Entry[V], size),
		size:    size,
	}
}

func (h *HashMap[V]) hash(key string) int {
	hash := 0
	for _, c := range key {
		hash = (hash*31 + int(c)) % h.size
	}
	return hash
}

func (h *HashMap[V]) Set(key string, value V) {
	h.mu.Lock()
	defer h.mu.Unlock()

	idx := h.hash(key)
	entry := &Entry[V]{Key: key, Value: value}

	if h.buckets[idx] == nil {
		h.buckets[idx] = entry
		h.count++
		return
	}

	current := h.buckets[idx]
	for {
		if current.Key == key {
			current.Value = value
			return
		}
		if current.Next == nil {
			current.Next = entry
			h.count++
			return
		}
		current = current.Next
	}
}

func (h *HashMap[V]) Get(key string) (V, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	idx := h.hash(key)
	current := h.buckets[idx]
	for current != nil {
		if current.Key == key {
			return current.Value, true
		}
		current = current.Next
	}
	var zero V
	return zero, false
}

func (h *HashMap[V]) Delete(key string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	idx := h.hash(key)
	current := h.buckets[idx]

	if current == nil {
		return false
	}

	if current.Key == key {
		h.buckets[idx] = current.Next
		h.count--
		return true
	}

	for current.Next != nil {
		if current.Next.Key == key {
			current.Next = current.Next.Next
			h.count--
			return true
		}
		current = current.Next
	}
	return false
}

func (h *HashMap[V]) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.count
}
