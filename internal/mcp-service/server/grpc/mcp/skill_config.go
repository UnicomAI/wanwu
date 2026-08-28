package mcp

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) UpdateCustomSkillBasicMeta(ctx context.Context, req *mcp_service.UpdateCustomSkillBasicMetaReq) (*emptypb.Empty, error) {
	if err := s.cli.UpdateCustomSkillBasicMeta(ctx, req.SkillId, req.Name, req.Desc); err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdateCustomSkillThreadMeta(ctx context.Context, req *mcp_service.UpdateCustomSkillThreadMetaReq) (*emptypb.Empty, error) {
	if err := s.cli.UpdateCustomSkillThreadMeta(ctx, req.SkillId, req.WgaThreadId, req.PreviewThreadId); err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) CreateCustomSkillVar(ctx context.Context, req *mcp_service.CreateCustomSkillVarReq) (*mcp_service.SkillVariableCreateResp, error) {
	customSkill, st := s.cli.GetCustomSkill(ctx, req.SkillId)
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	id, err := s.cli.CreateCustomSkillVar(ctx, customSkill.UserID, customSkill.OrgID, toCustomSkillVariable(req.SkillId, req.Variable))
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return &mcp_service.SkillVariableCreateResp{Id: util.Int2Str(id)}, nil
}

func (s *Service) GetCustomSkillVars(ctx context.Context, req *mcp_service.GetCustomSkillVarsReq) (*mcp_service.CustomSkillVars, error) {
	customSkill, st := s.cli.GetCustomSkill(ctx, req.SkillId)
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	variables, err := s.cli.GetCustomSkillVars(ctx, customSkill.UserID, customSkill.OrgID, req.SkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return &mcp_service.CustomSkillVars{
		SkillId:   req.SkillId,
		Variables: toProtoVariablesFromCustom(variables),
		Total:     int64(len(variables)),
	}, nil
}

func (s *Service) UpdateCustomSkillVar(ctx context.Context, req *mcp_service.UpdateCustomSkillVarReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, toErrStatus("mcp_custom_skill_var_update", "identity is empty"))
	}
	id := util.MustU32(req.GetId())
	variable, varSt := s.cli.GetCustomSkillVarByID(ctx, id)
	if varSt != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, varSt)
	}
	customSkill, st := s.cli.GetCustomSkill(ctx, variable.SkillID)
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	if err := s.cli.UpdateCustomSkillVar(ctx, customSkill.UserID, customSkill.OrgID, id, protoVariableToCustomModel(req.Variable)); err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteCustomSkillVar(ctx context.Context, req *mcp_service.DeleteCustomSkillVarReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, toErrStatus("mcp_custom_skill_var_delete", "identity is empty"))
	}
	id := util.MustU32(req.GetId())
	variable, varSt := s.cli.GetCustomSkillVarByID(ctx, id)
	if varSt != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, varSt)
	}
	customSkill, st := s.cli.GetCustomSkill(ctx, variable.SkillID)
	if st != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, st)
	}
	if err := s.cli.DeleteCustomSkillVar(ctx, customSkill.UserID, customSkill.OrgID, id); err != nil {
		return nil, errStatus(errs.Code_MCPCustomSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) CreateAcquiredSkillVar(ctx context.Context, req *mcp_service.CreateAcquiredSkillVarReq) (*mcp_service.SkillVariableCreateResp, error) {
	acquired, st := s.cli.GetAcquiredSkill(ctx, req.AcquiredSkillId)
	if st != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, st)
	}
	id, err := s.cli.CreateAcquiredSkillVar(ctx, acquired.UserID, acquired.OrgID, toAcquiredSkillVariable(req.AcquiredSkillId, req.Variable))
	if err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	return &mcp_service.SkillVariableCreateResp{Id: util.Int2Str(id)}, nil
}

func (s *Service) GetAcquiredSkillVars(ctx context.Context, req *mcp_service.GetAcquiredSkillVarsReq) (*mcp_service.AcquiredSkillVars, error) {
	acquired, st := s.cli.GetAcquiredSkill(ctx, req.AcquiredSkillId)
	if st != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, st)
	}
	variables, err := s.cli.GetAcquiredSkillVars(ctx, acquired.UserID, acquired.OrgID, req.AcquiredSkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	protoVars := make([]*mcp_service.Variable, 0, len(variables))
	for _, v := range variables {
		protoVars = append(protoVars, toProtoVariableFromAcquired(v))
	}
	return &mcp_service.AcquiredSkillVars{
		AcquiredSkillId: req.AcquiredSkillId,
		Variables:       protoVars,
		Total:           int64(len(variables)),
	}, nil
}

func (s *Service) UpdateAcquiredSkillVar(ctx context.Context, req *mcp_service.UpdateAcquiredSkillVarReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, toErrStatus("mcp_acquired_skill_var_update", "identity is empty"))
	}
	id := util.MustU32(req.GetId())
	variable, varSt := s.cli.GetAcquiredSkillVarByID(ctx, id)
	if varSt != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, varSt)
	}
	acquired, st := s.cli.GetAcquiredSkill(ctx, variable.AcquiredSkillID)
	if st != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, st)
	}
	// 校验归属
	if acquired.UserID != req.GetIdentity().GetUserId() || acquired.OrgID != req.GetIdentity().GetOrgId() {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, toErrStatus("mcp_acquired_skill_var_update", "identity mismatch"))
	}
	if err := s.cli.UpdateAcquiredSkillVar(ctx, acquired.UserID, acquired.OrgID, id, protoVariableToAcquiredModel(req.Variable)); err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteAcquiredSkillVar(ctx context.Context, req *mcp_service.DeleteAcquiredSkillVarReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, toErrStatus("mcp_acquired_skill_var_delete", "identity is empty"))
	}
	id := util.MustU32(req.GetId())
	variable, varSt := s.cli.GetAcquiredSkillVarByID(ctx, id)
	if varSt != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, varSt)
	}
	acquired, st := s.cli.GetAcquiredSkill(ctx, variable.AcquiredSkillID)
	if st != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, st)
	}
	// 校验归属
	if acquired.UserID != req.GetIdentity().GetUserId() || acquired.OrgID != req.GetIdentity().GetOrgId() {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, toErrStatus("mcp_acquired_skill_var_delete", "identity mismatch"))
	}
	if err := s.cli.DeleteAcquiredSkillVar(ctx, acquired.UserID, acquired.OrgID, id); err != nil {
		return nil, errStatus(errs.Code_MCPAcquiredSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) CreateBuiltinSkillVar(ctx context.Context, req *mcp_service.CreateBuiltinSkillVarReq) (*mcp_service.SkillVariableCreateResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, toErrStatus("mcp_builtin_skill_var_create", "identity is empty"))
	}
	id, err := s.cli.CreateBuiltinSkillVar(ctx, req.GetIdentity().GetUserId(), req.GetIdentity().GetOrgId(), toBuiltinSkillVariable(req.SkillId, req.Variable))
	if err != nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, err)
	}
	return &mcp_service.SkillVariableCreateResp{Id: util.Int2Str(id)}, nil
}

func (s *Service) UpdateBuiltinSkillVar(ctx context.Context, req *mcp_service.UpdateBuiltinSkillVarReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, toErrStatus("mcp_builtin_skill_var_update", "identity is empty"))
	}
	id := util.MustU32(req.GetId())
	if err := s.cli.UpdateBuiltinSkillVar(ctx, req.GetIdentity().GetUserId(), req.GetIdentity().GetOrgId(), id, protoVariableToBuiltinModel(req.Variable)); err != nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteBuiltinSkillVar(ctx context.Context, req *mcp_service.DeleteBuiltinSkillVarReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, toErrStatus("mcp_builtin_skill_var_delete", "identity is empty"))
	}
	id := util.MustU32(req.GetId())
	if err := s.cli.DeleteBuiltinSkillVar(ctx, req.GetIdentity().GetUserId(), req.GetIdentity().GetOrgId(), id); err != nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetBuiltinSkillVars(ctx context.Context, req *mcp_service.GetBuiltinSkillVarsReq) (*mcp_service.BuiltinSkillVars, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, toErrStatus("mcp_builtin_skill_var_list", "identity is empty"))
	}
	variables, total, err := s.cli.GetBuiltinSkillVars(ctx, req.GetIdentity().GetUserId(), req.GetIdentity().GetOrgId(), req.SkillId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, err)
	}
	protoVars := make([]*mcp_service.Variable, 0, len(variables))
	for _, variable := range variables {
		protoVars = append(protoVars, toProtoVariableFromBuiltin(variable))
	}
	return &mcp_service.BuiltinSkillVars{
		SkillId:   req.SkillId,
		Variables: protoVars,
		Total:     total,
	}, nil
}

func (s *Service) GetBuiltinSkillsVars(ctx context.Context, req *mcp_service.GetBuiltinSkillsVarsReq) (*mcp_service.GetBuiltinSkillsVarsResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, toErrStatus("mcp_builtin_skills_var_list", "identity is empty"))
	}
	skills, err := s.cli.GetBuiltinSkillsVars(ctx, req.GetIdentity().GetUserId(), req.GetIdentity().GetOrgId(), req.SkillIds)
	if err != nil {
		return nil, errStatus(errs.Code_MCPBuiltinSkillErr, err)
	}
	ret := &mcp_service.GetBuiltinSkillsVarsResp{}
	for skillId, variables := range skills {
		protoVars := make([]*mcp_service.Variable, 0, len(variables))
		for _, variable := range variables {
			protoVars = append(protoVars, toProtoVariableFromBuiltin(variable))
		}
		ret.Skills = append(ret.Skills, &mcp_service.BuiltinSkillVars{
			SkillId:   skillId,
			Variables: protoVars,
			Total:     int64(len(variables)),
		})
	}
	return ret, nil
}

// --- internal ---

func toCustomSkillVariable(skillId string, variable *mcp_service.Variable) *model.CustomSkillVariable {
	if variable == nil {
		return &model.CustomSkillVariable{SkillID: skillId}
	}
	return &model.CustomSkillVariable{
		SkillID:       skillId,
		Name:          variable.Name,
		Desc:          variable.Desc,
		VariableKey:   variable.VariableKey,
		VariableValue: variable.VariableValue,
	}
}

func protoVariableToCustomModel(variable *mcp_service.Variable) *model.CustomSkillVariable {
	if variable == nil {
		return nil
	}
	return &model.CustomSkillVariable{
		Name:          variable.Name,
		Desc:          variable.Desc,
		VariableKey:   variable.VariableKey,
		VariableValue: variable.VariableValue,
	}
}

func toAcquiredSkillVariable(acquiredSkillId string, variable *mcp_service.Variable) *model.AcquiredSkillVariable {
	if variable == nil {
		return &model.AcquiredSkillVariable{AcquiredSkillID: acquiredSkillId}
	}
	return &model.AcquiredSkillVariable{
		AcquiredSkillID: acquiredSkillId,
		Name:            variable.Name,
		Desc:            variable.Desc,
		VariableKey:     variable.VariableKey,
		VariableValue:   variable.VariableValue,
	}
}

func protoVariableToAcquiredModel(variable *mcp_service.Variable) *model.AcquiredSkillVariable {
	if variable == nil {
		return nil
	}
	return &model.AcquiredSkillVariable{
		Name:          variable.Name,
		Desc:          variable.Desc,
		VariableKey:   variable.VariableKey,
		VariableValue: variable.VariableValue,
	}
}

func toBuiltinSkillVariable(skillId string, variable *mcp_service.Variable) *model.BuiltinSkillVariable {
	if variable == nil {
		return &model.BuiltinSkillVariable{SkillID: skillId}
	}
	return &model.BuiltinSkillVariable{
		SkillID:       skillId,
		Name:          variable.Name,
		Desc:          variable.Desc,
		VariableKey:   variable.VariableKey,
		VariableValue: variable.VariableValue,
	}
}

func protoVariableToBuiltinModel(variable *mcp_service.Variable) *model.BuiltinSkillVariable {
	if variable == nil {
		return nil
	}
	return &model.BuiltinSkillVariable{
		Name:          variable.Name,
		Desc:          variable.Desc,
		VariableKey:   variable.VariableKey,
		VariableValue: variable.VariableValue,
	}
}
