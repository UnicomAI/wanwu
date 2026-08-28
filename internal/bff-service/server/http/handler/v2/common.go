package v2

import (
	"net/http"
	"net/url"

	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

//	@title						AI Agent Productivity Platform API V2
//	@version					v0.0.1
//	@securityDefinitions.apikey	JWT
//	@in							header
//	@name						Authorization

//	@BasePath	/v2

func getUserID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.USER_ID)
}
func getOrgID(ctx *gin.Context) string {
	return ctx.GetHeader(gin_util.X_ORG_ID)
}
func isAdmin(ctx *gin.Context) bool {
	return ctx.GetBool(gin_util.IS_ADMIN)
}
func isSystem(ctx *gin.Context) bool {
	return ctx.GetBool(gin_util.IS_SYSTEM)
}

// writeExcelExport 将 Excel 以附件形式写入响应。
func writeExcelExport(ctx *gin.Context, wb *util.Workbook, fileName string) {
	if wb == nil {
		return
	}
	defer func() { _ = wb.Close() }()
	ctx.Writer.WriteHeader(http.StatusOK)
	ctx.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.QueryEscape(fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	if _, err := wb.WriteTo(ctx.Writer); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}
