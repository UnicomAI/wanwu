package service

import (
	"encoding/json"
	"fmt"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_persistent "github.com/UnicomAI/wanwu/pkg/wga-persistent"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/gin-gonic/gin"
)

func GetGeneralAgentSubList(ctx *gin.Context) (*response.GetGeneralAgentSubListResp, error) {
	result := &response.GetGeneralAgentSubListResp{}

	// 获取wga所支持的子智能体
	for _, agent := range config.WgaCfg().SubAgents {
		if agent.AgentID == "DIP Agent" && config.Cfg().Ontology.Enable == 0 {
			continue
		}
		result.WgaAgentList = append(result.WgaAgentList, response.GeneralAgentInfo{
			AgentID:     agent.AgentID,
			AgentName:   agent.AgentName,
			Avatar:      request.Avatar{Path: agent.AvatarPath},
			Placeholder: agent.Placeholder,
		})
	}

	return result, nil
}

func GetGeneralAgentUploadLimit(ctx *gin.Context, userId, orgId string) (*response.GeneralAgentUploadLimitResp, error) {
	cfg := config.WgaCfg()
	uploadLimit := cfg.UploadLimit

	retList := make([]*response.GeneralAgentUploadLimit, 0)
	retList = append(retList, buildGeneralAgentUploadLimit("image", uploadLimit.ImageTypes, uploadLimit.MaxImageSize))
	retList = append(retList, buildGeneralAgentUploadLimit("file", uploadLimit.FileTypes, uploadLimit.MaxFileSize))

	return &response.GeneralAgentUploadLimitResp{
		UploadLimitList: retList,
	}, nil
}

func buildGeneralAgentUploadLimit(fileType, extStr string, maxSize int) *response.GeneralAgentUploadLimit {
	var extList []string
	if extStr != "" {
		extList = strings.Split(extStr, ";")
	}
	return &response.GeneralAgentUploadLimit{
		FileType: fileType,
		MaxSize:  maxSize,
		ExtList:  extList,
	}
}

func UpdateGeneralAgentConfig(ctx *gin.Context, userId, orgId string, req request.UpdateGeneralAgentConfigReq) error {
	// 解析请求，转换为内部格式
	var toolList []*assistant_service.WgaConfigTool
	var mcpList []*assistant_service.WgaConfigMcp
	var workflowList []*assistant_service.WgaConfigWorkflow
	var skillList []*assistant_service.WgaConfigSkill
	var assistantList []*assistant_service.WgaConfigAssistant
	var knowledgeList []*assistant_service.WgaConfigKnowledge
	var ontologyList []*assistant_service.WgaConfigOntologyKnowledge
	var toolIds []string

	// 处理 tool
	for _, item := range req.Tool {
		toolList = append(toolList, &assistant_service.WgaConfigTool{
			ToolId:   item.ID,
			ToolType: item.Type,
		})
		if item.Type == constant.ToolTypeBuiltIn {
			toolIds = append(toolIds, item.ID)
		}
	}

	// 处理 mcp
	for _, item := range req.Mcp {
		mcpList = append(mcpList, &assistant_service.WgaConfigMcp{
			McpId:   item.ID,
			McpType: item.Type,
		})
	}

	// 处理 workflow
	for _, item := range req.Workflow {
		workflowList = append(workflowList, &assistant_service.WgaConfigWorkflow{
			WorkflowId: item.ID,
		})
	}

	// 处理 skill
	for _, item := range req.Skill {
		skillList = append(skillList, &assistant_service.WgaConfigSkill{
			SkillId:   item.ID,
			SkillType: item.Type,
		})
	}

	// 处理 assistant
	for _, item := range req.Assistant {
		assistantList = append(assistantList, &assistant_service.WgaConfigAssistant{
			AssistantId:   item.ID,
			AssistantType: util.Int2Str(constant.AgentCategorySingle), // 默认单智能体
		})
	}

	// 处理 knowledge
	for _, item := range req.Knowledge {
		knowledgeList = append(knowledgeList, &assistant_service.WgaConfigKnowledge{
			KnowledgeId: item.ID,
		})
	}

	// 处理 ontology
	for _, item := range req.Ontology {
		ontologyList = append(ontologyList, &assistant_service.WgaConfigOntologyKnowledge{
			OntologyKnowledgeId: item.ID,
		})
	}

	// 校验 tool 配置
	validToolIds, _ := getValidToolIds(ctx, userId, orgId, toolIds)
	for _, t := range toolList {
		if t.ToolType == constant.ToolTypeBuiltIn {
			if !validToolIds[t.ToolId] {
				return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("tool not found or invalid: %s", t.ToolId))
			}
		}
	}

	// 校验 mcp 配置
	if err := checkWgaMCPConfig(ctx, userId, orgId, mcpList); err != nil {
		return err
	}

	// 校验 workflow 配置
	if err := checkWgaWorkflowConfig(ctx, userId, orgId, workflowList); err != nil {
		return err
	}

	// 校验 skill 配置
	validSkillList, err := checkWgaSkillConfig(ctx, userId, orgId, skillList)
	if err != nil {
		return err
	}

	// 校验 assistant 配置
	if err := checkWgaAssistantConfig(ctx, userId, orgId, assistantList); err != nil {
		return err
	}

	// 校验 knowledge 配置
	if err := checkWgaKnowledgeConfig(ctx, userId, orgId, knowledgeList); err != nil {
		return err
	}

	// 校验 ontology 配置
	if err := checkWgaOntologyKnowledgeConfig(ctx, userId, orgId, ontologyList); err != nil {
		return err
	}

	_, err = assistant.UpdateWgaConfig(ctx.Request.Context(), &assistant_service.UpdateWgaConfigReq{
		ToolList:              toolList,
		McpList:               mcpList,
		WorkflowList:          workflowList,
		SkillList:             validSkillList,
		AssistantList:         assistantList,
		KnowledgeList:         knowledgeList,
		OntologyKnowledgeList: ontologyList,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

func GetGeneralAgentConfig(ctx *gin.Context, userId, orgId string) (response.GetGeneralAgentConfigResp, error) {
	resp, err := assistant.GetWgaConfig(ctx.Request.Context(), &assistant_service.GetWgaConfigReq{
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	result := make(response.GetGeneralAgentConfigResp, 0)

	// 过滤存在的 tool
	toolIds := make([]string, 0, len(resp.Config.ToolList))
	for _, t := range resp.Config.ToolList {
		if t.ToolType == constant.ToolTypeBuiltIn {
			toolIds = append(toolIds, t.ToolId)
		}
	}
	validToolIds, _ := getValidToolIds(ctx, userId, orgId, toolIds)
	var toolItems []*response.GeneralAgentConfigToolItem
	for _, t := range resp.Config.ToolList {
		if t.ToolType == constant.ToolTypeBuiltIn {
			if validToolIds[t.ToolId] {
				toolItems = append(toolItems, &response.GeneralAgentConfigToolItem{
					ID:   t.ToolId,
					Type: t.ToolType,
				})
			}
		}
	}
	if len(toolItems) > 0 {
		result = append(result, &response.GeneralAgentConfigList{
			ListType: "tool",
			List:     toolItems,
		})
	}

	// 过滤存在的 mcp
	var mcpCustomIds, mcpServerIds []string
	for _, m := range resp.Config.McpList {
		switch m.McpType {
		case constant.MCPTypeMCP:
			mcpCustomIds = append(mcpCustomIds, m.McpId)
		case constant.MCPTypeMCPServer:
			mcpServerIds = append(mcpServerIds, m.McpId)
		}
	}
	validMcpIds, mcpTypes, _ := getValidMcpIds(ctx, mcpCustomIds, mcpServerIds)
	var mcpItems []*response.GeneralAgentConfigItem
	for _, m := range resp.Config.McpList {
		// 验证 MCP 存在且类型匹配
		if validMcpIds[m.McpId] && mcpTypes[m.McpId] == m.McpType {
			mcpItems = append(mcpItems, &response.GeneralAgentConfigItem{
				ID:   m.McpId,
				Type: m.McpType,
			})
		}
	}
	if len(mcpItems) > 0 {
		result = append(result, &response.GeneralAgentConfigList{
			ListType: "mcp",
			List:     mcpItems,
		})
	}

	// 过滤存在的 workflow
	workflowIds := make([]string, 0, len(resp.Config.WorkflowList))
	for _, w := range resp.Config.WorkflowList {
		workflowIds = append(workflowIds, w.WorkflowId)
	}
	validWorkflowIds, _ := getValidWorkflowIds(ctx, workflowIds)
	var workflowItems []*response.GeneralAgentConfigItem
	for _, w := range resp.Config.WorkflowList {
		if validWorkflowIds[w.WorkflowId] {
			workflowItems = append(workflowItems, &response.GeneralAgentConfigItem{
				ID: w.WorkflowId,
			})
		}
	}
	if len(workflowItems) > 0 {
		result = append(result, &response.GeneralAgentConfigList{
			ListType: "workflow",
			List:     workflowItems,
		})
	}

	// 过滤存在的 skill
	var customSkillIds, acquiredSkillIds []string
	for _, s := range resp.Config.SkillList {
		switch s.SkillType {
		case constant.SkillTypeCustom:
			customSkillIds = append(customSkillIds, s.SkillId)
		case constant.SkillTypeAcquired:
			acquiredSkillIds = append(acquiredSkillIds, s.SkillId)
		}
	}
	validCustomSkillIds, _ := getValidSkillIds(ctx, customSkillIds)
	validAcquiredSkillMap, _ := getAcquiredSkillByIDMap(ctx, acquiredSkillIds)
	var skillItems []*response.GeneralAgentConfigItem
	for _, s := range resp.Config.SkillList {
		switch s.SkillType {
		case constant.SkillTypeCustom:
			if validCustomSkillIds[s.SkillId] {
				skillItems = append(skillItems, &response.GeneralAgentConfigItem{
					ID:   s.SkillId,
					Type: s.SkillType,
				})
			}
		case constant.SkillTypeAcquired:
			if _, exists := validAcquiredSkillMap[s.SkillId]; exists {
				skillItems = append(skillItems, &response.GeneralAgentConfigItem{
					ID:   s.SkillId,
					Type: s.SkillType,
				})
			}
		}
	}
	if len(skillItems) > 0 {
		result = append(result, &response.GeneralAgentConfigList{
			ListType: "skill",
			List:     skillItems,
		})
	}

	// 过滤存在的 assistant
	assistantIds := make([]string, 0, len(resp.Config.AssistantList))
	for _, a := range resp.Config.AssistantList {
		assistantIds = append(assistantIds, a.AssistantId)
	}
	validAssistantIds, _, _ := getValidAssistantIds(ctx, userId, orgId, assistantIds)
	var assistantItems []*response.GeneralAgentConfigItem
	for _, a := range resp.Config.AssistantList {
		if validAssistantIds[a.AssistantId] {
			assistantItems = append(assistantItems, &response.GeneralAgentConfigItem{
				ID:   a.AssistantId,
				Type: a.AssistantType,
			})
		}
	}
	if len(assistantItems) > 0 {
		result = append(result, &response.GeneralAgentConfigList{
			ListType: "assistant",
			List:     assistantItems,
		})
	}

	// 过滤存在的 knowledge
	knowledgeIds := make([]string, 0, len(resp.Config.KnowledgeList))
	for _, k := range resp.Config.KnowledgeList {
		knowledgeIds = append(knowledgeIds, k.KnowledgeId)
	}
	validKnowledgeIds, _ := getValidKnowledgeIds(ctx, userId, orgId, knowledgeIds)
	var knowledgeItems []*response.GeneralAgentConfigItem
	for _, k := range resp.Config.KnowledgeList {
		if validKnowledgeIds[k.KnowledgeId] {
			knowledgeItems = append(knowledgeItems, &response.GeneralAgentConfigItem{
				ID: k.KnowledgeId,
			})
		}
	}
	if len(knowledgeItems) > 0 {
		result = append(result, &response.GeneralAgentConfigList{
			ListType: "knowledge",
			List:     knowledgeItems,
		})
	}

	// 过滤存在的 ontology
	ontologyIds := make([]string, 0, len(resp.Config.OntologyKnowledgeList))
	for _, o := range resp.Config.OntologyKnowledgeList {
		ontologyIds = append(ontologyIds, o.OntologyKnowledgeId)
	}
	validOntologyIds, _ := getValidOntologyIds(ctx, userId, orgId, ontologyIds)
	var ontologyItems []*response.GeneralAgentConfigItem
	for _, o := range resp.Config.OntologyKnowledgeList {
		if validOntologyIds[o.OntologyKnowledgeId] {
			ontologyItems = append(ontologyItems, &response.GeneralAgentConfigItem{
				ID: o.OntologyKnowledgeId,
			})
		}
	}
	if len(ontologyItems) > 0 {
		result = append(result, &response.GeneralAgentConfigList{
			ListType: "ontology",
			List:     ontologyItems,
		})
	}

	return result, nil
}

func GeneralAgentWorkspaceDownload(ctx *gin.Context, userId, orgId string, req request.GeneralAgentWorkspaceDownloadReq) (string, []byte, error) {
	store, err := newGeneralAgentWorkspaceStoreByThreadID(ctx, userId, orgId, req.ThreadID)
	if err != nil {
		return "", nil, err
	}
	result, err := DownloadWgaWorkspace(store, req.RunID, req.Path)
	if err != nil {
		return "", nil, err
	}
	return result.FileName, result.Data, nil
}

func GeneralAgentWorkspacePreview(ctx *gin.Context, userId, orgId string, req request.GeneralAgentWorkspacePreviewReq) (string, []byte, string, error) {
	store, err := newGeneralAgentWorkspaceStoreByThreadID(ctx, userId, orgId, req.ThreadID)
	if err != nil {
		return "", nil, "", err
	}
	result, err := PreviewWgaWorkspace(store, req.RunID, req.Path)
	if err != nil {
		return "", nil, "", err
	}
	return result.FileName, result.Data, result.ContentType, nil
}

func GeneralAgentWorkspaceInfo(ctx *gin.Context, userId, orgId string, req request.GeneralAgentWorkspaceReq) (*response.GeneralAgentWorkspaceResp, error) {
	store, err := newGeneralAgentWorkspaceStoreByThreadID(ctx, userId, orgId, req.ThreadID)
	if err != nil {
		return nil, err
	}
	result, err := GetWgaWorkspaceTree(store, req.RunID)
	if err != nil {
		return nil, err
	}

	return &response.GeneralAgentWorkspaceResp{
		GeneralAgentConversationWorkspaceInfo: response.GeneralAgentConversationWorkspaceInfo{
			ThreadID:  req.ThreadID,
			RunID:     req.RunID,
			FileCount: int32(result.FileCount),
			TotalSize: result.TotalSize,
			IsDisplay: true,
		},
		Path:  "",
		Files: result.Files,
	}, nil
}

func newGeneralAgentWorkspaceStoreByThreadID(ctx *gin.Context, userId, orgId, threadID string) (*wga_persistent.Store, error) {
	if !config.WgaCfg().Persistent.Enabled {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "persistent not enabled")
	}

	resp, err := mcp.GetCustomSkillByThreadID(ctx.Request.Context(), &mcp_service.GetCustomSkillByThreadIDReq{
		WgaThreadId: threadID,
	})
	if err != nil {
		return nil, err
	}
	if customSkill := resp.GetSkill(); customSkill != nil {
		return NewGeneralAgentSkillWorkspaceStore(customSkill.SkillId)
	}
	return NewGeneralAgentWorkspaceStore(threadID)
}

// --- internal wga model ---

// checkModelConfig 校验请求中的模型配置（用于创建/更新对话配置）
func checkModelConfig(ctx *gin.Context, modelConfig *request.AppModelConfig) error {
	if modelConfig == nil {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "modelConfig is required")
	}
	if modelConfig.ModelId == "" {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "modelId is required")
	}
	if modelConfig.Model == "" {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "model is required")
	}
	// 校验模型是否存在
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelConfig.ModelId})
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("model not found: %s", modelConfig.ModelId))
	}
	// 校验模型是否已启用
	if !modelInfo.IsActive {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("model is not active: %s", modelConfig.ModelId))
	}
	// 校验 model 名称是否匹配
	if modelInfo.Model != modelConfig.Model {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("model name mismatch: expected %s, got %s", modelInfo.Model, modelConfig.Model))
	}
	return nil
}

// checkModelConfigFromProto 校验proto类型的模型配置（用于运行时检查）
func checkModelConfigFromProto(ctx *gin.Context, modelConfig *common.AppModelConfig) error {
	if modelConfig == nil || modelConfig.ModelId == "" {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "modelConfig is required for conversation")
	}
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelConfig.ModelId})
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("model not found: %s", modelConfig.ModelId))
	}
	// 校验模型是否已启用
	if !modelInfo.IsActive {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("model is not active: %s", modelConfig.ModelId))
	}
	if modelInfo.Model != modelConfig.Model {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("model name mismatch: expected %s, got %s", modelInfo.Model, modelConfig.Model))
	}
	return nil
}

// buildWgaModelOption 构建模型配置选项
func buildWgaModelOption(ctx *gin.Context, modelConfig *common.AppModelConfig) (wga_option.Option, error) {
	if modelConfig == nil || modelConfig.ModelId == "" {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "modelConfig is required")
	}

	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelConfig.ModelId})
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("model not found: %s", modelConfig.ModelId))
	}
	// 校验模型是否已启用
	if !modelInfo.IsActive {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("model is not active: %s", modelConfig.ModelId))
	}

	endpoint := mp.ToModelEndpoint(modelConfig.ModelId, modelConfig.Model)
	modelURL, _ := endpoint["model_url"].(string)

	var apiKey string
	cfg, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err == nil {
		var cfgMap map[string]any
		if b, err := json.Marshal(cfg); err == nil {
			if err = json.Unmarshal(b, &cfgMap); err == nil {
				if k, ok := cfgMap["apiKey"].(string); ok {
					apiKey = k
				}
			}
		}
	}

	var modelParams *mp_common.LLMParams
	if modelConfig.Config != "" {
		llmParams, _, err := mp.ToModelParams(modelConfig.Provider, modelConfig.ModelType, modelConfig.Config)
		if err == nil && llmParams != nil {
			modelParams = llmParams.(*mp_common.LLMParams)
		}
	}

	return wga_option.WithModelConfig(wga_option.ModelConfig{
		Provider:     modelConfig.Provider,
		ProviderName: modelConfig.Provider,
		BaseURL:      modelURL,
		APIKey:       apiKey,
		Model:        modelConfig.Model,
		ModelName:    modelConfig.Model,
		Params:       modelParams,
	}), nil
}

// ResolveModelConfigByUuid 通过模型 UUID 解析为 AppModelConfig。
// 用于 OpenAPI 的创建对话和 chat 接口，统一通过 modelUuid 指定模型。
func ResolveModelConfigByUuid(ctx *gin.Context, modelUuid string) (*request.AppModelConfig, error) {
	modelInfo, err := model.GetModelByUuid(ctx.Request.Context(), &model_service.GetModelByUuidReq{Uuid: modelUuid})
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("model not found by uuid: %s", modelUuid))
	}
	return &request.AppModelConfig{
		Provider:  modelInfo.Provider,
		Model:     modelInfo.Model,
		ModelId:   modelInfo.ModelId,
		ModelType: modelInfo.ModelType,
	}, nil
}

// UpdateConversationModelByUuid 通过模型 UUID 更新对话的模型配置。
// 用于 chat 接口传入 modelUuid 参数时，在对话开始前切换模型。
func UpdateConversationModelByUuid(ctx *gin.Context, userId, orgId, threadId, modelUuid string) error {
	modelConfig, err := ResolveModelConfigByUuid(ctx, modelUuid)
	if err != nil {
		return err
	}
	req := request.UpdateGeneralAgentConversationConfigReq{
		ThreadID:    threadId,
		ModelConfig: modelConfig,
	}
	return UpdateGeneralAgentConversationConfig(ctx, userId, orgId, req)
}
