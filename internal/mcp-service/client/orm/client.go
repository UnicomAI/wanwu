package orm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/UnicomAI/wanwu/pkg/util"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"

	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"

	"gorm.io/gorm"
)

const (
	// skill 发布改造后 acquired_skill 仅保留 custom_skill_id，清理无关联键的历史数据
	initLegacyAcquiredSkillFlagKey = "v0.5.5_acquired_skill_legacy_cleared"
	// skill 统计字段初始化：根据 acquired_skill 历史 COUNT 填充 custom_skill.add_count
	initCustomSkillCountsFlagKey = "v0.5.6_custom_skill_counts_initialized"
)

type Metadata struct {
	MetaKey   string `gorm:"primaryKey;column:key"`
	MetaValue string `gorm:"column:value"`
	CreatedAt int64  `gorm:"autoCreateTime:milli"`
	UpdatedAt int64  `gorm:"autoUpdateTime:milli"`
}

type Client struct {
	db *gorm.DB
}

func NewClient(ctx context.Context, db *gorm.DB) (*Client, error) {
	if err := db.AutoMigrate(&Metadata{}); err != nil {
		return nil, err
	}

	// auto migrate
	if err := db.AutoMigrate(
		model.MCPClient{},
		model.CustomTool{},
		model.MCPServer{},
		model.MCPServerTool{},
		model.BuiltinTool{},
		model.CustomSkill{},
		model.AcquiredSkill{},
		model.CustomSkillVariable{},
		model.AcquiredSkillVariable{},
		model.BuiltinSkillVariable{},
		model.CustomSkillPublish{},
		model.BuiltinSkill{},
	); err != nil {
		return nil, err
	}
	// 迁移数据
	if err := initCustomToolAuthJson(db); err != nil {
		return nil, err
	}
	if err := initMCPAuthJson(db); err != nil {
		return nil, err
	}
	if err := initMCPClientTransport(db); err != nil {
		return nil, err
	}
	if err := initLegacyAcquiredSkillCleanup(db); err != nil {
		return nil, err
	}
	if err := initCustomSkillCounts(db); err != nil {
		return nil, err
	}
	return &Client{
		db: db,
	}, nil
}

func initLegacyAcquiredSkillCleanup(db *gorm.DB) error {
	var meta Metadata
	err := db.Where(&Metadata{MetaKey: initLegacyAcquiredSkillFlagKey}).First(&meta).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("query metadata failed: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("custom_skill_id = '' OR custom_skill_id IS NULL").
			Delete(&model.AcquiredSkill{}).Error; err != nil {
			return fmt.Errorf("delete legacy acquired_skill failed: %w", err)
		}
		if err := tx.Create(&Metadata{MetaKey: initLegacyAcquiredSkillFlagKey}).Error; err != nil {
			return fmt.Errorf("failed to set init flag: %w", err)
		}
		return nil
	})
}

func initCustomToolAuthJson(dbClient *gorm.DB) error {
	var customToolBaseList []model.CustomTool
	//数据量不会太大直接getAll
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

	// 清理脏数据
	err = dbClient.Model(&model.CustomTool{}).
		Where("tool_square_id != ''").Delete(&model.CustomTool{}).Error
	if err != nil {
		return err
	}

	return nil
}

func initMCPAuthJson(dbClient *gorm.DB) error {
	//数据量不会太大直接getAll
	apiAuth := &util.ApiAuthWebRequest{
		AuthType: util.AuthTypeNone,
	}
	apiAuthBytes, err := json.Marshal(apiAuth)
	if err != nil {
		return err
	}
	updateMap := map[string]interface{}{
		"auth_json": string(apiAuthBytes),
	}
	err = dbClient.Model(&model.MCPClient{}).
		Where("auth_json = '' OR auth_json IS NULL").
		Updates(updateMap).Error
	if err != nil {
		return err
	}
	return nil
}

func initMCPClientTransport(dbClient *gorm.DB) error {
	err := dbClient.Model(&model.MCPClient{}).
		Where("transport = '' OR transport IS NULL").
		Where("sse_url != ''").Update("transport", "sse").Error
	if err != nil {
		return err
	}
	return nil
}

// initCustomSkillCounts 根据 acquired_skill 表数据填充
func initCustomSkillCounts(db *gorm.DB) error {
	var meta Metadata
	if err := db.Where(&Metadata{MetaKey: initCustomSkillCountsFlagKey}).First(&meta).Error; err == nil {
		return nil
	}

	// 按 custom_skill_id 分组统计 acquired 数量
	type countResult struct {
		CustomSkillID string
		Cnt           int32
	}
	var results []countResult
	if err := db.Model(&model.AcquiredSkill{}).
		Select("custom_skill_id, COUNT(*) as cnt").
		Where("custom_skill_id != '' AND custom_skill_id IS NOT NULL").
		Group("custom_skill_id").
		Find(&results).Error; err != nil {
		return fmt.Errorf("query acquired_skill count failed: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, r := range results {
			if err := tx.Model(&model.CustomSkill{}).
				Where("id = ?", r.CustomSkillID).
				UpdateColumn("acquired_count", r.Cnt).Error; err != nil {
				return fmt.Errorf("update custom_skill add_count failed: %w", err)
			}
		}
		if err := tx.Create(&Metadata{MetaKey: initCustomSkillCountsFlagKey}).Error; err != nil {
			return fmt.Errorf("failed to set init flag: %w", err)
		}
		return nil
	})
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
