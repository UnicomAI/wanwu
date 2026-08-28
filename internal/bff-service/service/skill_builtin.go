package service

import (
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

func GetAgentSkillDetail(ctx *gin.Context, skillId string) (*response.SkillDetail, error) {
	skillsCfg, exist := config.Cfg().AgentSkill(skillId)
	if !exist {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "bff_agent_skill_detail", "get skill detail empty")
	}
	return buildSkillTempDetail(skillsCfg, true), nil
}

// GetAgentSkillListDetail 返回内置 skill 详情列表。当 userId / orgId 非空时，
// 会额外通过 mcp.GetBuiltinSkillVars 拉取每个 skill 的 per-user 变量集填入 Variables；
// userId / orgId 缺失时跳过 vars 填充（保持老 caller 行为）。
// 单个 skill 拉 vars 失败只打 Warn 跳过，不影响别的 skill。
func GetAgentSkillListDetail(ctx *gin.Context, userId, orgId string, skillIdList []string) (*response.SkillDetailListResp, error) {
	var skillDetailList []*response.SkillDetail
	for _, skillId := range skillIdList {
		skillsCfg, exist := config.Cfg().AgentSkill(skillId)
		if !exist {
			continue
		}
		detail := buildSkillTempDetail(skillsCfg, false)
		detail.SkillPath = skillsCfg.SkillPath

		if userId != "" && orgId != "" {
			vars, varsErr := getBuiltinSkillVariables(ctx, userId, orgId, skillId)
			if varsErr != nil {
				log.Warnf("callback builtin skill list failed to fetch variables, skillId: %s, err: %v", skillId, varsErr)
			} else {
				detail.Variables = toSkillVariables(vars)
			}
		}

		skillDetailList = append(skillDetailList, detail)
	}
	return &response.SkillDetailListResp{SkillList: skillDetailList}, nil
}

func GetBuiltinSkillList(ctx *gin.Context, name string) (*response.ListResult, error) {
	var list []*response.BuiltinSkillInfo
	var skillIds []string
	for _, skillsCfg := range config.Cfg().AgentSkills {
		if name != "" && !strings.Contains(skillsCfg.Name, name) {
			continue
		}
		info := buildBuiltinSkillInfo(*skillsCfg)
		list = append(list, &info)
		skillIds = append(skillIds, skillsCfg.SkillId)
	}
	// 批量填充下载计数
	downloadMap := getBuiltinSkillDownloadCounts(ctx, skillIds)
	for i, info := range list {
		if info != nil {
			list[i].DownloadCount = downloadMap[info.SkillId]
		}
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

func GetBuiltinSkillDetail(ctx *gin.Context, userId, orgId, skillId string) (*response.BuiltinSkillDetail, error) {
	skillsCfg, exist := config.Cfg().AgentSkill(skillId)
	if !exist {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_builtin_not_found", "skill not found in builtin skills")
	}

	detail := &response.BuiltinSkillDetail{
		BuiltinSkillInfo: buildBuiltinSkillInfo(skillsCfg),
		SkillMarkdown:    string(skillsCfg.SkillMarkdown),
	}
	// 填充下载计数
	downloadMap := getBuiltinSkillDownloadCounts(ctx, []string{skillId})
	detail.DownloadCount = downloadMap[skillId]

	configResp, err := mcp.GetBuiltinSkillVars(ctx.Request.Context(), &mcp_service.GetBuiltinSkillVarsReq{
		SkillId:  skillId,
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}
	if configResp != nil {
		detail.Variables = append(detail.Variables, toSkillVariables(configResp.Variables)...)
	}
	return detail, nil
}

func DownloadBuiltinSkill(ctx *gin.Context, skillId string) ([]byte, error) {
	skillsCfg, exist := config.Cfg().AgentSkill(skillId)
	if !exist {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_builtin_not_found", "skill not found in builtin skills")
	}
	// 递增下载计数
	incrementBuiltinSkillDownload(ctx, skillId)
	return skillsCfg.AgentSkillZipToBytes(skillId)
}

// --- internal ---
func buildBuiltinSkillInfo(skillsCfg config.SkillsConfig) response.BuiltinSkillInfo {
	iconUrl := config.Cfg().DefaultIcon.SkillIcon
	if skillsCfg.Avatar != "" {
		iconUrl = skillsCfg.Avatar
	}
	return response.BuiltinSkillInfo{
		SkillBasicInfo: response.SkillBasicInfo{
			SkillId: skillsCfg.SkillId,
			Name:    skillsCfg.Name,
			Avatar:  request.Avatar{Path: iconUrl},
			Author:  skillsCfg.Author,
			Desc:    skillsCfg.Desc,
		},
	}
}

func buildSkillTempDetail(skillsCfg config.SkillsConfig, needMd bool) *response.SkillDetail {
	iconUrl := config.Cfg().DefaultIcon.SkillIcon
	if skillsCfg.Avatar != "" {
		iconUrl = skillsCfg.Avatar
	}
	ret := &response.SkillDetail{
		SkillBasicInfo: response.SkillBasicInfo{
			SkillId: skillsCfg.SkillId,
			Name:    skillsCfg.Name,
			Avatar:  request.Avatar{Path: iconUrl},
			Author:  skillsCfg.Author,
			Desc:    skillsCfg.Desc,
		},
	}
	if needMd {
		ret.SkillMarkdown = string(skillsCfg.SkillMarkdown)
	}
	return ret
}

// getBuiltinSkillVariables 通过 mcp gRPC 拉取单个内置 skill 的变量集（per-user）。
// userId/orgId 缺一不可（mcp.GetBuiltinSkillVars 要求 Identity）；缺时直接返回 nil，
// 让 callback 退化为不带 vars 的行为（向后兼容旧 caller）。
func getBuiltinSkillVariables(ctx *gin.Context, userId, orgId, skillId string) ([]*mcp_service.Variable, error) {
	if userId == "" || orgId == "" {
		return nil, nil
	}
	varResp, err := mcp.GetBuiltinSkillVars(ctx.Request.Context(), &mcp_service.GetBuiltinSkillVarsReq{
		SkillId:  skillId,
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}
	if varResp == nil {
		return nil, nil
	}
	return varResp.GetVariables(), nil
}
