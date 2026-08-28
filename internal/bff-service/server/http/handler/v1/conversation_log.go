package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetAssistantConversationLogList
//
//	@Tags			agent
//	@Summary		获取会话日志列表
//	@Description	获取会话日志列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GetConversationLogListRequest	true	"会话日志列表"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.ConversationLogInfo}}
//	@Router			/assistant/draft/conversation/log/list [post]
func GetAssistantConversationLogList(ctx *gin.Context) {
	var req request.GetConversationLogListRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetConversationLogList(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetAssistantConversationLogDetail
//
//	@Tags			agent
//	@Summary		获取会话日志详情
//	@Description	获取会话日志详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GetConversationLogDetailRequest	true	"会话日志详情"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.ConversationDetailInfo}}
//	@Router			/assistant/draft/conversation/log/detail [post]
func GetAssistantConversationLogDetail(ctx *gin.Context) {
	var req request.GetConversationLogDetailRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetAssistantConversationLogDetail(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetAssistantConversationLogUserSelect
//
//	@Tags			agent
//	@Summary		获取会话日志使用者列表
//	@Description	获取会话日志使用者列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GetConversationLogUserSelectRequest	true	"会话日志使用者列表"
//	@Success		200		{object}	response.Response{data=response.Users}
//	@Router			/assistant/draft/conversation/log/user/select [post]
func GetAssistantConversationLogUserSelect(ctx *gin.Context) {
	var req request.GetConversationLogUserSelectRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetConversationLogUserSelect(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetRagConversationLogList
//
//	@Tags			rag
//	@Summary		获取会话日志列表
//	@Description	获取会话日志列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GetConversationLogListRequest	true	"会话日志列表"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.ConversationLogInfo}}
//	@Router			/rag/draft/conversation/log/list [post]
func GetRagConversationLogList(ctx *gin.Context) {
	var req request.GetConversationLogListRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetConversationLogList(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetRagConversationLogDetail
//
//	@Tags			rag
//	@Summary		获取会话日志详情
//	@Description	获取会话日志详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GetConversationLogDetailRequest	true	"会话日志详情"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.RagConversationDetailInfo}}
//	@Router			/rag/draft/conversation/log/detail [post]
func GetRagConversationLogDetail(ctx *gin.Context) {
	//var req request.GetConversationLogDetailRequest
	//if !gin_util.Bind(ctx, &req) {
	//	return
	//}
	//resp, err := service.GetAssistantConversationLogDetail(ctx, getUserID(ctx), getOrgID(ctx), req)
	//gin_util.Response(ctx, resp, err)
}

// GetRagConversationLogUserSelect
//
//	@Tags			rag
//	@Summary		获取会话日志使用者列表
//	@Description	获取会话日志使用者列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GetConversationLogUserSelectRequest	true	"会话日志使用者列表"
//	@Success		200		{object}	response.Response{data=response.Users}
//	@Router			/rag/draft/conversation/log/user/select [post]
func GetRagConversationLogUserSelect(ctx *gin.Context) {
	var req request.GetConversationLogUserSelectRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetConversationLogUserSelect(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// ExportConversationLog
//
//	@Tags			conversation.log
//	@Summary		对话日志导出
//	@Description	对话日志导出（异步，立即返回）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConversationLogExportReq	true	"对话日志导出请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/conversation/log/export [post]
func ExportConversationLog(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationLogExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ExportConversationLog(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetConversationLogExportRecordList
//
//	@Tags			conversation.log
//	@Summary		获取对话日志导出记录列表
//	@Description	获取对话日志导出记录列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.ConversationLogExportRecordListReq	true	"获取对话日志导出记录列表请求参数"
//	@Success		200		{object}	response.Response{data=response.PageResult{list=[]response.ConversationLogExportRecordResp}}
//	@Router			/conversation/log/export/record/list [get]
func GetConversationLogExportRecordList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ConversationLogExportRecordListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetConversationLogExportRecordList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// DeleteConversationLogExportRecord
//
//	@Tags			conversation.log
//	@Summary		删除对话日志导出记录
//	@Description	删除对话日志导出记录
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteConversationLogExportRecordReq	true	"删除对话日志导出记录请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/conversation/log/export/record [delete]
func DeleteConversationLogExportRecord(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteConversationLogExportRecordReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteConversationLogExportRecord(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
