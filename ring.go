package main

// Ring is a manager for circular, fixed-size buffers.
// Keeps a pointer to the head and tail positions in the buffer
// - Head is the index where next item should be stored
// - Tail is the index of the oldest index stored
type Ring struct {
	head int  // offset of the next item to push to the buffer
	tail int  // offset of the oldest item in the buffer
	full bool // true if buffer is full
	Size int  // total size of the buffer
}

// Iterator on a Ring
type Iterator struct {
	size      int
	countdown int
	At        int
}

// HeadNext returns the current Head pointer, and moves it forward.
func (r *Ring) HeadNext() int {
	prevHead := r.head
	r.head++
	if r.head >= r.Size {
		r.head = 0
	}
	// If it was already full, drag tail along with head
	if r.full {
		r.tail = r.head
	} else if r.head == r.tail {
		// Otherwise, check if we wrapped around
		r.full = true
	}
	return prevHead
}

// TailNext returns the current tail pointer, and moves it forward.
func (r *Ring) TailNext() int {
	if r.head == r.tail && !r.full {
		return -1
	}
	prevTail := r.tail
	r.tail++
	if r.tail >= r.Size {
		r.tail = 0
	}
	r.full = false
	return prevTail
}

// Full returns true if the queue is full
func (r *Ring) Full() bool {
	return r.full
}

// Reset the ring
func (r *Ring) Reset() {
	r.head, r.tail = 0, 0
	r.full = false
}

// Each returns an iterator to traverse the ring in insertion order,
// from oldest to newest.
func (r *Ring) Each() Iterator {
	if r.full {
		return Iterator{
			size:      r.Size,
			countdown: r.Size,
			At:        r.tail - 1, // Iterator.Next will increment this
		}
	}
	// If head > tail, the buffer does not wrap
	countdown := r.head - r.tail
	if countdown < 0 {
		// If tail > head, the buffer has wrapped around
		countdown = r.Size - countdown
	}
	return Iterator{
		size:      r.Size,
		countdown: countdown,
		At:        r.tail - 1, // Iterator.Next will increment this
	}
}

// Next returns true if there are more items to iterate
func (i *Iterator) Next() bool {
	if i.countdown <= 0 {
		return false
	}
	i.countdown--
	i.At++
	if i.At >= i.size {
		i.At = 0
	}
	return true
}

type intRing struct {
	Items []int
	Ring
}

func makeIntRing(size int) intRing {
	return intRing{
		Items: make([]int, size),
		Ring:  Ring{Size: size},
	}
}

// Push the item at Head, and moves Head forward.
// returns the value evicted.
func (r *intRing) Push(val int) int {
	// Must check Full() before pushing. If it is full *after* pushing,
	// but not before, we have not evicted anything
	full := r.Full()
	evicted, head := 0, r.HeadNext()
	if full {
		evicted = r.Items[head]
	}
	r.Items[head] = val
	return evicted
}

// Pop the item at Tail, and move Tail forward.
// returns the former Tail, and the value evicted.
func (r *intRing) Pop() int {
	evicted, tail := 0, r.TailNext()
	if tail >= 0 {
		evicted = r.Items[tail]
	}
	return evicted
}
