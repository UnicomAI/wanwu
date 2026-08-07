package v1

import (
	"net/http"

	v1 "github.com/UnicomAI/wanwu/internal/bff-service/server/http/handler/v1"
	mid "github.com/UnicomAI/wanwu/pkg/gin-util/mid-wrap"
	"github.com/gin-gonic/gin"
)

func registerStatistic(apiV1 *gin.RouterGroup) {
	// filter & select（V2 看板仍复用；V1 统计读写已下线）
	mid.Sub("app_observability.statistic").Reg(apiV1, "/statistic/orgs/select", http.MethodGet, v1.GetStatisticOrgsSelect, "获取统计看板组织下拉列表")
	mid.Sub("app_observability.statistic").Reg(apiV1, "/statistic/users/select", http.MethodGet, v1.GetStatisticUsersSelect, "获取统计看板用户下拉列表")
}
