package orm

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/model"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/client/orm/sqlopt"
	async_task "github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/db"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
)

// SelectKnowledgeRunningImportTask Query import information
func SelectKnowledgeRunningImportTask(ctx context.Context, knowledgeId string) error {
	var count int64
	err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId), sqlopt.WithStatusList([]int{model.KnowledgeImportAnalyze})).
		Apply(db.GetHandle(ctx), &model.KnowledgeImportTask{}).
		Count(&count).Error
	if err != nil {
		log.Errorf("SelectKnowledgeRunningImportTask knowledgeId %s err: %v", knowledgeId, err)
		return util.ErrCode(errs.Code_KnowledgeBaseDeleteFailed)
	}
	if count > 0 {
		return util.ErrCode(errs.Code_KnowledgeBaseDeleteDuringUpload)
	}
	return nil
}

// SelectKnowledgeLatestImportTask queries the latest import tasks
func SelectKnowledgeLatestImportTask(ctx context.Context, knowledgeId string) ([]*model.KnowledgeImportTask, error) {
	var importTaskList []*model.KnowledgeImportTask
	err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeImportTask{}).
		Order("create_at desc").
		Limit(1).
		Find(&importTaskList).Error
	if err != nil {
		log.Errorf("SelectKnowledgeLatestImportTask knowledgeId %s err: %v", knowledgeId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseDeleteFailed)
	}
	return importTaskList, nil
}

// SelectKnowledgeImportTaskById Query import information based on id
func SelectKnowledgeImportTaskById(ctx context.Context, importId string) (*model.KnowledgeImportTask, error) {
	var importTask model.KnowledgeImportTask
	err := sqlopt.SQLOptions(sqlopt.WithImportID(importId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeImportTask{}).
		First(&importTask).Error
	if err != nil {
		log.Errorf("SelectKnowledgeRunningImportTask importId %s err: %v", importId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseDeleteFailed)
	}
	return &importTask, nil
}

// SelectKnowledgeImportTaskByIdList queries import information based on id list
func SelectKnowledgeImportTaskByIdList(ctx context.Context, importId []string) ([]*model.KnowledgeImportTask, error) {
	var importTask []*model.KnowledgeImportTask
	err := sqlopt.SQLOptions(sqlopt.WithImportIDs(importId)).
		Apply(db.GetHandle(ctx), &model.KnowledgeImportTask{}).
		Find(&importTask).Error
	if err != nil {
		log.Errorf("SelectKnowledgeImportTaskByIdList importId %s err: %v", importId, err)
		return nil, util.ErrCode(errs.Code_KnowledgeBaseDeleteFailed)
	}
	return importTask, nil
}

// CreateKnowledgeImportTask import task
func CreateKnowledgeImportTask(ctx context.Context, importTask *model.KnowledgeImportTask) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1. Create a knowledge base import task
		err := createKnowledgeImportTask(tx, importTask)
		if err != nil {
			return err
		}
		//2. Notify rag to update the knowledge base
		return async_task.SubmitTask(ctx, async_task.DocImportTaskType, &async_task.DocImportTaskParams{
			TaskId: importTask.ImportId,
		})
	})
}

// UpdateKnowledgeImportTaskStatus updates the import task status
func UpdateKnowledgeImportTaskStatus(ctx context.Context, tx *gorm.DB, id uint32, status int, errMsg string) error {
	if tx == nil {
		tx = db.GetHandle(ctx)
	}
	return tx.Model(&model.KnowledgeImportTask{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":    status,
			"error_msg": errMsg,
		}).Error
}

// DeleteImportTaskByKnowledgeId Delete the import task based on the knowledge base id
func DeleteImportTaskByKnowledgeId(tx *gorm.DB, knowledgeId string) error {
	var count int64
	err := sqlopt.SQLOptions(sqlopt.WithKnowledgeID(knowledgeId)).
		Apply(tx, &model.KnowledgeImportTask{}).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return tx.Unscoped().Model(&model.KnowledgeImportTask{}).Where("knowledge_id = ?", knowledgeId).Delete(&model.KnowledgeImportTask{}).Error
	}
	return nil
}

func createKnowledgeImportTask(tx *gorm.DB, importTask *model.KnowledgeImportTask) error {
	return tx.Model(&model.KnowledgeImportTask{}).Create(importTask).Error
}
