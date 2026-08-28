package orm

import (
	"context"
	"errors"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) CreateCustomSkillVar(ctx context.Context, userId, orgId string, variable *model.CustomSkillVariable) (uint32, *errs.Status) {
	if variable == nil || variable.SkillID == "" || variable.Name == "" {
		return 0, toErrStatus("mcp_skill_var_invalid_arg")
	}
	_, st := c.GetCustomSkill(ctx, variable.SkillID)
	if st != nil {
		return 0, st
	}
	cnt, err := c.countSkillVarByName(ctx, &model.CustomSkillVariable{}, variable.SkillID, userId, orgId, variable.Name, 0)
	if err != nil {
		return 0, toErrStatus("mcp_custom_skill_var_create", err.Error())
	}
	if cnt > 0 {
		return 0, toErrStatus("mcp_skill_var_duplicate_name")
	}
	variable.UserID = userId
	variable.OrgID = orgId
	if err := c.db.WithContext(ctx).Create(variable).Error; err != nil {
		return 0, toErrStatus("mcp_custom_skill_var_create", err.Error())
	}
	return variable.ID, nil
}

func (c *Client) UpdateCustomSkillVar(ctx context.Context, userId, orgId string, id uint32, variable *model.CustomSkillVariable) *errs.Status {
	if id == 0 || variable == nil || variable.Name == "" {
		return toErrStatus("mcp_skill_var_invalid_arg")
	}
	var row model.CustomSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_custom_skill_var_not_found")
		}
		return toErrStatus("mcp_custom_skill_var_update", err.Error())
	}
	_, st := c.GetCustomSkill(ctx, row.SkillID)
	if st != nil {
		return st
	}
	cnt, err := c.countSkillVarByName(ctx, &model.CustomSkillVariable{}, row.SkillID, userId, orgId, variable.Name, id)
	if err != nil {
		return toErrStatus("mcp_custom_skill_var_update", err.Error())
	}
	if cnt > 0 {
		return toErrStatus("mcp_skill_var_duplicate_name")
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx).Model(&model.CustomSkillVariable{})).
		Updates(map[string]any{
			"name":           variable.Name,
			"desc":           variable.Desc,
			"variable_key":   variable.VariableKey,
			"variable_value": variable.VariableValue,
		}).Error; err != nil {
		return toErrStatus("mcp_custom_skill_var_update", err.Error())
	}
	return nil
}

func (c *Client) DeleteCustomSkillVar(ctx context.Context, userId, orgId string, id uint32) *errs.Status {
	if id == 0 {
		return toErrStatus("mcp_skill_var_invalid_arg")
	}
	var row model.CustomSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_custom_skill_var_not_found")
		}
		return toErrStatus("mcp_custom_skill_var_delete", err.Error())
	}
	_, st := c.GetCustomSkill(ctx, row.SkillID)
	if st != nil {
		return st
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).Delete(&model.CustomSkillVariable{}).Error; err != nil {
		return toErrStatus("mcp_custom_skill_var_delete", err.Error())
	}
	return nil
}

func (c *Client) GetCustomSkillVarByID(ctx context.Context, id uint32) (*model.CustomSkillVariable, *errs.Status) {
	if id == 0 {
		return nil, toErrStatus("mcp_skill_var_invalid_arg")
	}
	var row model.CustomSkillVariable
	if err := sqlopt.SQLOptions(sqlopt.WithID(id)).Apply(c.db).WithContext(ctx).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toErrStatus("mcp_custom_skill_var_not_found")
		}
		return nil, toErrStatus("mcp_custom_skill_var_get", err.Error())
	}
	return &row, nil
}

func (c *Client) GetCustomSkillVars(ctx context.Context, userId, orgId, skillId string) ([]*model.CustomSkillVariable, *errs.Status) {
	if skillId == "" {
		return nil, toErrStatus("mcp_skill_var_invalid_arg")
	}
	var list []*model.CustomSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithSkillID(skillId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_custom_skill_var_list", err.Error())
	}
	return list, nil
}

func (c *Client) GetCustomSkillVarsBySkillIDs(ctx context.Context, userId, orgId string, skillIds []string) (map[string][]*model.CustomSkillVariable, *errs.Status) {
	if len(skillIds) == 0 {
		return map[string][]*model.CustomSkillVariable{}, nil
	}
	var list []*model.CustomSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithSkillIDs(skillIds),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_custom_skill_var_list", err.Error())
	}
	out := make(map[string][]*model.CustomSkillVariable)
	for _, v := range list {
		out[v.SkillID] = append(out[v.SkillID], v)
	}
	return out, nil
}
