package callback

import (
	"errors"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

//	@title		AI Agent Productivity Platform - Callback
//	@version	v0.0.1

//	@BasePath	/callback/v1

// FileUrlConvertBase64
//
//	@Tags		callback
//	@Summary	文件Url转换为base64
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.FileUrlConvertBase64Req	true	"文件Url转换base64请求参数"
//	@Success	200		{object}	response.Response{data=string}
//	@Router		/file/url/base64 [post]
func FileUrlConvertBase64(ctx *gin.Context) {
	var req request.FileUrlConvertBase64Req
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.FileUrlConvertBase64(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// UploadFileByBase64
//
//	@Tags		callback
//	@Summary	通过base64上传文件
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.UploadFileByBase64Req	true	"通过base64格式上传文件参数"
//	@Success	200		{object}	response.Response{data=response.UploadFileByBase64Resp}
//	@Router		/file/upload/base64 [post]
func UploadFileByBase64(ctx *gin.Context) {
	var req request.UploadFileByBase64Req
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.UploadFileByBase64(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// UnarchiveFile
//
//	@Tags			callback
//	@Summary		解压压缩包
//	@Description	解压MinIO中的压缩包，将文件上传到MinIO并返回目录结构
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UnarchiveFileReq	true	"解压压缩包请求参数"
//	@Success		200		{object}	response.Response{data=response.UnarchiveFileResp}
//	@Router			/file/unarchive [post]
func UnarchiveFile(ctx *gin.Context) {
	var req request.UnarchiveFileReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.UnarchiveFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// GetUserListByUserIds
//
//	@Tags		callback
//	@Summary	根据userIds获取用户信息
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.GetUserListByUserIdsReq	true	"根据userIds获取用户信息参数"
//	@Success	200		{object}	response.Response{data=response.ListResult{List=[]response.IDNameWithAvatar}}
//	@Router		/user/list [post]
func GetUserListByUserIds(ctx *gin.Context) {
	var req request.GetUserListByUserIdsReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetUserListByUserIds(ctx, req.UserIds)
	gin_util.Response(ctx, resp, err)
}

// UpdateDocStatus
//
//	@Tags			callback
//	@Summary		更新文档状态（模型扩展调用）
//	@Description	更新文档状态（模型扩展调用）
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CallbackUpdateDocStatusReq	true	"更新文档状态请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/api/docstatus [post]
func UpdateDocStatus(ctx *gin.Context) {
	var req request.CallbackUpdateDocStatusReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDocStatus(ctx, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateKnowledgeStatus
//
//	@Tags			callback
//	@Summary		更新知识库状态
//	@Description	更新知识库状态
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CallbackUpdateDocStatusReq	true	"更新知识库状态请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/api/knowledge/status [post]
func UpdateKnowledgeStatus(ctx *gin.Context) {
	var req request.CallbackUpdateKnowledgeStatusReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledgeStatus(ctx, &req)
	gin_util.Response(ctx, nil, err)
}

// DocStatusInit
//
//	@Tags			callback
//	@Summary		将正在解析的文档设置为解析失败
//	@Description	将正在解析的文档设置为解析失败
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{}
//	@Router			/api/doc_status_init [get]
func DocStatusInit(ctx *gin.Context) {
	resp, err := service.DocStatusInit(ctx, "", "")
	gin_util.Response(ctx, resp, err)
}

// GetDeployInfo
//
//	@Tags			callback
//	@Summary		获取Maas平台部署信息（模型扩展调用）
//	@Description	获取Maas平台部署信息（模型扩展调用）
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{}
//	@Router			/api/deploy/info [get]
func GetDeployInfo(ctx *gin.Context) {
	resp, err := service.GetDeployInfo(ctx)
	gin_util.Response(ctx, resp, err)
}

// SearchKnowledgeBase
//
//	@Tags			callback
//	@Summary		查询知识库列表（命中测试）
//	@Description	查询知识库列表（命中测试）
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagSearchKnowledgeBaseReq	true	"查询知识库列表请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/rag/search-knowledge-base [post]
func SearchKnowledgeBase(ctx *gin.Context) {
	var req request.RagSearchKnowledgeBaseReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.RagKnowledgeHit(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// KnowledgeStreamSearch
//
//	@Tags			callback
//	@Summary		知识库流式问答
//	@Description	知识库流式问答
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagKnowledgeChatReq	true	"知识库流式问答请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/rag/knowledge/stream/search [post]
func KnowledgeStreamSearch(ctx *gin.Context) {
	userId := ctx.GetHeader("X-uid")
	if userId == "" {
		resp, httpStatus := response.CommonRagKnowledgeError(errors.New("callback: empty X-uid"))
		gin_util.ResponseRawByte(ctx, httpStatus, resp)
		return
	}
	var req request.RagKnowledgeChatReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	req.UserId = userId
	err := service.KnowledgeStreamSearch(ctx, &req)
	if err != nil {
		resp, httpStatus := response.CommonRagKnowledgeError(err)
		gin_util.ResponseRawByte(ctx, httpStatus, resp)
		return
	}
}

// SearchQABase
//
//	@Tags			callback
//	@Summary		查询问答列表（命中测试）
//	@Description	查询问答列表（命中测试）
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RagSearchQABaseReq	true	"查询知识库列表请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/rag/search-qa-base [post]
func SearchQABase(ctx *gin.Context) {
	var req request.RagSearchQABaseReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, httpStatus := service.RagSearchQABase(ctx, &req)
	gin_util.ResponseRawByte(ctx, httpStatus, resp)
}
