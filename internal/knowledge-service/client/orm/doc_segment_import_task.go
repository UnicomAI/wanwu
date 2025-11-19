package orm

import (
	"context"

	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	async_task "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

// SelectDocSegmentImportTaskById 根据id查询导入信息 [EN] SelectDocSegmentImportTaskById Query import information based on id
func SelectDocSegmentImportTaskById(ctx context.Context, importId string) (*model.DocSegmentImportTask, error) {
	var importTask model.DocSegmentImportTask
	err := sqlopt.SQLOptions(sqlopt.WithImportID(importId)).
		Apply(db.GetHandle(ctx), &model.DocSegmentImportTask{}).
		First(&importTask).Error
	if err != nil {
		log.Errorf("SelectDocSegmentImportTaskById importId %s err: %v", importId, err)
		//todo 错误码 [EN] todo error code
		return nil, err
	}
	return &importTask, nil
}

// SelectSegmentLatestImportTaskByDocID 查询文档最新的导入信息 [EN] SelectSegmentLatestImportTaskByDocID Query the latest import information of the document
func SelectSegmentLatestImportTaskByDocID(ctx context.Context, docId string) (*model.DocSegmentImportTask, error) {
	var importTask model.DocSegmentImportTask
	err := sqlopt.SQLOptions(sqlopt.WithDocID(docId)).
		Apply(db.GetHandle(ctx), &model.DocSegmentImportTask{}).
		Order("create_at desc").
		First(&importTask).Error
	if err != nil {
		log.Errorf("SelectSegmentLatestImportTaskByDocID docId %s err: %v", docId, err)
		return nil, err
	}
	return &importTask, nil
}

// CreateDocSegmentImportTask 导入任务 [EN] CreateDocSegmentImportTask import task
func CreateDocSegmentImportTask(ctx context.Context, importTask *model.DocSegmentImportTask) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.创建知识库导入任务 [EN] 1. Create a knowledge base import task
		err := createDocSegmentImportTask(tx, importTask)
		if err != nil {
			return err
		}
		//2.通知rag更新知识库 [EN] 2. Notify rag to update the knowledge base
		return async_task.SubmitTask(ctx, async_task.DocSegmentImportTaskType, &async_task.DocSegmentImportTaskParams{
			TaskId: importTask.ImportId,
		})
	})
}

// CreateOneDocSegment 创建一个分片 [EN] CreateOneDocSegment creates a segment
func CreateOneDocSegment(ctx context.Context, importTask *model.DocSegmentImportTask, importParams *service.RagCreateDocSegmentParams) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.创建知识库导入任务 [EN] 1. Create a knowledge base import task
		err := tx.Model(&model.DocSegmentImportTask{}).Where("import_id = ?", importTask.ImportId).
			Update("success_count", gorm.Expr("success_count + ?", 1)).Error
		if err != nil {
			return err
		}
		//2.通知rag创建分段 [EN] 2. Notify rag to create segments
		return service.RagCreateDocSegment(ctx, importParams)
	})
}

// UpdateDocSegmentImportTaskStatus 更新导入任务状态 [EN] UpdateDocSegmentImportTaskStatus updates the import task status
func UpdateDocSegmentImportTaskStatus(ctx context.Context, taskId string, status int, errMsg string, totalCount int) error {
	return db.GetHandle(ctx).Model(&model.DocSegmentImportTask{}).
		Where("import_id = ?", taskId).
		Updates(map[string]interface{}{
			"status":      status,
			"error_msg":   errMsg,
			"total_count": totalCount,
		}).Error
}

func createDocSegmentImportTask(tx *gorm.DB, importTask *model.DocSegmentImportTask) error {
	return tx.Model(&model.DocSegmentImportTask{}).Create(importTask).Error
}
