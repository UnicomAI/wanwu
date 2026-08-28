package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GreateKnowledgeQAPair
//
//	@Tags			knowledge.qa
//	@Summary		新增问答对
//	@Description	新增问答对
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateKnowledgeQAPairReq	true	"新增问答对请求参数"
//	@Success		200		{object}	response.Response{data=response.CreateKnowledgeQAPairResp}
//	@Router			/knowledge/qa/pair [post]
func GreateKnowledgeQAPair(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.CreateKnowledgeQAPairReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateKnowledgeQAPair(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// UpdateKnowledgeQAPair
//
//	@Tags			knowledge.qa
//	@Summary		编辑问答对
//	@Description	编辑问答对
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateKnowledgeQAPairReq	true	"编辑问答对请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/qa/pair [put]
func UpdateKnowledgeQAPair(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateKnowledgeQAPairReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledgeQAPair(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// UpdateKnowledgeQAPairSwitch
//
//	@Tags			knowledge.qa
//	@Summary		启停问答对
//	@Description	启停问答对
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateKnowledgeQAPairSwitchReq	true	"启停问答对请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/qa/pair/switch [put]
func UpdateKnowledgeQAPairSwitch(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.UpdateKnowledgeQAPairSwitchReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateKnowledgeQAPairSwitch(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteKnowledgeQAPair
//
//	@Tags			knowledge.qa
//	@Summary		刪除问答对
//	@Description	刪除问答对
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteKnowledgeQAPairReq	true	"刪除问答对请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/qa/pair [delete]
func DeleteKnowledgeQAPair(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.DeleteKnowledgeQAPairReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteKnowledgeQAPair(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetKnowledgeQAPairList
//
//	@Tags			knowledge.qa
//	@Summary		获取问答对列表
//	@Description	获取问答对列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.KnowledgeQAPairListReq	true	"问答对列表查询请求参数"
//	@Success		200		{object}	response.PageResult{list=response.ListKnowledgeQAPairResp}
//	@Router			/knowledge/qa/pair/list [post]
func GetKnowledgeQAPairList(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeQAPairListReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.GetKnowledgeQAPairList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// ImportKnowledgeQAPair
//
//	@Tags			knowledge.qa
//	@Summary		问答库文档导入
//	@Description	问答库文档导入
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeQAPairImportReq	true	"问答库文档导入请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/qa/pair/import [post]
func ImportKnowledgeQAPair(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeQAPairImportReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.ImportKnowledgeQAPair(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}

// GetKnowledgeQAPairImportTip
//
//	@Tags			knowledge.qa
//	@Summary		获取问答库异步上传任务提示
//	@Description	获取问答库异步上传任务提示：有正在执行的异步上传任务/最近一次上传任务的失败信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.KnowledgeQAPairImportTipReq	true	"获取问答库异步上传任务提示请求参数"
//	@Success		200		{object}	response.Response{data=response.KnowledgeQAPairImportTipResp}
//	@Router			/knowledge/qa/pair/import/tip [get]
func GetKnowledgeQAPairImportTip(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeQAPairImportTipReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetKnowledgeQAPairImportTip(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// KnowledgeQAHit
//
//	@Tags			knowledge
//	@Summary		问答库命中测试
//	@Description	问答库命中测试
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.KnowledgeHitReq	true	"问答库命中测试请求参数"
//	@Success		200		{object}	response.Response{data=response.KnowledgeQAHitResp}
//	@Router			/knowledge/qa/hit [post]
func KnowledgeQAHit(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeQAHitReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.KnowledgeQAHit(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// ExportKnowledgeQAPair
//
//	@Tags			knowledge.qa
//	@Summary		问答库文档导出
//	@Description	问答库文档导出
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.KnowledgeQAPairExportReq	true	"问答库文档导出请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/knowledge/qa/export [get]
func ExportKnowledgeQAPair(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeQAPairExportReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	err := service.ExportKnowledgeQAPair(ctx, userId, orgId, &req)
	gin_util.Response(ctx, nil, err)
}
