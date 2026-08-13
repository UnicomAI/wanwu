package message_builder

import (
	"context"

	rag_manage_service "github.com/UnicomAI/wanwu/internal/rag-service/service/rag-manage-service"
	"github.com/UnicomAI/wanwu/pkg/log"
)

type KnowledgeContentBuilder struct {
}

func (KnowledgeContentBuilder) MessageType() RagMessageType {
	return KnowledgeContent
}

func (KnowledgeContentBuilder) Build(ctx context.Context, ragContext *RagContext) *RagEvent {
	//  请求rag
	buildParams, err := rag_manage_service.BuildChatConsultParams(ragContext.Req, ragContext.Rag, ragContext.KnowledgeIDToName, ragContext.KnowledgeIds)
	if err != nil {
		log.Errorf("errk = %s", err.Error())
		return &RagEvent{
			Error: err,
		}
	}
	ragContext.MarkChatStart()
	chatChan, err := rag_manage_service.RagStreamChat(ctx, ragContext.Rag.UserID, buildParams)
	if err != nil {
		log.Errorf("errk = %s", err.Error())
		return &RagEvent{
			Error: err,
		}
	}
	return &RagEvent{
		Streaming:     true,
		Stop:          true,
		StreamMessage: chatChan,
	}
}
