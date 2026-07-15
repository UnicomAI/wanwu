package sqlopt

import (
	"gorm.io/gorm"
)

type sqlOptions []SQLOption

func SQLOptions(opts ...SQLOption) sqlOptions {
	return opts
}

func (s sqlOptions) Apply(db *gorm.DB) *gorm.DB {
	for _, opt := range s {
		db = opt.Apply(db)
	}
	return db
}

type SQLOption interface {
	Apply(db *gorm.DB) *gorm.DB
}

type funcSQLOption func(db *gorm.DB) *gorm.DB

func (f funcSQLOption) Apply(db *gorm.DB) *gorm.DB {
	return f(db)
}

func WithID(id uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func WithMultiAgentID(id uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if id == 0 {
			return db
		}
		return db.Where("multi_agent_id = ?", id)
	})
}

func WithAgentID(id uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if id == 0 {
			return db
		}
		return db.Where("agent_id = ?", id)
	})
}

func WithIDs(ids []uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	})
}

func WithOrgID(orgId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("org_id = ?", orgId)
	})
}

func WithUserID(userId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userId)
	})
}

func DataPerm(userId, orgId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if userId != "" && orgId == "" {
			//数据权限：所有org内本人，userId传有效值，orgId不传有效值
			return SQLOptions(
				WithUserID(userId),
			).Apply(db)
		} else if userId != "" && orgId != "" {
			//数据权限：本org内本人，userId和orgId都需要传有效值
			return SQLOptions(
				WithUserID(userId),
				WithOrgID(orgId),
			).Apply(db)
		} else if userId == "" && orgId != "" {
			//数据权限：本org内全部，userId不传有效值，orgId传有效值
			return SQLOptions(
				WithOrgID(orgId),
			).Apply(db)
		} else {
			//数据权限：全部
			return db
		}
	})
}

func WithAssistantID(assistantId uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("assistant_id = ?", assistantId)
	})
}

func WithAssistantIDs(assistantIds []uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(assistantIds) > 0 {
			return db.Where("assistant_id IN ?", assistantIds)
		}
		return db
	})
}

func WithToolId(toolId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("tool_id = ?", toolId)
	})
}

func WithToolType(toolType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("tool_type = ?", toolType)
	})
}

func WithActionName(actionName string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("action_name = ?", actionName)
	})
}

func WithMCPID(mcpId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("mcp_id = ?", mcpId)
	})
}

func WithMCPType(mcpType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("mcp_type = ?", mcpType)
	})
}

func WithWorkflowID(workflowId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("workflow_id = ?", workflowId)
	})
}

func WithCustomPromptNotID(id uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("id != ?", id)
	})
}

func WithCustomPromptName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	})
}

func WithCustomPromptLikeName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("name LIKE ?", "%"+name+"%")
	})
}

func WithVersion(version string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if version != "" {
			return db.Where("version = ?", version)
		}
		return db
	})
}

func WithVersionNonEmpty() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("version != ?", "")
	})
}

func WithVersionIsEmpty() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("version = ?", "")
	})
}

func WithUuid(uuid string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("uuid = ?", uuid)
	})
}

func WithConversationType(conversationType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if conversationType != "" {
			return db.Where("conversation_type = ?", conversationType)
		}
		return db
	})
}

func WithConversationId(conversationId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if conversationId != "" {
			return db.Where("conversation_id = ?", conversationId)
		}
		return db
	})
}

func WithSkillId(skillId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("skill_id = ?", skillId)
	})
}

func WithSkillType(skillType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("skill_type = ?", skillType)
	})
}

func WithThreadID(threadId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if threadId != "" {
			return db.Where("thread_id = ?", threadId)
		}
		return db
	})
}

func WithTitleLike(title string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if title != "" {
			return db.Where("title LIKE ? ", "%"+title+"%")
		}
		return db
	})
}
