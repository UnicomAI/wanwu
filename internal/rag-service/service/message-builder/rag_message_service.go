package message_builder

import (
	"context"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"google.golang.org/grpc"
)

type RagMessageType string

const (
	QAStart          RagMessageType = "qa_start"
	QAFinish         RagMessageType = "qa_finish"
	QAError          RagMessageType = "qa_error"
	KnowledgeStart   RagMessageType = "knowledge_start"
	KnowledgeContent RagMessageType = "knowledge_content"
)

var ragMessageProcessChain = []RagMessageBuilder{
	QAStartBuilder{},
	QAFinishBuilder{},
	KnowledgeStartBuilder{},
	KnowledgeContentBuilder{},
}

type RagContext struct {
	MessageId string
	// chatStartTime 统计耗时的基准，由真正请求 rag-wanwu 的 builder 打点
	chatStartTime     time.Time
	Req               *rag_service.ChatRagReq
	Rag               *model.RagInfo
	KnowledgeIDToName map[string]string
	KnowledgeIds      []string
	QAIds             []string
}

// MarkChatStart 在向 rag-wanwu 发起请求前调用
func (c *RagContext) MarkChatStart() {
	c.chatStartTime = time.Now()
}

type RagEvent struct {
	Skip          bool
	Stop          bool
	Streaming     bool
	Message       []string      //当非流式的时候读这个
	StreamMessage <-chan string // 单流式的时候读这个
	Error         error
}

type RagMessageBuilder interface {
	MessageType() RagMessageType
	Build(ctx context.Context, ragContext *RagContext) *RagEvent
}

type RagMessage struct {
	Code       int            `json:"code"`
	Message    string         `json:"message"`
	MsgId      string         `json:"msg_id"`
	MsgType    RagMessageType `json:"msg_type"`
	Data       *RagData       `json:"data"`
	History    []*RagHistory  `json:"history"`
	Finish     int            `json:"finish"`
	ErrMessage string         `json:"errMessage,omitempty"`
}

type RagData struct {
	Output     string        `json:"output"`
	SearchList []interface{} `json:"searchList"`
}

type RagHistory struct {
	Query       string `json:"query"`
	Response    string `json:"response"`
	NeedHistory bool   `json:"needHistory"`
}

func BuildMessage(ctx context.Context, ragContext *RagContext, stream grpc.ServerStreamingServer[rag_service.ChatRagResp]) (err error) {
	// 兜底基准：builder 还没来得及打点就出错时用它，正常路径由 MarkChatStart 覆盖
	ragContext.chatStartTime = time.Now()
	recorder := &ragChatRecorder{}
	//无论正常结束、上游报错还是客户端断流，都落一条问答明细
	defer func() {
		recorder.save(ctx, ragContext, err)
	}()
	for _, builder := range ragMessageProcessChain {
		ragEvent := builder.Build(ctx, ragContext)
		if ragEvent.Skip {
			continue
		}
		if ragEvent.Error != nil {
			return ragEvent.Error
		}
		//发送消息
		if err = sendMessage(ragEvent, stream, recorder); err != nil {
			return err
		}
		if ragEvent.Stop {
			break
		}
	}
	return nil
}

func sendMessage(ragEvent *RagEvent, stream grpc.ServerStreamingServer[rag_service.ChatRagResp], recorder *ragChatRecorder) error {
	if ragEvent.Streaming {
		for message := range ragEvent.StreamMessage {
			if err := sendLine(message, stream, recorder); err != nil {
				return err
			}
		}
	} else {
		for _, message := range ragEvent.Message {
			if err := sendLine(message, stream, recorder); err != nil {
				return err
			}
		}
	}
	return nil
}

func sendLine(message string, stream grpc.ServerStreamingServer[rag_service.ChatRagResp], recorder *ragChatRecorder) error {
	recorder.record(message)
	if err := stream.Send(&rag_service.ChatRagResp{Content: message}); err != nil {
		return grpc_util.ErrorStatusWithKey(errs.Code_RagChatErr, "rag_chat_err", err.Error())
	}
	return nil
}

func LineData(data []byte) string {
	return "data: " + string(data) + "\n\n"
}
