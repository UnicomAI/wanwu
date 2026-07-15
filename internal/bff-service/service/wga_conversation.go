package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga"
	wga_persistent "github.com/UnicomAI/wanwu/pkg/wga-persistent"
	wga_sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/gin-gonic/gin"
)

const (
	generalAgentSkillChatModeNormal  = "normal"
	generalAgentSkillChatModeImport  = "import"
	generalAgentSkillChatModeConvert = "convert"
	generalAgentSkillChatModePreview = "preview"

	generalAgentSkillChatNormalAgentID  = "Skill Chat Agent"
	generalAgentSkillChatImportAgentID  = "Skill Import Agent"
	generalAgentSkillChatPreviewAgentID = "Skill Preview Agent"
)

func CreateGeneralAgentConversation(ctx *gin.Context, userId, orgId string, req request.CreateGeneralAgentConversationReq) (*response.CreateGeneralAgentConversationResp, error) {
	if err := checkModelConfig(ctx, req.ModelConfig); err != nil {
		return nil, err
	}

	resp, err := assistant.WgaConversationCreate(ctx.Request.Context(), &assistant_service.WgaConversationCreateReq{
		Prompt: req.Title,
		ModelConfig: &common.AppModelConfig{
			ModelId:   req.ModelConfig.ModelId,
			Provider:  req.ModelConfig.Provider,
			Model:     req.ModelConfig.Model,
			ModelType: req.ModelConfig.ModelType,
		},
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.CreateGeneralAgentConversationResp{ThreadID: resp.ThreadId}, nil
}

func CreateGeneralAgentSkillConversation(ctx *gin.Context, userId, orgId string, req request.CreateGeneralAgentSkillConversationReq) (*response.CreateGeneralAgentSkillConversationResp, error) {
	generalConversationResp, err := CreateGeneralAgentConversation(ctx, userId, orgId, request.CreateGeneralAgentConversationReq(req))
	if err != nil {
		return nil, err
	}

	previewID := util.GenUUID()
	customSkillResp, err := mcp.CustomSkillCreate(ctx.Request.Context(), &mcp_service.CustomSkillCreateReq{
		Name:            req.Title,
		Author:          getUserNameById(ctx, userId),
		WgaThreadId:     generalConversationResp.ThreadID,
		PreviewThreadId: previewID,
		Identity:        &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		_, _ = assistant.WgaConversationDelete(ctx.Request.Context(), &assistant_service.WgaConversationDeleteReq{
			ThreadId: generalConversationResp.ThreadID,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		return nil, err
	}

	return &response.CreateGeneralAgentSkillConversationResp{
		CustomSkillID: customSkillResp.SkillId,
		ThreadID:      generalConversationResp.ThreadID,
		PreviewID:     previewID,
	}, nil
}

func RefreshGeneralAgentSkillConversation(ctx *gin.Context, userId, orgId string, req request.RefreshGeneralAgentSkillConversationReq) (*response.RefreshGeneralAgentSkillConversationResp, error) {
	publish, err := mcp.CustomSkillGet(ctx.Request.Context(), &mcp_service.CustomSkillGetReq{
		SkillId: req.SkillID,
	})
	if err != nil {
		return nil, err
	}
	skill := publish.GetSkill()
	if skill == nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_skill_custom_not_found")
	}

	log.Infof("[wga-skill-legacy] refresh skill %v start, objectPathExists=%v, currentThreadID=%s, currentPreviewID=%s", req.SkillID, strings.TrimSpace(skill.ObjectPath) != "", skill.WgaThreadId, skill.PreviewThreadId)
	if err := ensureLegacyCustomSkillWorkspace(ctx, skill); err != nil {
		return nil, err
	}

	title := strings.TrimSpace(skill.Name)
	if title == "" {
		title = "Skill Conversation"
	}

	threadResp, err := assistant.WgaConversationCreate(ctx.Request.Context(), &assistant_service.WgaConversationCreateReq{
		Prompt: title,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	previewID := strings.TrimSpace(skill.PreviewThreadId)
	if previewID == "" {
		previewID = util.GenUUID()
	}
	_, err = mcp.UpdateCustomSkillThreadMeta(ctx.Request.Context(), &mcp_service.UpdateCustomSkillThreadMetaReq{
		SkillId:         req.SkillID,
		WgaThreadId:     threadResp.ThreadId,
		PreviewThreadId: previewID,
	})
	if err != nil {
		_, _ = assistant.WgaConversationDelete(ctx.Request.Context(), &assistant_service.WgaConversationDeleteReq{
			ThreadId: threadResp.ThreadId,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		return nil, err
	}
	log.Infof("[wga-skill-legacy] refresh skill %v success, newThreadID=%s, previewID=%s", req.SkillID, threadResp.ThreadId, previewID)

	return &response.RefreshGeneralAgentSkillConversationResp{
		CustomSkillID: req.SkillID,
		ThreadID:      threadResp.ThreadId,
		PreviewID:     previewID,
	}, nil
}

func DeleteGeneralAgentConversation(ctx *gin.Context, userId, orgId string, req request.DeleteGeneralAgentConversationReq) error {
	threadID := strings.TrimSpace(req.ThreadID)
	binding, err := getCustomSkillThreadBinding(ctx, userId, orgId, threadID)
	if err != nil {
		return err
	}
	if binding != nil {
		if err := clearCustomSkillThreadBinding(ctx, binding); err != nil {
			return err
		}
		if err := deleteWgaConversationHistory(ctx, userId, orgId, binding.previewThreadID); err != nil {
			return err
		}
	}
	return deleteGeneralAgentConversationCore(ctx, userId, orgId, threadID)
}

func deleteGeneralAgentConversationForSkillDelete(ctx *gin.Context, userId, orgId string, req request.DeleteGeneralAgentConversationReq) error {
	return deleteGeneralAgentConversationCore(ctx, userId, orgId, strings.TrimSpace(req.ThreadID))
}

type customSkillThreadBinding struct {
	threadID        string
	skillID         string
	previewThreadID string
}

func getCustomSkillThreadBinding(ctx *gin.Context, userId, orgId, threadID string) (*customSkillThreadBinding, error) {
	threadID = strings.TrimSpace(threadID)
	if threadID == "" {
		return nil, nil
	}

	resp, err := mcp.GetCustomSkillByThreadID(ctx.Request.Context(), &mcp_service.GetCustomSkillByThreadIDReq{
		WgaThreadId: threadID,
	})
	if err != nil {
		return nil, err
	}

	skill := resp.GetSkill()
	if skill == nil || strings.TrimSpace(skill.WgaThreadId) != threadID {
		return nil, nil
	}
	skillID := strings.TrimSpace(skill.SkillId)
	if skillID == "" {
		return nil, nil
	}
	return &customSkillThreadBinding{
		threadID:        threadID,
		skillID:         skillID,
		previewThreadID: strings.TrimSpace(skill.PreviewThreadId),
	}, nil
}

func clearCustomSkillThreadBinding(ctx *gin.Context, binding *customSkillThreadBinding) error {
	if binding == nil || binding.threadID == "" || binding.skillID == "" {
		return nil
	}
	_, err := mcp.UpdateCustomSkillThreadMeta(ctx.Request.Context(), &mcp_service.UpdateCustomSkillThreadMetaReq{
		SkillId:         binding.skillID,
		WgaThreadId:     "",
		PreviewThreadId: "",
	})
	if err != nil {
		return err
	}
	return nil
}

func deleteGeneralAgentConversationCore(ctx *gin.Context, userId, orgId, threadID string) error {
	_, err := assistant.WgaConversationDelete(ctx.Request.Context(), &assistant_service.WgaConversationDeleteReq{
		ThreadId: threadID,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return err
	}

	_, err = assistant.DeleteFromES(ctx.Request.Context(), &assistant_service.DeleteFromESReq{
		IndexName: wgaConversationHistoryEventESIndexName,
		Conditions: map[string]string{
			"threadId": threadID,
			"userId":   userId,
			"orgId":    orgId,
		},
	})
	if err != nil && !wgaConversationHistoryEventESIndexNotFound(err) {
		log.Errorf("[wga] thread %v delete chat history from ES err: %v", threadID, err)
	}

	return nil
}

func GetGeneralAgentConversationList(ctx *gin.Context, userId, orgId string, req request.GetGeneralAgentConversationListReq) (*response.ListResult, error) {
	resp, err := assistant.WgaConversationList(ctx.Request.Context(), &assistant_service.WgaConversationListReq{
		PageSize:   int32(req.PageSize),
		PageNo:     int32(req.PageNo),
		SearchText: strings.TrimSpace(req.SearchText),
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	threadIDs := collectWgaConversationThreadIDs(resp.Data)
	skillByThreadID, err := getGeneralAgentSkillConversationMap(ctx, userId, orgId, threadIDs)
	if err != nil {
		return nil, err
	}

	result := make([]response.GeneralAgentConversationInfo, 0, len(resp.Data))
	for _, info := range resp.Data {
		skill := skillByThreadID[info.ThreadId]
		item := response.GeneralAgentConversationInfo{
			ThreadID:            info.ThreadId,
			Title:               info.Title,
			CreatedAt:           util.Time2Str(info.CreatedAt),
			UpdatedAt:           util.Time2Str(info.UpdatedAt),
			IsSkillConversation: skill != nil,
		}
		if skill != nil {
			item.SkillID = skill.SkillId
			item.PreviewID = skill.PreviewThreadId
		}
		result = append(result, item)
	}
	return &response.ListResult{List: result, Total: resp.Total}, nil
}

func getGeneralAgentSkillConversationMap(ctx *gin.Context, userId, orgId string, threadIDs []string) (map[string]*mcp_service.CustomSkill, error) {
	skillByThreadID := make(map[string]*mcp_service.CustomSkill)
	if len(threadIDs) == 0 {
		return skillByThreadID, nil
	}

	resp, err := mcp.GetCustomSkillListByThreadIDList(ctx.Request.Context(), &mcp_service.GetCustomSkillListByThreadIDListReq{
		WgaThreadIdList: threadIDs,
		Identity:        &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return skillByThreadID, nil
	}
	for _, publish := range resp.List {
		skill := publish.GetSkill()
		if skill == nil || strings.TrimSpace(skill.WgaThreadId) == "" || strings.TrimSpace(skill.SkillId) == "" {
			continue
		}
		skillByThreadID[skill.WgaThreadId] = skill
	}
	return skillByThreadID, nil
}

func collectWgaConversationThreadIDs(conversations []*assistant_service.WgaConversationInfo) []string {
	threadIDs := make([]string, 0, len(conversations))
	seen := make(map[string]struct{}, len(conversations))
	for _, conversation := range conversations {
		if conversation == nil {
			continue
		}
		threadID := strings.TrimSpace(conversation.ThreadId)
		if threadID == "" {
			continue
		}
		if _, ok := seen[threadID]; ok {
			continue
		}
		seen[threadID] = struct{}{}
		threadIDs = append(threadIDs, threadID)
	}
	return threadIDs
}

func GetGeneralAgentConversationDetail(ctx *gin.Context, userId, orgId, threadId string) (*response.ListResult, error) {
	exist, err := assistant.WgaConversationExists(ctx.Request.Context(), &assistant_service.WgaConversationExistsReq{
		ThreadId: threadId,
		Identity: &assistant_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}
	if !exist.Exists {
		return &response.ListResult{}, nil
	}

	return getWgaConversationDetailFromES(ctx, userId, orgId, threadId)
}

func GetGeneralAgentSkillPreviewConversationDetail(ctx *gin.Context, userId, orgId string, req request.GetGeneralAgentSkillPreviewConversationDetailReq) (*response.ListResult, error) {
	return getWgaConversationDetailFromES(ctx, userId, orgId, req.PreviewID)
}

func getWgaConversationDetailFromES(ctx *gin.Context, userId, orgId, threadId string) (*response.ListResult, error) {
	resp, err := assistant.SearchFromES(ctx.Request.Context(), &assistant_service.SearchFromESReq{
		IndexName: wgaConversationHistoryEventESIndexName,
		Conditions: map[string]string{
			"threadId": threadId,
			"userId":   userId,
			"orgId":    orgId,
		},
		SortOrder: "asc",
		PageNo:    1,
		PageSize:  1000,
	})
	if err != nil {
		if wgaConversationHistoryEventESIndexNotFound(err) {
			return &response.ListResult{}, nil
		}
		return nil, err
	}

	result := make([]response.GeneralAgentConversationDetailInfo, 0, len(resp.DocJsonList))
	for _, docJson := range resp.DocJsonList {
		var doc map[string]interface{}
		if err := json.Unmarshal([]byte(docJson), &doc); err != nil {
			continue
		}

		createdAt, _ := doc["createdAt"].(int64)
		runId, _ := doc["runId"].(string)

		info := response.GeneralAgentConversationDetailInfo{
			ThreadID:  threadId,
			RunID:     runId,
			CreatedAt: createdAt,
		}
		if eventsStr, ok := doc["events"].(string); ok {
			var events []interface{}
			if err := json.Unmarshal([]byte(eventsStr), &events); err != nil {
				log.Errorf("[wga] thread %v unmarshal events err: %v", threadId, err)
				continue
			}
			info.Events = events
		}
		result = append(result, info)
	}

	return &response.ListResult{List: result, Total: int64(len(result))}, nil
}

// GetGeneralAgentConversationConfig 获取WGA对话配置
func GetGeneralAgentConversationConfig(ctx *gin.Context, userId, orgId string, req request.GetGeneralAgentConversationConfigReq) (*response.GetGeneralAgentConversationConfigResp, error) {
	resp, err := assistant.GetWgaConversationConfig(ctx.Request.Context(), &assistant_service.GetWgaConversationConfigReq{
		ThreadId: req.ThreadID,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	wgaConfig := resp.Config
	result := &response.GetGeneralAgentConversationConfigResp{
		ThreadID:    wgaConfig.ThreadId,
		ModelConfig: request.AppModelConfig{},
	}

	// 处理模型配置 - 需要验证模型是否存在且已启用
	if wgaConfig.ModelConfig != nil && wgaConfig.ModelConfig.ModelId != "" {
		// 验证模型是否存在
		modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: wgaConfig.ModelConfig.ModelId})
		if err == nil && modelInfo != nil && modelInfo.IsActive {
			// 模型存在且已启用，返回配置
			result.ModelConfig = request.AppModelConfig{
				Provider:    wgaConfig.ModelConfig.Provider,
				Model:       wgaConfig.ModelConfig.Model,
				ModelId:     wgaConfig.ModelConfig.ModelId,
				ModelType:   wgaConfig.ModelConfig.ModelType,
				DisplayName: modelInfo.DisplayName,
			}
		}
		// 如果模型不存在或未启用，返回空的 ModelConfig
	}

	return result, nil
}

func UpdateGeneralAgentConversationConfig(ctx *gin.Context, userId, orgId string, req request.UpdateGeneralAgentConversationConfigReq) error {
	if err := checkModelConfig(ctx, req.ModelConfig); err != nil {
		return err
	}
	_, err := assistant.UpdateWgaConversationConfig(ctx.Request.Context(), &assistant_service.UpdateWgaConversationConfigReq{
		ThreadId: req.ThreadID,
		ModelConfig: &common.AppModelConfig{
			ModelId:   req.ModelConfig.ModelId,
			Provider:  req.ModelConfig.Provider,
			Model:     req.ModelConfig.Model,
			ModelType: req.ModelConfig.ModelType,
		},
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	return err
}

// CheckGeneralAgentConversationConfig 检查WGA对话配置
func CheckGeneralAgentConversationConfig(ctx *gin.Context, userId, orgId string, req request.GeneralAgentConfigCheckRequest) (*response.GeneralAgentConfigCheckResponse, error) {
	// 查询对话配置
	conversationConfigResp, err := assistant.GetWgaConversationConfig(ctx.Request.Context(), &assistant_service.GetWgaConversationConfigReq{
		ThreadId: req.ThreadID,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	wgaConversationConfig := conversationConfigResp.Config

	// 查询配置
	configResp, err := assistant.GetWgaConfig(ctx.Request.Context(), &assistant_service.GetWgaConfigReq{
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}
	wgaConfig := configResp.Config

	var opts []wga_option.Option

	// 构建模型配置选项
	if wgaConversationConfig.GetModelConfig() != nil && wgaConversationConfig.ModelConfig.ModelId != "" {
		modelOpt, err := buildWgaModelOption(ctx, wgaConversationConfig.ModelConfig)
		if err != nil {
			return nil, err
		}
		opts = append(opts, modelOpt)
	}

	if req.AgentID == "" {
		// 如果agentId为空，则不限制工具配置，直接返回模型检查结果
		return &response.GeneralAgentConfigCheckResponse{
			Meet:      true,
			ModelMeet: true,
		}, nil
	}

	// 构建工具配置选项
	toolOpts, err := buildWgaToolOptions(ctx, userId, orgId, wgaConfig.ToolList)
	if err != nil {
		return nil, err
	}
	opts = append(opts, toolOpts...)

	// 检查工具配置
	checkResult, err := wga.CheckToolOptions(ctx.Request.Context(), req.AgentID, opts...)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("check tool options err: %v", err))
	}

	result := &response.GeneralAgentConfigCheckResponse{
		ModelMeet: true,
		ToolsMeet: make([]response.GeneralAgentToolCategories, 0, len(checkResult.ToolCategories)),
	}

	// 处理工具检查结果
	for _, tc := range checkResult.ToolCategories {
		category := response.GeneralAgentToolCategories{
			Category:  tc.Category,
			Condition: tc.Condition,
			Meet:      tc.Meet,
			Tools:     make([]response.GeneralAgentCheckTool, 0, len(tc.Tools)),
		}
		for _, t := range tc.Tools {
			category.Tools = append(category.Tools, response.GeneralAgentCheckTool{
				ToolID: t.Title,
				Meet:   t.Meet,
			})
		}
		result.ToolsMeet = append(result.ToolsMeet, category)
	}

	result.Meet = result.ModelMeet
	for _, tc := range result.ToolsMeet {
		if !tc.Meet {
			result.Meet = false
			break
		}
	}

	return result, nil
}

// GeneralAgentConversationChat 通用智能体对话接口
func GeneralAgentConversationChat(ctx *gin.Context, userId, orgId, clientId string, req request.GeneralAgentConversationChatReq, enableHumanInTheLoop bool) error {
	// 获取 threadId 的 ModelConfig
	configResp, err := assistant.GetWgaConversationConfig(ctx.Request.Context(), &assistant_service.GetWgaConversationConfigReq{
		ThreadId: req.ThreadID,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return err
	}
	var modelConfig *common.AppModelConfig
	if configResp.Config != nil {
		modelConfig = configResp.Config.ModelConfig
	}

	// 获取 threadId 的 workspace store
	var workspaceStore *wga_persistent.Store
	if config.WgaCfg().Persistent.Enabled {
		store, err := NewGeneralAgentWorkspaceStore(req.ThreadID)
		if err != nil {
			log.Errorf("[wga] thread %v failed to create persistent store: %v", req.ThreadID, err)
		} else {
			workspaceStore = store
		}
	}

	agentID := req.AgentID
	if agentID == "" {
		agentID = config.WgaCfg().AgentID
	}

	return WgaConversationChat(ctx, &WgaChatParams{
		UserID:               userId,
		OrgID:                orgId,
		AgentID:              agentID,
		ThreadID:             req.ThreadID,
		Messages:             req.Messages,
		ClientID:             clientId,
		ModelConfig:          modelConfig,
		WorkspaceStore:       workspaceStore,
		SendWorkspaceEvent:   true,
		EnableHumanInTheLoop: enableHumanInTheLoop,
	})
}

// GeneralAgentSkillConversationChat Skill对话接口
// 根据 mode 选择对应 Skill Agent，workspace 使用 overwrite 模式
func GeneralAgentSkillConversationChat(ctx *gin.Context, userId, orgId, clientId string, req request.GeneralAgentSkillConversationChatReq) error {
	mode := strings.TrimSpace(strings.ToLower(req.Mode))
	chatThreadID := req.ThreadID
	if mode == generalAgentSkillChatModePreview {
		if strings.TrimSpace(req.PreviewID) == "" {
			return grpc_util.ErrorStatus(errs.Code_BFFGeneral, "previewId is required")
		}
		chatThreadID = req.PreviewID
	}

	// 使用 customSkillID 创建 overwrite 模式的 workspace store
	var workspaceStore *wga_persistent.Store
	if config.WgaCfg().Persistent.Enabled {
		store, err := NewGeneralAgentSkillWorkspaceStore(req.CustomSkillID)
		if err != nil {
			log.Errorf("[wga-skill] customSkillID %v failed to create persistent store: %v", req.CustomSkillID, err)
			return err
		} else {
			workspaceStore = store
		}
	}

	var agentID string
	switch mode {
	case "", generalAgentSkillChatModeNormal:
		agentID = generalAgentSkillChatNormalAgentID
	case generalAgentSkillChatModeImport, generalAgentSkillChatModeConvert:
		agentID = generalAgentSkillChatImportAgentID
	case generalAgentSkillChatModePreview:
		agentID = generalAgentSkillChatPreviewAgentID
	default:
		return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("unsupported skill chat mode: %s", mode))
	}

	configResp, err := assistant.GetWgaConversationConfig(ctx.Request.Context(), &assistant_service.GetWgaConversationConfigReq{
		ThreadId: req.ThreadID,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return err
	}
	var modelConfig *common.AppModelConfig
	if configResp.Config != nil {
		modelConfig = configResp.Config.ModelConfig
	}

	if err := WgaConversationChat(ctx, &WgaChatParams{
		UserID:               userId,
		OrgID:                orgId,
		AgentID:              agentID,
		ThreadID:             chatThreadID,
		Messages:             req.Messages,
		ClientID:             clientId,
		ModelConfig:          modelConfig,
		WorkspaceStore:       workspaceStore,
		WorkspaceReadOnly:    mode == generalAgentSkillChatModePreview,
		EnableHumanInTheLoop: true,
	}); err != nil {
		return err
	}
	if mode != generalAgentSkillChatModePreview && workspaceStore != nil {
		if err := normalizeCustomSkillWorkspaceNestedSkill(req.CustomSkillID); err != nil {
			log.Errorf("[wga-skill] customSkillID %v normalize nested skill workspace err: %v", req.CustomSkillID, err)
			return grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("normalize nested skill workspace failed: %v", err))
		}
	}
	if mode != generalAgentSkillChatModePreview {
		detachedCtx := trace_util.DetachContext(ctx.Request.Context())
		go func() {
			defer util.PrintPanicStack()
			ctx, cancel := context.WithTimeout(detachedCtx, 30*time.Second)
			defer cancel()
			updateCustomSkillMetaFromWorkspace(ctx, req.CustomSkillID)
		}()
	}
	return nil
}

func updateCustomSkillMetaFromWorkspace(ctx context.Context, customSkillID string) {
	frontMatter, err := findGeneratedSkillFrontMatter(customSkillID)
	if err != nil {
		log.Warnf("[wga-skill] customSkillID %v generated skill metadata not ready: %s", customSkillID, formatGeneratedSkillMetaError(err))
		return
	}
	_, err = mcp.UpdateCustomSkillBasicMeta(ctx, &mcp_service.UpdateCustomSkillBasicMetaReq{
		SkillId: customSkillID,
		Name:    frontMatter.Name,
		Desc:    frontMatter.Description,
	})
	if err != nil {
		log.Warnf("[wga-skill] customSkillID %v update generated skill metadata err: %v", customSkillID, err)
		return
	}
	log.Infof("[wga-skill] customSkillID %v updated generated skill metadata from workspace", customSkillID)
}

func GeneralAgentReplyQuestion(ctx *gin.Context, runID string, questionID string, answers [][]string) error {
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	sandboxCfg, err := getWgaSandboxConfig()
	if err != nil {
		return err
	}

	if config.Cfg().Ontology.Enable != 0 {
		go func() {
			defer util.PrintPanicStack()
			ontologySandboxCfg, err := getWgaSandboxOntologyConfig()
			if err != nil {
				log.Errorf("get ontology sandbox config failed: %v", err)
				return
			}
			if err := wga_sandbox.ReplyQuestion(detachedCtx, ontologySandboxCfg, runID, questionID, answers); err != nil {
				log.Warnf("reply ontology question failed, runID: %s, questionID: %s, err: %v", runID, questionID, err)
			}
		}()
	}

	return wga_sandbox.ReplyQuestion(ctx.Request.Context(), sandboxCfg, runID, questionID, answers)
}

func GeneralAgentRejectQuestion(ctx *gin.Context, runID string, questionID string) error {
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	sandboxCfg, err := getWgaSandboxConfig()
	if err != nil {
		return err
	}

	if config.Cfg().Ontology.Enable != 0 {
		go func() {
			defer util.PrintPanicStack()
			ontologySandboxCfg, err := getWgaSandboxOntologyConfig()
			if err != nil {
				log.Errorf("get ontology sandbox config failed: %v", err)
				return
			}
			if err := wga_sandbox.RejectQuestion(detachedCtx, ontologySandboxCfg, runID, questionID); err != nil {
				log.Warnf("reject ontology question failed, runID: %s, questionID: %s, err: %v", runID, questionID, err)
			}
		}()
	}

	return wga_sandbox.RejectQuestion(ctx.Request.Context(), sandboxCfg, runID, questionID)
}
