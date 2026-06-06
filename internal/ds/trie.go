package ds

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	docID    string
}

type Trie struct {
	root *TrieNode
	size int
}

func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{children: make(map[rune]*TrieNode)},
	}
}

func (t *Trie) Insert(word string, docID string) {
	node := t.root
	for _, ch := range word {
		if _, ok := node.children[ch]; !ok {
			node.children[ch] = &TrieNode{children: make(map[rune]*TrieNode)}
		}
		node = node.children[ch]
	}
	if !node.isEnd {
		node.isEnd = true
		node.docID = docID
		t.size++
	}
}

func (t *Trie) Search(prefix string) []string {
	node := t.root
	for _, ch := range prefix {
		if _, ok := node.children[ch]; !ok {
			return nil
		}
		node = node.children[ch]
	}

	var results []string
	t.collect(node, &results)
	return results
}

func (t *Trie) SearchWithIDs(prefix string) []struct {
	Word  string
	DocID string
} {
	node := t.root
	for _, ch := range prefix {
		if _, ok := node.children[ch]; !ok {
			return nil
		}
		node = node.children[ch]
	}

	var results []struct {
		Word  string
		DocID string
	}
	t.collectWithIDs(node, prefix, &results)
	return results
}

func (t *Trie) collect(node *TrieNode, results *[]string) {
	if node.isEnd {
		*results = append(*results, node.docID)
	}
	for _, child := range node.children {
		t.collect(child, results)
	}
}

func (t *Trie) collectWithIDs(node *TrieNode, prefix string, results *[]struct {
	Word  string
	DocID string
}) {
	if node.isEnd {
		*results = append(*results, struct {
			Word  string
			DocID string
		}{prefix, node.docID})
	}
	for ch, child := range node.children {
		t.collectWithIDs(child, prefix+string(ch), results)
	}
}

func (t *Trie) Size() int {
	return t.size
}

func (t *Trie) Clear() {
	t.root = &TrieNode{children: make(map[rune]*TrieNode)}
	t.size = 0
}
