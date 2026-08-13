package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

// registerNotice 消息中心。
//
// 挂在 common 命名空间（PermNeedEnable + JWTUser + CheckUserEnable）而不是新建权限 sub：
// 消息中心对所有登录用户开放，不是需要 RBAC 勾选的菜单项。
func registerNotice(apiV1 *gin.RouterGroup) {
	mid.Sub("common").Reg(apiV1, "/notice/unread/count", http.MethodGet, v1.GetNoticeUnreadCount, "未读总数+分类角标")
	mid.Sub("common").Reg(apiV1, "/notice/list", http.MethodGet, v1.ListNotice, "消息中心整页列表")
	mid.Sub("common").Reg(apiV1, "/notice/read", http.MethodPut, v1.ReadNotice, "消息已读（单条）")
	mid.Sub("common").Reg(apiV1, "/notice/read-all", http.MethodPut, v1.ReadAllNotice, "一键已读")
	mid.Sub("common").Reg(apiV1, "/notice", http.MethodDelete, v1.DeleteNotice, "批量删除消息")
}
