package sqlopt

import (
	"gorm.io/gorm"
)

// docFailStatus 文档解析失败相关的 status 码（上传/切片/向量化等环节失败）
var docFailStatus = []int{5, 51, 52, 53, 54, 55, 56, 61, 62}

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

func WithQAPairID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("qa_pair_id = ?", id)
	})
}

func WithQuestionMd5(questionMd5 string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(questionMd5) > 0 {
			return db.Where("question_md5 = ?", questionMd5)
		}
		return db
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

func WithCategory(category int) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("category = ?", category)
	})
}

func WithCategoryList(categoryList []int) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(categoryList) == 0 {
			return db
		}
		return db.Where("category IN ?", categoryList)
	})
}

func WithExternal(external int) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if external == -1 {
			return db
		}
		return db.Where("external = ?", external)
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

func WithExportID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("export_id = ?", id)
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

func WithDocIDsNonEmpty(ids []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(ids) == 0 {
			return db
		}
		return db.Where("doc_id in ?", ids)
	})
}

func WithQAPairIDsNonEmpty(ids []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(ids) == 0 {
			return db
		}
		return db.Where("qa_pair_id in ?", ids)
	})
}

func WithFileTypeFilter(fileType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(fileType) > 0 {
			return db.Where("file_type != ?", fileType)
		}
		return db
	})
}

func WithQAPairIDs(ids []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("qa_pair_id in ?", ids)
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

func WithNonType(metaValueType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("value_type != ?", metaValueType)
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

// WithPermit 权限查询条件
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

// WithPermitList 权限查询条件
func WithPermitList(orgID, userID []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(orgID) > 0 {
			db = db.Where("org_id IN ?", orgID)
		}
		if len(userID) > 0 {
			db = db.Where("user_id IN ?", userID)
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

func WithGraphStatusList(graphStatus []int) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(graphStatus) == 0 {
			return db
		}
		return db.Where("graph_status IN ? AND status NOT IN ?", graphStatus, docFailStatus)
	})
}

func WithStatus(status int) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if status == -1 {
			return db
		}
		return db.Where("status = ?", status)
	})
}

func WithQAStatusList(status []int32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(status) == 0 {
			return db
		} else if len(status) == 1 {
			if status[0] == -1 {
				return db
			}
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

func WithoutDocId(docId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(docId) > 0 {
			return db.Where("doc_id != ?", docId)
		}
		return db
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

func WithRagName(name string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(name) > 0 {
			return db.Where("rag_name = ?", name)
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
			// 使用 OR 条件组合模糊查询
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

func LikeMetaValue(value string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if value != "" {
			return db.Where("value_main LIKE ?", "%"+value+"%")
		}
		return db
	})
}

// WithValueType 按元数据类型(value_type)精确过滤
func WithValueType(valueType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if valueType != "" {
			return db.Where("value_type = ?", valueType)
		}
		return db
	})
}

// BetweenMetaValueTime 按时间区间过滤 value_main(存储为毫秒时间戳字符串)
// 时间戳为等宽非负十进制(13位毫秒在 2286 年前均为 13 位)，字符串比较与数值比较等价，可跨数据库方言。
// start/end 为空时对应一侧不限制；闭区间 [start, end]。
func BetweenMetaValueTime(start, end string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if start != "" {
			db = db.Where("value_main >= ?", start)
		}
		if end != "" {
			db = db.Where("value_main <= ?", end)
		}
		return db
	})
}

func LikeQuestion(question string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if question != "" {
			return db.Where("question LIKE ?", "%"+question+"%")
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

func WithExternalAPIId(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(id) > 0 {
			return db.Where("external_api_id = ?", id)
		}
		return db
	})
}

func WithExternalAPIIdList(ids []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(ids) > 0 {
			return db.Where("external_api_id IN ?", ids)
		}
		return db
	})
}
