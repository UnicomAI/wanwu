package v1

import (
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	jwt_util "github.com/UnicomAI/wanwu/pkg/jwt-util"
	svc_auth "github.com/UnicomAI/wanwu/pkg/svc-auth"
	"github.com/gin-gonic/gin"
)

// WorkflowAuth
//
//	@Tags			guest
//	@Summary		工作流边界鉴权（供 nginx auth_request 调用）
//	@Description	校验平台 JWT + 用户有效性 + 组织归属；通过则以响应头回传已认证身份，由 nginx 覆写到转发给 workflow 的请求上。仅从 Authorization 取身份，不读入站 X-User-Id/X-Org-Id
//	@Param			Authorization	header	string	false	"Bearer <平台用户 JWT>"
//	@Param			X-Org-Id		header	string	false	"当前组织"
//	@Param			X-Original-URI	header	string	false	"原始请求 URI（含查询串）"
//	@Success		200				"认证通过（响应头携带 X-Auth-User-Id / X-Auth-Org-Id，启用内部鉴权时另附 X-Auth-Sign / X-Auth-Ts）"
//	@Failure		401				"未认证"
//	@Failure		403				"组织越权"
//	@Router			/v1/workflow/auth [get]
func WorkflowAuth(ctx *gin.Context) {
	// 取原始请求 URI（去查询串）用于路径分类
	origPath := ctx.GetHeader("X-Original-URI")
	if i := strings.IndexByte(origPath, '?'); i >= 0 {
		origPath = origPath[:i]
	}

	// 自鉴权路径（聊天/上传等）放行：交 workflow 的 OpenapiAuthMW 真鉴权，此处不注入身份
	if isWorkflowSelfAuthPath(origPath) {
		ctx.Status(http.StatusOK)
		return
	}

	// 解析 Authorization: Bearer <token>
	token := ""
	if auth := ctx.GetHeader("Authorization"); auth != "" {
		if parts := strings.SplitN(auth, " ", 2); len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
			token = strings.TrimSpace(parts[1])
		}
	}

	// 校验平台用户 JWT（HS256，Subject 必须为 user）
	claims, err := jwt_util.ParseToken(token)
	if err != nil || claims == nil || claims.Subject != jwt_util.SUBJECT_USER || claims.UserID == "" {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	// 校验用户有效性（token 失效即拒绝）
	genTokenTS := "0"
	if claims.ExpiresAt != nil {
		genTokenTS = strconv.Itoa(int(claims.ExpiresAt.Unix() * 1000))
	}
	if _, err := service.CheckUserEnable(ctx, claims.UserID, genTokenTS); err != nil {
		ctx.Status(http.StatusUnauthorized)
		return
	}

	// 校验组织归属：X-Org-Id 必须属于该用户可选组织，防跨组织伪造
	orgID := ctx.GetHeader(gin_util.X_ORG_ID)
	if orgID != "" {
		sel, err := service.GetOrgSelect(ctx, claims.UserID)
		inOrg := false
		if err == nil && sel != nil {
			for _, o := range sel.Select {
				if o.ID == orgID {
					inOrg = true
					break
				}
			}
		}
		if !inOrg {
			ctx.Status(http.StatusForbidden)
			return
		}
	}

	// 以响应头回传已认证身份，供 nginx auth_request_set 覆写下游请求头
	ctx.Header("X-Auth-User-Id", claims.UserID)
	ctx.Header("X-Auth-Org-Id", orgID)

	// 启用内部鉴权时附网关签名，供 workflow 正向信任校验
	if secret := os.Getenv("WANWU_INTERNAL_AUTH_SECRET"); secret != "" {
		ts := time.Now().Unix()
		ctx.Header(svc_auth.HeaderTS, strconv.Itoa(int(ts)))
		ctx.Header(svc_auth.HeaderSign, svc_auth.Sign(secret, claims.UserID, orgID, ts))
	}
	ctx.Status(http.StatusOK)
}

// workflowSelfAuthPaths/Regex 镜像 wanwu-workflow 的 needAuthPath/needAuthFunc（backend/api/middleware/openapi_auth.go）；
// 变更需同步（建议加 CI 交叉校验）。
var workflowSelfAuthPaths = map[string]bool{
	"/v1/conversations":                true,
	"/v1/conversation/create":          true,
	"/v1/conversation/message/list":    true,
	"/v1/conversation/retrieve":        true,
	"/v1/files/upload":                 true,
	"/v1/workflow/run":                 true,
	"/v1/workflow/stream_run":          true,
	"/v1/workflow/stream_resume":       true,
	"/v1/workflow/get_run_history":     true,
	"/v1/workflow/conversation/create": true,
	"/v1/bot/get_online_info":          true,
	"/v1/workflows/chat":               true,
	"/v1/datasets":                     true,
	// 例外：工作流编辑器「试运行」表单的文件直传（ImageX apply_upload_action 领取上传凭证 + 文件写入
	// /api/common/upload/<tos_uri>）由浏览器裸 XHR 发起、不携带平台 JWT，无法通过边界 JWT 校验；
	// 此处显式放行，交 workflow 侧处理（workflow 的 upload 路径本就不在 needAuthPath 内）。
	"/api/common/upload/apply_upload_action": true,
}

var workflowSelfAuthPathRegex = []*regexp.Regexp{
	regexp.MustCompile(`^/v1/conversations/[0-9]+$`),
	regexp.MustCompile(`^/v1/conversations/[0-9]+/clear$`),
	regexp.MustCompile(`^/v1/bots/[0-9]+$`),
	regexp.MustCompile(`^/v1/workflows/[0-9]+$`),
	regexp.MustCompile(`^/v1/apps/[0-9]+$`),
	regexp.MustCompile(`^/v1/datasets/[0-9]+$`),
	regexp.MustCompile(`^/v1/datasets/[0-9]+/images$`),
	regexp.MustCompile(`^/v1/datasets/[0-9]+/process$`),
	regexp.MustCompile(`^/api/common/upload/`),
}

func isWorkflowSelfAuthPath(path string) bool {
	if path == "" {
		return false
	}
	if workflowSelfAuthPaths[path] {
		return true
	}
	for _, re := range workflowSelfAuthPathRegex {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}
