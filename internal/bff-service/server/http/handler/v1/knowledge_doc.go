package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetDocList
//
//	@Tags			knowledge.doc
//	@Summary		获取文档列表 [EN] @Summary Get the document list
//	@Description	获取知识库文档列表，不展示状态为无效（-1）的文档数据 [EN] @Description Gets the list of knowledge base documents and does not display document data with invalid status (-1)
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.DocListReq	true	"文档列表查询请求参数" [EN] @Param data query request.DocListReq true "Document list query request parameters"
//	@Success		200		{object}	response.Response{data=response.DocPageResult}
//	@Router			/knowledge/doc/list [get]
func GetDocList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDocList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// ImportDoc
//
//	@Tags			knowledge.doc
//	@Summary		上传文档 [EN] @Summary Upload document
//	@Description	上传文档 [EN] @Description upload document
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocImportReq	true	"文档上传请求参数" [EN] @Param data body request.DocImportReq true "Document upload request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/import [post]
func ImportDoc(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocImportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ImportDoc(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteDoc
//
//	@Tags			knowledge.doc
//	@Summary		删除文档 [EN] @Summary Delete document
//	@Description	删除文档 [EN] @Description delete document
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteDocReq	true	"删除文档请求参数" [EN] @Param data body request.DeleteDocReq true "Delete document request parameters"
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

// UpdateDocMetaData
//
//	@Tags			knowledge.doc
//	@Summary		更新文档元数据 [EN] @Summary updates document metadata
//	@Description	更新文档元数据 [EN] @Description updates document metadata
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocMetaDataReq	true	"文档更新元数据请求参数" [EN] @Param data body request.DocMetaDataReq true "Document update metadata request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/meta [post]
func UpdateDocMetaData(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocMetaDataReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDocMetaData(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// BatchUpdateDocMetaData
//
//	@Tags			knowledge.doc
//	@Summary		批量更新文档元数据 [EN] @Summary Batch update document metadata
//	@Description	批量更新文档元数据 [EN] @Description Batch update document metadata
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.BatchDocMetaDataReq	true	"批量文档更新元数据请求参数" [EN] @Param data body request.BatchDocMetaDataReq true "Batch document update metadata request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/meta/batch [post]
func BatchUpdateDocMetaData(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.BatchDocMetaDataReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.BatchUpdateDocMetaData(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetDocImportTip
//
//	@Tags			knowledge.doc
//	@Summary		获取知识库异步上传任务提示 [EN] @Summary Get knowledge base asynchronous upload task prompts
//	@Description	获取知识库异步上传任务提示：有正在执行的异步上传任务/最近一次上传任务的失败信息 [EN] @Description Obtain knowledge base asynchronous upload task prompt: there is an asynchronous upload task being executed/failure information of the latest upload task
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.QueryKnowledgeReq	true	"获取知识库异步上传任务提示请求参数" [EN] @Param data query request.QueryKnowledgeReq true "Get knowledge base asynchronous upload task prompt request parameters"
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

// GetDocSegmentList
//
//	@Tags			knowledge.doc
//	@Summary		获取文档切分结果 [EN] @Summary Get the document segmentation results
//	@Description	获取文档切分结果 [EN] @Description Get the document segmentation results
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.DocSegmentListReq	true	"获取文档切分结果请求参数" [EN] @Param data query request.DocSegmentListReq true "Get document segmentation result request parameters"
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

// UpdateDocSegmentStatus
//
//	@Tags			knowledge.doc
//	@Summary		更新文档切片启用状态 [EN] @Summary Update document slicing enabled status
//	@Description	更新文档切片启用状态 [EN] @Description updates document slicing enabled status
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateDocSegmentStatusReq	true	"更新文档切片启用状态请求参数" [EN] @Param data body request.UpdateDocSegmentStatusReq true "Update document slice enable status request parameter"
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

// AnalysisDocUrl
//
//	@Tags			knowledge.doc
//	@Summary		解析url [EN] @Summary parse url
//	@Description	解析url [EN] @Description parse url
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AnalysisUrlDocReq	true	"解析url请求参数" [EN] @Param data body request.AnalysisUrlDocReq true "Analysis url request parameters"
//	@Success		200		{object}	response.Response{data=response.AnalysisDocUrlResp}
//	@Router			/knowledge/doc/url/analysis [post]
func AnalysisDocUrl(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.AnalysisUrlDocReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.AnalysisDocUrl(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// UpdateDocSegmentLabels
//
//	@Tags			knowledge.doc
//	@Summary		更新文档切片标签 [EN] @Summary Update document slice tags
//	@Description	更新文档切片标签 [EN] @Description updates the document slice tag
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocSegmentLabelsReq	true	"更新文档切片标签请求参数" [EN] @Param data body request.DocSegmentLabelsReq true "Update document slice label request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/labels [post]
func UpdateDocSegmentLabels(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocSegmentLabelsReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDocSegmentLabels(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// CreateDocSegment
//
//	@Tags			knowledge.doc
//	@Summary		新增文档切片 [EN] @Summary New document slice
//	@Description	新增文档切片 [EN] @Description Add a new document slice
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateDocSegmentReq	true	"新增文档切片请求参数" [EN] @Param data body request.CreateDocSegmentReq true "New document slice request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/create [post]
func CreateDocSegment(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.CreateDocSegmentReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateDocSegment(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// BatchCreateDocSegment
//
//	@Tags			knowledge.doc
//	@Summary		批量新增文档切片 [EN] @Summary Add document slices in batches
//	@Description	批量新增文档切片 [EN] @Description Add document slices in batches
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.BatchCreateDocSegmentReq	true	"批量新增文档切片请求参数" [EN] @Param data body request.BatchCreateDocSegmentReq true "Batch add document slice request parameters"
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

// DeleteDocSegment
//
//	@Tags			knowledge.doc
//	@Summary		删除文档切片 [EN] @Summary Delete document slices
//	@Description	删除文档切片 [EN] @Description delete document slice
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteDocSegmentReq	true	"删除文档切片请求参数" [EN] @Param data body request.DeleteDocSegmentReq true "Delete document slice request parameters"
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

// UpdateDocSegment
//
//	@Tags			knowledge.doc
//	@Summary		更新文档切片 [EN] @Summary Update document slice
//	@Description	更新文档切片 [EN] @Description updates document slice
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateDocSegmentReq	true	"更新文档切片请求参数" [EN] @Param data body request.UpdateDocSegmentReq true "Update document slice request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/update [post]
func UpdateDocSegment(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateDocSegmentReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDocSegment(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetDocChildSegmentList
//
//	@Tags			knowledge.doc
//	@Summary		获取子分段列表 [EN] @Summary Get the list of sub-segments
//	@Description	获取子分段列表 [EN] @Description Get the list of sub-segments
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.DocChildListReq	true	"获取子分段列表查询请求参数" [EN] @Param data query request.DocChildListReq true "Get sub-segment list query request parameters"
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

// CreateDocChildSegment
//
//	@Tags			knowledge.doc
//	@Summary		新增文档子分片 [EN] @Summary Add new document sub-shards
//	@Description	新增文档子分片 [EN] @Description Add a new document sub-shard
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateDocChildSegmentReq	true	"新增文档子切片请求参数" [EN] @Param data body request.CreateDocChildSegmentReq true "Add document sub-slice request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/child/create [post]
func CreateDocChildSegment(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.CreateDocChildSegmentReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.CreateDocChildSegment(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateDocChildSegment
//
//	@Tags			knowledge.doc
//	@Summary		更新文档子切片 [EN] @Summary Update document subslice
//	@Description	更新文档子切片 [EN] @Description updates document subslice
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateDocChildSegmentReq	true	"更新文档子切片请求参数" [EN] @Param data body request.UpdateDocChildSegmentReq true "Update document subsegment request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/child/update [post]
func UpdateDocChildSegment(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateDocChildSegmentReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDocChildSegment(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteDocChildSegment
//
//	@Tags			knowledge.doc
//	@Summary		删除文档子切片 [EN] @Summary Delete document subslice
//	@Description	删除文档子切片 [EN] @Description delete document subslice
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteDocChildSegmentReq	true	"删除文档切片请求参数" [EN] @Param data body request.DeleteDocChildSegmentReq true "Delete document slice request parameters"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/segment/child/delete [delete]
func DeleteDocChildSegment(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteDocChildSegmentReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteDocChildSegment(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
