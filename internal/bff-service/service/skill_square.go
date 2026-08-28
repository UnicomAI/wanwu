package service

import (
	"strings"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

func GetSquareBuiltinSkillList(ctx *gin.Context, userId, orgId, name string) (*response.ListResult, error) {
	var skillsCfgList []*config.SkillsConfig
	for _, skillsCfg := range config.Cfg().AgentSkills {
		if name != "" && !strings.Contains(skillsCfg.Name, name) {
			continue
		}
		skillsCfgList = append(skillsCfgList, skillsCfg)
	}

	list := make([]*response.BuiltinSkillInfo, 0, len(skillsCfgList))
	var skillIds []string
	for _, skillsCfg := range skillsCfgList {
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

func GetSquareBuiltinSkillDetail(ctx *gin.Context, skillId string) (*response.BuiltinSkillDetail, error) {
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
	return detail, nil
}

func GetSquareShareSkillList(ctx *gin.Context, userId, orgId, name string) (*response.ListResult, error) {
	// 获取 skill 列表
	appResp, err := app.GetExplorationAppList(ctx.Request.Context(), &app_service.GetExplorationAppListReq{
		Name:       name,
		AppType:    constant.AppTypeSkill,
		SearchType: "all",
		UserId:     userId,
		OrgId:      orgId,
	})
	if err != nil {
		return nil, err
	}
	var skillIds []string
	for _, appInfo := range appResp.GetInfos() {
		if isSquareShareSkill(appInfo, userId, orgId) {
			skillIds = append(skillIds, appInfo.GetAppId())
		}
	}
	if len(skillIds) == 0 {
		return &response.ListResult{List: []*response.SharedSkillInfo{}, Total: 0}, nil
	}

	// 获取 skill 详情
	detailResp, err := mcp.GetCustomSkillDetailByIdList(ctx.Request.Context(), &mcp_service.CustomSkillDetailByIdListReq{
		SkillIds: skillIds,
	})
	if err != nil {
		return nil, err
	}

	// 获取 acquired 状态
	acquiredMap := make(map[string]bool)
	if len(skillIds) > 0 {
		acquiredResp, err := mcp.CheckAcquiredSkill(ctx.Request.Context(), &mcp_service.CheckAcquiredSkillReq{
			SkillIds: skillIds,
			Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
		})
		if err == nil && acquiredResp != nil {
			acquiredMap = acquiredResp.GetAcquiredMap()
		}
	}

	detailByID := make(map[string]*mcp_service.CustomSkill, len(detailResp.GetSkillDetails()))
	for _, skill := range detailResp.GetSkillDetails() {
		detailByID[skill.GetSkillId()] = skill
	}

	list := make([]*response.SharedSkillInfo, 0, len(skillIds))
	for _, skillId := range skillIds {
		skill := detailByID[skillId]
		if skill == nil {
			continue
		}
		if name != "" && !strings.Contains(skill.GetName(), name) {
			continue
		}
		list = append(list, toSharedSkillInfo(ctx, skill, acquiredMap[skillId]))
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

func ShareSquareSkill(ctx *gin.Context, userId, orgId, skillId string) error {
	_, err := mcp.AcquiredSkillCreate(ctx.Request.Context(), &mcp_service.AcquiredSkillCreateReq{
		Identity:      &mcp_service.Identity{UserId: userId, OrgId: orgId},
		CustomSkillId: skillId,
	})
	if err != nil {
		return err
	}
	// 递增添加计数
	incrementCustomSkillAcquired(ctx, skillId)
	return nil
}

func GetSquareShareSkillDetail(ctx *gin.Context, userId, orgId, skillId string) (*response.SharedSkillDetail, error) {
	// 获取最新版本的 skill markdown
	if skillId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "skillId is required")
	}
	publish, err := mcp.GetPublishCustomSkillByLatest(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillByLatestReq{
		SkillId: skillId,
	})
	if err != nil {
		return nil, err
	}
	if publish.GetMarkdown() == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "skill_publish_markdown_not_available", "latest skill markdown is empty")
	}
	skill := publish.GetSkill()
	if skill == nil || skill.GetSkillId() == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_custom_not_found")
	}

	// 获取 acquired 状态
	isAcquired := false
	acquiredResp, err := mcp.CheckAcquiredSkill(ctx.Request.Context(), &mcp_service.CheckAcquiredSkillReq{
		SkillIds: []string{skillId},
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err == nil && acquiredResp != nil {
		isAcquired = acquiredResp.GetAcquiredMap()[skillId]
	}

	return toSharedSkillDetail(ctx, publish, isAcquired), nil
}

func DownloadSquareShareSkill(ctx *gin.Context, skillId string) ([]byte, error) {
	if skillId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "skillId is required")
	}
	publish, err := mcp.GetPublishCustomSkillByLatest(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillByLatestReq{SkillId: skillId})
	if err != nil {
		return nil, err
	}
	if publish.GetObjectPath() == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "skill_publish_package_not_available", "latest skill package objectPath is empty")
	}
	// 递增下载计数
	incrementCustomSkillDownload(ctx, skillId)
	return downloadCustomSkillZip(ctx, publish.GetObjectPath())
}

// GetSquareShareSkillVersionList 获取共享skill版本列表
func GetSquareShareSkillVersionList(ctx *gin.Context, skillId string) (*response.ListResult, error) {
	if skillId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "skillId is required")
	}
	return GetSkillVersionList(ctx, skillId, "", "")
}

// GetSquareCreatedSkillList 获取我发布的skill列表（当前用户、当前组织下所有已发布 skill，含私密/组织内/全局）
func GetSquareCreatedSkillList(ctx *gin.Context, userId, orgId, name string) (*response.ListResult, error) {
	// 必须用 GetAppList（按 userId+orgId 精确过滤），不能用 GetExplorationAppList：
	// 探索列表 searchType=all 会包含「本人所有组织的 private/public」及他人公开，切换组织后仍会看到其他组织的发布。
	appResp, err := app.GetAppList(ctx.Request.Context(), &app_service.GetAppListReq{
		OrgIds:  []string{orgId},
		UserIds: []string{userId},
		AppType: constant.AppTypeSkill,
	})
	if err != nil {
		return nil, err
	}
	skillIds := make([]string, 0, len(appResp.GetInfos()))
	for _, appInfo := range appResp.GetInfos() {
		if appInfo == nil || appInfo.GetAppId() == "" || appInfo.GetPublishType() == "" {
			continue
		}
		skillIds = append(skillIds, appInfo.GetAppId())
	}
	if len(skillIds) == 0 {
		return &response.ListResult{List: []*response.PublishedSkillInfo{}, Total: 0}, nil
	}

	// 获取 skill 详情
	detailResp, err := mcp.GetCustomSkillDetailByIdList(ctx.Request.Context(), &mcp_service.CustomSkillDetailByIdListReq{
		SkillIds: skillIds,
	})
	if err != nil {
		return nil, err
	}
	publishTypeBySkill := make(map[string]string, len(appResp.GetInfos()))
	for _, appInfo := range appResp.GetInfos() {
		if appInfo != nil && appInfo.GetAppId() != "" {
			publishTypeBySkill[appInfo.GetAppId()] = appInfo.GetPublishType()
		}
	}

	// 获取每个 skill 的最新版本
	publishBySkill, err := getCustomSkillPublishMap(ctx, skillIds)
	if err != nil {
		log.Errorf("GetSquareCreatedSkillList version list failed, userId=%s orgId=%s err=%v", userId, orgId, err)
	}

	list := make([]*response.PublishedSkillInfo, 0, len(skillIds))
	for _, skill := range detailResp.GetSkillDetails() {
		if skill == nil {
			continue
		}
		if name != "" && !strings.Contains(skill.GetName(), name) {
			continue
		}
		item := toSquareCreatedSkillItem(ctx, skill, publishTypeBySkill[skill.GetSkillId()])
		if publish := publishBySkill[skill.GetSkillId()]; publish != nil {
			item.Version = publish.GetVersion()
		}
		list = append(list, item)
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

// GetSquareCreatedSkillDetail 获取我发布的skill详情
func GetSquareCreatedSkillDetail(ctx *gin.Context, userId, orgId, customSkillId string) (*response.PublishedSkillDetail, error) {
	if customSkillId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "customSkillId is required")
	}
	return GetCustomSkill(ctx, userId, orgId, customSkillId)
}

// GetSquareCreatedSkillVersionList 获取我发布skill版本列表
func GetSquareCreatedSkillVersionList(ctx *gin.Context, userId, orgId, customSkillId string) (*response.ListResult, error) {
	if customSkillId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "customSkillId is required")
	}
	return GetSkillVersionList(ctx, customSkillId, userId, orgId)
}

// --- internal ---

// isSquareShareSkill 判断 skill 是否应出现在广场「共享」列表。
// 包含：他人全局发布；他人在当前组织的组织内/全局发布；本人在其他组织的全局发布。
// 不包含：本人在当前组织的任意发布；私密发布。
func isSquareShareSkill(appInfo *app_service.ExplorationAppInfo, userId, orgId string) bool {
	if appInfo == nil {
		return false
	}
	publishType := appInfo.GetPublishType()
	if publishType == "" || publishType == constant.AppPublishPrivate {
		return false
	}
	isSelf := appInfo.GetUserId() == userId
	isCurrentOrg := appInfo.GetOrgId() == orgId
	if isSelf && isCurrentOrg {
		return false
	}
	if isSelf {
		return publishType == constant.AppPublishPublic
	}
	if publishType == constant.AppPublishPublic {
		return true
	}
	return publishType == constant.AppPublishOrganization && isCurrentOrg
}

func toSharedSkillDetail(ctx *gin.Context, publish *mcp_service.PublishCustomSkill, isAcquired bool) *response.SharedSkillDetail {
	skill := publish.GetSkill()
	return &response.SharedSkillDetail{
		SharedSkillInfo: response.SharedSkillInfo{
			SkillBasicInfo: response.SkillBasicInfo{
				SkillId: skill.GetSkillId(),
				Name:    skill.GetName(),
				Avatar:  cacheSkillAvatar(ctx, skill.GetAvatar()),
				Author:  skill.GetAuthor(),
				Desc:    skill.GetDesc(),
			},
			IsShared:      isAcquired,
			DownloadCount: skill.GetDownloadCount(),
			AcquiredCount: skill.GetAcquiredCount(),
		},
		SkillMarkdown: config.FixFrontMatterFormat(publish.GetMarkdown()),
	}
}

func toSharedSkillInfo(ctx *gin.Context, skill *mcp_service.CustomSkill, isAcquired bool) *response.SharedSkillInfo {
	return &response.SharedSkillInfo{
		SkillBasicInfo: response.SkillBasicInfo{
			SkillId: skill.GetSkillId(),
			Name:    skill.GetName(),
			Avatar:  cacheSkillAvatar(ctx, skill.GetAvatar()),
			Author:  skill.GetAuthor(),
			Desc:    skill.GetDesc(),
		},
		IsShared:      isAcquired,
		DownloadCount: skill.GetDownloadCount(),
		AcquiredCount: skill.GetAcquiredCount(),
	}
}

func toSquareCreatedSkillItem(ctx *gin.Context, skill *mcp_service.CustomSkill, publishType string) *response.PublishedSkillInfo {
	return &response.PublishedSkillInfo{
		SkillBasicInfo: response.SkillBasicInfo{
			SkillId: skill.GetSkillId(),
			Name:    skill.GetName(),
			Avatar:  cacheSkillAvatar(ctx, skill.GetAvatar()),
			Author:  skill.GetAuthor(),
			Desc:    skill.GetDesc(),
		},
		ThreadID:      skill.GetWgaThreadId(),
		PreviewID:     skill.GetPreviewThreadId(),
		IsPublished:   true,
		PublishType:   publishType,
		Version:       "",
		DownloadCount: skill.GetDownloadCount(),
		AcquiredCount: skill.GetAcquiredCount(),
	}
}
