package ds

import "sync"

type ListNode[T any] struct {
	Value T
	Next  *ListNode[T]
	Prev  *ListNode[T]
}

type LinkedList[T any] struct {
	Head *ListNode[T]
	Tail *ListNode[T]
	size int
	mu   sync.RWMutex
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

func (l *LinkedList[T]) Append(value T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	node := &ListNode[T]{Value: value}
	if l.Tail == nil {
		l.Head = node
		l.Tail = node
	} else {
		node.Prev = l.Tail
		l.Tail.Next = node
		l.Tail = node
	}
	l.size++
}

func (l *LinkedList[T]) LastN(n int) []T {
	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]T, 0, n)
	current := l.Tail
	for current != nil && len(result) < n {
		result = append([]T{current.Value}, result...)
		current = current.Prev
	}
	return result
}

func (l *LinkedList[T]) Size() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.size
}
