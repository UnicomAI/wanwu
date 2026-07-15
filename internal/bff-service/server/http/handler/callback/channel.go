package callback

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// SendMessage 内部服务发消息（无鉴权，callback 接口）。
// 供内部其他服务给指定通道发消息：channelId/content 必填，userId 可选（缺省由 channel-service 自动取最近互动用户）。
// 成功无返回体（仅 code/msg），失败透传 gRPC err_code。
func SendMessage(ctx *gin.Context) {
	var req request.SendMessageRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.SendMessage(ctx.Request.Context(), &req); err != nil {
		gin_util.ResponseErr(ctx, err)
		return
	}
	gin_util.ResponseOK(ctx)
}
