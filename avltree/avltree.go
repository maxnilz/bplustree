package avltree

import (
	"bytes"
	"fmt"
	"io"
)

type node[T any] struct {
	value       T
	height      int
	left, right *node[T]
}

// insert inserts a value into the subtree rooted at this node
func (n *node[T]) insert(value T, less LessFunc[T]) (*node[T], bool) {
	if n == nil {
		return &node[T]{
			value:  value,
			height: 1, // new node is initially added at leaf
		}, true
	}

	var ok bool
	isEqual := true
	if less(value, n.value) {
		isEqual = false
		n.left, ok = n.left.insert(value, less)
	}
	if less(n.value, value) {
		isEqual = false
		n.right, ok = n.right.insert(value, less)
	}
	if isEqual {
		return n, false // skipping equal value
	}

	// update height
	n.height = max(height(n.left), height(n.right)) + 1

	bf := n.balanceFactor()

	// left-left case
	if bf < -1 && less(value, n.left.value) {
		return n.rightRotate(), ok
	}
	// right-right case
	if bf > 1 && less(n.right.value, value) {
		return n.leftRotate(), ok
	}
	// left-right case
	if bf < -1 && less(n.left.value, value) {
		//      z                               z                           x
		//     / \                            /   \                        /  \
		//    y   T4  Left Rotate (y)        x    T4  Right Rotate(z)    y      z
		//   / \      - - - - - - - - ->    /  \      - - - - - - - ->  / \    / \
		// T1   x                          y    T3                    T1  T2 T3  T4
		//     / \                        / \
		//   T2   T3                    T1   T2
		z, y := n, n.left
		z.left = y.leftRotate()
		return z.rightRotate(), ok
	}
	// right-left case
	if bf > 1 && less(value, n.right.value) {
		//    z                            z                            x
		//   / \                          / \                          /  \
		// T1   y   Right Rotate (y)    T1   x      Left Rotate(z)   z      y
		//     / \  - - - - - - - - ->     /  \   - - - - - - - ->  / \    / \
		//    x   T4                      T2   y                  T1  T2  T3  T4
		//   / \                              /  \
		// T2   T3                           T3   T4
		z, y := n, n.right
		z.right = y.rightRotate()
		return z.leftRotate(), ok
	}
	return n, ok
}

// remove removes a value from the subtree rooted at this node,
// return the new root node and an indicator that indicate whether
// the given value was found or not.
func (n *node[T]) remove(value T, less LessFunc[T]) (*node[T], bool) {
	if n == nil {
		return n, false
	}

	var ok bool
	isEqual := true
	if less(value, n.value) {
		isEqual = false
		n.left, ok = n.left.remove(value, less)
	}
	if less(n.value, value) {
		isEqual = false
		n.right, ok = n.right.remove(value, less)
	}
	if isEqual {
		r := n
		// no valid child
		if r.left == nil && r.right == nil {
			n = nil
			ok = true
		}
		// left child is valid
		if r.left != nil && r.right == nil {
			n = n.left
			ok = true
		}
		// right child is valid
		if r.left == nil && r.right != nil {
			n = n.right
			ok = true
		}
		// both left and right child are valid
		if r.left != nil && r.right != nil {
			cur := n // point left most node
			for cur.left != nil {
				cur = cur.left
			}
			n.value = cur.value
			n.right, ok = n.right.remove(cur.value, less)
		}
	}
	if n == nil {
		return nil, ok
	}

	// update height
	n.height = max(height(n.left), height(n.right)) + 1

	bf := n.balanceFactor()
	// left-left case
	if bf < -1 && n.left.balanceFactor() < 0 {
		return n.rightRotate(), ok
	}
	// right-right case
	if bf > 1 && n.right.balanceFactor() > 0 {
		return n.leftRotate(), ok
	}
	// left-right case
	if bf < -1 && n.left.balanceFactor() > 0 {
		//      z                               z                           x
		//     / \                            /   \                        /  \
		//    y   T4  Left Rotate (y)        x    T4  Right Rotate(z)    y      z
		//   / \      - - - - - - - - ->    /  \      - - - - - - - ->  / \    / \
		// T1   x                          y    T3                    T1  T2 T3  T4
		//     / \                        / \
		//   T2   T3                    T1   T2
		z, y := n, n.left
		z.left = y.leftRotate()
		return z.rightRotate(), ok
	}
	// right-left case
	if bf > 1 && n.right.balanceFactor() < 0 {
		//    z                            z                            x
		//   / \                          / \                          /  \
		// T1   y   Right Rotate (y)    T1   x      Left Rotate(z)   z      y
		//     / \  - - - - - - - - ->     /  \   - - - - - - - ->  / \    / \
		//    x   T4                      T2   y                  T1  T2  T3  T4
		//   / \                              /  \
		// T2   T3                           T3   T4
		z, y := n, n.right
		z.right = y.rightRotate()
		return z.leftRotate(), ok
	}
	return n, ok
}

func (n *node[T]) balanceFactor() int {
	if n == nil {
		return 0
	}
	return height(n.right) - height(n.left)
}

// leftRotate perform a left rotate on the
// tree rooted with this node itself return new root
// node after the rotation.
//      y                             x
//     / \                           / \
//    T1  x    --> leftRotate(y)    y  T3
//       / \                       / \
//      T2 T3                     T1 T2
func (n *node[T]) leftRotate() *node[T] {
	y, x := n, n.right
	t2 := x.left

	// rotate
	y.right, x.left = t2, y

	// update height
	x.height = max(height(x.left), height(x.right)) + 1
	y.height = max(height(y.left), height(y.right)) + 1

	return x
}

// rightRotate perform a right rotate on the
// tree rooted with this node itself return new root
// node after the rotation.
//      y                             x
//     / \                           / \
//    x  T3    --> rightRotate(y)   T1  y
//   / \                               / \
//  T1 T2                             T2 T3
func (n *node[T]) rightRotate() *node[T] {
	y, x := n, n.left
	t2 := x.right

	// rotate
	y.left, x.right = t2, y

	// update height
	x.height = max(height(x.left), height(x.right)) + 1
	y.height = max(height(y.left), height(y.right)) + 1

	return x
}

func (n *node[T]) print(w io.Writer) error {
	if n == nil {
		return nil
	}

	out := &bytes.Buffer{}
	out.WriteString(fmt.Sprintf("%v", n.value))
	leftPointer := "└──"
	if n.right != nil {
		leftPointer = "├──"
	}
	n.left.prettyPrint(out, "", leftPointer, n.right != nil)
	rightPointer := "└──"
	n.right.prettyPrint(out, "", rightPointer, false)

	out.WriteString("\n")

	if _, err := io.Copy(w, out); err != nil {
		return err
	}
	return nil
}

func (n *node[T]) prettyPrint(sb *bytes.Buffer, padding, pointer string, hasRightSibling bool) {
	if n == nil {
		return
	}
	sb.WriteString("\n")
	sb.WriteString(padding)
	sb.WriteString(pointer)
	sb.WriteString(fmt.Sprintf("%v", n.value))

	paddingBuilder := bytes.NewBufferString(padding)
	if hasRightSibling {
		paddingBuilder.WriteString("│  ")
	} else {
		paddingBuilder.WriteString("   ")
	}
	padding = paddingBuilder.String()

	leftPointer := "└──"
	if n.right != nil {
		leftPointer = "├──"
	}
	n.left.prettyPrint(sb, padding, leftPointer, n.right != nil)
	rightPointer := "└──"
	n.right.prettyPrint(sb, padding, rightPointer, false)
}

func height[T any](n *node[T]) int {
	if n == nil {
		return 0
	}
	return n.height
}

func max(a, b int) int {
	if a >= b {
		return a
	}
	return b
}

// LessFunc determines how to order a type 'T'.  It should implement a strict
// ordering, and should return true if within that ordering, 'a' < 'b'.
type LessFunc[T any] func(a, b T) bool

type AVLTree[T any] struct {
	less LessFunc[T]
	root *node[T]
}

func New[T any](less LessFunc[T]) *AVLTree[T] {
	return &AVLTree[T]{less: less}
}

func (a *AVLTree[T]) Insert(value T) bool {
	var ok bool
	a.root, ok = a.root.insert(value, a.less)
	return ok
}

func (a *AVLTree[T]) Remove(value T) (_ T, _ bool) {
	var found bool
	a.root, found = a.root.remove(value, a.less)
	if found {
		return value, true
	}
	return
}

func (a *AVLTree[T]) Print(w io.Writer) error {
	if a.root == nil {
		return nil
	}
	return a.root.print(w)
}
