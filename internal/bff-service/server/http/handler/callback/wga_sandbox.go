package callback

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// WgaSandboxRun
//
//	@Tags			wga-sandbox
//	@Summary		WGA沙箱运行
//	@Description	在沙箱容器中执行智能体任务，SSE流式返回结果
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			data	body		request.WgaSandboxRunReq	true	"请求参数"
//	@Success		200		{object}	string						"SSE流式返回"
//	@Router			/wga/sandbox/run [post]
func WgaSandboxRun(ctx *gin.Context) {
	var req request.WgaSandboxRunReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.WgaSandboxRun(ctx, &req); err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
}

// WgaSandboxCleanup
//
//	@Tags			wga-sandbox
//	@Summary		WGA沙箱清理
//	@Description	清理沙箱资源
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.WgaSandboxCleanupReq	true	"请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/wga/sandbox/cleanup [post]
func WgaSandboxCleanup(ctx *gin.Context) {
	var req request.WgaSandboxCleanupReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.WgaSandboxCleanup(ctx, &req)
	gin_util.Response(ctx, nil, err)
}
