package orm

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/UnicomAI/wanwu/pkg/util"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"

	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"

	"gorm.io/gorm"
)

type Client struct {
	db *gorm.DB
}

func NewClient(ctx context.Context, db *gorm.DB) (*Client, error) {
	// auto migrate
	if err := db.AutoMigrate(
		model.MCPClient{},
		model.CustomTool{},
		model.MCPServer{},
		model.MCPServerTool{},
		model.BuiltinTool{},
	); err != nil {
		return nil, err
	}
	// 迁移数据 [EN] Migrate data
	if err := initCustomToolAuthJson(db); err != nil {
		return nil, err
	}
	return &Client{
		db: db,
	}, nil
}

func initCustomToolAuthJson(dbClient *gorm.DB) error {
	var customToolBaseList []model.CustomTool
	//数据量不会太大直接getAll [EN] The amount of data will not be too large, just getAll
	err := dbClient.Model(&model.CustomTool{}).
		Where("tool_square_id = '' OR tool_square_id IS NULL").
		Where("auth_json = '' OR auth_json IS NULL").
		Find(&customToolBaseList).Error
	if err != nil {
		return err
	}

	for _, customTool := range customToolBaseList {
		if len(customTool.ToolSquareId) > 0 || customTool.AuthJSON != "" {
			continue
		}
		apiAuth := &util.ApiAuthWebRequest{
			AuthType: util.AuthTypeNone,
		}
		if customTool.Type == "API Key" {
			apiAuth.AuthType = util.AuthTypeAPIKeyHeader
			apiAuth.ApiKeyHeaderPrefix = util.ApiKeyHeaderPrefixBearer
			apiAuth.ApiKeyHeader = util.ApiKeyHeaderDefault
			apiAuth.ApiKeyValue = customTool.APIKey
		}
		apiAuthBytes, err := json.Marshal(apiAuth)
		if err != nil {
			return err
		}
		updateMap := map[string]interface{}{
			"auth_json": string(apiAuthBytes),
		}
		err = dbClient.Model(&model.CustomTool{}).Where("id = ?", customTool.ID).Updates(updateMap).Error
		if err != nil {
			return err
		}
	}

	// 清理脏数据 [EN] Clean dirty data
	err = dbClient.Model(&model.CustomTool{}).
		Where("tool_square_id != ''").Delete(&model.CustomTool{}).Error
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) transaction(ctx context.Context, fc func(tx *gorm.DB) *err_code.Status) *err_code.Status {
	var status *err_code.Status
	_ = c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if status = fc(tx); status != nil {
			return errors.New(status.String())
		}
		return nil
	})
	return status
}

func toErrStatus(key string, args ...string) *err_code.Status {
	return &err_code.Status{
		TextKey: key,
		Args:    args,
	}
}
