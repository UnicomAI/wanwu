package sqlopt

import "gorm.io/gorm"

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

func WithClientID(userID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if userID != "" {
			return db.Where("client_id = ?", userID)
		}
		return db
	})
}

func WithKey(key string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if key != "" {
			return db.Where("`key` = ?", key)
		}
		return db
	})
}

func WithMessageID(messageID int64) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("message_id = ?", messageID)
	})
}

func InMessageIDs(messageIDs []int64) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(messageIDs) > 0 {
			return db.Where("message_id IN ?", messageIDs)
		}
		return db
	})
}

func StartDate(date string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if date != "" {
			return db.Where("date >= ?", date)
		}
		return db
	})
}

func EndDate(date string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if date != "" {
			return db.Where("date <= ?", date)
		}
		return db
	})
}
