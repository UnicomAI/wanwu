package orm

import (
	"context"
	"strconv"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) CreateAssistant(ctx context.Context, assistant *model.Assistant) *err_code.Status {
	if assistant.ID != 0 {
		return toErrStatus("assistant_create", "create assistant but id not 0")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		if err := tx.Create(assistant).Error; err != nil {
			return toErrStatus("assistant_create", err.Error())
		}
		return nil
	})
}

func (c *Client) UpdateAssistant(ctx context.Context, assistant *model.Assistant) *err_code.Status {
	if assistant.ID == 0 {
		return toErrStatus("assistant_update", "update assistant but id 0")
	}
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		if err := tx.Model(assistant).Updates(map[string]interface{}{
			"avatar_path":          assistant.AvatarPath,
			"name":                 assistant.Name,
			"desc":                 assistant.Desc,
			"instructions":         assistant.Instructions,
			"prologue":             assistant.Prologue,
			"recommend_question":   assistant.RecommendQuestion,
			"model_config":         assistant.ModelConfig,
			"knowledgebase_config": assistant.KnowledgebaseConfig,
			"scope":                assistant.Scope,
			"rerank_config":        assistant.RerankConfig,
			"safety_config":        assistant.SafetyConfig,
			"vision_config":        assistant.VisionConfig,
		}).Error; err != nil {
			return toErrStatus("assistant_update", err.Error())
		}
		return nil
	})
}

func (c *Client) DeleteAssistant(ctx context.Context, assistantID uint32) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		if err := sqlopt.WithID(assistantID).Apply(tx).Delete(&model.Assistant{}).Error; err != nil {
			return toErrStatus("assistant_delete", err.Error())
		}
		if err := sqlopt.WithAssistantID(assistantID).Apply(tx).Delete(&model.AssistantWorkflow{}).Error; err != nil {
			return toErrStatus("assistant_delete", err.Error())
		}
		if err := sqlopt.WithAssistantID(assistantID).Apply(tx).Delete(&model.AssistantMCP{}).Error; err != nil {
			return toErrStatus("assistant_delete", err.Error())
		}
		if err := sqlopt.WithAssistantID(assistantID).Apply(tx).Delete(&model.AssistantTool{}).Error; err != nil {
			return toErrStatus("assistant_delete", err.Error())
		}
		return nil
	})
}

func (c *Client) GetAssistant(ctx context.Context, assistantID uint32, userID, orgID string) (*model.Assistant, *err_code.Status) {
	var assistant *model.Assistant
	return assistant, c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		assistant = &model.Assistant{}
		query := sqlopt.SQLOptions(
			sqlopt.WithID(assistantID),
			sqlopt.DataPerm(userID, orgID),
		).Apply(tx)
		if err := query.First(assistant).Error; err != nil {
			return toErrStatus("assistant_get", err.Error())
		}
		return nil
	})
}

func (c *Client) GetAssistantsByIDs(ctx context.Context, assistantIDs []uint32) ([]*model.Assistant, *err_code.Status) {
	var assistants []*model.Assistant
	return assistants, c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		if err := sqlopt.WithIDs(assistantIDs).Apply(tx).Find(&assistants).Error; err != nil {
			return toErrStatus("assistants_get_by_ids", err.Error())
		}
		return nil
	})
}

func (c *Client) GetAssistantList(ctx context.Context, userID, orgID string, name string) ([]*model.Assistant, int64, *err_code.Status) {
	var assistants []*model.Assistant
	var count int64
	return assistants, count, c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		query := sqlopt.DataPerm(userID, orgID).Apply(tx.Model(&model.Assistant{}))

		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
		}

		if err := query.Count(&count).Error; err != nil {
			return toErrStatus("assistants_get_list", err.Error())
		}

		if err := query.Order("created_at DESC").Find(&assistants).Error; err != nil {
			return toErrStatus("assistants_get_list", err.Error())
		}

		return nil
	})
}

func (c *Client) CheckSameAssistantName(ctx context.Context, userID, orgID, name, assistantID string) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		query := sqlopt.SQLOptions(
			sqlopt.WithUserId(userID),
			sqlopt.WithOrgID(orgID),
		).Apply(tx.Model(&model.Assistant{}))

		if assistantID != "" {
			id, _ := strconv.ParseUint(assistantID, 10, 32)
			query = query.Where("id != ?", uint32(id))
		}

		if name != "" {
			query = query.Where("name = ?", name)
		}
		var count int64
		if err := query.Count(&count).Error; err != nil {
			return toErrStatus("assistant_get_by_name", err.Error())
		}

		// An agent with the same name exists
		if count > 0 {
			return toErrStatus("assistant_same_name", name)
		}
		return nil
	})
}

func (c *Client) CopyAssistant(ctx context.Context, assistant *model.Assistant, workflows []*model.AssistantWorkflow, mcps []*model.AssistantMCP, customTools []*model.AssistantTool) (uint32, *err_code.Status) {
	// Agent name prefix
	prefix := assistant.Name + "_"

	// Query all names prefixed with "original name_"
	var existingNames []string
	err := c.db.WithContext(ctx).Model(&model.Assistant{}).
		Where("name LIKE ?", prefix+"%").
		Pluck("name", &existingNames).Error

	if err != nil {
		return 0, toErrStatus("assistant_copy", err.Error())
	}

	// Parse name
	maxNum := 0
	for _, name := range existingNames {
		numStr := strings.TrimPrefix(name, prefix)
		num, err := strconv.Atoi(numStr)
		if err != nil {
			continue
		}
		if num > maxNum {
			maxNum = num
		}
	}

	// Generate new name
	newName := prefix + strconv.Itoa(maxNum+1)

	var newAssistantId uint32
	return newAssistantId, c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// Copy and save the new agent
		newAssistant := *assistant
		newAssistant.ID = 0
		newAssistant.Name = newName
		if err = tx.Create(&newAssistant).Error; err != nil {
			return toErrStatus("assistant_create", err.Error())
		}
		newAssistantId = newAssistant.ID

		// Copy and save the new agent workflow
		for _, workflow := range workflows {
			workflow.ID = 0
			workflow.AssistantId = newAssistantId
			if err = tx.Create(&workflow).Error; err != nil {
				return toErrStatus("assistant_workflow_create", err.Error())
			}
		}

		// Copy and save the new agent MCP
		for _, mcp := range mcps {
			mcp.ID = 0
			mcp.AssistantId = newAssistantId
			if err = tx.Create(&mcp).Error; err != nil {
				return toErrStatus("assistant_mcp_create", err.Error())
			}
		}

		// Copy and save the new agent customizer
		for _, tool := range customTools {
			tool.ID = 0
			tool.AssistantId = newAssistantId
			if err = tx.Create(&tool).Error; err != nil {
				return toErrStatus("assistant_tool_create", err.Error())
			}
		}
		return nil
	})
}
