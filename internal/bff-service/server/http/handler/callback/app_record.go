package callback

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/gin-gonic/gin"
)

// AppRecord
//
//	@Tags			callback
//	@Summary		应用使用记录
//	@Description	应用使用记录
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.AppRecordRequest	true	"应用使用记录请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/app/record [post]
func AppRecord(ctx *gin.Context) {
	var req request.AppRecordRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	service.RecordAppStatistic(trace_util.DetachContext(ctx.Request.Context()), req.UserID, req.OrgID, req.AppID, req.AppType, req.Module,
		req.StatusCode, req.FailureReason, req.IsStream, req.StreamCosts, req.NonStreamCosts, req.Source, req.RequestBody, req.ResponseBody, req.Question, req.Answer)
	gin_util.Response(ctx, nil, nil)
}
