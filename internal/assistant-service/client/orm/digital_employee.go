package orm

import (
	"context"
	"errors"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

// CreateDigitalEmployeeConversationConfig 创建数字员工发布会话
func (c *Client) CreateDigitalEmployeeConversationConfig(ctx context.Context, config *model.DigitalEmployeeConversationConfig) *err_code.Status {
	if err := c.db.WithContext(ctx).Create(config).Error; err != nil {
		return toErrStatus("de_conversation_create", err.Error())
	}
	return nil
}

// GetDigitalEmployeeConversationConfig 按会话ID获取数字员工发布会话配置
func (c *Client) GetDigitalEmployeeConversationConfig(ctx context.Context, threadId string, userId, orgId string) (*model.DigitalEmployeeConversationConfig, *err_code.Status) {
	var config model.DigitalEmployeeConversationConfig
	if err := sqlopt.SQLOptions(
		sqlopt.WithThreadID(threadId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Model(&model.DigitalEmployeeConversationConfig{}).First(&config).Error; err != nil {
		return nil, toErrStatus("de_conversation_get", err.Error())
	}
	return &config, nil
}

// GetDigitalEmployeeConversationConfigList 获取数字员工发布会话列表（按 employeeId 维度过滤）
func (c *Client) GetDigitalEmployeeConversationConfigList(ctx context.Context, userID, orgID, employeeID, searchText string, offset, limit int32) ([]*model.DigitalEmployeeConversationConfig, int64, *err_code.Status) {
	var configs []*model.DigitalEmployeeConversationConfig
	var count int64

	db := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithEmployeeID(employeeID),
		sqlopt.WithTitleLike(searchText),
	).Apply(c.db.WithContext(ctx).Model(&model.DigitalEmployeeConversationConfig{}))

	if err := db.Count(&count).Error; err != nil {
		return nil, 0, toErrStatus("de_conversation_list", err.Error())
	}

	if err := db.Offset(int(offset)).Limit(int(limit)).Order("updated_at DESC").Find(&configs).Error; err != nil {
		return nil, 0, toErrStatus("de_conversation_list", err.Error())
	}

	return configs, count, nil
}

// DeleteDigitalEmployeeConversationConfig 删除数字员工发布会话
func (c *Client) DeleteDigitalEmployeeConversationConfig(ctx context.Context, threadId string, userId, orgId string) *err_code.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithThreadID(threadId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Delete(&model.DigitalEmployeeConversationConfig{}).Error; err != nil {
		return toErrStatus("de_conversation_delete", err.Error())
	}
	return nil
}

// TouchDigitalEmployeeConversationConfig 刷新数字员工发布会话的 updated_at（列表按 updated_at DESC 排序，对话后需 touch 以重排到最新）
func (c *Client) TouchDigitalEmployeeConversationConfig(ctx context.Context, threadId, userId, orgId string) *err_code.Status {
	var existing model.DigitalEmployeeConversationConfig
	if err := sqlopt.SQLOptions(
		sqlopt.WithThreadID(threadId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).First(&existing).Error; err != nil {
		return toErrStatus("de_conversation_touch", err.Error())
	}
	if err := c.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"updated_at": time.Now().UnixMilli(),
	}).Error; err != nil {
		return toErrStatus("de_conversation_touch", err.Error())
	}
	return nil
}

// DigitalEmployeeConversationConfigExists 判断数字员工发布会话是否存在
func (c *Client) DigitalEmployeeConversationConfigExists(ctx context.Context, threadId, userID, orgID string) (bool, *err_code.Status) {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithThreadID(threadId),
	).Apply(c.db.WithContext(ctx).Model(&model.DigitalEmployeeConversationConfig{})).Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, toErrStatus("de_conversation_exists", err.Error())
	}
	return count > 0, nil
}

// UpdateDigitalEmployeeConversationConfig 更新数字员工发布会话的模型配置（同时刷新 updated_at）
func (c *Client) UpdateDigitalEmployeeConversationConfig(ctx context.Context, threadId, modelConfig, userId, orgId string) *err_code.Status {
	var existing model.DigitalEmployeeConversationConfig
	if err := sqlopt.SQLOptions(
		sqlopt.WithThreadID(threadId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).First(&existing).Error; err != nil {
		return toErrStatus("de_conversation_update", err.Error())
	}
	if err := c.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
		"model_config": modelConfig,
		"updated_at":   time.Now().UnixMilli(),
	}).Error; err != nil {
		return toErrStatus("de_conversation_update", err.Error())
	}
	return nil
}
