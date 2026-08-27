package openapi

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// ListAgents
//
//	@Tags			openapi
//	@Summary		智能体列表OpenAPI
//	@Description	获取当前 API Key 所属用户下的全部智能体（含草稿与已发布），可按名称模糊筛选。返回的 uuid 字段可用于其他智能体接口。
//	@Produce		json
//	@Param			name	query		string	false	"按名称模糊筛选（可选）"
//	@Success		200		{object}	response.Response{data=response.OpenAPIAgentListResponse}
//	@Router			/agent/list [get]
func ListAgents(ctx *gin.Context) {
	var req request.OpenAPIAgentListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	resp, err := service.GetAgentListForOpenAPI(ctx, userID, orgID, req.Name)
	gin_util.Response(ctx, resp, err)
}

// DeleteAgent
//
//	@Tags			openapi
//	@Summary		删除智能体OpenAPI
//	@Description	删除指定智能体，同时删除其已发布版本。删除后不可恢复，请谨慎操作。
//	@Produce		json
//	@Param			uuid	query		string	true	"智能体UUID"
//	@Success		200		{object}	response.Response
//	@Router			/agent [delete]
func DeleteAgent(ctx *gin.Context) {
	var req request.OpenAPIAgentDeleteRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	err := service.DeleteAppSpaceApp(ctx, userID, orgID, req.UUID, constant.AppTypeAgent)
	gin_util.Response(ctx, nil, err)
}

// UpdateAgent
//
//	@Tags			openapi
//	@Summary		更新智能体基本信息OpenAPI
//	@Description	修改智能体的名称、描述、图标等基本信息（不涉及开场白、大模型等高级配置，后者请使用 /agent/config）。通过 assistantUuid 定位智能体。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIUpdateAgentRequest	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/agent [put]
func UpdateAgent(ctx *gin.Context) {
	var req request.OpenAPIUpdateAgentRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	_, err := service.AssistantUpdate(ctx, userID, orgID, request.AssistantUpdateReq{
		AssistantId:    req.AssistantUUID,
		AppBriefConfig: req.AppBriefConfig,
	})
	gin_util.Response(ctx, nil, err)
}

// --- 已发布智能体对话管理 ---

// ListAgentConversations
//
//	@Tags			openapi
//	@Summary		智能体对话列表OpenAPI
//	@Description	获取指定智能体的已发布对话列表，按创建时间降序排列。
//	@Produce		json
//	@Param			uuid		query		string	true	"智能体UUID"
//	@Param			pageNo		query		int		true	"页码，从 1 开始"
//	@Param			pageSize	query		int		true	"每页条数，从 1 开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.ConversationInfo}}
//	@Router			/agent/conversation/list [get]
func ListAgentConversations(ctx *gin.Context) {
	var req request.OpenAPIAgentConversationListRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	resp, err := service.GetConversationList(ctx, userID, orgID, request.ConversationGetListRequest{
		AssistantId: req.UUID,
		PageNo:      req.PageNo,
		PageSize:    req.PageSize,
	})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp.PageNo, resp.PageSize = req.PageNo, req.PageSize
	gin_util.Response(ctx, resp, nil)
}

// GetAgentConversationDetail
//
//	@Tags			openapi
//	@Summary		智能体对话历史消息OpenAPI
//	@Description	获取指定对话的历史消息列表（问答明细），分页返回。
//	@Produce		json
//	@Param			conversation_id	query		string	true	"对话ID（由创建对话接口返回）"
//	@Param			pageNo			query		int		true	"页码，从 1 开始"
//	@Param			pageSize		query		int		true	"每页条数，从 1 开始"
//	@Success		200				{object}	response.Response{data=response.PageResult{list=[]response.ConversationDetailInfo}}
//	@Router			/agent/conversation/detail [get]
func GetAgentConversationDetail(ctx *gin.Context) {
	var req request.OpenAPIAgentConversationDetailRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	resp, err := service.GetConversationDetailList(ctx, userID, orgID, request.ConversationGetDetailListRequest{
		ConversationId: req.ConversationID,
		PageNo:         req.PageNo,
		PageSize:       req.PageSize,
	})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp.PageNo, resp.PageSize = req.PageNo, req.PageSize
	gin_util.Response(ctx, resp, nil)
}

// DeleteAgentConversation
//
//	@Tags			openapi
//	@Summary		删除智能体对话OpenAPI
//	@Description	删除整个对话（含 conversation 主体及所有历史消息）。如需按条删除消息或仅清空消息保留对话 ID，请使用 /agent/conversation/clear。
//	@Produce		json
//	@Param			conversation_id	query		string	true	"对话ID"
//	@Success		200				{object}	response.Response
//	@Router			/agent/conversation [delete]
func DeleteAgentConversation(ctx *gin.Context) {
	var req request.OpenAPIAgentConversationDeleteRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	_, err := service.ConversationDelete(ctx, userID, orgID, request.ConversationIdRequest{
		ConversationId: req.ConversationID,
	})
	gin_util.Response(ctx, nil, err)
}

// ClearAgentConversation
//
//	@Tags			openapi
//	@Summary		清空/按条删除智能体对话消息OpenAPI
//	@Description	对指定对话的消息（ES 数据）进行清理：传入 detail_id 则仅删除该条消息；不传则清空整个对话的全部消息。两种情况下 conversation_id 均保留，可继续发起新的问答。
//	@Produce		json
//	@Param			conversation_id	query		string	true	"对话ID"
//	@Param			detail_id		query		string	false	"消息ID（不传则清空整个对话的全部消息）"
//	@Success		200				{object}	response.Response
//	@Router			/agent/conversation/clear [delete]
func ClearAgentConversation(ctx *gin.Context) {
	var req request.OpenAPIAgentConversationClearRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	_, err := service.ClearPublishedConversationES(ctx, userID, orgID, request.ConversationIdRequest{
		ConversationId: req.ConversationID,
		DetailId:       req.DetailID,
	})
	gin_util.Response(ctx, nil, err)
}

// --- 草稿态智能体对话管理 ---

// GetAgentDraftConversationDetail
//
//	@Tags			openapi
//	@Summary		草稿态智能体对话历史消息OpenAPI
//	@Description	获取指定草稿智能体的对话历史消息。草稿态每个智能体只维护一条会话，通过 uuid 定位后分页返回消息列表。
//	@Produce		json
//	@Param			uuid		query		string	true	"智能体UUID"
//	@Param			pageNo		query		int		true	"页码，从 1 开始"
//	@Param			pageSize	query		int		true	"每页条数，从 1 开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.ConversationDetailInfo}}
//	@Router			/agent/conversation/draft/detail [get]
func GetAgentDraftConversationDetail(ctx *gin.Context) {
	var req request.OpenAPIAgentDraftConversationDetailRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	convResp, err := service.GetDraftConversationIdByAssistantID(ctx, userID, orgID, request.ConversationGetListRequest{
		AssistantId: req.UUID,
	})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	if convResp == nil {
		gin_util.Response(ctx, response.PageResult{
			List:     []response.ConversationDetailInfo{},
			PageNo:   req.PageNo,
			PageSize: req.PageSize,
		}, nil)
		return
	}

	resp, err := service.GetConversationDetailList(ctx, userID, orgID, request.ConversationGetDetailListRequest{
		ConversationId: convResp.ConversationId,
		PageNo:         req.PageNo,
		PageSize:       req.PageSize,
	})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	resp.PageNo, resp.PageSize = req.PageNo, req.PageSize
	gin_util.Response(ctx, resp, nil)
}

// DeleteAgentDraftConversation
//
//	@Tags			openapi
//	@Summary		删除草稿态智能体对话历史OpenAPI
//	@Description	删除草稿智能体对话历史。传入 detail_id 则只删除该条消息，不传则清空全部历史（会话 ID 保留）。
//	@Produce		json
//	@Param			uuid		query		string	true	"智能体UUID"
//	@Param			detail_id	query		string	false	"消息ID（不传则清空全部）"
//	@Success		200			{object}	response.Response
//	@Router			/agent/conversation/draft [delete]
func DeleteAgentDraftConversation(ctx *gin.Context) {
	var req request.OpenAPIAgentDraftConversationDeleteRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	_, err := service.DraftConversationDeleteByAssistantID(ctx, userID, orgID, request.ConversationDeleteRequest{
		AssistantId: req.UUID,
		DetailId:    req.DetailID,
	})
	gin_util.Response(ctx, nil, err)
}
