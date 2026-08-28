package v1

import (
	"net/http"

	"github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/callback"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerV1Callback(callbackV1 *gin.RouterGroup) {
	mid.Sub("callback").Reg(callbackV1, "/api/deploy/info", http.MethodGet, callback.GetDeployInfo, "获取Maas平台部署信息（模型扩展调用）")
}
