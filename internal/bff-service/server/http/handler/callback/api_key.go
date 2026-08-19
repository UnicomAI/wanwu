package callback

import (
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// GetUserInfoByApiKey
//
//	@Tags		callback
//	@Summary	通过apikey获取用户信息（内部接口）
//	@Accept		json
//	@Produce	json
//	@Param		apikey	query		string	true	"api key"
//	@Success	200		{object}	response.Response{data=response.UserInfoByApiKey}
//	@Router		/api/key/user [get]
func GetUserInfoByApiKey(ctx *gin.Context) {
	apiKey := ctx.Query("apikey")
	if apiKey == "" {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, "apikey is empty"))
		return
	}
	resp, err := service.GetUserInfoByApiKey(ctx, apiKey)
	gin_util.Response(ctx, resp, err)
}
