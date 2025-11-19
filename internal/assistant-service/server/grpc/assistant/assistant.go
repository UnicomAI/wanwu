package assistant

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetAssistantByIds Gets the list of agents based on the agent id set
func (s *Service) GetAssistantByIds(ctx context.Context, req *assistant_service.GetAssistantByIdsReq) (*assistant_service.AppBriefList, error) {
	// Convert string ID to uint32
	var assistantIDs []uint32
	for _, idStr := range req.AssistantIdList {
		if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
			assistantIDs = append(assistantIDs, uint32(id))
		}
	}

	// Call the client method to get the list of agents
	assistants, status := s.cli.GetAssistantsByIDs(ctx, assistantIDs)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// Convert to responsive format
	var appBriefs []*common.AppBrief
	for _, assistant := range assistants {
		appBriefs = append(appBriefs, &common.AppBrief{
			OrgId:      assistant.OrgId,
			UserId:     assistant.UserId,
			AppId:      strconv.FormatUint(uint64(assistant.ID), 10),
			AppType:    "agent",
			Name:       assistant.Name,
			AvatarPath: assistant.AvatarPath,
			Desc:       assistant.Desc,
			CreatedAt:  assistant.CreatedAt,
			UpdatedAt:  assistant.UpdatedAt,
		})
	}

	return &assistant_service.AppBriefList{
		AssistantInfos: appBriefs,
	}, nil
}

// AssistantCreate creates an agent
func (s *Service) AssistantCreate(ctx context.Context, req *assistant_service.AssistantCreateReq) (*assistant_service.AssistantCreateResp, error) {
	// Assemble model parameters
	assistant := &model.Assistant{
		AvatarPath: req.AssistantBrief.AvatarPath,
		Name:       req.AssistantBrief.Name,
		Desc:       req.AssistantBrief.Desc,
		Scope:      1,
		UserId:     req.Identity.UserId,
		OrgId:      req.Identity.OrgId,
	}
	// Find if an agent with the same name exists
	if err := s.cli.CheckSameAssistantName(ctx, req.Identity.UserId, req.Identity.OrgId, req.AssistantBrief.Name, ""); err != nil {
		return nil, errStatus(errs.Code_AssistantErr, err)
	}
	// Call the client method to create an agent
	if status := s.cli.CreateAssistant(ctx, assistant); status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	return &assistant_service.AssistantCreateResp{
		AssistantId: strconv.FormatUint(uint64(assistant.ID), 10),
	}, nil
}

// AssistantUpdate modifies the agent
func (s *Service) AssistantUpdate(ctx context.Context, req *assistant_service.AssistantUpdateReq) (*emptypb.Empty, error) {
	// Conversion ID
	assistantID, err := strconv.ParseUint(req.AssistantId, 10, 32)
	if err != nil {
		return nil, err
	}

	// Get existing agent information
	existingAssistant, status := s.cli.GetAssistant(ctx, uint32(assistantID), "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// Find if an agent with the same name exists
	if err := s.cli.CheckSameAssistantName(ctx, req.Identity.UserId, req.Identity.OrgId, req.AssistantBrief.Name, req.AssistantId); err != nil {
		return nil, errStatus(errs.Code_AssistantErr, err)
	}

	existingAssistant.AvatarPath = req.AssistantBrief.AvatarPath
	existingAssistant.Name = req.AssistantBrief.Name
	existingAssistant.Desc = req.AssistantBrief.Desc

	// Call the client method to update the agent
	if status := s.cli.UpdateAssistant(ctx, existingAssistant); status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	return &emptypb.Empty{}, nil
}

// AssistantDelete deletes the agent
func (s *Service) AssistantDelete(ctx context.Context, req *assistant_service.AssistantDeleteReq) (*emptypb.Empty, error) {
	// Conversion ID
	assistantID, err := strconv.ParseUint(req.AssistantId, 10, 32)
	if err != nil {
		return nil, err
	}

	// Call the client method to delete the agent
	if status := s.cli.DeleteAssistant(ctx, uint32(assistantID)); status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	return &emptypb.Empty{}, nil
}

// AssistantConfigUpdate modifies the agent configuration
func (s *Service) AssistantConfigUpdate(ctx context.Context, req *assistant_service.AssistantConfigUpdateReq) (*emptypb.Empty, error) {
	// Conversion ID
	assistantID, err := strconv.ParseUint(req.AssistantId, 10, 32)
	if err != nil {
		return nil, err
	}

	// First obtain existing agent information
	existingAssistant, status := s.cli.GetAssistant(ctx, uint32(assistantID), "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// Update configuration fields
	existingAssistant.Instructions = req.Instructions
	existingAssistant.Prologue = req.Prologue
	existingAssistant.RecommendQuestion = strings.Join(req.RecommendQuestion, "@#@")

	// Process modelConfig, convert it into a json string and then update it
	if req.ModelConfig != nil {
		modelConfigBytes, err := json.Marshal(req.ModelConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_modelConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.ModelConfig = string(modelConfigBytes)
	}

	// Process rerankConfig, convert it into a json string and then update it
	if req.RerankConfig != nil {
		rerankConfigBytes, err := json.Marshal(req.RerankConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_rerankConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.RerankConfig = string(rerankConfigBytes)
	}

	// Process knowledgeBaseConfig, convert it into a json string and then update it
	if req.KnowledgeBaseConfig != nil {
		knowledgeBaseConfigBytes, err := json.Marshal(req.KnowledgeBaseConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_knowledgeBaseConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.KnowledgebaseConfig = string(knowledgeBaseConfigBytes)
		log.Debugf("knowConfig = %s", existingAssistant.KnowledgebaseConfig)
	}

	// Process safetyConfig, convert it into a json string and then update it
	if req.SafetyConfig != nil {
		safetyConfigBytes, err := json.Marshal(req.SafetyConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_safetyConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.SafetyConfig = string(safetyConfigBytes)
	}

	// Process visionConfig, convert it into a json string and then update it
	if req.VisionConfig != nil {
		visionConfigBytes, err := json.Marshal(req.VisionConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_visionConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.VisionConfig = string(visionConfigBytes)
	}

	// Call the client method to update the agent
	if status := s.cli.UpdateAssistant(ctx, existingAssistant); status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	return &emptypb.Empty{}, nil
}

// GetAssistantListMyAll agent list
func (s *Service) GetAssistantListMyAll(ctx context.Context, req *assistant_service.GetAssistantListMyAllReq) (*assistant_service.AppBriefList, error) {
	// Call the client method to get the list of agents
	assistants, _, status := s.cli.GetAssistantList(ctx, req.Identity.UserId, req.Identity.OrgId, req.Name)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// Convert to responsive format
	var appBriefs []*common.AppBrief
	for _, assistant := range assistants {
		appBriefs = append(appBriefs, &common.AppBrief{
			OrgId:      assistant.OrgId,
			UserId:     assistant.UserId,
			AppId:      strconv.FormatUint(uint64(assistant.ID), 10),
			AppType:    "agent",
			Name:       assistant.Name,
			AvatarPath: assistant.AvatarPath,
			Desc:       assistant.Desc,
			CreatedAt:  assistant.CreatedAt,
			UpdatedAt:  assistant.UpdatedAt,
		})
	}

	return &assistant_service.AppBriefList{
		AssistantInfos: appBriefs,
	}, nil
}

// GetAssistantInfo View agent details
func (s *Service) GetAssistantInfo(ctx context.Context, req *assistant_service.GetAssistantInfoReq) (*assistant_service.AssistantInfo, error) {
	// Conversion ID
	assistantId, err := util.U32(req.AssistantId)
	if err != nil {
		return nil, err
	}

	// Null judgment processing, using different parameters depending on whether the Identity is empty
	var assistant *model.Assistant
	var status *errs.Status
	if req.Identity == nil {
		assistant, status = s.cli.GetAssistant(ctx, assistantId, "", "")
	} else {
		assistant, status = s.cli.GetAssistant(ctx, assistantId, req.Identity.UserId, req.Identity.OrgId)
	}
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// Get associated WorkFlows
	workflows, _ := s.cli.GetAssistantWorkflowsByAssistantID(ctx, assistantId)

	// Convert WorkFlows
	var workFlowInfos []*assistant_service.AssistantWorkFlowInfos
	for _, workflow := range workflows {
		workFlowInfos = append(workFlowInfos, &assistant_service.AssistantWorkFlowInfos{
			Id:         strconv.FormatUint(uint64(workflow.ID), 10),
			WorkFlowId: workflow.WorkflowId,
			Enable:     workflow.Enable,
		})
	}

	// Get associated MCP
	mcps, _ := s.cli.GetAssistantMCPList(ctx, assistantId)
	// Convert MCP
	var mcpInfos []*assistant_service.AssistantMCPInfos
	for _, mcp := range mcps {
		mcpInfos = append(mcpInfos, &assistant_service.AssistantMCPInfos{
			Id:         strconv.FormatUint(uint64(mcp.ID), 10),
			McpId:      mcp.MCPId,
			McpType:    mcp.MCPType,
			ActionName: mcp.ActionName,
			Enable:     mcp.Enable,
		})
	}

	// Get the associated Tool
	tools, _ := s.cli.GetAssistantToolList(ctx, assistantId)
	// Convert Tool
	var toolInfos []*assistant_service.AssistantToolInfos
	for _, tool := range tools {
		toolInfos = append(toolInfos, &assistant_service.AssistantToolInfos{
			Id:         strconv.FormatUint(uint64(tool.ID), 10),
			ToolId:     tool.ToolId,
			ToolType:   tool.ToolType,
			ActionName: tool.ActionName,
			Enable:     tool.Enable,
			ToolConfig: tool.ToolConfig,
		})
	}

	// Process assistant.ModelConfig and convert to common.AppModelConfig
	var modelConfig *common.AppModelConfig
	if assistant.ModelConfig != "" {
		modelConfig = &common.AppModelConfig{}
		if err := json.Unmarshal([]byte(assistant.ModelConfig), modelConfig); err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_modelConfig_unmarshal",
				Args:    []string{err.Error()},
			})
		}
	}

	// Process assistant.RerankConfig and convert to common.AppModelConfig
	var rerankConfig *common.AppModelConfig
	if assistant.RerankConfig != "" {
		rerankConfig = &common.AppModelConfig{}
		if err := json.Unmarshal([]byte(assistant.RerankConfig), rerankConfig); err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_rerankConfig_unmarshal",
				Args:    []string{err.Error()},
			})
		}
	}

	// Process assistant.KnowledgebaseConfig and convert to AssistantKnowledgeBaseConfig
	var knowledgeBaseConfig *assistant_service.AssistantKnowledgeBaseConfig
	if assistant.KnowledgebaseConfig != "" {
		knowledgeBaseConfig = &assistant_service.AssistantKnowledgeBaseConfig{}
		if err := json.Unmarshal([]byte(assistant.KnowledgebaseConfig), knowledgeBaseConfig); err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_knowledgeBaseConfig_unmarshal",
				Args:    []string{err.Error()},
			})
		}
	}

	// Process assistant.SafetyConfig and convert to AssistantSafetyConfig
	var safetyConfig *assistant_service.AssistantSafetyConfig
	if assistant.SafetyConfig != "" {
		safetyConfig = &assistant_service.AssistantSafetyConfig{}
		if err := json.Unmarshal([]byte(assistant.SafetyConfig), safetyConfig); err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_safetyConfig_unmarshal",
				Args:    []string{err.Error()},
			})
		}
	}

	// Process assistant.VisionConfig and convert to AssistantVisionConfig
	var visionConfig *assistant_service.AssistantVisionConfig
	if assistant.VisionConfig != "" {
		visionConfig = &assistant_service.AssistantVisionConfig{}
		if err := json.Unmarshal([]byte(assistant.VisionConfig), visionConfig); err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_visionConfig_unmarshal",
				Args:    []string{err.Error()},
			})
		}
		visionConfig.MaxPicNum = config.Cfg().Assistant.MaxPicNum
	}

	return &assistant_service.AssistantInfo{
		AssistantId: strconv.FormatUint(uint64(assistant.ID), 10),
		Identity: &assistant_service.Identity{
			UserId: assistant.UserId,
			OrgId:  assistant.OrgId,
		},
		AssistantBrief: &common.AppBriefConfig{
			Name:       assistant.Name,
			AvatarPath: assistant.AvatarPath,
			Desc:       assistant.Desc,
		},
		Prologue:            assistant.Prologue,
		Instructions:        assistant.Instructions,
		RecommendQuestion:   strings.Split(assistant.RecommendQuestion, "@#@"),
		ModelConfig:         modelConfig,
		KnowledgeBaseConfig: knowledgeBaseConfig,
		RerankConfig:        rerankConfig,
		SafetyConfig:        safetyConfig,
		VisionConfig:        visionConfig,
		Scope:               int32(assistant.Scope),
		WorkFlowInfos:       workFlowInfos,
		McpInfos:            mcpInfos,
		ToolInfos:           toolInfos,
		CreatTime:           assistant.CreatedAt,
		UpdateTime:          assistant.UpdatedAt,
	}, nil
}

func (s *Service) AssistantCopy(ctx context.Context, req *assistant_service.AssistantCopyReq) (*assistant_service.AssistantCreateResp, error) {
	assistantId, err := util.U32(req.AssistantId)
	if err != nil {
		return nil, err
	}

	// Get parent agent information
	parentAssistant, status := s.cli.GetAssistant(ctx, assistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// Get the associated workflow
	workflows, status := s.cli.GetAssistantWorkflowsByAssistantID(ctx, assistantId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// Get associated mcp
	mcps, status := s.cli.GetAssistantMCPList(ctx, assistantId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// Get the associated tool
	tools, status := s.cli.GetAssistantToolList(ctx, assistantId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// Copy agent
	assistantID, status := s.cli.CopyAssistant(ctx, parentAssistant, workflows, mcps, tools)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}
	return &assistant_service.AssistantCreateResp{
		AssistantId: util.Int2Str(assistantID),
	}, nil
}
