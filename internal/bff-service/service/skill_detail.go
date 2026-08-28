package service

import (
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// GetCallbackSkillDetail 根据 skillType 和 skillId 获取 skill 详情。
func GetCallbackSkillDetail(ctx *gin.Context, skillId, skillType string) (*response.CallbackSkillDetail, error) {
	switch skillType {
	case constant.SkillTypeCustom:
		return getCustomCallbackSkillDetail(ctx, skillId)
	case constant.SkillTypeBuiltIn:
		return getBuiltinCallbackSkillDetail(skillId)
	default:
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "skill_detail_unknown_type", "unsupported skill type: "+skillType)
	}
}

func getCustomCallbackSkillDetail(ctx *gin.Context, skillId string) (*response.CallbackSkillDetail, error) {
	resp, err := mcp.GetCustomSkillDetailByIdList(ctx.Request.Context(), &mcp_service.CustomSkillDetailByIdListReq{
		SkillIds: []string{skillId},
	})
	if err != nil {
		return nil, err
	}
	if len(resp.SkillDetails) == 0 {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "skill_detail_not_found", "custom skill not found: "+skillId)
	}
	skill := resp.SkillDetails[0]
	return &response.CallbackSkillDetail{
		SkillId:    skill.SkillId,
		SkillType:  constant.SkillTypeCustom,
		Name:       skill.Name,
		Desc:       skill.Desc,
		Avatar:     skill.Avatar,
		ObjectPath: skill.ObjectPath,
	}, nil
}

func getBuiltinCallbackSkillDetail(skillId string) (*response.CallbackSkillDetail, error) {
	skillsCfg, exist := config.Cfg().AgentSkill(skillId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "skill_detail_not_found", "builtin skill not found: "+skillId)
	}
	iconUrl := config.Cfg().DefaultIcon.SkillIcon
	if skillsCfg.Avatar != "" {
		iconUrl = skillsCfg.Avatar
	}
	return &response.CallbackSkillDetail{
		SkillId:    skillsCfg.SkillId,
		SkillType:  constant.SkillTypeBuiltIn,
		Name:       skillsCfg.Name,
		Desc:       skillsCfg.Desc,
		Avatar:     iconUrl,
		ObjectPath: skillsCfg.SkillPath,
	}, nil
}
