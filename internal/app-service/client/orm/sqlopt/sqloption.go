package sqlopt

import (
	"github.com/UnicomAI/wanwu/pkg/constant"
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

func WithOrgIDs(orgIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(orgIDs) > 0 {
			return db.Where("org_id IN ?", orgIDs)
		}
		return db
	})
}

func WithUserIDs(userIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(userIDs) > 0 {
			return db.Where("user_id IN ?", userIDs)
		}
		return db
	})
}

func WithAppIDs(Ids []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("app_id IN (?)", Ids)
	})
}

// WithAppIDsForStatistic 统计查询时使用，增加app_id过滤，但不强制要求app_id必须存在，目前应用于应用统计
func WithAppIDsForStatistic(Ids []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(Ids) > 0 {
			return db.Where("app_id IN (?)", Ids)
		}
		return db
	})
}

func WithExcludeUserID(userID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if userID != "" {
			return db.Where("user_id != ?", userID)
		}
		return db
	})
}

func WithSuffix(suffix string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("suffix = ?", suffix)
	})
}

func WithAppID(appID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if appID != "" {
			return db.Where("app_id = ?", appID)
		}
		return db
	})
}

func WithAppType(appType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if appType != "" {
			return db.Where("app_type = ?", appType)
		}
		return db
	})
}

func WithAppTypes(appTypes []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(appTypes) > 0 {
			return db.Where("app_type IN ?", appTypes)
		}
		return db
	})
}

func WithID(id uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
}

func WithIDs(Ids []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("id IN (?)", Ids)
	})
}

func WithAppKey(appKey string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if appKey != "" {
			return db.Where("api_key = ?", appKey)
		}
		return db
	})
}

func WithKey(apiKey string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if apiKey != "" {
			return db.Where("`key` = ?", apiKey)
		}
		return db
	})
}

func InAppIds(appIds []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(appIds) > 0 {
			return db.Where("app_id IN ?", appIds)
		}
		return db
	})
}

func Private() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("publish_type = ? ", constant.AppPublishPrivate)
	})
}

func StartCreatedAt(startCreatedAt int64) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if startCreatedAt != 0 {
			return db.Where("created_at >= ?", startCreatedAt)
		}
		return db
	})
}

func EndCreatedAt(endCreatedAt int64) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if endCreatedAt != 0 {
			return db.Where("created_at <= ?", endCreatedAt)
		}
		return db
	})
}

func StartUpdatedAt(startUpdatedAt int64) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if startUpdatedAt != 0 {
			return db.Where("updated_at >= ?", startUpdatedAt)
		}
		return db
	})
}

func EndUpdatedAt(endUpdatedAt int64) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if endUpdatedAt != 0 {
			return db.Where("updated_at <= ?", endUpdatedAt)
		}
		return db
	})
}

func WithSearchType(userID, orgID, searchType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		var query string
		var args []interface{}
		switch searchType {
		case "", "all":
			query = "(user_id = ? AND org_id = ? AND publish_type = ?) OR (publish_type = ?) OR (publish_type = ? AND org_id = ?)"
			args = append(args, userID, orgID, constant.AppPublishPrivate, constant.AppPublishPublic, constant.AppPublishOrganization, orgID)
		case "private":
			query = "user_id = ? AND org_id = ? AND publish_type = ?"
			args = append(args, userID, orgID, constant.AppPublishPrivate)
		}
		return db.Where(query, args...)
	})
}

func WithTableID(tableID uint32) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("table_id = ?", tableID)
	})
}

func WithTableIDs(tableIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("table_id IN (?)", tableIDs)
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

func WithContent(content string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if content != "" {
			return db.Where("content = ?", content)
		}
		return db
	})
}

func WithContents(contents []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("content IN ?", contents)
	})
}

func WithConversationID(id string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("conversation_id = ?", id)
	})
}

func WithWorkflowID(workflowId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if workflowId != "" {
			return db.Where("workflow_id = ?", workflowId)
		}
		return db
	})
}

func WithApplicationID(applicationId string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if applicationId != "" {
			return db.Where("application_id = ?", applicationId)
		}
		return db
	})
}

func WithTableType(tableType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("table_type = ?", tableType)
	})
}

func WithModelID(modelID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if modelID != "" {
			return db.Where("model_id = ?", modelID)
		}
		return db
	})
}

func WithModelIds(modelIds []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(modelIds) > 0 {
			return db.Where("model_id IN ?", modelIds)
		}
		return db
	})
}

func WithProvider(provider string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if provider != "" {
			return db.Where("provider = ?", provider)
		}
		return db
	})
}

func WithDate(date string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if date != "" {
			return db.Where("date = ?", date)
		}
		return db
	})
}

func WithModelType(modelType string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if modelType != "" {
			return db.Where("model_type = ?", modelType)
		}
		return db
	})
}

func StartDate(startDate string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if startDate != "" {
			return db.Where("date >= ?", startDate)
		}
		return db
	})
}

func EndDate(endDate string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if endDate != "" {
			return db.Where("date <= ?", endDate)
		}
		return db
	})
}

func WithAPIKeyIds(apiKeyIds []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(apiKeyIds) > 0 {
			return db.Where("api_key_id IN ?", apiKeyIds)
		}
		return db
	})
}

func WithMethodPaths(methodPaths []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(methodPaths) > 0 {
			return db.Where("method_path IN ?", methodPaths)
		}
		return db
	})
}

func WithModule(module string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if module != "" {
			return db.Where("module = ?", module)
		}
		return db
	})
}

func WithSource(source string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if source != "" {
			return db.Where("source = ?", source)
		}
		return db
	})
}

func WithModuleCreatorOrgIDs(orgIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(orgIDs) > 0 {
			return db.Where("module_creator_org_id IN ?", orgIDs)
		}
		return db
	})
}

func WithModuleCreatorUserIDs(userIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(userIDs) > 0 {
			return db.Where("module_creator_user_id IN ?", userIDs)
		}
		return db
	})
}

func WithModelCreatorOrgIDs(orgIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(orgIDs) > 0 {
			return db.Where("model_creator_org_id IN ?", orgIDs)
		}
		return db
	})
}

func WithModelCreatorUserIDs(userIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(userIDs) > 0 {
			return db.Where("model_creator_user_id IN ?", userIDs)
		}
		return db
	})
}

// WithStatisticAppIDFilter 应用统计查询的 app_id 。
// - 允许空 appId 的板块（wga/model/skill/knowledge/prompt）：不加限制
// - module 空（查全部）：保留有 appId 的行，以及允许空 appId 的板块级行
// - 其它指定 module：要求 app_id 非空
func WithStatisticAppIDFilter(module string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if constant.StatisticModuleAllowsEmptyAppID(module) {
			return db
		}
		if module == "" {
			return db.Where("(app_id != '' AND app_id IS NOT NULL) OR module IN ?", []string{
				constant.BizModuleWGA,
				constant.BizModuleModel,
				constant.BizModuleResourceSkill,
				constant.BizModuleResourceKnowledge,
				constant.BizModuleResourcePrompt,
			})
		}
		return db.Where("app_id != '' AND app_id IS NOT NULL")
	})
}

// WithStatisticAppID 钻取定位主表行：非空按 app_id 精确匹配；空串匹配空 app_id（板块级行）。
func WithStatisticAppID(appID string) SQLOption {
	if appID != "" {
		return WithAppID(appID)
	}
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("(app_id = '' OR app_id IS NULL)")
	})
}

func WithNonEmptyModelID() SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		return db.Where("model_id != '' AND model_id IS NOT NULL")
	})
}

func WithAPIKeyID(apiKeyID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if apiKeyID != "" {
			return db.Where("api_key_id = ?", apiKeyID)
		}
		return db
	})
}

func WithMethodPath(methodPath string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if methodPath != "" {
			return db.Where("method_path = ?", methodPath)
		}
		return db
	})
}

func WithSources(sources []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(sources) > 0 {
			return db.Where("source IN ?", sources)
		}
		return db
	})
}

func WithExportIDs(exportIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(exportIDs) > 0 {
			return db.Where("export_id IN ?", exportIDs)
		}
		return db
	})
}

func LikeTitle(title string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if title != "" {
			return db.Where("title LIKE ?", "%"+title+"%")
		}
		return db
	})
}

func WithExportID(exportID string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if exportID != "" {
			return db.Where("export_id = ?", exportID)
		}
		return db
	})
}

func WithLogIDs(logIDs []string) SQLOption {
	return funcSQLOption(func(db *gorm.DB) *gorm.DB {
		if len(logIDs) > 0 {
			return db.Where("log_id IN ?", logIDs)
		}
		return db
	})
}
