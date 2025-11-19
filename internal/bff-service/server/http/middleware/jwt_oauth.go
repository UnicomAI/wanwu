package middleware

import (
	"net/http"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	oauth2_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/oauth2-util"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
)

func JWTOAuthAccess(ctx *gin.Context) {
	token, err := getJWTToken(ctx)
	if err != nil {
		gin_util.ResponseDetail(ctx, http.StatusUnauthorized, codes.Code(err_code.Code_BFFJWT), nil, err.Error())
		ctx.Abort()
		return
	}
	jwtOAuthAccessAuth(ctx, token)
}

func jwtOAuthAccessAuth(ctx *gin.Context, token string) {
	httpStatus := http.StatusUnauthorized
	claims, err := oauth2_util.ParseAccessToken(token)
	if err != nil {
		gin_util.ResponseDetail(ctx, httpStatus, codes.Code(err_code.Code_BFFJWT), nil, err.Error())
		ctx.Abort()
		return
	}
	//验证sub，是否是access token [EN] Verify whether sub is an access token
	if claims.Subject != oauth2_util.SUBJECT_ACCESS {
		gin_util.ResponseDetail(ctx, httpStatus, codes.Code(err_code.Code_BFFJWT), nil, "token subject错误")
		ctx.Abort()
		return
	}

	ctx.Set(gin_util.USER_ID, claims.UserID)
	ctx.Set(gin_util.CLAIMS, claims)
	ctx.Next()
}
