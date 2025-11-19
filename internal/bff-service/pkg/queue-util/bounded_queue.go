package queue_util

// BoundedQueue 基础有界队列结构 [EN] BoundedQueue basic bounded queue structure
type BoundedQueue struct {
	items []string
	head  int
	tail  int
	size  int
	cap   int
}

// NewBoundedQueue 创建有界队列，容量必须为正整数 [EN] NewBoundedQueue creates a bounded queue, the capacity must be a positive integer
func NewBoundedQueue(cap int) *BoundedQueue {
	if cap <= 0 {
		panic("队列容量必须为正整数")
	}
	return &BoundedQueue{
		items: make([]string, cap),
		cap:   cap,
	}
}

// Enqueue 非阻塞入队，返回是否成功 [EN] Enqueue non-blocking enqueue, returns whether it is successful
func (q *BoundedQueue) Enqueue(item string) bool {
	if q.IsFull() {
		return false
	}

	q.items[q.tail] = item
	q.tail = (q.tail + 1) % q.cap
	q.size++
	return true
}

// Dequeue 非阻塞出队，返回元素和是否成功 [EN] Dequeue non-blocking dequeue, returns the element and whether it is successful
func (q *BoundedQueue) Dequeue() (string, bool) {
	if q.IsEmpty() {
		return "", false
	}

	item := q.items[q.head]
	q.items[q.head] = "" // 防止内存泄漏 [EN] Prevent memory leaks
	q.head = (q.head + 1) % q.cap
	q.size--
	return item, true
}

// Size 当前元素数量 [EN] Size current number of elements
func (q *BoundedQueue) Size() int {
	return q.size
}

// IsEmpty 是否为空队列 [EN] IsEmpty Is the queue empty?
func (q *BoundedQueue) IsEmpty() bool {
	return q.size == 0
}

// IsFull 是否已满 [EN] IsFull Is it full?
func (q *BoundedQueue) IsFull() bool {
	return q.size == q.cap
}

// Cap 获取队列容量 [EN] Cap gets the queue capacity
func (q *BoundedQueue) Cap() int {
	return q.cap
}

// AllValue 获取	全部元素 [EN] AllValue gets all elements
func (q *BoundedQueue) AllValue() []string {
	var retList []string
	for i := 0; i < q.size; i++ {
		pos := (q.head + i) % q.cap
		retList = append(retList, q.items[pos])
	}
	return retList
}
