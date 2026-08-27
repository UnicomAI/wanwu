package async_task

import "context"

// 任务类型常量。app-service 仅对话日志导出一种异步任务，从 1 开始。
const (
	ConversationLogExportTaskType = 1 // 对话日志导出
)

// ConversationLogExportTaskParams 对话日志导出任务参数。
type ConversationLogExportTaskParams struct {
	TaskId string `json:"taskId"` // 导出任务 exportId
}

// BusinessTaskService 业务任务服务接口，每种异步任务实现该接口并在 init 中注册。
type BusinessTaskService interface {
	BuildServiceType() uint32
	// InitTask 初始化任务（向 go-async 注册任务工厂）
	InitTask() error
	// SubmitTask 提交任务（创建 go-async 任务记录）
	SubmitTask(ctx context.Context, params interface{}) error
}
