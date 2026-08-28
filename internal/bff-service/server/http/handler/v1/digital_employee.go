package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateDigitalEmployeeConversation
//
//	@Tags			digital-employee
//	@Summary		创建数字员工发布会话
//	@Description	创建数字员工发布会话（绑定数字员工，创建后不可切换；模型配置可选，落独立表）。
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateDigitalEmployeeConversationReq	true	"请求参数"
//	@Success		200		{object}	response.Response{data=response.CreateDigitalEmployeeConversationResp}
//	@Router			/digital-employee/conversation [post]
func CreateDigitalEmployeeConversation(ctx *gin.Context) {
	var req request.CreateDigitalEmployeeConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateDigitalEmployeeConversation(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DeleteDigitalEmployeeConversation
//
//	@Tags			digital-employee
//	@Summary		删除数字员工发布会话
//	@Description	删除数字员工发布会话及其历史（DB 行 + 独立 ES 索引级联删除）。
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteDigitalEmployeeConversationReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/digital-employee/conversation [delete]
func DeleteDigitalEmployeeConversation(ctx *gin.Context) {
	var req request.DeleteDigitalEmployeeConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteDigitalEmployeeConversation(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// GetDigitalEmployeeConversationList
//
//	@Tags			digital-employee
//	@Summary		数字员工发布会话列表
//	@Description	获取与数字员工绑定的发布会话列表（按 employeeId 维度、updated_at DESC，最近对话排最前）。
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			employeeId	query		string	true	"数字员工ID"
//	@Param			pageNo		query		int		true	"页码"
//	@Param			pageSize	query		int		true	"每页数量"
//	@Param			searchText	query		string	false	"标题关键词"
//	@Success		200			{object}	response.Response{data=response.ListResult}
//	@Router			/digital-employee/conversation/list [get]
func GetDigitalEmployeeConversationList(ctx *gin.Context) {
	var req request.GetDigitalEmployeeConversationListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDigitalEmployeeConversationList(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetDigitalEmployeeConversationDetail
//
//	@Tags			digital-employee
//	@Summary		数字员工发布会话详情
//	@Description	按会话ID读取发布对话历史消息（独立 ES 索引 digital_employee_chat_history_event）。
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			conversationId	query		string	true	"会话ID"
//	@Success		200				{object}	response.Response{data=response.ListResult{list=[]response.GeneralAgentConversationDetailInfo}}
//	@Router			/digital-employee/conversation/detail [get]
func GetDigitalEmployeeConversationDetail(ctx *gin.Context) {
	var req request.GetDigitalEmployeeConversationDetailReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDigitalEmployeeConversationDetail(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DigitalEmployeeChat
//
//	@Tags			digital-employee
//	@Summary		数字员工发布对话
//	@Description	应用广场数字员工发布对话（wga 模式，固定 DIP Agent，SSE 流式返回；会话须先创建，conversationId 必填）。
//	@Security		JWT
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			data	body		request.DigitalEmployeeChatReq	true	"请求参数"
//	@Success		200		{object}	string							"SSE流式返回"
//	@Router			/digital-employee/chat [post]
func DigitalEmployeeChat(ctx *gin.Context) {
	var req request.DigitalEmployeeChatReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DigitalEmployeeChat(ctx, getUserID(ctx), getOrgID(ctx), getClientID(ctx), req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// GetDigitalEmployeeSquareDetail
//
//	@Tags			digital-employee
//	@Summary		数字员工广场详情
//	@Description	应用广场数字员工详情展示（实时调外部详情 + 发布者信息，决策 D6 不做缓存）。
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			employeeId	query		string	true	"数字员工ID"
//	@Success		200			{object}	response.Response{data=response.DigitalEmployeeSquareDetail}
//	@Router			/digital-employee/detail [get]
func GetDigitalEmployeeSquareDetail(ctx *gin.Context) {
	var req request.GetDigitalEmployeeSquareDetailReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDigitalEmployeeSquareDetail(ctx, getUserID(ctx), getOrgID(ctx), req.EmployeeId)
	gin_util.Response(ctx, resp, err)
}

// GetDigitalEmployeeConversationConfig
//
//	@Tags			digital-employee
//	@Summary		数字员工发布会话配置
//	@Description	读回数字员工发布会话的模型配置（对齐通用智能体 GET /general/agent/conversation/config；模型未配置/无效时 modelConfig 为空）。
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			conversationId	query		string	true	"会话ID"
//	@Success		200				{object}	response.Response{data=response.GetDigitalEmployeeConversationConfigResp}
//	@Router			/digital-employee/conversation/config [get]
func GetDigitalEmployeeConversationConfig(ctx *gin.Context) {
	var req request.GetDigitalEmployeeConversationConfigReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDigitalEmployeeConversationConfig(ctx, getUserID(ctx), getOrgID(ctx), req.ConversationId)
	gin_util.Response(ctx, resp, err)
}

// UpdateDigitalEmployeeConversationConfig
//
//	@Tags			digital-employee
//	@Summary		修改数字员工发布会话配置
//	@Description	更新数字员工发布会话的模型配置（对齐通用智能体 PUT /general/agent/conversation/config）。
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateDigitalEmployeeConversationConfigReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/digital-employee/conversation/config [put]
func UpdateDigitalEmployeeConversationConfig(ctx *gin.Context) {
	var req request.UpdateDigitalEmployeeConversationConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDigitalEmployeeConversationConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}
