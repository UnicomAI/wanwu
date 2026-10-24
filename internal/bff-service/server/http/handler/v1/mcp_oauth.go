package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	mcpoauth "github.com/UnicomAI/wanwu/pkg/mcp-oauth"
	"github.com/gin-gonic/gin"
	"net/http"
)

func MCPOAuthProviders(ctx *gin.Context) { gin_util.Response(ctx, service.MCPProviders(), nil) }
func MCPOAuthAction(ctx *gin.Context) {
	// Explicitly bound JSON only; owner is always derived from verified JWT.
	var req struct {
		Provider string `json:"provider"`
		State    string `json:"state"`
		Code     string `json:"code"`
	}
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 16384)
	if e := ctx.ShouldBindJSON(&req); e != nil {
		gin_util.Response(ctx, nil, mcpoauth.ErrInvalidConfig)
		return
	}
	o := mcpoauth.Owner{Organization: getOrgID(ctx), User: getUserID(ctx)}
	var data interface{}
	var err error
	switch ctx.Param("action") {
	case "prepare":
		var u string
		u, err = service.MCPPrepareOAuth(ctx, o, req.Provider)
		data = map[string]string{"authorizationUrl": u}
	case "complete":
		err = service.MCPCompleteOAuth(ctx, o, req.State, req.Code)
	case "status":
		var authorized bool
		authorized, err = service.MCPOAuthStatus(ctx, o, req.Provider)
		data = map[string]bool{"authorized": authorized}
	case "connect":
		err = service.MCPConnectOAuth(ctx, o, req.Provider)
	case "disconnect":
		err = service.MCPDisconnectOAuth(ctx, o, req.Provider)
	default:
		err = mcpoauth.ErrInvalidConfig
	}
	gin_util.Response(ctx, data, err)
}
