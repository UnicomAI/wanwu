package queue_util

import (
	"fmt"
	"strings"
)

// OverridableCircularQueue loop can overwrite the queue and is not thread safe
type OverridableCircularQueue struct {
	data     []string // underlying storage
	capacity int      // Total queue capacity (actual number of elements that can be stored)
	head     int      // head pointer
	tail     int      // tail pointer
	size     int      // Current number of elements
}

// NewOverridableQueue initializes the queue and sets the actual storage capacity to k
func NewOverridableQueue(k int) *OverridableCircularQueue {
	return &OverridableCircularQueue{
		data:     make([]string, k),
		capacity: k,
		head:     0,
		tail:     -1, // Initialized to -1 to facilitate processing of the first element
		size:     0,
	}
}

// EnQueue enqueue operation (automatically overwrites the oldest element)
func (q *OverridableCircularQueue) EnQueue(value string) {
	// Calculate the next tail pointer position
	q.tail = (q.tail + 1) % q.capacity

	if q.size < q.capacity {
		q.size++
	} else { // Move head pointer when queue is full
		q.head = (q.head + 1) % q.capacity
	}

	q.data[q.tail] = value
}

// DeQueue dequeue operation
func (q *OverridableCircularQueue) DeQueue() bool {
	if q.IsEmpty() {
		return false
	}
	q.head = (q.head + 1) % q.capacity
	q.size--
	return true
}

// Front gets the first element of the team
func (q *OverridableCircularQueue) Front() string {
	if q.IsEmpty() {
		return ""
	}
	return q.data[q.head]
}

// Rear gets the rear element of the queue
func (q *OverridableCircularQueue) Rear() string {
	if q.IsEmpty() {
		return ""
	}
	return q.data[q.tail]
}

// IsEmpty checks if the queue is empty
func (q *OverridableCircularQueue) IsEmpty() bool {
	return q.size == 0
}

// IsFull checks if the queue is full
func (q *OverridableCircularQueue) IsFull() bool {
	return q.size == q.capacity
}

// Size gets the current number of elements
func (q *OverridableCircularQueue) Size() int {
	return q.size
}

// AllValue gets all elements
func (q *OverridableCircularQueue) AllValue() string {
	var builder strings.Builder
	for i := 0; i < q.size; i++ {
		pos := (q.head + i) % q.capacity
		builder.WriteString(q.data[pos])
	}
	return builder.String()
}

// Print print queue status (for debugging)
func (q *OverridableCircularQueue) Print() {
	fmt.Print("Queue [")
	for i := 0; i < q.size; i++ {
		pos := (q.head + i) % q.capacity
		fmt.Printf("%s ", q.data[pos])
	}
	fmt.Println("]")
}
