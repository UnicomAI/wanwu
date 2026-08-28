package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"sort"
	"strings"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_doc_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-doc-service"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	prompt_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/prompt-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	successCode = 0
	errorCode   = 1
	difyHitRate = 0.6
)

var ragKnowHttp = http_client.CreateDefault()

type KnowledgeHitParams struct {
}

type RagKnowledgeHitResp struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Data    *KnowledgeHitData `json:"data"`
}

type KnowledgeHitData struct {
	Prompt     string             `json:"prompt"`
	SearchList []*ChunkSearchList `json:"searchList"`
	Score      []float64          `json:"score"`
	UseGraph   bool               `json:"use_graph"`
}

type ChunkSearchList struct {
	Title            string           `json:"title"`
	Snippet          string           `json:"snippet"`
	KbName           string           `json:"kb_name"`
	UserKbName       string           `json:"user_kb_name"`
	MetaData         interface{}      `json:"meta_data"`
	ChildContentList []*ChildContent  `json:"child_content_list"`
	ChildScore       []float64        `json:"child_score"`
	ContentType      string           `json:"content_type"` // graph：知识图谱（文本）, text：文档分段（文本）, community_report：社区报告（markdown）
	Score            float64          `json:"score"`
	RerankInfo       []*RagRerankInfo `json:"rerank_info"`
}

type ChildContent struct {
	ChildSnippet string  `json:"child_snippet"`
	Score        float64 `json:"score"`
}

type RagRerankInfo struct {
	Type    string  `json:"type"`
	FileUrl string  `json:"file_url"`
	Score   float64 `json:"score"`
}

type RagKnowledgeSearchContext struct {
	HasLocal           bool
	LocalKnowledgeData *KnowledgeHitData `json:"localKnowledgeData"`
	LocalHitErr        error
	HasDify            bool
	DifyKnowledgeData  *KnowledgeHitData `json:"difyKnowledgeData"`
	DifyHitErr         error
}

type ChunkSearchData struct {
	Search *ChunkSearchList
	Score  float64
}

func RagKnowledgeHit(ctx *gin.Context, req *request.RagSearchKnowledgeBaseReq) (*KnowledgeHitData, error) {
	//查询知识库详情
	list, err := selectKnowledgeListByIdList(ctx, &request.KnowledgeBatchSelectReq{UserId: req.UserId, KnowledgeIdList: req.KnowledgeIdList})
	if err != nil {
		return nil, err
	}
	if len(list.KnowledgeList) == 0 {
		return nil, grpc_util.ErrorStatus(err_code.Code_KnowledgePermissionDeny, "")
	}
	//并发进行命中测试
	searchResult := BatchRagKnowledgeSearch(ctx, req, list)
	err = checkHitErr(searchResult)
	if err != nil {
		log.Errorf("命中测试失败 err %v", err)
		return nil, grpc_util.ErrorStatus(err_code.Code_KnowledgeBaseHitFailed, err.Error())
	}
	//合并结果返回
	return fillUserKbName(mergeHitResult(req, searchResult), list.KnowledgeList), nil
}

// BatchRagKnowledgeSearch 批量rag知识库查询，后面多智能体版本合并后可以使用并发框架改造
func BatchRagKnowledgeSearch(ctx *gin.Context, req *request.RagSearchKnowledgeBaseReq, list *response.KnowledgeListResp) *RagKnowledgeSearchContext {
	//先把架子搭出来，后续优化
	ragSearchContext := &RagKnowledgeSearchContext{}
	knowledgeList, extendKnowledgeList := splitKnowledgeList(list)
	localHit := localKnowledgeHit(ctx, req, ragSearchContext, knowledgeList)
	difyHit := difyKnowledgeHit(ctx, req, ragSearchContext, extendKnowledgeList)
	//并发调用
	safe_go_util.SageGoWaitGroup(localHit, difyHit)
	return ragSearchContext
}

// RagLocalKnowledgeHit rag本地知识库命中测试
func RagLocalKnowledgeHit(ctx context.Context, knowledgeHitParams *request.RagSearchKnowledgeBaseReq) (*RagKnowledgeHitResp, error) {
	knowledgeConfig := config.Cfg().RagKnowledgeConfig
	url := knowledgeConfig.Endpoint + knowledgeConfig.SearchKnowledgeBaseUri
	paramsByte, err := json.Marshal(knowledgeHitParams)
	if err != nil {
		return nil, err
	}
	result, err := ragKnowHttp.PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		Timeout:    time.Duration(knowledgeConfig.SearchKnowTimeout) * time.Second,
		MonitorKey: "rag_knowledge_hit",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var resp RagKnowledgeHitResp
	if err := json.Unmarshal(result, &resp); err != nil {
		log.Errorf(err.Error())
		return nil, err
	}
	if resp.Code != successCode {
		return nil, errors.New(resp.Message)
	}
	return &resp, nil
}

func BatchRagDifyKnowledgeHit(ctx context.Context, req *request.RagSearchKnowledgeBaseReq, knowledgeList []*response.KnowledgeInfo) (*RagKnowledgeHitResp, error) {
	var searchList []*ChunkSearchList
	var scores []float64
	var errCount int
	var hitErrors string
	hitDatas := make([]*RagKnowledgeHitResp, len(knowledgeList))
	var funcList = buildBatchDifyKnowHit(ctx, req, hitDatas, knowledgeList)
	//并发调用
	safe_go_util.SageGoWaitGroup(funcList...)
	searchDatas := make([]*ChunkSearchData, 0)
	for i, hit := range hitDatas {
		if hit.Code != successCode {
			errCount++
			hitErrors = hitErrors + fmt.Sprintf("knowledge: %v err: %v ;", knowledgeList[i].Name, hit.Message)
			continue
		}
		for idx := range hit.Data.SearchList {
			searchDatas = append(searchDatas, &ChunkSearchData{
				Search: hit.Data.SearchList[idx],
				Score:  hit.Data.Score[idx],
			})
		}
	}
	if errCount >= len(knowledgeList) {
		return nil, errors.New(hitErrors)
	}
	sort.Slice(searchDatas, func(i, j int) bool {
		return searchDatas[i].Score > searchDatas[j].Score
	})
	for idx, search := range searchDatas {
		if int32(idx) >= req.TopK {
			break
		}
		searchList = append(searchList, search.Search)
		scores = append(scores, search.Score)
	}
	return &RagKnowledgeHitResp{
		Code: successCode,
		Data: &KnowledgeHitData{
			SearchList: searchList,
			Score:      scores,
			Prompt:     buildPrompt(req.Question, searchList),
		},
		Message: "",
	}, nil
}

func RagDifyKnowledgeHit(ctx context.Context, req *request.RagSearchKnowledgeBaseReq, knowledge *response.KnowledgeInfo) (*KnowledgeHitData, error) {
	// 1.查询外部知识库信息
	externalKnowledgeInfo, err := knowledgeBase.SelectKnowledgeExternalInfo(ctx, &knowledgebase_service.KnowledgeExternalInfoSelectReq{
		KnowledgeId: knowledge.KnowledgeId,
	})
	if err != nil {
		return nil, err
	}
	// 2.调用dify命中测试接口
	hitData := &KnowledgeHitData{}
	params := buildDifyHitParams(req, externalKnowledgeInfo)
	difyKnowledgeRetrieveResp, err := DifyKnowledgeRetrieve(ctx, externalKnowledgeInfo.ExternalAPIUrl, externalKnowledgeInfo.ExternalAPIKey, externalKnowledgeInfo.ExternalKnowledgeId, params)
	if err != nil {
		return nil, err
	}
	for _, retrieveDataRecord := range difyKnowledgeRetrieveResp.Records {
		chunkSearchList, score := buildChunkSearchList(retrieveDataRecord)
		hitData.SearchList = append(hitData.SearchList, chunkSearchList)
		hitData.Score = append(hitData.Score, score)
	}
	hitData.Prompt = buildPrompt(req.Question, hitData.SearchList)
	return hitData, err
}

func buildChunkSearchList(retrievalData *DifyKnowledgeHitRecords) (*ChunkSearchList, float64) {
	var childContents []*ChildContent
	var childScores []float64
	for _, childChunk := range retrievalData.ChildChunks {
		childContents = append(childContents, &ChildContent{
			ChildSnippet: childChunk.Content,
			Score:        childChunk.Score,
		})
		childScores = append(childScores, childChunk.Score)
	}
	chunkSearchList := &ChunkSearchList{
		Title:            retrievalData.Segment.Document.Name,
		Snippet:          retrievalData.Segment.Content,
		ChildContentList: childContents,
		ChildScore:       childScores,
	}
	return chunkSearchList, retrievalData.Score
}

func buildBatchDifyKnowHit(ctx context.Context, req *request.RagSearchKnowledgeBaseReq, hitDatas []*RagKnowledgeHitResp, knowledgeList []*response.KnowledgeInfo) []func() {
	var funcList []func()
	for idx, knowledgeInfo := range knowledgeList {
		funcList = append(funcList, func() {
			if len(knowledgeList) == 0 {
				return
			}
			hit, err := RagDifyKnowledgeHit(ctx, req, knowledgeInfo)
			if err != nil {
				hitDatas[idx] = &RagKnowledgeHitResp{
					Code:    errorCode,
					Message: err.Error(),
				}
				return
			}
			if hit != nil {
				hitDatas[idx] = &RagKnowledgeHitResp{
					Code: successCode,
					Data: hit,
				}
			}
		})
	}
	return funcList
}

func difyKnowledgeHit(ctx *gin.Context, req *request.RagSearchKnowledgeBaseReq, ragSearchContext *RagKnowledgeSearchContext, extendKnowledgeList []*response.KnowledgeInfo) func() {
	return func() {
		if len(extendKnowledgeList) == 0 {
			return
		}
		ragSearchContext.HasDify = true
		hit, err := BatchRagDifyKnowledgeHit(ctx, req, extendKnowledgeList)
		if err != nil {
			ragSearchContext.DifyHitErr = err
			return
		}
		if hit != nil {
			ragSearchContext.DifyKnowledgeData = hit.Data
		}
	}
}

func localKnowledgeHit(ctx *gin.Context, req *request.RagSearchKnowledgeBaseReq, ragSearchContext *RagKnowledgeSearchContext, knowledgeList []*response.KnowledgeInfo) func() {
	return func() {
		if len(knowledgeList) == 0 {
			return
		}
		ragSearchContext.HasLocal = true
		// 增加多模态知识库的校验
		err := checkRerank(ctx, req.RerankModelId, req.Question, len(req.AttachmentFiles) > 0)
		if err != nil {
			ragSearchContext.LocalHitErr = err
			return
		}
		hit, err := RagLocalKnowledgeHit(ctx, buildLocalHitParams(req, knowledgeList))
		if err != nil {
			ragSearchContext.LocalHitErr = err
			return
		}
		if hit != nil {
			hit = rebuildFileName(hit)
			ragSearchContext.LocalKnowledgeData = hit.Data
		}
	}
}

// rebuildFileName 重新构造文件名，主要处理url类型的文件
func rebuildFileName(hit *RagKnowledgeHitResp) (hit1 *RagKnowledgeHitResp) {
	defer util.PrintPanicStackWithCall(func(panicOccur bool, recoverError error) {
		if panicOccur {
			hit1 = hit
		}
	})
	if hit.Data != nil && len(hit.Data.SearchList) > 0 {
		var docIdMap = make(map[string]bool)
		var docIdList []string
		//根据文件名去重获取docId
		for _, item := range hit.Data.SearchList {
			//url 类会写死.txt
			if filepath.Ext(item.Title) != ".txt" {
				continue
			}
			baseName := strings.TrimSuffix(filepath.Base(item.Title), ".txt")
			if !docIdMap[baseName] {
				docIdMap[baseName] = true
				docIdList = append(docIdList, baseName)
			}
		}
		if len(docIdList) > 0 {
			//根据docId 批量查询文件详情
			docListResp, _ := knowledgeBaseDoc.GetDocListByDocIdList(context.Background(), &knowledgebase_doc_service.GetDocListByDocIdListReq{
				DocIdList: docIdList,
			})
			docMap := make(map[string]string)
			if docListResp != nil && len(docListResp.Docs) > 0 {
				for _, doc := range docListResp.Docs {
					docMap[doc.DocId] = doc.DocName
				}
			}
			if len(docMap) > 0 {
				//替换文件名
				for _, item := range hit.Data.SearchList {
					//url 类会写死.txt
					if filepath.Ext(item.Title) != ".txt" {
						continue
					}
					baseName := strings.TrimSuffix(filepath.Base(item.Title), ".txt")
					docName := docMap[baseName]
					if docName != "" {
						item.Title = docName + filepath.Ext(item.Title)
					}
				}
			}
		}
	}
	return hit
}

// buildLocalHitParams 构造本地查查询请求参数，注意深copy问题
func buildLocalHitParams(req *request.RagSearchKnowledgeBaseReq, knowledgeList []*response.KnowledgeInfo) *request.RagSearchKnowledgeBaseReq {
	knowledgeUser, enableVision := buildUserKnowledgeList(knowledgeList)
	req.KnowledgeUser = knowledgeUser
	req.EnableVision = enableVision
	if req.AttachmentFiles == nil || !req.EnableVision {
		req.AttachmentFiles = make([]*request.RagKnowledgeAttachment, 0)
	}
	return req
}

func buildDifyHitParams(req *request.RagSearchKnowledgeBaseReq, externalKnowledgeInfo *knowledgebase_service.KnowledgeExternalInfo) *DifyKnowledgeRetrieveParams {
	difyKnowledgeRetrieveParams := &DifyKnowledgeRetrieveParams{
		Query: req.Question,
		RetrievalModel: &DifyRetrievalModel{
			SearchMethod:          externalKnowledgeInfo.RetrievalModelInfo.SearchMethod,
			RerankingEnable:       externalKnowledgeInfo.RetrievalModelInfo.RerankingEnable,
			RerankingMode:         externalKnowledgeInfo.RetrievalModelInfo.RerankingMode,
			RerankingModel:        buildDifyRerankingModel(externalKnowledgeInfo.RetrievalModelInfo),
			TopK:                  int64(req.TopK),
			ScoreThresholdEnabled: true,
			ScoreThreshold:        float32(req.Threshold),
			Weights:               buildDifyWeights(externalKnowledgeInfo.RetrievalModelInfo.Weights),
		},
	}
	return difyKnowledgeRetrieveParams
}

func buildDifyRerankingModel(model *knowledgebase_service.RetrievalModelInfo) *DifyRerankingModel {
	if model == nil || model.RerankingModel == nil {
		return nil
	}
	return &DifyRerankingModel{
		RerankingProviderName: model.RerankingModel.RerankingProviderName,
		RerankingModelName:    model.RerankingModel.RerankingModelName,
	}
}

func buildDifyWeights(weight *knowledgebase_service.Weights) *DifyWeights {
	if weight == nil {
		return nil
	}
	return &DifyWeights{
		WeightType: weight.WeightType,
		KeywordSetting: &DifyKeywordSetting{
			KeywordWeight: weight.KeywordSetting.KeywordWeight,
		},
		VectorSetting: &DifyVectorSetting{
			VectorWeight:          weight.VectorSetting.VectorWeight,
			EmbeddingProviderName: weight.VectorSetting.EmbeddingProviderName,
			EmbeddingModelName:    weight.VectorSetting.EmbeddingModelName,
		},
	}
}

func fillUserKbName(knowledgeHitData *KnowledgeHitData, knowledgeList []*response.KnowledgeInfo) *KnowledgeHitData {
	if knowledgeHitData == nil || len(knowledgeHitData.SearchList) == 0 {
		return knowledgeHitData
	}
	knowledgeNameMap := buildKnowledgeNameMap(knowledgeList)
	if len(knowledgeNameMap) == 0 {
		return knowledgeHitData
	}
	//填充用户知识库名称
	for _, chunk := range knowledgeHitData.SearchList {
		userKbName := knowledgeNameMap[chunk.KbName]
		if len(userKbName) > 0 {
			chunk.UserKbName = userKbName
		} else {
			chunk.UserKbName = chunk.KbName
		}
	}
	return knowledgeHitData
}

// mergeHitResult 合并命中测试结果
func mergeHitResult(req *request.RagSearchKnowledgeBaseReq, searchContext *RagKnowledgeSearchContext) *KnowledgeHitData {
	if searchContext.LocalKnowledgeData == nil {
		return searchContext.DifyKnowledgeData
	}
	if searchContext.DifyKnowledgeData == nil {
		return searchContext.LocalKnowledgeData
	}
	//小于等于0 先返回全部
	if req.TopK <= 0 {
		return mergeHitData(req.Question, searchContext.LocalKnowledgeData, searchContext.DifyKnowledgeData)
	}
	//处理topK 按比例划分
	difyHitLimit := int(math.Round(float64(req.TopK) * difyHitRate))
	difyData, difyLen := limitHitData(searchContext.DifyKnowledgeData, difyHitLimit)
	localData, _ := limitHitData(searchContext.LocalKnowledgeData, int(req.TopK)-difyLen)
	return mergeHitData(req.Question, localData, difyData)
}

func buildKnowledgeNameMap(knowledgeList []*response.KnowledgeInfo) map[string]string {
	knowledgeNameMap := make(map[string]string)
	for _, knowledge := range knowledgeList {
		knowledgeNameMap[knowledge.RagName] = knowledge.Name
	}
	return knowledgeNameMap
}

func mergeHitData(question string, localHitData *KnowledgeHitData, difyHitData *KnowledgeHitData) *KnowledgeHitData {
	var hitData = &KnowledgeHitData{}
	hitData.SearchList = append(hitData.SearchList, localHitData.SearchList...)
	hitData.Score = append(hitData.Score, localHitData.Score...)
	hitData.SearchList = append(hitData.SearchList, difyHitData.SearchList...)
	hitData.Score = append(hitData.Score, difyHitData.Score...)

	hitData.Prompt = buildPrompt(question, hitData.SearchList)
	return localHitData
}

func limitHitData(knowledgeData *KnowledgeHitData, limit int) (*KnowledgeHitData, int) {
	searchList := knowledgeData.SearchList
	score := knowledgeData.Score
	if len(searchList) > 0 {
		searchList = searchList[:limit]
		score = score[:limit]
	}
	return &KnowledgeHitData{SearchList: searchList, Score: score}, len(searchList)
}

// buildPrompt 构造提示词
func buildPrompt(question string, chunkList []*ChunkSearchList) string {
	var dataBuilder = &strings.Builder{}
	for _, list := range chunkList {
		dataBuilder.WriteString(list.Snippet)
		dataBuilder.WriteString("\n")
	}
	return prompt_util.RagPrompt(question, dataBuilder.String())
}

// 拆分知识库列表
func splitKnowledgeList(knowledgeList *response.KnowledgeListResp) (localKnowledgeList []*response.KnowledgeInfo, extendKnowledgeList []*response.KnowledgeInfo) {
	for _, knowledge := range knowledgeList.KnowledgeList {
		if knowledge.External == 0 {
			localKnowledgeList = append(localKnowledgeList, knowledge)
		} else {
			extendKnowledgeList = append(extendKnowledgeList, knowledge)
		}
	}
	return
}

// checkHitErr 校验命中测试失败
func checkHitErr(searchResult *RagKnowledgeSearchContext) error {
	if searchResult.LocalHitErr != nil && searchResult.DifyHitErr != nil {
		//都有错返回本地的
		return searchResult.LocalHitErr
	}
	//只有本地且报错了
	if searchResult.LocalHitErr != nil && !searchResult.HasDify {
		return searchResult.LocalHitErr
	}
	//只有dify且报错了
	if searchResult.DifyHitErr != nil && !searchResult.HasLocal {
		return searchResult.DifyHitErr
	}
	return nil
}
