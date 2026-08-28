package async_task

import (
	"context"
	"errors"
)

var taskServiceMap = make(map[uint32]BusinessTaskService)

// AddContainer 注册业务任务服务（在 worker 的 init 中调用）。
func AddContainer(service BusinessTaskService) {
	taskServiceMap[service.BuildServiceType()] = service
}

// InitAllService 初始化所有已注册业务任务（向 go-async 注册任务工厂）。
func InitAllService() error {
	for _, service := range taskServiceMap {
		if err := service.InitTask(); err != nil {
			return err
		}
	}
	return nil
}

// SubmitTask 按任务类型提交任务。
func SubmitTask(ctx context.Context, taskType uint32, params interface{}) error {
	service := taskServiceMap[taskType]
	if service == nil {
		return errors.New("未找到对应任务类型")
	}
	return service.SubmitTask(ctx, params)
}
