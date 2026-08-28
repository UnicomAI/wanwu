package orm

import (
	"context"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IncrementCustomSkillDownloadCount 递增 custom_skill 的 download_count
func (c *Client) IncrementCustomSkillDownloadCount(ctx context.Context, skillId string) *err_code.Status {
	if err := sqlopt.WithID(util.MustU32(skillId)).Apply(c.db.WithContext(ctx)).Model(&model.CustomSkill{}).
		UpdateColumn("download_count", gorm.Expr("download_count + 1")).Error; err != nil {
		return toErrStatus("mcp_custom_skill_increment_download_count", err.Error())
	}
	return nil
}

// IncrementCustomSkillAcquiredCount 递增 custom_skill 的 acquired_count
func (c *Client) IncrementCustomSkillAcquiredCount(ctx context.Context, skillId string) *err_code.Status {
	if err := sqlopt.WithID(util.MustU32(skillId)).Apply(c.db.WithContext(ctx)).Model(&model.CustomSkill{}).
		UpdateColumn("acquired_count", gorm.Expr("acquired_count + 1")).Error; err != nil {
		return toErrStatus("mcp_custom_skill_increment_acquired_count", err.Error())
	}
	return nil
}

// IncrementBuiltinSkillDownloadCount 递增内置skill下载计数，不存在则创建
func (c *Client) IncrementBuiltinSkillDownloadCount(ctx context.Context, skillId string) *err_code.Status {
	stat := model.BuiltinSkill{
		SkillId:       skillId,
		DownloadCount: 1,
	}
	if err := sqlopt.WithID(util.MustU32(skillId)).Apply(c.db.WithContext(ctx)).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "skill_id"}},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"download_count": gorm.Expr("download_count + 1"),
		}),
	}).Create(&stat).Error; err != nil {
		return toErrStatus("mcp_builtin_skill_increment_download_count", err.Error())
	}
	return nil
}

// GetBuiltinSkillDownloadCounts 批量获取内置skill下载计数
func (c *Client) GetBuiltinSkillDownloadCounts(ctx context.Context, skillIds []string) (map[string]int32, *err_code.Status) {
	if len(skillIds) == 0 {
		return make(map[string]int32), nil
	}
	var stats []model.BuiltinSkill
	if err := c.db.WithContext(ctx).Where("skill_id IN ?", skillIds).Find(&stats).Error; err != nil {
		return nil, toErrStatus("mcp_builtin_skill_get_download_counts", err.Error())
	}
	result := make(map[string]int32, len(stats))
	for _, s := range stats {
		result[s.SkillId] = s.DownloadCount
	}
	return result, nil
}
