package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateSensitiveWordTable
//
//	@Tags			safety
//	@Summary Create a sensitive word list
//	@Description Create a sensitive word list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.CreateSensitiveWordTableReq true "Create sensitive word table request parameters"
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
//	@Summary Edit sensitive word list
//	@Description Edit sensitive word list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UpdateSensitiveWordTableReq true "Edit sensitive word table request parameters"
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
//	@Summary Delete sensitive word list
//	@Description Delete sensitive word list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.DeleteSensitiveWordTableReq true "Delete sensitive word table request parameter"
//	@Success		200		{object}	response.Response
//	@Router			/safe/sensitive/table [delete]
func DeleteSensitiveWordTable(ctx *gin.Context) {
	var req request.DeleteSensitiveWordTableReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteSensitiveWordTable(ctx, &req)
	gin_util.Response(ctx, nil, err)
}

// GetSensitiveWordTableList
//
//	@Tags			safety
//	@Summary Get a list of sensitive words
//	@Description Get the list of sensitive words
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.ListResult{list=[]response.SensitiveWordTableDetail}}
//	@Router			/safe/sensitive/table/list [get]
func GetSensitiveWordTableList(ctx *gin.Context) {
	resp, err := service.GetSensitiveWordTableList(ctx, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetSensitiveVocabularyList
//
//	@Tags			safety
//	@Summary Get the vocabulary data list
//	@Description Get the vocabulary data list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data query request.GetSensitiveVocabularyReq true "Query vocabulary data list parameters"
//	@Param pageNo query int true "Page number, starting from 1"
//	@Param pageSize query int true "Number of single pages, starting from 1"
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
//	@Summary Upload sensitive words
//	@Description Upload sensitive words
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UploadSensitiveVocabularyReq true "Upload sensitive word parameters"
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
//	@Summary Delete sensitive words
//	@Description Delete sensitive words
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.DeleteSensitiveVocabularyReq true "Delete sensitive word parameters"
//	@Success		200		{object}	response.Response
//	@Router			/safe/sensitive/word [delete]
func DeleteSensitiveVocabulary(ctx *gin.Context) {
	var req request.DeleteSensitiveVocabularyReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteSensitiveVocabulary(ctx, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateSensitiveWordTableReply
//
//	@Tags			safety
//	@Summary Edit reply settings
//	@Description Edit reply settings
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data body request.UpdateSensitiveWordTableReplyReq true "Edit reply set request parameters"
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
//	@Summary Get a list of sensitive words (for drop-down selection)
//	@Description Get the list of sensitive words (for drop-down selection)
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.ListResult{list=[]response.SensitiveWordTableDetail}}
//	@Router			/safe/sensitive/table/select [get]
func GetSensitiveWordTableSelect(ctx *gin.Context) {
	resp, err := service.GetSensitiveWordTableList(ctx, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetSensitiveWordTable
//
//	@Tags			safety
//	@Summary Get the sensitive word list
//	@Description Get the sensitive word list
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param data query request.GetSensitiveVocabularyReq true "Query sensitive vocabulary parameters"
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
