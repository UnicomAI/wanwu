package message_builder

import (
	"context"
	"encoding/json"

	rag_manage_service "github.com/UnicomAI/wanwu/internal/rag-service/service/rag-manage-service"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// qaNoAnswerOutput 问答库未命中且没配知识库时的兜底答复
const qaNoAnswerOutput = "根据已知信息，无法回答您的问题。"

type QAFinishBuilder struct {
}

func (QAFinishBuilder) MessageType() RagMessageType {
	return QAFinish
}

func (QAFinishBuilder) Build(ctx context.Context, ragContext *RagContext) *RagEvent {
	if len(ragContext.QAIds) == 0 {
		return &RagEvent{Skip: true}
	}
	search, err := qaSearch(ctx, ragContext)
	if err != nil {
		log.Errorf("RagQASearch error %s", err)
		return buildQAErrorEvent(ragContext, err)
	}
	if len(search.Data.SearchList) > 0 {
		return buildQAHitEvent(ragContext, search)
	}
	// 未命中：配了知识库就转过去，没配就在这里兜底答复
	if len(ragContext.KnowledgeIds) == 0 {
		return buildQANoAnswerEvent(ragContext)
	}
	return &RagEvent{Skip: true}
}

func qaSearch(ctx context.Context, ragContext *RagContext) (*rag_manage_service.RagKnowledgeHitResp, error) {
	params, err := rag_manage_service.BuildQaHitParams(ragContext.Req, ragContext.Rag, ragContext.KnowledgeIDToName, ragContext.QAIds)
	if err != nil {
		return nil, err
	}
	ragContext.MarkChatStart()
	return rag_manage_service.RagQASearch(ctx, params)
}

// buildQAErrorEvent 问答库检索失败（模型不可用、参数构造失败等）
func buildQAErrorEvent(ragContext *RagContext, searchErr error) *RagEvent {
	errLine, err := marshalRagMessage(&RagMessage{
		Code:       0,
		Message:    "success",
		MsgId:      ragContext.MessageId,
		MsgType:    QAError,
		History:    make([]*RagHistory, 0),
		Data:       emptyRagData(),
		ErrMessage: searchErr.Error(),
	})
	if err != nil {
		return &RagEvent{Error: err}
	}
	if len(ragContext.KnowledgeIds) > 0 {
		return &RagEvent{Message: []string{errLine}}
	}
	noAnswer := buildQANoAnswerEvent(ragContext)
	if noAnswer.Error != nil {
		return noAnswer
	}
	return &RagEvent{
		Stop:    true,
		Message: append([]string{errLine}, noAnswer.Message...),
	}
}

// buildQAHitEvent 命中问答库，直接把答案作为终帧返回，不再走知识库
func buildQAHitEvent(ragContext *RagContext, search *rag_manage_service.RagKnowledgeHitResp) *RagEvent {
	list := search.Data.SearchList
	scoreList := search.Data.Score
	scoreLen := len(scoreList)
	resultSearchList := make([]interface{}, 0, len(list))
	for index, item := range list {
		if index < scoreLen {
			item.Score = scoreList[index]
		}
		resultSearchList = append(resultSearchList, item)
	}
	data, err := marshalRagMessage(&RagMessage{
		Code:    0,
		Message: "success",
		MsgId:   ragContext.MessageId,
		MsgType: QAFinish,
		History: make([]*RagHistory, 0),
		Data: &RagData{
			Output:     list[0].Answer,
			SearchList: resultSearchList,
		},
		Finish: 1,
	})
	if err != nil {
		return &RagEvent{Error: err}
	}
	return &RagEvent{
		Stop:    true,
		Message: []string{data},
	}
}

func buildQANoAnswerEvent(ragContext *RagContext) *RagEvent {
	data, err := marshalRagMessage(&RagMessage{
		Code:    0,
		Message: "success",
		MsgId:   ragContext.MessageId,
		MsgType: QAFinish,
		History: make([]*RagHistory, 0),
		Data: &RagData{
			Output:     qaNoAnswerOutput,
			SearchList: make([]interface{}, 0),
		},
		Finish: 1,
	})
	if err != nil {
		return &RagEvent{Error: err}
	}
	return &RagEvent{
		Stop:    true,
		Message: []string{data},
	}
}

func emptyRagData() *RagData {
	return &RagData{
		Output:     "",
		SearchList: make([]interface{}, 0),
	}
}

func marshalRagMessage(ragMessage *RagMessage) (string, error) {
	data, err := json.Marshal(ragMessage)
	if err != nil {
		return "", err
	}
	return LineData(data), nil
}
