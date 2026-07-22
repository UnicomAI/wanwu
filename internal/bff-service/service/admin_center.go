package service

import (
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/gin-gonic/gin"
)

const (
	SkillTypeBuiltin = "builtin"
)

// AdminKnowledgePageList 管理员中心知识库全局分页列表（跨用户，按系统权限过滤）
func AdminKnowledgePageList(ctx *gin.Context, req *request.AdminKnowledgePageListReq) (*response.PageResult, error) {
	pageNo, pageSize := normalizePage(req.PageNo, req.PageSize)
	resp, err := knowledgeBase.AdminKnowledgePageList(ctx.Request.Context(), &knowledgebase_service.AdminKnowledgePageListReq{
		Name:     req.Name,
		Category: req.Category,
		External: req.External,
		UserId:   req.UserIdList,
		OrgId:    req.OrgIdList,
		PageNum:  int32(pageNo),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return nil, err
	}
	list := buildAdminKnowledgeList(ctx, resp.KnowledgeList)
	//填充用户的名称和组织名称
	fillOwnerList(ctx, list)
	return &response.PageResult{
		List:     list,
		Total:    resp.Total,
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// AdminSelectKnowledgeBase 管理员中心根据 knowledgeId 查询知识库详情（跨用户，不做权限收窄）
func AdminSelectKnowledgeBase(ctx *gin.Context, req *request.AdminKnowledgeDetailReq) (*response.AdminKnowledgeBase, error) {
	docKnowledgeInfo, err := GetDocKnowledgeDetail(ctx, "", "", &request.DocKnowledgeDetailReq{
		KnowledgeId: req.KnowledgeId,
	}, &DocKnowledgeParams{
		NeedOwner: true,
	})
	if err != nil {
		return nil, err
	}
	return buildAdminKnowledgeBase(ctx, docKnowledgeInfo), nil
}

func AdminKnowledgeFileDetail(ctx *gin.Context, req *request.AdminKnowledgeFileDetailReq) (*response.DocSegmentResp, error) {
	return GetDocSegmentList(ctx, "", "", &request.DocSegmentListReq{
		DocId: req.DocId,
		PageSearch: request.PageSearch{
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		},
	})
}

// AdminWorkflowPageList 工作流列表
func AdminWorkflowPageList(ctx *gin.Context, req *request.AdminWorkflowPageListReq) (*response.PageResult, error) {
	return &response.PageResult{}, nil
}

// AdminSkillPageList skill分页列表（跨用户，按 userId[]/orgId[]/name 过滤后分页）
func AdminSkillPageList(ctx *gin.Context, req *request.AdminSkillPageListReq) (*response.PageResult, error) {
	pageNo, pageSize := normalizePage(req.PageNo, req.PageSize)
	resp, err := mcp.AdminCustomSkillPageList(ctx.Request.Context(), &mcp_service.AdminCustomSkillPageListReq{
		Name:     req.Name,
		UserId:   req.UserIdList,
		OrgId:    req.OrgIdList,
		PageNum:  int32(pageNo),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return nil, err
	}

	return &response.PageResult{
		List:     buildSkillList(ctx, resp.List),
		Total:    resp.Total,
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// AdminSkillBase skill基础信息
func AdminSkillBase(ctx *gin.Context, req *request.AdminSkillDetailReq) (*response.AdminAppBaseInfo, error) {
	skillPublish, err := GetCustomSkill(ctx, "", "", req.SkillId)
	if err != nil {
		return nil, err
	}
	var publishStatus = "draft"
	if skillPublish.IsPublished {
		publishStatus = "publish"
	}
	retBaseInfo := &response.AdminAppBaseInfo{
		Avatar:        skillPublish.Avatar,
		UpdatedAt:     skillPublish.UpdatedAt,
		CreatedAt:     skillPublish.CreatedAt,
		Desc:          skillPublish.Desc,
		Name:          skillPublish.Name,
		PublishStatus: publishStatus,
		PublishScope:  skillPublish.PublishType,
		OwnerHolder:   response.CreateOwnerHolder(skillPublish.UserId, skillPublish.OrgId),
	}
	//填充用户的名称和组织名称
	fillOwner(ctx, retBaseInfo)
	return retBaseInfo, nil
}

// AdminSkillDetail skill详情
func AdminSkillDetail(ctx *gin.Context, req *request.AdminSkillDetailReq) (*response.PublishedSkillDetail, error) {
	if req.SkillType == SkillTypeBuiltin {
		detail, err := GetSquareBuiltinSkillDetail(ctx, req.SkillId)
		if err != nil {
			return nil, err
		}
		return buildAdminBuiltinSkillDetail(detail), nil
	} else {
		skill, err := GetCustomSkill(ctx, "", "", req.SkillId)
		if err != nil {
			return nil, err
		}
		if !skill.IsPublished {
			skill.SkillMarkdown = fillSkillMarkdown(ctx, req.SkillId, skill.Name)
		}

		return skill, nil
	}
}

// AdminSkillVersionList skill详情
func AdminSkillVersionList(ctx *gin.Context, req *request.AdminSkillDetailReq) (*response.ListResult, error) {
	return GetSquareCreatedSkillVersionList(ctx, "", "", req.SkillId)
}

// normalizePage 规范化分页参数：pageNo<1 置 1，pageSize<=0 置默认 10，避免下游负 offset
func normalizePage(pageNo, pageSize int) (int, int) {
	if pageNo < 1 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return pageNo, pageSize
}

func buildAdminKnowledgeBase(ctx *gin.Context, docKnowledgeInfo *response.DocKnowledgeInfo) *response.AdminKnowledgeBase {
	var retKnowledgeInfo = &response.AdminKnowledgeBase{
		KnowledgeId:    docKnowledgeInfo.KnowledgeId,
		Name:           docKnowledgeInfo.KnowledgeName,
		Keywords:       docKnowledgeInfo.Keywords,
		Description:    docKnowledgeInfo.Description,
		EmbeddingModel: docKnowledgeInfo.EmbeddingModel,
		Category:       docKnowledgeInfo.Category,
		Avatar:         docKnowledgeInfo.Avatar,
		LlmModelId:     docKnowledgeInfo.LlmModelId,
		GraphSwitch:    docKnowledgeInfo.GraphSwitch,
		AdminAppBaseInfo: response.AdminAppBaseInfo{
			CreatedAt:   docKnowledgeInfo.CreatedAt,
			UpdatedAt:   docKnowledgeInfo.UpdatedAt,
			OwnerHolder: response.CreateOwnerHolder(docKnowledgeInfo.OwnerUserId, docKnowledgeInfo.OwnerOrgId),
		},
	}

	//填充用户的名称和组织名称
	fillOwner(ctx, retKnowledgeInfo)
	return retKnowledgeInfo
}

func buildAdminKnowledgeList(ctx *gin.Context, list []*knowledgebase_service.KnowledgeInfo) []*response.AdminKnowledge {
	retList := make([]*response.AdminKnowledge, 0, len(list))
	for _, k := range list {
		retList = append(retList, &response.AdminKnowledge{
			KnowledgeId: k.KnowledgeId,
			Name:        k.Name,
			Description: k.Description,
			Category:    k.Category,
			External:    k.External,
			Avatar:      cacheKnowledgeAvatar(ctx, k.AvatarPath, k.Category),
			OwnerHolder: response.CreateOwnerHolder(k.OwnerUserId, k.OwnerOrgId),
			CreatedAt:   k.CreatedAt,
			UpdatedAt:   k.UpdatedAt,
		})
	}
	return retList
}

// buildSkillList skill列表
func buildSkillList(ctx *gin.Context, list []*mcp_service.PublishCustomSkill) []*response.AdminSkillDetail {
	var userIdMap = make(map[string]bool)
	var orgIdMap = make(map[string]bool)
	skillList := make([]*response.PublishedSkillInfo, 0, len(list))
	for _, item := range list {
		skillList = append(skillList, toCustomSkillListItem(ctx, item))
		userIdMap[item.Skill.UserId] = true
		orgIdMap[item.Skill.OrgId] = true
	}
	//填充发布信息
	fillCustomSkillPublishInfo(ctx, skillList)
	//构造结果数据
	adminSkillList := fillAdminSkillList(skillList)
	//填充用户和组织名称
	fillOwnerList(ctx, adminSkillList)
	return adminSkillList
}

// fillAdminSkillList 填充用户和组织名称
func fillAdminSkillList(skillList []*response.PublishedSkillInfo) []*response.AdminSkillDetail {
	var retSkillList []*response.AdminSkillDetail
	for _, skill := range skillList {
		retSkill := &response.AdminSkillDetail{
			PublishedSkillInfo: *skill,
			OwnerHolder:        response.CreateOwnerHolder(skill.UserId, skill.OrgId),
		}
		retSkillList = append(retSkillList, retSkill)
	}
	return retSkillList
}

// fillSkillMarkdown 填充技能的markdown
func fillSkillMarkdown(ctx *gin.Context, skillId, name string) string {
	file, err := GetSkillWorkspaceFile(ctx, "", "", request.GetSkillWorkspaceFileReq{
		CustomSkillID: skillId,
		Path:          name + "/SKILL.md",
	})
	if err != nil {
		return ""
	}
	return file.Content
}

// buildAdminBuiltinSkillDetail 内置skill详情转换为管理员中心已发布skill详情
func buildAdminBuiltinSkillDetail(detail *response.BuiltinSkillDetail) *response.PublishedSkillDetail {
	return &response.PublishedSkillDetail{
		PublishedSkillInfo: response.PublishedSkillInfo{
			SkillBasicInfo: detail.SkillBasicInfo,
			DownloadCount:  detail.DownloadCount,
		},
		Variables:     detail.Variables,
		SkillMarkdown: detail.SkillMarkdown,
	}
}
