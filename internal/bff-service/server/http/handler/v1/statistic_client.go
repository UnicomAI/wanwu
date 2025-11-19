package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetClientStatistic
//
//	@Tags			statistic_client
//	@Summary Get client statistics
//	@Description Get client statistics
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param startDate query string true "Start time (format yyyy-mm-dd)"
//	@Param endDate query string true "End time (format yyyy-mm-dd)"
//	@Success		200			{object}	response.Response{data=response.ClientStatistic}
//	@Router			/statistic/client [get]
func GetClientStatistic(ctx *gin.Context) {
	resp, err := service.GetClientStatistic(ctx, ctx.Query("startDate"), ctx.Query("endDate"))
	gin_util.Response(ctx, resp, err)
}
