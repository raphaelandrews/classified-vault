package ds

import "sync"

type AVLNode struct {
	Key    int
	DocIDs []string
	Height int
	Left   *AVLNode
	Right  *AVLNode
}

type AVLTree struct {
	Root *AVLNode
	mu   sync.RWMutex
}

func NewAVLTree() *AVLTree {
	return &AVLTree{}
}

func height(n *AVLNode) int {
	if n == nil {
		return 0
	}
	return n.Height
}

func balanceFactor(n *AVLNode) int {
	if n == nil {
		return 0
	}
	return height(n.Left) - height(n.Right)
}

func updateHeight(n *AVLNode) {
	lh := height(n.Left)
	rh := height(n.Right)
	if lh > rh {
		n.Height = lh + 1
	} else {
		n.Height = rh + 1
	}
}

func rotateRight(y *AVLNode) *AVLNode {
	x := y.Left
	t := x.Right
	x.Right = y
	y.Left = t
	updateHeight(y)
	updateHeight(x)
	return x
}

func rotateLeft(x *AVLNode) *AVLNode {
	y := x.Right
	t := y.Left
	y.Left = x
	x.Right = t
	updateHeight(x)
	updateHeight(y)
	return y
}

func insert(node *AVLNode, key int, docID string) *AVLNode {
	if node == nil {
		return &AVLNode{Key: key, DocIDs: []string{docID}, Height: 1}
	}

	if key < node.Key {
		node.Left = insert(node.Left, key, docID)
	} else if key > node.Key {
		node.Right = insert(node.Right, key, docID)
	} else {
		node.DocIDs = append(node.DocIDs, docID)
		return node
	}

	updateHeight(node)
	bf := balanceFactor(node)

	if bf > 1 && key < node.Left.Key {
		return rotateRight(node)
	}
	if bf < -1 && key > node.Right.Key {
		return rotateLeft(node)
	}
	if bf > 1 && key > node.Left.Key {
		node.Left = rotateLeft(node.Left)
		return rotateRight(node)
	}
	if bf < -1 && key < node.Right.Key {
		node.Right = rotateRight(node.Right)
		return rotateLeft(node)
	}

	return node
}

func (t *AVLTree) Insert(clearanceLevel int, docID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Root = insert(t.Root, clearanceLevel, docID)
}

func (t *AVLTree) QueryUpTo(maxLevel int) []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var result []string
	var traverse func(node *AVLNode)
	traverse = func(node *AVLNode) {
		if node == nil {
			return
		}
		if node.Key <= maxLevel {
			traverse(node.Left)
			result = append(result, node.DocIDs...)
			traverse(node.Right)
		} else {
			traverse(node.Left)
		}
	}
	traverse(t.Root)
	return result
}

func (t *AVLTree) Remove(clearanceLevel int, docID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Root = remove(t.Root, clearanceLevel, docID)
}

func remove(node *AVLNode, key int, docID string) *AVLNode {
	if node == nil {
		return nil
	}

	if key < node.Key {
		node.Left = remove(node.Left, key, docID)
	} else if key > node.Key {
		node.Right = remove(node.Right, key, docID)
	} else {
		newDocIDs := make([]string, 0)
		for _, id := range node.DocIDs {
			if id != docID {
				newDocIDs = append(newDocIDs, id)
			}
		}
		node.DocIDs = newDocIDs

		if len(node.DocIDs) > 0 {
			return node
		}

		if node.Left == nil && node.Right == nil {
			return nil
		}
		if node.Left == nil {
			return node.Right
		}
		if node.Right == nil {
			return node.Left
		}

		succ := minNode(node.Right)
		node.Key = succ.Key
		node.DocIDs = succ.DocIDs
		node.Right = remove(node.Right, succ.Key, "")
	}
	return rebalance(node)
}

func minNode(node *AVLNode) *AVLNode {
	for node.Left != nil {
		node = node.Left
	}
	return node
}

func rebalance(node *AVLNode) *AVLNode {
	if node == nil {
		return nil
	}

	updateHeight(node)
	bf := balanceFactor(node)

	if bf > 1 {
		if balanceFactor(node.Left) < 0 {
			node.Left = rotateLeft(node.Left)
		}
		return rotateRight(node)
	}
	if bf < -1 {
		if balanceFactor(node.Right) > 0 {
			node.Right = rotateRight(node.Right)
		}
		return rotateLeft(node)
	}
	return node
}
