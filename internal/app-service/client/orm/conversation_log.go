package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

// GetConversationLog 根据 appId、appType、conversationId 查询单条会话日志。
func (c *Client) GetConversationLog(ctx context.Context, appId, appType, conversationId string) (*model.ConversationLog, *errs.Status) {
	var log model.ConversationLog
	if err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
		sqlopt.WithConversationID(conversationId),
	).Apply(c.db.WithContext(ctx)).First(&log).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toErrStatus("app_conversation_log_not_found", fmt.Sprintf("appId=%s appType=%s conversationId=%s", appId, appType, conversationId))
		}
		return nil, toErrStatus("app_conversation_log_get", fmt.Errorf("get conversation log err: %v", err).Error())
	}
	return &log, nil
}

// GetConversationLogByLogIds 根据 logId 列表批量查询会话日志。
func (c *Client) GetConversationLogByLogIds(ctx context.Context, logIds []string) ([]*model.ConversationLog, *errs.Status) {
	var logs []*model.ConversationLog
	if err := c.db.WithContext(ctx).Where("log_id IN (?)", logIds).Find(&logs).Error; err != nil {
		return nil, toErrStatus("app_conversation_log_get", fmt.Errorf("get conversation log by logIds err: %v", err).Error())
	}
	return logs, nil
}

// GetConversationLogList 获取会话日志列表（分页），返回 logs 与满足过滤条件的总条数 total
func (c *Client) GetConversationLogList(ctx context.Context, orgIds, userIds []string, appId, appType, name string, sources []string, startDate, endDate string, orderBy, orderType string, offset, limit int32) ([]*model.ConversationLog, int64, *errs.Status) {
	opts := []sqlopt.SQLOption{
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
		sqlopt.LikeTitle(name),
		sqlopt.WithSources(sources),
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithUserIDs(userIds),
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

	var total int64
	if err := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).Model(&model.ConversationLog{}).Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("app_conversation_log_list", err.Error())
	}

	var logs []*model.ConversationLog
	query := sqlopt.SQLOptions(opts...).Apply(c.db).WithContext(ctx).
		Model(&model.ConversationLog{}).
		Order(conversationLogOrder(orderBy, orderType)).Offset(int(offset)).Limit(int(limit))
	if err := query.Find(&logs).Error; err != nil {
		return nil, 0, toErrStatus("app_conversation_log_list", fmt.Errorf("get conversation log list err: %v", err).Error())
	}
	return logs, total, nil
}

// conversationLogOrder 将排序参数转换为安全的 ORDER BY 子句。
func conversationLogOrder(orderBy, orderType string) string {
	// 排序字段白名单：前端字段名 -> 数据库列名
	columnMap := map[string]string{
		"avgCosts":             "costs",
		"avgFirstTokenLatency": "first_token_latency",
		"createAt":             "created_at",
		"updateAt":             "updated_at",
	}
	column, ok := columnMap[orderBy]
	if !ok {
		// 默认按 updateAt 降序
		return "updated_at DESC"
	}
	direction := "DESC"
	if orderType == "asc" {
		direction = "ASC"
	}
	// des 即为降序（DESC）；其他非 asc 值统一按降序处理
	return fmt.Sprintf("%s %s", column, direction)
}

// GetConversationLogUserIds 根据 appId 和 appType 查询去重后的 userId 列表，可按 orgIds、userIds 筛选。
func (c *Client) GetConversationLogUserIds(ctx context.Context, appId, appType string, orgIds, userIds []string) ([]string, *errs.Status) {
	var result []string
	err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithUserIDs(userIds),
	).Apply(c.db).WithContext(ctx).Model(&model.ConversationLog{}).Distinct("user_id").Pluck("user_id", &result).Error
	if err != nil {
		return nil, toErrStatus("app_conversation_log_user_select", fmt.Errorf("get conversation log user ids err: %v", err).Error())
	}
	return result, nil
}

// GetConversationLogListByLogIds 按 logId 列表查询对话日志（导出按勾选日志导出用）。
// appId/appType 用于校验 logId 属于指定应用；orgIds/userIds 用于校验调用方对该 logId 的可见权限（空=不限）。
func (c *Client) GetConversationLogListByLogIds(ctx context.Context, logIds []string, appId, appType string, orgIds, userIds []string) ([]*model.ConversationLog, *errs.Status) {
	var logs []*model.ConversationLog
	if err := sqlopt.SQLOptions(
		sqlopt.WithLogIDs(logIds),
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithUserIDs(userIds)).
		Apply(c.db).WithContext(ctx).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, toErrStatus("app_conversation_log_list_by_log_ids", fmt.Errorf("get conversation log list by log ids err: %v", err).Error())
	}
	return logs, nil
}

// DeleteConversationLogByAppId 删除指定应用下的全部会话日志，用于删除应用时清场。
func (c *Client) DeleteConversationLogByAppId(ctx context.Context, appId, appType string) *errs.Status {
	if appId == "" || appType == "" {
		return toErrStatus("app_conversation_log_delete", fmt.Sprintf("invalid parameters: appId=%s appType=%s", appId, appType))
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
	).Apply(c.db.WithContext(ctx)).Delete(&model.ConversationLog{}).Error; err != nil {
		return toErrStatus("app_conversation_log_delete", fmt.Errorf("delete conversation log by app err: %v", err).Error())
	}
	return nil
}

// RecordConversationLog 记录会话日志（upsert）。
func (c *Client) RecordConversationLog(ctx context.Context, log *model.ConversationLog) *errs.Status {
	if log.AppId == "" || log.AppType == "" || log.ConversationId == "" {
		return toErrStatus("app_conversation_log_record", fmt.Sprintf("invalid parameters: appId=%s appType=%s conversationId=%s", log.AppId, log.AppType, log.ConversationId))
	}

	// 新建时生成 logId（logId 唯一索引，且调用方可能未填充）
	if log.LogId == "" {
		log.LogId = util.NewID()
	}
	// 标题超长截断（数据库列 varchar(128)）
	if len(log.Title) > 128 {
		log.Title = log.Title[:128]
	}

	var existing model.ConversationLog
	err := sqlopt.SQLOptions(
		sqlopt.WithAppID(log.AppId),
		sqlopt.WithAppType(log.AppType),
		sqlopt.WithConversationID(log.ConversationId),
	).Apply(c.db.WithContext(ctx)).First(&existing).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("app_conversation_log_record", fmt.Errorf("query conversation log err: %v", err).Error())
		}
		// 不存在：新建
		if err := c.db.WithContext(ctx).Create(log).Error; err != nil {
			return toErrStatus("app_conversation_log_record", fmt.Errorf("create conversation log err: %v", err).Error())
		}
		return nil
	}
	if log.Ext == "" {
		log.Ext = existing.Ext
	}

	// 存在：更新（保留主键与 logId、创建时间，更新业务字段）
	updates := map[string]interface{}{
		"title":               log.Title,
		"source":              log.Source,
		"version":             log.Version,
		"user_id":             log.UserId,
		"org_id":              log.OrgId,
		"message_count":       log.MessageCount,
		"costs":               log.Costs,
		"first_token_latency": log.FirstTokenLatency,
		"like_count":          log.LikeCount,
		"dislike_count":       log.DisLikeCount,
		"error_count":         log.ErrorCount,
		"ext":                 log.Ext,
	}
	if err := c.db.WithContext(ctx).Model(&existing).Where("id = ?", existing.ID).Updates(updates).Error; err != nil {
		return toErrStatus("app_conversation_log_record", fmt.Errorf("update conversation log err: %v", err).Error())
	}
	return nil
}
