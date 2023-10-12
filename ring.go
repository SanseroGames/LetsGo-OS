package main

// I don't know how to make this datastructure generic
// After a lot of experimenting I made more or less generic structure that handles
// the index management. But in's not nice like this...

type GenericRing struct {
	head int
	tail int
	Cap  int // Cannot use function as it requires memory allocation for a method reference
}

func (r *GenericRing) Len() int {
	t := r.tail
	h := r.head
	if t < h {
		t += r.Cap
	}
	return t - h
}

func (r *GenericRing) Push() int {
	// Not thread safe
	l := r.Cap
	if r.tail == r.head-1 || (r.head == 0 && r.tail == l-1) {
		return -1
	}
	index := r.tail
	r.tail = (index + 1) % l
	return index
}

func (r *GenericRing) Pop() int {
	// Not thread safe
	if r.head == r.tail {
		return -1
	}
	index := r.head
	r.head = (index + 1) % r.Cap
	return index
}
