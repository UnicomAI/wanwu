package service

import (
	"context"
	"encoding/json"
	"time"

	ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
)

// --- AG-UI 事件名常量（RAG 专属）---
const (
	// EventNameRagSearchList CUSTOM 事件名：知识库检索命中列表
	EventNameRagSearchList = "rag_search_list"
	// EventNameRagKnowledgeStart CUSTOM 事件名：即将进入知识库检索（前端据此来创建"知识库检索"卡片）
	EventNameRagKnowledgeStart = "rag_knowledge_start"
	// EventNameRagQAStart CUSTOM 事件名：即将进入问答库检索（前端据此来创建"问答库检索"卡片）
	EventNameRagQAStart = "rag_qa_start"
	// EventNameRagQASearchList CUSTOM 事件名：问答库检索结果列表（与 KB 分开，前端独立渲染 QA 卡片）。
	// 即使命中为空也会发一次（payload=[]），以便前端把"问答库检索"卡片从 running 切到 done（未命中）。
	EventNameRagQASearchList = "rag_qa_search_list"
)

// --- rag-service SSE msg_type 常量（对应 rag-service 的 RagMessageType）---
const (
	ragMsgTypeQAStart        = "qa_start"
	ragMsgTypeQAFinish       = "qa_finish"
	ragMsgTypeKnowledgeStart = "knowledge_start"
)

// --- RAG RUN_ERROR 错误码 ---
// 前端通过该 code 查 vue-i18n 文案。新增码时须同步更新
// web/src/mixins/sseMethod.js 的 RAG_ERROR_CODE_I18N 映射表。
const (
	RagErrCodeSensitiveBlock = "sensitive_block" // 上游 finish=2：敏感词拦截
	RagErrCodeUpstream       = "upstream_error"  // 上游返回非零业务错误码
	RagErrCodeUnknown        = "unknown_error"   // 未分类错误（兜底）
)

// ragStreamConverter 封装 rag-service 原始 SSE → AG-UI 事件的转换器。
//
// 拆分原因：原来 convertRagStream2AGUIEvents 把 "goroutine 骨架 / 事件发射 /
// 首 token 埋点 / chunk 业务分支" 四种关注点揉在 150 行单函数里，嵌 6 层缩进。
// 拆成 struct 后每个方法 5–15 行，handleChunk 可独立单测。
//
// 转换规则：
//  1. 首条事件：RUN_STARTED
//  2. msg_type=qa_start 状态帧：CUSTOM(rag_qa_start) — 通知前端懒创建"问答库检索"卡片
//  3. msg_type=qa_finish：CUSTOM(rag_qa_search_list)（命中/未命中都发，未命中 payload=[]）+ 透传 output
//     —— 不走通用 searchList 分支，QA 结果与 KB 结果在前端独立渲染
//  4. msg_type=knowledge_start 状态帧：CUSTOM(rag_knowledge_start) — 通知前端懒创建"知识库检索"卡片
//  5. 收到第一个非空 searchList（knowledge_content）时：CUSTOM(rag_search_list) — 必须在任何文字事件之前
//  6. reasoning_content 非空：REASONING_MESSAGE_START / CONTENT（逐 token）
//  7. output 非空（reasoning 阶段已结束）：REASONING_MESSAGE_END + TEXT_MESSAGE_START / CONTENT
//  8. finish=1（终止帧）：BaseState.FinishBase —— 自动关闭所有开放消息 + RUN_FINISHED
//  9. 错误 / finish=2（敏感词拦截）：RUN_ERROR（带 code 字段供前端 i18n 查表）
//
// 事件序列化由调用方通过 ag_ui_util.EventsToJSONChannel 完成。
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
	// hasFinalized 标记是否已发出 RUN_FINISHED 或 RUN_ERROR；用于在上游 channel 异常关闭时
	// 兜底补发 RUN_ERROR，避免前端收到无收尾事件的 SSE 流（症状：前端一直 loading）。
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
	c.emit(c.state.FinishBase()...)
	c.hasFinalized = true
}

// recordTTFT 记录首 token 延迟。
// 口径：首个"生成内容" token（reasoning_content 或 output 首字符），
// 不包含连接延迟与检索延迟——业界通行的 TTFT 定义。
// 注：此口径与旧版（首条 SSE 帧即记录）有差异，旧版会把检索时延算进来。
func (c *ragStreamConverter) recordTTFT() {
	if c.streamParams.hasRecorded {
		return
	}
	c.streamParams.firstTokenLatency = time.Since(c.streamParams.startTime).Milliseconds()
	c.streamParams.hasRecorded = true
	if c.streamParams.ctx != nil {
		c.streamParams.ctx.Set(gin_util.FIRST_RESP_LATENCY, c.streamParams.firstTokenLatency)
	}
}

// handleChunk 处理单条 chunk。返回 true 表示流已终止（error / finish=1/2），调用方应退出循环。
func (c *ragStreamConverter) handleChunk(chunk ragChunkData) (done bool) {
	// 非零 code（0外）视为错误
	if chunk.Code != 0 {
		c.streamParams.hasErr = true
		code := RagErrCodeUnknown
		if chunk.Code == ragChunkCodeBusinessError {
			code = RagErrCodeUpstream
		}
		c.finalizeError(code, chunk.Message)
		return true
	}

	// finish=2：敏感词拦截
	if chunk.Finish == 2 {
		c.streamParams.hasErr = true
		c.finalizeError(RagErrCodeSensitiveBlock, "Content blocked by sensitive word filter")
		return true
	}

	// 状态帧（Data 通常为空）：通知前端懒创建对应检索卡片。
	// 必须在 Data==nil 短路之前处理。
	switch chunk.MsgType {
	case ragMsgTypeQAStart:
		c.emitQAStartOnce()
	case ragMsgTypeKnowledgeStart:
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

	// qa_finish（问答库检索结束）：独立发 QA 搜索列表（未命中 payload=[]，前端据此把卡片从 running 切到 done），
	// 然后透传 output（命中则 output 是答案，未命中且无 KB 时 output 是"无法回答"兜底文案）。
	// 不走通用 searchList 分支，避免 QA 结果混入 KB 的 rag_search_list 事件。
	// 错误路径（chunk.Code 非零 / finish=2）在上方已处理，QA 阶段的错误仍会正常透出。
	if chunk.MsgType == ragMsgTypeQAFinish {
		c.emitQASearchListOnce(chunk.Data.SearchList)
		c.emitOutput(chunk.Data.Output)
		if chunk.Finish == 1 {
			c.emit(c.state.FinishBase()...)
			return true
		}
		return false
	}

	c.emitSearchListOnce(chunk.Data.SearchList)
	c.emitReasoning(chunk.Data.ReasoningContent)
	c.emitOutput(chunk.Data.Output)

	if chunk.Finish == 1 {
		c.emit(c.state.FinishBase()...)
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

// emitQAStartOnce 在首次收到 qa_start 状态帧时发 CUSTOM 事件，通知前端创建"问答库检索"卡片。
func (c *ragStreamConverter) emitQAStartOnce() {
	if c.hasSentQAStart {
		return
	}
	c.emit(aguievents.NewCustomEvent(EventNameRagQAStart,
		aguievents.WithValue(json.RawMessage("null"))))
	c.hasSentQAStart = true
}

// emitQASearchListOnce 在首次收到 qa_finish 时发 QA 搜索列表事件。
// 与 emitSearchListOnce 不同：空数组也发（payload=[]），让前端把"问答库检索"卡片从 running 切到 done（未命中态）。
func (c *ragStreamConverter) emitQASearchListOnce(raw json.RawMessage) {
	if c.hasSentQASearchList {
		return
	}
	// 非空时复用 KB 端的富化逻辑（补 user_kb_name）；空/解析失败回落为空数组。
	payload := enrichSearchListWithUserKbName(raw, c.kbNameMap)
	c.emit(aguievents.NewCustomEvent(EventNameRagQASearchList,
		aguievents.WithValue(payload)))
	c.hasSentQASearchList = true
}

// emitSearchListOnce 在首次收到非空 searchList 时发 CUSTOM 事件。
// 用 raw JSON 长度快速过滤空数组（"[]" 只有 2 字节），避免反序列化两次。
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
