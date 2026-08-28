package orm

import (
	"context"
	"errors"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

func (c *Client) CreateAcquiredSkill(ctx context.Context, acquiredSkill *model.AcquiredSkill) (string, *err_code.Status) {
	if acquiredSkill.CustomSkillID == "" {
		return "", toErrStatus("mcp_acquired_skill_create", "customSkillId is empty")
	}
	if _, err := c.GetCustomSkill(ctx, acquiredSkill.CustomSkillID); err != nil {
		return "", err
	}
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(acquiredSkill.UserID),
		sqlopt.WithOrgID(acquiredSkill.OrgID),
		sqlopt.WithCustomSkillID(acquiredSkill.CustomSkillID),
	).Apply(c.db).WithContext(ctx).Model(&model.AcquiredSkill{}).
		Count(&count).Error; err != nil {
		return "", toErrStatus("mcp_acquired_skill_create", err.Error())
	}
	if count > 0 {
		return "", toErrStatus("mcp_acquired_skill_already_exists")
	}
	if err := c.db.WithContext(ctx).Create(acquiredSkill).Error; err != nil {
		return "", toErrStatus("mcp_acquired_skill_create", err.Error())
	}
	return util.Int2Str(acquiredSkill.ID), nil
}

func (c *Client) DeleteAcquiredSkill(ctx context.Context, acquiredSkillId, userId, orgId string) *err_code.Status {
	id := util.MustU32(acquiredSkillId)
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		if err := sqlopt.SQLOptions(
			sqlopt.WithAcquiredSkillID(acquiredSkillId),
			sqlopt.WithUserID(userId),
			sqlopt.WithOrgID(orgId),
		).Apply(tx).Delete(&model.AcquiredSkillVariable{}).Error; err != nil {
			return toErrStatus("mcp_acquired_skill_delete_variables", err.Error())
		}
		if err := sqlopt.SQLOptions(
			sqlopt.WithID(id),
			sqlopt.WithUserID(userId),
			sqlopt.WithOrgID(orgId),
		).Apply(tx).Delete(&model.AcquiredSkill{}).Error; err != nil {
			return toErrStatus("mcp_acquired_skill_delete", err.Error())
		}
		return nil
	})
}

func (c *Client) GetAcquiredSkill(ctx context.Context, acquiredSkillId string) (*model.AcquiredSkill, *err_code.Status) {
	var as model.AcquiredSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(util.MustU32(acquiredSkillId)),
	).Apply(c.db).WithContext(ctx).First(&as).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toErrStatus("mcp_acquired_skill_not_found", acquiredSkillId)
		}
		return nil, toErrStatus("mcp_acquired_skill_get", acquiredSkillId, err.Error())
	}
	return &as, nil
}

func (c *Client) GetAcquiredSkillByIDList(ctx context.Context, acquiredSkillIdList []string) ([]*model.AcquiredSkill, *err_code.Status) {
	if len(acquiredSkillIdList) == 0 {
		return []*model.AcquiredSkill{}, nil
	}
	ids := make([]uint32, 0, len(acquiredSkillIdList))
	for _, id := range acquiredSkillIdList {
		if id == "" {
			continue
		}
		ids = append(ids, util.MustU32(id))
	}
	if len(ids) == 0 {
		return []*model.AcquiredSkill{}, nil
	}
	var list []*model.AcquiredSkill
	if err := sqlopt.WithIDs(ids).Apply(c.db).WithContext(ctx).Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_acquired_skill_get_by_id_list", err.Error())
	}
	return list, nil
}

func (c *Client) GetAcquiredSkillList(ctx context.Context, userId, orgId, name string) ([]*model.AcquiredSkill, int64, *err_code.Status) {
	var list []*model.AcquiredSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, toErrStatus("mcp_acquired_skill_list", err.Error())
	}
	if name == "" || len(list) == 0 {
		return list, int64(len(list)), nil
	}
	customSkillIDs := make([]string, 0, len(list))
	for _, as := range list {
		if as.CustomSkillID != "" {
			customSkillIDs = append(customSkillIDs, as.CustomSkillID)
		}
	}
	matched, st := c.GetCustomSkillListByIDs(ctx, customSkillIDs, name)
	if st != nil {
		return nil, 0, st
	}
	if len(matched) == 0 {
		return []*model.AcquiredSkill{}, 0, nil
	}
	matchedIDs := make(map[string]bool, len(matched))
	for _, cs := range matched {
		matchedIDs[util.Int2Str(cs.ID)] = true
	}
	filtered := make([]*model.AcquiredSkill, 0, len(matched))
	for _, as := range list {
		if matchedIDs[as.CustomSkillID] {
			filtered = append(filtered, as)
		}
	}
	return filtered, int64(len(filtered)), nil
}
