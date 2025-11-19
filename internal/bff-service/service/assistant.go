package service

import (
	"encoding/json"
	"sort"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	knowledgeBase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	safety_service "github.com/UnicomAI/wanwu/api/proto/safety-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	bff_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func AssistantCreate(ctx *gin.Context, userId, orgId string, req request.AppBriefConfig) (*response.AssistantCreateResp, error) {
	resp, err := assistant.AssistantCreate(ctx.Request.Context(), &assistant_service.AssistantCreateReq{
		AssistantBrief: appBriefConfigModel2Proto(req),
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

func AssistantUpdate(ctx *gin.Context, userId, orgId string, req request.AssistantBrief) (interface{}, error) {
	_, err := assistant.AssistantUpdate(ctx.Request.Context(), &assistant_service.AssistantUpdateReq{
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
	modelConfig, err := appModelConfigModel2Proto(req.ModelConfig)
	if err != nil {
		return nil, err
	}
	rerankConfig, err := appModelConfigModel2Proto(req.RerankConfig)
	if err != nil {
		return nil, err
	}
	_, err = assistant.AssistantConfigUpdate(ctx.Request.Context(), &assistant_service.AssistantConfigUpdateReq{
		AssistantId:         req.AssistantId,
		Prologue:            req.Prologue,
		Instructions:        req.Instructions,
		RecommendQuestion:   req.RecommendQuestion,
		ModelConfig:         modelConfig,
		KnowledgeBaseConfig: transKnowledgebases2Proto(req.KnowledgeBaseConfig),
		RerankConfig:        rerankConfig,
		SafetyConfig: &assistant_service.AssistantSafetyConfig{
			Enable:         req.SafetyConfig.Enable,
			SensitiveTable: transSafetyConfig2Proto(req.SafetyConfig.Tables),
		},
		VisionConfig: &assistant_service.AssistantVisionConfig{
			PicNum: req.VisionConfig.PicNum,
		},
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})

	return nil, err
}

func GetAssistantInfo(ctx *gin.Context, userId, orgId string, req request.AssistantIdRequest) (*response.Assistant, error) {
	resp, err := assistant.GetAssistantInfo(ctx.Request.Context(), &assistant_service.GetAssistantInfoReq{
		AssistantId: req.AssistantId,
	})
	if err != nil {
		return nil, err
	}
	return transAssistantResp2Model(ctx, resp)
}

func GetAssistantDraftInfo(ctx *gin.Context, userId, orgId string, req request.AssistantIdRequest) (*response.Assistant, error) {
	resp, err := assistant.GetAssistantInfo(ctx.Request.Context(), &assistant_service.GetAssistantInfoReq{
		AssistantId: req.AssistantId,
		Identity: &assistant_service.Identity{ //草稿只能看自己的 [EN] You can only read your own drafts
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return transAssistantResp2Model(ctx, resp)
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

func assistantMCPConvert(ctx *gin.Context, assistantMCPInfos []*assistant_service.AssistantMCPInfos) ([]*response.AssistantMCPInfo, error) {
	// 若查询结果为空，返回空列表 [EN] If the query result is empty, return an empty list
	if len(assistantMCPInfos) == 0 {
		return nil, nil
	}

	// 提取MCP ID列表 [EN] Extract MCP ID list
	var MCPCustomIds, MCPServerIds []string
	for _, m := range assistantMCPInfos {
		if m.McpType == constant.MCPTypeMCP {
			MCPCustomIds = append(MCPCustomIds, m.McpId)
		} else if m.McpType == constant.MCPTypeMCPServer {
			MCPServerIds = append(MCPServerIds, m.McpId)
		}
	}

	// 批量查询MCP详情 [EN] Query MCP details in batches
	mcpResp, err := mcp.GetMCPByMCPIdList(ctx.Request.Context(), &mcp_service.GetMCPByMCPIdListReq{
		McpIdList:       MCPCustomIds,
		McpServerIdList: MCPServerIds,
	})

	// 构建MCP详情映射 [EN] Build MCP details mapping
	mcpDetailMap := make(map[string]*mcp_service.CustomMCPInfo)
	if err == nil && mcpResp != nil { // 仅当查询成功且响应有效时才构建映射 [EN] The map is only built if the query is successful and the response is valid
		for _, item := range mcpResp.Infos {
			mcpDetailMap[item.McpId] = item
		}
	}
	// 构建MCPServer详情映射 [EN] Build MCPServer details mapping
	mcpserverDetailMap := make(map[string]*mcp_service.MCPServerInfo)
	if err == nil && mcpResp != nil { // 仅当查询成功且响应有效时才构建映射 [EN] The map is only built if the query is successful and the response is valid
		for _, item := range mcpResp.Servers {
			mcpserverDetailMap[item.McpServerId] = item
		}
	}

	// 构建返回结果 [EN] Build return results
	var retMCPInfos []*response.AssistantMCPInfo
	for _, info := range assistantMCPInfos {
		var exists bool
		var mcpName string
		var avatar request.Avatar

		switch info.McpType {
		case constant.MCPTypeMCP:
			if item, ok := mcpDetailMap[info.McpId]; ok {
				exists = true
				mcpName = item.Info.Name
				avatar = cacheMCPAvatar(ctx, item.Info.AvatarPath, item.AvatarPath)
			}
		case constant.MCPTypeMCPServer:
			if item, ok := mcpserverDetailMap[info.McpId]; ok {
				exists = true
				mcpName = item.Name
				avatar = cacheMCPServerAvatar(ctx, item.AvatarPath)
			}
		}

		if exists {
			retMCPInfos = append(retMCPInfos, &response.AssistantMCPInfo{
				UniqueId:   bff_util.ConcatAssistantToolUniqueId(info.McpType, info.McpId),
				MCPId:      info.McpId,
				MCPType:    info.McpType,
				MCPName:    mcpName,
				ActionName: info.ActionName,
				Enable:     info.Enable,
				Valid:      true,
				Avatar:     avatar,
			})
		}
	}

	return retMCPInfos, nil
}

func assistantToolsConvert(ctx *gin.Context, assistantToolInfos []*assistant_service.AssistantToolInfos) ([]*response.AssistantToolInfo, error) {
	// 若查询为空，返回空列表 [EN] If the query is empty, an empty list is returned
	if len(assistantToolInfos) == 0 {
		return nil, nil
	}

	// 提取工具ID列表 [EN] Extract tool ID list
	var customToolIds, builtinToolIds []string
	for _, tool := range assistantToolInfos {
		switch tool.ToolType {
		case constant.ToolTypeCustom:
			customToolIds = append(customToolIds, tool.ToolId)
		case constant.ToolTypeBuiltIn:
			builtinToolIds = append(builtinToolIds, tool.ToolId)
		}
	}

	// 批量查询 [EN] Batch query
	toolInfoResp, err := mcp.GetToolByIdList(ctx.Request.Context(), &mcp_service.GetToolByToolIdListReq{
		BuiltInToolIdList: builtinToolIds,
		CustomToolIdList:  customToolIds,
	})

	// 构建ID到工具信息的映射 [EN] Build a mapping from ID to tool information
	customToolMap := make(map[string]*mcp_service.GetCustomToolItem)
	if err == nil && toolInfoResp != nil { // 仅当查询成功且响应有效时才构建映射 [EN] The map is only built if the query is successful and the response is valid
		for _, item := range toolInfoResp.List {
			customToolMap[item.CustomToolId] = item
		}
	}
	builtinToolMap := make(map[string]*mcp_service.ToolSquareInfo)
	if err == nil && toolInfoResp != nil { // 仅当查询成功且响应有效时才构建映射 [EN] The map is only built if the query is successful and the response is valid
		for _, item := range toolInfoResp.ToolSquareInfoList {
			builtinToolMap[item.ToolSquareId] = item
		}
	}

	// 组装返回结果 [EN] Assemble return results
	var retToolInfos []*response.AssistantToolInfo
	for _, info := range assistantToolInfos {
		var exists bool
		var toolName string
		var avatar request.Avatar

		switch info.ToolType {
		case constant.ToolTypeCustom:
			if item, ok := customToolMap[info.ToolId]; ok {
				exists = true
				toolName = item.Name
				avatar = cacheToolAvatar(ctx, constant.ToolTypeCustom, item.AvatarPath)
			}
		case constant.ToolTypeBuiltIn:
			if item, ok := builtinToolMap[info.ToolId]; ok {
				exists = true
				toolName = item.Name
				avatar = cacheToolAvatar(ctx, constant.ToolTypeBuiltIn, item.AvatarPath)
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
				UniqueId:   bff_util.ConcatAssistantToolUniqueId(info.ToolType, info.ToolId),
				ToolId:     info.ToolId,
				ToolType:   info.ToolType,
				ToolName:   toolName,
				ActionName: info.ActionName,
				Enable:     info.Enable,
				Valid:      true,
				ToolConfig: toolConfig,
				Avatar:     avatar,
			})
		}
	}
	return retToolInfos, nil

}

func ConversationCreate(ctx *gin.Context, userId, orgId string, req request.ConversationCreateRequest) (response.ConversationCreateResp, error) {
	resp, err := assistant.ConversationCreate(ctx.Request.Context(), &assistant_service.ConversationCreateReq{
		AssistantId: req.AssistantId,
		Prompt:      req.Prompt,
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

func GetConversationList(ctx *gin.Context, userId, orgId string, req request.ConversationGetListRequest) (response.PageResult, error) {
	resp, err := assistant.GetConversationList(ctx.Request.Context(), &assistant_service.GetConversationListReq{
		AssistantId: req.AssistantId,
		PageSize:    int32(req.PageSize),
		PageNo:      int32(req.PageNo),
		Identity: &assistant_service.Identity{
			UserId: userId,
		},
	})
	if err != nil {
		return response.PageResult{}, err
	}
	return response.PageResult{Total: resp.Total, List: resp.Data}, nil
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

	// 转换resp.Data为自定义的ConversionDetailInfo结构体数组 [EN] Convert resp.Data into a custom ConversionDetailInfo structure array
	var convertedList []response.ConversationDetailInfo
	for _, item := range resp.Data {
		convertedItem := response.ConversationDetailInfo{
			Id:             item.Id,
			AssistantId:    item.AssistantId,
			ConversationId: item.ConversationId,
			Prompt:         item.Prompt,
			SysPrompt:      item.SysPrompt,
			Response:       item.Response,
			QaType:         item.QaType,
			CreatedBy:      item.CreatedBy,
			CreatedAt:      item.CreatedAt,
			UpdatedAt:      item.UpdatedAt,
			RequestFiles:   transRequestFiles(item.RequestFiles),
			FileSize:       item.FileSize,
			FileName:       item.FileName,
		}

		// 将SearchList从string转换为interface{} [EN] Convert SearchList from string to interface{}
		if item.SearchList != "" {
			var searchList interface{}
			if err := json.Unmarshal([]byte(item.SearchList), &searchList); err != nil {
				log.Warnf("解析SearchList失败，使用原始字符串，error: %v, searchList: %s", err, item.SearchList)
				convertedItem.SearchList = item.SearchList
			} else {
				convertedItem.SearchList = searchList
			}
		} else {
			convertedItem.SearchList = nil
		}

		convertedList = append(convertedList, convertedItem)

		// 对切片进行排序 [EN] Sort slices
		sort.Slice(convertedList, func(i, j int) bool {
			// CreatedAt值小的时间更早，排在前面 [EN] The time with a smaller CreatedAt value is earlier and ranked first.
			return convertedList[i].CreatedAt < convertedList[j].CreatedAt
		})
	}

	return response.PageResult{Total: resp.Total, List: convertedList, PageNo: req.PageNo, PageSize: req.PageSize}, nil
}

func transKnowledgebases2Proto(kbConfig request.AppKnowledgebaseConfig) *assistant_service.AssistantKnowledgeBaseConfig {
	var knowIds []string
	if len(kbConfig.Knowledgebases) > 0 {
		for _, v := range kbConfig.Knowledgebases {
			knowIds = append(knowIds, v.ID)
		}
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
		AppKnowledgeBaseList: transKnowledgeParams(kbConfig.Knowledgebases),
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
	var modelConfig request.AppModelConfig
	if resp.ModelConfig != nil && resp.ModelConfig.ModelId != "" {
		log.Debugf("检测到模型配置，模型ID: %s", resp.ModelConfig.ModelId)
		modelInfo, err := model.GetModelById(ctx.Request.Context(), &model_service.GetModelByIdReq{ModelId: resp.ModelConfig.ModelId})
		if err != nil {
			log.Errorf("获取模型信息失败，模型ID: %s, 错误: %v", resp.ModelConfig.ModelId, err)
		}
		if modelInfo != nil {
			modelConfig, err = appModelConfigProto2Model(resp.ModelConfig, modelInfo.DisplayName)
			if err != nil {
				log.Errorf("模型配置Proto转换到模型失败，模型ID: %s, 错误: %v", resp.ModelConfig.ModelId, err)
				return nil, err
			}
			log.Debugf("模型配置转换成功: %+v", modelConfig)
		}
	} else {
		log.Debugf("模型配置为空或模型ID为空")
	}
	var rerankConfig request.AppModelConfig
	if resp.RerankConfig != nil && resp.RerankConfig.ModelId != "" {
		log.Debugf("检测到Rerank配置，模型ID: %s", resp.RerankConfig.ModelId)
		modelInfo, err := model.GetModelById(ctx.Request.Context(), &model_service.GetModelByIdReq{ModelId: resp.RerankConfig.ModelId})
		if err != nil {
			log.Errorf("获取Rerank模型信息失败，模型ID: %s, 错误: %v", resp.RerankConfig.ModelId, err)
			return nil, err
		}
		rerankConfig, err = appModelConfigProto2Model(resp.RerankConfig, modelInfo.DisplayName)
		if err != nil {
			log.Errorf("Rerank配置Proto转换到模型失败，模型ID: %s, 错误: %v", resp.RerankConfig.ModelId, err)
			return nil, err
		}
		log.Debugf("Rerank配置转换成功: %+v", rerankConfig)
	} else {
		log.Debugf("Rerank配置为空或模型ID为空")
	}

	var assistantWorkFlowInfos []*response.AssistantWorkFlowInfo
	if len(resp.WorkFlowInfos) > 0 {
		var workflowIds []string
		for _, wf := range resp.WorkFlowInfos {
			workflowIds = append(workflowIds, wf.WorkFlowId)
		}
		cozeWorkflowList, err := ListWorkflowByIDs(ctx, "", workflowIds)
		if err != nil {
			return nil, err
		}
		for _, wf := range resp.WorkFlowInfos {
			workFlowInfo := &response.AssistantWorkFlowInfo{
				WorkFlowId: wf.WorkFlowId,
				ApiName:    wf.ApiName,
				Enable:     wf.Enable,
				UniqueId:   bff_util.ConcatAssistantToolUniqueId("workflow", wf.WorkFlowId),
			}

			for _, info := range cozeWorkflowList.Workflows {
				if info.WorkflowId == wf.WorkFlowId {
					// 找到匹配的工作流，设置名称和描述 [EN] Find matching workflow, set name and description
					workFlowInfo.WorkFlowName = info.Name
					workFlowInfo.WorkFlowDesc = info.Desc
					workFlowInfo.AvatarPath = cacheWorkflowAvatar(info.URL, constant.AppTypeWorkflow)
				}
			}

			assistantWorkFlowInfos = append(assistantWorkFlowInfos, workFlowInfo)
			log.Debugf("添加工作流信息: WorkFlowId=%s, ApiName=%s", wf.WorkFlowId, wf.ApiName)
		}
		log.Debugf("总共添加 %d 个工作流信息", len(assistantWorkFlowInfos))
	} else {
		log.Debugf("工作流信息为空")
	}

	// 查询该用户所有权限的所有 MCP [EN] Query all MCPs with all permissions of this user
	assistantMCPInfos, err := assistantMCPConvert(ctx, resp.McpInfos)
	if err != nil {
		return nil, err
	}
	// 查询该用户所有权限的 custom、builtin 工具 [EN] Query custom and builtin tools for all permissions of the user
	assistantToolInfos, err := assistantToolsConvert(ctx, resp.ToolInfos)
	if err != nil {
		return nil, err
	}

	var sensitiveWordTable *safety_service.SensitiveWordTables
	if len(resp.SafetyConfig.GetSensitiveTable()) != 0 {
		var tableIds []string
		for _, table := range resp.SafetyConfig.SensitiveTable {
			tableIds = append(tableIds, table.TableId)
		}
		sensitiveWordTable, _ = safety.GetSensitiveWordTableListByIDs(ctx, &safety_service.GetSensitiveWordTableListByIDsReq{TableIds: tableIds})
	}
	knowledgeBaseConfig, err := transKnowledgeBases2Model(ctx, resp.KnowledgeBaseConfig)
	if err != nil {
		return nil, err
	}
	var visionConfig response.VisionConfig
	if resp.VisionConfig != nil {
		visionConfig = response.VisionConfig{
			MaxPicNum: resp.VisionConfig.MaxPicNum,
			PicNum:    resp.VisionConfig.PicNum,
		}
	}
	assistantModel := response.Assistant{
		AssistantId:         resp.AssistantId,
		AppBriefConfig:      appBriefConfigProto2Model(ctx, resp.AssistantBrief, constant.AppTypeAgent),
		Prologue:            resp.Prologue,
		Instructions:        resp.Instructions,
		RecommendQuestion:   resp.RecommendQuestion,
		KnowledgeBaseConfig: knowledgeBaseConfig,
		ModelConfig:         modelConfig,
		RerankConfig:        rerankConfig,
		SafetyConfig:        request.AppSafetyConfig{Enable: resp.SafetyConfig.GetEnable()},
		VisionConfig:        visionConfig,
		Scope:               resp.Scope,
		WorkFlowInfos:       assistantWorkFlowInfos,
		MCPInfos:            assistantMCPInfos,
		ToolInfos:           assistantToolInfos,
		CreatedAt:           util.Time2Str(resp.CreatTime),
		UpdatedAt:           util.Time2Str(resp.UpdateTime),
	}
	if sensitiveWordTable != nil {
		var sensitiveTableList []request.SensitiveTable
		for _, table := range sensitiveWordTable.List {
			sensitiveTableList = append(sensitiveTableList, request.SensitiveTable{
				TableId:   table.TableId,
				TableName: table.TableName,
			})
		}
		assistantModel.SafetyConfig.Tables = sensitiveTableList
	}
	log.Debugf("Assistant响应到模型转换完成，结果: %+v", assistantModel)
	return &assistantModel, nil
}

func transKnowledgeBases2Model(ctx *gin.Context, kbConfig *assistant_service.AssistantKnowledgeBaseConfig) (request.AppKnowledgebaseConfig, error) {
	if kbConfig == nil {
		log.Debugf("知识库配置为空")
		return request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
		}, nil
	}
	if len(kbConfig.KnowledgeBaseIds) == 0 {
		log.Debugf("知识库配置为空")
		return request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
		}, nil
	}

	// 获取知识库详情列表 [EN] Get a list of knowledge base details
	kbInfoList, err := knowledgeBase.SelectKnowledgeDetailByIdList(ctx, &knowledgeBase_service.KnowledgeDetailSelectListReq{
		KnowledgeIds: kbConfig.KnowledgeBaseIds,
	})
	if err != nil {
		return request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
		}, err
	}

	knowledgeBases := buildKnowledgeBases(kbInfoList, kbConfig.KnowledgeBaseIds, kbConfig.AppKnowledgeBaseList)

	return request.AppKnowledgebaseConfig{
		Knowledgebases: knowledgeBases,
		Config: request.AppKnowledgebaseParams{
			MaxHistory:        kbConfig.MaxHistory,
			Threshold:         kbConfig.Threshold,
			TopK:              kbConfig.TopK,
			MatchType:         kbConfig.MatchType,
			PriorityMatch:     kbConfig.PriorityMatch,
			SemanticsPriority: kbConfig.SemanticsPriority,
			KeywordPriority:   kbConfig.KeywordPriority,
			TermWeight:        kbConfig.TermWeight,
			TermWeightEnable:  kbConfig.TermWeightEnable,
			UseGraph:          kbConfig.UseGraph,
		},
	}, nil

}

func buildKnowledgeBases(kbInfoList *knowledgeBase_service.KnowledgeDetailSelectListResp, kbIdList []string, kbConfigList []*assistant_service.AppKnowledgeBase) []request.AppKnowledgeBase {
	if len(kbInfoList.List) == 0 {
		return make([]request.AppKnowledgeBase, 0)
	}
	var knowledgeMap = make(map[string]*knowledgeBase_service.KnowledgeInfo)
	for _, kbInfo := range kbInfoList.List {
		knowledgeMap[kbInfo.KnowledgeId] = kbInfo
	}
	var knowledgeBases = make([]request.AppKnowledgeBase, 0)
	if len(kbConfigList) > 0 {
		for _, kbConfig := range kbConfigList {
			info := knowledgeMap[kbConfig.KnowledgeBaseId]
			if info == nil {
				continue
			}
			params := buildAssistantMetaDataFilterParams(kbConfig)
			knowledgeBases = append(knowledgeBases, request.AppKnowledgeBase{
				ID:                   kbConfig.KnowledgeBaseId,
				Name:                 info.Name,
				GraphSwitch:          info.GraphSwitch,
				MetaDataFilterParams: params,
			})
		}
	} else {
		for _, kbId := range kbIdList {
			info := knowledgeMap[kbId]
			if info == nil {
				continue
			}
			knowledgeBases = append(knowledgeBases, request.AppKnowledgeBase{
				ID:   kbId,
				Name: info.Name,
			})
		}
	}

	return knowledgeBases
}

func buildAssistantMetaDataFilterParams(kbConfig *assistant_service.AppKnowledgeBase) *request.MetaDataFilterParams {
	params := kbConfig.MetaDataFilterParams
	if params == nil {
		return nil
	}
	return &request.MetaDataFilterParams{
		FilterEnable:     params.FilterEnable,
		FilterLogicType:  params.FilterLogicType,
		MetaFilterParams: buildAssistantMetaFilterParams(params.MetaFilterParams),
	}
}

func buildAssistantMetaFilterParams(metaFilterList []*assistant_service.MetaFilterParams) []*request.MetaFilterParams {
	if metaFilterList == nil {
		return nil
	}
	var metaList []*request.MetaFilterParams
	for _, m := range metaFilterList {
		metaList = append(metaList, &request.MetaFilterParams{
			Condition: m.Condition,
			Key:       m.Key,
			Type:      m.Type,
			Value:     m.Value,
		})
	}
	return metaList
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
