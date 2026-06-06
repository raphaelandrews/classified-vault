package ds

import (
	"sync"
	"time"
)

type lruEntry struct {
	key     string
	value   interface{}
	expires time.Time
}

type LRUCache struct {
	capacity int
	ttl      time.Duration
	entries  map[string]*lruEntry
	order    *LinkedList[string]
	mu       sync.Mutex
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		entries:  make(map[string]*lruEntry),
		order:    NewLinkedList[string](),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.entries[key]
	if !ok {
		return nil, false
	}

	if time.Now().After(entry.expires) {
		delete(c.entries, key)
		c.removeFromOrder(key)
		return nil, false
	}

	c.moveToFront(key)
	return entry.value, true
}

func (c *LRUCache) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.entries[key]; ok {
		existing.value = value
		existing.expires = time.Now().Add(c.ttl)
		c.moveToFront(key)
		return
	}

	if len(c.entries) >= c.capacity {
		oldest := c.order.Head
		if oldest != nil {
			delete(c.entries, oldest.Value)
			c.removeFromOrder(oldest.Value)
		}
	}

	c.entries[key] = &lruEntry{
		key:     key,
		value:   value,
		expires: time.Now().Add(c.ttl),
	}
	c.order.Append(key)
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
	c.removeFromOrder(key)
}

func (c *LRUCache) Keys() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.order.LastN(c.order.Size())
}

func (c *LRUCache) Size() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.entries)
}

func (c *LRUCache) moveToFront(key string) {
	c.removeFromOrder(key)
	c.order.Append(key)
}

func (c *LRUCache) removeFromOrder(key string) {
	node := c.order.Head
	for node != nil {
		if node.Value == key {
			if node.Prev != nil {
				node.Prev.Next = node.Next
			} else {
				c.order.Head = node.Next
			}
			if node.Next != nil {
				node.Next.Prev = node.Prev
			} else {
				c.order.Tail = node.Prev
			}
			return
		}
		node = node.Next
	}
}
