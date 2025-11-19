package async_task

import (
	"context"
)

const (
	KnowledgeDeleteTaskType  = 1 //知识库删除 [EN] Knowledge base deletion
	DocDeleteTaskType        = 2 // 文档列表删除 [EN] Document list delete
	DocImportTaskType        = 3 // 文档导入 [EN] Document import
	DocSegmentImportTaskType = 4 // 文档分片导入 [EN] Document segmentation import
)

type KnowledgeDeleteParams struct {
	KnowledgeId string `json:"knowledgeId"`
}

type DocDeleteParams struct {
	DocIdList []uint32 `json:"docIdList"`
}

type DocImportTaskParams struct {
	TaskId string `json:"taskId"`
}

type DocSegmentImportTaskParams struct {
	TaskId string `json:"taskId"`
}

type BusinessTaskService interface {
	BuildServiceType() uint32
	//InitTask 初始化任务 [EN] InitTask initialization task
	InitTask() error
	//SubmitTask 提交任务 [EN] SubmitTask Submit task
	SubmitTask(ctx context.Context, params interface{}) error
}
