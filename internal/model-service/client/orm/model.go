package orm

import (
	"context"
	"database/sql"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_client "github.com/UnicomAI/wanwu/internal/model-service/client/model"
	"github.com/UnicomAI/wanwu/internal/model-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/db"
	"gorm.io/gorm"
)

func (c *Client) ImportModel(ctx context.Context, tab *model_client.ModelImported) *errs.Status {
	var err error
	db := c.db.Begin(&sql.TxOptions{Isolation: sql.LevelSerializable}).WithContext(ctx)
	defer func() {
		if err == nil {
			db.Commit()
			return
		}
		db.Rollback()
	}()
	// 查询是否已存在相同的模型(0.2.1版本下掉）
	//if err := sqlopt.SQLOptions(
	//	sqlopt.WithProvider(tab.Provider),
	//	sqlopt.WithModelType(tab.ModelType),
	//	sqlopt.WithModel(tab.Model),
	//	sqlopt.WithOrgID(tab.OrgID),
	//	sqlopt.WithUserID(tab.UserID),
	//).Apply(db).Select("id").First(&model_client.ModelImported{}).Error; err == nil {
	//	return toErrStatus("model_create_err", "model with same identifier exist")
	//} else if err != gorm.ErrRecordNotFound {
	//	// 其他错误
	//	return toErrStatus("model_create_err", err.Error())
	//}

	if tab.DisplayName != "" {
		if err := sqlopt.SQLOptions(
			sqlopt.WithProvider(tab.Provider),
			sqlopt.WithDisplayName(tab.DisplayName),
			sqlopt.WithOrgID(tab.OrgID),
			sqlopt.WithUserID(tab.UserID),
		).Apply(db).Select("id").First(&model_client.ModelImported{}).Error; err == nil {
			return toErrStatus("model_create_err", "model with same display name exist")
		} else if err != gorm.ErrRecordNotFound {
			// 其他错误
			return toErrStatus("model_create_err", err.Error())
		}
	}

	if err = db.Create(tab).Error; err != nil {
		return toErrStatus("model_create_err", err.Error())
	}
	return nil
}

func (c *Client) DeleteModel(ctx context.Context, tab *model_client.ModelImported) *errs.Status {
	// 查询
	var existing model_client.ModelImported
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tab.ID),
		sqlopt.WithOrgID(tab.OrgID),
		sqlopt.WithUserID(tab.UserID),
	).Apply(c.db).WithContext(ctx).Select("id").First(&existing).Error; err != nil {
		return toErrStatus("model_delete_err", err.Error())
	}
	if err := c.db.WithContext(ctx).Delete(existing).Error; err != nil {
		return toErrStatus("model_delete_err", err.Error())
	}
	return nil
}

func (c *Client) UpdateModel(ctx context.Context, tab *model_client.ModelImported) *errs.Status {
	// 模型显示名称判重
	var existing model_client.ModelImported
	if err := sqlopt.SQLOptions(
		sqlopt.WithProvider(tab.Provider),
		sqlopt.WithDisplayName(tab.DisplayName),
		sqlopt.WithOrgID(tab.OrgID),
		sqlopt.WithUserID(tab.UserID),
	).Apply(c.db).WithContext(ctx).Select("id").First(&existing).Error; err == nil {
		if tab.ID != existing.ID {
			return toErrStatus("model_update_err", "model with same display name exist")
		}
	} else if err != gorm.ErrRecordNotFound {
		return toErrStatus("model_update_err", err.Error())
	}
	// 更新
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tab.ID),
	).Apply(c.db).WithContext(ctx).Model(tab).Updates(map[string]interface{}{
		"display_name":    tab.DisplayName,
		"model_desc":      tab.ModelDesc,
		"model_icon_path": tab.ModelIconPath,
		"publish_date":    tab.PublishDate,
		"provider_config": tab.ProviderConfig,
		"scope_type":      tab.ScopeType,
	}).Error; err != nil {
		return toErrStatus("model_update_err", err.Error())
	}
	return nil
}

func (c *Client) ChangeModelStatus(ctx context.Context, tab *model_client.ModelImported) *errs.Status {
	// 查询
	var existing model_client.ModelImported
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tab.ID),
		sqlopt.WithOrgID(tab.OrgID),
		sqlopt.WithUserID(tab.UserID),
	).Apply(c.db).WithContext(ctx).Select("id").First(&existing).Error; err != nil {
		return toErrStatus("model_change_model_status_err", err.Error())
	}
	// 更新
	if err := c.db.WithContext(ctx).Model(existing).Updates(map[string]interface{}{
		"is_active": tab.IsActive,
	}).Error; err != nil {
		return toErrStatus("model_change_model_status_err", err.Error())
	}
	return nil
}

func (c *Client) GetModel(ctx context.Context, tab *model_client.ModelImported) (*model_client.ModelImported, *errs.Status) {
	info := &model_client.ModelImported{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tab.ID),
	).Apply(c.db).WithContext(ctx).First(info).Error; err != nil {
		return nil, toErrStatus("model_get_err", err.Error())
	}
	return info, nil
}

func (c *Client) GetModelByUUID(ctx context.Context, uuid string) (*model_client.ModelImported, *errs.Status) {
	info := &model_client.ModelImported{}
	if err := sqlopt.WithUUID(uuid).Apply(c.db).WithContext(ctx).First(info).Error; err != nil {
		return nil, toErrStatus("model_get_by_uuid_err", err.Error())
	}
	return info, nil
}

func (c *Client) ListModelsByUuids(ctx context.Context, uuids []string) ([]*model_client.ModelImported, *errs.Status) {
	var models []*model_client.ModelImported
	if err := sqlopt.WithUUIDs(uuids).Apply(c.db).WithContext(ctx).Find(&models).Error; err != nil {
		return nil, toErrStatus("model_get_by_uuids_err", err.Error())
	}
	return models, nil
}

func (c *Client) ListModelsByIds(ctx context.Context, modelIds []uint32) ([]*model_client.ModelImported, *errs.Status) {
	var models []*model_client.ModelImported
	if err := sqlopt.WithIDs(modelIds).Apply(c.db).WithContext(ctx).Find(&models).Error; err != nil {
		return nil, toErrStatus("model_get_by_ids_err", err.Error())
	}
	return models, nil
}

func (c *Client) ListModels(ctx context.Context, tab *model_client.ModelImported) ([]*model_client.ModelImported, *errs.Status) {
	var modelInfos []*model_client.ModelImported
	db := sqlopt.SQLOptions(
		sqlopt.WithUserOrgOrPublicScope(tab.UserID, tab.OrgID),
		sqlopt.WithProvider(tab.Provider),
		sqlopt.WithScopeType(tab.ScopeType),
		sqlopt.WithModelType(tab.ModelType),
		sqlopt.LikeDisplayNameOrModel(tab.DisplayName),
	).Apply(c.db.WithContext(ctx))
	if tab.IsActive {
		db = sqlopt.WithIsActive(true).Apply(db)
	}
	if err := db.Order("updated_at DESC").Find(&modelInfos).Error; err != nil {
		return nil, toErrStatus("model_list_models_err", err.Error())
	}
	return modelInfos, nil
}

// ListModelsInStatisticScope 统计看板：org_id IN + user_id IN，与 app 统计 scope 语义一致。
func (c *Client) ListModelsInStatisticScope(ctx context.Context, orgIds, userIds []string, modelType string) ([]*model_client.ModelImported, *errs.Status) {
	var modelInfos []*model_client.ModelImported
	db := sqlopt.SQLOptions(
		sqlopt.WithUserIDs(userIds),
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithModelType(modelType),
	).Apply(c.db.WithContext(ctx))
	if err := db.Order("updated_at DESC").Find(&modelInfos).Error; err != nil {
		return nil, toErrStatus("model_list_models_in_statistic_scope_err", err.Error())
	}
	return modelInfos, nil
}

// AdminModelPageList 管理员中心模型全局分页列表：跨用户/组织（userIds/orgIds IN），
// 支持按显示名模糊、供应商、模型类型集合过滤，按 updated_at 倒序分页，返回当前页数据与总数。
func (c *Client) AdminModelPageList(ctx context.Context, name, provider string, modelTypes, userIds, orgIds []string, scopeTypes []uint32, pageNum, pageSize int) ([]*model_client.ModelImported, int64, *errs.Status) {
	query := sqlopt.SQLOptions(
		sqlopt.WithUserIDs(userIds),
		sqlopt.WithOrgIDs(orgIds),
		sqlopt.WithProvider(provider),
		sqlopt.LikeDisplayNameOrModel(db.EscapeLike(name)),
		sqlopt.WithModelTypes(modelTypes),
		sqlopt.WithScopeTypes(scopeTypes),
	).Apply(c.db.WithContext(ctx)).Model(&model_client.ModelImported{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("model_admin_page_list_err", err.Error())
	}

	var modelInfos []*model_client.ModelImported
	if err := query.
		Order("updated_at DESC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&modelInfos).Error; err != nil {
		return nil, 0, toErrStatus("model_admin_page_list_err", err.Error())
	}
	return modelInfos, total, nil
}

func (c *Client) ListTypeModels(ctx context.Context, tab *model_client.ModelImported) ([]*model_client.ModelImported, *errs.Status) {
	var modelInfos []*model_client.ModelImported
	modelRerankTypes := []string{"rerank", "multimodal-rerank"}
	modelEmbedTypes := []string{"embedding", "multimodal-embedding"}
	var modelTypeOpt sqlopt.SQLOption
	switch tab.ModelType {
	case "rerank":
		modelTypeOpt = sqlopt.WithModelTypes(modelRerankTypes)
	case "embedding":
		modelTypeOpt = sqlopt.WithModelTypes(modelEmbedTypes)
	default:
		modelTypeOpt = sqlopt.WithModelType(tab.ModelType)
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserOrgOrPublicScope(tab.UserID, tab.OrgID),
		modelTypeOpt,
		sqlopt.WithIsActive(true),
	).Apply(c.db.WithContext(ctx)).Order("provider DESC").Find(&modelInfos).Error; err != nil {
		return nil, toErrStatus("model_list_type_models_err", err.Error())
	}
	return modelInfos, nil
}
