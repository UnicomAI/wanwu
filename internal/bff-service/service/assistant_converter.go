package service

import (
	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	knowledgeBase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	"github.com/gin-gonic/gin"
)

type AssistantConverterError struct {
	Error error
}

// AssistantConverter 把 AssistantInfo 的某个子配置转换并写入 response.Assistant 的对应字段。
type AssistantConverter interface {
	NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool
	Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error
}

// assistantConverters 是 transAssistantResp2Model 顺序执行的转换链。
// 顺序敏感：KnowledgeBaseConverter 必须先于 RerankConverter 执行，
// 因为 RerankConverter 依赖已写入的 resp.KnowledgeBaseConfig。
var assistantConverters = []AssistantConverter{
	&AppConverter{},
	&KnowledgeBaseConverter{},
	&ModelConverter{},
	&RerankConverter{},
	&WorkFlowConverter{},
	&MCPConverter{},
	&ToolsConverter{},
	&SkillConverter{},
	&SafetyConverter{},
	&RecommendConverter{},
	&MultiAgentConverter{},
}

// ConcurrentConverter 并发执行各子配置转换，写入 assistantModel 对应字段。
// 对实现了 ConverterPreflight 的 Converter，若预判无需执行则跳过，不创建协程。
func ConcurrentConverter(ctx *gin.Context, resp *assistant_service.AssistantInfo, assistantModel *response.Assistant) error {
	var funcList []func()
	converterErr := AssistantConverterError{}
	for _, c := range assistantConverters {
		// 预判：若 Converter 声明无需转换，跳过本次执行
		if !c.NeedConvert(resp, assistantModel) {
			continue
		}
		funcList = append(funcList, func() {
			err := c.Convert(ctx, resp, assistantModel)
			if err != nil {
				converterErr.Error = err
			}
		})
	}
	safe_go_util.SageGoWaitGroup(funcList...)
	return converterErr.Error
}

// convertAppModelConfig 取模型信息并完成 proto->model 转换，供 Model/Rerank/Recommend 复用。
// 入参为空或 ModelId 为空时返回零值，与原 assistantModelConvert 行为一致。
func convertAppModelConfig(ctx *gin.Context, modelConfigInfo *common.AppModelConfig) (request.AppModelConfig, error) {
	var modelConfig request.AppModelConfig
	if modelConfigInfo != nil && modelConfigInfo.ModelId != "" {
		modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelConfigInfo.ModelId})
		if err != nil {
			log.Errorf("获取模型信息失败，模型ID: %s, 错误: %v", modelConfigInfo.ModelId, err)
		}
		if modelInfo != nil {
			modelConfig, err = appModelConfigProto2Model(modelConfigInfo, modelInfo.DisplayName)
			if err != nil {
				log.Errorf("模型配置Proto转换到模型失败，模型ID: %s, 错误: %v", modelConfigInfo.ModelId, err)
				return modelConfig, err
			}
		}
	}
	return modelConfig, nil
}

type AppConverter struct {
}

// NeedConvert 与 convertAppModelConfig 的取模型条件一致：配置为空或 ModelId 为空时无需执行。
func (*AppConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	return true
}

func (*AppConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	// 获取app发布信息，可能没有发布过，不返回错误
	appInfo, _ := app.GetAppInfo(ctx, &app_service.GetAppInfoReq{AppId: resp.AssistantId, AppType: constant.AppTypeAgent})
	if appInfo != nil {
		resp.PublishType = appInfo.PublishType
	}
	return nil
}

type ModelConverter struct {
}

func (*ModelConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	modelConfig, err := convertAppModelConfig(ctx, assistant.ModelConfig)
	if err != nil {
		return err
	}
	resp.ModelConfig = modelConfig
	return nil
}

// NeedConvert 与 convertAppModelConfig 的取模型条件一致：配置为空或 ModelId 为空时无需执行。
func (*ModelConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	return assistant.ModelConfig != nil && assistant.ModelConfig.ModelId != ""
}

type RerankConverter struct {
}

func (*RerankConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	rerankConfig, err := convertAppModelConfig(ctx, assistant.RerankConfig)
	if err != nil {
		return err
	}
	resp.RerankConfig = rerankConfig
	return nil
}

// NeedConvert 复用 convertAppModelConfig 的取模型条件；KB 为空时 Convert 内部亦会短路。
func (*RerankConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	// 仅当存在知识库配置时才转换 Rerank
	if assistant.KnowledgeBaseConfig == nil || len(assistant.KnowledgeBaseConfig.KnowledgeBaseIds) == 0 {
		return false
	}
	return assistant.RerankConfig != nil && assistant.RerankConfig.ModelId != ""
}

type KnowledgeBaseConverter struct {
}

// NeedConvert 复用 convertAppModelConfig 的取模型条件；KB 为空时 Convert 内部亦会短路。
func (*KnowledgeBaseConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	kbConfig := assistant.KnowledgeBaseConfig
	if kbConfig == nil || len(kbConfig.KnowledgeBaseIds) == 0 {
		log.Debugf("知识库配置为空")
		resp.KnowledgeBaseConfig = request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
		}
		return false
	}
	return true
}
func (*KnowledgeBaseConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	knowledgeBaseConfig, err := transKnowledgeBases2Model(ctx, assistant.KnowledgeBaseConfig)
	if err != nil {
		return err
	}
	resp.KnowledgeBaseConfig = knowledgeBaseConfig
	return nil
}

type WorkFlowConverter struct {
}

func (*WorkFlowConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	workFlowInfos, err := assistantWorkFlowConvert(ctx, assistant.WorkFlowInfos)
	if err != nil {
		return err
	}
	resp.WorkFlowInfos = workFlowInfos
	return nil
}

// NeedConvert 与 assistantWorkFlowConvert 的空入参短路一致：无工作流时不发 RPC。
func (*WorkFlowConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	return len(assistant.WorkFlowInfos) > 0
}

type MCPConverter struct {
}

func (*MCPConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	mcpInfos, err := assistantMCPConvert(ctx, assistant.McpInfos)
	if err != nil {
		return err
	}
	resp.MCPInfos = mcpInfos
	return nil
}

// NeedConvert 与 assistantMCPConvert 的空入参短路一致：无 MCP 时不发 RPC。
func (*MCPConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	return len(assistant.McpInfos) > 0
}

type ToolsConverter struct {
}

func (*ToolsConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	toolInfos, err := assistantToolsConvert(ctx, assistant.ToolInfos)
	if err != nil {
		return err
	}
	resp.ToolInfos = toolInfos
	return nil
}

// NeedConvert 与 assistantToolsConvert 的空入参短路一致：无工具时不发 RPC。
func (*ToolsConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	return len(assistant.ToolInfos) > 0
}

type SkillConverter struct {
}

func (*SkillConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	skillInfos, err := assistantSkillConvert(ctx, assistant.SkillInfos)
	if err != nil {
		return err
	}
	resp.SkillInfos = skillInfos
	return nil
}

// NeedConvert 与 assistantSkillConvert 的空入参短路一致：无技能时不发 RPC。
func (*SkillConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	return len(assistant.SkillInfos) > 0
}

type SafetyConverter struct {
}

func (*SafetyConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	safetyConfig, err := assistantSafetyConvert(ctx, assistant.SafetyConfig)
	if err != nil {
		return err
	}
	resp.SafetyConfig = safetyConfig
	return nil
}

// NeedConvert 与 assistantSafetyConvert 的空入参行为一致：无敏感词表配置时不发 RPC，结果与零值相同。
func (*SafetyConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	return assistant.SafetyConfig != nil && len(assistant.SafetyConfig.SensitiveTable) > 0
}

type RecommendConverter struct {
}

func (*RecommendConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	if assistant.RecommendConfig == nil {
		return nil
	}
	modelConfig, err := convertAppModelConfig(ctx, assistant.RecommendConfig.ModelConfig)
	if err != nil {
		return err
	}
	recommendConfig := response.RecommendConfig{
		ModelConfig:     modelConfig,
		MaxHistory:      assistant.RecommendConfig.MaxHistory,
		RecommendEnable: assistant.RecommendConfig.RecommendEnable,
		Prompt:          assistant.RecommendConfig.SystemPrompt,
		PromptEnable:    assistant.RecommendConfig.PromptEnable,
	}
	// prompt 为空时回填默认提示词
	if recommendConfig.Prompt == "" {
		recommendConfig.Prompt = systemPrompt
	}
	resp.RecommendConfig = recommendConfig
	return nil
}

// NeedConvert 与 Convert 入口的 nil 短路一致：无追问配置时无需执行。
func (*RecommendConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	recommendConfig := assistant.RecommendConfig
	return recommendConfig != nil && recommendConfig.ModelConfig != nil && len(recommendConfig.ModelConfig.ModelId) > 0
}

type MultiAgentConverter struct {
}

func (*MultiAgentConverter) Convert(ctx *gin.Context, assistant *assistant_service.AssistantInfo, resp *response.Assistant) error {
	assistantMultiAgents := make([]*response.AssistantAgentInfo, 0)
	for _, agent := range assistant.MultiAgentInfos {
		multiAgent := &response.AssistantAgentInfo{
			AgentId: agent.AgentId,
			Name:    agent.Name,
			Desc:    agent.Desc,
			Avatar:  cacheAppAvatar(ctx, agent.AvatarPath, constant.AppTypeAgent),
			Enable:  agent.Enable,
		}
		assistantMultiAgents = append(assistantMultiAgents, multiAgent)
	}
	resp.MultiAgentInfos = assistantMultiAgents
	return nil
}

// NeedConvert 有子智能体才需要执行
func (*MultiAgentConverter) NeedConvert(assistant *assistant_service.AssistantInfo, resp *response.Assistant) bool {
	return len(assistant.MultiAgentInfos) > 0
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

	// 获取知识库详情列表
	kbInfoList, err := knowledgeBase.SelectKnowledgeDetailByIdList(ctx, &knowledgeBase_service.KnowledgeDetailSelectListReq{
		KnowledgeIds: kbConfig.KnowledgeBaseIds,
	})

	if err != nil || kbInfoList == nil || len(kbInfoList.List) == 0 {
		return request.AppKnowledgebaseConfig{
			Knowledgebases: make([]request.AppKnowledgeBase, 0),
		}, err
	}

	knowledgeBases := buildKnowledgeBases(ctx, kbInfoList, kbConfig.KnowledgeBaseIds, kbConfig.AppKnowledgeBaseList)

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

func buildKnowledgeBases(ctx *gin.Context, kbInfoList *knowledgeBase_service.KnowledgeDetailSelectListResp, kbIdList []string, kbConfigList []*assistant_service.AppKnowledgeBase) []request.AppKnowledgeBase {
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
			share := info.ShareCount > 1
			var orgName string
			if share {
				orgInfo, err := iam.GetOrgInfo(ctx, &iam_service.GetOrgInfoReq{OrgId: info.CreateOrgId})
				if err != nil {
					log.Errorf("get org info error: %v", err)
				} else {
					orgName = buildShareOrgName(share, orgInfo.Name)
				}
			}
			params := buildAssistantMetaDataFilterParams(kbConfig)
			knowledgeBases = append(knowledgeBases, request.AppKnowledgeBase{
				ID:                   kbConfig.KnowledgeBaseId,
				Name:                 info.Name,
				GraphSwitch:          info.GraphSwitch,
				External:             info.External,
				Category:             info.Category,
				Share:                share,
				OrgName:              orgName,
				MetaDataFilterParams: params,
				Avatar:               cacheKnowledgeAvatar(ctx, info.AvatarPath, info.Category),
				Description:          info.Description,
			})
		}
	} else {
		for _, kbId := range kbIdList {
			info := knowledgeMap[kbId]
			if info == nil {
				continue
			}
			knowledgeBases = append(knowledgeBases, request.AppKnowledgeBase{
				ID:          kbId,
				Name:        info.Name,
				Description: info.Description,
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
