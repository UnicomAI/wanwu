package queue_util

// BoundedQueue basic bounded queue structure
type BoundedQueue struct {
	items []string
	head  int
	tail  int
	size  int
	cap   int
}

// NewBoundedQueue creates a bounded queue, the capacity must be a positive integer
func NewBoundedQueue(cap int) *BoundedQueue {
	if cap <= 0 {
		panic("Queue capacity must be a positive integer")
	}
	return &BoundedQueue{
		items: make([]string, cap),
		cap:   cap,
	}
}

// Enqueue non-blocking enqueue, returns whether it is successful
func (q *BoundedQueue) Enqueue(item string) bool {
	if q.IsFull() {
		return false
	}

	q.items[q.tail] = item
	q.tail = (q.tail + 1) % q.cap
	q.size++
	return true
}

// Dequeue non-blocking dequeue, returns the element and whether it is successful
func (q *BoundedQueue) Dequeue() (string, bool) {
	if q.IsEmpty() {
		return "", false
	}

	item := q.items[q.head]
	q.items[q.head] = "" // Prevent memory leaks
	q.head = (q.head + 1) % q.cap
	q.size--
	return item, true
}

// Size current number of elements
func (q *BoundedQueue) Size() int {
	return q.size
}

// IsEmpty Is the queue empty?
func (q *BoundedQueue) IsEmpty() bool {
	return q.size == 0
}

// IsFull Is it full?
func (q *BoundedQueue) IsFull() bool {
	return q.size == q.cap
}

// Cap gets the queue capacity
func (q *BoundedQueue) Cap() int {
	return q.cap
}

// AllValue gets all elements
func (q *BoundedQueue) AllValue() []string {
	var retList []string
	for i := 0; i < q.size; i++ {
		pos := (q.head + i) % q.cap
		retList = append(retList, q.items[pos])
	}
	return retList
}
