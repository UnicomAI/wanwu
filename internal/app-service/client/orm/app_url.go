package orm

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
)

func (c *Client) CreateAppUrl(ctx context.Context, appUrl *model.AppUrl) *err_code.Status {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appUrl.AppID),
		sqlopt.WithAppType(appUrl.AppType),
		sqlopt.WithName(appUrl.Name),
	).Apply(c.db.WithContext(ctx)).Model(&model.AppUrl{}).Count(&count).Error; err != nil {
		return toErrStatus("app_url_get_by_name", appUrl.Name, err.Error())
	}
	// There is an agent Url with the same name
	if count > 0 {
		return toErrStatus("app_url_same_name", appUrl.Name)
	}
	if err := c.db.WithContext(ctx).Create(appUrl).Error; err != nil {
		return toErrStatus("app_url_create", appUrl.Name, err.Error())
	}
	return nil
}

func (c *Client) DeleteAppUrl(ctx context.Context, urlID uint32) *err_code.Status {
	if err := sqlopt.WithID(urlID).Apply(c.db.WithContext(ctx)).Delete(&model.AppUrl{}).Error; err != nil {
		return toErrStatus("app_url_delete", util.Int2Str(urlID), err.Error())
	}
	return nil

}

func (c *Client) UpdateAppUrl(ctx context.Context, appUrl *model.AppUrl) *err_code.Status {
	var count int64
	appUrlCfg := &model.AppUrl{}
	if err := sqlopt.WithID(appUrl.ID).Apply(c.db.WithContext(ctx)).First(appUrlCfg).Error; err != nil {
		return toErrStatus("app_url_get")
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithAppID(appUrlCfg.AppID),
		sqlopt.WithAppType(appUrlCfg.AppType),
		sqlopt.WithName(appUrl.Name),
	).Apply(c.db.WithContext(ctx)).Where("id != ?", appUrl.ID).Model(&model.AppUrl{}).Count(&count).Error; err != nil {
		return toErrStatus("app_url_get_by_name", appUrl.Name, err.Error())
	}
	// There is an agent Url with the same name
	if count > 0 {
		return toErrStatus("app_url_same_name", appUrl.Name)
	}
	if err := sqlopt.WithID(appUrl.ID).Apply(c.db.WithContext(ctx)).Model(appUrl).Updates(map[string]interface{}{
		"name":                  appUrl.Name,
		"expired_at":            appUrl.ExpiredAt,
		"copyright":             appUrl.Copyright,
		"copyright_enable":      appUrl.CopyrightEnable,
		"privacy_policy":        appUrl.PrivacyPolicy,
		"privacy_policy_enable": appUrl.PrivacyPolicyEnable,
		"disclaimer":            appUrl.Disclaimer,
		"disclaimer_enable":     appUrl.DisclaimerEnable,
		"description":           appUrl.Description,
	}).Error; err != nil {
		return toErrStatus("app_url_update", appUrl.Name, err.Error())
	}
	return nil

}

func (c *Client) GetAppUrlInfoBySuffix(ctx context.Context, suffix string) (*model.AppUrl, *err_code.Status) {
	appUrl := &model.AppUrl{}
	if err := sqlopt.WithSuffix(suffix).Apply(c.db.WithContext(ctx)).First(appUrl).Error; err != nil {
		return nil, toErrStatus("app_url_get")
	}
	return appUrl, nil

}

func (c *Client) GetAppUrlList(ctx context.Context, appID, appType string) ([]*model.AppUrl, *err_code.Status) {
	var appUrls []*model.AppUrl
	if err := sqlopt.SQLOptions(sqlopt.WithAppID(appID), sqlopt.WithAppType(appType)).Apply(c.db.WithContext(ctx)).Find(&appUrls).Error; err != nil {
		return nil, toErrStatus("app_url_get_list", appID, appType, err.Error())
	}
	return appUrls, nil
}

func (c *Client) AppUrlStatusSwitch(ctx context.Context, urlID uint32, status bool) *err_code.Status {
	if err := sqlopt.WithID(urlID).Apply(c.db.WithContext(ctx)).Model(&model.AppUrl{}).Updates(map[string]interface{}{
		"status": status,
	}).Error; err != nil {
		return toErrStatus("app_url_update_switch", util.Int2Str(urlID), err.Error())
	}
	return nil
}
