package agent_message_processor

import (
	"fmt"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/cloudwego/eino/adk"
	"github.com/gin-gonic/gin"
)

// AgentMessage 智能体消息处理
func AgentMessage(ctx *gin.Context, iter *adk.AsyncIterator[*adk.AgentEvent], req *request.AgentChatContext) (*response.AgentChatRespContext, error) {
	respContext := response.NewAgentChatRespContext(false, req.Order, req.AgentChatReq)
	//1.读取enio结果
	rawCh := safe_go_util.SafeChannelReceiveByIter(ctx, EnioAgentEventIteratorReader(iter, respContext, req))
	//2.流式返回结果
	err := sseWriter(ctx, req).WriteStream(rawCh, nil, WanWuAgentChatRespLineProcessor(), nil)
	return respContext, err
}

// sseWriter 根据请求构造sse写入器
func sseWriter(ctx *gin.Context, req *request.AgentChatContext) *sse_util.SSEWriter[string] {
	return sse_util.NewSSEWriter(ctx, sseLogLabel(req), sse_util.DONE_EMPTY)
}

// sseLogLabel sse 输出日志标签
func sseLogLabel(req *request.AgentChatContext) string {
	return fmt.Sprintf("[Agent] %v ", req.AgentChatReq.Input)
}
