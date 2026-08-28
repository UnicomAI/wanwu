package sqlopt

import "gorm.io/gorm"

// WithChannelID 按通道 ID 查询
func WithChannelID(channelID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if channelID != "" {
			return db.Where("channel_id = ?", channelID)
		}
		return db
	})
}

// WithChannelType 按通道类型查询
func WithChannelType(channelType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if channelType != "" {
			return db.Where("channel_type = ?", channelType)
		}
		return db
	})
}

// WithEnabled 按启用状态查询
func WithEnabled(enabled bool) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("enabled = ?", enabled)
	})
}

// WithStatus 按状态查询
func WithStatus(status string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if status != "" {
			return db.Where("status = ?", status)
		}
		return db
	})
}

// WithChannelName 按渠道名称模糊查询
func WithChannelName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if name != "" {
			return db.Where("name LIKE ?", "%"+name+"%")
		}
		return db
	})
}

// WithOrgID 按组织 ID 查询
func WithOrgID(orgID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if orgID != "" {
			return db.Where("org_id = ?", orgID)
		}
		return db
	})
}

// WithUserID 按用户 ID 查询
func WithUserID(userID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if userID != "" {
			return db.Where("user_id = ?", userID)
		}
		return db
	})
}

// --- 以下为 sqlopt 通用选项，从 app-service 复制基础结构 ---

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
