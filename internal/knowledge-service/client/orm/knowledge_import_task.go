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
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"gorm.io/gorm"
)

// SelectKnowledgeRunningImportTask 查询导入信息
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

// SelectKnowledgeLatestImportTask 查询最近导入任务
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

// SelectKnowledgeImportTaskById 根据id查询导入信息
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

// SelectKnowledgeImportTaskByIdList 根据id 列表查询导入信息
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

// CreateKnowledgeImportTask 导入任务
func CreateKnowledgeImportTask(ctx context.Context, importTask *model.KnowledgeImportTask) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.创建知识库导入任务
		err := createKnowledgeImportTask(tx, importTask)
		if err != nil {
			return err
		}
		//2.通知rag更新知识库
		return async_task.SubmitTask(ctx, async_task.DocImportTaskType, &async_task.DocImportTaskParams{
			TaskId: importTask.ImportId,
			TaskTraceParams: async_task.TaskTraceParams{
				TraceID: trace_util.GetTraceID(ctx),
				SpanID:  trace_util.GetSpanID(ctx),
			},
		})
	})
}

// CreateKnowledgeReImportTask 导入任务
func CreateKnowledgeReImportTask(ctx context.Context, importTask *model.KnowledgeImportTask, knowledgeDoc *model.KnowledgeDoc) error {
	return db.GetHandle(ctx).Transaction(func(tx *gorm.DB) error {
		//1.创建知识库导入任务
		err := createKnowledgeImportTask(tx, importTask)
		if err != nil {
			return err
		}
		//2.更新文档导入任务id
		err = UpdateDocInfo(tx, knowledgeDoc.DocId, -1, "", importTask.ImportId)
		if err != nil {
			return err
		}
		//3.通知rag更新知识库
		return async_task.SubmitTask(ctx, async_task.DocReImportTaskType, &async_task.DocReImportTaskParams{
			TaskId: importTask.ImportId,
			DocId:  knowledgeDoc.DocId,
			TaskTraceParams: async_task.TaskTraceParams{
				TraceID: trace_util.GetTraceID(ctx),
				SpanID:  trace_util.GetSpanID(ctx),
			},
		})
	})
}

// UpdateKnowledgeImportTaskStatus 更新导入任务状态
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

// DeleteImportTaskByKnowledgeId 根据知识库id 删除导入任务
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
