package queue

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

type Queue[T any] interface {
	PopFront() T
	PushBack(item T)
	Size() int
}

func New[T any]() Queue[T] {
	return &queue[T]{}
}

type queue[T any] struct {
	items items[T]
}

func (q *queue[T]) PopFront() (_ T) {
	if len(q.items) == 0 {
		return
	}
	return q.items.removeAt(0)
}

func (q *queue[T]) PushBack(item T) {
	idx := len(q.items)
	q.items.insertAt(idx, item)
}

func (q *queue[T]) Size() int {
	return len(q.items)
}
