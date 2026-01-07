package assistant

import (
	"context"
	"encoding/json"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetAssistantByIds 根据智能体id集合获取智能体列表
func (s *Service) GetAssistantByIds(ctx context.Context, req *assistant_service.GetAssistantByIdsReq) (*assistant_service.AppBriefList, error) {
	// 转换字符串ID为uint32
	var assistantIDs []uint32
	for _, idStr := range req.AssistantIdList {
		assistantIDs = append(assistantIDs, util.MustU32(idStr))
	}

	// 调用client方法获取智能体列表
	assistants, status := s.cli.GetAssistantsByIDs(ctx, assistantIDs)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 转换为响应格式
	var appBriefs []*common.AppBrief
	for _, assistant := range assistants {
		appBriefs = append(appBriefs, &common.AppBrief{
			OrgId:      assistant.OrgId,
			UserId:     assistant.UserId,
			AppId:      util.Int2Str(assistant.ID),
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

// AssistantCreate 创建智能体
func (s *Service) AssistantCreate(ctx context.Context, req *assistant_service.AssistantCreateReq) (*assistant_service.AssistantCreateResp, error) {
	// 组装model参数
	assistant := &model.Assistant{
		UUID:       util.NewID(),
		AvatarPath: req.AssistantBrief.AvatarPath,
		Name:       req.AssistantBrief.Name,
		Desc:       req.AssistantBrief.Desc,
		Scope:      1,
		UserId:     req.Identity.UserId,
		OrgId:      req.Identity.OrgId,
	}
	// 查找否存在相同名称智能体
	if err := s.cli.CheckSameAssistantName(ctx, req.Identity.UserId, req.Identity.OrgId, req.AssistantBrief.Name, ""); err != nil {
		return nil, errStatus(errs.Code_AssistantErr, err)
	}
	// 调用client方法创建智能体
	if status := s.cli.CreateAssistant(ctx, assistant); status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	return &assistant_service.AssistantCreateResp{
		AssistantId: util.Int2Str(assistant.ID),
	}, nil
}

// AssistantUpdate 修改智能体
func (s *Service) AssistantUpdate(ctx context.Context, req *assistant_service.AssistantUpdateReq) (*emptypb.Empty, error) {
	// 转换ID
	assistantID := util.MustU32(req.AssistantId)

	// 获取现有智能体信息
	existingAssistant, status := s.cli.GetAssistant(ctx, assistantID, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 查找否存在相同名称智能体
	if err := s.cli.CheckSameAssistantName(ctx, req.Identity.UserId, req.Identity.OrgId, req.AssistantBrief.Name, req.AssistantId); err != nil {
		return nil, errStatus(errs.Code_AssistantErr, err)
	}

	existingAssistant.AvatarPath = req.AssistantBrief.AvatarPath
	existingAssistant.Name = req.AssistantBrief.Name
	existingAssistant.Desc = req.AssistantBrief.Desc

	// 调用client方法更新智能体
	if status := s.cli.UpdateAssistant(ctx, existingAssistant); status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	return &emptypb.Empty{}, nil
}

// AssistantDelete 删除智能体
func (s *Service) AssistantDelete(ctx context.Context, req *assistant_service.AssistantDeleteReq) (*emptypb.Empty, error) {
	// 转换ID
	assistantID := util.MustU32(req.AssistantId)

	// 调用client方法删除智能体
	if status := s.cli.DeleteAssistant(ctx, assistantID); status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	return &emptypb.Empty{}, nil
}

// AssistantConfigUpdate 修改智能体配置
func (s *Service) AssistantConfigUpdate(ctx context.Context, req *assistant_service.AssistantConfigUpdateReq) (*emptypb.Empty, error) {
	// 转换ID
	assistantID := util.MustU32(req.AssistantId)

	// 先获取现有智能体信息
	existingAssistant, status := s.cli.GetAssistant(ctx, assistantID, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 更新配置字段
	existingAssistant.Instructions = req.Instructions
	existingAssistant.Prologue = req.Prologue
	existingAssistant.RecommendQuestion = strings.Join(req.RecommendQuestion, "@#@")

	// 处理modelConfig，转换成json字符串之后再更新
	if req.ModelConfig != nil {
		modelConfigBytes, err := json.Marshal(req.ModelConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_modelConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.ModelConfig = db.LongText(modelConfigBytes)
	}

	// 处理rerankConfig，转换成json字符串之后再更新
	var knowledgeBaseIds []string
	if req.KnowledgeBaseConfig != nil {
		knowledgeBaseIds = req.KnowledgeBaseConfig.GetKnowledgeBaseIds()
	}

	if req.KnowledgeBaseConfig == nil || len(knowledgeBaseIds) == 0 {
		existingAssistant.RerankConfig = ""
	} else {
		if req.RerankConfig != nil {
			rerankConfigBytes, err := json.Marshal(req.RerankConfig)
			if err != nil {
				return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
					TextKey: "assistant_rerankConfig_marshal",
					Args:    []string{err.Error()},
				})
			}
			existingAssistant.RerankConfig = db.LongText(rerankConfigBytes)
		}
	}

	// 处理knowledgeBaseConfig，转换成json字符串之后再更新
	if req.KnowledgeBaseConfig != nil {
		knowledgeBaseConfigBytes, err := json.Marshal(req.KnowledgeBaseConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_knowledgeBaseConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.KnowledgebaseConfig = db.LongText(knowledgeBaseConfigBytes)
		log.Debugf("knowConfig = %s", existingAssistant.KnowledgebaseConfig)
	}

	// 处理safetyConfig，转换成json字符串之后再更新
	if req.SafetyConfig != nil {
		safetyConfigBytes, err := json.Marshal(req.SafetyConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_safetyConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.SafetyConfig = db.LongText(safetyConfigBytes)
	}

	// 处理visionConfig，转换成json字符串之后再更新
	if req.VisionConfig != nil {
		visionConfigBytes, err := json.Marshal(req.VisionConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_visionConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.VisionConfig = db.LongText(visionConfigBytes)
	}

	// 调用client方法更新智能体
	if status := s.cli.UpdateAssistant(ctx, existingAssistant); status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	return &emptypb.Empty{}, nil
}

// GetAssistantListMyAll 智能体列表
func (s *Service) GetAssistantListMyAll(ctx context.Context, req *assistant_service.GetAssistantListMyAllReq) (*assistant_service.AppBriefList, error) {
	// 调用client方法获取智能体列表
	assistants, _, status := s.cli.GetAssistantList(ctx, req.Identity.UserId, req.Identity.OrgId, req.Name)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 转换为响应格式
	var appBriefs []*common.AppBrief
	for _, assistant := range assistants {
		appBriefs = append(appBriefs, &common.AppBrief{
			OrgId:      assistant.OrgId,
			UserId:     assistant.UserId,
			AppId:      util.Int2Str(assistant.ID),
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

// GetAssistantInfo 查看智能体详情
func (s *Service) GetAssistantInfo(ctx context.Context, req *assistant_service.GetAssistantInfoReq) (*assistant_service.AssistantInfo, error) {
	// 转换ID
	assistantId, err := util.U32(req.AssistantId)
	if err != nil {
		return nil, err
	}

	// 判空处理，根据Identity是否为空使用不同参数
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

	// 获取关联的WorkFlows
	workflows, _ := s.cli.GetAssistantWorkflowsByAssistantID(ctx, assistantId)

	// 转换WorkFlows
	var workFlowInfos []*assistant_service.AssistantWorkFlowInfos
	for _, workflow := range workflows {
		workFlowInfos = append(workFlowInfos, &assistant_service.AssistantWorkFlowInfos{
			Id:         util.Int2Str(workflow.ID),
			WorkFlowId: workflow.WorkflowId,
			Enable:     workflow.Enable,
		})
	}

	// 获取关联的 MCP
	mcps, _ := s.cli.GetAssistantMCPList(ctx, assistantId)
	// 转换MCP
	var mcpInfos []*assistant_service.AssistantMCPInfos
	for _, mcp := range mcps {
		mcpInfos = append(mcpInfos, &assistant_service.AssistantMCPInfos{
			Id:         util.Int2Str(mcp.ID),
			McpId:      mcp.MCPId,
			McpType:    mcp.MCPType,
			ActionName: mcp.ActionName,
			Enable:     mcp.Enable,
		})
	}

	// 获取关联的 Tool
	tools, _ := s.cli.GetAssistantToolList(ctx, assistantId)
	// 转换 Tool
	var toolInfos []*assistant_service.AssistantToolInfos
	for _, tool := range tools {
		toolInfos = append(toolInfos, &assistant_service.AssistantToolInfos{
			Id:         util.Int2Str(tool.ID),
			ToolId:     tool.ToolId,
			ToolType:   tool.ToolType,
			ActionName: tool.ActionName,
			Enable:     tool.Enable,
			ToolConfig: tool.ToolConfig,
		})
	}

	// 处理assistant.ModelConfig，转换成common.AppModelConfig
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

	// 处理assistant.RerankConfig，转换成common.AppModelConfig
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

	// 处理assistant.KnowledgebaseConfig，转换成AssistantKnowledgeBaseConfig
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

	// 处理assistant.SafetyConfig，转换成AssistantSafetyConfig
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

	// 处理assistant.VisionConfig，转换成AssistantVisionConfig
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
		AssistantId: util.Int2Str(assistant.ID),
		Identity: &assistant_service.Identity{
			UserId: assistant.UserId,
			OrgId:  assistant.OrgId,
		},
		Uuid: assistant.UUID,
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

func (s *Service) GetAssistantIdByUuid(ctx context.Context, req *assistant_service.GetAssistantIdByUuidReq) (*assistant_service.GetAssistantIdByUuidResp, error) {
	assistant, status := s.cli.GetAssistantByUuid(ctx, req.Uuid)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}
	return &assistant_service.GetAssistantIdByUuidResp{
		AssistantId: util.Int2Str(assistant.ID),
	}, nil
}

func (s *Service) AssistantCopy(ctx context.Context, req *assistant_service.AssistantCopyReq) (*assistant_service.AssistantCreateResp, error) {
	assistantId, err := util.U32(req.AssistantId)
	if err != nil {
		return nil, err
	}

	// 获取父智能体信息
	parentAssistant, status := s.cli.GetAssistant(ctx, assistantId, "", "")
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 获取关联的 workflow
	workflows, status := s.cli.GetAssistantWorkflowsByAssistantID(ctx, assistantId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 获取关联的 mcp
	mcps, status := s.cli.GetAssistantMCPList(ctx, assistantId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 获取关联的 tool
	tools, status := s.cli.GetAssistantToolList(ctx, assistantId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 复制智能体
	assistantID, status := s.cli.CopyAssistant(ctx, parentAssistant, workflows, mcps, tools)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}
	return &assistant_service.AssistantCreateResp{
		AssistantId: util.Int2Str(assistantID),
	}, nil
}
