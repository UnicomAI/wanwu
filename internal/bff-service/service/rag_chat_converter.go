package service

import (
	"context"
	"encoding/json"

	ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
)

// --- AG-UI 事件名常量（RAG 专属）---
const (
	EventNameRagSearchList     = "rag_search_list"     // CUSTOM 事件名：知识库检索命中列表
	EventNameRagKnowledgeStart = "rag_knowledge_start" // CUSTOM 事件名：即将进入知识库检索
	EventNameRagQAStart        = "rag_qa_start"        // CUSTOM 事件名：即将进入问答库检索
	EventNameRagQASearchList   = "rag_qa_search_list"  // CUSTOM 事件名：问答库检索结果列表
	EventNameRagQAEnd          = "rag_qa_end"          // CUSTOM 事件名：问答库检索结束
	EventNameRagKnowledgeEnd   = "rag_knowledge_end"   // CUSTOM 事件名：知识库检索结束
	EventNameRagQAError        = "rag_qa_error"        // CUSTOM 事件名：问答库检索报错
)

// --- rag-service SSE msg_type 常量（对应 rag-service 的 RagMessageType）---
const (
	ragMsgTypeQAStart        = "qa_start"
	ragMsgTypeQAFinish       = "qa_finish"
	ragMsgTypeQAError        = "qa_error"
	ragMsgTypeKnowledgeStart = "knowledge_start"
)

// --- RAG RUN_ERROR 错误码 ---
const (
	RagErrCodeSensitiveBlock = "sensitive_block" // 上游 finish=2：敏感词拦截
	RagErrCodeUpstream       = "upstream_error"  // 上游返回非零业务错误码
	RagErrCodeUnknown        = "unknown_error"   // 未分类错误（兜底）
)

// ragStreamConverter 把 rag-service 原始 SSE 帧转成 AG-UI 事件
type ragStreamConverter struct {
	ctx                   context.Context
	out                   chan<- aguievents.Event
	state                 *ag_ui_util.BaseState
	runID                 string
	streamParams          *ragChatStreamParams
	kbNameMap             map[string]string
	hasSentSearchList     bool
	hasSentQASearchList   bool
	hasSentKnowledgeStart bool
	hasSentQAStart        bool
	hasSentQAError        bool
	// 两个检索阶段的结束事件是否已发
	hasSentQAEnd        bool
	hasSentKnowledgeEnd bool
	// hasFinalized 标记是否已发出 RUN_FINISHED 或 RUN_ERROR
	hasFinalized bool
}

// emit 将一批事件写入输出 channel；ctx 取消时返回 false 让调用方提前退出。
func (c *ragStreamConverter) emit(events ...aguievents.Event) bool {
	for _, evt := range events {
		select {
		case c.out <- evt:
		case <-c.ctx.Done():
			return false
		}
	}
	return true
}

// finalizeError 关闭所有活跃消息后发 RUN_ERROR（不发 RUN_FINISHED）。
// 幂等：若已 finalize（RUN_FINISHED/RUN_ERROR 已发）则直接返回，避免重复事件。
func (c *ragStreamConverter) finalizeError(code, msg string) {
	if c.hasFinalized {
		return
	}
	if msg == "" {
		msg = code // 兜底：至少让 Message 非空满足协议 Validate
	}
	c.emitPendingPhaseEnds()
	c.emit(c.state.EnsureRunStarted()...)
	c.emit(c.state.EndAll()...)
	c.emit(aguievents.NewRunErrorEvent(msg,
		aguievents.WithErrorCode(code),
		aguievents.WithRunID(c.runID)))
	c.hasFinalized = true
}

// finalizeSuccess 正常收尾：发 RUN_FINISHED（经由 BaseState.FinishBase，会自动关闭所有开放消息）。
func (c *ragStreamConverter) finalizeSuccess() {
	if c.hasFinalized {
		return
	}
	c.emitPendingPhaseEnds()
	c.emit(c.state.FinishBase()...)
	c.hasFinalized = true
}

// recordTTFT 记录首 token 延迟。
// 口径：首个"生成内容" token（reasoning_content 或 output 首字符），
// 不包含连接延迟与检索延迟——业界通行的 TTFT 定义。
// 注：此口径与旧版（首条 SSE 帧即记录）有差异，旧版会把检索时延算进来。
func (c *ragStreamConverter) recordTTFT() {
	c.streamParams.recordFirstToken()
}

// handleChunk 处理单条 chunk。返回 true 表示流已终止（error / finish=1/2），调用方应退出循环。
func (c *ragStreamConverter) handleChunk(chunk ragChunkData) (done bool) {
	// 非零 code（0外）视为错误
	if chunk.Code != 0 {
		c.streamParams.errMsg = chunk.Message
		code := RagErrCodeUnknown
		if chunk.Code == ragChunkCodeBusinessError {
			code = RagErrCodeUpstream
		}
		c.finalizeError(code, chunk.Message)
		return true
	}

	// finish=2：敏感词拦截
	if chunk.Finish == 2 {
		c.streamParams.errMsg = "Content blocked by sensitive word filter"
		c.finalizeError(RagErrCodeSensitiveBlock, "Content blocked by sensitive word filter")
		return true
	}

	// 状态帧（Data 通常为空）：通知前端懒创建对应检索卡片。
	// 必须在 Data==nil 短路之前处理。
	switch chunk.MsgType {
	case ragMsgTypeQAStart:
		c.emitQAStartOnce()
	case ragMsgTypeQAError:
		c.emitQAStartOnce()
		c.emitQAError(chunk.ErrMessage)
		c.emitQAEndOnce()
	case ragMsgTypeKnowledgeStart:
		// 未命中转知识库时上游不发 qa_finish，问答库检索就结束在这一刻
		c.emitQAEndOnce()
		c.emitKnowledgeStartOnce()
	}

	// 纯状态帧：Data 为 nil 时只看 finish 决定是否终止
	if chunk.Data == nil {
		if chunk.Finish == 1 {
			c.finalizeSuccess()
			return true
		}
		return false
	}

	if chunk.MsgType == ragMsgTypeQAFinish {
		c.emitQASearchListOnce(chunk.Data.SearchList)
		c.emitQAEndOnce()
		c.emitOutput(chunk.Data.Output)
		if chunk.Finish == 1 {
			c.finalizeSuccess()
			return true
		}
		return false
	}

	c.emitSearchListOnce(chunk.Data.SearchList)
	if c.hasSentSearchList {
		c.emitKnowledgeEndOnce() // 首个非空 searchList 即知识库检索结束
	}
	if chunk.Data.Output != "" || chunk.Data.ReasoningContent != "" {
		c.emitQAEndOnce()
		c.emitKnowledgeEndOnce()
	}
	c.emitReasoning(chunk.Data.ReasoningContent)
	c.emitOutput(chunk.Data.Output)

	if chunk.Finish == 1 {
		c.finalizeSuccess()
		return true
	}
	return false
}

// emitKnowledgeStartOnce 在首次收到 knowledge_start 状态帧时发 CUSTOM 事件。
// 幂等：即使后端重复发也只发一次，避免前端重复创建卡片。
func (c *ragStreamConverter) emitKnowledgeStartOnce() {
	if c.hasSentKnowledgeStart {
		return
	}
	c.emit(aguievents.NewCustomEvent(EventNameRagKnowledgeStart,
		aguievents.WithValue(json.RawMessage("null"))))
	c.hasSentKnowledgeStart = true
}

// emitQAError 透出问答库检索失败原因
func (c *ragStreamConverter) emitQAError(message string) {
	c.emit(aguievents.NewCustomEvent(EventNameRagQAError,
		aguievents.WithValue(message)))
	c.hasSentQAError = true
}

// emitQAEndOnce 问答库检索结束
func (c *ragStreamConverter) emitQAEndOnce() {
	if c.hasSentQAEnd || !c.hasSentQAStart {
		return
	}
	c.emit(aguievents.NewCustomEvent(EventNameRagQAEnd,
		aguievents.WithValue(json.RawMessage("null"))))
	c.hasSentQAEnd = true
}

// emitKnowledgeEndOnce 知识库检索结束，正常路径是首个非空 searchList 到达时。
func (c *ragStreamConverter) emitKnowledgeEndOnce() {
	if c.hasSentKnowledgeEnd || !c.hasSentKnowledgeStart {
		return
	}
	c.emit(aguievents.NewCustomEvent(EventNameRagKnowledgeEnd,
		aguievents.WithValue(json.RawMessage("null"))))
	c.hasSentKnowledgeEnd = true
}

// emitPendingPhaseEnds 收尾前给已开始却没收到结束帧的检索阶段补发 end
func (c *ragStreamConverter) emitPendingPhaseEnds() {
	c.emitQAEndOnce()
	c.emitKnowledgeEndOnce()
}

// emitQAStartOnce 在首次收到 qa_start 状态帧时发 CUSTOM 事件，通知前端创建"问答库检索"卡片。
func (c *ragStreamConverter) emitQAStartOnce() {
	if c.hasSentQAStart {
		return
	}
	c.emit(aguievents.NewCustomEvent(EventNameRagQAStart,
		aguievents.WithValue(json.RawMessage("null"))))
	c.hasSentQAStart = true
}

// emitQASearchListOnce 在首次收到 qa_finish 时发 QA 搜索列表事件
func (c *ragStreamConverter) emitQASearchListOnce(raw json.RawMessage) {
	if c.hasSentQASearchList {
		return
	}
	if c.hasSentQAError && len(raw) <= 2 {
		c.hasSentQASearchList = true
		return
	}
	// 非空时补 user_kb_name；空/解析失败回落为空数组
	payload := enrichSearchListWithUserKbName(raw, c.kbNameMap)
	c.emit(aguievents.NewCustomEvent(EventNameRagQASearchList,
		aguievents.WithValue(payload)))
	c.hasSentQASearchList = true
}

// emitSearchListOnce 在首次收到非空 searchList 时发 CUSTOM 事件
func (c *ragStreamConverter) emitSearchListOnce(raw json.RawMessage) {
	if c.hasSentSearchList || len(raw) <= 2 {
		return
	}
	payload := enrichSearchListWithUserKbName(raw, c.kbNameMap)
	if len(payload) == 0 {
		return
	}
	c.emit(aguievents.NewCustomEvent(EventNameRagSearchList,
		aguievents.WithValue(payload)))
	c.hasSentSearchList = true
}

// emitReasoning 发推理内容事件（若非空）。
func (c *ragStreamConverter) emitReasoning(reasoning string) {
	if reasoning == "" {
		return
	}
	c.recordTTFT()
	c.emit(c.state.StartReasoningMessage()...)
	c.emit(aguievents.NewReasoningMessageContentEvent(
		c.state.ReasoningMessageID(), reasoning))
}

// emitOutput 发正文内容事件（若非空）；首次 output 到达即视为 reasoning 阶段结束。
func (c *ragStreamConverter) emitOutput(output string) {
	if output == "" {
		return
	}
	c.recordTTFT()
	c.emit(c.state.EndReasoningMessage()...)
	c.emit(c.state.StartTextMessage()...)
	c.emit(aguievents.NewTextMessageContentEvent(
		c.state.MessageID(), output))
}
