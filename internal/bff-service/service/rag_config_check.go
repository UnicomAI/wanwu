package service

import (
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// 知识问答「可发布 / 可问答」的配置前置校验。
type ragConfigCheckInput struct {
	modelID         string
	kbCount         int
	kbRerankID      string
	kbMatchType     string
	kbPriorityMatch int32
	qaCount         int
	qaRerankID      string
}

// CheckRagPublishReady 按 ragId 拉草稿配置做发布前校验
func CheckRagPublishReady(ctx *gin.Context, userId, orgId, ragId string) error {
	ragInfo, err := GetRag(ctx, userId, orgId, request.RagReq{RagID: ragId}, false)
	if err != nil {
		return err
	}
	return CheckRagConfigReady(ragInfo)
}

// CheckRagConfigReady 校验响应模型形态的知识问答配置
func CheckRagConfigReady(ragInfo *response.RagInfo) error {
	if ragInfo == nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_model_not_configured")
	}
	return checkRagConfig(ragConfigCheckInput{
		modelID:         ragInfo.ModelConfig.ModelId,
		kbCount:         len(ragInfo.KnowledgeBaseConfig.Knowledgebases),
		kbRerankID:      ragInfo.RerankConfig.ModelId,
		kbMatchType:     ragInfo.KnowledgeBaseConfig.Config.MatchType,
		kbPriorityMatch: ragInfo.KnowledgeBaseConfig.Config.PriorityMatch,
		qaCount:         len(ragInfo.QAKnowledgeBaseConfig.Knowledgebases),
		qaRerankID:      ragInfo.QARerankConfig.ModelId,
	})
}

// CheckRagChatFileReady 带文件提问的前置校验，口径与页面「视觉」配置块一致：
// 视觉开关已开、对话模型支持图文问答、至少绑定一个多模态知识库、知识库 rerank 为多模态 rerank
func CheckRagChatFileReady(ctx *gin.Context, userId, orgId, ragID string, needLatestPublished bool) error {
	var identity *rag_service.Identity
	if !needLatestPublished {
		identity = &rag_service.Identity{UserId: userId, OrgId: orgId}
	}
	ragInfo, err := rag.GetRagDetail(ctx.Request.Context(), &rag_service.RagDetailReq{
		RagId:    ragID,
		Publish:  util.IfElse(needLatestPublished, int32(1), int32(0)),
		Identity: identity,
	})
	if err != nil {
		return err
	}
	if ragInfo.GetVisionConfig().GetPicNum() <= 0 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_vision_not_enabled")
	}
	if err := checkRagChatModelVision(ctx, ragInfo.GetModelConfig().GetModelId()); err != nil {
		return err
	}
	if err := checkRagChatRerankMultimodal(ctx, ragInfo.GetRerankConfig().GetModelId()); err != nil {
		return err
	}
	return checkRagChatKnowledgeMultimodal(ctx, userId, orgId, ragInfo.GetKnowledgeBaseConfig().GetPerKnowledgeConfigs())
}

// checkRagChatModelVision 对话模型须带「图文问答」标签
func checkRagChatModelVision(ctx *gin.Context, modelID string) error {
	if modelID == "" {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_model_not_configured")
	}
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelID})
	if err != nil {
		return err
	}
	visionSupport, err := checkModelExperienceModelVisionSupport(modelInfo)
	if err != nil {
		return err
	}
	if !visionSupport {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_model_not_vision")
	}
	return nil
}

// checkRagChatRerankMultimodal 知识库 rerank 模型须为 multimodal-rerank
func checkRagChatRerankMultimodal(ctx *gin.Context, rerankModelID string) error {
	if rerankModelID == "" {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_rerank_not_multimodal")
	}
	rerankInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: rerankModelID})
	if err != nil {
		return err
	}
	if rerankInfo.GetModelType() != mp.ModelTypeMultiRerank {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_rerank_not_multimodal")
	}
	return nil
}

// checkRagChatKnowledgeMultimodal 绑定的知识库至少有一个是多模态，与页面 getCategory 同口径
func checkRagChatKnowledgeMultimodal(ctx *gin.Context, userId, orgId string, kbConfigs []*rag_service.RagPerKnowledgeConfig) error {
	for _, kb := range kbConfigs {
		multimodal, err := isMultimodalKnowledge(ctx, userId, orgId, kb.GetKnowledgeId())
		if err != nil {
			return err
		}
		if multimodal {
			return nil
		}
	}
	return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_knowledge_not_multimodal")
}

// checkRagProtoConfigReady 校验 proto 形态的知识问答配置，供问答链路复用已拉到的详情，不额外发一次查询
func checkRagProtoConfigReady(ragInfo *rag_service.RagInfo) error {
	if ragInfo == nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_model_not_configured")
	}
	return checkRagConfig(ragConfigCheckInput{
		modelID:         ragInfo.GetModelConfig().GetModelId(),
		kbCount:         len(ragInfo.GetKnowledgeBaseConfig().GetPerKnowledgeConfigs()),
		kbRerankID:      ragInfo.GetRerankConfig().GetModelId(),
		kbMatchType:     ragInfo.GetKnowledgeBaseConfig().GetGlobalConfig().GetMatchType(),
		kbPriorityMatch: ragInfo.GetKnowledgeBaseConfig().GetGlobalConfig().GetPriorityMatch(),
		qaCount:         len(ragInfo.GetQAknowledgeBaseConfig().GetPerKnowledgeConfigs()),
		qaRerankID:      ragInfo.GetQArerankConfig().GetModelId(),
	})
}

// checkRagConfig 对话模型必配 → 知识库/问答库至少绑一个 →
// 绑了知识库要有 rerank（混合检索+权重匹配可免）→ 绑了问答库要有 rerank
func checkRagConfig(in ragConfigCheckInput) error {
	if in.modelID == "" {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_model_not_configured")
	}
	if in.kbCount == 0 && in.qaCount == 0 {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_knowledge_not_configured")
	}
	// 混合检索且开了权重匹配时由权重决定排序，可以不配 rerank
	if in.kbCount > 0 && in.kbRerankID == "" && !isRagMixPriorityMatch(in.kbMatchType, in.kbPriorityMatch) {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_rerank_not_configured")
	}
	if in.qaCount > 0 && in.qaRerankID == "" {
		return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_rag_qa_rerank_not_configured")
	}
	return nil
}

// isRagMixPriorityMatch 混合检索 + 权重匹配
func isRagMixPriorityMatch(matchType string, priorityMatch int32) bool {
	return matchType == "mix" && priorityMatch == 1
}
