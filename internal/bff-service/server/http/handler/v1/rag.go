package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// ChatDraftRag
//
//	@Tags		rag
//	@Summary	私域 草稿RAG 问答
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.ChatRagRequest	true	"RAG问答请求参数"
//	@Success	200		{object}	response.Response
//	@Router		/rag/chat/draft [post]
func ChatDraftRag(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.ChatRagRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// 草稿态每个知识问答仅维护一条会话，调用方不传则 get-or-create，确保历史能落库
	if req.ConversationID == "" {
		conversationId, err := service.GetOrCreateDraftRagConversation(ctx, userId, orgId, req.RagID, req.Question)
		if err != nil {
			gin_util.Response(ctx, nil, err)
			return
		}
		req.ConversationID = conversationId
	}
	//启用链接保持
	req.SseHold = true
	if err := service.ChatRagStream(ctx, userId, orgId, clientId, req, false, constant.BizSourceWeb); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// RagUpload
//
//	@Tags		rag
//	@Summary	文件直接上传到rag
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.RagUploadParams	true	"RAG上传文档请求参数"
//	@Success	200		{object}	response.Response{data=response.RagUploadResponse}
//	@Router		/rag/upload [post]
func RagUpload(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagUploadParams
	if !gin_util.BindForm(ctx, &req) {
		return
	}
	upload, err := service.RagUpload(ctx, userId, orgId, req)
	gin_util.Response(ctx, upload, err)
}

// ChatPublishedRag
//
//	@Tags		rag
//	@Summary	已发布 RAG 问答
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.ChatRagRequest	true	"RAG问答请求参数"
//	@Success	200		{object}	response.Response
//	@Router		/rag/chat [post]
func ChatPublishedRag(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.ChatRagRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	//启用链接保持
	req.SseHold = true
	if err := service.ChatRagStream(ctx, userId, orgId, clientId, req, true, constant.BizSourceWeb); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// GetRagPendingConversation
//
//	@Tags			rag
//	@Summary		获取知识问答运行中会话
//	@Description	获取知识问答运行中会话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagPendingConversationReq	true	"获取知识问答运行中会话请求参数"
//	@Success		200		{object}	response.Response{data=response.PendingConversationResp}
//	@Router			/rag/pending/conversation [post]
func GetRagPendingConversation(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.RagPendingConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	conversation, err := service.GetRagPendingConversation(ctx, userId, orgId, clientId, req)
	gin_util.Response(ctx, conversation, err)
}

// ChatRagStreamConnect
//
//	@Tags			rag
//	@Summary		知识问答流式问答断开后重连
//	@Description	知识问答流式问答断开后重连
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagStreamConnectReq	true	"知识问答流式问答重连参数"
//	@Success		200		{object}	response.Response
//	@Router			/rag/stream/connect [post]
func ChatRagStreamConnect(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.RagStreamConnectReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.ChatRagStreamConnect(ctx, userId, orgId, clientId, req); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// ChatRagStreamCancel
//
//	@Tags			rag
//	@Summary		知识问答流式问答手动停止
//	@Description	知识问答流式问答手动停止
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagStreamCancelReq	true	"知识问答流式问答手动停止参数"
//	@Success		200		{object}	response.Response
//	@Router			/rag/stream/cancel [post]
func ChatRagStreamCancel(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.RagStreamCancelReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.ChatRagStreamCancel(ctx, userId, orgId, clientId, req); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, nil, nil)
}

// CreateRag
//
//	@Tags		rag
//	@Summary	创建RAG
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.RagCreateReq	true	"创建RAG的请求参数"
//	@Success	200		{object}	response.Response{data=request.RagReq}
//	@Router		/appspace/rag [post]
func CreateRag(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagCreateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateRag(ctx, userId, orgId, req.AppBriefConfig)
	gin_util.Response(ctx, resp, err)
}

// UpdateRag
//
//	@Tags		rag
//	@Summary	更新RAG基本信息
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.RagUpdateReq	true	"更新RAG基本信息的请求参数"
//	@Success	200		{object}	response.Response
//	@Router		/appspace/rag [put]
func UpdateRag(ctx *gin.Context) {
	var req request.RagUpdateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	err := service.UpdateRag(ctx, req, userId, orgId)
	gin_util.Response(ctx, nil, err)
}

// UpdateRagConfig
//
//	@Tags		rag
//	@Summary	更新RAG配置信息
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.RagConfig	true	"更新RAG配置信息的请求参数"
//	@Success	200		{object}	response.Response
//	@Router		/appspace/rag/config [put]
func UpdateRagConfig(ctx *gin.Context) {
	var req request.RagConfig
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateRagConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// DeleteRag
//
//	@Tags		rag
//	@Summary	删除RAG
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.RagReq	true	"删除RAG的请求参数"
//	@Success	200		{object}	response.Response
//	@Router		/appspace/rag [delete]
func DeleteRag(ctx *gin.Context) {
	var req request.RagReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteRag(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// GetDraftRag
//
//	@Tags		rag
//	@Summary	获取草稿RAG信息
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	query		request.RagReq	true	"获取RAG信息的请求参数"
//	@Success	200		{object}	response.Response{data=response.RagInfo}
//	@Router		/appspace/rag/draft [get]
func GetDraftRag(ctx *gin.Context) {
	var req request.RagReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetRag(ctx, getUserID(ctx), getOrgID(ctx), req, false)
	gin_util.Response(ctx, resp, err)
}

// GetPublishedRag
//
//	@Tags		rag
//	@Summary	获取已发布RAG信息
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	query		request.RagReq	true	"获取RAG信息的请求参数"
//	@Success	200		{object}	response.Response{data=response.RagInfo}
//	@Router		/appspace/rag [get]
func GetPublishedRag(ctx *gin.Context) {
	var req request.RagReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetRag(ctx, getUserID(ctx), getOrgID(ctx), req, true)
	gin_util.Response(ctx, resp, err)
}

// CopyRag
//
//	@Tags		rag
//	@Summary	复制RAG
//	@Description
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.RagReq	true	"复制RAG的请求参数"
//	@Success	200		{object}	response.Response{data=request.RagReq}
//	@Router		/appspace/rag/copy [post]
func CopyRag(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CopyRag(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// RagConversationCreate
//
//	@Tags			rag
//	@Summary		创建知识问答会话
//	@Description	创建知识问答会话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagConversationCreateReq	true	"知识问答会话创建参数"
//	@Success		200		{object}	response.Response{data=response.RagConversationCreateResp}
//	@Router			/rag/conversation [post]
func RagConversationCreate(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagConversationCreateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.RagConversationCreate(ctx, userId, orgId, req, constant.ConversationTypePublished)
	gin_util.Response(ctx, resp, err)
}

// RagConversationDelete
//
//	@Tags			rag
//	@Summary		删除知识问答会话
//	@Description	删除知识问答会话，传 detailId 则只删单条对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagConversationIDReq	true	"知识问答会话id"
//	@Success		200		{object}	response.Response
//	@Router			/rag/conversation [delete]
func RagConversationDelete(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagConversationIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.RagConversationDelete(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// RagConversationClear
//
//	@Tags			rag
//	@Summary		清空知识问答会话记录
//	@Description	清空会话下的对话记录但保留会话，传 detailId 则只删单条
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagConversationIDReq	true	"知识问答会话id"
//	@Success		200		{object}	response.Response
//	@Router			/rag/conversation/clear [delete]
func RagConversationClear(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagConversationIDReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.RagConversationClear(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// RagConversationList
//
//	@Tags			rag
//	@Summary		知识问答会话列表
//	@Description	知识问答会话列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.RagConversationListReq	true	"知识问答会话列表参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.RagConversationInfo}}
//	@Router			/rag/conversation/list [get]
func RagConversationList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagConversationListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.RagConversationList(ctx, userId, orgId, req, constant.ConversationTypePublished)
	gin_util.Response(ctx, resp, err)
}

// RagConversationDetailList
//
//	@Tags			rag
//	@Summary		知识问答对话详情历史列表
//	@Description	知识问答对话详情历史列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.RagConversationDetailListReq	true	"知识问答对话详情列表参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.RagConversationDetailInfo}}
//	@Router			/rag/conversation/detail [get]
func RagConversationDetailList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagConversationDetailListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.RagConversationDetailList(ctx, userId, orgId, req)
	gin_util.Response(ctx, resp, err)
}

// RagConversationDraftDetailList
//
//	@Tags			rag
//	@Summary		草稿知识问答对话详情历史列表
//	@Description	草稿态每个知识问答仅一条会话，按 ragId 取
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.RagConversationDraftReq	true	"知识问答id"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.RagConversationDetailInfo}}
//	@Router			/rag/conversation/draft/detail [get]
func RagConversationDraftDetailList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagConversationDraftReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	conversationId, err := service.GetDraftRagConversationId(ctx, userId, orgId, req.RagID)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 草稿尚未发起过对话：返回空列表而不是报错
	if conversationId == "" {
		gin_util.Response(ctx, response.PageResult{List: []response.RagConversationDetailInfo{}}, nil)
		return
	}
	// 草稿只有一条会话、条数有限，不传分页时给默认值，避免 size=0 查出空列表
	if req.PageNo <= 0 {
		req.PageNo = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	resp, err := service.RagConversationDetailList(ctx, userId, orgId, request.RagConversationDetailListReq{
		RagID:          req.RagID,
		ConversationID: conversationId,
		PageNo:         req.PageNo,
		PageSize:       req.PageSize,
	})
	gin_util.Response(ctx, resp, err)
}

// RagConversationDraftDelete
//
//	@Tags			rag
//	@Summary		删除草稿知识问答对话
//	@Description	不传 detailId 删除整条草稿会话，传了则只删单条对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagConversationDraftReq	true	"知识问答id"
//	@Success		200		{object}	response.Response
//	@Router			/rag/conversation/draft [delete]
func RagConversationDraftDelete(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.RagConversationDraftReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	conversationId, err := service.GetDraftRagConversationId(ctx, userId, orgId, req.RagID)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	// 草稿会话尚未创建：删除请求幂等成功
	if conversationId == "" {
		gin_util.Response(ctx, nil, nil)
		return
	}
	resp, err := service.RagConversationDelete(ctx, userId, orgId, request.RagConversationIDReq{
		RagID:          req.RagID,
		ConversationID: conversationId,
		DetailID:       req.DetailID,
	})
	gin_util.Response(ctx, resp, err)
}
