package service

import (
	"slices"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	SkillTypeBuiltin = "builtin"

	adminRagFilterFetchPageNum  = 1
	adminRagFilterFetchPageSize = 2000
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

func AdminModelPageList(ctx *gin.Context, req *request.AdminModelPageListReq) (*response.PageResult, error) {
	pageNo, pageSize := normalizePage(req.PageNo, req.PageSize)
	resp, err := model.AdminModelPageList(ctx.Request.Context(), &model_service.AdminModelPageListReq{
		Name:      req.Name,
		ModelType: req.ModelType,
		Provider:  req.Provider,
		UserId:    req.UserIdList,
		OrgId:     req.OrgIdList,
		ScopeType: publishScopeToModelScopeTypes(req.PublishScope),
		PageNum:   int32(pageNo),
		PageSize:  int32(pageSize),
	})
	if err != nil {
		return nil, err
	}
	list, err := buildAdminModelList(ctx, resp.Models)
	if err != nil {
		return nil, err
	}
	return &response.PageResult{
		List:     list,
		Total:    resp.Total,
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// AdminModelBase 管理员中心模型基础信息（模型无草稿概念，创建即发布；发布范围由 scopeType 映射，无特定用户列表）
func AdminModelBase(ctx *gin.Context, req *request.AdminModelDetailReq) (*response.AdminModelBase, error) {
	modelInfo, err := GetModelById(ctx, &request.GetModelRequest{
		BaseModelRequest: request.BaseModelRequest{ModelId: req.ModelId},
	})
	if err != nil {
		return nil, err
	}
	base := &response.AdminModelBase{
		AdminAppBaseInfo: response.AdminAppBaseInfo{
			Avatar:        modelInfo.Avatar,
			Name:          modelInfo.DisplayName,
			Desc:          modelInfo.ModelDesc,
			CreatedAt:     modelInfo.CreatedAt,
			UpdatedAt:     modelInfo.UpdatedAt,
			PublishStatus: "publish", // 模型没有发布概念，创建=发布
			PublishScope:  modelScopeToPublishScope(modelInfo.ScopeType),
			OwnerHolder:   response.CreateOwnerHolder(modelInfo.UserId, modelInfo.OrgId),
		},
		Provider: modelInfo.Provider,
	}
	fillOwner(ctx, &base.AdminAppBaseInfo)
	return base, nil
}

// buildAdminModelList 将模型 proto 列表转换为管理员列表项，并补全拥有者用户/组织名称
func buildAdminModelList(ctx *gin.Context, models []*model_service.ModelInfo) ([]*response.AdminModel, error) {
	modelInfos, err := toModelInfos(ctx, models)
	if err != nil {
		return nil, err
	}
	retList := make([]*response.AdminModel, 0, len(modelInfos))
	for _, m := range modelInfos {
		retList = append(retList, &response.AdminModel{
			ModelInfo:   *m,
			OwnerHolder: response.CreateOwnerHolder(m.UserId, m.OrgId),
		})
	}
	fillOwnerList(ctx, retList)
	return retList, nil
}

// modelScopeToPublishScope 将模型 scopeType(1私有/2公开/3组织) 映射为发布范围字符串
func modelScopeToPublishScope(scopeType string) string {
	switch scopeType {
	case "1":
		return constant.AppPublishPrivate
	case "2":
		return constant.AppPublishPublic
	case "3":
		return constant.AppPublishOrganization
	default:
		return ""
	}
}

// publishScopeToModelScopeTypes 将发布范围字符串映射为模型 scopeType(1私有/2公开/3组织)；不适用于模型的范围（如特定用户）被忽略
func publishScopeToModelScopeTypes(scopes []string) []uint32 {
	var ret []uint32
	for _, s := range scopes {
		switch s {
		case constant.AppPublishPrivate:
			ret = append(ret, 1)
		case constant.AppPublishPublic:
			ret = append(ret, 2)
		case constant.AppPublishOrganization:
			ret = append(ret, 3)
		}
	}
	return ret
}

// AdminRagPageList 管理员中心知识问答全局分页列表（跨用户，按 userId[]/orgId[]/name + 发布状态/范围）
func AdminRagPageList(ctx *gin.Context, req *request.AdminRagPageListReq) (*response.PageResult, error) {
	pageNo, pageSize := normalizePage(req.PageNo, req.PageSize)
	// 不进行发布筛选
	if len(req.PublishStatus) == 0 && len(req.PublishScope) == 0 {
		return adminRagPageByDB(ctx, req, pageNo, pageSize)
	}
	// 全量筛选
	return adminRagPageByFilter(ctx, req, pageNo, pageSize)
}

func adminRagPageByDB(ctx *gin.Context, req *request.AdminRagPageListReq, pageNo, pageSize int) (*response.PageResult, error) {
	resp, err := rag.AdminRagPageList(ctx.Request.Context(), &rag_service.AdminRagPageListReq{
		Name:     req.Name,
		UserId:   req.UserIdList,
		OrgId:    req.OrgIdList,
		PageNum:  int32(pageNo),
		PageSize: int32(pageSize),
	})
	if err != nil {
		return nil, err
	}
	list, err := toAdminRagList(ctx, resp.RagInfos)
	if err != nil {
		return nil, err
	}
	return &response.PageResult{List: list, Total: resp.Total, PageNo: pageNo, PageSize: pageSize}, nil
}

func adminRagPageByFilter(ctx *gin.Context, req *request.AdminRagPageListReq, pageNo, pageSize int) (*response.PageResult, error) {
	resp, err := rag.AdminRagPageList(ctx.Request.Context(), &rag_service.AdminRagPageListReq{
		Name:     req.Name,
		UserId:   req.UserIdList,
		OrgId:    req.OrgIdList,
		PageNum:  adminRagFilterFetchPageNum,
		PageSize: adminRagFilterFetchPageSize,
	})
	if err != nil {
		return nil, err
	}
	allList, err := toAdminRagList(ctx, resp.RagInfos)
	if err != nil {
		return nil, err
	}
	// 按发布状态/范围过滤
	matched := make([]*response.AdminRag, 0, len(allList))
	for _, item := range allList {
		if matchPublishFilter(item.PublishType, req.PublishStatus, req.PublishScope) {
			matched = append(matched, item)
		}
	}
	return &response.PageResult{
		List:     util.PageSlice(matched, pageNo, pageSize),
		Total:    int64(len(matched)),
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// toAdminRagList 把 rag 数据转成列表项，并补上发布状态/范围、版本号、拥有者的用户名和组织名。
func toAdminRagList(ctx *gin.Context, briefs []*common.AppBrief) ([]*response.AdminRag, error) {
	ownerMap := make(map[string]response.OwnerInfo, len(briefs))
	apps := make([]response.AppBriefInfo, 0, len(briefs))
	for _, brief := range briefs {
		apps = append(apps, appBriefProto2Model(ctx, brief, 0))
		ownerMap[brief.AppId] = response.OwnerInfo{OwnerUserId: brief.UserId, OwnerOrgId: brief.OrgId}
	}
	listResult, err := fillAppPublishInfo(ctx, "", "", apps)
	if err != nil {
		return nil, err
	}
	publishedApps := listResult.List.([]response.AppBriefInfo)
	ret := make([]*response.AdminRag, 0, len(publishedApps))
	for _, appInfo := range publishedApps {
		owner := ownerMap[appInfo.AppId]
		ret = append(ret, &response.AdminRag{
			AppBriefInfo: appInfo,
			OwnerHolder:  response.CreateOwnerHolder(owner.OwnerUserId, owner.OwnerOrgId),
		})
	}
	fillOwnerList(ctx, ret)
	return ret, nil
}

// publishStatusOf 有发布范围(publishType)即已发布，否则为草稿
func publishStatusOf(publishType string) string {
	if publishType != "" {
		return "publish"
	}
	return "draft"
}

// matchPublishFilter 判断某发布类型是否满足发布状态(draft/publish)与发布范围(=PublishType)过滤；空过滤视为全部通过
func matchPublishFilter(publishType string, statusFilter, scopeFilter []string) bool {
	if len(statusFilter) > 0 {
		if !slices.Contains(statusFilter, publishStatusOf(publishType)) {
			return false
		}
	}
	// 草稿无发布范围，指定发布范围时一律排除
	if len(scopeFilter) > 0 && (publishType == "" || !slices.Contains(scopeFilter, publishType)) {
		return false
	}
	return true
}

// AdminRagBase 管理员中心知识问答基础信息（拥有者取自 rag 归属，发布范围取自 app 发布类型）
func AdminRagBase(ctx *gin.Context, req *request.AdminRagDetailReq) (*response.AdminRagBase, error) {
	resp, err := rag.GetRagDetail(ctx.Request.Context(), &rag_service.RagDetailReq{RagId: req.RagId})
	if err != nil {
		return nil, err
	}
	appInfo, _ := app.GetAppInfo(ctx.Request.Context(), &app_service.GetAppInfoReq{AppId: req.RagId, AppType: constant.AppTypeRag})
	publishType := appInfo.GetPublishType()
	base := &response.AdminRagBase{
		AdminAppBaseInfo: response.AdminAppBaseInfo{
			Avatar:        cacheAppAvatar(ctx, resp.BriefConfig.AvatarPath, constant.AppTypeRag),
			Name:          resp.BriefConfig.Name,
			Desc:          resp.BriefConfig.Desc,
			CreatedAt:     util.Time2Str(resp.CreateTime),
			UpdatedAt:     util.Time2Str(resp.UpdateTime),
			PublishStatus: publishStatusOf(publishType),
			PublishScope:  publishType,
		},
	}
	if resp.Identity != nil {
		base.OwnerHolder = response.CreateOwnerHolder(resp.Identity.UserId, resp.Identity.OrgId)
		fillOwner(ctx, &base.AdminAppBaseInfo)
	}
	return base, nil
}

// AdminRagDetail 管理员中心知识问答详情：在全量 RagInfo 基础上，补全 llm/知识库Rerank/问答库Rerank 三个模型的头像与标签
func AdminRagDetail(ctx *gin.Context, req *request.AdminRagDetailReq) (*response.AdminRagDetail, error) {
	ragInfo, err := GetRag(ctx, request.RagReq{RagID: req.RagId}, false)
	if err != nil {
		return nil, err
	}
	return &response.AdminRagDetail{
		RagInfo:       *ragInfo,
		LlmModel:      adminModelInfoByID(ctx, ragInfo.ModelConfig.ModelId),
		RerankModel:   adminModelInfoByID(ctx, ragInfo.RerankConfig.ModelId),
		QaRerankModel: adminModelInfoByID(ctx, ragInfo.QARerankConfig.ModelId),
	}, nil
}

// adminModelInfoByID 按 modelId 查询模型完整信息（含头像、标签）；
// modelId 为空返回 nil，查询失败仅记日志不阻断详情返回
func adminModelInfoByID(ctx *gin.Context, modelId string) *response.ModelInfo {
	if modelId == "" {
		return nil
	}
	modelInfo, err := GetModelById(ctx, &request.GetModelRequest{
		BaseModelRequest: request.BaseModelRequest{ModelId: modelId},
	})
	if err != nil {
		log.Warnf("admin rag detail get model %v info err: %v", modelId, err)
		return nil
	}
	return modelInfo
}

// AdminMCPPageList 管理员中心MCP全局列表（单类型，按 Type 路由：mcp=导入MCP，mcpserver=创建MCP）。
func AdminMCPPageList(ctx *gin.Context, req request.AdminMCPPageListReq) (*response.PageResult, error) {
	pageNo, pageSize := normalizePage(req.PageNo, req.PageSize)
	// 按类型路由：mcpserver 查 MCP服务，mcp 查自定义MCP
	if slices.Contains(req.Type, constant.MCPTypeMCP) {
		return adminCustomMCPPageList(ctx, req, pageNo, pageSize)
	}
	if slices.Contains(req.Type, constant.MCPTypeMCPServer) {
		return adminMCPServerPageList(ctx, req, pageNo, pageSize)
	}
	return nil, nil
}

// adminCustomMCPPageList 自定义MCP（导入MCP）列表
func adminCustomMCPPageList(ctx *gin.Context, req request.AdminMCPPageListReq, pageNo, pageSize int) (*response.PageResult, error) {
	resp, err := mcp.GetAdminCustomMCPPageList(ctx.Request.Context(), &mcp_service.GetAdminCustomMCPPageListReq{
		Name:       req.Name,
		OrgIdList:  req.OrgIdList,
		UserIdList: req.UserIdList,
		PageNo:     int32(pageNo),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return nil, err
	}

	list := make([]*response.AdminMCP, 0, len(resp.List))
	for _, item := range resp.List {
		mcpItem := &response.AdminMCP{
			Avatar:      cacheMCPAvatar(ctx, "", item.AvatarPath),
			MCPID:       item.McpId,
			Type:        constant.MCPTypeMCP,
			Name:        item.Name,
			Description: item.Desc,
			ServerFrom:  item.From,
			UpdatedAt:   util.Time2Str(item.UpdatedAt),
			OwnerHolder: response.CreateOwnerHolder(item.Owner.UserId, item.Owner.OrgId),
		}
		list = append(list, mcpItem)
	}
	fillOwnerList(ctx, list)
	return &response.PageResult{
		List:     list,
		Total:    resp.Total,
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// adminMCPServerPageList MCP服务（创建MCP）列表
func adminMCPServerPageList(ctx *gin.Context, req request.AdminMCPPageListReq, pageNo, pageSize int) (*response.PageResult, error) {
	resp, err := mcp.GetAdminMCPServerPageList(ctx.Request.Context(), &mcp_service.GetAdminMCPServerPageListReq{
		Name:       req.Name,
		OrgIdList:  req.OrgIdList,
		UserIdList: req.UserIdList,
		PageNo:     int32(pageNo),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return nil, err
	}

	list := make([]*response.AdminMCP, 0, len(resp.List))
	for _, item := range resp.List {
		mcpItem := &response.AdminMCP{
			Avatar:      cacheMCPServerAvatar(ctx, item.AvatarPath),
			MCPID:       item.McpServerId,
			Type:        constant.MCPTypeMCPServer,
			Name:        item.Name,
			Description: item.Desc,
			UpdatedAt:   util.Time2Str(item.UpdatedAt),
			OwnerHolder: response.CreateOwnerHolder(item.Owner.UserId, item.Owner.OrgId),
		}
		list = append(list, mcpItem)
	}
	fillOwnerList(ctx, list)
	return &response.PageResult{
		List:     list,
		Total:    resp.Total,
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// AdminMCPBase 管理员中心MCP详情基础信息（按 Type 路由：mcp=导入MCP，mcpserver=创建MCP）。
func AdminMCPBase(ctx *gin.Context, req request.AdminMCPBaseReq) (*response.AdminMCPBase, error) {
	switch req.Type {
	case constant.MCPTypeMCP:
		return adminCustomMCPBase(ctx, req)
	case constant.MCPTypeMCPServer:
		return adminMCPServerBase(ctx, req)
	}
	return nil, nil
}

// adminCustomMCPBase 导入MCP（自定义MCP）基础信息
func adminCustomMCPBase(ctx *gin.Context, req request.AdminMCPBaseReq) (*response.AdminMCPBase, error) {
	detail, err := mcp.GetCustomMCP(ctx.Request.Context(), &mcp_service.GetCustomMCPReq{
		McpId: req.McpId,
	})
	if err != nil {
		return nil, err
	}
	resp := &response.AdminMCPBase{
		AdminAppBaseInfo: response.AdminAppBaseInfo{
			Avatar:      cacheMCPAvatar(ctx, "", detail.AvatarPath),
			Name:        detail.Info.Name,
			Desc:        detail.Info.Desc,
			CreatedAt:   util.Time2Str(detail.CreatedAt),
			UpdatedAt:   util.Time2Str(detail.UpdatedAt),
			OwnerHolder: response.CreateOwnerHolder(detail.Owner.UserId, detail.Owner.OrgId),
		},
		Type: constant.MCPTypeMCP,
	}
	fillOwner(ctx, resp)
	return resp, nil
}

// adminMCPServerBase 创建MCP（MCP服务）基础信息
func adminMCPServerBase(ctx *gin.Context, req request.AdminMCPBaseReq) (*response.AdminMCPBase, error) {
	info, err := mcp.GetMCPServer(ctx.Request.Context(), &mcp_service.GetMCPServerReq{
		McpServerId: req.McpId,
	})
	if err != nil {
		return nil, err
	}
	resp := &response.AdminMCPBase{
		AdminAppBaseInfo: response.AdminAppBaseInfo{
			Avatar:      cacheMCPServerAvatar(ctx, info.AvatarPath),
			Name:        info.Name,
			Desc:        info.Desc,
			CreatedAt:   util.Time2Str(info.CreatedAt),
			UpdatedAt:   util.Time2Str(info.UpdatedAt),
			OwnerHolder: response.CreateOwnerHolder(info.Owner.UserId, info.Owner.OrgId),
		},
		Type: constant.MCPTypeMCPServer,
	}
	fillOwner(ctx, resp)
	return resp, nil
}

// AdminMCPDetail 管理员中心导入MCP（自定义MCP）详情。
func AdminMCPDetail(ctx *gin.Context, req request.AdminMCPDetailReq) (*response.MCPDetail, error) {
	return GetMCP(ctx, req.McpId)
}

// AdminMCPServerDetail 管理员中心创建MCP（MCP服务）详情。
func AdminMCPServerDetail(ctx *gin.Context, req request.AdminMCPDetailReq) (*response.MCPServerDetail, error) {
	return GetMCPServerDetail(ctx, req.McpId)
}

// AdminMCPToolList 管理员中心获取MCP工具列表详情。
func AdminMCPToolList(ctx *gin.Context, req request.AdminMCPToolListReq) (*response.MCPToolList, error) {
	mcpToolsReq := request.MCPToolListReq{
		MCPID: req.McpId,
		Type:  req.Type,
	}
	return GetMCPToolList(ctx, mcpToolsReq)
}

// AdminToolPageList 管理员中心工具（自定义工具）全局列表。
func AdminToolPageList(ctx *gin.Context, req request.AdminToolPageListReq) (*response.PageResult, error) {
	pageNo, pageSize := normalizePage(req.PageNo, req.PageSize)
	resp, err := mcp.GetAdminCustomToolPageList(ctx.Request.Context(), &mcp_service.GetAdminCustomToolPageListReq{
		Name:       req.Name,
		OrgIdList:  req.OrgIdList,
		UserIdList: req.UserIdList,
		PageNo:     int32(pageNo),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return nil, err
	}

	// 构造列表项
	list := make([]*response.AdminTool, 0, len(resp.List))
	for _, item := range resp.List {
		toolItem := &response.AdminTool{
			Avatar:      cacheMCPServerAvatar(ctx, item.AvatarPath),
			ToolID:      item.CustomToolId,
			Name:        item.Name,
			Description: item.Description,
			UpdatedAt:   util.Time2Str(item.UpdatedAt),
			OwnerHolder: response.CreateOwnerHolder(item.Owner.UserId, item.Owner.OrgId),
		}
		list = append(list, toolItem)
	}
	fillOwnerList(ctx, list)
	return &response.PageResult{
		List:     list,
		Total:    resp.Total,
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// AdminToolBase 管理员中心工具基础信息。
func AdminToolBase(ctx *gin.Context, req request.AdminToolBaseReq) (*response.AdminToolBase, error) {
	switch req.Type {
	case constant.ToolTypeCustom:
		return AdminCustomToolBase(ctx, req)
	case constant.ToolTypeBuiltIn:
		return AdminBuiltinToolBase(ctx, req)
	}
	return nil, nil
}

// AdminCustomToolBase 管理员中心工具（自定义工具）基础信息。
func AdminCustomToolBase(ctx *gin.Context, req request.AdminToolBaseReq) (*response.AdminToolBase, error) {
	info, err := mcp.GetCustomToolInfo(ctx.Request.Context(), &mcp_service.GetCustomToolInfoReq{
		CustomToolId: req.ToolId,
		Identity:     &mcp_service.Identity{},
	})
	if err != nil {
		return nil, err
	}
	resp := &response.AdminToolBase{
		AdminAppBaseInfo: response.AdminAppBaseInfo{
			Avatar:      cacheToolAvatar(ctx, constant.ToolTypeCustom, info.AvatarPath),
			Name:        info.Name,
			Desc:        info.Description,
			CreatedAt:   util.Time2Str(info.CreatedAt),
			UpdatedAt:   util.Time2Str(info.UpdatedAt),
			OwnerHolder: response.CreateOwnerHolder(info.Owner.UserId, info.Owner.OrgId),
		},
	}
	fillOwner(ctx, resp)
	return resp, nil
}

// AdminBuiltinToolBase 管理员中心工具（内置工具）基础信息。
func AdminBuiltinToolBase(ctx *gin.Context, req request.AdminToolBaseReq) (*response.AdminToolBase, error) {
	squareTool, err := mcp.GetSquareTool(ctx.Request.Context(), &mcp_service.GetSquareToolReq{
		ToolSquareId: req.ToolId,
		Identity:     &mcp_service.Identity{UserId: "", OrgId: ""},
	})
	if err != nil {
		return nil, err
	}
	resp := &response.AdminToolBase{
		AdminAppBaseInfo: response.AdminAppBaseInfo{
			Avatar: cacheToolAvatar(ctx, constant.ToolTypeBuiltIn, squareTool.Info.AvatarPath),
			Name:   squareTool.Info.Name,
			Desc:   squareTool.Info.Desc,
		},
	}
	return resp, nil
}

// AdminToolDetail 管理员中心工具详情。
func AdminToolDetail(ctx *gin.Context, req request.AdminToolDetailReq) (*response.CustomToolDetail, error) {
	switch req.Type {
	case constant.ToolTypeCustom:
		return GetCustomTool(ctx, "", "", req.ToolId)
	case constant.ToolTypeBuiltIn:
		return GetSquareToolDetail(ctx, "", "", req.ToolId)
	}
	return nil, nil
}

func GetSquareToolDetail(ctx *gin.Context, userId string, orgId string, toolId string) (*response.CustomToolDetail, error) {
	squareDetail, err := mcp.GetSquareTool(ctx.Request.Context(), &mcp_service.GetSquareToolReq{
		ToolSquareId: toolId,
		Identity:     &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}
	doc, err := openapi3_util.LoadFromData(ctx.Request.Context(), []byte(squareDetail.Schema))
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	ret := &response.CustomToolDetail{
		CustomToolInfo: response.CustomToolInfo{
			CustomToolId: squareDetail.Info.ToolSquareId,
			Name:         squareDetail.Info.Name,
			Description:  squareDetail.Info.Desc,
			Avatar:       cacheToolAvatar(ctx, constant.ToolTypeBuiltIn, squareDetail.Info.AvatarPath),
		},
		Schema: squareDetail.Schema,
		ApiAuth: util.ApiAuthWebRequest{
			AuthType:           squareDetail.BuiltInTools.ApiAuth.AuthType,
			ApiKeyHeaderPrefix: squareDetail.BuiltInTools.ApiAuth.ApiKeyHeaderPrefix,
			ApiKeyHeader:       squareDetail.BuiltInTools.ApiAuth.ApiKeyHeader,
			ApiKeyQueryParam:   squareDetail.BuiltInTools.ApiAuth.ApiKeyQueryParam,
			ApiKeyValue:        squareDetail.BuiltInTools.ApiAuth.ApiKeyValue,
		},
		ApiList: openapiSchema2ToolList(doc),
	}
	return ret, nil
}

// AdminPromptPageList 管理员中心提示词（自定义提示词）全局列表。
func AdminPromptPageList(ctx *gin.Context, req request.AdminPromptPageListReq) (*response.PageResult, error) {
	pageNo, pageSize := normalizePage(req.PageNo, req.PageSize)
	resp, err := assistant.GetAdminCustomPromptPageList(ctx.Request.Context(), &assistant_service.GetAdminCustomPromptPageListReq{
		Name:       req.Name,
		OrgIdList:  req.OrgIdList,
		UserIdList: req.UserIdList,
		PageNo:     int32(pageNo),
		PageSize:   int32(pageSize),
	})
	if err != nil {
		return nil, err
	}

	// 构造列表项
	list := make([]*response.AdminPrompt, 0, len(resp.CustomPromptInfos))
	for _, item := range resp.CustomPromptInfos {
		promptItem := &response.AdminPrompt{
			Avatar:      cacheMCPServerAvatar(ctx, item.AvatarPath),
			PromptID:    item.CustomPromptId,
			Name:        item.Name,
			Description: item.Desc,
			UpdatedAt:   util.Time2Str(item.UpdatedAt),
			OwnerHolder: response.CreateOwnerHolder(item.Identity.UserId, item.Identity.OrgId),
		}
		list = append(list, promptItem)
	}
	fillOwnerList(ctx, list)

	return &response.PageResult{
		List:     list,
		Total:    resp.Total,
		PageNo:   pageNo,
		PageSize: pageSize,
	}, nil
}

// AdminPromptBase 管理员中心提示词（自定义提示词）详情基础信息。
func AdminPromptBase(ctx *gin.Context, req request.AdminPromptBaseReq) (*response.AdminPromptBase, error) {
	info, err := assistant.CustomPromptGet(ctx.Request.Context(), &assistant_service.CustomPromptGetReq{
		CustomPromptId: req.CustomPromptId,
		Identity:       &assistant_service.Identity{},
	})
	if err != nil {
		return nil, err
	}
	resp := &response.AdminPromptBase{
		AdminAppBaseInfo: response.AdminAppBaseInfo{
			Avatar:      cachePromptAvatar(ctx, info.AvatarPath),
			Name:        info.Name,
			Desc:        info.Desc,
			CreatedAt:   util.Time2Str(info.CreatedAt),
			UpdatedAt:   util.Time2Str(info.UpdatedAt),
			OwnerHolder: response.CreateOwnerHolder(info.Identity.UserId, info.Identity.OrgId),
		},
	}
	fillOwner(ctx, resp)
	return resp, nil
}

// AdminPromptDetail 管理员中心提示词（自定义提示词）详情。
func AdminPromptDetail(ctx *gin.Context, req request.AdminPromptDetailReq) (*response.CustomPrompt, error) {
	return GetCustomPrompt(ctx, "", "", req.CustomPromptId)
}
