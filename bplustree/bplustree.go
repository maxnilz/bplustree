package bplustree

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"sort"
)

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

// front returns the fist item or nil if it is empty.
func (s *items[T]) front() (_ T) {
	if len(*s) == 0 {
		return
	}
	return (*s)[0]
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

func (n *node[kT, vT]) maxKeys() int {
	if !n.isLeaf {
		return n.order - 1
	}
	return n.order
}

func (n *node[kT, vT]) minKeys() int {
	degree := int(math.Ceil(float64(n.order) / 2.0))
	if !n.isLeaf {
		return degree - 1
	}
	return degree
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
	newNode.order = n.order
	newNode.parent = n.parent
	newNode.isLeaf = n.isLeaf
	if len(n.values) > 0 {
		newNode.values = append(newNode.values, n.values[i:]...)
		n.values.truncate(i)
	}
	if n.next != nil {
		n.next.prev = newNode
	}
	newNode.prev = n
	newNode.next = n.next
	n.next = newNode
	return key, newNode
}

// insert inserts a key-value pair into the subtree rooted at this node,
// making sure no nodes in the subtree exceed order-1 keys. it will replace the value
// if the given key existed already and return false to indicate that no new key is
// inserted, otherwise, return the true and newly created node if a split happens.
func (n *node[kT, vT]) insert(key kT, value vT, less LessFunc[kT]) (*node[kT, vT], bool) {
	if n.isLeaf {
		return n.insertIntoLeaf(key, value, less)
	}
	i, _ := n.keys.find(key, less)
	return n.children[i].insert(key, value, less)
}

func (n *node[kT, vT]) insertIntoLeaf(key kT, value vT, less LessFunc[kT]) (*node[kT, vT], bool) {
	index, found := n.keys.find(key, less)
	if found {
		n.values[index] = value
		return nil, false
	}
	n.keys.insertAt(index, key)
	n.values.insertAt(index, value)
	return n.mayGrowUp(less)
}

func (n *node[kT, vT]) mayGrowUp(less LessFunc[kT]) (*node[kT, vT], bool) {
	if len(n.keys) <= n.maxKeys() {
		return nil, false
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
		return root, true
	}

	index, _ := parent.keys.find(promotedKey, less)
	parent.keys.insertAt(index, promotedKey)
	parent.children.insertAt(index+1, newNode)
	return parent.mayGrowUp(less)
}

// remove removes an item from the subtree rooted at this node.
// if no key found in the leaf node of the subtree, return false, otherwise, remove
// it from leaf node, then the stop node(if merge happens), removed value and true.
func (n *node[kT, vT]) remove(key kT, less LessFunc[kT]) (_ *node[kT, vT], _ vT, _ bool) {
	if n.isLeaf {
		return n.removeFromLeaf(key, less)
	}
	i, _ := n.keys.find(key, less)
	return n.children[i].remove(key, less)
}

func (n *node[kT, vT]) removeFromLeaf(key kT, less LessFunc[kT]) (stopAt *node[kT, vT], out vT, found bool) {
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
	stopAt, _ = n.mayMergeWithNeighbor(key, less)
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

func (n *node[kT, vT]) mayMergeWithNeighbor(key kT, less LessFunc[kT]) (*node[kT, vT], bool) {
	first, second := n.prev, n
	if n.prev == nil {
		first, second = n, n.next
	}
	if len(first.keys)+len(second.keys) > first.maxKeys() {
		return n, false
	}

	first.keys = append(first.keys, second.keys...)
	first.children = append(first.children, second.children...)
	first.values = append(first.values, second.values...)
	first.next = second.next
	if second.next != nil {
		second.next.prev = first
	}

	parent := n.parent
	i, _ := parent.keys.find(key, less)
	if i == len(parent.keys) {
		i = i - 1
	}
	parent.keys.removeAt(i)
	parent.children.removeAt(i + 1)
	if len(parent.keys) == 0 {
		first.parent = nil
		parent = nil
		return first, true
	}
	stopAt, _ := parent.mayMergeWithNeighbor(key, less)
	return stopAt, true
}

func (n *node[kT, vT]) print(w io.Writer) error {
	if n == nil {
		return nil
	}
	q := items[*node[kT, vT]]{}
	q = append(q, n)
	out := &bytes.Buffer{}
	for len(q) > 0 {
		cnt := len(q)
		for ; cnt > 0; cnt-- {
			out.WriteString("| ")
			a := q.removeAt(0)
			for i, key := range a.keys {
				if a.isLeaf {
					out.WriteString(fmt.Sprintf("%v-%v ", key, a.values[i]))
					continue
				}
				out.WriteString(fmt.Sprintf("%v ", key))
			}
			out.WriteString("|")
			q = append(q, a.children...)
		}
		out.WriteString("\n")
	}
	if _, err := io.Copy(w, out); err != nil {
		return err
	}
	return nil
}

// LessFunc determines how to order a type 'T'.  It should implement a strict
// ordering, and should return true if within that ordering, 'a' < 'b'.
type LessFunc[T any] func(a, b T) bool

type BPlusTree[kT, vT any] struct {
	order int
	less  LessFunc[kT]
	root  *node[kT, vT]
}

func New[kT, vT any](order int, less LessFunc[kT]) *BPlusTree[kT, vT] {
	return &BPlusTree[kT, vT]{order: order, less: less}
}

func (t *BPlusTree[kT, vT]) Insert(key kT, value vT) bool {
	if t.root == nil {
		t.root = &node[kT, vT]{order: t.order, isLeaf: true}
		t.root.keys = append(t.root.keys, key)
		t.root.values = append(t.root.values, value)
		return false
	}
	root, found := t.root.insert(key, value, t.less)
	if root != nil {
		t.root = root
	}
	return found
}

func (t *BPlusTree[kT, vT]) Remove(key kT) (_ vT, _ bool) {
	if t.root == nil {
		return
	}
	stopNode, out, found := t.root.remove(key, t.less)
	if stopNode != nil && stopNode.parent == nil {
		t.root = stopNode
	}
	return out, found
}

func (t *BPlusTree[kt, vT]) Print(w io.Writer) error {
	if t.root == nil {
		return nil
	}
	return t.root.print(w)
}
