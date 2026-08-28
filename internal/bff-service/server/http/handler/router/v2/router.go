package v2

import (
	"github.com/gin-gonic/gin"
)

func Register(apiV2 *gin.RouterGroup) {
	registerStatistic(apiV2)
}
