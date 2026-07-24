package orm

import (
	"context"
	"errors"

	"github.com/UnicomAI/wanwu/pkg/db"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

func (c *Client) CreateCustomSkill(ctx context.Context, customSkill *model.CustomSkill) (string, *err_code.Status) {
	if err := c.db.WithContext(ctx).Create(customSkill).Error; err != nil {
		return "", toErrStatus("mcp_custom_skill_create", err.Error())
	}
	return util.Int2Str(customSkill.ID), nil
}

func (c *Client) DeleteCustomSkill(ctx context.Context, skillId string) *err_code.Status {
	id := util.MustU32(skillId)
	return c.transaction(ctx, func(tx *gorm.DB) *err_code.Status {
		var acquiredList []*model.AcquiredSkill
		if err := sqlopt.WithCustomSkillID(skillId).Apply(tx).Find(&acquiredList).Error; err != nil {
			return toErrStatus("mcp_custom_skill_delete_acquired_list", err.Error())
		}
		if len(acquiredList) > 0 {
			acquiredSkillIDs := make([]string, 0, len(acquiredList))
			for _, as := range acquiredList {
				acquiredSkillIDs = append(acquiredSkillIDs, util.Int2Str(as.ID))
			}
			if err := sqlopt.WithAcquiredSkillIDs(acquiredSkillIDs).Apply(tx).
				Delete(&model.AcquiredSkillVariable{}).Error; err != nil {
				return toErrStatus("mcp_custom_skill_delete_acquired_variables", err.Error())
			}
			if err := sqlopt.WithCustomSkillID(skillId).Apply(tx).Delete(&model.AcquiredSkill{}).Error; err != nil {
				return toErrStatus("mcp_custom_skill_delete_acquired", err.Error())
			}
		}
		if err := sqlopt.WithSkillID(skillId).Apply(tx).Delete(&model.CustomSkillVariable{}).Error; err != nil {
			return toErrStatus("mcp_custom_skill_delete_variables", err.Error())
		}
		if err := sqlopt.WithSkillID(skillId).Apply(tx).Delete(&model.CustomSkillPublish{}).Error; err != nil {
			return toErrStatus("mcp_custom_skill_delete_publish", err.Error())
		}
		if err := sqlopt.WithID(id).Apply(tx).Delete(&model.CustomSkill{}).Error; err != nil {
			return toErrStatus("mcp_custom_skill_delete", err.Error())
		}
		return nil
	})
}

func (c *Client) GetCustomSkill(ctx context.Context, skillId string) (*model.CustomSkill, *err_code.Status) {
	if skillId == "" {
		return nil, toErrStatus("mcp_custom_skill_invalid_arg", "skillId is empty")
	}
	var cs model.CustomSkill
	result := sqlopt.WithID(util.MustU32(skillId)).Apply(c.db).WithContext(ctx).First(&cs)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, toErrStatus("mcp_custom_skill_not_found", skillId)
		}
		return nil, toErrStatus("mcp_custom_skill_get", skillId, result.Error.Error())
	}
	return &cs, nil
}

// GetCustomSkillByPreviewThreadID 仅匹配列 preview_thread_id；previewThreadId 为空时返回 (nil, nil)；未找到返回 (nil, nil)；仅数据库失败返回 Status。
func (c *Client) GetCustomSkillByPreviewThreadID(ctx context.Context, previewThreadID string) (*model.CustomSkill, *err_code.Status) {
	if previewThreadID == "" {
		return nil, nil
	}
	var cs model.CustomSkill
	err := sqlopt.SQLOptions(
		sqlopt.WithCustomSkillPreviewThreadId(previewThreadID),
	).Apply(c.db).WithContext(ctx).Model(&model.CustomSkill{}).First(&cs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, toErrStatus("mcp_custom_skill_get_by_preview_thread", err.Error())
	}
	return &cs, nil
}

// GetCustomSkillByWgaThreadID 按 wga_thread_id 查找；wgaThreadId 为空时返回 (nil, nil)；未找到返回 (nil, nil)；仅数据库失败返回 Status。
func (c *Client) GetCustomSkillByWgaThreadID(ctx context.Context, wgaThreadID string) (*model.CustomSkill, *err_code.Status) {
	if wgaThreadID == "" {
		return nil, nil
	}
	var cs model.CustomSkill
	err := sqlopt.SQLOptions(
		sqlopt.WithCustomSkillWgaThreadId(wgaThreadID),
	).Apply(c.db).WithContext(ctx).Model(&model.CustomSkill{}).First(&cs).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, toErrStatus("mcp_custom_skill_get_by_wga_thread", err.Error())
	}
	return &cs, nil
}

// GetCustomSkillListByWgaThreadIDList 在 identity 下按 wga_thread_id IN 批量查询；去空后列表为空或 userId/orgId 为空时返回空切片；仅数据库失败返回 Status。
func (c *Client) GetCustomSkillListByWgaThreadIDList(ctx context.Context, userId, orgId string, wgaThreadIDList []string) ([]*model.CustomSkill, *err_code.Status) {
	nonEmpty := make([]string, 0, len(wgaThreadIDList))
	for _, id := range wgaThreadIDList {
		if id != "" {
			nonEmpty = append(nonEmpty, id)
		}
	}
	if len(nonEmpty) == 0 || userId == "" || orgId == "" {
		return []*model.CustomSkill{}, nil
	}
	var list []*model.CustomSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
		sqlopt.WithCustomSkillWgaThreadIdList(nonEmpty),
	).Apply(c.db).WithContext(ctx).Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_custom_skill_get_by_wga_thread_list", err.Error())
	}
	return list, nil
}

// GetCustomSkillList 返回的 total：无分页，表示当前筛选条件下全量条数，与 len(list) 一致。
func (c *Client) GetCustomSkillList(ctx context.Context, userId, orgId, name string) ([]*model.CustomSkill, int64, *err_code.Status) {
	var list []*model.CustomSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
		sqlopt.LikeName(name),
	).Apply(c.db).WithContext(ctx).Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, 0, toErrStatus("mcp_custom_skill_list", err.Error())
	}
	return list, int64(len(list)), nil
}

// GetCustomSkillPageList 管理员中心分页查询：按 userId[]/orgId[]/name 过滤后分页，total 为符合条件全量条数。
// userId 与 orgId 均为空时不按拥有者过滤（全局列表）。
func (c *Client) GetCustomSkillPageList(ctx context.Context, userId, orgId []string, name string, pageNum, pageSize int) ([]*model.CustomSkill, int64, *err_code.Status) {
	query := sqlopt.SQLOptions(
		sqlopt.WithUserIDList(userId),
		sqlopt.WithOrgIDList(orgId),
		sqlopt.LikeName(db.EscapeLike(name)),
	).Apply(c.db).WithContext(ctx).Model(&model.CustomSkill{})

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, toErrStatus("mcp_custom_skill_page_list", err.Error())
	}
	if total == 0 {
		return []*model.CustomSkill{}, 0, nil
	}
	var list []*model.CustomSkill
	if err := query.
		Order("created_at DESC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, toErrStatus("mcp_custom_skill_page_list", err.Error())
	}
	return list, total, nil
}

// GetCustomSkillListByIDs 按主键 id 批量查询；name 非空时做名称模糊过滤。
func (c *Client) GetCustomSkillListByIDs(ctx context.Context, ids []string, name string) ([]*model.CustomSkill, *err_code.Status) {
	if len(ids) == 0 {
		return []*model.CustomSkill{}, nil
	}
	uids := make([]uint32, 0, len(ids))
	for _, id := range ids {
		if id != "" {
			uids = append(uids, util.MustU32(id))
		}
	}
	if len(uids) == 0 {
		return []*model.CustomSkill{}, nil
	}
	var list []*model.CustomSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithIDs(uids),
		sqlopt.LikeName(name),
	).Apply(c.db).WithContext(ctx).Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_custom_skill_list_by_ids", err.Error())
	}
	return list, nil
}

func (c *Client) GetCustomSkillBySkillIds(ctx context.Context, skillIds []string) ([]*model.CustomSkill, *err_code.Status) {
	if len(skillIds) == 0 {
		return []*model.CustomSkill{}, nil
	}
	var list []*model.CustomSkill
	if err := sqlopt.SQLOptions(
		sqlopt.WithCustomSkillSkillId(skillIds),
	).Apply(c.db).WithContext(ctx).Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_custom_skill_get_by_skill_ids", err.Error())
	}
	return list, nil
}

func (c *Client) UpdateCustomSkillBasicMeta(ctx context.Context, skillId, name, desc string) *err_code.Status {
	updates := map[string]any{
		"name": name,
		"desc": desc,
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(util.MustU32(skillId)),
	).Apply(c.db).WithContext(ctx).Model(&model.CustomSkill{}).Updates(updates).Error; err != nil {
		return toErrStatus("mcp_custom_skill_update_basic_meta", skillId, err.Error())
	}
	return nil
}

func (c *Client) UpdateCustomSkillThreadMeta(ctx context.Context, skillId, wgaThreadId, previewThreadId string) *err_code.Status {
	updates := map[string]any{
		"wga_thread_id":     wgaThreadId,
		"preview_thread_id": previewThreadId,
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(util.MustU32(skillId)),
	).Apply(c.db).WithContext(ctx).Model(&model.CustomSkill{}).Updates(updates).Error; err != nil {
		return toErrStatus("mcp_custom_skill_update_thread_meta", skillId, err.Error())
	}
	return nil
}
