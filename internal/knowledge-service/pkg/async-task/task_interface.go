package async_task

import (
	"context"
)

const (
	KnowledgeDeleteTaskType       = 1  // 知识库删除
	DocDeleteTaskType             = 2  // 文档列表删除
	DocImportTaskType             = 3  // 文档导入
	DocSegmentImportTaskType      = 4  // 文档分片导入
	KnowledgeReportTaskType       = 5  // 知识库社区报告批量导入
	QADeleteTaskType              = 6  // 问答库删除
	KnowledgeQAPairImportTaskType = 7  // 问答库导入
	QAPairDeleteTaskType          = 8  // 问答对列表删除
	KnowledgeQAPairExportTaskType = 9  // 问答库导出
	DocReImportTaskType           = 10 // 文档重新导入
	KnowledgeDocExportTaskType    = 11 // 知识库导出
)

type TaskTraceParams struct {
	TraceID string `json:"traceID"`
	SpanID  string `json:"spanID"`
}

type KnowledgeDeleteParams struct {
	KnowledgeId string `json:"knowledgeId"`
}

type KnowledgeReportImportTaskParams struct {
	TaskId string `json:"taskId"`
}

type KnowledgeQAPairImportTaskParams struct {
	TaskId string `json:"taskId"`
}

type KnowledgeQAPairExportTaskParams struct {
	TaskId string `json:"taskId"`
}

type KnowledgeDocExportTaskParams struct {
	TaskId string `json:"taskId"`
}

type DocDeleteParams struct {
	DocIdList []uint32 `json:"docIdList"`
}

type QAPairDeleteParams struct {
	QaPairId string `json:"qaPairId"`
}

type DocImportTaskParams struct {
	TaskId string `json:"taskId"`
	TaskTraceParams
}

type DocReImportTaskParams struct {
	TaskId string `json:"taskId"`
	DocId  string `json:"docId"`
	TaskTraceParams
}

type DocSegmentImportTaskParams struct {
	TaskId string `json:"taskId"`
}

type BusinessTaskService interface {
	BuildServiceType() uint32
	//InitTask 初始化任务
	InitTask() error
	//SubmitTask 提交任务
	SubmitTask(ctx context.Context, params interface{}) error
}
