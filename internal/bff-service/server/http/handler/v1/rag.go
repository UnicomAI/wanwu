package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
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
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ChatRagRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.ChatRagStream(ctx, userId, orgId, req, false, constant.BizSourceWeb); err != nil {
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
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.ChatRagRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.ChatRagStream(ctx, userId, orgId, req, true, constant.BizSourceWeb); err != nil {
		gin_util.Response(ctx, nil, err)
	}
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
