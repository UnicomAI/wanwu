package service

import (
	"encoding/json"
	"sort"
	"strings"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	safety_service "github.com/UnicomAI/wanwu/api/proto/safety-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	bff_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/status"
)

func AssistantCreate(ctx *gin.Context, userId, orgId string, req request.AssistantCreateReq) (*response.AssistantCreateResp, error) {
	resp, err := assistant.AssistantCreate(ctx.Request.Context(), &assistant_service.AssistantCreateReq{
		AssistantBrief: assistantBriefConfigModel2Proto(req),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
		Category: int32(req.Category),
	})
	if err != nil {
		return nil, err
	}
	return &response.AssistantCreateResp{
		AssistantId: resp.AssistantId,
	}, nil
}

func AssistantUpdate(ctx *gin.Context, userId, orgId string, req request.AssistantUpdateReq) (interface{}, error) {
	existingAssistant, err := assistant.GetAssistantInfo(ctx.Request.Context(), &assistant_service.GetAssistantInfoReq{
		AssistantId: req.AssistantId,
		Identity:    &assistant_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}
	if err := util.ValidateBriefUpdate(&req.Name, existingAssistant.AssistantBrief.Name, &req.Desc, existingAssistant.AssistantBrief.Desc, util.SubjectAssistant); err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	_, err = assistant.AssistantUpdate(ctx.Request.Context(), &assistant_service.AssistantUpdateReq{
		AssistantId:    req.AssistantId,
		AssistantBrief: appBriefConfigModel2Proto(req.AppBriefConfig),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return nil, err
}

func AssistantConfigUpdate(ctx *gin.Context, userId, orgId string, req request.AssistantConfig) (interface{}, error) {
	var modelConfig, rerankConfig *common.AppModelConfig
	var err error
	if req.ModelConfig != nil {
		if req.ModelConfig.ModelId == "" {
			// 如果 modelId 为空，则认为用户想要清空模型配置，构造一个空的 AppModelConfig
			modelConfig = &common.AppModelConfig{}
		} else {
			modelConfig, err = appModelConfigModel2Proto(*req.ModelConfig)
			if err != nil {
				return nil, err
			}
		}
	}
	if req.RerankConfig != nil {
		if req.RerankConfig.ModelId == "" {
			// 如果 modelId 为空，则认为用户想要清空模型配置，构造一个空的 AppModelConfig
			rerankConfig = &common.AppModelConfig{}
		} else {
			rerankConfig, err = appModelConfigModel2Proto(*req.RerankConfig)
			if err != nil {
				return nil, err
			}
		}
	}
	var recommendConfig *assistant_service.AssistantRecommendConfig
	if req.RecommendConfig != nil {
		recommendConfig, err = recommendConfigModel2Proto(*req.RecommendConfig)
		if err != nil {
			return nil, err
		}
	}
	var safetyConfig *assistant_service.AssistantSafetyConfig
	if req.SafetyConfig != nil {
		safetyConfig = &assistant_service.AssistantSafetyConfig{
			Enable:         req.SafetyConfig.Enable,
			SensitiveTable: transSafetyConfig2Proto(req.SafetyConfig.Tables),
		}
	}

	var visionConfig *assistant_service.AssistantVisionConfig
	if req.VisionConfig != nil {
		visionConfig = &assistant_service.AssistantVisionConfig{
			PicNum: req.VisionConfig.PicNum,
		}
	}

	var memoryConfig *assistant_service.AssistantMemoryConfig
	if req.MemoryConfig != nil {
		memoryConfig = &assistant_service.AssistantMemoryConfig{
			MaxHistoryLength: req.MemoryConfig.MaxHistoryLength,
		}
	}

	var knowledgeBaseConfig *assistant_service.AssistantKnowledgeBaseConfig
	if req.KnowledgeBaseConfig != nil {
		knowledgeBaseConfig = transKnowledgebases2Proto(*req.KnowledgeBaseConfig)
	}

	_, err = assistant.AssistantConfigUpdate(ctx.Request.Context(), &assistant_service.AssistantConfigUpdateReq{
		AssistantId:         req.AssistantId,
		Prologue:            req.Prologue,
		Instructions:        req.Instructions,
		RecommendQuestion:   req.RecommendQuestion,
		ModelConfig:         modelConfig,
		KnowledgeBaseConfig: knowledgeBaseConfig,
		RerankConfig:        rerankConfig,
		SafetyConfig:        safetyConfig,
		VisionConfig:        visionConfig,
		MemoryConfig:        memoryConfig,
		RecommendConfig:     recommendConfig,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})

	return nil, err
}

func GetAssistantInfo(ctx *gin.Context, userId, orgId string, req request.AssistantIdRequest, needPublished bool) (*response.Assistant, error) {
	var resp *assistant_service.AssistantInfo
	var err error
	if needPublished {
		resp, err = assistant.AssistantSnapshotInfo(ctx.Request.Context(), &assistant_service.AssistantSnapshotInfoReq{
			AssistantId: req.AssistantId,
			Version:     req.Version,
		})
	} else {
		resp, err = assistant.GetAssistantInfo(ctx.Request.Context(), &assistant_service.GetAssistantInfoReq{
			AssistantId: req.AssistantId,
			Identity: &assistant_service.Identity{ //草稿只能看自己的
				UserId: userId,
				OrgId:  orgId,
			},
		})
	}
	if err != nil {
		return nil, err
	}
	return transAssistantResp2Model(ctx, resp)
}

func GetAssistantIdByUuid(ctx *gin.Context, uuid string) (string, error) {
	resp, err := assistant.GetAssistantIdByUuid(ctx.Request.Context(), &assistant_service.GetAssistantIdByUuidReq{
		Uuid: uuid,
	})
	if err != nil {
		return "", err
	}
	return resp.AssistantId, nil
}

func AssistantCopy(ctx *gin.Context, userId, orgId string, req request.AssistantIdRequest) (*response.AssistantCreateResp, error) {
	resp, err := assistant.AssistantCopy(ctx.Request.Context(), &assistant_service.AssistantCopyReq{
		AssistantId: req.AssistantId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.AssistantCreateResp{
		AssistantId: resp.AssistantId,
	}, nil
}

func AssistantWorkFlowCreate(ctx *gin.Context, userId, orgId string, req request.AssistantWorkFlowAddRequest) error {
	_, err := assistant.AssistantWorkFlowCreate(ctx.Request.Context(), &assistant_service.AssistantWorkFlowCreateReq{
		AssistantId: req.AssistantId,
		WorkFlowId:  req.WorkFlowId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantWorkFlowDelete(ctx *gin.Context, userId, orgId string, req request.AssistantWorkFlowDelRequest) error {
	_, err := assistant.AssistantWorkFlowDelete(ctx.Request.Context(), &assistant_service.AssistantWorkFlowDeleteReq{
		AssistantId: req.AssistantId,
		WorkFlowId:  req.WorkFlowId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantWorkFlowEnableSwitch(ctx *gin.Context, userId, orgId string, req request.AssistantWorkFlowToolEnableRequest) error {
	_, err := assistant.AssistantWorkFlowEnableSwitch(ctx.Request.Context(), &assistant_service.AssistantWorkFlowEnableSwitchReq{
		AssistantId: req.AssistantId,
		WorkFlowId:  req.WorkFlowId,
		Enable:      req.Enable,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantMCPCreate(ctx *gin.Context, userId, orgId string, req request.AssistantMCPToolAddRequest) error {
	_, err := assistant.AssistantMCPCreate(ctx.Request.Context(), &assistant_service.AssistantMCPCreateReq{
		AssistantId: req.AssistantId,
		McpId:       req.MCPId,
		McpType:     req.MCPType,
		ActionName:  req.ActionName,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantMCPDelete(ctx *gin.Context, userId, orgId string, req request.AssistantMCPToolDelRequest) error {
	_, err := assistant.AssistantMCPDelete(ctx.Request.Context(), &assistant_service.AssistantMCPDeleteReq{
		AssistantId: req.AssistantId,
		McpId:       req.MCPId,
		McpType:     req.MCPType,
		ActionName:  req.ActionName,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantMCPEnableSwitch(ctx *gin.Context, userId, orgId string, req request.AssistantMCPToolEnableRequest) error {
	_, err := assistant.AssistantMCPEnableSwitch(ctx.Request.Context(), &assistant_service.AssistantMCPEnableSwitchReq{
		AssistantId: req.AssistantId,
		McpId:       req.MCPId,
		McpType:     req.MCPType,
		ActionName:  req.ActionName,
		Enable:      req.Enable,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantToolCreate(ctx *gin.Context, userId, orgId string, req request.AssistantToolAddRequest) error {
	_, err := assistant.AssistantToolCreate(ctx.Request.Context(), &assistant_service.AssistantToolCreateReq{
		AssistantId: req.AssistantId,
		ToolId:      req.ToolId,
		ToolType:    req.ToolType,
		ActionName:  req.ActionName,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantToolDelete(ctx *gin.Context, userId, orgId string, req request.AssistantToolDelRequest) error {
	_, err := assistant.AssistantToolDelete(ctx.Request.Context(), &assistant_service.AssistantToolDeleteReq{
		AssistantId: req.AssistantId,
		ToolId:      req.ToolId,
		ToolType:    req.ToolType,
		ActionName:  req.ActionName,
	})
	return err
}

func AssistantToolEnableSwitch(ctx *gin.Context, userId, orgId string, req request.AssistantToolEnableRequest) error {
	_, err := assistant.AssistantToolEnableSwitch(ctx.Request.Context(), &assistant_service.AssistantToolEnableSwitchReq{
		AssistantId: req.AssistantId,
		ToolId:      req.ToolId,
		ToolType:    req.ToolType,
		ActionName:  req.ActionName,
		Enable:      req.Enable,
	})
	return err
}

func AssistantToolConfig(ctx *gin.Context, userId, orgId string, req request.AssistantToolConfigRequest) error {
	toolConfigJSON, err := json.Marshal(req.ToolConfig)
	if err != nil {
		return err
	}
	_, err = assistant.AssistantToolConfig(ctx.Request.Context(), &assistant_service.AssistantToolConfigReq{
		AssistantId: req.AssistantId,
		ToolId:      req.ToolId,
		ToolConfig:  string(toolConfigJSON),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantSkillCreate(ctx *gin.Context, userId, orgId string, req request.AssistantSkillAddRequest) error {
	_, err := assistant.AssistantSkillCreate(ctx.Request.Context(), &assistant_service.AssistantSkillCreateReq{
		AssistantId: req.AssistantId,
		SkillId:     req.SkillId,
		SkillType:   req.SkillType,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantSkillDelete(ctx *gin.Context, userId, orgId string, req request.AssistantSkillDelRequest) error {
	_, err := assistant.AssistantSkillDelete(ctx.Request.Context(), &assistant_service.AssistantSkillDeleteReq{
		AssistantId: req.AssistantId,
		SkillId:     req.SkillId,
		SkillType:   req.SkillType,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func AssistantSkillEnableSwitch(ctx *gin.Context, userId, orgId string, req request.AssistantSkillEnableSwitchRequest) error {
	_, err := assistant.AssistantSkillEnableSwitch(ctx.Request.Context(), &assistant_service.AssistantSkillEnableSwitchReq{
		AssistantId: req.AssistantId,
		SkillId:     req.SkillId,
		SkillType:   req.SkillType,
		Enable:      req.Enable,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func MultiAgentCreate(ctx *gin.Context, userId, orgId string, req request.MultiAgentCreateReq) error {
	_, err := assistant.MultiAgentCreate(ctx.Request.Context(), &assistant_service.MultiAgentCreateReq{
		AssistantId: req.AssistantId,
		AgentId:     req.AgentId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func MultiAgentDelete(ctx *gin.Context, userId, orgId string, req request.MultiAgentCreateReq) error {
	_, err := assistant.MultiAgentDelete(ctx.Request.Context(), &assistant_service.MultiAgentCreateReq{
		AssistantId: req.AssistantId,
		AgentId:     req.AgentId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func MultiAgentEnableSwitch(ctx *gin.Context, userId, orgId string, req request.MultiAgentEnableSwitchReq) error {
	_, err := assistant.MultiAgentEnableSwitch(ctx.Request.Context(), &assistant_service.MultiAgentEnableSwitchReq{
		AssistantId: req.AssistantId,
		AgentId:     req.AgentId,
		Enable:      req.Enable,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func MultiAgentConfigUpdate(ctx *gin.Context, userId, orgId string, req request.MultiAgentConfigUpdateReq) error {
	_, err := assistant.MultiAgentConfigUpdate(ctx.Request.Context(), &assistant_service.MultiAgentConfigUpdateReq{
		AssistantId: req.AssistantId,
		AgentId:     req.AgentId,
		Desc:        req.Desc,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func GetAssistantSelect(ctx *gin.Context, userId, orgId string, req request.GetExplorationAppListRequest) (*response.ListResult, error) {
	return getAssistantSelect(ctx, userId, orgId, req, false)
}

func GetAssistantSelectSingle(ctx *gin.Context, userId, orgId string, req request.GetExplorationAppListRequest) (*response.ListResult, error) {
	return getAssistantSelect(ctx, userId, orgId, req, true)
}

func getAssistantSelect(ctx *gin.Context, userId, orgId string, req request.GetExplorationAppListRequest, singleOnly bool) (*response.ListResult, error) {
	wlist, err := GetExplorationAppList(ctx, userId, orgId, request.GetExplorationAppListRequest{
		Name:       req.Name,
		AppType:    constant.AppTypeAgent,
		SearchType: "all",
	})
	if err != nil {
		return nil, err
	}
	var appList []*response.ExplorationAppInfo
	if wlistSlice, ok := wlist.List.([]*response.ExplorationAppInfo); ok {
		for _, w := range wlistSlice {
			if w.User.UserId == userId {
				if singleOnly && w.Category == constant.AgentCategoryMulti {
					continue
				}
				appList = append(appList, w)
			}
		}
	}
	return &response.ListResult{
		List:  appList,
		Total: int64(len(appList)),
	}, nil
}

func assistantWorkFlowConvert(ctx *gin.Context, workFlowInfos []*assistant_service.AssistantWorkFlowInfos) ([]*response.AssistantWorkFlowInfo, error) {
	var assistantWorkFlowInfos []*response.AssistantWorkFlowInfo
	if len(workFlowInfos) > 0 {
		var workflowIds []string
		for _, wf := range workFlowInfos {
			workflowIds = append(workflowIds, wf.WorkFlowId)
		}
		cozeWorkflowList, err := ListWorkflowByIDs(ctx, "", workflowIds)
		if err != nil {
			return nil, err
		}
		for _, wf := range workFlowInfos {
			workFlowInfo := &response.AssistantWorkFlowInfo{
				WorkFlowId: wf.WorkFlowId,
				ApiName:    wf.ApiName,
				Enable:     wf.Enable,
				UniqueId:   bff_util.ConcatAssistantToolUniqueId("workflow", wf.WorkFlowId),
			}

			for _, info := range cozeWorkflowList.Workflows {
				if info.WorkflowId == wf.WorkFlowId {
					// 找到匹配的工作流，设置名称和描述
					workFlowInfo.WorkFlowName = info.Name
					workFlowInfo.WorkFlowDesc = info.Desc
					workFlowInfo.AvatarPath = cacheWorkflowAvatar(info.URL, constant.AppTypeWorkflow)
				}
			}

			// 仅当工作流名称非空时才添加
			if workFlowInfo.WorkFlowName != "" {
				assistantWorkFlowInfos = append(assistantWorkFlowInfos, workFlowInfo)
				log.Debugf("添加工作流信息: WorkFlowId=%s, ApiName=%s", wf.WorkFlowId, wf.ApiName)
			}
		}
		log.Debugf("总共添加 %d 个工作流信息", len(assistantWorkFlowInfos))
	} else {
		log.Debugf("工作流信息为空")
	}
	return assistantWorkFlowInfos, nil
}

func assistantMCPConvert(ctx *gin.Context, assistantMCPInfos []*assistant_service.AssistantMCPInfos) ([]*response.AssistantMCPInfo, error) {
	// 若查询结果为空，返回空列表
	if len(assistantMCPInfos) == 0 {
		return nil, nil
	}

	// 提取MCP ID列表
	var MCPCustomIds, MCPServerIds []string
	for _, m := range assistantMCPInfos {
		switch m.McpType {
		case constant.MCPTypeMCP:
			MCPCustomIds = append(MCPCustomIds, m.McpId)
		case constant.MCPTypeMCPServer:
			MCPServerIds = append(MCPServerIds, m.McpId)
		}
	}

	// 批量查询MCP详情
	mcpResp, err := mcp.GetMCPByMCPIdList(ctx.Request.Context(), &mcp_service.GetMCPByMCPIdListReq{
		McpIdList:       MCPCustomIds,
		McpServerIdList: MCPServerIds,
	})

	// 构建MCP详情映射
	mcpDetailMap := make(map[string]*mcp_service.CustomMCPInfo)
	if err == nil && mcpResp != nil { // 仅当查询成功且响应有效时才构建映射
		for _, item := range mcpResp.Infos {
			mcpDetailMap[item.McpId] = item
		}
	}
	// 构建MCPServer详情映射
	mcpserverDetailMap := make(map[string]*mcp_service.MCPServerInfo)
	if err == nil && mcpResp != nil { // 仅当查询成功且响应有效时才构建映射
		for _, item := range mcpResp.Servers {
			mcpserverDetailMap[item.McpServerId] = item
		}
	}

	// 构建返回结果
	var retMCPInfos []*response.AssistantMCPInfo
	for _, info := range assistantMCPInfos {
		var exists bool
		var mcpName string
		var avatar request.Avatar
		var desc string

		switch info.McpType {
		case constant.MCPTypeMCP:
			if item, ok := mcpDetailMap[info.McpId]; ok {
				exists = true
				mcpName = item.Info.Name
				avatar = cacheMCPAvatar(ctx, item.Info.AvatarPath, item.AvatarPath)
				desc = item.GetInfo().GetDesc()
			}
		case constant.MCPTypeMCPServer:
			if item, ok := mcpserverDetailMap[info.McpId]; ok {
				exists = true
				mcpName = item.Name
				avatar = cacheMCPServerAvatar(ctx, item.AvatarPath)
				desc = item.GetDesc()
			}
		}

		if exists {
			retMCPInfos = append(retMCPInfos, &response.AssistantMCPInfo{
				UniqueId:    bff_util.ConcatAssistantToolUniqueId(info.McpType, info.McpId),
				MCPId:       info.McpId,
				MCPType:     info.McpType,
				MCPName:     mcpName,
				ActionName:  info.ActionName,
				Enable:      info.Enable,
				Valid:       true,
				Avatar:      avatar,
				Description: desc,
			})
		}
	}

	return retMCPInfos, nil
}

func assistantToolsConvert(ctx *gin.Context, assistantToolInfos []*assistant_service.AssistantToolInfos) ([]*response.AssistantToolInfo, error) {
	// 若查询为空，返回空列表
	if len(assistantToolInfos) == 0 {
		return nil, nil
	}

	// 提取工具ID列表
	var customToolIds, builtinToolIds []string
	for _, tool := range assistantToolInfos {
		switch tool.ToolType {
		case constant.ToolTypeCustom:
			customToolIds = append(customToolIds, tool.ToolId)
		case constant.ToolTypeBuiltIn:
			builtinToolIds = append(builtinToolIds, tool.ToolId)
		}
	}

	// 批量查询
	toolInfoResp, err := mcp.GetToolByIdList(ctx.Request.Context(), &mcp_service.GetToolByToolIdListReq{
		BuiltInToolIdList: builtinToolIds,
		CustomToolIdList:  customToolIds,
	})

	// 构建ID到工具信息的映射
	customToolMap := make(map[string]*mcp_service.GetCustomToolItem)
	if err == nil && toolInfoResp != nil { // 仅当查询成功且响应有效时才构建映射
		for _, item := range toolInfoResp.List {
			customToolMap[item.CustomToolId] = item
		}
	}
	builtinToolMap := make(map[string]*mcp_service.ToolSquareInfo)
	if err == nil && toolInfoResp != nil { // 仅当查询成功且响应有效时才构建映射
		for _, item := range toolInfoResp.ToolSquareInfoList {
			builtinToolMap[item.ToolSquareId] = item
		}
	}

	// 组装返回结果
	var retToolInfos []*response.AssistantToolInfo
	for _, info := range assistantToolInfos {
		var exists bool
		var toolName string
		var avatar request.Avatar
		var desc string

		switch info.ToolType {
		case constant.ToolTypeCustom:
			if item, ok := customToolMap[info.ToolId]; ok {
				exists = true
				toolName = item.Name
				avatar = cacheToolAvatar(ctx, constant.ToolTypeCustom, item.AvatarPath)
				desc = item.GetDescription()
			}
		case constant.ToolTypeBuiltIn:
			if item, ok := builtinToolMap[info.ToolId]; ok {
				exists = true
				toolName = item.Name
				avatar = cacheToolAvatar(ctx, constant.ToolTypeBuiltIn, item.AvatarPath)
				desc = item.GetDesc()
			}
		}

		if exists {
			var toolConfig request.AssistantToolConfig
			if info.ToolConfig != "" {
				if err := json.Unmarshal([]byte(info.ToolConfig), &toolConfig); err != nil {
					log.Warnf("解析ToolConfig失败，使用空配置，error: %v, toolConfig: %s", err, info.ToolConfig)
				}
			}
			retToolInfos = append(retToolInfos, &response.AssistantToolInfo{
				UniqueId:    bff_util.ConcatAssistantToolUniqueId(info.ToolType, info.ToolId),
				ToolId:      info.ToolId,
				ToolType:    info.ToolType,
				ToolName:    toolName,
				ActionName:  info.ActionName,
				Enable:      info.Enable,
				Valid:       true,
				ToolConfig:  toolConfig,
				Avatar:      avatar,
				Description: desc,
			})
		}
	}
	return retToolInfos, nil

}

func assistantSkillConvert(ctx *gin.Context, assistantSkillInfos []*assistant_service.AssistantSkillInfo) ([]*response.AssistantSkillInfo, error) {
	if len(assistantSkillInfos) == 0 {
		return nil, nil
	}

	var customSkillIds []string
	var acquiredSkillIds []string
	for _, skill := range assistantSkillInfos {
		if skill.SkillId == "" {
			continue
		}
		switch skill.SkillType {
		case constant.SkillTypeCustom:
			customSkillIds = append(customSkillIds, skill.SkillId)
		case constant.SkillTypeAcquired:
			acquiredSkillIds = append(acquiredSkillIds, skill.SkillId)
		}
	}

	// 获取自定义技能（只获取已发布的）
	customSkillMap := make(map[string]*mcp_service.PublishCustomSkill)
	if len(customSkillIds) > 0 {
		var err error
		customSkillMap, err = getCustomSkillPublishMap(ctx, customSkillIds)
		if err != nil {
			log.Warnf("获取自定义 skill 详情失败，err: %v", err)
			customSkillMap = make(map[string]*mcp_service.PublishCustomSkill)
		}
	}

	acquiredSkillMap := make(map[string]*mcp_service.AcquiredSkill)
	if len(acquiredSkillIds) > 0 {
		var err error
		acquiredSkillMap, err = getAcquiredSkillByIDMap(ctx, acquiredSkillIds)
		if err != nil {
			log.Warnf("获取我添加的 skill 详情失败，err: %v", err)
			acquiredSkillMap = make(map[string]*mcp_service.AcquiredSkill)
		}
	}

	// 批量获取 acquired skill 源发布信息用于权限校验
	var acquiredAppInfoMap map[string][]*app_service.AppInfo
	if len(acquiredSkillMap) > 0 {
		sourceSkillIds := make([]string, 0, len(acquiredSkillMap))
		for _, acquired := range acquiredSkillMap {
			if acquired == nil {
				continue
			}
			if sourceSkillId := acquired.GetSkill().GetSkill().GetSkillId(); sourceSkillId != "" {
				sourceSkillIds = append(sourceSkillIds, sourceSkillId)
			}
		}
		var err error
		acquiredAppInfoMap, err = getSourceSkillAppInfoMap(ctx, sourceSkillIds)
		if err != nil {
			log.Warnf("获取我添加的 skill 源发布信息失败，err: %v", err)
			acquiredAppInfoMap = make(map[string][]*app_service.AppInfo)
		}
	}

	var retSkillInfos []*response.AssistantSkillInfo
	for _, info := range assistantSkillInfos {
		var exists bool
		var skillName, author string
		var avatar request.Avatar
		var desc string

		switch info.SkillType {
		case constant.SkillTypeCustom:
			if publish, ok := customSkillMap[info.SkillId]; ok {
				// 只返回已发布的 custom skill
				if strings.TrimSpace(publish.GetVersion()) == "" {
					continue
				}
				customSkill := publish.GetSkill()
				if customSkill != nil {
					exists = true
					skillName = customSkill.GetName()
					author = customSkill.GetAuthor()
					avatar = cacheSkillAvatar(ctx, customSkill.GetAvatar())
					desc = customSkill.GetDesc()
				}
			}
		case constant.SkillTypeAcquired:
			if item, ok := acquiredSkillMap[info.SkillId]; ok {
				// 过滤无权限访问的 acquired skill
				if !isAcquiredSkillAccessible(ctx, item, acquiredAppInfoMap) {
					continue
				}
				customSkill := item.GetSkill().GetSkill()
				if customSkill != nil {
					exists = true
					skillName = customSkill.GetName()
					author = customSkill.GetAuthor()
					avatar = cacheSkillAvatar(ctx, customSkill.GetAvatar())
					desc = customSkill.GetDesc()
				}
			}
		case constant.SkillTypeBuiltIn:
			skillDetail, err := GetAgentSkillDetail(ctx, info.SkillId)
			if err == nil && skillDetail != nil {
				exists = true
				skillName = skillDetail.Name
				author = skillDetail.Author
				avatar = skillDetail.Avatar
				desc = skillDetail.Desc
			}
		}

		if exists {
			retSkillInfos = append(retSkillInfos, &response.AssistantSkillInfo{
				SkillId:     info.SkillId,
				SkillType:   info.SkillType,
				SkillName:   skillName,
				Author:      author,
				Enable:      info.Enable,
				Valid:       true,
				Avatar:      avatar,
				Description: desc,
			})
		}
	}
	return retSkillInfos, nil
}

func assistantSafetyConvert(ctx *gin.Context, resp *assistant_service.AssistantSafetyConfig) (request.AppSafetyConfig, error) {
	var exists bool
	enable := resp.GetEnable()
	var sensitiveTableList []request.SensitiveTable

	if len(resp.GetSensitiveTable()) != 0 {
		var tableIds []string
		for _, table := range resp.GetSensitiveTable() {
			tableIds = append(tableIds, table.TableId)
		}
		sensitiveWordTable, err := safety.GetSensitiveWordTableListByIDs(ctx.Request.Context(), &safety_service.GetSensitiveWordTableListByIDsReq{TableIds: tableIds})

		if err == nil && sensitiveWordTable != nil {
			exists = true
		} else {
			enable = false
		}

		if exists {
			for _, table := range sensitiveWordTable.List {
				sensitiveTableList = append(sensitiveTableList, request.SensitiveTable{
					TableId:   table.TableId,
					TableName: table.TableName,
				})
			}
		}
	}

	if len(sensitiveTableList) == 0 {
		enable = false
	}

	return request.AppSafetyConfig{
		Enable: enable,
		Tables: sensitiveTableList,
	}, nil
}

func ConversationCreate(ctx *gin.Context, userId, orgId string, req request.ConversationCreateRequest, conversationType string) (response.ConversationCreateResp, error) {
	resp, err := assistant.ConversationCreate(ctx.Request.Context(), &assistant_service.ConversationCreateReq{
		AssistantId:      req.AssistantId,
		Prompt:           req.Prompt,
		ConversationType: conversationType,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return response.ConversationCreateResp{}, err
	}
	return response.ConversationCreateResp{
		ConversationId: resp.ConversationId,
	}, nil
}

func ConversationDelete(ctx *gin.Context, userId, orgId string, req request.ConversationIdRequest) (interface{}, error) {
	_, err := assistant.ConversationDelete(ctx.Request.Context(), &assistant_service.ConversationDeleteReq{
		ConversationId: req.ConversationId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func ClearPublishedConversationES(ctx *gin.Context, userId, orgId string, req request.ConversationIdRequest) (interface{}, error) {
	// 清空已发布智能体的对话ES数据（不删除会话ID）
	_, err := assistant.ClearConversationES(ctx.Request.Context(), &assistant_service.ClearConversationESReq{
		ConversationId: req.ConversationId,
		DetailId:       req.DetailId,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func GetDraftConversationIdByAssistantID(ctx *gin.Context, userId, orgId string, req request.ConversationGetListRequest) (*response.ConversationIdResp, error) {
	resp, err := assistant.GetConversationIdByAssistantId(ctx.Request.Context(), &assistant_service.GetConversationIdByAssistantIdReq{
		AssistantId:      req.AssistantId,
		ConversationType: constant.ConversationTypeDraft,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		// 空状态（草稿态尚未发起过对话）vs 系统级故障
		if isRecordNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	return &response.ConversationIdResp{
		ConversationId: resp.ConversationId,
	}, nil
}

// isRecordNotFoundErr 判定 gRPC 错误是否由 gorm.ErrRecordNotFound 引起
func isRecordNotFoundErr(err error) bool {
	if err == nil {
		return false
	}
	st, ok := status.FromError(err)
	if !ok {
		// 非 gRPC status error，兜底字符串匹配
		return strings.Contains(err.Error(), "record not found")
	}
	for _, detail := range st.Details() {
		if s, ok := detail.(*err_code.Status); ok {
			for _, arg := range s.Args {
				if strings.Contains(arg, "record not found") {
					return true
				}
			}
		}
	}
	return false
}

func DraftConversationDeleteByAssistantID(ctx *gin.Context, userId, orgId string, req request.ConversationDeleteRequest) (interface{}, error) {
	// 获取 conversation_id
	conversationIdResp, err := assistant.GetConversationIdByAssistantId(ctx.Request.Context(), &assistant_service.GetConversationIdByAssistantIdReq{
		AssistantId:      req.AssistantId,
		ConversationType: constant.ConversationTypeDraft,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})

	if err != nil {
		// 草稿对话尚未创建：删除请求幂等成功，不向调用方抛 5xx。其它错误原样上抛。
		if isRecordNotFoundErr(err) {
			return nil, nil
		}
		return nil, err
	}
	if conversationIdResp == nil {
		return nil, nil
	}

	if req.DetailId != "" {
		// 传了 detailId：删除单条对话详情
		_, err = assistant.ClearConversationES(ctx.Request.Context(), &assistant_service.ClearConversationESReq{
			ConversationId: conversationIdResp.ConversationId,
			DetailId:       req.DetailId,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
	} else {
		// 未传 detailId：删除全部对话
		_, err = assistant.ConversationDelete(ctx.Request.Context(), &assistant_service.ConversationDeleteReq{
			ConversationId: conversationIdResp.ConversationId,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
	}
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func GetConversationList(ctx *gin.Context, userId, orgId string, req request.ConversationGetListRequest) (response.PageResult, error) {
	resp, err := assistant.GetConversationList(ctx.Request.Context(), &assistant_service.GetConversationListReq{
		AssistantId:      req.AssistantId,
		ConversationType: constant.ConversationTypePublished,
		PageSize:         int32(req.PageSize),
		PageNo:           int32(req.PageNo),
		SearchText:       strings.TrimSpace(req.SearchText),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return response.PageResult{}, err
	}
	var list []response.ConversationInfo
	for _, item := range resp.Data {
		list = append(list, response.ConversationInfo{
			ConversationId: item.ConversationId,
			AssistantId:    item.AssistantId,
			Title:          item.Title,
			CreatedAt:      util.Time2Str(item.CreatedAt),
			UpdatedAt:      util.Time2Str(item.UpdatedAt),
		})
	}
	return response.PageResult{Total: resp.Total, List: list}, nil
}

func GetConversationDetailList(ctx *gin.Context, userId, orgId string, req request.ConversationGetDetailListRequest) (response.PageResult, error) {
	resp, err := assistant.GetConversationDetailList(ctx.Request.Context(), &assistant_service.GetConversationDetailListReq{
		ConversationId: req.ConversationId,
		PageSize:       int32(req.PageSize),
		PageNo:         int32(req.PageNo),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return response.PageResult{}, err
	}

	// 转换resp.Data为自定义的ConversionDetailInfo结构体数组
	var convertedList []response.ConversationDetailInfo
	for _, item := range resp.Data {
		convertedItem := response.ConversationDetailInfo{
			Id:                  item.Id,
			AssistantId:         item.AssistantId,
			ConversationId:      item.ConversationId,
			Prompt:              item.Prompt,
			SysPrompt:           item.SysPrompt,
			Response:            item.Response,
			ResponseList:        buildResponseList(item.ConversationResponse),
			QaType:              item.QaType,
			CreatedBy:           item.CreatedBy,
			CreatedAt:           item.CreatedAt,
			UpdatedAt:           item.UpdatedAt,
			RequestFiles:        transRequestFiles(item.RequestFiles),
			FileSize:            item.FileSize,
			FileName:            item.FileName,
			SubConversationList: buildSubConversationList(item.SubConversationList),
			ResponseFiles:       transResponseFiles(item.ResponseFiles),
		}

		// 将SearchList从string转换为interface{}
		convertedItem.SearchList = buildSearchList(item.SearchList)

		convertedList = append(convertedList, convertedItem)

		// 对切片进行排序
		sort.Slice(convertedList, func(i, j int) bool {
			// CreatedAt值小的时间更早，排在前面
			return convertedList[i].CreatedAt < convertedList[j].CreatedAt
		})
	}

	return response.PageResult{Total: resp.Total, List: convertedList, PageNo: req.PageNo, PageSize: req.PageSize}, nil
}

func buildResponseList(conversationResponse []*assistant_service.ConversationResponse) []*response.ConversationResponse {
	if len(conversationResponse) == 0 {
		return make([]*response.ConversationResponse, 0)
	}
	var retList []*response.ConversationResponse
	for _, resp := range conversationResponse {
		retList = append(retList, &response.ConversationResponse{
			Response:    resp.Response,
			Order:       resp.Order,
			ErrMessage:  resp.ErrMessage,
			ErrResponse: resp.ErrResponse,
		})
	}
	return retList
}

func buildSearchList(searchListStr string) interface{} {
	// 将SearchList从string转换为interface{}
	if searchListStr != "" {
		var searchList interface{}
		if err := json.Unmarshal([]byte(searchListStr), &searchList); err != nil {
			log.Warnf("解析SearchList失败，使用原始字符串，error: %v, searchList: %s", err, searchListStr)
			return searchList
		} else {
			return searchListStr
		}
	}
	return nil
}

func buildSubConversationList(conversationList []*assistant_service.SubConversation) []*response.SubConversation {
	if len(conversationList) == 0 {
		return make([]*response.SubConversation, 0)
	}
	var subConversationList []*response.SubConversation
	for _, conversation := range conversationList {
		subConversationList = append(subConversationList, &response.SubConversation{
			Response:         conversation.Response,
			SearchList:       buildSearchList(conversation.SearchList),
			Id:               conversation.Id,
			Name:             conversation.Name,
			Profile:          conversation.Profile,
			TimeCost:         conversation.TimeCost,
			Status:           conversation.Status,
			ConversationType: conversation.ConversationType,
			Order:            conversation.Order,
			ParentId:         conversation.ParentId,
			ErrMessage:       conversation.ErrMessage,
		})
	}
	return subConversationList
}

func transKnowledgebases2Proto(kbConfig request.AppKnowledgebaseConfig) *assistant_service.AssistantKnowledgeBaseConfig {
	validKbs := make([]request.AppKnowledgeBase, 0, len(kbConfig.Knowledgebases))
	var knowIds []string
	for _, v := range kbConfig.Knowledgebases {
		if v.ID == "" {
			continue
		}
		validKbs = append(validKbs, v)
		knowIds = append(knowIds, v.ID)
	}
	return &assistant_service.AssistantKnowledgeBaseConfig{
		KnowledgeBaseIds:     knowIds,
		MaxHistory:           kbConfig.Config.MaxHistory,
		Threshold:            kbConfig.Config.Threshold,
		TopK:                 kbConfig.Config.TopK,
		MatchType:            kbConfig.Config.MatchType,
		KeywordPriority:      kbConfig.Config.KeywordPriority,
		PriorityMatch:        kbConfig.Config.PriorityMatch,
		SemanticsPriority:    kbConfig.Config.SemanticsPriority,
		TermWeight:           kbConfig.Config.TermWeight,
		TermWeightEnable:     kbConfig.Config.TermWeightEnable,
		UseGraph:             kbConfig.Config.UseGraph,
		AppKnowledgeBaseList: transKnowledgeParams(validKbs),
	}
}

func transKnowledgeParams(paramsList []request.AppKnowledgeBase) []*assistant_service.AppKnowledgeBase {
	if len(paramsList) == 0 {
		return nil
	}
	var retList []*assistant_service.AppKnowledgeBase
	for _, base := range paramsList {
		retList = append(retList, &assistant_service.AppKnowledgeBase{
			KnowledgeBaseId:      base.ID,
			KnowledgeBaseName:    base.Name,
			GraphSwitch:          base.GraphSwitch,
			Category:             base.Category,
			MetaDataFilterParams: transKnowledgeMetaParams(base.MetaDataFilterParams),
		})
	}
	return retList
}

func transKnowledgeMetaParams(baseInfo *request.MetaDataFilterParams) *assistant_service.MetaDataFilterParams {
	if baseInfo == nil {
		return nil
	}
	return &assistant_service.MetaDataFilterParams{
		FilterEnable:     baseInfo.FilterEnable,
		FilterLogicType:  baseInfo.FilterLogicType,
		MetaFilterParams: transMetaFilterParams(baseInfo.MetaFilterParams),
	}
}

func transMetaFilterParams(metaFilterList []*request.MetaFilterParams) []*assistant_service.MetaFilterParams {
	if metaFilterList == nil {
		return nil
	}
	var metaList []*assistant_service.MetaFilterParams
	for _, m := range metaFilterList {
		metaList = append(metaList, &assistant_service.MetaFilterParams{
			Condition: m.Condition,
			Key:       m.Key,
			Type:      m.Type,
			Value:     m.Value,
		})
	}
	return metaList
}

func transSafetyConfig2Proto(tables []request.SensitiveTable) []*assistant_service.SensitiveTable {
	if tables == nil {
		return nil
	}
	result := make([]*assistant_service.SensitiveTable, 0, len(tables))
	for _, table := range tables {
		result = append(result, &assistant_service.SensitiveTable{
			TableId:   table.TableId,
			TableName: table.TableName,
		})
	}
	return result
}

func transAssistantResp2Model(ctx *gin.Context, resp *assistant_service.AssistantInfo) (*response.Assistant, error) {
	log.Debugf("开始转换Assistant响应到模型，响应内容: %+v", resp)
	if resp == nil {
		log.Debugf("Assistant响应为空，返回空Assistant模型")
		return nil, nil
	}
	//构造智能体数据
	assistantModel := buildAssistant(ctx, resp)
	//并发填充各个智能体数据
	err := ConcurrentConverter(ctx, resp, assistantModel)
	if err != nil {
		log.Errorf("Assistant响应到模型转换失败: %v", err)
		return nil, err
	}
	log.Debugf("Assistant响应到模型转换完成，结果: %+v", *assistantModel)
	return assistantModel, nil
}

// 构建Assistant模型
func buildAssistant(ctx *gin.Context, resp *assistant_service.AssistantInfo) *response.Assistant {
	assistantModel := response.Assistant{
		AssistantId:       resp.AssistantId,
		UUID:              resp.Uuid,
		AppBriefConfig:    appBriefConfigProto2Model(ctx, resp.AssistantBrief, constant.AppTypeAgent),
		Prologue:          resp.Prologue,
		Instructions:      resp.Instructions,
		RecommendQuestion: resp.RecommendQuestion,
		Scope:             resp.Scope,
		CreatedAt:         util.Time2Str(resp.CreatTime),
		UpdatedAt:         util.Time2Str(resp.UpdateTime),
		NewAgent:          true,
		Category:          resp.Category,
		MemoryConfig: request.MemoryConfig{
			MaxHistoryLength: resp.MemoryConfig.MaxHistoryLength,
		},
	}
	// 转换Vision配置
	if resp.VisionConfig != nil {
		assistantModel.VisionConfig = response.VisionConfig{
			MaxPicNum: resp.VisionConfig.MaxPicNum,
			PicNum:    resp.VisionConfig.PicNum,
		}
	}
	return &assistantModel
}

func transRequestFiles(files []*assistant_service.RequestFile) []response.AssistantRequestFile {
	if files == nil {
		return nil
	}
	var result []response.AssistantRequestFile
	for _, file := range files {
		result = append(result, response.AssistantRequestFile{
			FileName: file.FileName,
			FileSize: file.FileSize,
			FileUrl:  file.FileUrl,
		})
	}
	return result
}

func transResponseFiles(files []*assistant_service.AgentFile) []*response.AgentResponseFile {
	if files == nil {
		return nil
	}
	var result []*response.AgentResponseFile
	for _, file := range files {
		var metadata = &response.AgentFileMeta{}
		if file.Metadata != nil {
			metadata.Desc = file.Metadata.Desc
			metadata.CreateAt = file.Metadata.CreateAt
			metadata.Name = file.Metadata.Name
		}
		result = append(result, &response.AgentResponseFile{
			FileName: file.Name,
			FileSize: int64(file.Size),
			FileUrl:  file.FileUrl,
			FileType: file.FileType,
			MetaData: metadata,
		})
	}
	return result
}

func assistantBriefConfigModel2Proto(appBrief request.AssistantCreateReq) *common.AppBriefConfig {
	return &common.AppBriefConfig{
		Name:       appBrief.Name,
		Desc:       appBrief.Desc,
		AvatarPath: appBrief.Avatar.Key,
	}

}

func recommendConfigModel2Proto(recommendConfig request.RecommendConfig) (ret *assistant_service.AssistantRecommendConfig, err error) {
	modelConfig := &common.AppModelConfig{}
	if recommendConfig.ModelConfig.ModelId != "" {
		modelConfig, err = appModelConfigModel2Proto(recommendConfig.ModelConfig)
		if err != nil {
			return nil, err
		}
	}
	ret = &assistant_service.AssistantRecommendConfig{
		RecommendEnable: recommendConfig.RecommendEnable,
		ModelConfig:     modelConfig,
		SystemPrompt:    strings.TrimSpace(recommendConfig.Prompt),
		PromptEnable:    recommendConfig.PromptEnable,
		MaxHistory:      recommendConfig.MaxHistory,
	}

	return ret, nil
}
