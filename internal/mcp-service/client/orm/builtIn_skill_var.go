package orm

import (
	"context"
	"errors"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) CreateBuiltinSkillVar(ctx context.Context, userId, orgId string, variable *model.BuiltinSkillVariable) (uint32, *errs.Status) {
	if variable == nil || variable.SkillID == "" || variable.Name == "" {
		return 0, toErrStatus("mcp_skill_var_invalid_arg")
	}
	cnt, err := c.countSkillVarByName(ctx, &model.BuiltinSkillVariable{}, variable.SkillID, userId, orgId, variable.Name, 0)
	if err != nil {
		return 0, toErrStatus("mcp_builtin_skill_var_create", err.Error())
	}
	if cnt > 0 {
		return 0, toErrStatus("mcp_skill_var_duplicate_name")
	}
	variable.UserID = userId
	variable.OrgID = orgId
	if err := c.db.WithContext(ctx).Create(variable).Error; err != nil {
		return 0, toErrStatus("mcp_builtin_skill_var_create", err.Error())
	}
	return variable.ID, nil
}

func (c *Client) UpdateBuiltinSkillVar(ctx context.Context, userId, orgId string, id uint32, variable *model.BuiltinSkillVariable) *errs.Status {
	if id == 0 || variable == nil || variable.Name == "" {
		return toErrStatus("mcp_skill_var_invalid_arg")
	}
	var row model.BuiltinSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_builtin_skill_var_not_found")
		}
		return toErrStatus("mcp_builtin_skill_var_update", err.Error())
	}
	cnt, err := c.countSkillVarByName(ctx, &model.BuiltinSkillVariable{}, row.SkillID, userId, orgId, variable.Name, id)
	if err != nil {
		return toErrStatus("mcp_builtin_skill_var_update", err.Error())
	}
	if cnt > 0 {
		return toErrStatus("mcp_skill_var_duplicate_name")
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx).Model(&model.BuiltinSkillVariable{})).
		Updates(map[string]any{
			"name":           variable.Name,
			"desc":           variable.Desc,
			"variable_key":   variable.VariableKey,
			"variable_value": variable.VariableValue,
		}).Error; err != nil {
		return toErrStatus("mcp_builtin_skill_var_update", err.Error())
	}
	return nil
}

func (c *Client) DeleteBuiltinSkillVar(ctx context.Context, userId, orgId string, id uint32) *errs.Status {
	if id == 0 {
		return toErrStatus("mcp_skill_var_invalid_arg")
	}
	var row model.BuiltinSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_builtin_skill_var_not_found")
		}
		return toErrStatus("mcp_builtin_skill_var_delete", err.Error())
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).Delete(&model.BuiltinSkillVariable{}).Error; err != nil {
		return toErrStatus("mcp_builtin_skill_var_delete", err.Error())
	}
	return nil
}

// GetBuiltinSkillVars 返回的 total：无分页，表示当前 skill 下变量全量条数，与 len(variables) 一致。
func (c *Client) GetBuiltinSkillVars(ctx context.Context, userId, orgId, skillId string) ([]*model.BuiltinSkillVariable, int64, *errs.Status) {
	if skillId == "" {
		return nil, 0, toErrStatus("mcp_skill_var_invalid_arg")
	}
	var list []*model.BuiltinSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithSkillID(skillId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).Find(&list).Error; err != nil {
		return nil, 0, toErrStatus("mcp_builtin_skill_var_list", err.Error())
	}
	return list, int64(len(list)), nil
}

func (c *Client) GetBuiltinSkillsVars(ctx context.Context, userId, orgId string, skillIds []string) (map[string][]*model.BuiltinSkillVariable, *errs.Status) {
	var list []*model.BuiltinSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithSkillIDs(skillIds),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_builtin_skills_var_list", err.Error())
	}
	result := make(map[string][]*model.BuiltinSkillVariable)
	for _, v := range list {
		result[v.SkillID] = append(result[v.SkillID], v)
	}
	return result, nil
}

// --- internal ---

func (c *Client) countSkillVarByName(ctx context.Context, tab interface{}, skillID, userID, orgID, name string, excludeID uint32) (int64, error) {
	var cnt int64
	db := sqlopt.SQLOptions(
		sqlopt.WithSkillID(skillID),
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithVariableName(name),
	).Apply(c.db.WithContext(ctx).Model(tab))
	if excludeID > 0 {
		db = db.Where("id <> ?", excludeID)
	}
	err := db.Count(&cnt).Error
	return cnt, err
}
