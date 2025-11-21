package orm

import (
	"context"
	"errors"
	"fmt"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) PublishApp(ctx context.Context, userId, orgId, appId, appType, publishType string) *errs.Status {
	var existingApp model.App
	err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
	).Apply(c.db.WithContext(ctx)).First(&existingApp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newApp := model.App{
				UserID:      userId,
				OrgID:       orgId,
				AppID:       appId,
				AppType:     appType,
				PublishType: publishType,
			}
			if err := c.db.WithContext(ctx).Create(&newApp).Error; err != nil {
				return toErrStatus("app_publish_app_create", appId, err.Error())
			}
			return nil
		}
		return toErrStatus("app_publish_app_query", appId, err.Error())
	}
	return toErrStatus("app_publish_app_exist", appId)
}

func (c *Client) UnPublishApp(ctx context.Context, appId, appType, userId string) *errs.Status {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete other users' ApiKeys when unpublishing
		if err := sqlopt.SQLOptions(
			sqlopt.WithAppID(appId),
			sqlopt.WithAppType(appType),
			sqlopt.WithExcludeUserID(userId),
		).Apply(tx).Delete(&model.ApiKey{}).Error; err != nil {
			return fmt.Errorf("failed to delete api key: %v", err)
		}
		if err := deleteAppRelatedDataByUnPublish(tx, appId, appType); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return toErrStatus("app_unpublish", appId, err.Error())
	}
	return nil

}

func (c *Client) GetAppList(ctx context.Context, userId, orgId, appType string) ([]*model.App, *errs.Status) {
	var publishApps []*model.App
	query := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
		sqlopt.WithAppType(appType),
	).Apply(c.db.WithContext(ctx))
	if err := query.Order("id DESC").Find(&publishApps).Error; err != nil {
		return nil, toErrStatus("app_publish_apps_get", err.Error())
	}
	return publishApps, nil
}

func (c *Client) DeleteApp(ctx context.Context, appId, appType string) *errs.Status {
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := sqlopt.SQLOptions(
			sqlopt.WithAppID(appId),
			sqlopt.WithAppType(appType),
		).Apply(tx).Delete(&model.ApiKey{}).Error; err != nil {
			return fmt.Errorf("failed to delete api key: %v", err)
		}
		if err := deleteAppRelatedDataByUnPublish(tx, appId, appType); err != nil {
			return err
		}
		if err := sqlopt.SQLOptions(
			sqlopt.WithAppID(appId),
			sqlopt.WithAppType(appType),
		).Apply(tx).Delete(&model.AppUrl{}).Error; err != nil {
			return fmt.Errorf("failed to delete app open url: %v", err)
		}
		return nil
	})
	if err != nil {
		return toErrStatus("app_config_delete", appId, err.Error())
	}
	return nil
}

func (c *Client) GetAppListByIds(ctx context.Context, ids []string) ([]*model.App, *errs.Status) {
	if len(ids) == 0 {
		return nil, nil
	}
	var publishApps []*model.App
	if err := sqlopt.InAppIds(ids).Apply(c.db.WithContext(ctx)).Order("id DESC").Find(&publishApps).Error; err != nil {
		return nil, toErrStatus("app_publish_apps_get_by_ids", err.Error())
	}
	return publishApps, nil
}

func (c *Client) RecordAppHistory(ctx context.Context, userId, appId, appType string) *errs.Status {
	var app model.App
	err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
	).Apply(c.db.WithContext(ctx)).First(&app).Error
	if err != nil {
		return toErrStatus("app_publish_status_query", appId, err.Error())
	}
	var appRecord model.AppHistory
	err = sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
	).Apply(c.db.WithContext(ctx)).First(&appRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newAppRecord := model.AppHistory{
				UserID:  userId,
				AppID:   appId,
				AppType: appType,
			}
			if err := c.db.WithContext(ctx).Create(&newAppRecord).Error; err != nil {
				return toErrStatus("app_record_history_create", appId, err.Error())
			}
			return nil
		}
		return toErrStatus("app_record_history_query", appId, err.Error())
	}
	if err = c.db.WithContext(ctx).Model(&appRecord).Update("updated_at", time.Now().UnixMilli()).Error; err != nil {
		return toErrStatus("app_record_history_update", appId, err.Error())
	}
	return nil
}

// Does not include apiKey
func deleteAppRelatedDataByUnPublish(tx *gorm.DB, appId, appType string) error {
	if err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
	).Apply(tx).Delete(&model.App{}).Error; err != nil {
		return fmt.Errorf("failed to delete app: %v", err)
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
	).Apply(tx).Delete(&model.AppHistory{}).Error; err != nil {
		return fmt.Errorf("failed to delete app history: %v", err)
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appId),
		sqlopt.WithAppType(appType),
	).Apply(tx).Delete(&model.AppFavorite{}).Error; err != nil {
		return fmt.Errorf("failed to delete app favorite: %v", err)
	}
	return nil
}
