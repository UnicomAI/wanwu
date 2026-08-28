package middleware

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func AppHistoryRecord(filedId, appType string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		appID := getFieldValue(ctx, filedId)
		userID, _ := getUserID(ctx)
		detachedCtx := trace_util.DetachContext(ctx.Request.Context())
		ctx.Next()
		if appID == "" || userID == "" || appType == "" {
			log.Errorf("record user %v app %v type %v history err", userID, appID, appType)
			return
		}
		go func() {
			defer util.PrintPanicStack()
			if err := service.AddAppHistoryRecord(detachedCtx, userID, appID, appType); err != nil {
				log.Errorf("record user %v app %v type %v history err: %v", userID, appID, appType, err)
			}
		}()
	}
}
