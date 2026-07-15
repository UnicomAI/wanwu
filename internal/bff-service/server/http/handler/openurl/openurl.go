package openurl

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

//	@title		AI Agent Productivity Platform - OpenUrl
//	@version	v0.0.1

//	@BasePath	/openurl/v1

// GetUrlAgentDetail
//
//	@Tags			openurl
//	@Summary		获取智能体url信息
//	@Description	获取智能体url信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-ID			header		string	true	"临时唯一标识"
//	@Param			suffix				path		string	true	"Url后缀"
//	@Success		200					{object}	response.Response{data=response.AppUrlConfig}
//	@Router			/agent/{suffix} 	[get]
func GetUrlAgentDetail(ctx *gin.Context) {
	resp, err := service.GetAppUrlInfo(ctx, ctx.Param("suffix"))
	gin_util.Response(ctx, resp, err)
}

// GetUrlAgentLLM
//
//	@Tags			openurl
//	@Summary		智能体模型配置
//	@Description	智能体模型配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-ID				header		string	true	"临时唯一标识"
//	@Param			suffix					path		string	true	"Url后缀"
//	@Success		200						{object}	response.Response{data=response.ModelInfo}
//	@Router			/agent/{suffix}/llm 	[get]
func GetUrlAgentLLM(ctx *gin.Context) {
	resp, err := service.GetAppUrlModel(ctx, ctx.Param("suffix"))
	gin_util.Response(ctx, resp, err)
}

// UrlConversationCreate
//
//	@Tags			openurl
//	@Summary		创建智能体对话
//	@Description	创建智能体对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-ID						header		string									true	"临时唯一标识"
//	@Param			suffix							path		string									true	"Url后缀"
//	@Param			data							body		request.UrlConversationCreateRequest	true	"智能体对话创建参数"
//	@Success		200								{object}	response.Response{data=response.ConversationCreateResp}
//	@Router			/agent/{suffix}/conversation 	[post]
func UrlConversationCreate(ctx *gin.Context) {
	var req request.UrlConversationCreateRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.UrlConversationCreate(ctx, req, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix"))
	gin_util.Response(ctx, resp, err)
}

// UrlConversationDelete
//
//	@Tags			openurl
//	@Summary		删除智能体对话
//	@Description	删除智能体对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-ID						header		string							true	"临时唯一标识"
//	@Param			suffix							path		string							true	"Url后缀"
//	@Param			data							body		request.ConversationIdRequest	true	"智能体对话的id"
//	@Success		200								{object}	response.Response
//	@Router			/agent/{suffix}/conversation 	[delete]
func UrlConversationDelete(ctx *gin.Context) {
	var req request.UrlConversationIdRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UrlConversationDelete(ctx, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix"), req)
	gin_util.Response(ctx, nil, err)
}

// UrlConversationClear
//
//	@Tags			openurl
//	@Summary		清空智能体对话
//	@Description	清空智能体对话ES数据，不删除会话ID
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-ID							header		string								true	"临时唯一标识"
//	@Param			suffix								path		string								true	"Url后缀"
//	@Param			data								body		request.UrlConversationIdRequest	true	"智能体对话清空参数"
//	@Success		200									{object}	response.Response
//	@Router			/agent/{suffix}/conversation/clear 	[delete]
func UrlConversationClear(ctx *gin.Context) {
	var req request.UrlConversationIdRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UrlConversationClear(ctx, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix"), req)
	gin_util.Response(ctx, nil, err)
}

// GetUrlConversationList
//
//	@Tags			openurl
//	@Summary		获取智能体对话列表
//	@Description	获取智能体对话列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-ID							header		string	true	"临时唯一标识"
//	@Param			suffix								path		string	true	"Url后缀"
//	@Param			pageNo								query		int		true	"页码"
//	@Param			pageSize							query		int		true	"每页数量"
//	@Param			searchText							query		string	false	"检索词"
//	@Success		200									{object}	response.Response{data=response.ListResult{list=[]response.ConversationInfo}}
//	@Router			/agent/{suffix}/conversation/list 								[get]
func GetUrlConversationList(ctx *gin.Context) {
	var req request.GetUrlConversationListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetUrlConversationList(ctx, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix"), req)
	gin_util.Response(ctx, resp, err)
}

// GetUrlConversationDetailList
//
//	@Tags			openurl
//	@Summary		智能体对话详情历史列表
//	@Description	智能体对话详情历史列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-ID								header		string	true	"临时唯一标识"
//	@Param			suffix									path		string	true	"Url后缀"
//	@Param			conversationId							query		string	true	"智能体对话id"
//	@Success		200										{object}	response.Response{data=response.ListResult{list=[]response.ConversationDetailInfo}}
//	@Router			/agent/{suffix}/conversation/detail 	[get]
func GetUrlConversationDetailList(ctx *gin.Context) {
	var req request.UrlConversationIdRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetUrlConversationDetailList(ctx, req, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix"))
	gin_util.Response(ctx, resp, err)
}

// AssistantUrlConversionStream
//
//	@Tags			openurl
//	@Summary		智能体流式问答
//	@Description	智能体流式问答
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-ID				header		string								true	"临时唯一标识"
//	@Param			suffix					path		string								true	"Url后缀"
//	@Param			data					body		request.UrlConversionStreamRequest	true	"智能体流式问答参数"
//	@Success		200						{object}	response.Response
//	@Router			/agent/{suffix}/stream 	[post]
func AssistantUrlConversionStream(ctx *gin.Context) {
	var req request.UrlConversionStreamRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.AppUrlConversionStream(ctx, req, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix")); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// GetAssistantPendingConversion
//
//	@Tags			openurl
//	@Summary		获取智能体运行中会话
//	@Description	获取智能体运行中会话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UrlPendingConversionRequest	true	"获取智能体运行中会话请求参数"
//	@Success		200		{object}	response.Response{data=response.PendingConversationResp}
//	@Router			/agent/:suffix/pending/conversation [get]
func GetAssistantPendingConversion(ctx *gin.Context) {
	var req request.UrlPendingConversionRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	conversation, err := service.AppUrlGetPendingConversation(ctx, req, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix"))
	if err != nil {
		gin_util.Response(ctx, nil, err)
	}
	gin_util.Response(ctx, conversation, err)
}

// AssistantConversionStreamConnect
//
//	@Tags			openurl
//	@Summary		草稿智能体流式问答断开后重连
//	@Description	草稿智能体流式问答断开后重连
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UrlConversionStreamConnectRequest	true	"草稿智能体流式问答断开后重连参数"
//	@Success		200		{object}	response.Response
//	@Router			/agent/:suffix/stream/connect [post]
func AssistantConversionStreamConnect(ctx *gin.Context) {
	var req request.UrlConversionStreamConnectRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}

	if err := service.AppUrlConversionStreamConnect(ctx, req, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix")); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// AssistantConversionStreamCancel
//
//	@Tags			openurl
//	@Summary		智能体流式问答手动停止
//	@Description	智能体流式问答手动停止
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UrlConversionStreamCancelRequest	true	"智能体流式问答手动停止参数"
//	@Success		200		{object}	response.Response
//	@Router			/agent/:suffix/stream/cancel [post]
func AssistantConversionStreamCancel(ctx *gin.Context) {
	var req request.UrlConversionStreamCancelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}

	if err := service.AppUrlConversionStreamCancel(ctx, req, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix")); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, nil, nil)
}

// AssistantUrlQuestionRecommend
//
//	@Tags			openurl
//	@Summary		智能体推荐问题
//	@Description	智能体推荐问题
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			X-Client-ID					header		string								true	"临时唯一标识"
//	@Param			suffix						path		string								true	"Url后缀"
//	@Param			data						body		request.UrlQuestionRecommendRequest	true	"智能体推荐问题参数"
//	@Success		200							{object}	response.Response
//	@Router			/agent/{suffix}/recommend 	[post]
func AssistantUrlQuestionRecommend(ctx *gin.Context) {
	var req request.UrlQuestionRecommendRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.AppUrlQuestionRecommend(ctx, req, ctx.GetHeader("X-Client-ID"), ctx.Param("suffix")); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// --- 文件上传（匿名访问） ---

// UploadFile
//
//	@Tags			openurl.file
//	@Summary		文件上传
//	@Description	分片文件上传（匿名访问）
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			fileName	formData	string	true	"原始文件名"
//	@Param			sequence	formData	int		true	"分片文件序号"
//	@Param			chunkName	formData	string	true	"上传批次标识"
//	@Param			files		formData	file	true	"文件"
//	@Success		200			{object}	response.Response{data=response.UploadFileResp}
//	@Router			/file/upload [post]
func UploadFile(ctx *gin.Context) {
	var req request.UploadFileReq
	if !gin_util.BindForm(ctx, &req) {
		return
	}
	resp, err := service.UploadFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// MergeFile
//
//	@Tags			openurl.file
//	@Summary		文件合并
//	@Description	合并分片文件（匿名访问）
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MergeFileReq	true	"文件合并参数"
//	@Success		200		{object}	response.Response{data=response.MergeFileResp}
//	@Router			/file/merge [post]
func MergeFile(ctx *gin.Context) {
	var req request.MergeFileReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.MergeFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// CleanFile
//
//	@Tags			openurl.file
//	@Summary		文件清除
//	@Description	清除已上传的分片文件（匿名访问）
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CleanFileReq	true	"文件清除参数"
//	@Success		200		{object}	response.Response{data=response.CleanFileResp}
//	@Router			/file/clean [post]
func CleanFile(ctx *gin.Context) {
	var req request.CleanFileReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CleanFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}
