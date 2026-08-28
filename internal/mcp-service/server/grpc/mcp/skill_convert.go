package mcp

import (
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/util"
)

func toProtoCustomSkill(cs *model.CustomSkill) *mcp_service.CustomSkill {
	if cs == nil {
		return nil
	}
	return &mcp_service.CustomSkill{
		SkillId:         util.Int2Str(cs.ID),
		Name:            cs.Name,
		Avatar:          cs.Avatar,
		Author:          cs.Author,
		Desc:            cs.Desc,
		ObjectPath:      cs.ObjectPath,
		WgaThreadId:     cs.WgaThreadID,
		PreviewThreadId: cs.PreviewThreadID,
		Identity:        &mcp_service.Identity{UserId: cs.UserID, OrgId: cs.OrgID},
		CreatedAt:       cs.CreatedAt,
		UpdatedAt:       cs.UpdatedAt,
		DownloadCount:   cs.DownloadCount,
		AcquiredCount:   cs.AcquiredCount,
		UserId:          cs.UserID,
		OrgId:           cs.OrgID,
	}
}

func toProtoPublishCustomSkill(cs *model.CustomSkill, publish *model.CustomSkillPublish) *mcp_service.PublishCustomSkill {
	if cs == nil {
		return nil
	}
	out := &mcp_service.PublishCustomSkill{
		Skill: toProtoCustomSkill(cs),
	}
	if publish != nil {
		out.ObjectPath = publish.ObjectPath
		out.Markdown = publish.Markdown
		out.Version = publish.Version
		out.VersionDesc = publish.VersionDescription
		out.CreatedAt = publish.CreatedAt
		out.UpdatedAt = publish.UpdatedAt
	}
	return out
}

func toProtoVariableFromCustom(v *model.CustomSkillVariable) *mcp_service.Variable {
	if v == nil {
		return nil
	}
	return &mcp_service.Variable{
		Id:            util.Int2Str(v.ID),
		Name:          v.Name,
		Desc:          v.Desc,
		VariableKey:   v.VariableKey,
		VariableValue: v.VariableValue,
	}
}

func toProtoVariablesFromCustom(vars []*model.CustomSkillVariable) []*mcp_service.Variable {
	ret := make([]*mcp_service.Variable, 0, len(vars))
	for _, v := range vars {
		ret = append(ret, toProtoVariableFromCustom(v))
	}
	return ret
}

func toProtoVariableFromAcquired(v *model.AcquiredSkillVariable) *mcp_service.Variable {
	if v == nil {
		return nil
	}
	return &mcp_service.Variable{
		Id:            util.Int2Str(v.ID),
		Name:          v.Name,
		Desc:          v.Desc,
		VariableKey:   v.VariableKey,
		VariableValue: v.VariableValue,
	}
}

func toProtoVariableFromBuiltin(v *model.BuiltinSkillVariable) *mcp_service.Variable {
	if v == nil {
		return nil
	}
	return &mcp_service.Variable{
		Id:            util.Int2Str(v.ID),
		Name:          v.Name,
		Desc:          v.Desc,
		VariableKey:   v.VariableKey,
		VariableValue: v.VariableValue,
	}
}
