package async_task

import (
	"context"
)

const (
	KnowledgeDeleteTaskType  = 1 //Knowledge base deletion
	DocDeleteTaskType        = 2 // Document list delete
	DocImportTaskType        = 3 // Document import
	DocSegmentImportTaskType = 4 // Document segmentation import
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
	//InitTask initialization task
	InitTask() error
	//SubmitTask Submit task
	SubmitTask(ctx context.Context, params interface{}) error
}
