package orm

import (
	"context"
	"errors"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/orm/sqlopt"
	"gorm.io/gorm"
)

func (c *Client) CreateAcquiredSkillVar(ctx context.Context, userId, orgId string, variable *model.AcquiredSkillVariable) (uint32, *errs.Status) {
	if variable == nil || variable.AcquiredSkillID == "" || variable.Name == "" {
		return 0, toErrStatus("mcp_skill_var_invalid_arg")
	}
	cnt, err := c.countAcquiredSkillVarByName(ctx, variable.AcquiredSkillID, userId, orgId, variable.Name, 0)
	if err != nil {
		return 0, toErrStatus("mcp_acquired_skill_var_create", err.Error())
	}
	if cnt > 0 {
		return 0, toErrStatus("mcp_skill_var_duplicate_name")
	}
	variable.UserID = userId
	variable.OrgID = orgId
	if err := c.db.WithContext(ctx).Create(variable).Error; err != nil {
		return 0, toErrStatus("mcp_acquired_skill_var_create", err.Error())
	}
	return variable.ID, nil
}

func (c *Client) UpdateAcquiredSkillVar(ctx context.Context, userId, orgId string, id uint32, variable *model.AcquiredSkillVariable) *errs.Status {
	if id == 0 || variable == nil || variable.Name == "" {
		return toErrStatus("mcp_skill_var_invalid_arg")
	}
	var row model.AcquiredSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_acquired_skill_var_not_found")
		}
		return toErrStatus("mcp_acquired_skill_var_update", err.Error())
	}
	cnt, err := c.countAcquiredSkillVarByName(ctx, row.AcquiredSkillID, userId, orgId, variable.Name, id)
	if err != nil {
		return toErrStatus("mcp_acquired_skill_var_update", err.Error())
	}
	if cnt > 0 {
		return toErrStatus("mcp_skill_var_duplicate_name")
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx).Model(&model.AcquiredSkillVariable{})).
		Updates(map[string]any{
			"name":           variable.Name,
			"desc":           variable.Desc,
			"variable_key":   variable.VariableKey,
			"variable_value": variable.VariableValue,
		}).Error; err != nil {
		return toErrStatus("mcp_acquired_skill_var_update", err.Error())
	}
	return nil
}

func (c *Client) DeleteAcquiredSkillVar(ctx context.Context, userId, orgId string, id uint32) *errs.Status {
	if id == 0 {
		return toErrStatus("mcp_skill_var_invalid_arg")
	}
	var row model.AcquiredSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return toErrStatus("mcp_acquired_skill_var_not_found")
		}
		return toErrStatus("mcp_acquired_skill_var_delete", err.Error())
	}
	if err := sqlopt.SQLOptions(
		sqlopt.WithID(id),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).Delete(&model.AcquiredSkillVariable{}).Error; err != nil {
		return toErrStatus("mcp_acquired_skill_var_delete", err.Error())
	}
	return nil
}

func (c *Client) GetAcquiredSkillVarByID(ctx context.Context, id uint32) (*model.AcquiredSkillVariable, *errs.Status) {
	if id == 0 {
		return nil, toErrStatus("mcp_skill_var_invalid_arg")
	}
	var row model.AcquiredSkillVariable
	if err := sqlopt.SQLOptions(sqlopt.WithID(id)).Apply(c.db).WithContext(ctx).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, toErrStatus("mcp_acquired_skill_var_not_found")
		}
		return nil, toErrStatus("mcp_acquired_skill_var_get", err.Error())
	}
	return &row, nil
}

func (c *Client) GetAcquiredSkillVars(ctx context.Context, userId, orgId, acquiredSkillId string) ([]*model.AcquiredSkillVariable, *errs.Status) {
	if acquiredSkillId == "" {
		return nil, toErrStatus("mcp_skill_var_invalid_arg")
	}
	var list []*model.AcquiredSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithAcquiredSkillID(acquiredSkillId),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db).WithContext(ctx).Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_acquired_skill_var_list", err.Error())
	}
	return list, nil
}

func (c *Client) GetAcquiredSkillVarsByAcquiredSkillIDs(ctx context.Context, userId, orgId string, acquiredSkillIDs []string) (map[string][]*model.AcquiredSkillVariable, *errs.Status) {
	if len(acquiredSkillIDs) == 0 {
		return map[string][]*model.AcquiredSkillVariable{}, nil
	}
	var list []*model.AcquiredSkillVariable
	if err := sqlopt.SQLOptions(
		sqlopt.WithAcquiredSkillIDs(acquiredSkillIDs),
		sqlopt.WithUserID(userId),
		sqlopt.WithOrgID(orgId),
	).Apply(c.db.WithContext(ctx)).Find(&list).Error; err != nil {
		return nil, toErrStatus("mcp_acquired_skill_var_list", err.Error())
	}
	out := make(map[string][]*model.AcquiredSkillVariable)
	for _, v := range list {
		out[v.AcquiredSkillID] = append(out[v.AcquiredSkillID], v)
	}
	return out, nil
}

func (c *Client) countAcquiredSkillVarByName(ctx context.Context, acquiredSkillID, userID, orgID, name string, excludeID uint32) (int64, error) {
	var cnt int64
	db := sqlopt.SQLOptions(
		sqlopt.WithAcquiredSkillID(acquiredSkillID),
		sqlopt.WithUserID(userID),
		sqlopt.WithOrgID(orgID),
		sqlopt.WithVariableName(name),
	).Apply(c.db.WithContext(ctx).Model(&model.AcquiredSkillVariable{}))
	if excludeID > 0 {
		db = db.Where("id <> ?", excludeID)
	}
	err := db.Count(&cnt).Error
	return cnt, err
}
