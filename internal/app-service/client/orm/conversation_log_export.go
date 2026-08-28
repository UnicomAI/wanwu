package orm

import (
	"context"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/internal/app-service/config"
	async_task "github.com/UnicomAI/wanwu/internal/app-service/pkg/async-task"
	"github.com/UnicomAI/wanwu/internal/app-service/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
	"time"
)

// CreateConversationLogExportTask 创建对话日志导出任务。
func (c *Client) CreateConversationLogExportTask(ctx context.Context, exportTask *model.ConversationLogExportTask) *errs.Status {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1.创建导出任务
		if err := tx.Model(&model.ConversationLogExportTask{}).Create(exportTask).Error; err != nil {
			return err
		}
		// 2.提交异步导出任务
		return async_task.SubmitTask(ctx, async_task.ConversationLogExportTaskType, &async_task.ConversationLogExportTaskParams{
			TaskId: exportTask.ExportId,
		})
	})
	if err != nil {
		log.Errorf("create conversation log export task err: %v", err)
		return toErrStatus("conversation_log_export_create", exportTask.ExportId, err.Error())
	}
	return nil
}

// GetConversationLogExportTaskList 分页查询导出任务列表（含总数）。
func (c *Client) GetConversationLogExportTaskList(ctx context.Context, appId, appType string, startDate, endDate, searchTitle string, userIds, orgIds []string, pageSize, pageNum int32) ([]*model.ConversationLogExportTask, int64, *errs.Status) {
	limit := pageSize
	offset := pageSize * (pageNum - 1)
	var exportTasks []*model.ConversationLogExportTask
	var total int64
	opts := []sqlopt.SQLOption{
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
		sqlopt.LikeTitle(searchTitle),
		sqlopt.WithUserIDs(userIds),
		sqlopt.WithOrgIDs(orgIds),
	}

	const dayMs = int64(24 * time.Hour / time.Millisecond)
	if startDate != "" {
		if startTs, err := util.Date2Time(startDate); err == nil {
			opts = append(opts, sqlopt.StartUpdatedAt(startTs))
		}
	}
	if endDate != "" {
		if endTs, err := util.Date2Time(endDate); err == nil {
			opts = append(opts, sqlopt.EndUpdatedAt(endTs+dayMs-1))
		}
	}

	if err := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.ConversationLogExportTask{}).Count(&total).Error; err != nil {
		log.Errorf("GetConversationLogExportTaskList count appId %s err: %v", appId, err)
		return nil, 0, toErrStatus("conversation_log_export_count", appId, err.Error())
	}
	if err := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.ConversationLogExportTask{}).
		Order("created_at desc").Limit(int(limit)).Offset(int(offset)).Find(&exportTasks).Error; err != nil {
		log.Errorf("GetConversationLogExportTaskList find appId %s err: %v", appId, err)
		return nil, 0, toErrStatus("conversation_log_export_find", appId, err.Error())
	}
	return exportTasks, total, nil
}

// SelectConversationLogExportTaskById 根据导出id查询导出任务（worker 用）。
func (c *Client) SelectConversationLogExportTaskById(ctx context.Context, exportId string) (*model.ConversationLogExportTask, error) {
	var exportTask model.ConversationLogExportTask
	err := sqlopt.SQLOptions(sqlopt.WithExportID(exportId)).
		Apply(c.db.WithContext(ctx)).First(&exportTask).Error
	if err != nil {
		log.Errorf("SelectConversationLogExportTaskById exportId %s err: %v", exportId, err)
		return nil, err
	}
	return &exportTask, nil
}

// TrySetConvLogExpTask 乐观锁抢占导出任务：
// 仅当任务处于待处理(Init)或 导出中 状态时，置为导出中。
// 返回 (rowsAffected, err)：rowsAffected=0 表示已被抢/已完成/已删除等，调用方应跳过执行。
func (c *Client) TrySetConvLogExpTask(ctx context.Context, exportId string) (int64, error) {
	res := c.db.WithContext(ctx).Model(&model.ConversationLogExportTask{}).
		Where("export_id = ? AND status  IN ?", exportId, []int{model.ConversationLogExportInit, model.ConversationLogExportExporting}).
		Update("status", model.ConversationLogExportExporting)
	if res.Error != nil {
		log.Errorf("TrySetConvLogExpTask exportId %s err: %v", exportId, res.Error)
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// UpdateConversationLogExportTask 更新导出任务状态/文件信息（worker 用）。
func (c *Client) UpdateConversationLogExportTask(ctx context.Context, exportId string, status int, errMsg string, totalCount, successCount int, filePath string, fileSize int64) error {
	return c.db.WithContext(ctx).Model(&model.ConversationLogExportTask{}).
		Where("export_id = ?", exportId).
		Updates(map[string]interface{}{
			"status":           status,
			"error_msg":        errMsg,
			"success_count":    successCount,
			"total_count":      totalCount,
			"export_file_path": filePath,
			"export_file_size": fileSize,
		}).Error
}

// DeleteConversationLogExportTaskByIds 根据导出id删除导出任务。
func (c *Client) DeleteConversationLogExportTaskByIds(ctx context.Context, exportIds []string, userId, orgId string) *errs.Status {
	// 1.查询待删除的导出任务。
	var exportTasks []*model.ConversationLogExportTask
	if err := sqlopt.SQLOptions(
		sqlopt.WithExportIDs(exportIds),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Model(&model.ConversationLogExportTask{}).Find(&exportTasks).Error; err != nil {
		log.Errorf("query conversation log export tasks before delete exportIds %v err: %v", exportIds, err)
	}

	// 2.拦截导出中的任务：导出中(status=1)的任务仍在异步导出，删除会导致状态写回丢失与孤儿文件，故整批拒绝。
	var exportingIds []string
	for _, t := range exportTasks {
		if t != nil && t.Status == model.ConversationLogExportExporting {
			exportingIds = append(exportingIds, t.ExportId)
		}
	}
	if len(exportingIds) > 0 {
		log.Errorf("DeleteConversationLogExportTaskByIds blocked, tasks still exporting exportIds %v", exportingIds)
		return toErrStatus("conversation_log_export_delete_exporting", strings.Join(exportingIds, ","))
	}

	// 3.硬删行
	if err := c.db.WithContext(ctx).Unscoped().Model(&model.ConversationLogExportTask{}).
		Where("export_id IN ?", exportIds).
		Where("user_id = ?", userId).
		Where("org_id = ?", orgId).
		Delete(&model.ConversationLogExportTask{}).Error; err != nil {
		log.Errorf("DeleteConversationLogExportTaskByIds delete rows exportIds %v err: %v", exportIds, err)
		return toErrStatus("conversation_log_export_delete", err.Error())
	}

	// 4.协程异步删 MinIO 文件。
	filePaths := make([]string, 0, len(exportTasks))
	for _, t := range exportTasks {
		if t != nil && t.ExportFilePath != "" {
			filePaths = append(filePaths, t.ExportFilePath)
		}
	}
	if len(filePaths) > 0 {
		safe_go_util.SafeGo(func() {
			for _, p := range filePaths {
				filePath := "http://" + config.Cfg().Minio.Endpoint + "/" + p
				if err := minio.DeleteFile(context.Background(), filePath); err != nil {
					log.Errorf("delete conversation log export minio file %s err: %v", p, err)
				}
			}
		})
	}
	return nil
}
