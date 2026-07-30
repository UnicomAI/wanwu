package orm

import (
	"context"
	"fmt"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
)

// CreateApiKey 创建API Key
func (c *Client) CreateApiKey(ctx context.Context, userId, orgId, name, desc string, expiredAt int64, apiKey string) (*model.OpenApiKey, *errs.Status) {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
		sqlopt.WithName(name),
	).Apply(c.db.WithContext(ctx)).Model(&model.OpenApiKey{}).Count(&count).Error; err != nil {
		return nil, toErrStatus("api_key_create", err.Error())
	}
	if count > 0 {
		return nil, toErrStatus("api_key_create", fmt.Sprintf("name %v duplicate", name))
	}
	newKey := &model.OpenApiKey{
		OrgID:       orgId,
		UserID:      userId,
		Name:        name,
		Description: desc,
		ExpiredAt:   expiredAt,
		Key:         apiKey,
		Status:      true, // 默认启用
	}
	if err := c.db.WithContext(ctx).Create(newKey).Error; err != nil {
		return nil, toErrStatus("api_key_create", err.Error())
	}
	return newKey, nil
}

// DeleteApiKey 删除API Key
func (c *Client) DeleteApiKey(ctx context.Context, keyId uint32, userId, orgId string) *errs.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(keyId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Delete(&model.OpenApiKey{}).Error; err != nil {
		return toErrStatus("api_key_delete", util.Int2Str(keyId), err.Error())
	}
	return nil
}

// UpdateApiKey 更新API Key信息
func (c *Client) UpdateApiKey(ctx context.Context, keyId uint32, userId, orgId, name, desc string, expiredAt int64) *errs.Status {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
		sqlopt.WithName(name),
	).Apply(c.db.WithContext(ctx)).Where("id != ?", keyId).Model(&model.OpenApiKey{}).Count(&count).Error; err != nil {
		return toErrStatus("api_key_update", err.Error())
	}
	if count > 0 {
		return toErrStatus("api_key_update", fmt.Sprintf("name %v duplicate", name))
	}
	updateFields := map[string]any{
		"name":        name,
		"description": desc,
	}
	if expiredAt >= 0 {
		updateFields["expired_at"] = expiredAt
	}
	result := sqlopt.WithID(keyId).Apply(c.db.WithContext(ctx)).Model(&model.OpenApiKey{}).Updates(updateFields)
	if result.Error != nil {
		return toErrStatus("api_key_update", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return toErrStatus("api_key_update", fmt.Sprintf("api key not found with id: %d", keyId))
	}
	return nil
}

// ListApiKeys 查询API Key列表
func (c *Client) ListApiKeys(ctx context.Context, orgIds, userIds []string, offset, limit int32) ([]*model.OpenApiKey, int64, *errs.Status) {
	var count int64
	var keys []*model.OpenApiKey
	opts := []sqlopt.SQLOption{
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithUserIDs(userIds),
	}
	if err := sqlopt.SQLOptions(opts...).Apply(c.db.WithContext(ctx)).
		Offset(int(offset)).Limit(int(limit)).Order("id DESC").Find(&keys).
		Offset(-1).Limit(-1).Count(&count).Error; err != nil {
		return nil, 0, toErrStatus("api_key_list", err.Error())
	}
	return keys, count, nil
}

// UpdateApiKeyStatus 更新API Key状态
func (c *Client) UpdateApiKeyStatus(ctx context.Context, keyId uint32, status bool) *errs.Status {
	result := sqlopt.WithID(keyId).Apply(c.db.WithContext(ctx)).Model(&model.OpenApiKey{}).Updates(map[string]interface{}{
		"status": status,
	})
	if result.Error != nil {
		return toErrStatus("api_key_update_status", result.Error.Error())
	}
	if result.RowsAffected == 0 {
		return toErrStatus("api_key_update_status", fmt.Sprintf("api key not found with id: %d", keyId))
	}
	return nil
}

// GetApiKeyByKey 根据Key获取API Key信息
func (c *Client) GetApiKeyByKey(ctx context.Context, key string) (*model.OpenApiKey, *errs.Status) {
	var apiKey model.OpenApiKey
	if err := sqlopt.SQLOptions(
		sqlopt.WithKey(key),
	).Apply(c.db.WithContext(ctx)).First(&apiKey).Error; err != nil {
		return nil, toErrStatus("api_key_get_by_key", err.Error())
	}
	return &apiKey, nil
}
