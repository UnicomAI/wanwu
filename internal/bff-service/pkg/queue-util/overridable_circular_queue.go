package queue_util

import (
	"fmt"
	"strings"
)

// OverridableCircularQueue 循环可覆盖队列，非线程安全 [EN] OverridableCircularQueue loop can overwrite the queue and is not thread safe
type OverridableCircularQueue struct {
	data     []string // 底层存储 [EN] underlying storage
	capacity int      // 队列总容量（实际可存储元素数量） [EN] Total queue capacity (actual number of elements that can be stored)
	head     int      // 头指针 [EN] head pointer
	tail     int      // 尾指针 [EN] tail pointer
	size     int      // 当前元素数量 [EN] Current number of elements
}

// NewOverridableQueue 初始化队列，设置实际存储容量为k [EN] NewOverridableQueue initializes the queue and sets the actual storage capacity to k
func NewOverridableQueue(k int) *OverridableCircularQueue {
	return &OverridableCircularQueue{
		data:     make([]string, k),
		capacity: k,
		head:     0,
		tail:     -1, // 初始化为-1便于处理第一个元素 [EN] Initialized to -1 to facilitate processing of the first element
		size:     0,
	}
}

// EnQueue 入队操作（自动覆盖最旧元素） [EN] EnQueue enqueue operation (automatically overwrites the oldest element)
func (q *OverridableCircularQueue) EnQueue(value string) {
	// 计算下一个尾指针位置 [EN] Calculate the next tail pointer position
	q.tail = (q.tail + 1) % q.capacity

	if q.size < q.capacity {
		q.size++
	} else { // 队列已满时移动头指针 [EN] Move head pointer when queue is full
		q.head = (q.head + 1) % q.capacity
	}

	q.data[q.tail] = value
}

// DeQueue 出队操作 [EN] DeQueue dequeue operation
func (q *OverridableCircularQueue) DeQueue() bool {
	if q.IsEmpty() {
		return false
	}
	q.head = (q.head + 1) % q.capacity
	q.size--
	return true
}

// Front 获取队首元素 [EN] Front gets the first element of the team
func (q *OverridableCircularQueue) Front() string {
	if q.IsEmpty() {
		return ""
	}
	return q.data[q.head]
}

// Rear 获取队尾元素 [EN] Rear gets the rear element of the queue
func (q *OverridableCircularQueue) Rear() string {
	if q.IsEmpty() {
		return ""
	}
	return q.data[q.tail]
}

// IsEmpty 检查队列是否为空 [EN] IsEmpty checks if the queue is empty
func (q *OverridableCircularQueue) IsEmpty() bool {
	return q.size == 0
}

// IsFull 检查队列是否已满 [EN] IsFull checks if the queue is full
func (q *OverridableCircularQueue) IsFull() bool {
	return q.size == q.capacity
}

// Size 获取当前元素数量 [EN] Size gets the current number of elements
func (q *OverridableCircularQueue) Size() int {
	return q.size
}

// AllValue 获取	全部元素 [EN] AllValue gets all elements
func (q *OverridableCircularQueue) AllValue() string {
	var builder strings.Builder
	for i := 0; i < q.size; i++ {
		pos := (q.head + i) % q.capacity
		builder.WriteString(q.data[pos])
	}
	return builder.String()
}

// Print 打印队列状态（调试用） [EN] Print print queue status (for debugging)
func (q *OverridableCircularQueue) Print() {
	fmt.Print("Queue [")
	for i := 0; i < q.size; i++ {
		pos := (q.head + i) % q.capacity
		fmt.Printf("%s ", q.data[pos])
	}
	fmt.Println("]")
}
