package orm

import (
	"context"
	"errors"
	"strconv"
	"strings"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
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
			"memory_config":        assistant.MemoryConfig,
			"recommend_config":     assistant.RecommendConfig,
		}).Error; err != nil {
			return toErrStatus("assistant_update", err.Error())
		}
		return nil
	})
}

func (c *Client) DeleteAssistant(ctx context.Context, assistantID uint32, userId, orgId string) *err_code.Status {
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// 1. Main delete with ownership filter — guard against 0 rows to prevent cascading deletes on unowned resources
		result := sqlopt.SQLOptions(
			sqlopt.WithID(assistantID),
			sqlopt.WithUserID(userId),
			sqlopt.WithOrgID(orgId),
		).Apply(tx).Delete(&model.Assistant{})
		if result.Error != nil {
			return toErrStatus("assistant_delete", result.Error.Error())
		}
		if result.RowsAffected == 0 {
			return toErrStatus("assistant_delete", "assistant not found or not owned by user")
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
		if err := sqlopt.WithMultiAgentID(assistantID).Apply(tx).Delete(&model.MultiAgentRelation{}).Error; err != nil {
			return toErrStatus("assistant_delete", err.Error())
		}
		if err := sqlopt.WithAgentID(assistantID).Apply(tx).Delete(&model.MultiAgentRelation{}).Error; err != nil {
			return toErrStatus("assistant_delete", err.Error())
		}
		// 同步删除智能体多版本信息
		if err := sqlopt.WithAssistantID(assistantID).Apply(tx).Delete(&model.AssistantSnapshot{}).Error; err != nil {
			return toErrStatus("assistant_delete", err.Error())
		}
		// 同步删除智能体对话
		if err := sqlopt.WithAssistantID(assistantID).Apply(tx).Delete(&model.Conversation{}).Error; err != nil {
			return toErrStatus("assistant_delete", err.Error())
		}
		return nil
	})
}

func (c *Client) GetAssistant(ctx context.Context, assistantID uint32, userID, orgID string) (*model.Assistant, *err_code.Status) {
	var assistant model.Assistant
	query := sqlopt.SQLOptions(
		sqlopt.WithID(assistantID),
		sqlopt.DataPerm(userID, orgID),
	).Apply(c.db.WithContext(ctx))
	if err := query.First(&assistant).Error; err != nil {
		return nil, toErrStatus("assistant_get", err.Error())
	}
	return &assistant, nil
}

func (c *Client) GetAssistantsByIDs(ctx context.Context, assistantIDs []uint32) ([]*model.Assistant, *err_code.Status) {
	var assistants []*model.Assistant
	if err := sqlopt.WithIDs(assistantIDs).Apply(c.db.WithContext(ctx)).Find(&assistants).Error; err != nil {
		return nil, toErrStatus("assistants_get_by_ids", err.Error())
	}
	return assistants, nil
}

func (c *Client) GetAssistantsByUuids(ctx context.Context, uuids []string) ([]*model.Assistant, *err_code.Status) {
	var assistants []*model.Assistant
	return assistants, c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		if err := sqlopt.WithUuids(uuids).Apply(tx).Find(&assistants).Error; err != nil {
			return toErrStatus("assistants_get_by_uuids", err.Error())
		}
		return nil
	})
}

func (c *Client) GetAssistantByUuid(ctx context.Context, uuid string) (*model.Assistant, *err_code.Status) {
	var assistant model.Assistant
	if err := sqlopt.WithUuid(uuid).Apply(c.db.WithContext(ctx)).
		First(&assistant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toErrStatus("assistant_get_by_uuid", "assistant not found")
		}
		return nil, toErrStatus("assistant_get_by_uuid", err.Error())
	}
	return &assistant, nil
}

func (c *Client) GetAssistantByUuidWithPerm(ctx context.Context, uuid, userId, orgId string) (*model.Assistant, *err_code.Status) {
	var assistant model.Assistant
	if err := sqlopt.SQLOptions(
		sqlopt.WithUuid(uuid),
		sqlopt.DataPerm(userId, orgId),
	).Apply(c.db.WithContext(ctx)).First(&assistant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toErrStatus("assistant_get_by_uuid", "assistant not found or no permission")
		}
		return nil, toErrStatus("assistant_get_by_uuid", err.Error())
	}
	return &assistant, nil
}

func (c *Client) GetAssistantList(ctx context.Context, userID, orgID string, name string) ([]*model.Assistant, int64, *err_code.Status) {
	var assistants []*model.Assistant
	var count int64
	query := sqlopt.SQLOptions(
		sqlopt.WithNameLike(name),
		sqlopt.DataPerm(userID, orgID),
	).Apply(c.db.WithContext(ctx).Model(&model.Assistant{}))

	if err := query.Count(&count).Error; err != nil {
		return nil, 0, toErrStatus("assistants_get_list", err.Error())
	}

	if err := query.Order("updated_at DESC").Find(&assistants).Error; err != nil {
		return nil, 0, toErrStatus("assistants_get_list", err.Error())
	}

	return assistants, count, nil
}

func (c *Client) CheckSameAssistantName(ctx context.Context, userID, orgID, name, assistantID string) *err_code.Status {
	var count int64
	err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithoutUuid(assistantID),
		sqlopt.WithName(name),
	).Apply(c.db.WithContext(ctx)).Model(&model.Assistant{}).Count(&count).Error

	if err != nil {
		return toErrStatus("assistant_get_by_name", err.Error())
	}
	// 存在同名智能体
	if count > 0 {
		return toErrStatus("assistant_same_name", name)
	}
	return nil
}

func (c *Client) CopyAssistant(ctx context.Context, assistant *model.Assistant, workflows []*model.AssistantWorkflow, mcps []*model.AssistantMCP, customTools []*model.AssistantTool, subAgents []*model.MultiAgentRelation, skills []*model.AssistantSkill) (uint32, string, *err_code.Status) {
	// 智能体名称前缀
	prefix := assistant.Name + "_"

	// 查询所有以"原名称_"为前缀的名称
	var existingNames []string
	err := sqlopt.DataPerm(assistant.UserId, assistant.OrgId).Apply(
		c.db.WithContext(ctx).Model(&model.Assistant{}),
	).Where("name LIKE ?", prefix+"%").
		Pluck("name", &existingNames).Error

	if err != nil {
		return 0, "", toErrStatus("assistant_copy", err.Error())
	}

	// 解析名称
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

	// 生成新名称
	newName := util.GenCopyName(assistant.Name, maxNum+1)

	var newAssistantId uint32
	var newAssistantUuid string
	return newAssistantId, newAssistantUuid, c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		// 复制并保存新智能体
		newAssistant := *assistant
		newAssistant.ID = 0
		newAssistant.Name = newName
		newAssistant.UUID = util.NewID() // 生成新的UUID
		if err = tx.Create(&newAssistant).Error; err != nil {
			return toErrStatus("assistant_create", err.Error())
		}
		newAssistantId = newAssistant.ID
		newAssistantUuid = newAssistant.UUID

		// 复制并保存新智能体工作流
		for _, workflow := range workflows {
			workflow.ID = 0
			workflow.AssistantId = newAssistantId
			if err = tx.Create(&workflow).Error; err != nil {
				return toErrStatus("assistant_workflow_create", err.Error())
			}
		}

		// 复制并保存新智能体MCP
		for _, mcp := range mcps {
			mcp.ID = 0
			mcp.AssistantId = newAssistantId
			if err = tx.Create(&mcp).Error; err != nil {
				return toErrStatus("assistant_mcp_create", err.Error())
			}
		}

		// 复制并保存新智能体自定义工具
		for _, tool := range customTools {
			tool.ID = 0
			tool.AssistantId = newAssistantId
			if err = tx.Create(&tool).Error; err != nil {
				return toErrStatus("assistant_tool_create", err.Error())
			}
		}

		// 复制并保存新智能体 -- 多智能体配置
		for _, relation := range subAgents {
			relation.Id = 0
			relation.MultiAgentId = newAssistantId
			if err = tx.Create(&relation).Error; err != nil {
				return toErrStatus("assistant_multi_agent_create", err.Error())
			}
		}

		// 复制并保存新智能体skills
		for _, skill := range skills {
			skill.ID = 0
			skill.AssistantId = newAssistantId
			if err = tx.Create(&skill).Error; err != nil {
				return toErrStatus("assistant_skill_create", err.Error())
			}
		}
		return nil
	})
}

// AdminGetAssistantListAll 管理员中心智能体列表（返回全部）
func (c *Client) AdminGetAssistantListAll(ctx context.Context, userIds, orgIds []string, name string, categories []int32) ([]*model.AdminAssistantItem, int64, *err_code.Status) {
	var assistants []*model.AdminAssistantItem
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserIDs(userIds),
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithNameLike(name),
		sqlopt.WithCategories(categories),
	).Apply(c.db.WithContext(ctx).Model(&model.Assistant{})).
		Order("updated_at DESC").
		Find(&assistants).Error; err != nil {
		return nil, 0, toErrStatus("assistants_admin_list", err.Error())
	}

	return assistants, int64(len(assistants)), nil
}

// AdminGetAssistantListPage 管理员中心智能体列表（分页返回）
func (c *Client) AdminGetAssistantListPage(ctx context.Context, userIds, orgIds []string, name string, categories []int32, pageNum, pageSize int) ([]*model.AdminAssistantItem, int64, *err_code.Status) {
	var assistants []*model.AdminAssistantItem
	var total int64
	query := sqlopt.SQLOptions(
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithUserIDs(userIds),
		sqlopt.WithNameLike(name),
		sqlopt.WithCategories(categories),
	).Apply(c.db.WithContext(ctx).Model(&model.Assistant{}))
	if err := query.
		Count(&total).
		Order("updated_at DESC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&assistants).Error; err != nil {
		return nil, 0, toErrStatus("assistants_admin_list", err.Error())
	}
	return assistants, total, nil
}
