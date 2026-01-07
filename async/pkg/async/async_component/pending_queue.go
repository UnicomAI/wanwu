package async_component

import "context"

type IQueue interface {
	// AddTask 向队列中添加一个任务
	AddTask(ctx context.Context, user, group string, taskID, taskType uint32) error
	// DelTask 从队列中删除一个任务（直接删除，只有正确删除指定任务才不返回error）
	DelTask(ctx context.Context, taskID uint32) error
	// PopOne 从队列中获取一个任务（出队列）
	PopOne(ctx context.Context, taskTypes []uint32) (uint32, error)
}
