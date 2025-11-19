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

// SelectKnowledgeList queries the knowledge base list, mainly querying all user knowledge bases based on userId.
func SelectKnowledgeList(ctx *gin.Context, userId, orgId string, req *request.KnowledgeSelectReq) (*response.KnowledgeListResp, error) {
	resp, err := knowledgeBase.SelectKnowledgeList(ctx.Request.Context(), &knowledgebase_service.KnowledgeSelectReq{
		UserId:    userId,
		OrgId:     orgId,
		Name:      req.Name,
		TagIdList: req.TagIdList,
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeInfoList(ctx, resp), nil
}

// RagSearchKnowledgeBase queries the knowledge base list (hit test)
func RagSearchKnowledgeBase(ctx *gin.Context, req *request.RagSearchKnowledgeBaseReq) ([]byte, int) {
	list, err := selectKnowledgeListByIdList(ctx, &request.KnowledgeBatchSelectReq{UserId: req.UserId, KnowledgeIdList: req.KnowledgeIdList})
	if err != nil {
		return response.CommonRagKnowledgeError(err)
	}
	if len(list.KnowledgeList) == 0 {
		return response.CommonRagKnowledgeError(errors.New("no knowledge permit"))
	}
	req.KnowledgeUser = buildUserKnowledgeList(list)
	// Build rag request
	return requestRagSearchKnowledgeBase(ctx, req)
}

// KnowledgeStreamSearch Knowledge base streaming Q&A
func KnowledgeStreamSearch(ctx *gin.Context, req *request.RagKnowledgeChatReq) error {
	list, err := selectKnowledgeListByIdList(ctx, &request.KnowledgeBatchSelectReq{UserId: req.UserId, KnowledgeIdList: req.KnowledgeIdList})
	if err != nil {
		return err
	}
	if len(list.KnowledgeList) == 0 {
		return errors.New("no knowledge permit")
	}
	req.KnowledgeUser = buildUserKnowledgeList(list)
	// Build rag request
	return requestRagKnowledgeStreamChat(ctx, req)
}

// SelectKnowledgeInfoByName Query knowledge base information based on the knowledge base name
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

// GetDeployInfo Query deployment information
func GetDeployInfo(ctx *gin.Context) (interface{}, error) {
	cfgServer := config.Cfg().Server
	return map[string]string{
		"webBaseUrl": cfgServer.WebBaseUrl + "/minio/download/api/",
	}, nil
}

// CreateKnowledge Create a knowledge base
func CreateKnowledge(ctx *gin.Context, userId, orgId string, r *request.CreateKnowledgeReq) (*response.CreateKnowledgeResp, error) {
	var llmModelId, schemaUrl string
	if r.KnowledgeGraph.Switch {
		llmModelId = r.KnowledgeGraph.LLMModelId
		schemaUrl = r.KnowledgeGraph.SchemaUrl
	}
	resp, err := knowledgeBase.CreateKnowledge(ctx.Request.Context(), &knowledgebase_service.CreateKnowledgeReq{
		Name:        r.Name,
		Description: r.Description,
		UserId:      userId,
		OrgId:       orgId,
		EmbeddingModelInfo: &knowledgebase_service.EmbeddingModelInfo{
			ModelId: r.EmbeddingModel.ModelId,
		},
		KnowledgeGraph: &knowledgebase_service.KnowledgeGraph{
			Switch:     r.KnowledgeGraph.Switch,
			LlmModelId: llmModelId,
			SchemaUrl:  schemaUrl,
		},
	})
	if err != nil {
		return nil, err
	}
	return &response.CreateKnowledgeResp{KnowledgeId: resp.KnowledgeId}, nil
}

// UpdateKnowledge Update knowledge base
func UpdateKnowledge(ctx *gin.Context, userId, orgId string, r *request.UpdateKnowledgeReq) error {
	_, err := knowledgeBase.UpdateKnowledge(ctx.Request.Context(), &knowledgebase_service.UpdateKnowledgeReq{
		KnowledgeId: r.KnowledgeId,
		Name:        r.Name,
		Description: r.Description,
		UserId:      userId,
		OrgId:       orgId,
	})
	return err
}

// DeleteKnowledge Delete knowledge base
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

// KnowledgeHit knowledge base hit
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
	})
	if err != nil {
		return nil, err
	}
	return buildKnowledgeHitResp(resp), nil
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

// GetKnowledgeMetaValueList Gets the document metadata list
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

// UpdateKnowledgeMetaValue updates the knowledge base metadata value
func UpdateKnowledgeMetaValue(ctx *gin.Context, userId, orgId string, r *request.UpdateMetaValueReq) error {
	_, err := knowledgeBase.UpdateKnowledgeMetaValue(ctx.Request.Context(), &knowledgebase_service.UpdateKnowledgeMetaValueReq{
		UserId:          userId,
		OrgId:           orgId,
		ApplyToSelected: r.ApplyToSelected,
		DocIdList:       r.DocIdList,
		MetaList:        buildKnowledgeMetaValueReqList(r.MetaValueList),
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

// GetKnowledgeGraph Query knowledge graph details
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

func buildUserKnowledgeList(knowledgeList *response.KnowledgeListResp) map[string][]*request.RagKnowledgeInfo {
	retMap := make(map[string][]*request.RagKnowledgeInfo)
	for _, knowledge := range knowledgeList.KnowledgeList {
		knowledgeInfos, exist := retMap[knowledge.CreateUserId]
		if !exist {
			knowledgeInfos = make([]*request.RagKnowledgeInfo, 0)
		}
		knowledgeInfos = append(knowledgeInfos, &request.RagKnowledgeInfo{
			KnowledgeId:   knowledge.KnowledgeId,
			KnowledgeName: knowledge.RagName,
		})
		retMap[knowledge.CreateUserId] = knowledgeInfos
	}
	return retMap
}

// selectKnowledgeListByIdList queries the knowledge base list, mainly querying all user knowledge bases based on userId
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

// buildKnowledgeMetaList constructs a knowledge base metadata list
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

// buildKnowledgeListReq construct hit test - knowledge base list parameter
func buildKnowledgeListReq(r *request.KnowledgeHitReq) []*knowledgebase_service.KnowledgeParams {
	var knowledgeList []*knowledgebase_service.KnowledgeParams
	for _, k := range r.KnowledgeList {
		knowledgeList = append(knowledgeList, &knowledgebase_service.KnowledgeParams{
			KnowledgeId: k.ID,
			MetaDataFilterParams: &knowledgebase_service.MetaDataFilterParams{
				FilterEnable:     k.MetaDataFilterParams.FilterEnable,
				FilterLogicType:  k.MetaDataFilterParams.FilterLogicType,
				MetaFilterParams: buildMetaFilterParams(k.MetaDataFilterParams.MetaFilterParams),
			},
		})
	}
	return knowledgeList
}

// buildKnowledgeInfoList constructs the knowledge base list results
func buildKnowledgeInfoList(ctx *gin.Context, knowledgeListResp *knowledgebase_service.KnowledgeSelectListResp) *response.KnowledgeListResp {
	if knowledgeListResp == nil || len(knowledgeListResp.KnowledgeList) == 0 {
		return &response.KnowledgeListResp{}
	}
	orgMap := buildOtherOrgInfoMap(ctx, knowledgeListResp)

	var list []*response.KnowledgeInfo
	for _, knowledge := range knowledgeListResp.KnowledgeList {
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
			Share:            share, //Sharing is only possible if the number is greater than 1, because one of the permission records is the permission of the record creator.
			RagName:          knowledge.RagName,
			GraphSwitch:      knowledge.GraphSwitch,
		})
	}
	return &response.KnowledgeListResp{KnowledgeList: list}
}

//nolint:staticcheck
func buildShareOrgName(share bool, orgName string) string {
	if share {
		if strings.Contains(orgName, "---") {
			// "--- System ---" => "System"
			return strings.TrimSpace(strings.Trim(orgName, "---"))
		}
		return orgName
	}
	return ""
}

// The buildOtherOrgInfoMap structure deletes the organization information of the current organization.
func buildOtherOrgInfoMap(ctx *gin.Context, knowledgeListResp *knowledgebase_service.KnowledgeSelectListResp) map[string]string {
	var shareOrgIdList []string
	for _, knowledge := range knowledgeListResp.KnowledgeList {
		if knowledge.ShareCount > 1 {
			shareOrgIdList = append(shareOrgIdList, knowledge.CreateOrgId)
		}
	}
	var dataMap = make(map[string]string)
	if len(shareOrgIdList) > 0 {
		orgInfoList, err := iam.GetOrgByOrgIDs(ctx, &iam_service.GetOrgByOrgIDsReq{
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

// buildTagList constructs a knowledge base tag list
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

// buildKnowledgeHitResp constructs knowledge base hit return
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

func buildMetaFilterParams(meta []*request.MetaFilterParams) []*knowledgebase_service.MetaFilterParams {
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

// requestRagSearchKnowledgeBase request rag
func requestRagSearchKnowledgeBase(ctx context.Context, req *request.RagSearchKnowledgeBaseReq) ([]byte, int) {
	url := config.Cfg().RagKnowledgeConfig.Endpoint + config.Cfg().RagKnowledgeConfig.SearchKnowledgeBaseUri
	paramsByte, err := json.Marshal(req)
	if err != nil {
		return response.CommonRagKnowledgeError(err)
	}
	result, err := knowHttp.PostJsonOriResp(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       paramsByte,
		MonitorKey: "rag_search_knowledge_base",
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
	// Catch panic and log (do not rethrow to avoid crash)
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
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: 调用下游服务异常: %s", resp.Status)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("error: 响应体关闭异常: %v", err)
		}
	}(resp.Body) // Make sure the response body is closed

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
