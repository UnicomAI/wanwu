package orm

import (
	"context"
	"database/sql"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_client "github.com/UnicomAI/wanwu/internal/model-service/client/model"
	"github.com/UnicomAI/wanwu/internal/model-service/client/orm/sqlopt"
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
	// Query whether the same model already exists (dropped in version 0.2.1)
	//if err := sqlopt.SQLOptions(
	//	sqlopt.WithProvider(tab.Provider),
	//	sqlopt.WithModelType(tab.ModelType),
	//	sqlopt.WithModel(tab.Model),
	//	sqlopt.WithOrgID(tab.OrgID),
	//	sqlopt.WithUserID(tab.UserID),
	//).Apply(db).Select("id").First(&model_client.ModelImported{}).Error; err == nil {
	//	return toErrStatus("model_create_err", "model with same identifier exist")
	//} else if err != gorm.ErrRecordNotFound {
	//	// other errors
	//	return toErrStatus("model_create_err", err.Error())
	//}

	if tab.DisplayName != "" {
		if err := sqlopt.SQLOptions(
			sqlopt.WithProvider(tab.Provider),
			sqlopt.WithModelType(tab.ModelType),
			sqlopt.WithDisplayName(tab.DisplayName),
			sqlopt.WithOrgID(tab.OrgID),
			sqlopt.WithUserID(tab.UserID),
		).Apply(db).Select("id").First(&model_client.ModelImported{}).Error; err == nil {
			return toErrStatus("model_create_err", "model with same display name exist")
		} else if err != gorm.ErrRecordNotFound {
			// Other errors
			return toErrStatus("model_create_err", err.Error())
		}
	}

	if err = db.Create(tab).Error; err != nil {
		return toErrStatus("model_create_err", err.Error())
	}
	return nil
}

func (c *Client) DeleteModel(ctx context.Context, tab *model_client.ModelImported) *errs.Status {
	// Query
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
	// Query
	var existing model_client.ModelImported
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tab.ID),
	).Apply(c.db).WithContext(ctx).First(&existing).Error; err != nil {
		return toErrStatus("model_update_err", err.Error())
	}
	if existing.ModelType != tab.ModelType || existing.Model != tab.Model || existing.Provider != tab.Provider {
		return toErrStatus("model_update_err", "type,model,provider can not update!")
	}
	// renew
	if err := c.db.WithContext(ctx).Model(existing).Updates(map[string]interface{}{
		"display_name":    tab.DisplayName,
		"model_icon_path": tab.ModelIconPath,
		"publish_date":    tab.PublishDate,
		"provider_config": tab.ProviderConfig,
	}).Error; err != nil {
		return toErrStatus("model_update_err", err.Error())
	}
	return nil
}

func (c *Client) ChangeModelStatus(ctx context.Context, tab *model_client.ModelImported) *errs.Status {
	// Query
	var existing model_client.ModelImported
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tab.ID),
		sqlopt.WithOrgID(tab.OrgID),
		sqlopt.WithUserID(tab.UserID),
	).Apply(c.db).WithContext(ctx).Select("id").First(&existing).Error; err != nil {
		return toErrStatus("model_change_model_status_err", err.Error())
	}
	// renew
	if err := c.db.WithContext(ctx).Model(existing).Updates(map[string]interface{}{
		"is_active": tab.IsActive,
	}).Error; err != nil {
		return toErrStatus("model_change_model_status_err", err.Error())
	}
	return nil
}

func (c *Client) GetModelById(ctx context.Context, modelId uint32) (*model_client.ModelImported, *errs.Status) {
	info := &model_client.ModelImported{}
	if err := sqlopt.WithID(modelId).Apply(c.db).WithContext(ctx).First(info).Error; err != nil {
		return nil, toErrStatus("model_get_by_id_err", err.Error())
	}
	return info, nil
}

func (c *Client) GetModelByIds(ctx context.Context, modelIds []uint32) ([]*model_client.ModelImported, *errs.Status) {
	var models []*model_client.ModelImported
	if err := sqlopt.WithIDs(modelIds).Apply(c.db).WithContext(ctx).Find(&models).Error; err != nil {
		return nil, toErrStatus("model_get_by_ids_err", err.Error())
	}
	return models, nil
}

func (c *Client) GetModel(ctx context.Context, tab *model_client.ModelImported) (*model_client.ModelImported, *errs.Status) {
	info := &model_client.ModelImported{}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(tab.ID),
		sqlopt.WithOrgID(tab.OrgID),
		sqlopt.WithUserID(tab.UserID),
	).Apply(c.db).WithContext(ctx).First(info).Error; err != nil {
		return nil, toErrStatus("model_get_err", err.Error())
	}
	return info, nil
}

func (c *Client) ListModels(ctx context.Context, tab *model_client.ModelImported) ([]*model_client.ModelImported, *errs.Status) {
	var modelInfos []*model_client.ModelImported
	db := sqlopt.SQLOptions(
		sqlopt.WithOrgID(tab.OrgID),
		sqlopt.WithUserID(tab.UserID),
		sqlopt.WithProvider(tab.Provider),
		sqlopt.WithModelType(tab.ModelType),
		sqlopt.LikeDisplayNameOrModel(tab.DisplayName),
	).Apply(c.db.WithContext(ctx))
	if tab.IsActive {
		db = sqlopt.WithIsActive(true).Apply(db)
	}
	if err := db.Order("created_at DESC").Find(&modelInfos).Error; err != nil {
		return nil, toErrStatus("model_list_models_err", err.Error())
	}
	return modelInfos, nil
}

func (c *Client) ListTypeModels(ctx context.Context, tab *model_client.ModelImported) ([]*model_client.ModelImported, *errs.Status) {
	var modelInfos []*model_client.ModelImported
	if err := sqlopt.SQLOptions(
		sqlopt.WithOrgID(tab.OrgID),
		sqlopt.WithUserID(tab.UserID),
		sqlopt.WithModelType(tab.ModelType),
		sqlopt.WithIsActive(true),
	).Apply(c.db.WithContext(ctx)).Order("provider DESC").Find(&modelInfos).Error; err != nil {
		return nil, toErrStatus("model_list_type_models_err", err.Error())
	}
	return modelInfos, nil
}
