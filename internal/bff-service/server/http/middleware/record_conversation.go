package middleware

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func RecordConversation(adminCenterBiz *AdminCenterBiz, sourceFrom string) func(ctx *gin.Context) {
	return func(ctx *gin.Context) {
		defer util.PrintPanicStack()
		conversationID := getFieldValue(ctx, adminCenterBiz.BizId)
		ctx.Next()
		if len(conversationID) == 0 {
			conversationID = ctx.GetString(gin_util.CONVERSATION_ID)
		}
		if len(conversationID) == 0 {
			log.Errorf("record conversation log skip: conversationId empty, bizType %s", adminCenterBiz.BizType)
			return
		}

		ctxCopy := ctx.Copy()
		ctxCopy.Request = ctxCopy.Request.WithContext(trace_util.DetachContext(ctx.Request.Context()))
		safe_go_util.SafeGo(func() {
			executeConversationLog(ctxCopy, adminCenterBiz.BizType, conversationID, sourceFrom)
		})

	}
}

// executeConversationLog 执行会话记录
func executeConversationLog(ctx *gin.Context, bizType, conversationID, sourceFrom string) {
	conversationLog, err := service.ConversationLog(ctx, bizType, conversationID, sourceFrom)
	if err != nil {
		log.Errorf("conversation log search error %s", err)
		return
	}
	if err = service.RecordConversationLog(ctx, conversationLog); err != nil {
		log.Errorf("record conversation log: %v, err: %v", conversationLog, err)
	}
}
