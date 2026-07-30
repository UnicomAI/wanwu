package orm

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
)

func (c *Client) GetAppKeyList(ctx context.Context, userId, orgId, appId, appType string) ([]*model.ApiKey, *errs.Status) {
	var appKeys []*model.ApiKey
	if err := sqlopt.SQLOptions(
		sqlopt.WithAppType(appType),
		sqlopt.WithAppID(appId),
		sqlopt.WithOrgID(orgId),
		sqlopt.WithUserID(userId),
	).Apply(c.db.WithContext(ctx)).
		Find(&appKeys).Error; err != nil {
		return nil, toErrStatus("app_api_keys_get", err.Error())
	}
	return appKeys, nil
}

func (c *Client) DelAppKey(ctx context.Context, appKeyId uint32, userId, orgId string) *errs.Status {
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(appKeyId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Delete(&model.ApiKey{}).Error; err != nil {
		return toErrStatus("app_api_key_delete", util.Int2Str(appKeyId), err.Error())
	}
	return nil
}

func (c *Client) GenAppKey(ctx context.Context, userId, orgId, appId, appType, appKey string) (*model.ApiKey, *errs.Status) {
	newAppKey := &model.ApiKey{
		OrgID:   orgId,
		UserID:  userId,
		AppID:   appId,
		AppType: appType,
		ApiKey:  appKey,
	}
	if err := c.db.WithContext(ctx).Create(newAppKey).Error; err != nil {
		return nil, toErrStatus("app_api_keys_gen", err.Error())
	}
	return newAppKey, nil
}

func (c *Client) GetAppKeyByKey(ctx context.Context, appKey string) (*model.ApiKey, *errs.Status) {
	ret := &model.ApiKey{}
	if err := sqlopt.WithAppKey(appKey).Apply(c.db).WithContext(ctx).First(ret).Error; err != nil {
		return nil, toErrStatus("app_api_key_get_by_key", err.Error())
	}
	return ret, nil
}
