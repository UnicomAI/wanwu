package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerChannel(apiV1 *gin.RouterGroup) {
	// 万悟平台代理
	mid.Sub("operation").Reg(apiV1, "/channel/agent", http.MethodGet, v1.ListWanwuAgents, "获取万悟智能体列表")
	mid.Sub("operation").Reg(apiV1, "/channel/apikeys", http.MethodGet, v1.ListWanwuApiKeys, "获取万悟API Key列表")
	mid.Sub("operation").Reg(apiV1, "/channel/models", http.MethodGet, v1.ListWanwuModels, "获取万悟模型列表")

	// 扫码登录
	mid.Sub("operation").Reg(apiV1, "/channel/qrcode/:channelType", http.MethodPost, v1.CreateQRLogin, "发起扫码登录")
	mid.Sub("operation").Reg(apiV1, "/channel/qrcode/:channelType/status/:sessionId", http.MethodGet, v1.GetQRLoginStatus, "查询扫码状态")
	mid.Sub("operation").Reg(apiV1, "/channel/qrcode/:channelType/complete/:sessionId", http.MethodPost, v1.CompleteQRLogin, "完成扫码登录")
	mid.Sub("operation").Reg(apiV1, "/channel/qrcode/:channelType/:sessionId", http.MethodDelete, v1.CancelQRLogin, "取消扫码登录")

	// 通道管理
	mid.Sub("operation").Reg(apiV1, "/channel/channels", http.MethodPost, v1.CreateChannel, "创建通道")
	mid.Sub("operation").Reg(apiV1, "/channel/channels", http.MethodGet, v1.ListChannels, "获取通道列表")
	mid.Sub("operation").Reg(apiV1, "/channel/channels/:id", http.MethodGet, v1.GetChannel, "获取通道详情")
	mid.Sub("operation").Reg(apiV1, "/channel/channels/:id", http.MethodPut, v1.UpdateChannel, "更新通道")
	mid.Sub("operation").Reg(apiV1, "/channel/channels/:id/status", http.MethodPost, v1.UpdateChannelStatus, "启用/停用通道")
	mid.Sub("operation").Reg(apiV1, "/channel/channels/:id", http.MethodDelete, v1.DeleteChannel, "删除通道")
	mid.Sub("operation").Reg(apiV1, "/channel/channels/:id/disconnect", http.MethodPost, v1.DisconnectChannel, "断开通道")

	// WGA（通用智能体）子智能体列表
	mid.Sub("operation").Reg(apiV1, "/channel/wga/sub-agents", http.MethodGet, v1.ListWanwuWGASubAgents, "WGA子智能体列表")

	// DIP（数字员工）下拉
	mid.Sub("operation").Reg(apiV1, "/channel/dip/employees", http.MethodGet, v1.ListWanwuDIPAgents, "数字员工列表")
}
