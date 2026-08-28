package queue_util

import (
	"fmt"
	"strings"
)

// OverridableCircularQueue 循环可覆盖队列，非线程安全
type OverridableCircularQueue struct {
	data     []string // 底层存储
	capacity int      // 队列总容量（实际可存储元素数量）
	head     int      // 头指针
	tail     int      // 尾指针
	size     int      // 当前元素数量
}

// NewOverridableQueue 初始化队列，设置实际存储容量为k
func NewOverridableQueue(k int) *OverridableCircularQueue {
	return &OverridableCircularQueue{
		data:     make([]string, k),
		capacity: k,
		head:     0,
		tail:     -1, // 初始化为-1便于处理第一个元素
		size:     0,
	}
}

// EnQueue 入队操作（自动覆盖最旧元素）
func (q *OverridableCircularQueue) EnQueue(value string) {
	// 计算下一个尾指针位置
	q.tail = (q.tail + 1) % q.capacity

	if q.size < q.capacity {
		q.size++
	} else { // 队列已满时移动头指针
		q.head = (q.head + 1) % q.capacity
	}

	q.data[q.tail] = value
}

// DeQueue 出队操作
func (q *OverridableCircularQueue) DeQueue() bool {
	if q.IsEmpty() {
		return false
	}
	q.head = (q.head + 1) % q.capacity
	q.size--
	return true
}

// Front 获取队首元素
func (q *OverridableCircularQueue) Front() string {
	if q.IsEmpty() {
		return ""
	}
	return q.data[q.head]
}

// Rear 获取队尾元素
func (q *OverridableCircularQueue) Rear() string {
	if q.IsEmpty() {
		return ""
	}
	return q.data[q.tail]
}

// IsEmpty 检查队列是否为空
func (q *OverridableCircularQueue) IsEmpty() bool {
	return q.size == 0
}

// IsFull 检查队列是否已满
func (q *OverridableCircularQueue) IsFull() bool {
	return q.size == q.capacity
}

// Size 获取当前元素数量
func (q *OverridableCircularQueue) Size() int {
	return q.size
}

// AllValue 获取	全部元素
func (q *OverridableCircularQueue) AllValue() string {
	var builder strings.Builder
	for i := 0; i < q.size; i++ {
		pos := (q.head + i) % q.capacity
		builder.WriteString(q.data[pos])
	}
	return builder.String()
}

// Clear 清空队列
func (q *OverridableCircularQueue) Clear() {
	q.head = 0
	q.tail = -1
	q.size = 0
}

// Print 打印队列状态（调试用）
func (q *OverridableCircularQueue) Print() {
	fmt.Print("Queue [")
	for i := 0; i < q.size; i++ {
		pos := (q.head + i) % q.capacity
		fmt.Printf("%s ", q.data[pos])
	}
	fmt.Println("]")
}
