package callback

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// SendMessage
//
//	@Tags			callback
//	@Summary		内部服务发消息
//	@Description	内部服务给指定通道发消息（无鉴权，callback 接口）。channelId 必填；userId 可选，缺省由 channel-service 自动取该通道最近互动过的 IM 用户作收件人。msgType 支持 text（默认）/markdown/file；file 类型需 fileUrl+fileName，channel-service 下载字节后投递。成功无返回体，失败透传 gRPC err_code。
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.SendMessageRequest	true	"内部服务发消息请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/callback/v1/channel/send-message [post]
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
