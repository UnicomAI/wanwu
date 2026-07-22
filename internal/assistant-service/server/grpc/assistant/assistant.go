package assistant

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/UnicomAI/wanwu/internal/assistant-service/client"
	"github.com/UnicomAI/wanwu/internal/assistant-service/service"
	params_process "github.com/UnicomAI/wanwu/internal/assistant-service/service/params-process"
	"github.com/UnicomAI/wanwu/pkg/constant"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
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
	var appBriefs []*assistant_service.AssistantBrief
	for _, assistant := range assistants {
		appBriefs = append(appBriefs, transToAssistantBrief(assistant, ""))
	}

	return &assistant_service.AppBriefList{
		AssistantInfos: appBriefs,
	}, nil
}

// AssistantCreate 创建智能体
func (s *Service) AssistantCreate(ctx context.Context, req *assistant_service.AssistantCreateReq) (*assistant_service.AssistantCreateResp, error) {
	if req.Category == 0 {
		req.Category = model.SingleAgent
	}
	// 组装model参数
	assistant := &model.Assistant{
		UUID:       util.NewID(),
		AvatarPath: req.AssistantBrief.AvatarPath,
		Name:       req.AssistantBrief.Name,
		Desc:       req.AssistantBrief.Desc,
		Scope:      1,
		UserId:     req.Identity.UserId,
		OrgId:      req.Identity.OrgId,
		Category:   int(req.Category),
	}
	// 查找否存在相同名称智能体
	if err := s.cli.CheckSameAssistantName(ctx, req.Identity.UserId, req.Identity.OrgId, req.AssistantBrief.Name, ""); err != nil {
		return nil, errStatus(errs.Code_AssistantErr, err)
	}
	// 调用client方法创建智能体
	if status := s.cli.CreateAssistant(ctx, assistant); status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 创建智能体后自动绑定配置中的内置工具
	for _, tool := range config.Cfg().BuiltinTools {
		if status := s.cli.CreateAssistantTool(ctx, assistant.ID, tool.ToolId, tool.ToolType, tool.ActionName, req.Identity.UserId, req.Identity.OrgId); status != nil {
			log.Warnf("auto bind builtin tool %s for assistant %d failed: %v", tool.ToolId, assistant.ID, status)
		}
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
		existingAssistant.ModelConfig = string(modelConfigBytes)
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
			existingAssistant.RerankConfig = string(rerankConfigBytes)
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
		existingAssistant.KnowledgebaseConfig = string(knowledgeBaseConfigBytes)
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
		existingAssistant.SafetyConfig = string(safetyConfigBytes)
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
		existingAssistant.VisionConfig = string(visionConfigBytes)
	}

	// 处理memoryConfig，转换成json字符串之后再更新
	if req.MemoryConfig != nil {
		memoryConfigBytes, err := json.Marshal(req.MemoryConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_memoryConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.MemoryConfig = string(memoryConfigBytes)
	}

	// 处理recommendConfig，转换成json字符串之后再更新
	if req.RecommendConfig != nil {
		recommendConfigBytes, err := json.Marshal(req.RecommendConfig)
		if err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_recommendConfig_marshal",
				Args:    []string{err.Error()},
			})
		}
		existingAssistant.RecommendConfig = string(recommendConfigBytes)
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
	var appBriefs []*assistant_service.AssistantBrief
	for _, assistant := range assistants {
		appBriefs = append(appBriefs, &assistant_service.AssistantBrief{
			Info: &common.AppBrief{
				OrgId:      assistant.OrgId,
				UserId:     assistant.UserId,
				AppId:      util.Int2Str(assistant.ID),
				AppType:    "agent",
				AvatarPath: assistant.AvatarPath,
				Name:       assistant.Name,
				Desc:       assistant.Desc,
				CreatedAt:  assistant.CreatedAt,
				UpdatedAt:  assistant.UpdatedAt,
			},
			Category: int32(assistant.Category),
			Uuid:     assistant.UUID,
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

	// 获取关联的 Skill
	skills, _ := s.cli.GetAssistantSkillList(ctx, assistantId)
	// 转换 Skill
	var skillInfos []*assistant_service.AssistantSkillInfo
	for _, skill := range skills {
		skillInfos = append(skillInfos, &assistant_service.AssistantSkillInfo{
			Id:        util.Int2Str(skill.ID),
			SkillId:   skill.SkillId,
			SkillType: skill.SkillType,
			Enable:    skill.Enable,
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

	// 处理assistant.MemoryConfig，转换成AssistantMemoryConfig
	var memoryConfig *assistant_service.AssistantMemoryConfig
	if assistant.MemoryConfig != "" {
		memoryConfig = &assistant_service.AssistantMemoryConfig{}
		if err := json.Unmarshal([]byte(assistant.MemoryConfig), memoryConfig); err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_memoryConfig_unmarshal",
				Args:    []string{err.Error()},
			})
		}
	} else {
		memoryConfig = &assistant_service.AssistantMemoryConfig{
			MaxHistoryLength: config.DefaultMaxHistoryLength,
		}
	}

	// 处理assistant.RecommendConfig，转换成AssistantRecommendConfig
	var recommendConfig *assistant_service.AssistantRecommendConfig
	if assistant.RecommendConfig != "" {
		recommendConfig = &assistant_service.AssistantRecommendConfig{}
		if err := json.Unmarshal([]byte(assistant.RecommendConfig), recommendConfig); err != nil {
			return nil, errStatus(errs.Code_AssistantErr, &errs.Status{
				TextKey: "assistant_recommendConfig_unmarshal",
				Args:    []string{err.Error()},
			})
		}
	} else {
		recommendConfig = &assistant_service.AssistantRecommendConfig{
			MaxHistory: config.DefaultRecommendMaxHistory,
		}
	}

	// 构建多智能体信息
	userId, orgId := "", ""
	if req.Identity != nil {
		userId = req.Identity.UserId
		orgId = req.Identity.OrgId
	}
	multiAgentInfos, err := s.GetMultiAgentInfos(ctx, assistantId, userId, orgId, "", true)
	if err != nil {
		return nil, err
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
		MemoryConfig:        memoryConfig,
		RecommendConfig:     recommendConfig,
		Scope:               int32(assistant.Scope),
		WorkFlowInfos:       workFlowInfos,
		McpInfos:            mcpInfos,
		ToolInfos:           toolInfos,
		SkillInfos:          skillInfos,
		MultiAgentInfos:     multiAgentInfos,
		Category:            int32(assistant.Category),
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

func (s *Service) GetMultiAgentInfos(ctx context.Context, assistantId uint32, userId, orgId, version string, draft bool) ([]*assistant_service.AssistantMultiAgentInfos, error) {
	multiAgentInfos := make([]*assistant_service.AssistantMultiAgentInfos, 0)
	_, _, subAgents, err := s.cli.GetMultiAssistant(ctx, assistantId, userId, orgId, draft, version, false)
	if err != nil {
		return nil, errStatus(errs.Code_AssistantMultiAgentErr, &errs.Status{
			TextKey: "assistant_multi_agent_get",
			Args:    []string{err.Error()},
		})
	}
	if len(subAgents) > 0 {
		// 解析子智能体信息
		subAgentInfos, err := parseSubAgentInfos(subAgents)
		if err != nil {
			return nil, err
		}
		// 获取多智能体关系
		relations, errR := s.cli.FetchMultiAssistantRelationList(ctx, assistantId, version, draft)
		if errR != nil {
			return nil, errStatus(errs.Code_AssistantMultiAgentErr, errR)
		}
		relationMap := buildRelationMap(relations)
		for _, subAgent := range subAgentInfos {
			multiAgentInfos = append(multiAgentInfos, &assistant_service.AssistantMultiAgentInfos{
				AgentId:    util.Int2Str(subAgent.ID),
				Name:       subAgent.Name,
				Desc:       relationMap[subAgent.ID].Description,
				AvatarPath: subAgent.AvatarPath,
				Enable:     relationMap[subAgent.ID].Enable,
			})
		}
	}
	return multiAgentInfos, nil
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

	// 获取关联的多智能体配置
	subAgents, status := s.cli.FetchMultiAssistantRelationList(ctx, assistantId, "", true)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 获取关联的 skills
	skills, status := s.cli.GetAssistantSkillList(ctx, assistantId)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	// 复制智能体
	assistantID, status := s.cli.CopyAssistant(ctx, parentAssistant, workflows, mcps, tools, subAgents, skills)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}
	return &assistant_service.AssistantCreateResp{
		AssistantId: util.Int2Str(assistantID),
	}, nil
}

func (s *Service) GetAssistantDetailById(ctx context.Context, req *assistant_service.GetAssistantDetailByIdReq) (*assistant_service.AssistantDetailResp, error) {
	detail, snapshot, err := searchAssistantDetail(ctx, req.Draft, req.AssistantId, s.cli, req.Version)
	if err != nil {
		return nil, err
	}
	params, err := buildAgentParams(ctx, s.cli, detail, snapshot, req.ConversationId, req.Identity.UserId, req.Identity.OrgId)
	if err != nil {
		log.Errorf("Assistant服务获取智能体信息失败，assistantId: %d, error: %v", req.AssistantId, err)
		return nil, errCode(errs.Code_AssistantConversationErr)
	}
	return &assistant_service.AssistantDetailResp{
		AgentDetail: params,
	}, nil
}

// AdminAssistantPageList 管理员中心智能体列表
func (s *Service) AdminAssistantPageList(ctx context.Context, req *assistant_service.AdminAssistantPageListReq) (*assistant_service.AdminAssistantPageListResp, error) {
	pageNum := int(req.PageNum)
	pageSize := int(req.PageSize)

	var items []*model.AdminAssistantItem
	var total int64
	var status *errs.Status

	if req.PublishStatus == nil {
		// 查询所有智能体
		items, total, status = s.cli.AdminGetAssistantListAll(ctx, req.UserId, req.OrgId, req.Name, req.Category)
	} else {
		// 分页查询
		items, total, status = s.cli.AdminGetAssistantListPage(ctx, req.UserId, req.OrgId, req.Name, req.Category, pageNum, pageSize)
	}
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}

	var appBriefs []*assistant_service.AssistantBrief
	for _, item := range items {
		appBriefs = append(appBriefs, transToAssistantBrief(&item.Assistant, item.PublishType))
	}

	return &assistant_service.AdminAssistantPageListResp{
		Total: total,
		List:  appBriefs,
	}, nil
}

// SyncAssistantPublish 同步智能体发布状态到 assistant 数据库的关联表
// publishType 非空=发布记录，空=删除发布记录(取消发布)
func (s *Service) SyncAssistantPublish(ctx context.Context, req *assistant_service.SyncAssistantPublishReq) (*emptypb.Empty, error) {
	assistantId := util.MustU32(req.AssistantId)
	var status *errs.Status
	if req.PublishType == "" {
		status = s.cli.DeleteAssistantPublish(ctx, assistantId)
	} else {
		status = s.cli.UpsertAssistantPublish(ctx, assistantId, req.PublishType)
	}
	if status != nil {
		return nil, errStatus(errs.Code_AssistantErr, status)
	}
	return &emptypb.Empty{}, nil
}

// ---internal methods---

func searchAssistantDetail(ctx context.Context, draft bool, assistantId uint32, cli client.IClient, version string) (*model.Assistant, *model.AssistantSnapshot, error) {
	assistant := &model.Assistant{}
	var assistantSnapshot *model.AssistantSnapshot
	var status *errs.Status
	if draft {
		assistant, status = cli.GetAssistant(ctx, assistantId, "", "")
		if status != nil {
			log.Errorf("Assistant服务获取智能体信息失败，assistantId: %d, error: %v", assistantId, status)
			return nil, nil, errStatus(errs.Code_AssistantConversationErr, status)
		}
	} else {
		assistantSnapshot, status = cli.GetAssistantSnapshot(ctx, assistantId, version)
		if status != nil {
			log.Errorf("Assistant服务获取智能体快照失败，assistantId: %d, error: %v", assistantId, status)
			return nil, nil, errStatus(errs.Code_AssistantConversationErr, status)
		}

		if err := jsonToStruct(assistantSnapshot.AssistantInfo, &assistant); err != nil {
			return nil, nil, errStatus(errs.Code_AssistantErr, toErrStatus("assistant_snapshot", err.Error()))
		}
	}
	return assistant, assistantSnapshot, nil
}

func buildAgentParams(ctx context.Context, cli client.IClient, agent *model.Assistant, snapshot *model.AssistantSnapshot, conversationId, userId, orgId string) (*assistant_service.AgentDetail, error) {
	clientInfo := &params_process.ClientInfo{
		Cli:       cli,
		Knowledge: Knowledge,
		MCP:       MCP,
	}
	//传入了 ConversationId就会尝试构造历史数据
	userQueryParams := &params_process.UserQueryParams{
		ConversationId: conversationId,
		QueryUserId:    userId,
		QueryOrgId:     orgId,
		Ctx:            ctx,
	}
	return service.NewAgentChatParamsBuilder(&params_process.AgentInfo{
		Assistant:         agent,
		AssistantSnapshot: snapshot,
		Draft:             snapshot == nil,
	}, userQueryParams, clientInfo).
		AgentBaseParams().
		ModelParams().
		KnowledgeParams().
		ToolParams().
		SkillParams().
		Build()
}

func parseSubAgentInfos(subAgents []*model.AssistantSnapshot) ([]*model.Assistant, error) {
	var subAgentInfos []*model.Assistant
	for _, subAgent := range subAgents {
		// 解析subAgent
		var subAgentInfo *model.Assistant
		if err := jsonToStruct(subAgent.AssistantInfo, &subAgentInfo); err != nil {
			return nil, errStatus(errs.Code_AssistantErr, toErrStatus("assistant_snapshot", err.Error()))
		}
		if subAgentInfo == nil {
			return nil, errStatus(errs.Code_AssistantErr, toErrStatus("assistant_snapshot", "assistant info is nil"))
		}
		subAgentInfos = append(subAgentInfos, subAgentInfo)
	}
	return subAgentInfos, nil
}

func buildRelationMap(relations []*model.MultiAgentRelation) map[uint32]*model.MultiAgentRelation {
	relationMap := make(map[uint32]*model.MultiAgentRelation, len(relations))
	for _, relation := range relations {
		relationMap[relation.AgentId] = relation
	}
	return relationMap
}

func transToAssistantBrief(assistant *model.Assistant, publishType string) *assistant_service.AssistantBrief {
	assistantBrief := assistant_service.AssistantBrief{
		Info: &common.AppBrief{
			OrgId:      assistant.OrgId,
			UserId:     assistant.UserId,
			AppId:      util.Int2Str(assistant.ID),
			AppType:    constant.AppTypeAgent,
			AvatarPath: assistant.AvatarPath,
			Name:       assistant.Name,
			Desc:       assistant.Desc,
			CreatedAt:  assistant.CreatedAt,
			UpdatedAt:  assistant.UpdatedAt,
		},
		Category: int32(assistant.Category),
		Uuid:     assistant.UUID,
	}
	if publishType != "" {
		assistantBrief.PublishType = publishType
	}
	return &assistantBrief
}
