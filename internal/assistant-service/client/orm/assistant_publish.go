package orm

import (
	"context"
	"errors"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

// UpsertAssistantPublish 创建或更新智能体发布状态记录
// publishType 非空时 upsert，为空时调用 DeleteAssistantPublish
func (c *Client) UpsertAssistantPublish(ctx context.Context, assistantID uint32, publishType string) *err_code.Status {
	if publishType == "" {
		return c.DeleteAssistantPublish(ctx, assistantID)
	}
	var existing model.AssistantPublish
	err := sqlopt.WithAssistantID(assistantID).Apply(c.db.WithContext(ctx)).First(&existing).Error
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("assistant_publish_upsert_query", err.Error())
		}
		// 不存在则创建
		if err := c.db.WithContext(ctx).Create(&model.AssistantPublish{
			AssistantID: assistantID,
			PublishType: publishType,
		}).Error; err != nil {
			return toErrStatus("assistant_publish_upsert_create", err.Error())
		}
		return nil
	}
	// 已存在则更新
	if err := sqlopt.WithAssistantID(assistantID).Apply(c.db.WithContext(ctx).Model(&model.AssistantPublish{})).
		Update("publish_type", publishType).Error; err != nil {
		return toErrStatus("assistant_publish_upsert_update", err.Error())
	}
	return nil
}

// DeleteAssistantPublish 删除智能体发布状态记录（取消发布时调用）
func (c *Client) DeleteAssistantPublish(ctx context.Context, assistantID uint32) *err_code.Status {
	if err := sqlopt.WithAssistantID(assistantID).
		Apply(c.db.WithContext(ctx)).
		Delete(&model.AssistantPublish{}).Error; err != nil {
		return toErrStatus("assistant_publish_delete", err.Error())
	}
	return nil
}
