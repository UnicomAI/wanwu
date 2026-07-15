package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/pkg/constant"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_jina "github.com/UnicomAI/wanwu/pkg/model-provider/mp-jina"
	utils "github.com/UnicomAI/wanwu/pkg/util"
	"github.com/google/uuid"

	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	iam_service "github.com/UnicomAI/wanwu/api/proto/iam-service"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

var knowHttp = http_client.CreateDefault()

type RagResponseInfo struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	MsgID   string                 `json:"msg_id"`
	Data    *RagData               `json:"data"`
	History []*request.HistoryItem `json:"history"`
	Finish  int                    `json:"finish"`
}

type RagData struct {
	Score      []float64          `json:"score"`
	Output     string             `json:"output"`
	SearchList []*ChunkSearchList `json:"searchList"`
}

// SelectKnowledgeList 查询知识库列表，主要根据userId 查询用户所有知识库
func SelectKnowledgeList(ctx *gin.Context, userId, orgId string, req *request.KnowledgeSelectReq) (*response.KnowledgeListResp, error) {
	resp, err := knowledgeBase.SelectKnowledgeList(ctx.Request.Context(), &knowledgebase_service.KnowledgeSelectReq{
		UserId:    userId,
		OrgId:     orgId,
		Name:      strings.TrimSpace(req.Name),
		TagIdList: req.TagIdList,
		Category:  req.Category,
		External:  req.External,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeInfoList(ctx, resp), nil
}

// RagSearchQABase 查询问答列表（命中测试）
func RagSearchQABase(ctx *gin.Context, req *request.RagSearchQABaseReq) ([]byte, int) {
	list, err := selectKnowledgeListByIdList(ctx, &request.KnowledgeBatchSelectReq{UserId: req.UserId, KnowledgeIdList: req.KnowledgeIdList})
	if err != nil {
		return response.CommonRagKnowledgeError(err)
	}
	if len(list.KnowledgeList) == 0 {
		return response.CommonRagKnowledgeError(errors.New("no knowledge permit"))
	}
	req.QAUser = buildUserQAList(list)
	// 构建 rag 请求
	return requestRagSearchQABase(ctx, req)
}

// KnowledgeStreamSearch 知识库流式问答
func KnowledgeStreamSearch(ctx *gin.Context, req *request.RagKnowledgeChatReq) error {
	list, err := selectKnowledgeListByIdList(ctx, &request.KnowledgeBatchSelectReq{UserId: req.UserId, KnowledgeIdList: req.KnowledgeIdList})
	if err != nil {
		return err
	}
	_, extendKnowledgeList := splitKnowledgeList(list)
	if len(extendKnowledgeList) == 0 { //先走原逻辑，先不去掉，等新逻辑稳定几个版本再说
		if len(list.KnowledgeList) == 0 {
			return errors.New("no knowledge permit")
		}
		req.KnowledgeUser, req.EnableVision = buildUserKnowledgeList(list.KnowledgeList)
		// 构建 rag 请求
		if req.AttachmentFiles == nil || !req.EnableVision {
			req.AttachmentFiles = []*request.RagKnowledgeAttachment{}
		}
		return requestRagKnowledgeStreamChat(ctx, req)
	}
	//新逻辑，只走rag的命中测试，不走ragChat接口
	hit, err := RagKnowledgeHit(ctx, buildHitParams(req))
	if err != nil {
		return err
	}
	//查询模型
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: req.CustomModelInfo.LlmModelID})
	if err != nil {
		return err
	}
	//执行模型调用和结果转换
	llmReq, historyItems := buildKnowledgeLLMReq(req, hit, modelInfo)
	ModelChatCompletions(ctx, modelInfo.ModelId, llmReq, ragLineProcessor(uuid.New().String(), req.Question, hit, historyItems))
	return nil
}

// SelectKnowledgeInfoByName 根据知识库名称查询知识库信息
func SelectKnowledgeInfoByName(ctx *gin.Context, userId, orgId string, r *request.SearchKnowledgeInfoReq) (interface{}, error) {
	resp, err := knowledgeBase.SelectKnowledgeDetailByName(ctx.Request.Context(), &knowledgebase_service.KnowledgeDetailSelectReq{
		UserId:        userId,
		OrgId:         orgId,
		KnowledgeName: r.KnowledgeName,
	})
	if err != nil {
		return nil, err
	}
	return map[string]string{
		"categoryId": resp.KnowledgeId,
	}, nil
}

// SelectKnowledgeIdByRagName 根据ragName获取知识库ID
func SelectKnowledgeIdByRagName(ctx *gin.Context, ragName string) (string, error) {
	resp, err := knowledgeBase.SelectKnowledgeIdByRagName(ctx.Request.Context(), &knowledgebase_service.SelectKnowledgeIdByRagNameReq{
		RagName: ragName,
	})
	if err != nil {
		return "", err
	}
	return resp.KnowledgeId, nil
}

// CreateKnowledge 创建知识库
func CreateKnowledge(ctx *gin.Context, userId, orgId string, r *request.CreateKnowledgeReq) (*response.CreateKnowledgeResp, error) {
	knowledgeGraph := &knowledgebase_service.KnowledgeGraph{}
	if r.Category == request.CategoryKnowledge || r.Category == request.CategoryMultimodalKnowledge {
		if r.KnowledgeGraph.Switch {
			knowledgeGraph = &knowledgebase_service.KnowledgeGraph{
				Switch:     r.KnowledgeGraph.Switch,
				LlmModelId: r.KnowledgeGraph.LLMModelId,
				SchemaUrl:  r.KnowledgeGraph.SchemaUrl,
			}
		}
	}
	resp, err := knowledgeBase.CreateKnowledge(ctx.Request.Context(), &knowledgebase_service.CreateKnowledgeReq{
		Name:        r.Name,
		Description: r.Description,
		UserId:      userId,
		OrgId:       orgId,
		EmbeddingModelInfo: &knowledgebase_service.EmbeddingModelInfo{
			ModelId: r.EmbeddingModel.ModelId,
		},
		KnowledgeGraph: knowledgeGraph,
		Category:       r.Category,
		AvatarPath:     r.Avatar.Key,
	})
	if err != nil {
		return nil, err
	}
	return &response.CreateKnowledgeResp{KnowledgeId: resp.KnowledgeId}, nil
}

func CreateKnowledgeOpenapi(ctx *gin.Context, userId, orgId string, r *request.CreateKnowledgeReq) (*response.CreateKnowledgeResp, error) {
	embModel, err := model.GetModelByUuid(ctx.Request.Context(), &model_service.GetModelByUuidReq{Uuid: r.EmbeddingModel.ModelId})
	if err != nil {
		return nil, err
	}
	// 校验所选模型为对应类型的 embedding 模型：多模态知识库需多模态 embedding，其余需文本 embedding
	if err := checkEmbeddingModelType(r.Category, embModel.ModelType); err != nil {
		return nil, err
	}
	r.EmbeddingModel.ModelId = embModel.ModelId
	if r.Category == request.CategoryKnowledge || r.Category == request.CategoryMultimodalKnowledge {
		if r.KnowledgeGraph.Switch {
			llmModel, err := model.GetModelByUuid(ctx.Request.Context(), &model_service.GetModelByUuidReq{Uuid: r.KnowledgeGraph.LLMModelId})
			if err != nil {
				return nil, err
			}
			// 校验知识图谱所选模型为 LLM 类型
			if llmModel.ModelType != mp.ModelTypeLLM {
				return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "knowledge graph requires an llm model")
			}
			r.KnowledgeGraph.LLMModelId = llmModel.ModelId
		}
	}
	return CreateKnowledge(ctx, userId, orgId, r)
}

// checkEmbeddingModelType 校验知识库 embedding 模型类型与知识库类型匹配
func checkEmbeddingModelType(category int32, modelType string) error {
	if category == request.CategoryMultimodalKnowledge {
		if modelType != mp.ModelTypeMultiEmbedding {
			return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "multimodal knowledge requires a multimodal-embedding model")
		}
		return nil
	}
	if modelType != mp.ModelTypeTextEmbedding {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "knowledge requires an embedding model")
	}
	return nil
}

// UpdateKnowledge 更新知识库
func UpdateKnowledge(ctx *gin.Context, userId, orgId string, r *request.UpdateKnowledgeReq) error {
	existingKnowledge, err := knowledgeBase.SelectKnowledgeDetailById(ctx.Request.Context(), &knowledgebase_service.KnowledgeDetailSelectReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: r.KnowledgeId,
	})
	if err != nil {
		return err
	}
	if err := utils.ValidateBriefUpdate(&r.Name, existingKnowledge.Name, &r.Description, existingKnowledge.Description, utils.SubjectKnowledge); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	_, err = knowledgeBase.UpdateKnowledge(ctx.Request.Context(), &knowledgebase_service.UpdateKnowledgeReq{
		KnowledgeId: r.KnowledgeId,
		Name:        r.Name,
		Description: r.Description,
		UserId:      userId,
		OrgId:       orgId,
		AvatarPath:  r.Avatar.Key,
	})
	return err
}

// DeleteKnowledge 删除知识库
func DeleteKnowledge(ctx *gin.Context, userId, orgId string, r *request.DeleteKnowledge) (interface{}, error) {
	resp, err := knowledgeBase.DeleteKnowledge(ctx.Request.Context(), &knowledgebase_service.DeleteKnowledgeReq{
		KnowledgeId: r.KnowledgeId,
		UserId:      userId,
		OrgId:       orgId,
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// KnowledgeHit 知识库命中
func KnowledgeHit(ctx *gin.Context, userId, orgId string, r *request.KnowledgeHitReq) (*response.KnowledgeHitResp, error) {
	matchParams := r.KnowledgeMatchParams
	resp, err := knowledgeBase.KnowledgeHit(ctx.Request.Context(), &knowledgebase_service.KnowledgeHitReq{
		Question:      r.Question,
		UserId:        userId,
		OrgId:         orgId,
		KnowledgeList: buildKnowledgeListReq(r),
		KnowledgeMatchParams: &knowledgebase_service.KnowledgeMatchParams{
			MatchType:         matchParams.MatchType,
			RerankModelId:     matchParams.RerankModelId,
			PriorityMatch:     matchParams.PriorityMatch,
			SemanticsPriority: matchParams.SemanticsPriority,
			KeywordPriority:   matchParams.KeywordPriority,
			TopK:              matchParams.TopK,
			Score:             matchParams.Threshold,
			TermWeight:        matchParams.TermWeight,
			TermWeightEnable:  matchParams.TermWeightEnable,
			UseGraph:          matchParams.UseGraph,
		},
		DocInfoList: buildKnowledgeHitDocInfoList(r.DocInfo),
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeHitResp(resp), nil
}

func checkRerank(ctx *gin.Context, rerankModelId, question string, hasImage bool) error {
	// 获取rerank模型信息
	var rerankModel *model_service.ModelInfo
	var err error
	// 纯图片搜索必选多模态rerank
	if question == "" && rerankModelId == "" {
		return errors.New("只输入图片必须选择多模态reranker")
	}
	if rerankModelId != "" {
		rerankModel, err = model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: rerankModelId})
		if err != nil {
			return err
		}
		if rerankModel == nil {
			return errors.New("所选reranker模型无法解析")
		}
		// 纯图片搜索必选多模态rerank
		if question == "" {
			if rerankModel.ModelType != mp.ModelTypeMultiRerank {
				return errors.New("只输入图片必须选择多模态reranker")
			}
		}
		// 包含图片搜索 - 若用户选了多模态rerank - 需查看模型是否支持图片搜索
		if hasImage {
			if rerankModel.ModelType == mp.ModelTypeMultiRerank {
				cfg := mp_jina.MultiModalRerank{}
				if err := json.Unmarshal([]byte(rerankModel.ProviderConfig), &cfg); err != nil {
					return errors.New("所选多模态reranker模型无法解析")
				}
				if !cfg.SupportImageInQuery {
					return errors.New("所选多模态reranker模型不支持输入图片")
				}
			}
		}
	}
	return nil
}

func KnowledgeHitOpenapi(ctx *gin.Context, userId, orgId string, r *request.KnowledgeHitReq) (*response.KnowledgeHitResp, error) {
	if r.KnowledgeMatchParams.RerankModelId != "" {
		rerankModelId, err := GetModelIdByUuid(ctx, r.KnowledgeMatchParams.RerankModelId)
		if err != nil {
			return nil, err
		}
		r.KnowledgeMatchParams.RerankModelId = rerankModelId
	}
	return KnowledgeHit(ctx, userId, orgId, r)
}

func GetKnowledgeMetaSelect(ctx *gin.Context, userId, orgId string, r *request.GetKnowledgeMetaSelectReq) (*response.GetKnowledgeMetaSelectResp, error) {
	metaList, err := knowledgeBase.GetKnowledgeMetaSelect(ctx.Request.Context(), &knowledgebase_service.SelectKnowledgeMetaReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: r.KnowledgeId,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeMetaList(metaList.MetaList), nil
}

// GetKnowledgeMetaValueList 获取文档元数据列表
func GetKnowledgeMetaValueList(ctx *gin.Context, userId, orgId string, r *request.KnowledgeMetaValueListReq) (*response.KnowledgeMetaValueListResp, error) {
	resp, err := knowledgeBase.GetKnowledgeMetaValueList(ctx.Request.Context(), &knowledgebase_service.KnowledgeMetaValueListReq{
		UserId:    userId,
		OrgId:     orgId,
		DocIdList: r.DocIdList,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeMetaValueRespList(resp), nil
}

// UpdateKnowledgeMetaValue 更新知识库元数据值
func UpdateKnowledgeMetaValue(ctx *gin.Context, userId, orgId string, r *request.UpdateMetaValueReq) error {
	_, err := knowledgeBase.UpdateKnowledgeMetaValue(ctx.Request.Context(), &knowledgebase_service.UpdateKnowledgeMetaValueReq{
		UserId:          userId,
		OrgId:           orgId,
		ApplyToSelected: r.ApplyToSelected,
		DocIdList:       r.DocIdList,
		MetaList:        buildKnowledgeMetaValueReqList(r.MetaValueList),
		KnowledgeId:     r.KnowledgeId,
	})
	if err != nil {
		return err
	}
	return nil
}

func UpdateKnowledgeStatus(ctx *gin.Context, r *request.CallbackUpdateKnowledgeStatusReq) error {
	_, err := knowledgeBase.UpdateKnowledgeStatus(ctx.Request.Context(), &knowledgebase_service.UpdateKnowledgeStatusReq{
		KnowledgeId:  r.KnowledgeId,
		ReportStatus: r.ReportStatus,
	})
	return err
}

// GetKnowledgeGraph 查询知识图谱详情
func GetKnowledgeGraph(ctx *gin.Context, userId, orgId string, req *request.KnowledgeGraphReq) (*response.KnowledgeGraphResp, error) {
	resp, err := knowledgeBase.GetKnowledgeGraph(ctx.Request.Context(), &knowledgebase_service.KnowledgeGraphReq{
		UserId:      userId,
		OrgId:       orgId,
		KnowledgeId: req.KnowledgeId,
	})
	if err != nil {
		return nil, err
	}
	graph := &response.KnowledgeGraphResp{
		ProcessingCount: resp.ProcessingCount,
		SuccessCount:    resp.SuccessCount,
		FailCount:       resp.FailedCount,
		Total:           resp.Total,
	}
	if err = json.Unmarshal([]byte(resp.Schema), graph); err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("knowledge graph unmarshal err: %v", err))
	}
	return graph, nil
}

func buildUserKnowledgeList(knowledgeList []*response.KnowledgeInfo) (map[string][]*request.RagKnowledgeInfo, bool) {
	retMap := make(map[string][]*request.RagKnowledgeInfo)
	var enableVision bool
	for _, knowledge := range knowledgeList {
		knowledgeInfos, exist := retMap[knowledge.CreateUserId]
		if !exist {
			knowledgeInfos = make([]*request.RagKnowledgeInfo, 0)
		}
		knowledgeInfos = append(knowledgeInfos, &request.RagKnowledgeInfo{
			KnowledgeId:   knowledge.KnowledgeId,
			KnowledgeName: knowledge.RagName,
		})
		retMap[knowledge.CreateUserId] = knowledgeInfos
		if knowledge.Category == constant.KnowledgeMultiModal {
			enableVision = true
		}
	}
	return retMap, enableVision
}

func buildUserQAList(knowledgeList *response.KnowledgeListResp) map[string][]*request.RagQaInfo {
	retMap := make(map[string][]*request.RagQaInfo)
	for _, knowledge := range knowledgeList.KnowledgeList {
		knowledgeInfos, exist := retMap[knowledge.CreateUserId]
		if !exist {
			knowledgeInfos = make([]*request.RagQaInfo, 0)
		}
		knowledgeInfos = append(knowledgeInfos, &request.RagQaInfo{
			QaBaseId:   knowledge.KnowledgeId,
			QaBaseName: knowledge.RagName,
		})
		retMap[knowledge.CreateUserId] = knowledgeInfos
	}
	return retMap
}

// selectKnowledgeListByIdList 查询知识库列表，主要根据userId 查询用户所有知识库
func selectKnowledgeListByIdList(ctx *gin.Context, req *request.KnowledgeBatchSelectReq) (*response.KnowledgeListResp, error) {
	resp, err := knowledgeBase.SelectKnowledgeListByIdList(ctx.Request.Context(), &knowledgebase_service.BatchKnowledgeSelectReq{
		UserId:          req.UserId,
		KnowledgeIdList: req.KnowledgeIdList,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeInfoList(ctx, resp), nil
}

// buildKnowledgeMetaList 构造知识库元数据列表
func buildKnowledgeMetaList(metaList []*knowledgebase_service.KnowledgeMetaData) *response.GetKnowledgeMetaSelectResp {
	var retMetaList []*response.KnowledgeMetaItem
	for _, meta := range metaList {
		retMetaList = append(retMetaList, &response.KnowledgeMetaItem{
			MetaKey:       meta.Key,
			MetaValueType: meta.Type,
			MetaId:        meta.MetaId,
		})
	}
	return &response.GetKnowledgeMetaSelectResp{MetaList: retMetaList}
}

// buildKnowledgeListReq 构造命中测试 - 知识库列表参数
func buildKnowledgeListReq(r *request.KnowledgeHitReq) []*knowledgebase_service.KnowledgeParams {
	var knowledgeList []*knowledgebase_service.KnowledgeParams
	for _, k := range r.KnowledgeList {
		filterParams := k.MetaDataFilterParams
		knowledgeList = append(knowledgeList, &knowledgebase_service.KnowledgeParams{
			KnowledgeId:          k.ID,
			MetaDataFilterParams: buildMetaDataFilterParams(filterParams),
		})
	}
	return knowledgeList
}

func buildMetaDataFilterParams(filterParams *request.MetaDataFilterParams) *knowledgebase_service.MetaDataFilterParams {
	if filterParams == nil {
		return &knowledgebase_service.MetaDataFilterParams{}
	}
	return &knowledgebase_service.MetaDataFilterParams{
		FilterEnable:     filterParams.FilterEnable,
		FilterLogicType:  filterParams.FilterLogicType,
		MetaFilterParams: buildMetaFilterParams(filterParams.MetaFilterParams),
	}
}

// buildKnowledgeInfoList 构造知识库列表结果
func buildKnowledgeInfoList(ctx *gin.Context, knowledgeListResp *knowledgebase_service.KnowledgeSelectListResp) *response.KnowledgeListResp {
	if knowledgeListResp == nil || len(knowledgeListResp.KnowledgeList) == 0 {
		return &response.KnowledgeListResp{}
	}
	orgMap := buildOtherOrgInfoMap(ctx, knowledgeListResp)
	var list []*response.KnowledgeInfo
	for _, knowledge := range knowledgeListResp.KnowledgeList {
		avatar := cacheKnowledgeAvatar(ctx, knowledge.AvatarPath, knowledge.Category)
		if avatar.Path != "" {
			avatar.Path = config.Cfg().Server.ApiBaseUrl + avatar.Path
		}
		share := knowledge.ShareCount > 1
		list = append(list, &response.KnowledgeInfo{
			KnowledgeId: knowledge.KnowledgeId,
			Name:        knowledge.Name,
			OrgName:     buildShareOrgName(share, orgMap[knowledge.CreateOrgId]),
			Description: knowledge.Description,
			DocCount:    int(knowledge.DocCount),
			EmbeddingModelInfo: &response.EmbeddingModelInfo{
				ModelId: knowledge.EmbeddingModelInfo.ModelId,
			},
			KnowledgeTagList: buildTagList(knowledge.KnowledgeTagInfoList),
			CreateAt:         knowledge.CreatedAt,
			PermissionType:   knowledge.PermissionType,
			CreateUserId:     knowledge.CreateUserId,
			Share:            share, //数量大于1才是分享，因为权限记录中有一条是记录创建者权限
			RagName:          knowledge.RagName,
			GraphSwitch:      knowledge.GraphSwitch,
			Category:         knowledge.Category,
			LlmModelId:       knowledge.LlmModelId,
			UpdatedAt:        knowledge.UpdatedAt,
			External:         knowledge.External,
			ExternalKnowledgeInfo: &response.KnowledgeExternalInfo{
				ExternalKnowledgeId:   knowledge.KnowledgeExternalInfo.ExternalKnowledgeId,
				ExternalKnowledgeName: knowledge.KnowledgeExternalInfo.ExternalKnowledgeName,
				ExternalSource:        knowledge.KnowledgeExternalInfo.Provider,
				ExternalApiId:         knowledge.KnowledgeExternalInfo.ExternalAPIId,
				ExternalApiName:       knowledge.KnowledgeExternalInfo.ExternalAPIName,
			},
			Avatar: avatar,
		})
	}
	return &response.KnowledgeListResp{KnowledgeList: list}
}

//nolint:staticcheck
func buildShareOrgName(share bool, orgName string) string {
	if share {
		if strings.Contains(orgName, "---") {
			// "--- 系统 ---" => "系统"
			return strings.TrimSpace(strings.Trim(orgName, "---"))
		}
		return orgName
	}
	return ""
}

// buildOtherOrgInfoMap 构造刨除当前组织的组织信息
func buildOtherOrgInfoMap(ctx *gin.Context, knowledgeListResp *knowledgebase_service.KnowledgeSelectListResp) map[string]string {
	var shareOrgIdList []string
	for _, knowledge := range knowledgeListResp.KnowledgeList {
		if knowledge.ShareCount > 1 {
			shareOrgIdList = append(shareOrgIdList, knowledge.CreateOrgId)
		}
	}
	var dataMap = make(map[string]string)
	if len(shareOrgIdList) > 0 {
		orgInfoList, err := iam.GetOrgByOrgIDs(ctx.Request.Context(), &iam_service.GetOrgByOrgIDsReq{
			OrgIds: shareOrgIdList,
		})
		if err != nil {
			log.Errorf("get share org info error: %v", err)
		}
		if orgInfoList != nil && len(orgInfoList.Orgs) > 0 {
			for _, org := range orgInfoList.Orgs {
				dataMap[org.Id] = org.Name
			}
		}
	}
	return dataMap
}

// buildTagList 构造知识库标签列表
func buildTagList(tagList []*knowledgebase_service.KnowledgeTagInfo) []*response.KnowledgeTag {
	var retTagList = make([]*response.KnowledgeTag, 0)
	if len(tagList) > 0 {
		for _, tag := range tagList {
			retTagList = append(retTagList, &response.KnowledgeTag{
				TagId:    tag.TagId,
				TagName:  tag.TagName,
				Selected: true,
			})
		}
	}
	return retTagList
}

// buildKnowledgeHitResp 构造知识库命中返回
func buildKnowledgeHitResp(resp *knowledgebase_service.KnowledgeHitResp) *response.KnowledgeHitResp {
	var searchList = make([]*response.ChunkSearchList, 0)
	if len(resp.SearchList) > 0 {
		for _, search := range resp.SearchList {
			childContentList := make([]*response.ChildContent, 0)
			for _, child := range search.ChildContentList {
				childContentList = append(childContentList, &response.ChildContent{
					ChildSnippet: child.ChildSnippet,
					Score:        float64(child.Score),
				})
			}
			childScore := make([]float64, 0)
			for _, score := range search.ChildScore {
				childScore = append(childScore, float64(score))
			}
			searchList = append(searchList, &response.ChunkSearchList{
				Title:            search.Title,
				Snippet:          search.Snippet,
				KnowledgeName:    search.KnowledgeName,
				ChildContentList: childContentList,
				ChildScore:       childScore,
				ContentType:      search.ContentType,
				Score:            float64(search.Score),
				RerankInfo:       buildRerankInfo(search.RerankInfo),
			})
		}
	}
	return &response.KnowledgeHitResp{
		Prompt:     resp.Prompt,
		Score:      resp.Score,
		SearchList: searchList,
		UseGraph:   resp.UseGraph,
	}
}

func buildRerankInfo(rerankInfo []*knowledgebase_service.RerankInfo) []*response.RerankInfo {
	rerankInfoList := make([]*response.RerankInfo, 0)
	if len(rerankInfo) > 0 {
		for _, r := range rerankInfo {
			rerankInfoList = append(rerankInfoList, &response.RerankInfo{
				FileUrl: r.FileUrl,
				Score:   float64(r.Score),
				Type:    r.Type,
			})
		}
	}
	return rerankInfoList
}

func buildMetaFilterParams(meta []*request.MetaFilterParams) []*knowledgebase_service.MetaFilterParams {
	if len(meta) == 0 {
		return make([]*knowledgebase_service.MetaFilterParams, 0)
	}
	var metaList []*knowledgebase_service.MetaFilterParams
	for _, m := range meta {
		metaList = append(metaList, &knowledgebase_service.MetaFilterParams{
			Key:       m.Key,
			Value:     m.Value,
			Type:      m.Type,
			Condition: m.Condition,
		})
	}
	return metaList
}

func buildKnowledgeMetaValueRespList(resp *knowledgebase_service.KnowledgeMetaValueListResp) *response.KnowledgeMetaValueListResp {
	retList := make([]*response.KnowledgeMetaValues, 0)
	for _, meta := range resp.MetaList {
		retList = append(retList, &response.KnowledgeMetaValues{
			MetaId:        meta.MetaId,
			MetaKey:       meta.Key,
			MetaValue:     meta.ValueList,
			MetaValueType: meta.Type,
		})
	}
	return &response.KnowledgeMetaValueListResp{
		KnowledgeMetaValues: retList,
	}
}

func buildKnowledgeMetaValueReqList(req []*request.DocMetaData) []*knowledgebase_service.MetaValueOperation {
	metaList := make([]*knowledgebase_service.MetaValueOperation, 0)
	for _, meta := range req {
		metaList = append(metaList, &knowledgebase_service.MetaValueOperation{
			MetaInfo: &knowledgebase_service.KnowledgeMetaData{
				MetaId: meta.MetaId,
				Key:    meta.MetaKey,
				Value:  meta.MetaValue,
				Type:   meta.MetaValueType,
			},
			Option: meta.Option,
		})
	}
	return metaList
}

// requestRagSearchQABase 请求rag
func requestRagSearchQABase(ctx context.Context, req *request.RagSearchQABaseReq) ([]byte, int) {
	url := config.Cfg().RagKnowledgeConfig.Endpoint + config.Cfg().RagKnowledgeConfig.SearchQABaseUri
	paramsByte, err := json.Marshal(req)
	if err != nil {
		return response.CommonRagKnowledgeError(err)
	}
	result, err := knowHttp.PostJsonOriResp(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		MonitorKey: "rag_search_qa_base",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return response.CommonRagKnowledgeError(err)
	}
	body, err := http_client.ReadHttpResp(result)
	if err != nil {
		return response.CommonRagKnowledgeError(err)
	}
	return body, result.StatusCode
}

func requestRagKnowledgeStreamChat(ctx *gin.Context, req *request.RagKnowledgeChatReq) error {
	params, err := buildRagKnowledgeChatHttpParams(req)
	if err != nil {
		log.Errorf("build http params fail %s", err.Error())
		return err
	}
	// 捕获 panic 并记录日志（不重新抛出，避免崩溃）
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("RagStreamChat panic: %v", r)
		}
	}()

	resp, err := knowHttp.PostJsonOriResp(ctx, params)
	if err != nil {
		errMsg := fmt.Sprintf("error: 调用下游服务异常: %v", err)
		log.Errorf(errMsg)
		return err
	}
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: 调用下游服务异常: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error: 响应体关闭异常: %v", err)
		}
	}(resp.Body) // 确保响应体关闭

	if resp.StatusCode != http.StatusOK {
		errMsg := fmt.Sprintf("error: 调用下游服务异常: %s", resp.Status)
		log.Errorf(errMsg)
		return errors.New(errMsg)
	}
	return writeSSE(ctx, resp)
}

func buildRagKnowledgeChatHttpParams(req *request.RagKnowledgeChatReq) (*http_client.HttpRequestParams, error) {
	url := fmt.Sprintf("%s%s", config.Cfg().RagKnowledgeConfig.ChatEndpoint, config.Cfg().RagKnowledgeConfig.KnowledgeChatUri)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	return &http_client.HttpRequestParams{
		Url:        url,
		Body:       body,
		Headers:    map[string]string{"X-uid": req.UserId},
		Timeout:    time.Minute * 10,
		MonitorKey: "rag_search_service",
		LogLevel:   http_client.LogAll,
	}, nil
}

func buildKnowledgeExternalAPI(resp *knowledgebase_service.KnowledgeExternalAPISelectListResp) *response.KnowledgeExternalAPIListResp {
	var externalAPIList []*response.KnowledgeExternalAPIInfo
	for _, externalApi := range resp.ExternalAPIList {
		externalAPIList = append(externalAPIList, &response.KnowledgeExternalAPIInfo{
			ExternalAPIId: externalApi.ExternalAPIId,
			Name:          externalApi.Name,
			Description:   externalApi.Description,
			BaseUrl:       externalApi.BaseUrl,
			ApiKey:        externalApi.ApiKey,
		})
	}
	return &response.KnowledgeExternalAPIListResp{ExternalAPIList: externalAPIList}
}

func buildKnowledgeExternal(resp *knowledgebase_service.KnowledgeExternalSelectListResp) *response.KnowledgeExternalListResp {
	var externalKnowledgeList []*response.KnowledgeExternalBriefInfo
	for _, externalKnowledge := range resp.ExternalKnowledgeList {
		externalKnowledgeList = append(externalKnowledgeList, &response.KnowledgeExternalBriefInfo{
			ExternalKnowledgeId:   externalKnowledge.ExternalKnowledgeId,
			ExternalKnowledgeName: externalKnowledge.ExternalKnowledgeName,
			ExternalApiId:         externalKnowledge.ExternalAPIId,
		})
	}
	return &response.KnowledgeExternalListResp{ExternalKnowledgeList: externalKnowledgeList}
}

// buildHitParams 构造命中测试参数
func buildHitParams(req *request.RagKnowledgeChatReq) *request.RagSearchKnowledgeBaseReq {
	if req.AttachmentFiles == nil {
		req.AttachmentFiles = []*request.RagKnowledgeAttachment{}
	}
	return &request.RagSearchKnowledgeBaseReq{
		KnowledgeUser:        req.KnowledgeUser,
		UseGraph:             req.UseGraph,
		UserId:               req.UserId,
		Question:             req.Question,
		KnowledgeIdList:      req.KnowledgeIdList,
		Threshold:            float64(req.Threshold),
		TopK:                 req.TopK,
		RerankModelId:        req.RerankModelId,
		RerankMod:            req.RerankMod,
		RetrieveMethod:       req.RetrieveMethod,
		Weight:               req.Weight,
		TermWeight:           req.TermWeight,
		MetaFilter:           req.MetaFilter,
		MetaFilterConditions: req.MetaFilterConditions,
		AutoCitation:         req.AutoCitation,
		RewriteQuery:         req.RewriteQuery,
		EnableVision:         req.EnableVision,
		AttachmentFiles:      req.AttachmentFiles,
	}
}

// rag 的数据转行行处理器
func ragLineProcessor(messageId, query string, hitData *KnowledgeHitData, history []*request.HistoryItem) func(resp *mp_common.LLMResp) string {
	contentBuilder := strings.Builder{}
	return func(resp *mp_common.LLMResp) string {
		defer utils.PrintPanicStack()
		if len(resp.Choices) > 0 {
			var finish = 0
			switch finishReason := resp.Choices[0].FinishReason; finishReason {
			case "stop":
				finish = 1
			case "sensitive_cancel":
				finish = 4
			}

			choice := resp.Choices[0]
			var content = ""
			if choice.Delta != nil {
				content = choice.Delta.Content
			} else if choice.Message != nil {
				content = choice.Message.Content
			}

			contentBuilder.WriteString(content)

			var historyTemp []*request.HistoryItem
			if len(history) > 0 {
				historyTemp = append(historyTemp, history...)
			}
			historyTemp = append(historyTemp, &request.HistoryItem{
				Query:       query,
				Response:    contentBuilder.String(),
				NeedHistory: true,
			})

			// 构建响应信息
			responseInfo := &RagResponseInfo{
				Code:    0,
				Message: "success",
				MsgID:   messageId,
				Data: &RagData{
					Score:      hitData.Score,
					Output:     content,
					SearchList: hitData.SearchList,
				},
				History: historyTemp,
				Finish:  finish,
			}

			marshal, err := json.Marshal(responseInfo)
			if err == nil {
				return string(marshal)
			}
		}
		return ""
	}
}

// buildKnowledgeLLMReq 构造知识库llm请求
func buildKnowledgeLLMReq(req *request.RagKnowledgeChatReq, hitData *KnowledgeHitData, modelInfo *model_service.ModelInfo) (*mp_common.LLMReq, []*request.HistoryItem) {
	var streamValue = true
	message, historyItems := buildHistory(req)
	message = append(message, mp_common.OpenAIReqMsg{
		Role:    mp_common.MsgRoleUser,
		Content: hitData.Prompt,
	})
	return &mp_common.LLMReq{
		Model:             modelInfo.Model,
		Messages:          message,
		Stream:            &streamValue,
		Temperature:       buildFloatValue(req.Temperature),
		TopP:              buildFloatValue(req.TopP),
		TopK:              buildIntValue(req.TopK),
		RepetitionPenalty: buildFloatValue(req.RepetitionPenalty),
	}, historyItems
}

func buildHistory(req *request.RagKnowledgeChatReq) ([]mp_common.OpenAIReqMsg, []*request.HistoryItem) {
	messageList := make([]mp_common.OpenAIReqMsg, 0)
	historyLen := len(req.History)
	maxHistory := int(req.MaxHistory)
	var historyItems = req.History
	if maxHistory > 0 && historyLen > 0 {
		if historyLen <= maxHistory {
			historyItems = req.History[historyLen-maxHistory:]
		}
		for _, v := range historyItems {
			messageList = append(messageList, mp_common.OpenAIReqMsg{
				Role:    mp_common.MsgRoleUser,
				Content: v.Query,
			})
			messageList = append(messageList, mp_common.OpenAIReqMsg{
				Role:    mp_common.MsgRoleAssistant,
				Content: v.Response,
			})
		}
	}
	return messageList, historyItems
}

func buildIntValue(value int32) *int {
	f := int(value)
	return &f
}

func buildFloatValue(value float32) *float64 {
	f := float64(value)
	return &f
}

func buildKnowledgeHitDocInfoList(docInfoList []*request.DocInfo) []*knowledgebase_service.DocFileInfo {
	retList := make([]*knowledgebase_service.DocFileInfo, 0)
	if len(docInfoList) > 0 {
		for _, docInfo := range docInfoList {
			retList = append(retList, &knowledgebase_service.DocFileInfo{
				DocId:   docInfo.DocId,
				DocName: docInfo.DocName,
				DocSize: docInfo.DocSize,
				DocType: docInfo.DocType,
				DocUrl:  docInfo.DocUrl,
			})
		}
	}
	return retList
}

// GetDocByKnowledgeNameAndDocName 根据知识库名称和文档名称获取文档信息
func GetDocByKnowledgeNameAndDocName(ctx *gin.Context, userId, orgId string, req *request.GetDocByKnowledgeNameAndDocNameReq) (*response.GetDocByKnowledgeNameAndDocNameResp, error) {
	resp, err := knowledgeBase.GetDocByKnowledgeNameAndDocName(ctx.Request.Context(), &knowledgebase_service.GetDocByKnowledgeNameAndDocNameReq{
		KnowledgeName: req.KnowledgeName,
		DocName:       req.DocName,
		UserId:        userId,
		OrgId:         orgId,
	})
	if err != nil {
		return nil, err
	}
	return &response.GetDocByKnowledgeNameAndDocNameResp{
		KnowledgeId:   resp.KnowledgeId,
		KnowledgeName: resp.KnowledgeName,
		DocId:         resp.DocId,
		DocType:       resp.DocType,
		DocName:       resp.DocName,
		Disable:       false,
	}, nil
}
