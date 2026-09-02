package middleware

import (
	"time"

	"github.com/UnicomAI/wanwu/api/proto/common"
	"github.com/gin-gonic/gin"
)

const (
	GeneralAgentChatApi          = "/v1/general/agent/conversation/chat"
	KnowledgeDocImportApi        = "/v1/knowledge/doc/import"         //知识库文档导入：异步任务+RAG解析耗时长，trace_user 需长 TTL 才能让 OCR/ASR 回调统计命中
	KnowledgeDocReImportApi      = "/v1/knowledge/doc/reimport"       //知识库文档重新解析
	OpenapiKnowledgeDocImportApi = "/openapi/v1/knowledge/doc/import" //openapi 文档导入
	TraceUserTimeout             = 30 * time.Minute
	TraceUserGeneralAgentTimeout = 6 * time.Hour //6个小时，理论上可以涵盖所有通用智能体会话的超时，如果发现bad case 提高这个值
)

var longTimeoutPathMap = map[string]bool{
	GeneralAgentChatApi:          true,
	KnowledgeDocImportApi:        true,
	KnowledgeDocReImportApi:      true,
	OpenapiKnowledgeDocImportApi: true,
}

// buildKeyTimeout 构造redis key 的过期时间
func buildKeyTimeout(ctx *gin.Context) time.Duration {
	var expiration = TraceUserTimeout
	if longTimeoutPathMap[ctx.Request.URL.Path] {
		expiration = TraceUserGeneralAgentTimeout
	}
	return expiration
}

// buildTraceInfo 构建追踪信息。
// TraceApp.AppId/AppType 由 mergeGinTraceStatistic 从 TraceExtra.moduleResource* 同步，此处不预填。
func buildTraceInfo(ctx *gin.Context, traceID string) *common.TraceInfo {
	userID, orgID, _ := getUserInfo(ctx)
	return &common.TraceInfo{
		TraceId: traceID,
		TraceUser: &common.TraceUser{
			UserId: userID,
			OrgId:  orgID,
		},
		TraceApp: &common.TraceApp{},
		TraceApi: &common.TraceApi{
			// 与 APIKeyRecord 的 methodPath 格式对齐：METHOD-/path（如 POST-/v1/chat/completions）
			ApiPath: ctx.Request.Method + "-" + ctx.Request.URL.Path,
		},
	}
}

// getUserInfo 获取用户信息
func getUserInfo(ctx *gin.Context) (userID, orgID string, err error) {
	// userID
	userID, err = getUserID(ctx)
	if err != nil {
		return "", "", err
	}

	// orgID
	orgID, err = getOrgID(ctx)
	if err != nil {
		return "", "", err
	}
	return userID, orgID, nil
}
