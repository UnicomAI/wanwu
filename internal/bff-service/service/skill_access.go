package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

// checkAccessibleSkillApp 从 AppInfo 列表中过滤用户可访问的第一条记录。
// 同一 skillId 可能有多条发布记录（不同组织发布），按顺序返回第一条可访问的。
// 发布类型权限判断：public 全部可访问，organization 仅同组织可访问，private 仅创建者可访问。
func checkAccessibleSkillApp(appInfos []*app_service.AppInfo, userID, orgID string) *app_service.AppInfo {
	for _, appInfo := range appInfos {
		if appInfo == nil {
			continue
		}
		switch appInfo.GetPublishType() {
		case constant.AppPublishPublic:
			return appInfo
		case constant.AppPublishOrganization:
			if appInfo.GetOrgId() == orgID {
				return appInfo
			}
		case constant.AppPublishPrivate:
			if appInfo.GetUserId() == userID {
				return appInfo
			}
		}
	}
	return nil
}

// getSourceSkillAppInfoMap 批量获取源 Skill 的发布信息。
// 返回 sourceSkillId 到 AppInfo 列表的映射（同一 Skill 可能有多个发布记录）。
func getSourceSkillAppInfoMap(ctx *gin.Context, sourceSkillIds []string) (map[string][]*app_service.AppInfo, error) {
	if len(sourceSkillIds) == 0 {
		return make(map[string][]*app_service.AppInfo), nil
	}
	appResp, err := app.GetAppListByIds(ctx.Request.Context(), &app_service.GetAppListByIdsReq{
		AppIdsList: sourceSkillIds,
		AppType:    constant.AppTypeSkill,
	})
	if err != nil {
		return nil, err
	}
	result := make(map[string][]*app_service.AppInfo)
	if appResp == nil {
		return result, nil
	}
	for _, info := range appResp.GetInfos() {
		if info == nil || info.GetAppId() == "" {
			continue
		}
		result[info.GetAppId()] = append(result[info.GetAppId()], info)
	}
	return result, nil
}

// isAcquiredSkillAccessible 判断 acquired skill 是否可使用。
// 同时检查：(1) 源 Skill 的发布范围对添加者可见；(2) Skill 有有效的发布包。
// 访问主体为 acquired 记录的所有者（skill.UserId / skill.OrgId）。
func isAcquiredSkillAccessible(ctx *gin.Context, skill *mcp_service.AcquiredSkill, appInfoMap map[string][]*app_service.AppInfo) bool {
	if skill == nil {
		return false
	}

	// 必须有有效地发布包
	publish := skill.GetSkill()
	if publish == nil {
		return false
	}
	if publish.GetVersion() == "" || publish.GetObjectPath() == "" {
		return false
	}

	// 获取源 Skill ID
	sourceSkillId := skill.GetSkill().GetSkill().GetSkillId()
	if sourceSkillId == "" {
		return false
	}

	// 获取源 Skill 的发布信息列表
	appInfos := appInfoMap[sourceSkillId]
	if len(appInfos) == 0 {
		return false
	}

	// 判断是否有任一发布记录可访问
	return checkAccessibleSkillApp(appInfos, skill.GetUserId(), skill.GetOrgId()) != nil
}

// checkAcquiredSkillAccess 校验用户是否有权限访问 acquired skill 的源 custom skill。
func checkAcquiredSkillAccess(ctx *gin.Context, skill *mcp_service.AcquiredSkill) error {
	if skill == nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_not_found", "acquired skill not found")
	}
	sourceSkillId := skill.GetSkill().GetSkill().GetSkillId()
	if sourceSkillId == "" {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_source_id_empty", "acquired skill source custom skill id is empty")
	}
	// 获取源 Skill 的发布信息
	appInfoMap, err := getSourceSkillAppInfoMap(ctx, []string{sourceSkillId})
	if err != nil {
		return err
	}
	appInfos := appInfoMap[sourceSkillId]
	if len(appInfos) == 0 || checkAccessibleSkillApp(appInfos, skill.GetUserId(), skill.GetOrgId()) == nil {
		log.Warnf("checkAcquiredSkillAccess: no accessible appInfo, sourceSkillId=%s", sourceSkillId)
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_no_permission", "no permission to access this skill")
	}
	return nil
}
