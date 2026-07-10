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

func (c *Client) GetWgaConversationConfig(ctx context.Context, threadId string, userId, orgId string) (*model.WgaConversationConfig, *err_code.Status) {
	var config model.WgaConversationConfig
	if err := sqlopt.SQLOptions(
		sqlopt.WithThreadID(threadId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Model(&model.WgaConversationConfig{}).First(&config).Error; err != nil {
		return nil, toErrStatus("wga_config_get", err.Error())
	}
	return &config, nil
}

func (c *Client) UpdateWgaConversationConfig(ctx context.Context, config *model.WgaConversationConfig) *err_code.Status {
	var existing model.WgaConversationConfig
	err := sqlopt.SQLOptions(
		sqlopt.WithThreadID(config.ThreadID),
		sqlopt.WithUserID(config.UserID),
		sqlopt.WithOrgID(config.OrgID),
	).Apply(c.db.WithContext(ctx)).First(&existing).Error

	if err == nil {
		updates := map[string]interface{}{
			"updated_at": time.Now().UnixMilli(),
		}
		if config.ModelConfig != "" {
			updates["model_config"] = config.ModelConfig
		}
		if config.Title != "" {
			updates["title"] = config.Title
		}
		result := c.db.WithContext(ctx).Model(&existing).Updates(updates)
		if result.Error != nil {
			return toErrStatus("wga_config_update", result.Error.Error())
		}
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return toErrStatus("wga_config_get", err.Error())
	}

	if err := c.db.WithContext(ctx).Create(config).Error; err != nil {
		return toErrStatus("wga_config_create", err.Error())
	}
	return nil
}

func (c *Client) CreateWgaConversationConfig(ctx context.Context, config *model.WgaConversationConfig) *err_code.Status {
	if err := c.db.WithContext(ctx).Create(config).Error; err != nil {
		return toErrStatus("wga_conversation_create", err.Error())
	}
	return nil
}

func (c *Client) DeleteWgaConversationConfig(ctx context.Context, threadId string) *err_code.Status {
	if err := sqlopt.WithThreadID(threadId).Apply(c.db.WithContext(ctx)).Delete(&model.WgaConversationConfig{}).Error; err != nil {
		return toErrStatus("wga_conversation_delete", err.Error())
	}
	return nil
}

func (c *Client) GetWgaConversationConfigList(ctx context.Context, userID, orgID string, offset, limit int32) ([]*model.WgaConversationConfig, int64, *err_code.Status) {
	var configs []*model.WgaConversationConfig
	var count int64

	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
	).Apply(c.db.WithContext(ctx).Model(&model.WgaConversationConfig{})).Offset(int(offset)).Limit(int(limit)).Order("updated_at DESC").Find(&configs).Error; err != nil {
		return configs, count, toErrStatus("wga_conversation_list", err.Error())
	}

	return configs, int64(len(configs)), nil
}

func (c *Client) WgaConversationConfigExists(ctx context.Context, threadId, userID, orgID string) (bool, *err_code.Status) {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithThreadID(threadId),
	).Apply(c.db.WithContext(ctx).Model(&model.WgaConversationConfig{})).Count(&count).Error; err != nil {
		return false, toErrStatus("wga_conversation_exists", err.Error())
	}
	return count > 0, nil
}

func (c *Client) GetWgaConfig(ctx context.Context, userId, orgId string) (*model.WgaConfig, *err_code.Status) {
	var config model.WgaConfig
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Model(&model.WgaConfig{}).First(&config).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.WgaConfig{
				UserID: userId,
				OrgID:  orgId,
			}, nil
		}
		return nil, toErrStatus("general_agent_tool_config_get", err.Error())
	}
	return &config, nil
}

func (c *Client) UpdateWgaConfig(ctx context.Context, config *model.WgaConfig) *err_code.Status {
	var existing model.WgaConfig
	err := sqlopt.SQLOptions(
		sqlopt.WithUserID(config.UserID),
		sqlopt.WithOrgID(config.OrgID),
	).Apply(c.db.WithContext(ctx)).First(&existing).Error

	if err == nil {
		result := c.db.WithContext(ctx).Model(&existing).Updates(map[string]interface{}{
			"tool_list":               config.ToolList,
			"mcp_list":                config.McpList,
			"workflow_list":           config.WorkflowList,
			"skill_list":              config.SkillList,
			"assistant_list":          config.AssistantList,
			"knowledge_list":          config.KnowledgeList,
			"ontology_knowledge_list": config.OntologyKnowledgeList,
		})
		if result.Error != nil {
			return toErrStatus("general_agent_tool_config_update", result.Error.Error())
		}
		return nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return toErrStatus("general_agent_tool_config_get", err.Error())
	}

	if err := c.db.WithContext(ctx).Create(config).Error; err != nil {
		return toErrStatus("general_agent_tool_config_create", err.Error())
	}
	return nil
}
