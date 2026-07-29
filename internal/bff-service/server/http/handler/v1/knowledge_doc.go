package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetDocKnowledgeDetail
//
//	@Tags			knowledge.doc
//	@Summary		获取文档知识库信息
//	@Description	获取文档知识库信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.DocKnowledgeDetailReq	true	"获取文档知识库信息请求参数"
//	@Success		200		{object}	response.Response{data=response.DocKnowledgeInfo}
//	@Router			/knowledge/doc/detail [get]
func GetDocKnowledgeDetail(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocKnowledgeDetailReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	//无需校验权限，中间件有校验
	resp, err := service.GetDocKnowledgeDetail(ctx, userId, orgId, &req, &service.DocKnowledgeParams{
		NeedGraph:      true,
		NeedPermission: true,
	})
	gin_util.Response(ctx, resp, err)
}

// GetDocConfig
//
//	@Tags			knowledge.doc
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
//	@Tags			knowledge.doc
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
//	@Tags			knowledge.doc
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
	err := service.ImportDoc(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateDocConfig
//
//	@Tags			knowledge.doc
//	@Summary		更新文档配置
//	@Description	更新文档配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocConfigUpdateReq	true	"文更新文档配置请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/update/config [post]
func UpdateDocConfig(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocConfigUpdateReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateDocConfig(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// ReImportDoc
//
//	@Tags			knowledge.doc
//	@Summary		重新解析文档
//	@Description	重新解析文档
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocReImportReq	true	"重新导入文档请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/doc/reimport [post]
func ReImportDoc(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DocReImportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ReImportDoc(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteDoc
//
//	@Tags			knowledge.doc
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

// UpdateDocMetaData
//
//	@Tags			knowledge.doc
//	@Summary		更新文档元数据
//	@Description	更新文档元数据
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocMetaDataReq	true	"文档更新元数据请求参数"
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

// GetDocImportTip
//
//	@Tags			knowledge.doc
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

// GetDocSegmentList
//
//	@Tags			knowledge.doc
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

// UpdateDocSegmentStatus
//
//	@Tags			knowledge.doc
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

// AnalysisDocUrl 解析url，已下架，路由注册同步注释
// func AnalysisDocUrl(ctx *gin.Context) {
// 	userId, orgId := getUserID(ctx), getOrgID(ctx)
// 	var req request.AnalysisUrlDocReq
// 	if !gin_util.Bind(ctx, &req) {
// 		return
// 	}
// 	resp, err := service.AnalysisDocUrl(ctx, userId, orgId, &req)
// 	gin_util.Response(ctx, resp, err)
// }

// UpdateDocSegmentLabels
//
//	@Tags			knowledge.doc
//	@Summary		更新文档切片标签
//	@Description	更新文档切片标签
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DocSegmentLabelsReq	true	"更新文档切片标签请求参数"
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
//	@Summary		新增文档切片
//	@Description	新增文档切片
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
	err := service.CreateDocSegment(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// BatchCreateDocSegment
//
//	@Tags			knowledge.doc
//	@Summary		批量新增文档切片
//	@Description	批量新增文档切片
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

// DeleteDocSegment
//
//	@Tags			knowledge.doc
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

// UpdateDocSegment
//
//	@Tags			knowledge.doc
//	@Summary		更新文档切片
//	@Description	更新文档切片
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
	err := service.UpdateDocSegment(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetDocChildSegmentList
//
//	@Tags			knowledge.doc
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

// CreateDocChildSegment
//
//	@Tags			knowledge.doc
//	@Summary		新增文档子分片
//	@Description	新增文档子分片
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateDocChildSegmentReq	true	"新增文档子切片请求参数"
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
//	@Summary		更新文档子切片
//	@Description	更新文档子切片
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateDocChildSegmentReq	true	"更新文档子切片请求参数"
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
//	@Summary		删除文档子切片
//	@Description	删除文档子切片
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteDocChildSegmentReq	true	"删除文档切片请求参数"
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

// ExportKnowledgeDoc
//
//	@Tags			knowledge.doc
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

// GetDocUploadLimit
//
//	@Tags			knowledge.doc
//	@Summary		获取可上传文件类型
//	@Description	获取可上传文件类型
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.QueryKnowledgeReq	true	"获取可上传文件类型请求参数"
//	@Success		200		{object}	response.Response{data=response.DocUploadLimitResp}
//	@Router			/knowledge/doc/upload/limit [get]
func GetDocUploadLimit(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.QueryKnowledgeReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDocUploadLimit(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// GetDocByKnowledgeNameAndDocName
//
//	@Tags			knowledge.doc
//	@Summary		根据知识库名称和文档名称获取文档信息
//	@Description	根据知识库名称和文档名称获取文档信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.GetDocByKnowledgeNameAndDocNameReq	true	"根据知识库名称和文档名称获取文档信息请求参数"
//	@Success		200		{object}	response.Response{data=response.GetDocByKnowledgeNameAndDocNameResp}
//	@Router			/knowledge/doc/by/name [get]
func GetDocByKnowledgeNameAndDocName(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.GetDocByKnowledgeNameAndDocNameReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetDocByKnowledgeNameAndDocName(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}
