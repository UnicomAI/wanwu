package openapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// 知识问答 OpenAPI —— 会话管理与草稿态问答。

// --- 已发布会话管理 ---

// CreateRagConversation
//
//	@Tags			openapi
//	@Summary		知识问答创建会话OpenAPI
//	@Description	创建一条会话，返回的 conversation_id 用于 /rag/chat 携带上下文。不传 conversation_id 的问答不落历史。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIRagCreateConversationRequest	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.OpenAPIRagCreateConversationResponse}
//	@Router			/rag/conversation [post]
func CreateRagConversation(ctx *gin.Context) {
	var req request.OpenAPIRagCreateConversationRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	// OpenAPI 创建的会话与 web 已发布会话同类型
	resp, err := service.RagConversationCreate(ctx, userID, orgID, request.RagConversationCreateReq{
		RagID:  req.UUID,
		Prompt: req.Title,
	}, constant.ConversationTypePublished)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, response.OpenAPIRagCreateConversationResponse{ConversationID: resp.ConversationId}, nil)
}

// ListRagConversations
//
//	@Tags			openapi
//	@Summary		知识问答会话列表OpenAPI
//	@Description	获取指定知识问答下的会话列表，按创建时间降序排列。
//	@Produce		json
//	@Param			uuid		query		string	true	"知识问答UUID"
//	@Param			pageNo		query		int		true	"页码，从 1 开始"
//	@Param			pageSize	query		int		true	"每页条数，从 1 开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.RagConversationInfo}}
//	@Router			/rag/conversation/list [get]
func ListRagConversations(ctx *gin.Context) {
	var req request.OpenAPIRagConversationListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	resp, err := service.RagConversationList(ctx, userID, orgID, request.RagConversationListReq{
		RagID:      req.UUID,
		SearchText: req.SearchText,
		PageNo:     req.PageNo,
		PageSize:   req.PageSize,
	}, constant.ConversationTypePublished)
	gin_util.Response(ctx, resp, err)
}

// GetRagConversationDetail
//
//	@Tags			openapi
//	@Summary		知识问答会话历史消息OpenAPI
//	@Description	获取指定会话的历史问答明细，分页返回，含检索命中列表与耗时。
//	@Produce		json
//	@Param			uuid			query		string	true	"知识问答UUID"
//	@Param			conversation_id	query		string	true	"会话ID（由创建会话接口返回）"
//	@Param			pageNo			query		int		true	"页码，从 1 开始"
//	@Param			pageSize		query		int		true	"每页条数，从 1 开始"
//	@Success		200				{object}	response.Response{data=response.PageResult{list=[]response.RagConversationDetailInfo}}
//	@Router			/rag/conversation/detail [get]
func GetRagConversationDetail(ctx *gin.Context) {
	var req request.OpenAPIRagConversationDetailRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	resp, err := service.RagConversationDetailList(ctx, userID, orgID, request.RagConversationDetailListReq{
		RagID:          req.UUID,
		ConversationID: req.ConversationID,
		PageNo:         req.PageNo,
		PageSize:       req.PageSize,
	})
	gin_util.Response(ctx, resp, err)
}

// DeleteRagConversation
//
//	@Tags			openapi
//	@Summary		删除知识问答会话OpenAPI
//	@Description	删除整条会话（含全部历史消息）。只想清空消息、保留会话ID请用 /rag/conversation/clear。
//	@Produce		json
//	@Param			uuid			query		string	true	"知识问答UUID"
//	@Param			conversation_id	query		string	true	"会话ID"
//	@Success		200				{object}	response.Response
//	@Router			/rag/conversation [delete]
func DeleteRagConversation(ctx *gin.Context) {
	var req request.OpenAPIRagConversationDeleteRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	_, err := service.RagConversationDelete(ctx, userID, orgID, request.RagConversationIDReq{
		RagID:          req.UUID,
		ConversationID: req.ConversationID,
	})
	gin_util.Response(ctx, nil, err)
}

// ClearRagConversation
//
//	@Tags			openapi
//	@Summary		清空/按条删除知识问答会话消息OpenAPI
//	@Description	传 detail_id 只删该条问答，不传则清空整条会话的全部消息。两种情况会话ID都保留，可继续提问。
//	@Produce		json
//	@Param			uuid			query		string	true	"知识问答UUID"
//	@Param			conversation_id	query		string	true	"会话ID"
//	@Param			detail_id		query		string	false	"问答明细ID（不传则清空全部）"
//	@Success		200				{object}	response.Response
//	@Router			/rag/conversation/clear [delete]
func ClearRagConversation(ctx *gin.Context) {
	var req request.OpenAPIRagConversationClearRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	_, err := service.RagConversationClear(ctx, userID, orgID, request.RagConversationIDReq{
		RagID:          req.UUID,
		ConversationID: req.ConversationID,
		DetailID:       req.DetailID,
	})
	gin_util.Response(ctx, nil, err)
}

// --- 草稿态会话管理 ---

// GetRagDraftConversationDetail
//
//	@Tags			openapi
//	@Summary		草稿知识问答会话历史消息OpenAPI
//	@Description	草稿态每个知识问答只维护一条会话，按 uuid 定位后分页返回问答明细。尚未发起过对话时返回空列表。
//	@Produce		json
//	@Param			uuid		query		string	true	"知识问答UUID"
//	@Param			pageNo		query		int		true	"页码，从 1 开始"
//	@Param			pageSize	query		int		true	"每页条数，从 1 开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.RagConversationDetailInfo}}
//	@Router			/rag/conversation/draft/detail [get]
func GetRagDraftConversationDetail(ctx *gin.Context) {
	var req request.OpenAPIRagDraftConversationDetailRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	conversationID, err := service.GetDraftRagConversationId(ctx, userID, orgID, req.UUID)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if conversationID == "" {
		gin_util.Response(ctx, response.PageResult{List: []response.RagConversationDetailInfo{}}, nil)
		return
	}
	resp, err := service.RagConversationDetailList(ctx, userID, orgID, request.RagConversationDetailListReq{
		RagID:          req.UUID,
		ConversationID: conversationID,
		PageNo:         req.PageNo,
		PageSize:       req.PageSize,
	})
	gin_util.Response(ctx, resp, err)
}

// DeleteRagDraftConversation
//
//	@Tags			openapi
//	@Summary		删除草稿知识问答会话消息OpenAPI
//	@Description	传 detail_id 只删该条问答，不传则删除整条草稿会话。草稿会话尚未创建时直接返回成功。
//	@Produce		json
//	@Param			uuid		query		string	true	"知识问答UUID"
//	@Param			detail_id	query		string	false	"问答明细ID（不传则删除整条草稿会话）"
//	@Success		200			{object}	response.Response
//	@Router			/rag/conversation/draft [delete]
func DeleteRagDraftConversation(ctx *gin.Context) {
	var req request.OpenAPIRagDraftConversationDeleteRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	conversationID, err := service.GetDraftRagConversationId(ctx, userID, orgID, req.UUID)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 草稿会话尚未创建：删除幂等成功
	if conversationID == "" {
		gin_util.Response(ctx, nil, nil)
		return
	}
	_, err = service.RagConversationDelete(ctx, userID, orgID, request.RagConversationIDReq{
		RagID:          req.UUID,
		ConversationID: conversationID,
		DetailID:       req.DetailID,
	})
	gin_util.Response(ctx, nil, err)
}

// --- 草稿态问答 ---

// DraftChatRag
//
//	@Tags			openapi
//	@Summary		知识问答草稿态问答OpenAPI
//	@Description	基于草稿配置问答，不要求已发布，计入统计。响应格式与 /rag/chat 一致（stream=true 为 SSE，否则为一次性 JSON）。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIRagDraftChatRequest	true	"请求参数"
//	@Success		200		{object}	response.OpenAPIRagChatResponse
//	@Success		400		{object}	response.Response
//	@Router			/rag/chat/draft [post]
func DraftChatRag(ctx *gin.Context) {
	var req request.OpenAPIRagDraftChatRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID, orgID := getUserID(ctx), getOrgID(ctx)
	// 带文件提问的前置校验
	if len(req.FileInfo) > 0 {
		if err := service.CheckRagChatFileReady(ctx, userID, orgID, req.UUID, false); err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
	}
	// 归属与配置校验在 service.CallRagChatStream 的草稿分支里做
	// 草稿态每个知识问答仅一条会话，由服务端 get-or-create
	conversationID, err := service.GetOrCreateDraftRagConversation(ctx, userID, orgID, req.UUID, req.Query)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	chatReq := request.ChatRagRequest{
		RagID:          req.UUID,
		Question:       req.Query,
		FileInfo:       req.FileInfo,
		ConversationID: conversationID,
	}
	chatRagWithFormat(ctx, userID, orgID, chatReq, req.Stream, false, service.MarshalStatisticBody(req))
}

// chatRagWithFormat 流式走 legacy SSE 透传，非流式收敛成一次性 JSON
func chatRagWithFormat(ctx *gin.Context, userID, orgID string, chatReq request.ChatRagRequest, stream, needLatestPublished bool, statisticBody string) {
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	if stream {
		if err := service.ChatRagStreamLegacy(ctx, userID, orgID, chatReq, needLatestPublished, constant.BizSourceOpenAPI); err != nil {
			gin_util.Response(ctx, nil, err)
		}
		return
	}
	startTime := time.Now()
	chatCh, _, err := service.CallRagChatStream(ctx, ctx.Request.Context(), userID, orgID, chatReq, needLatestPublished)
	if err != nil {
		statusCode, failureReason := service.GrpcErrorToHTTPStatus(err)
		go func() {
			defer util.PrintPanicStack()
			service.RecordAppStatistic(detachedCtx, userID, orgID, chatReq.RagID, constant.AppTypeRag, "",
				statusCode, failureReason, false, 0, 0, constant.BizSourceOpenAPI, statisticBody, "", chatReq.Question, "")
		}()
		gin_util.Response(ctx, nil, err)
		return
	}
	var output strings.Builder
	resp := &response.OpenAPIRagChatResponse{}
	for chat := range chatCh {
		if !strings.HasPrefix(chat, "data:") || strings.HasPrefix(chat, strings.TrimSpace(sse_util.DONE_MSG)) {
			continue
		}
		curr := &response.OpenAPIRagChatResponse{}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(chat, "data:")), curr); err != nil {
			log.Errorf("[RAG] %v user %v org %v unmarshal err: %v", chatReq.RagID, userID, orgID, err)
			continue
		}
		resp = curr
		output.WriteString(curr.Data.Output)
	}
	resp.Data.Output = output.String()
	costs := time.Since(startTime).Milliseconds()
	go func() {
		defer util.PrintPanicStack()
		service.RecordAppStatistic(detachedCtx, userID, orgID, chatReq.RagID, constant.AppTypeRag, "",
			200, "", false, 0, costs, constant.BizSourceOpenAPI, statisticBody, service.MarshalStatisticBody(resp), chatReq.Question, resp.Data.Output)
	}()
	b, _ := json.Marshal(resp)
	status := http.StatusOK
	ctx.Set(gin_util.STATUS, status)
	ctx.Set(gin_util.RESULT, string(b))
	ctx.JSON(status, resp)
}
