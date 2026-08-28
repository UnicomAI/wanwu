package orm

import (
	"context"
	"errors"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) GetExplorationAppList(ctx context.Context, userId, orgId, name, appType, searchType string) ([]*ExplorationAppInfo, *errs.Status) {
	var apps []*model.App
	var ret []*ExplorationAppInfo
	var favoriteApps []*model.AppFavorite
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithAppType(appType),
	).Apply(c.db.WithContext(ctx)).Find(&favoriteApps).Error; err != nil {
		return nil, toErrStatus("app_explore_favorite_apps_get", err.Error())
	}
	switch searchType {
	case "", "all", "private":
		query := sqlopt.SQLOptions(
			sqlopt.WithAppType(appType),
			sqlopt.WithSearchType(userId, orgId, searchType),
		).Apply(c.db.WithContext(ctx))
		if err := query.Order("id DESC").Find(&apps).Error; err != nil {
			return nil, toErrStatus("app_explore_apps_get", err.Error())
		}
		if len(apps) == 0 {
			break
		}
		for _, app := range apps {
			appInfo := &ExplorationAppInfo{
				UserID:      app.UserID,
				OrgID:       app.OrgID,
				AppId:       app.AppID,
				AppType:     app.AppType,
				CreatedAt:   app.CreatedAt,
				UpdatedAt:   app.UpdatedAt,
				IsFavorite:  false,
				PublishType: app.PublishType,
			}
			for _, favoriteApp := range favoriteApps {
				if favoriteApp.AppID == app.AppID && favoriteApp.AppType == app.AppType {
					appInfo.IsFavorite = true
					break
				}
			}
			ret = append(ret, appInfo)
		}
	case "favorite":
		if len(favoriteApps) == 0 {
			break
		}
		var appIds []string
		for _, app := range favoriteApps {
			appIds = append(appIds, app.AppID)
		}
		apps, err := c.getAppByIds(ctx, appIds)
		if err != nil {
			return nil, err
		}
		for _, app := range favoriteApps {
			var isValid bool
			appInfo := &ExplorationAppInfo{
				UserID:      "",
				AppId:       app.AppID,
				AppType:     app.AppType,
				CreatedAt:   app.CreatedAt,
				UpdatedAt:   app.UpdatedAt,
				IsFavorite:  true,
				PublishType: "",
			}
			for _, info := range apps {
				if app.AppID == info.AppID && app.AppType == info.AppType {
					appInfo.UserID = info.UserID
				}
				if app.AppID != info.AppID || app.AppType != info.AppType || !canAccessApp(info, userId, orgId) {
					continue
				}
				isValid = true
			}
			if !isValid {
				continue
			}
			ret = append(ret, appInfo)
		}
	case "history":
		var historyApps []*model.AppHistory
		oneMonthAgo := time.Now().AddDate(0, -1, 0).UnixMilli()
		if err := sqlopt.SQLOptions(
			sqlopt.WithUserID(userId),
			sqlopt.WithAppType(appType),
			sqlopt.StartUpdatedAt(oneMonthAgo),
		).Apply(c.db.WithContext(ctx)).
			Order("updated_at DESC").Find(&historyApps).Error; err != nil {
			return nil, toErrStatus("app_history_apps_get", err.Error())
		}
		if len(historyApps) == 0 {
			break
		}
		var appIds []string
		for _, r := range historyApps {
			appIds = append(appIds, r.AppID)
		}
		apps, err := c.getAppByIds(ctx, appIds)
		if err != nil {
			return nil, err
		}
		for _, historyApp := range historyApps {
			var isValid bool
			appInfo := &ExplorationAppInfo{
				UserID:     "",
				AppId:      historyApp.AppID,
				AppType:    historyApp.AppType,
				CreatedAt:  historyApp.CreatedAt,
				UpdatedAt:  historyApp.UpdatedAt,
				IsFavorite: false,
			}
			for _, favoriteApp := range favoriteApps {
				if favoriteApp.AppID == historyApp.AppID && favoriteApp.AppType == historyApp.AppType {
					appInfo.IsFavorite = true
					break
				}
			}
			for _, info := range apps {
				if historyApp.AppID == info.AppID && historyApp.AppType == info.AppType {
					appInfo.UserID = info.UserID
				}
				if historyApp.AppID != info.AppID || historyApp.AppType != info.AppType || !canAccessApp(info, userId, orgId) {
					continue
				}
				isValid = true
			}
			if !isValid {
				continue
			}
			ret = append(ret, appInfo)
		}
	}
	return ret, nil
}

func (c *Client) ChangeExplorationAppFavorite(ctx context.Context, userId, orgId, appId, appType string, isFavorite bool) *errs.Status {
	var existingApp model.AppFavorite
	if isFavorite {
		err := sqlopt.SQLOptions(
			sqlopt.WithUserID(userId),
			sqlopt.WithAppID(appId),
			sqlopt.WithAppType(appType),
		).Apply(c.db.WithContext(ctx)).First(&existingApp).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				newApp := model.AppFavorite{
					UserID:  userId,
					AppID:   appId,
					AppType: appType,
				}
				if err := c.db.WithContext(ctx).Create(&newApp).Error; err != nil {
					return toErrStatus("app_favorite_app_create", appId, err.Error())
				}
				return nil
			}
			return toErrStatus("app_favorite_app_query", appId, err.Error())
		}
		return nil
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appId),
		sqlopt.WithUserID(userId),
		sqlopt.WithAppType(appType),
	).Apply(c.db.WithContext(ctx)).Delete(&existingApp).Error; err != nil {
		return toErrStatus("app_favorite_app_delete", appId, err.Error())
	}
	return nil
}

func (c *Client) getAppByIds(ctx context.Context, appIds []string) ([]*model.App, *errs.Status) {
	var apps []*model.App
	if err := sqlopt.WithAppIDs(appIds).Apply(c.db.WithContext(ctx)).Find(&apps).Error; err != nil {
		return nil, toErrStatus("app_explore_apps_get", err.Error())
	}
	return apps, nil
}
