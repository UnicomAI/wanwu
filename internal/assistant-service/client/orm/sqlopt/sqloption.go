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

func WithUserId(userId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userId)
	})
}

func DataPerm(userId, orgId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if userId != "" && orgId == "" {
			//数据权限：所有org内本人，userId传有效值，orgId不传有效值 [EN] Data permissions: For all individuals in the org, valid values ​​are passed for userId, and no valid values ​​are passed for orgId.
			return SQLOptions(
				WithUserId(userId),
			).Apply(db)
		} else if userId != "" && orgId != "" {
			//数据权限：本org内本人，userId和orgId都需要传有效值 [EN] Data permissions: The person in this org, userId and orgId all need to pass valid values.
			return SQLOptions(
				WithUserId(userId),
				WithOrgID(orgId),
			).Apply(db)
		} else if userId == "" && orgId != "" {
			//数据权限：本org内全部，userId不传有效值，orgId传有效值 [EN] Data permissions: All in this org, userId does not pass a valid value, orgId passes a valid value
			return SQLOptions(
				WithOrgID(orgId),
			).Apply(db)
		} else {
			//数据权限：全部 [EN] Data permissions: All
			return db
		}
	})
}

func WithAssistantID(assistantId uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("assistant_id = ?", assistantId)
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
