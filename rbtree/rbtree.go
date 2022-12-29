package rbtree

import (
	"bytes"
	"fmt"
	"io"
)

type direction int

const maxHeight = 128

const (
	leftDir direction = iota
	rightDir
)

type color int

func (c color) String() string {
	if c == red {
		return "red"
	}
	return "black"
}

const (
	red color = iota
	black
)

type node[T any] struct {
	data     T
	color    color
	children *children[T]
}

func (n *node[T]) get(dir direction) *node[T] {
	return n.children.get(dir)
}

func (n *node[T]) set(dir direction, node *node[T]) {
	n.children.set(dir, node)
}

func (n *node[T]) left() *node[T] {
	return n.children.get(leftDir)
}

func (n *node[T]) setLeft(node *node[T]) {
	n.children.set(leftDir, node)
}

func (n *node[T]) right() *node[T] {
	return n.children.get(rightDir)
}

func (n *node[T]) setRight(node *node[T]) {
	n.children.set(rightDir, node)
}

func (n *node[T]) print(w io.Writer) error {
	if n == nil {
		return nil
	}

	out := &bytes.Buffer{}
	out.WriteString(fmt.Sprintf("(%v %v)", n.data, n.color))
	leftPointer := "└──"
	if n.right() != nil {
		leftPointer = "├──"
	}
	n.left().prettyPrint(out, "", leftPointer, n.right() != nil)
	rightPointer := "└──"
	n.right().prettyPrint(out, "", rightPointer, false)

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
	sb.WriteString(fmt.Sprintf("(%v %v)", n.data, n.color))

	paddingBuilder := bytes.NewBufferString(padding)
	if hasRightSibling {
		paddingBuilder.WriteString("│  ")
	} else {
		paddingBuilder.WriteString("   ")
	}
	padding = paddingBuilder.String()

	leftPointer := "└──"
	if n.right() != nil {
		leftPointer = "├──"
	}
	n.left().prettyPrint(sb, padding, leftPointer, n.right() != nil)
	rightPointer := "└──"
	n.right().prettyPrint(sb, padding, rightPointer, false)
}

type children[T any] []*node[T]

func newChildren[T any]() *children[T] {
	nodes := make([]*node[T], rightDir+1)
	out := children[T](nodes)
	return &out
}

func (c *children[T]) set(dir direction, n *node[T]) {
	i := int(dir)
	(*c)[i] = n
}

func (c *children[T]) get(dir direction) *node[T] {
	i := int(dir)
	return (*c)[i]
}

func (c *children[T]) left() *node[T] {
	return c.get(leftDir)
}

func (c *children[T]) right() *node[T] {
	return c.get(rightDir)
}

// CompareFunc determines how to order a type 'T'.  It should implement a strict
// ordering, and when
//       'a' < 'b' -> return -1
//       'a' == 'b' -> return 0
//       'a' > 'b' -> return 1
type CompareFunc[T any] func(a, b T) int

type RBTree[T any] struct {
	root *node[T]

	compare CompareFunc[T]
}

func New[T any](compare CompareFunc[T]) *RBTree[T] {
	return &RBTree[T]{compare: compare}
}

func (t *RBTree[T]) Insert(item T) bool {
	pa := make([]*node[T], maxHeight)  // Nodes on stack.
	da := make([]direction, maxHeight) // Directions moved from stack nodes.
	var k int                          // Stack height

	var p *node[T] // Traverses tree looking for insertion point

	pa[0] = t.root
	da[0] = leftDir
	k = 1
	for p = t.root; p != nil; p = p.get(da[k-1]) {
		cmp := t.compare(item, p.data)
		if cmp == 0 {
			return false
		}
		dir := leftDir
		if cmp > 0 {
			dir = rightDir
		}
		pa[k] = p
		da[k] = dir
		k++
	}

	// Newly inserted node
	n := &node[T]{
		color:    red,
		data:     item,
		children: newChildren[T](),
	}

	if t.root == nil {
		t.root = n
		t.root.color = black
		return true
	}

	pa[k-1].set(da[k-1], n)

	for k >= 3 && pa[k-1].color == red {
		if da[k-2] == leftDir {
			y := pa[k-2].get(rightDir)
			if y != nil && y.color == red {
				// case I2: P, U is red, G is black(y === U, pa[k-1] === P)
				y.color = black
				pa[k-1].color = black
				pa[k-2].color = red
				k -= 2
			} else {
				var x *node[T]
				if da[k-1] == leftDir {
					// G-P-N forms an outer line
					y = pa[k-1]
				} else {
					x = pa[k-1]
					y = x.get(rightDir)
					// case I5: P is red, U, G is black and G-P-N forms a triangle
					// x === P, y === N
					// left rotation on x
					x.set(rightDir, y.get(leftDir))
					y.set(leftDir, x)
					pa[k-2].set(leftDir, y)
				}

				x = pa[k-2]
				// case I6, P is red, U, G is black and G-P-N forms a outer line
				// x === G, y === P
				// right rotation on x
				x.set(leftDir, y.get(rightDir))
				y.set(rightDir, x)
				if k-3 == 0 {
					t.root = y
				} else {
					pa[k-3].set(da[k-3], y)
				}

				x.color = red
				y.color = black
				break
			}
		} else {
			y := pa[k-2].get(leftDir)
			if y != nil && y.color == red {
				// case I2: P, U is red, G is black(y === U, pa[k-1] === P)
				y.color = black
				pa[k-1].color = black
				pa[k-2].color = red
				k -= 2
			} else {
				var x *node[T]
				if da[k-1] == rightDir {
					// G-P-N forms an outer line
					y = pa[k-1]
				} else {
					x = pa[k-1]
					y = x.get(leftDir)
					// case I5: P is red, U, G is black and G-P-N forms a triangle
					// x === P, y === N
					// left rotation on x
					x.set(leftDir, y.get(rightDir))
					y.set(rightDir, x)
					pa[k-2].set(rightDir, y)
				}
				x = pa[k-2]
				// case I6, P is red, U, G is black and G-P-N forms a outer line
				// x === G, y === P
				// right rotation on x
				x.set(rightDir, y.get(leftDir))
				y.set(leftDir, x)
				if k-3 == 0 {
					t.root = y
				} else {
					pa[k-3].set(da[k-3], y)
				}

				x.color = red
				y.color = black
				break
			}
		}
	}

	t.root.color = black

	return true
}

func (t *RBTree[T]) Remove(item T) (_ T, _ bool) {
	if t.root == nil {
		return
	}

	pa := make([]*node[T], maxHeight)  // Nodes on stack.
	da := make([]direction, maxHeight) // Directions moved from stack nodes.

	k := 0      // Stack height
	p := t.root // The node to delete, or a node part way to it.
	cmp := t.compare(item, p.data)
	for ; cmp != 0; cmp = t.compare(item, p.data) {
		dir := leftDir
		if cmp > 0 {
			dir = rightDir
		}

		pa[k] = p
		da[k] = dir
		k++

		p = p.get(dir)
		if p == nil {
			return
		}
	}
	item = p.data

	if p.get(rightDir) == nil { // p has no right child
		t.setLinkForPred(pa, da, k-1, p.get(leftDir))
	} else {
		r := p.get(rightDir)
		if r.get(leftDir) == nil { // p has right child `r` and `r` has no left child.
			r.set(leftDir, p.get(leftDir))
			t.setLinkForPred(pa, da, k-1, r)

			t := r.color
			r.color = p.color
			p.color = t

			pa[k] = r
			da[k] = rightDir
			k++
		} else { // p has two non-nil child
			var s *node[T] // the first in-order successor

			j := k
			k++
			for {
				da[k] = leftDir
				pa[k] = r
				k++
				s = r.get(leftDir)
				if s.get(leftDir) == nil {
					break
				}
				r = s
			}

			// swap p and s
			da[j] = rightDir
			pa[j] = s

			s.set(leftDir, p.get(leftDir))
			r.set(leftDir, s.get(rightDir))
			s.set(rightDir, p.get(rightDir))
			t.setLinkForPred(pa, da, j-1, s)

			t := s.color
			s.color = p.color
			p.color = t
		}
	}

	if p.color == black {
		for {
			if k == 0 {
				if t.root != nil {
					t.root.color = black
				}
				break
			}
			x := pa[k-1].get(da[k-1])
			if x != nil && x.color == red {
				x.color = black
				break
			}
			if k < 2 {
				break
			}
			if da[k-1] == leftDir {
				w := pa[k-1].get(rightDir)
				if w.color == red {
					// case D3: w === S, pa[k-1] === P
					// left rotation at P
					pa[k-1].set(rightDir, w.get(leftDir))
					w.set(leftDir, pa[k-1])
					pa[k-2].set(da[k-2], w)

					// recolor
					w.color = black
					pa[k-1].color = red

					pa[k] = pa[k-1]
					da[k] = leftDir
					pa[k-1] = w
					k++

					w = pa[k-1].get(rightDir)
				}
				if (w.left() == nil || w.left().color == black) &&
					(w.right() == nil || w.right().color == black) {
					// case D1 or D4: w === S, pa[k-1] === P
					// recolor S to red
					w.color = red
				} else {
					if w.right() == nil || w.right().color == black {
						y := w.get(leftDir)

						// case D5: w === S, y ==== C
						// right rotation at S
						w.set(leftDir, y.right())
						y.set(rightDir, w)
						pa[k-1].set(rightDir, y)

						// recolor
						y.color = black
						w.color = red

						w = pa[k-1].right()
					}
					// case D6: w === S, pa[k-1] === P
					// left rotation at P
					pa[k-1].set(rightDir, w.left())
					w.set(leftDir, pa[k-1])
					pa[k-2].set(da[k-2], w)

					// recolor
					w.color = pa[k-1].color
					pa[k-1].color = black
					w.right().color = black

					break
				}
			} else {
				w := pa[k-1].get(leftDir)
				if w.color == red {
					// case D3: w === S, pa[k-1] === P
					// right rotation at P
					pa[k-1].set(leftDir, w.get(rightDir))
					w.set(rightDir, pa[k-1])
					pa[k-2].set(da[k-2], w)

					// recolor
					w.color = black
					pa[k-1].color = red

					pa[k] = pa[k-1]
					da[k] = rightDir
					pa[k-1] = w
					k++

					w = pa[k-1].get(leftDir)
				}
				if (w.left() == nil || w.left().color == black) &&
					(w.right() == nil || w.right().color == black) {
					// case D1 or D4: w === S, pa[k-1] === P
					// recolor S to red
					w.color = red
				} else {
					if w.right() == nil || w.right().color == black {
						y := w.get(rightDir)

						// case D5: w === S, y ==== C
						// left rotation at S
						w.set(rightDir, y.left())
						y.set(leftDir, w)
						pa[k-1].set(leftDir, y)

						// recolor
						y.color = black
						w.color = red

						w = pa[k-1].left()
					}
					// case D6: w === S, pa[k-1] === P
					// left rotation at P
					pa[k-1].set(leftDir, w.right())
					w.set(rightDir, pa[k-1])
					pa[k-2].set(da[k-2], w)

					// recolor
					w.color = pa[k-1].color
					pa[k-1].color = black
					w.left().color = black

					break
				}
			}
			k--
		}
	}
	p = nil
	return item, true
}

func (t *RBTree[T]) setLinkForPred(pa []*node[T], da []direction, i int, n *node[T]) {
	if i < 0 {
		t.root = n
		return
	}
	pa[i].set(da[i], n)
}

func (t *RBTree[T]) Print(w io.Writer) error {
	if t.root == nil {
		return nil
	}
	return t.root.print(w)
}
