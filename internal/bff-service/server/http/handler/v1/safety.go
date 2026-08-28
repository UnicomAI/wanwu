package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateSensitiveWordTable
//
//	@Tags			safety
//	@Summary		创建敏感词表
//	@Description	创建敏感词表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateSensitiveWordTableReq	true	"创建敏感词表请求参数"
//	@Success		200		{object}	response.Response{data=response.CreateSensitiveWordTableResp}
//	@Router			/safe/sensitive/table [post]
func CreateSensitiveWordTable(ctx *gin.Context) {
	var req request.CreateSensitiveWordTableReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateSensitiveWordTable(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// UpdateSensitiveWordTable
//
//	@Tags			safety
//	@Summary		编辑敏感词表
//	@Description	编辑敏感词表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateSensitiveWordTableReq	true	"编辑敏感词表请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/safe/sensitive/table [put]
func UpdateSensitiveWordTable(ctx *gin.Context) {
	var req request.UpdateSensitiveWordTableReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateSensitiveWordTable(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteSensitiveWordTable
//
//	@Tags			safety
//	@Summary		删除敏感词表
//	@Description	删除敏感词表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteSensitiveWordTableReq	true	"删除敏感词表请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/safe/sensitive/table [delete]
func DeleteSensitiveWordTable(ctx *gin.Context) {
	var req request.DeleteSensitiveWordTableReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteSensitiveWordTable(ctx, &req, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, nil, err)
}

// GetSensitiveWordTableList
//
//	@Tags			safety
//	@Summary		获取敏感词表列表
//	@Description	获取敏感词表列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			type	query		string	true	"敏感词表类型，personal：个人，global：全局"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.SensitiveWordTableDetail}}
//	@Router			/safe/sensitive/table/list [get]
func GetSensitiveWordTableList(ctx *gin.Context) {
	resp, err := service.GetSensitiveWordTableList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("type"))
	gin_util.Response(ctx, resp, err)
}

// GetSensitiveVocabularyList
//
//	@Tags			safety
//	@Summary		获取词表数据列表
//	@Description	获取词表数据列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data		query		request.GetSensitiveVocabularyReq	true	"查询词表数据列表参数"
//	@Param			pageNo		query		int									true	"页面编号，从1开始"
//	@Param			pageSize	query		int									true	"单页数量，从1开始"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.SensitiveWordVocabularyDetail}}
//	@Router			/safe/sensitive/word/list [get]
func GetSensitiveVocabularyList(ctx *gin.Context) {
	var req request.GetSensitiveVocabularyReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetSensitiveVocabularyList(ctx, getUserID(ctx), getOrgID(ctx), getPageNo(ctx), getPageSize(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// UploadSensitiveVocabulary
//
//	@Tags			safety
//	@Summary		上传敏感词
//	@Description	上传敏感词
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UploadSensitiveVocabularyReq	true	"上传敏感词参数"
//	@Success		200		{object}	response.Response
//	@Router			/safe/sensitive/word [post]
func UploadSensitiveVocabulary(ctx *gin.Context) {
	var req request.UploadSensitiveVocabularyReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UploadSensitiveVocabulary(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteSensitiveVocabulary
//
//	@Tags			safety
//	@Summary		删除敏感词
//	@Description	删除敏感词
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteSensitiveVocabularyReq	true	"删除敏感词参数"
//	@Success		200		{object}	response.Response
//	@Router			/safe/sensitive/word [delete]
func DeleteSensitiveVocabulary(ctx *gin.Context) {
	var req request.DeleteSensitiveVocabularyReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteSensitiveVocabulary(ctx, &req, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, nil, err)
}

// UpdateSensitiveWordTableReply
//
//	@Tags			safety
//	@Summary		编辑回复设置
//	@Description	编辑回复设置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateSensitiveWordTableReplyReq	true	"编辑回复设置请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/safe/sensitive/table/reply [put]
func UpdateSensitiveWordTableReply(ctx *gin.Context) {
	var req request.UpdateSensitiveWordTableReplyReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateSensitiveWordTableReply(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// GetSensitiveWordTableSelect
//
//	@Tags			safety
//	@Summary		获取敏感词表列表（用于下拉选择）
//	@Description	获取敏感词表列表（用于下拉选择）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.ListResult{list=[]response.SensitiveWordTableDetail}}
//	@Router			/safe/sensitive/table/select [get]
func GetSensitiveWordTableSelect(ctx *gin.Context) {
	resp, err := service.GetSensitiveWordTableList(ctx, getUserID(ctx), getOrgID(ctx), constant.SensitiveTableTypePersonal)
	gin_util.Response(ctx, resp, err)
}

// GetSensitiveWordTable
//
//	@Tags			safety
//	@Summary		获取敏感词表
//	@Description	获取敏感词表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.GetSensitiveVocabularyReq	true	"查询敏感词表参数"
//	@Success		200		{object}	response.Response{data=response.SensitiveWordTableDetail}
//	@Router			/safe/sensitive/table [get]
func GetSensitiveWordTable(ctx *gin.Context) {
	var req request.GetSensitiveVocabularyReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetSensitiveWordTableByID(ctx, &req)
	gin_util.Response(ctx, resp, err)
}
