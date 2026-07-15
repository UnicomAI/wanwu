package assistant

import (
	"context"
	"encoding/json"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetWgaConversationConfig(ctx context.Context, req *assistant_service.GetWgaConversationConfigReq) (*assistant_service.GetWgaConversationConfigResp, error) {
	if req.ThreadId == "" {
		return nil, errStatus(errs.Code_WgaConversationGetErr, toErrStatus("wga_conversation_get", "thread_id is required"))
	}
	if req.Identity == nil {
		return nil, errStatus(errs.Code_WgaConversationGetErr, toErrStatus("wga_conversation_get", "identity is required"))
	}

	config, status := s.cli.GetWgaConversationConfig(ctx, req.ThreadId, req.Identity.UserId, req.Identity.OrgId)
	if status != nil {
		log.Errorf("获取WGA对话配置失败，conversationId: %s, error: %v", req.ThreadId, status)
		return nil, errStatus(errs.Code_WgaConversationGetErr, status)
	}

	return &assistant_service.GetWgaConversationConfigResp{
		Config: toProtoWgaConversationConfig(config),
	}, nil
}

func (s *Service) UpdateWgaConversationConfig(ctx context.Context, req *assistant_service.UpdateWgaConversationConfigReq) (*emptypb.Empty, error) {
	if req.ThreadId == "" {
		return nil, errStatus(errs.Code_WgaConversationUpdateErr, toErrStatus("wga_conversation_update", "thread_id is required"))
	}
	if req.Identity == nil {
		return nil, errStatus(errs.Code_WgaConversationUpdateErr, toErrStatus("wga_conversation_update", "identity is required"))
	}

	var modelConfig string
	if req.ModelConfig != nil {
		modelConfigBytes, _ := json.Marshal(req.ModelConfig)
		modelConfig = string(modelConfigBytes)
	}

	config := &model.WgaConversationConfig{
		ThreadID:    req.ThreadId,
		UserID:      req.Identity.UserId,
		OrgID:       req.Identity.OrgId,
		ModelConfig: modelConfig,
	}

	status := s.cli.UpdateWgaConversationConfig(ctx, config)
	if status != nil {
		log.Errorf("更新WGA对话配置失败，conversationId: %s, error: %v", req.ThreadId, status)
		return nil, errStatus(errs.Code_WgaConversationUpdateErr, status)
	}

	return &emptypb.Empty{}, nil
}

func (s *Service) GetWgaConfig(ctx context.Context, req *assistant_service.GetWgaConfigReq) (*assistant_service.GetWgaConfigResp, error) {
	if req.Identity == nil {
		return nil, errStatus(errs.Code_WgaConfigGetErr, toErrStatus("wga_config_get", "identity is required"))
	}

	config, status := s.cli.GetWgaConfig(ctx, req.Identity.UserId, req.Identity.OrgId)
	if status != nil {
		log.Errorf("获取WGA配置失败，userId: %s, orgId: %s, error: %v", req.Identity.UserId, req.Identity.OrgId, status)
		return nil, errStatus(errs.Code_WgaConfigGetErr, status)
	}

	return &assistant_service.GetWgaConfigResp{
		Config: toProtoWgaConfig(config),
	}, nil
}

func (s *Service) UpdateWgaConfig(ctx context.Context, req *assistant_service.UpdateWgaConfigReq) (*emptypb.Empty, error) {
	if req.Identity == nil {
		return nil, errStatus(errs.Code_WgaConfigUpdateErr, toErrStatus("wga_config_update", "identity is required"))
	}

	toolListJSON, _ := json.Marshal(req.ToolList)
	assistantListJSON, _ := json.Marshal(req.AssistantList)
	mcpListJSON, _ := json.Marshal(req.McpList)
	workflowListJSON, _ := json.Marshal(req.WorkflowList)
	skillListJSON, _ := json.Marshal(req.SkillList)
	knowledgeListJSON, _ := json.Marshal(req.KnowledgeList)
	ontologyKnowledgeListJSON, _ := json.Marshal(req.OntologyKnowledgeList)

	config := &model.WgaConfig{
		UserID:                req.Identity.UserId,
		OrgID:                 req.Identity.OrgId,
		AssistantList:         string(assistantListJSON),
		ToolList:              string(toolListJSON),
		McpList:               string(mcpListJSON),
		WorkflowList:          string(workflowListJSON),
		SkillList:             string(skillListJSON),
		KnowledgeList:         string(knowledgeListJSON),
		OntologyKnowledgeList: string(ontologyKnowledgeListJSON),
	}

	status := s.cli.UpdateWgaConfig(ctx, config)
	if status != nil {
		log.Errorf("更新WGA配置失败，userId: %s, orgId: %s, error: %v", req.Identity.UserId, req.Identity.OrgId, status)
		return nil, errStatus(errs.Code_WgaConfigUpdateErr, status)
	}

	return &emptypb.Empty{}, nil
}

func toProtoWgaConversationConfig(m *model.WgaConversationConfig) *assistant_service.WgaConversationConfig {
	config := &assistant_service.WgaConversationConfig{
		ThreadId:  m.ThreadID,
		Title:     m.Title,
		UserId:    m.UserID,
		OrgId:     m.OrgID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.ModelConfig != "" {
		var modelConfig common.AppModelConfig
		if err := json.Unmarshal([]byte(m.ModelConfig), &modelConfig); err == nil {
			config.ModelConfig = &modelConfig
		}
	}

	return config
}

func toProtoWgaConfig(m *model.WgaConfig) *assistant_service.WgaConfig {
	config := &assistant_service.WgaConfig{
		UserId:    m.UserID,
		OrgId:     m.OrgID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	if m.ToolList != "" {
		var tools []assistant_service.WgaConfigTool
		if err := json.Unmarshal([]byte(m.ToolList), &tools); err == nil {
			for i := range tools {
				config.ToolList = append(config.ToolList, &assistant_service.WgaConfigTool{
					ToolId:   tools[i].ToolId,
					ToolType: tools[i].ToolType,
				})
			}
		}
	}

	if m.AssistantList != "" {
		var assistants []assistant_service.WgaConfigAssistant
		if err := json.Unmarshal([]byte(m.AssistantList), &assistants); err == nil {
			for i := range assistants {
				config.AssistantList = append(config.AssistantList, &assistant_service.WgaConfigAssistant{
					AssistantId:   assistants[i].AssistantId,
					AssistantType: assistants[i].AssistantType,
				})
			}
		}
	}

	if m.McpList != "" {
		var mcps []assistant_service.WgaConfigMcp
		if err := json.Unmarshal([]byte(m.McpList), &mcps); err == nil {
			for i := range mcps {
				config.McpList = append(config.McpList, &assistant_service.WgaConfigMcp{
					McpId:   mcps[i].McpId,
					McpType: mcps[i].McpType,
				})
			}
		}
	}

	if m.WorkflowList != "" {
		var workflows []assistant_service.WgaConfigWorkflow
		if err := json.Unmarshal([]byte(m.WorkflowList), &workflows); err == nil {
			for i := range workflows {
				config.WorkflowList = append(config.WorkflowList, &assistant_service.WgaConfigWorkflow{
					WorkflowId: workflows[i].WorkflowId,
				})
			}
		}
	}

	if m.SkillList != "" {
		var skills []assistant_service.WgaConfigSkill
		if err := json.Unmarshal([]byte(m.SkillList), &skills); err == nil {
			for i := range skills {
				config.SkillList = append(config.SkillList, &assistant_service.WgaConfigSkill{
					SkillId:   skills[i].SkillId,
					SkillType: skills[i].SkillType,
				})
			}
		}
	}

	if m.KnowledgeList != "" {
		var knowledges []assistant_service.WgaConfigKnowledge
		if err := json.Unmarshal([]byte(m.KnowledgeList), &knowledges); err == nil {
			for i := range knowledges {
				config.KnowledgeList = append(config.KnowledgeList, &assistant_service.WgaConfigKnowledge{
					KnowledgeId: knowledges[i].KnowledgeId,
				})
			}
		}
	}

	if m.OntologyKnowledgeList != "" {
		var ontologyKnowledges []assistant_service.WgaConfigOntologyKnowledge
		if err := json.Unmarshal([]byte(m.OntologyKnowledgeList), &ontologyKnowledges); err == nil {
			for i := range ontologyKnowledges {
				config.OntologyKnowledgeList = append(config.OntologyKnowledgeList, &assistant_service.WgaConfigOntologyKnowledge{
					OntologyKnowledgeId: ontologyKnowledges[i].OntologyKnowledgeId,
				})
			}
		}
	}

	return config
}
