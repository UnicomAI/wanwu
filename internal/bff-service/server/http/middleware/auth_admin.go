package middleware

import (
	"errors"
	"net/http"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// CheckOrgAdmin 校验当前用户对请求中指定组织（orgId）是否拥有管理员权限。
// 从 query、json body、form 中依次提取 orgId，校验用户对目标组织的管理权。
// 作为 router 层中间件使用，避免在每个 handler 中重复 IsAdminInOrgs 校验。
func CheckOrgAdmin(ctx *gin.Context) {
	defer util.PrintPanicStack()

	userID, err := getUserID(ctx)
	if err != nil {
		gin_util.ResponseErrCodeKeyWithStatus(ctx, http.StatusForbidden, err_code.Code_BFFAuth, "", err.Error())
		ctx.Abort()
		return
	}

	orgID := getFieldValue(ctx, "orgId")
	if orgID == "" {
		gin_util.ResponseErrWithStatus(ctx, http.StatusBadRequest, errors.New("orgId is required"))
		ctx.Abort()
		return
	}

	if !service.IsAdminInOrgs(ctx, userID, orgID) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_org_admin_required"))
		ctx.Abort()
		return
	}

	ctx.Next()
}

// CheckSystemAdmin 校验当前组织是否为系统组织且当前用户是否为系统管理员。
// 用于 admin_center.setting 和 admin_center.oauth 等仅在系统组织下可操作的接口。
// 依赖 CheckUserPerm 中间件已在 ctx 中设置 IS_SYSTEM 和 IS_ADMIN。
func CheckSystemAdmin(ctx *gin.Context) {
	defer util.PrintPanicStack()

	if !ctx.GetBool(gin_util.IS_SYSTEM) || !ctx.GetBool(gin_util.IS_ADMIN) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_system_org_required"))
		ctx.Abort()
		return
	}

	ctx.Next()
}
