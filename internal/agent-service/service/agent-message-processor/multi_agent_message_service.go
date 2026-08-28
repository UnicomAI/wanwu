package agent_message_processor

import (
	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	"github.com/cloudwego/eino/adk"
	"github.com/gin-gonic/gin"
)

// MultiAgentMessage 多智能体消息处理
func MultiAgentMessage(ctx *gin.Context, iter *adk.AsyncIterator[*adk.AgentEvent], req *request.AgentChatContext) error {
	respContext := response.NewAgentChatRespContext(true, req.Order, req.AgentChatReq)
	//1.读取enio结果
	rawCh := safe_go_util.SafeChannelReceiveByIter(ctx, EnioAgentEventIteratorReader(iter, respContext, req))
	//2.流式返回结果
	return sseWriter(ctx, req).WriteStream(rawCh, nil, WanWuAgentChatRespLineProcessor(), nil)
}
