package service

import (
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// GetAcquiredSkillList 资源库-我添加的skill列表
func GetAcquiredSkillList(ctx *gin.Context, userId, orgId, name string) (*response.ListResult, error) {
	resp, err := mcp.AcquiredSkillGetList(ctx.Request.Context(), &mcp_service.AcquiredSkillGetListReq{
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
		Name:     name,
	})
	if err != nil {
		return nil, err
	}

	list := make([]*response.AcquiredSkillInfo, 0, len(resp.List))
	for _, skill := range resp.List {
		list = append(list, toAcquiredSkillInfo(ctx, skill))
	}

	return &response.ListResult{
		List:  list,
		Total: resp.Total,
	}, nil
}

// DeleteAcquiredSkill 资源库-删除已添加的skill
func DeleteAcquiredSkill(ctx *gin.Context, userId, orgId, acquiredSkillId string) error {
	_, err := mcp.AcquiredSkillDelete(ctx.Request.Context(), &mcp_service.AcquiredSkillDeleteReq{
		AcquiredSkillId: acquiredSkillId,
	})
	return err
}

// GetAcquiredSkill 资源库-获取已添加skill详情
func GetAcquiredSkill(ctx *gin.Context, userId, orgId, acquiredSkillId string) (*response.AcquiredSkillDetail, error) {
	if acquiredSkillId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "acquiredSkillId is required")
	}
	skill, err := mcp.AcquiredSkillGet(ctx.Request.Context(), &mcp_service.AcquiredSkillGetReq{
		AcquiredSkillId: acquiredSkillId,
	})
	if err != nil {
		return nil, err
	}
	if err := checkAcquiredSkillAccess(ctx, skill); err != nil {
		return nil, err
	}
	return toAcquiredSkillDetail(ctx, skill, true), nil
}

func GetCallbackAcquiredSkillListDetail(ctx *gin.Context, acquiredSkillIdList []string) (*response.CallbackAcquiredSkillDetailListResp, error) {
	if len(acquiredSkillIdList) == 0 {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "skillIdList is required")
	}
	filteredAcquiredSkillIdList := make([]string, 0, len(acquiredSkillIdList))
	for _, acquiredSkillId := range acquiredSkillIdList {
		if acquiredSkillId == "" {
			continue
		}
		filteredAcquiredSkillIdList = append(filteredAcquiredSkillIdList, acquiredSkillId)
	}
	if len(filteredAcquiredSkillIdList) == 0 {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFInvalidArg, "skillIdList is required")
	}

	skillByID, err := getAcquiredSkillByIDMap(ctx, filteredAcquiredSkillIdList)
	if err != nil {
		return nil, err
	}

	// 批量获取源 Skill 的发布信息用于权限校验
	sourceSkillIds := make([]string, 0, len(skillByID))
	for _, skill := range skillByID {
		if skill == nil {
			continue
		}
		if sourceSkillId := skill.GetSkill().GetSkill().GetSkillId(); sourceSkillId != "" {
			sourceSkillIds = append(sourceSkillIds, sourceSkillId)
		}
	}
	appInfoMap, err := getSourceSkillAppInfoMap(ctx, sourceSkillIds)
	if err != nil {
		return nil, err
	}

	skillDetailList := make([]*response.CallbackAcquiredSkillDetail, 0, len(filteredAcquiredSkillIdList))
	for _, acquiredSkillId := range filteredAcquiredSkillIdList {
		skill := skillByID[acquiredSkillId]
		if skill == nil {
			log.Warnf("callback acquired skill list ignored missing acquired skill, acquiredSkillId: %s", acquiredSkillId)
			continue
		}
		// 过滤无权限访问的 acquired skill
		if !isAcquiredSkillAccessible(ctx, skill, appInfoMap) {
			log.Warnf("callback acquired skill list ignored inaccessible acquired skill, acquiredSkillId: %s", acquiredSkillId)
			continue
		}
		detail, err := toCallbackAcquiredSkillDetail(ctx, skill)
		if err != nil {
			log.Warnf("callback acquired skill list ignored invalid acquired skill, acquiredSkillId: %s, err: %v", acquiredSkillId, err)
			continue
		}
		skillDetailList = append(skillDetailList, detail)
	}

	return &response.CallbackAcquiredSkillDetailListResp{SkillList: skillDetailList}, nil
}

func DownloadAcquiredSkill(ctx *gin.Context, userId, orgId, acquiredSkillId string) ([]byte, error) {
	skill, err := mcp.AcquiredSkillGet(ctx.Request.Context(), &mcp_service.AcquiredSkillGetReq{
		AcquiredSkillId: acquiredSkillId,
	})
	if err != nil {
		return nil, err
	}
	if err := checkAcquiredSkillAccess(ctx, skill); err != nil {
		return nil, err
	}
	// 闭源 skill：禁止下载 ZIP
	if util.IsSkillClosedSource(skill.GetSkill().GetMarkdown()) {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "skill_closed_source_download_forbidden", "该 skill 为闭源，不支持下载")
	}
	if skill.GetSkill().GetObjectPath() == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "skill_publish_package_not_available", "acquired skill package objectPath is empty")
	}
	// 递增下载计数
	sourceSkillId := skill.GetSkill().GetSkill().GetSkillId()
	if sourceSkillId != "" {
		incrementCustomSkillDownload(ctx, sourceSkillId)
	}
	return downloadCustomSkillZip(ctx, skill.GetSkill().GetObjectPath())
}

func GetAcquiredSkillVersionList(ctx *gin.Context, userId, orgId, acquiredSkillId string) (*response.ListResult, error) {
	skill, err := mcp.AcquiredSkillGet(ctx.Request.Context(), &mcp_service.AcquiredSkillGetReq{
		AcquiredSkillId: acquiredSkillId,
	})
	if err != nil {
		return nil, err
	}
	if err := checkAcquiredSkillAccess(ctx, skill); err != nil {
		return nil, err
	}
	sourceSkillId := skill.GetSkill().GetSkill().GetSkillId()
	if sourceSkillId == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_source_id_empty", "acquired skill source custom skill id is empty")
	}
	return GetSkillVersionList(ctx, sourceSkillId, userId, orgId)
}

// GetSkillVersionList 获取 skill 版本列表
func GetSkillVersionList(ctx *gin.Context, skillId, userId, orgId string) (*response.ListResult, error) {
	resp, err := mcp.GetPublishCustomSkillHistoryList(ctx.Request.Context(), &mcp_service.GetPublishCustomSkillHistoryListReq{
		SkillId: skillId,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*response.SkillVersionInfo, 0, len(resp.GetHistoryList()))
	for _, item := range resp.GetHistoryList() {
		list = append(list, &response.SkillVersionInfo{
			Version:   item.GetVersion(),
			Desc:      item.GetVersionDesc(),
			UpdatedAt: util.Time2Str(item.GetCreatedAt()),
		})
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

// --- internal ---

func getAcquiredSkillByIDMap(ctx *gin.Context, acquiredSkillIdList []string) (map[string]*mcp_service.AcquiredSkill, error) {
	filteredAcquiredSkillIdList := make([]string, 0, len(acquiredSkillIdList))
	for _, acquiredSkillId := range acquiredSkillIdList {
		if acquiredSkillId == "" {
			continue
		}
		filteredAcquiredSkillIdList = append(filteredAcquiredSkillIdList, acquiredSkillId)
	}

	ret := make(map[string]*mcp_service.AcquiredSkill, len(filteredAcquiredSkillIdList))
	if len(filteredAcquiredSkillIdList) == 0 {
		return ret, nil
	}

	resp, err := mcp.AcquiredSkillGetByIDList(ctx.Request.Context(), &mcp_service.AcquiredSkillGetByIDListReq{
		AcquiredSkillIdList: uniqueSkillIDs(filteredAcquiredSkillIdList),
	})
	if err != nil {
		return nil, err
	}

	for _, skill := range resp.GetList() {
		if skill == nil || skill.GetAcquiredSkillId() == "" {
			continue
		}
		ret[skill.GetAcquiredSkillId()] = skill
	}
	return ret, nil
}

func toAcquiredSkillDetail(ctx *gin.Context, skill *mcp_service.AcquiredSkill, includeVariables bool) *response.AcquiredSkillDetail {
	if skill == nil {
		return nil
	}
	publish := skill.GetSkill()
	customSkill := publish.GetSkill()
	markdown := config.FixFrontMatterFormat(publish.GetMarkdown())
	ret := &response.AcquiredSkillDetail{
		AcquiredSkillInfo: response.AcquiredSkillInfo{
			SkillBasicInfo: response.SkillBasicInfo{
				SkillId: skill.GetAcquiredSkillId(),
				Name:    customSkill.GetName(),
				Avatar:  cacheSkillAvatar(ctx, customSkill.GetAvatar()),
				Author:  customSkill.GetAuthor(),
				Desc:    customSkill.GetDesc(),
			},
			DownloadCount: customSkill.GetDownloadCount(),
		},
		SkillMarkdown: markdown,
	}
	// 闭源 skill：隐藏正文（保留 frontmatter 以便前端展示 name/description）
	if util.IsSkillClosedSource(publish.GetMarkdown()) {
		ret.IsClosedSource = true
		ret.SkillMarkdown = stripSkillBody(markdown)
	}
	if includeVariables {
		variables, err := getAcquiredSkillVariables(ctx, skill.GetAcquiredSkillId())
		if err == nil {
			ret.Variables = toSkillVariables(variables)
		}
	}
	return ret
}

func toCallbackAcquiredSkillDetail(ctx *gin.Context, skill *mcp_service.AcquiredSkill) (*response.CallbackAcquiredSkillDetail, error) {
	if skill == nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_acquired_not_found", "acquired skill not found")
	}
	publish := skill.GetSkill()
	customSkill := publish.GetSkill()
	if customSkill == nil || customSkill.GetSkillId() == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "skill_not_published", "custom skill is not published")
	}
	if publish.GetObjectPath() == "" {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "skill_publish_package_not_available", "latest skill package objectPath is empty")
	}
	return &response.CallbackAcquiredSkillDetail{
		SkillBasicInfo: response.SkillBasicInfo{
			SkillId: skill.GetAcquiredSkillId(),
			Name:    customSkill.GetName(),
			Avatar:  cacheSkillAvatar(ctx, customSkill.GetAvatar()),
			Author:  customSkill.GetAuthor(),
			Desc:    customSkill.GetDesc(),
		},
		ObjectPath: publish.GetObjectPath(),
	}, nil
}

func toAcquiredSkillInfo(ctx *gin.Context, skill *mcp_service.AcquiredSkill) *response.AcquiredSkillInfo {
	if skill == nil {
		return nil
	}
	customSkill := skill.GetSkill().GetSkill()
	return &response.AcquiredSkillInfo{
		SkillBasicInfo: response.SkillBasicInfo{
			SkillId: skill.GetAcquiredSkillId(),
			Name:    customSkill.GetName(),
			Avatar:  cacheSkillAvatar(ctx, customSkill.GetAvatar()),
			Author:  customSkill.GetAuthor(),
			Desc:    customSkill.GetDesc(),
		},
		DownloadCount:  customSkill.GetDownloadCount(),
		IsClosedSource: util.IsSkillClosedSource(skill.GetSkill().GetMarkdown()),
	}
}
