package sqlopt

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		if id > 0 {
			return db.Where("id = ?", id)
		}
		return db
	})
}

func WithIDs(ids []uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	})
}

func WithOrgID(orgID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if orgID != "" {
			return db.Where("org_id = ?", orgID)
		}
		return db
	})
}

func WithUserID(userID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if userID != "" {
			return db.Where("user_id = ?", userID)
		}
		return db
	})
}

// WithOrgIDList 按组织id列表过滤，空列表不限制
func WithOrgIDList(orgIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(orgIDs) > 0 {
			return db.Where("org_id IN ?", orgIDs)
		}
		return db
	})
}

// WithUserIDList 按用户id列表过滤，空列表不限制
func WithUserIDList(userIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(userIDs) > 0 {
			return db.Where("user_id IN ?", userIDs)
		}
		return db
	})
}

func WithMcpSquareId(mcpSquareId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if mcpSquareId != "" {
			return db.Where("mcp_square_id = ?", mcpSquareId)
		}
		return db
	})
}

func WithName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if name != "" {
			return db.Where("name = ?", name)
		}
		return db
	})
}

func LikeName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if name != "" {
			return db.Where("name LIKE ?", "%"+name+"%")
		}
		return db
	})
}

func WithFrom(from string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if from != "" {
			return db.Where("from = ?", from)
		}
		return db
	})
}

func WithUpdateLock() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Clauses(clause.Locking{
			Strength: "UPDATE",
		})
	})
}

func WithToolSquareID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if id != "" {
			return db.Where("tool_square_id = ?", id)
		}
		return db
	})
}

func WithToolSquareIDList(idList []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(idList) > 0 {
			return db.Where("tool_square_id IN ?", idList)
		}
		return db
	})
}

func WithToolSquareIDEmpty() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("tool_square_id = '' or tool_square_id IS NULL")
	})
}

func WithToolSquareIDNotEmpty() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("tool_square_id != '' and tool_square_id IS NOT NULL")
	})
}

func WithMcpServerId(mcpServerId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if mcpServerId != "" {
			return db.Where("mcp_server_id = ?", mcpServerId)
		}
		return db
	})
}

func WithMcpServerToolId(mcpServerToolId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if mcpServerToolId != "" {
			return db.Where("mcp_server_tool_id = ?", mcpServerToolId)
		}
		return db
	})
}

func WithMcpServerIdList(mcpServerIdList []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("mcp_server_id IN ?", mcpServerIdList)
	})
}

func WithCustomSkillSaveIds(saveIdList []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(saveIdList) > 0 {
			return db.Where("save_id IN ?", saveIdList)
		}
		return db
	})
}

func WithCustomSkillSaveId(saveId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if saveId != "" {
			return db.Where("save_id = ?", saveId)
		}
		return db
	})
}

func WithCustomSkillSourceType(sourceType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if sourceType != "" {
			return db.Where("source_type = ?", sourceType)
		}
		return db
	})
}

func WithCustomSkillPreviewThreadId(previewThreadID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if previewThreadID != "" {
			return db.Where("preview_thread_id = ?", previewThreadID)
		}
		return db
	})
}

func WithCustomSkillWgaThreadId(wgaThreadId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if wgaThreadId != "" {
			return db.Where("wga_thread_id = ?", wgaThreadId)
		}
		return db
	})
}

func WithCustomSkillWgaThreadIdList(wgaThreadIdList []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(wgaThreadIdList) > 0 {
			return db.Where("wga_thread_id IN ?", wgaThreadIdList)
		}
		return db
	})
}

func WithCustomSkillSkillId(skillIds []string) SQLOption {
	if len(skillIds) > 0 {
		return funcSQLOption(func(db *gorm.DB) *gorm.DB {
			return db.Where("id IN ?", skillIds)
		})
	}
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db
	})
}

func WithSkillID(skillID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if skillID != "" {
			return db.Where("skill_id = ?", skillID)
		}
		return db
	})
}

func WithStatType(statType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if statType != "" {
			return db.Where("stat_type = ?", statType)
		}
		return db
	})
}

func WithSkillIDs(skillIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(skillIDs) > 0 {
			return db.Where("skill_id IN ?", skillIDs)
		}
		return db
	})
}

func WithCustomSkillID(customSkillID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if customSkillID != "" {
			return db.Where("custom_skill_id = ?", customSkillID)
		}
		return db
	})
}

func WithCustomSkillIDs(customSkillIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(customSkillIDs) > 0 {
			return db.Where("custom_skill_id IN ?", customSkillIDs)
		}
		return db
	})
}

func WithAcquiredSkillID(acquiredSkillID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if acquiredSkillID != "" {
			return db.Where("acquired_skill_id = ?", acquiredSkillID)
		}
		return db
	})
}

func WithAcquiredSkillIDs(acquiredSkillIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(acquiredSkillIDs) > 0 {
			return db.Where("acquired_skill_id IN ?", acquiredSkillIDs)
		}
		return db
	})
}

func WithVariableName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if name != "" {
			return db.Where("name = ?", name)
		}
		return db
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
