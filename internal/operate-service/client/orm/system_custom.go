package orm

import (
	"context"
	"encoding/json"
	"errors"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (c *Client) CreateSystemCustom(ctx context.Context, orgID, userID string, key SystemCustomKey, mode SystemCustomMode, custom SystemCustom) *err_code.Status {
	var sysCustom model.SystemCustom
	if err := sqlopt.WithKey(key2WithModeKey(key, mode)).Apply(c.db.WithContext(ctx)).First(&sysCustom).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("ope_system_custom_query", string(key), err.Error())
		}
		if err = c.db.WithContext(ctx).Create(&model.SystemCustom{
			OrgID:  orgID,
			UserID: userID,
			Key:    key2WithModeKey(key, mode),
			Value:  db.LongText(mergeCustomFields(key, model.SystemCustom{}, custom)),
		}).Error; err != nil {
			return toErrStatus("ope_system_custom_create", string(key), err.Error())
		}
		return nil
	}

	if err := sqlopt.WithKey(key2WithModeKey(key, mode)).Apply(c.db.WithContext(ctx)).Model(&model.SystemCustom{}).
		Updates(map[string]interface{}{
			"value": mergeCustomFields(key, sysCustom, custom),
		}).Error; err != nil {
		return toErrStatus("ope_system_custom_update", string(key), err.Error())
	}
	return nil
}

func (c *Client) GetSystemCustom(ctx context.Context, mode SystemCustomMode) (*SystemCustom, *err_code.Status) {
	keys := []string{
		key2WithModeKey(SystemCustomTabKey, mode),
		key2WithModeKey(SystemCustomLoginKey, mode),
		key2WithModeKey(SystemCustomHomeKey, mode),
	}

	var records []model.SystemCustom
	expr := clause.Expr{
		SQL:  "? IN (?)",
		Vars: []interface{}{clause.Column{Name: "key"}, keys},
	}
	if err := c.db.WithContext(ctx).
		Where(expr).
		Find(&records).Error; err != nil {
		return nil, toErrStatus("ope_system_custom_get", err.Error())
	}

	result := &SystemCustom{}
	for _, record := range records {
		switch record.Key {
		case key2WithModeKey(SystemCustomTabKey, mode):
			if err := json.Unmarshal([]byte(record.Value), &result.Tab); err != nil {
				log.Errorf("failed to unmarshal TabConfig mode %s: %v", mode, err)
			}
		case key2WithModeKey(SystemCustomLoginKey, mode):
			if err := json.Unmarshal([]byte(record.Value), &result.Login); err != nil {
				log.Errorf("failed to unmarshal LoginConfig mode %s: %v", mode, err)
			}
		case key2WithModeKey(SystemCustomHomeKey, mode):
			if err := json.Unmarshal([]byte(record.Value), &result.Home); err != nil {
				log.Errorf("failed to unmarshal HomeConfig mode %s: %v", mode, err)
			}
		}
	}
	return result, nil
}

func mergeCustomFields(key SystemCustomKey, custom model.SystemCustom, newCustom SystemCustom) string {
	switch key {
	case SystemCustomTabKey:
		var ret TabConfig
		if err := json.Unmarshal([]byte(custom.Value), &ret); err != nil {
			log.Errorf("failed to unmarshal key %s: %v", key, err)
		}
		if newCustom.Tab.LogoPath != "" {
			ret.LogoPath = newCustom.Tab.LogoPath
		}
		if newCustom.Tab.Title != "" {
			ret.Title = newCustom.Tab.Title
		}
		value, _ := json.Marshal(ret)
		return string(value)

	case SystemCustomLoginKey:
		var ret LoginConfig
		if err := json.Unmarshal([]byte(custom.Value), &ret); err != nil {
			log.Errorf("failed to unmarshal key %s: %v", key, err)
		}
		if newCustom.Login.LoginBgPath != "" {
			ret.LoginBgPath = newCustom.Login.LoginBgPath
		}
		if newCustom.Login.LogoPath != "" {
			ret.LogoPath = newCustom.Login.LogoPath
		}
		if newCustom.Login.ButtonColor != "" {
			ret.ButtonColor = newCustom.Login.ButtonColor
		}
		if newCustom.Login.WelcomeText != "" {
			ret.WelcomeText = newCustom.Login.WelcomeText
		}
		value, _ := json.Marshal(ret)
		return string(value)

	case SystemCustomHomeKey:
		var ret HomeConfig
		if err := json.Unmarshal([]byte(custom.Value), &ret); err != nil {
			log.Errorf("failed to unmarshal key %s: %v", key, err)
		}
		if newCustom.Home.Name != "" {
			ret.Name = newCustom.Home.Name
		}
		if newCustom.Home.LogoPath != "" {
			ret.LogoPath = newCustom.Home.LogoPath
		}
		if newCustom.Home.BgColor != "" {
			ret.BgColor = newCustom.Home.BgColor
		}
		value, _ := json.Marshal(ret)
		return string(value)

	default:
		log.Errorf("unsupported key %s", key)
	}
	return ""
}

func key2WithModeKey(key SystemCustomKey, mode SystemCustomMode) string {
	return string(key) + "_" + string(mode)
}
