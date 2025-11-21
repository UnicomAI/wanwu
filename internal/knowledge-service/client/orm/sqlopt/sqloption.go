package sqlopt

import (
	"gorm.io/gorm"
)

type SqlOptions []SQLOption

func SQLOptions(opts ...SQLOption) SqlOptions {
	return opts
}

func (s SqlOptions) Apply(db *gorm.DB, model interface{}) *gorm.DB {
	if model != nil {
		db = db.Model(model)
	}
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

func WithKnowledgeID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("knowledge_id = ?", id)
	})
}

func WithOverKnowledgePermission(id int) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("permission_type >= ?", id)
	})
}

func WithPermissionId(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("permission_id = ?", id)
	})
}

func WithoutKnowledgeID(knowledgeId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(knowledgeId) == 0 {
			return db
		}
		return db.Where("knowledge_id != ?", knowledgeId)
	})
}

func WithKnowledgeIDList(idList []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(idList) > 0 {
			return db.Where("knowledge_id IN (?)", idList)
		}
		return db
	})
}

func WithTagID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("tag_id = ?", id)
	})
}

func WithSplitterID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("splitter_id = ?", id)
	})
}

func WithImportID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("import_id = ?", id)
	})
}

func WithImportIDs(idList []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("import_id in ?", idList)
	})
}

func WithDocIDs(ids []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("doc_id in ?", ids)
	})
}

func WithDocID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("doc_id = ?", id)
	})
}

func WithKey(key string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("`key` = ?", key)
	})
}

func WithType(metaValueType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("`type` = ?", metaValueType)
	})
}

func WithIDs(ids []uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN ?", ids)
	})
}

func WithMetaId(metaId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("meta_id = ?", metaId)
	})
}

func WithMetaIds(metaIds []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("meta_id IN ?", metaIds)
	})
}

func WithOrgID(orgID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("org_id = ?", orgID)
	})
}

func WithUserID(userID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id = ?", userID)
	})
}

// WithPermit permission query conditions
func WithPermit(orgID, userID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(orgID) > 0 {
			db = db.Where("org_id = ?", orgID)
		}
		if len(userID) > 0 {
			db = db.Where("user_id = ?", userID)
		}
		return db
	})
}

func WithStatusList(status []int) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(status) == 0 {
			return db
		} else if len(status) == 1 {
			return db.Where("status = ?", status[0])
		}
		return db.Where("status IN ?", status)
	})
}

func WithoutStatus(status int) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("status != ?", status)
	})
}

func WithName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(name) > 0 {
			return db.Where("name = ?", name)
		}
		return db
	})
}

func WithoutID(id uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if id != 0 {
			return db.Where("id != ?", id)
		}
		return db
	})
}

func WithValue(value string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(value) > 0 {
			return db.Where("value = ?", value)
		}
		return db
	})
}

func WithNonEmptyValue() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("value != ''")
	})
}

func WithNameOrValue(name, value string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(name) > 0 || len(value) > 0 {
			return db.Where("name = ? OR value = ?", name, value)
		}
		return db
	})
}

func WithNameOrAliasLike(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(name) > 0 {
			// Combine fuzzy queries using OR conditions
			return db.Where("name LIKE ? OR alias LIKE ?", "%"+name+"%", "%"+name+"%")
		}
		return db
	})
}

func WithFilePathMd5(filePathMd5 string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(filePathMd5) > 0 {
			return db.Where("file_path_md5 = ?", filePathMd5)
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

func LikeTag(tag string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if tag != "" {
			return db.Where("tag LIKE ?", "%"+tag+"%")
		}
		return db
	})
}

func WithDelete(deleted int) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("deleted = ?", deleted)
	})
}
