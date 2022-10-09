package bplustree

type KeyType int
type ValueType int

type Pair[Ty1, Ty2 any] struct {
	First  Ty1
	Second Ty2
}

type Node interface {
	Order() int
	Parent() Node
	SetParent(parent *Node) error
	IsRoot() bool
	IsLeaf() bool
	Size() int
	MinSize() int
	MaxSize() int
	FirstKey() (KeyType, error)
	String() string
}

type InternalNode struct {
	order    int
	parent   Node
	mappings Pair[KeyType, ValueType]
}

type LeafNode struct {
	order    int
	parent   Node
	mappings Pair[KeyType, ValueType]
	next     *LeafNode
}
