package openapi

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CreateKnowledge
//
//	@Tags			openapi
//	@Summary		新建知识库OpenAPI
//	@Description	新建知识库OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateKnowledgeReq	true	"创建知识库请求参数"
//	@Success		200		{object}	response.Response{data=response.CreateKnowledgeResp}
//	@Router			/knowledge [post]
func CreateKnowledge(ctx *gin.Context) {
	var req request.CreateKnowledgeReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	resp, err := service.CreateKnowledgeOpenapi(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// UpdateKnowledge
//
//	@Tags			openapi
//	@Summary		修改知识库openapi
//	@Description	修改知识库openapi
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateKnowledgeReq	true	"修改知识库请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge [put]
func UpdateKnowledge(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateKnowledgeReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledge(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteKnowledge
//
//	@Tags			openapi
//	@Summary		删除知识库openapi
//	@Description	删除知识库openapi
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteKnowledge	true	"删除知识库请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge [delete]
func DeleteKnowledge(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteKnowledge
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.DeleteKnowledge(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// GetKnowledgeSelect
//
//	@Tags			openapi
//	@Summary		查询知识库列表openapi
//	@Description	查询知识库列表openapi
//	@Accept			json
//	@Param			data	body	request.KnowledgeSelectReq	true	"查询知识库列表"
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.KnowledgeListResp}
//	@Router			/knowledge/select [post]
func GetKnowledgeSelect(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeSelectReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.SelectKnowledgeList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// DirectUploadFiles
//
//	@Tags			openapi
//	@Summary		文件上传
//	@Description	文件上传
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			fileName	formData	string	true	"原始文件名"
//	@Param			file		formData	file	true	"文件"
//	@Success		200			{object}	response.Response{data=response.UploadFileResp}
//	@Router			/file/upload/direct [post]
func DirectUploadFiles(ctx *gin.Context) {
	var req request.DirectUploadFilesReq
	if !gin_util.BindForm(ctx, &req) {
		return
	}
	resp, err := service.DirectUploadFiles(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// GetDocConfig
//
//	@Tags			openapi
//	@Summary		获取文档配置信息
//	@Description	获取文档配置信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.DocConfigReq	true	"文档配置信息查询请求参数"
//	@Success		200		{object}	response.Response{data=response.DocConfigResult}
//	@Router			/knowledge/doc/config [get]
func GetDocConfig(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocConfigReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDocConfig(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// GetDocList
//
//	@Tags			openapi
//	@Summary		获取文档列表
//	@Description	获取知识库文档列表，不展示状态为无效（-1）的文档数据
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocListReq	true	"文档列表查询请求参数"
//	@Success		200		{object}	response.PageResult{list=response.ListDocResp}
//	@Router			/knowledge/doc/list [post]
func GetDocList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetDocList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// ImportDoc
//
//	@Tags			openapi
//	@Summary		上传文档
//	@Description	上传文档
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocImportReq	true	"文档上传请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/import [post]
func ImportDoc(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocImportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ImportDocOpenapi(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateDocConfig
//
//	@Tags			openapi
//	@Summary		更新文档配置
//	@Description	更新文档配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocConfigUpdateReq	true	"更新文档配置请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/update/config [post]
func UpdateDocConfig(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocConfigUpdateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDocConfigOpenapi(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteDoc
//
//	@Tags			openapi
//	@Summary		删除文档
//	@Description	删除文档
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteDocReq	true	"删除文档请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc [delete]
func DeleteDoc(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteDocReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteDoc(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetDocImportTip
//
//	@Tags			openapi
//	@Summary		获取知识库异步上传任务提示
//	@Description	获取知识库异步上传任务提示：有正在执行的异步上传任务/最近一次上传任务的失败信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.QueryKnowledgeReq	true	"获取知识库异步上传任务提示请求参数"
//	@Success		200		{object}	response.Response(data=response.DocImportTipResp)
//	@Router			/knowledge/doc/import/tip [get]
func GetDocImportTip(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.QueryKnowledgeReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDocImportTip(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// ExportKnowledgeDoc
//
//	@Tags			openapi
//	@Summary		知识库文档导出
//	@Description	知识库文档导出
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeDocExportReq	true	"知识库文档导出请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/export [post]
func ExportKnowledgeDoc(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeDocExportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ExportKnowledgeDoc(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetKnowledgeExportRecordList
//
//	@Tags			openapi
//	@Summary		获取知识库导出记录列表
//	@Description	获取知识库导出记录列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.KnowledgeExportRecordListReq	true	"获取知识库导出记录列表请求参数"
//	@Success		200		{object}	response.Response{data=response.KnowledgeExportRecordPageResult}
//	@Router			/knowledge/export/record/list [get]
func GetKnowledgeExportRecordList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeExportRecordListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetKnowledgeExportRecordList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// DeleteKnowledgeExportRecord
//
//	@Tags			openapi
//	@Summary		删除知识库导出记录
//	@Description	删除知识库导出记录
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteKnowledgeExportRecordReq	true	"删除知识库导出记录请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/export/record [delete]
func DeleteKnowledgeExportRecord(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteKnowledgeExportRecordReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteKnowledgeExportRecord(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// KnowledgeHit
//
//	@Tags			openapi
//	@Summary		知识库命中测试
//	@Description	知识库命中测试
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeHitReq	true	"知识库命中测试请求参数"
//	@Success		200		{object}	response.Response{data=response.KnowledgeHitResp}
//	@Router			/knowledge/hit [post]
func KnowledgeHit(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeHitReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.KnowledgeHitOpenapi(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// GetDocSegmentList
//
//	@Tags			openapi
//	@Summary		获取文档切分结果
//	@Description	获取文档切分结果
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.DocSegmentListReq	true	"获取文档切分结果请求参数"
//	@Success		200		{object}	response.Response{data=response.DocSegmentResp}
//	@Router			/knowledge/doc/segment/list [get]
func GetDocSegmentList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocSegmentListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDocSegmentList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// GetDocChildSegmentList
//
//	@Tags			openapi
//	@Summary		获取子分段列表
//	@Description	获取子分段列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.DocChildListReq	true	"获取子分段列表查询请求参数"
//	@Success		200		{object}	response.Response{data=response.DocChildSegmentResp}
//	@Router			/knowledge/doc/segment/child/list [get]
func GetDocChildSegmentList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocChildListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDocChildSegmentList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// UpdateDocSegmentStatus
//
//	@Tags			openapi
//	@Summary		更新文档切片启用状态
//	@Description	更新文档切片启用状态
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateDocSegmentStatusReq	true	"更新文档切片启用状态请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/status/update [post]
func UpdateDocSegmentStatus(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateDocSegmentStatusReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDocSegmentStatus(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// CreateDocSegment
//
//	@Tags			openapi
//	@Summary		新增文档切片
//	@Description	新增单条文档切片；多模态知识库的切片内容可包含图片markdown，普通知识库不允许
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateDocSegmentReq	true	"新增文档切片请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/create [post]
func CreateDocSegment(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.CreateDocSegmentReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateDocSegmentOpenapi(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// BatchCreateDocSegment
//
//	@Tags			openapi
//	@Summary		批量新增文档切片
//	@Description	通过csv文件批量新增文档切片
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.BatchCreateDocSegmentReq	true	"批量新增文档切片请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/batch/create [post]
func BatchCreateDocSegment(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.BatchCreateDocSegmentReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.BatchCreateDocSegment(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateDocSegment
//
//	@Tags			openapi
//	@Summary		更新文档切片
//	@Description	更新文档切片；多模态知识库的切片内容可包含图片markdown，普通知识库不允许
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateDocSegmentReq	true	"更新文档切片请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/update [post]
func UpdateDocSegment(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateDocSegmentReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDocSegmentOpenapi(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteDocSegment
//
//	@Tags			openapi
//	@Summary		删除文档切片
//	@Description	删除文档切片
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteDocSegmentReq	true	"删除文档切片请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/delete [delete]
func DeleteDocSegment(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteDocSegmentReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteDocSegment(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// UploadDocSegmentImage
//
//	@Tags			openapi
//	@Summary		分段图片上传
//	@Description	上传分段图片并返回markdown格式url，仅多模态知识库可用；返回的url可拼接进切片content中
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			knowledgeId	formData	string	true	"知识库id"
//	@Param			files		formData	file	true	"图片文件(png/jpg/jpeg)"
//	@Success		200			{object}	response.Response{data=response.RagUploadResponse}
//	@Router			/knowledge/doc/segment/image/upload [post]
func UploadDocSegmentImage(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UploadDocSegmentImageReq
	if !gin_util.BindForm(ctx, &req) {
		return
	}
	resp, err := service.UploadDocSegmentImageOpenapi(ctx, userId, orgId, req.KnowledgeId)
	gin_util.Response(ctx, resp, err)
}

// GetKnowledgeMetaKeyList
//
//	@Tags			openapi
//	@Summary		获取知识库元数据key列表
//	@Description	获取知识库已定义的元数据key列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.GetKnowledgeMetaSelectReq	true	"获取知识库元数据key列表请求参数"
//	@Success		200		{object}	response.Response{data=response.GetKnowledgeMetaSelectResp}
//	@Router			/knowledge/meta/list [get]
func GetKnowledgeMetaKeyList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.GetKnowledgeMetaSelectReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetKnowledgeMetaSelect(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// UpdateKnowledgeMetaKey
//
//	@Tags			openapi
//	@Summary		定义知识库元数据key
//	@Description	新增/修改/删除知识库的元数据key定义；文档元数据值的读写见 /knowledge/doc/meta/list 与 /knowledge/doc/meta/update
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeMetaKeyOpenapiReq	true	"定义知识库元数据key请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/meta/update [post]
func UpdateKnowledgeMetaKey(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeMetaKeyOpenapiReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledgeMetaKeyOpenapi(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetKnowledgeDocMetaValueList
//
//	@Tags			openapi
//	@Summary		获取文档元数据值列表
//	@Description	按文档id列表查询各文档的元数据值
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeMetaValueListReq	true	"获取文档元数据值列表请求参数"
//	@Success		200		{object}	response.Response{data=response.KnowledgeMetaValueListResp}
//	@Router			/knowledge/doc/meta/list [post]
func GetKnowledgeDocMetaValueList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeMetaValueListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetKnowledgeDocMetaValueListOpenapi(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// UpdateKnowledgeDocMetaValue
//
//	@Tags			openapi
//	@Summary		更新文档元数据值
//	@Description	批量更新文档的元数据值；key需已在知识库元数据中定义，option为add时不存在则新增、已存在则覆盖，update仅在已存在时更新
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeDocMetaValueOpenapiReq	true	"更新文档元数据值请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/meta/update [post]
func UpdateKnowledgeDocMetaValue(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeDocMetaValueOpenapiReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledgeDocMetaValueOpenapi(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
