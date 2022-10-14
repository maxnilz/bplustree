package bplustree

import "sort"

// items stores items in a node.
type items[T any] []T

// insertAt inserts a value into the given index, pushing all subsequent values
// forward.
func (s *items[T]) insertAt(index int, item T) {
	var zero T
	*s = append(*s, zero)
	if index < len(*s) {
		copy((*s)[index+1:], (*s)[index:])
	}
	(*s)[index] = item
}

// removeAt removes a value at a given index, pulling all subsequent values
// back.
func (s *items[T]) removeAt(index int) T {
	item := (*s)[index]
	copy((*s)[index:], (*s)[index+1:])
	var zero T
	(*s)[len(*s)-1] = zero
	*s = (*s)[:len(*s)-1]
	return item
}

// pop removes and returns the last element in the list.
func (s *items[T]) pop() (out T) {
	index := len(*s) - 1
	out = (*s)[index]
	var zero T
	(*s)[index] = zero
	*s = (*s)[:index]
	return
}

// truncate truncates this instance at index so that it contains only the
// first index items. index must be less than or equal to length.
func (s *items[T]) truncate(index int) {
	var toClear items[T]
	*s, toClear = (*s)[:index], (*s)[index:]
	var zero T
	for i := 0; i < len(toClear); i++ {
		toClear[i] = zero
	}
}

// find returns the index where the given item should be inserted into this
// list.  'found' is true if the item already exists in the list at the given
// index.
func (s items[T]) find(item T, less func(T, T) bool) (index int, found bool) {
	i := sort.Search(len(s), func(i int) bool {
		return less(item, s[i])
	})
	if i > 0 && !less(s[i-1], item) {
		return i - 1, true
	}
	return i, false
}

// node representing a node in the B+ tree.
// This type is general enough to serve for both the
// leaf and the internal node.
//
// In a leaf, children would be nil and the index of
// each key equals the index of its corresponding values,
// with a maximum of order-1 key-value pairs.
//
// In an internal node, the first child refers to lower
// nodes with keys less than the smallest key in the keys
// array. Then, with indices i starting at 0, the children
// at i+1 points to the subtree with keys greater than or
// equal to the key in this node at index i.
// It must at all times maintain the invariant when
//   * len(children) == 0, len(keys) unconstrained
//   * len(children) == len(keys) + 1
type node[kT, vT any] struct {
	keys     items[kT]
	children items[*node[kT, vT]]
	parent   *node[kT, vT]

	order int
	next  *node[kT, vT]
	prev  *node[kT, vT]

	// leaf only
	isLeaf bool
	values items[vT]
}

func (n *node[kT, vT]) degree() int {
	return (n.order + 1) / 2
}

func (n *node[kT, vT]) maxKeys() int {
	return n.order - 1
}

func (n *node[kT, vT]) minKeys() int {
	return n.degree() - 1
}

// split splits the given node at the given index. The current node shrinks,
// if the current node is an internal node, it returns the key that existed
// at that index and a new node containing all keys/children after the given
// index, otherwise, it's a leaf node, it returns the key that existed at that
// index and a new node containing all keys/values at & after the given index.
func (n *node[kT, vT]) split(i int) (kT, *node[kT, vT]) {
	key := n.keys[i]
	newNode := &node[kT, vT]{}
	ik := i + 1
	if n.isLeaf {
		ik = i
	}
	newNode.keys = append(newNode.keys, n.keys[ik:]...)
	n.keys.truncate(i)
	if len(n.children) > 0 {
		newNode.children = append(newNode.children, n.children[i+1:]...)
		n.children.truncate(i + 1)
	}
	newNode.parent = n.parent
	newNode.isLeaf = n.isLeaf
	if len(n.values) > 0 {
		newNode.values = append(newNode.values, n.values[i:]...)
	}
	n.next.prev = newNode
	newNode.prev = n
	newNode.next = n.next
	n.next = newNode
	return key, newNode
}

// insert inserts a key-value pair into the subtree rooted at this node,
// making sure no nodes in the subtree exceed order-1 keys. If an equivalent
// key be found/replaced by insert, it will be returned.
func (n *node[kT, vT]) insert(key kT, value vT, less LessFunc[kT]) bool {
	if n.isLeaf {
		return n.insertIntoLeaf(key, value, less)
	}
	i, _ := n.keys.find(key, less)
	return n.children[i].insert(key, value, less)
}

func (n *node[kT, vT]) insertIntoLeaf(key kT, value vT, less LessFunc[kT]) (found bool) {
	var index int
	index, found = n.keys.find(key, less)
	if found {
		n.values[index] = value
		return
	}
	n.keys.insertAt(index, key)
	n.values.insertAt(index, value)
	n.mayGrowUp(less)
	return
}

func (n *node[kT, vT]) mayGrowUp(less LessFunc[kT]) bool {
	if len(n.keys) <= n.maxKeys() {
		return false
	}
	promotedKey, newNode := n.split(n.minKeys())
	parent := n.parent
	if parent == nil {
		root := &node[kT, vT]{
			order: n.order,
		}
		root.keys = append(root.keys, promotedKey)
		root.children = append(root.children, n, newNode)
		n.parent = root
		newNode.parent = root
		return true
	}

	index, found := parent.keys.find(promotedKey, less)
	if found {
		panic("unexpected key found at parent")
	}
	promotedKey, newNode = parent.split(index)
	parent.keys.insertAt(index, promotedKey)
	parent.children.insertAt(index, newNode)
	return parent.mayGrowUp(less)
}

// remove removes an item from the subtree rooted at this node.
func (n *node[kT, vT]) remove(key kT, less LessFunc[kT]) (_ vT, _ bool) {
	if n.isLeaf {
		return n.removeFromLeaf(key, less)
	}
	i, _ := n.keys.find(key, less)
	return n.children[i].remove(key, less)
}

func (n *node[kT, vT]) removeFromLeaf(key kT, less LessFunc[kT]) (out vT, found bool) {
	var index int
	index, found = n.keys.find(key, less)
	if !found {
		return
	}
	n.keys.removeAt(index)
	out = n.values.removeAt(index)
	if n.parent == nil || len(n.keys) >= n.minKeys() {
		return // still valid after the removal, return directly
	}
	if n.mayStealFromNeighborLeaf(key, less) {
		return
	}
	n.mayMergeWithNeighbor(key, less)
	return
}

func (n *node[kT, vT]) mayStealFromNeighborLeaf(key kT, less LessFunc[kT]) bool {
	if !n.isLeaf {
		panic("unexpected steal operation")
	}
	prev, next := n.prev, n.next
	if prev != nil && len(prev.keys) > prev.minKeys() {
		stolenKey := prev.keys.pop()
		stolenValue := prev.values.pop()
		n.keys.insertAt(0, stolenKey)
		n.values.insertAt(0, stolenValue)
		parent := n.parent
		index, _ := parent.keys.find(key, less)
		parent.keys.removeAt(index)
		parent.keys.insertAt(index, stolenKey)
		return true
	}
	if next != nil && len(next.keys) > next.minKeys() {
		stolenKey := next.keys.removeAt(0)
		stolenValue := next.values.removeAt(0)
		n.keys = append(n.keys, stolenKey)
		n.values = append(n.values, stolenValue)

		shiftUpKey := next.keys[0]
		parent := n.parent
		index, _ := parent.keys.find(key, less)
		parent.keys.removeAt(index)
		parent.keys.insertAt(index, shiftUpKey)
		return true
	}
	return false
}

func (n *node[kT, vT]) mayMergeWithNeighbor(key kT, less LessFunc[kT]) bool {
	parent := n.parent
	first, second := n.prev, n
	if n.prev == nil {
		first, second = n, n.next
	}
	if len(first.keys)+len(second.keys) > first.maxKeys() {
		return false
	}

	first.keys = append(first.keys, second.keys...)
	first.children = append(first.children, second.children...)
	first.values = append(first.values, first.values...)

	i, _ := n.keys.find(key, less)
	parent.keys.removeAt(i)
	parent.children.removeAt(i + 1)
	parent.mayMergeWithNeighbor(key, less)

	return true
}

// LessFunc determines how to order a type 'T'.  It should implement a strict
// ordering, and should return true if within that ordering, 'a' < 'b'.
type LessFunc[T any] func(a, b T) bool
