package callback

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// SyncDigitalEmployeePublish
//
//	@Tags			callback
//	@Summary		数字员工发布状态同步
//	@Description	外部系统（ontology/vega）发布数字员工时回调本接口，登记 app 表 appType=digitalemployee 行（upsert，幂等）。万悟自拟格式：employeeId/publishType/userId/orgId。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DigitalEmployeePublishSyncReq	true	"发布状态同步请求"
//	@Success		200		{object}	response.Response
//	@Router			/digital-employee/publish/sync [post]
func SyncDigitalEmployeePublish(ctx *gin.Context) {
	var req request.DigitalEmployeePublishSyncReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.SyncDigitalEmployeePublish(ctx, &req)
	gin_util.Response(ctx, nil, err)
}

// SyncDigitalEmployeeUnpublish
//
//	@Tags			callback
//	@Summary		数字员工删除/下架同步
//	@Description	外部系统（ontology/vega）删除数字员工时回调本接口，删除 app 表 appType=digitalemployee 行（以 employeeId 为准，幂等，目标不存在也返回成功）。万悟自拟格式：employeeId/userId。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteDigitalEmployeePublishSyncReq	true	"删除/下架同步请求"
//	@Success		200		{object}	response.Response
//	@Router			/digital-employee/publish/sync [delete]
func SyncDigitalEmployeeUnpublish(ctx *gin.Context) {
	var req request.DeleteDigitalEmployeePublishSyncReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.SyncDigitalEmployeeUnpublish(ctx, &req)
	gin_util.Response(ctx, nil, err)
}
