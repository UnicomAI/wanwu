package agent_message_processor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	agent_chat_builder "github.com/UnicomAI/wanwu/internal/agent-service/service/agent-chat-builder"
	"github.com/UnicomAI/wanwu/pkg/log"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/cloudwego/eino/adk"
)

//一次enio 智能体的返回包含多个event，一个event包含多个message, event 可能是流式的可能是同步的

// WanWuAgentChatRespLineProcessor 万悟对话输出的行处理器
var WanWuAgentChatRespLineProcessor = agentChatRespLineProcessor

// EnioAgentEventIteratorReader enio 返回的智能体数据处理器，将enio返回的iter处理成和万悟和端约定的数据结构
func EnioAgentEventIteratorReader(iter *adk.AsyncIterator[*adk.AgentEvent], respContext *response.AgentChatRespContext, req *request.AgentChatContext) *safe_go_util.IteratorReader[*adk.AgentEvent, string] {
	//event读取器
	var reader = func(ctx context.Context) safe_go_util.IteratorReaderResponse[*adk.AgentEvent, string] {
		event, ok := iter.Next()
		if !ok {
			return safe_go_util.IteratorResponseStop[*adk.AgentEvent, string]()
		}
		if event.Err != nil {
			log.Errorf("agent event result error %v", event.Err)
			errResp := &safe_go_util.IteratorError[string]{Err: event.Err, ErrMsg: response.AgentChatFailResp(req.AgentChatReq.DetailId, event.Err, respContext), OutputMessage: true}
			return safe_go_util.IteratorResponseErr[*adk.AgentEvent, string](errResp)
		}
		return safe_go_util.IteratorReaderResponse[*adk.AgentEvent, string]{Data: event}
	}
	// event数据处理器
	var processor = func(ctx context.Context, data *adk.AgentEvent, rawCh chan string) ([]string, *safe_go_util.IteratorError[string]) {
		output := data.Output.MessageOutput
		var err error
		if output.IsStreaming {
			stream := output.MessageStream
			var closer = func(ctx context.Context) {
				stream.Close()
			}
			err = safe_go_util.SafeCycleReceiveByIter(ctx, EnioStreamEventMessageIReader(stream, respContext, req), rawCh, closer)
		} else {
			err = safe_go_util.SafeCycleReceiveByIter(ctx, EnioEventMessageIReader(output.Message, respContext, req), rawCh, nil)
		}
		if err != nil {
			log.Errorf("agent stream receive error %v", data.Err)
		}
		return nil, nil
	}
	return &safe_go_util.IteratorReader[*adk.AgentEvent, string]{
		Reader:    reader,
		Processor: processor,
	}
}

func EnioStreamEventMessageIReader(stream adk.MessageStream, respContext *response.AgentChatRespContext, req *request.AgentChatContext) *safe_go_util.IteratorReader[adk.Message, string] {
	var reader = func(ctx context.Context) safe_go_util.IteratorReaderResponse[adk.Message, string] {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return safe_go_util.IteratorResponseStop[adk.Message, string]()
			}
			errResp := &safe_go_util.IteratorError[string]{Err: err, OutputMessage: false}
			return safe_go_util.IteratorResponseErr[adk.Message, string](errResp)
		}
		return safe_go_util.IteratorReaderResponse[adk.Message, string]{Data: msg}
	}

	return &safe_go_util.IteratorReader[adk.Message, string]{
		Reader:    reader,
		Processor: EnioMessageProcessor(respContext, req),
	}
}

func EnioEventMessageIReader(msg adk.Message, respContext *response.AgentChatRespContext, req *request.AgentChatContext) *safe_go_util.IteratorReader[adk.Message, string] {
	var reader = func(ctx context.Context) safe_go_util.IteratorReaderResponse[adk.Message, string] {
		return safe_go_util.IteratorResponseDataStop[adk.Message, string](msg)
	}
	return &safe_go_util.IteratorReader[adk.Message, string]{
		Reader:    reader,
		Processor: EnioMessageProcessor(respContext, req),
	}
}

func EnioMessageProcessor(respContext *response.AgentChatRespContext, req *request.AgentChatContext) func(ctx context.Context, data adk.Message, rawCh chan string) ([]string, *safe_go_util.IteratorError[string]) {
	return func(ctx context.Context, data adk.Message, rawCh chan string) ([]string, *safe_go_util.IteratorError[string]) {
		messageJSON, _ := json.Marshal(data)
		log.Infof("enio message %v", string(messageJSON))
		respList, err := agent_chat_builder.BuildChatMessage(req, respContext, data)
		if err != nil {
			log.Errorf("MessageOutput error %v", err)
			return nil, &safe_go_util.IteratorError[string]{
				Err:    err,
				ErrMsg: "",
			}
		}
		return respList, nil
	}
}

// agentChatRespLineProcessor 构造对话结果行处理器
func agentChatRespLineProcessor() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		if strings.HasPrefix(lineText, "error:") {
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			//return lineText + "\n", false, nil
			return lineText, false, nil
		}
		//return lineText + "\n", false, nil
		return lineText, false, nil
	}
}
