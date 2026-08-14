package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/types"
	"github.com/UnicomAI/wanwu/internal/channel-service/client/model"
	"github.com/UnicomAI/wanwu/internal/channel-service/wanwu"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// handleChatflowMessage 处理对话流（chatflow）消息。
// 与 handleAgentMessage 同构（会话复用、附件两步操作、SSE 处理），差异：
//   - 附件上传走对话流专用 /chatflow/file/upload（返回 presign URL，非 file_info）。
//   - 附件 URL 填进 parameters，key=对话流开始节点声明的入参变量名（由 8999 schema 获取）。
//   - SSE 是 Coze 风格（event:/data: + content 字段 + done），非 agent 的 data:{"response":"..."}。
//
// 支持"先发附件、再发文字"的分两步操作（与 agent/WGA 链路同构）：
//   - 纯附件消息（无有效文字指令）：上传 → 存待用附件缓存 → 回提示 → 不调 chatflow。
//   - 有文字指令：drain 暂存附件 + 本条附件 → 填 parameters 发 chatflow。
func (h *Handler) handleChatflowMessage(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage) error {
	apiKey := ch.ApiKey
	wanwuClient := wanwu.NewClient(h.cfg.BFF.ApiBaseUrl)

	// 上传本条附件到对话流 minio（纯附件消息也要上传，才能存 URL 进待用缓存）
	currentAtts, err := uploadChatflowAttachments(ctx, wanwuClient, apiKey, msg.ChannelID, msg.UserID, msg.Attachments)
	if err != nil {
		return fmt.Errorf("failed to upload chatflow attachments for channel %s user %s: %w", msg.ChannelID, msg.UserID, err)
	}

	// 提取有效文字指令（复用 agent 链路逻辑：排除空白、微信占位符、等于附件名的 Content）
	text := effectiveText(msg.Content, msg.Attachments)

	if text == "" {
		// 纯附件消息：存入待用缓存 + 回提示，不调 chatflow
		for _, a := range currentAtts {
			h.attachmentCache.Append(msg.ChannelID, msg.UserID, a)
		}
		tip := fmt.Sprintf("已收到 %d 个文件，请说明要做什么（%d 分钟内有效）",
			len(currentAtts), int(pendingAttachmentTTL.Minutes()))
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); err != nil {
			log.Warnf("[Chatflow] send attachment-received tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, err)
		}
		return nil
	}

	// 取出暂存附件，与本条附件合并
	pending := h.attachmentCache.Drain(msg.ChannelID, msg.UserID)
	atts := append(pending, currentAtts...)
	if len(pending) > 0 {
		log.Infof("[Chatflow] drained %d pending attachments for channel %s user %s",
			len(pending), msg.ChannelID, msg.UserID)
	}

	// 获取或创建对话流会话（隔离 key "chatflow"）
	conversationID, ok := h.convManager.GetConversationID(ctx, msg.ChannelID, msg.UserID, "chatflow")
	if !ok {
		convResp, err := wanwuClient.CreateChatflowConversation(ctx, apiKey, &wanwu.ChatflowCreateConversationRequest{
			UUID:             ch.AppID,
			ConversationName: truncate(msg.Content, 50),
		})
		if err != nil {
			// conversation_id 是 chat 必填字段，创建会话失败不能带空值继续调 chat
			// （BFF ChatflowChat 会先用 conversation_id 查 app-service 会话记录，空值必 not found）。
			// 回发提示并返回，不再静默继续。
			tip := "对话流会话创建失败，请稍后重试"
			if sendErr := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); sendErr != nil {
				log.Warnf("[Chatflow] send create-conv-fail tip failed: channel=%s user=%s err=%v",
					msg.ChannelID, msg.UserID, sendErr)
			}
			log.Warnf("failed to create chatflow conversation for channel %s user %s: %v", msg.ChannelID, msg.UserID, err)
			return nil
		}
		conversationID = convResp.ConversationID
		h.convManager.SetConversationID(ctx, msg.ChannelID, msg.UserID, "chatflow", conversationID)
		log.Infof("created chatflow conversation %s for channel %s user %s", conversationID, msg.ChannelID, msg.UserID)
	}

	// 查开始节点入参（带缓存）：用于检测必填入参 + 填附件
	inputs, inputErr := h.getChatflowInputs(ctx, ch.ChannelID, ch.AppID, ch.OrgID, ch.UserID)

	// 文字型必填入参：走交互式追问补全（存追问状态 + 回发首个追问 + return，不调后端）
	if inputErr == nil && len(inputs) > 0 {
		if h.startChatflowInputPending(ctx, ch, msg, conversationID, text, atts, inputs) {
			return nil
		}
	}

	// 文件型必填入参但无附件 / 附件填充：buildChatflowParameters 处理（含 schema 失败降级）
	parameters, paramTip, abort := h.buildChatflowParameters(ctx, ch, atts)
	if abort {
		// 文件型必填入参但没发文件：回发提示，不调后端
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, paramTip, msg.Extra); err != nil {
			log.Warnf("[Chatflow] send required-param tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, err)
		}
		return nil
	}
	if paramTip != "" {
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, paramTip, msg.Extra); err != nil {
			log.Warnf("[Chatflow] send dropped-attachment tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, err)
		}
	}

	// 调用对话流 chat（SSE 流式）
	chatReq := &wanwu.ChatflowChatRequest{
		UUID:           ch.AppID,
		ConversationID: conversationID,
		Query:          text,
		Parameters:     parameters,
	}

	resp, err := wanwuClient.ChatWithChatflow(ctx, apiKey, chatReq)
	if err != nil {
		// chat 调用失败（网络/鉴权/后端 4xx 5xx）：回发提示，不再静默
		tip := "对话流调用失败，请稍后重试"
		if sendErr := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); sendErr != nil {
			log.Warnf("[Chatflow] send chat-fail tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, sendErr)
		}
		return fmt.Errorf("failed to call chatflow chat api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	return h.handleChatflowSSEResponse(ctx, ch, msg, resp)
}

// chatflowReservedInputs 对话流系统保留入参名：用户输入走 query（对应 USER_INPUT），
// 会话名走 CONVERSATION_NAME，均不填进 parameters，也不参与必填检测。
var chatflowReservedInputs = map[string]bool{
	"USER_INPUT": true, "CONVERSATION_NAME": true,
}

// isFileInput 启发式判断入参是否文件型（图片/文件 URL 变量），用于"无附件时"的必填检测。
// 依据：入参名或 description 含 img/image/file/photo/picture/图片/文件 等关键词。
// schema 只有 type=string，无法精确区分文件型与文字型入参，启发式不可靠但优于按顺序盲填
// （避免把图片 URL 错填进文字入参如 name/num）。
func isFileInput(in wanwu.ChatflowInput) bool {
	if chatflowReservedInputs[in.Name] {
		return false
	}
	text := strings.ToLower(in.Name + " " + in.Description)
	keywords := []string{"img", "image", "file", "photo", "picture", "图片", "文件", "图像", "照片"}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}

// isFillableInput 判断入参是否可被附件 URL 填充。
// IM 场景下用户发了图片/文件时，非系统保留的 string 入参大概率就是要吃这个 URL 的文件型入参
// （画布里 image/file 类型入参在 workflow schema 生成时被降级成 type:string，文件子类型信号丢失，
// 通道拿不到，故有附件时不再靠名字启发式，统一把非保留 string 入参当可填）。
// hasAttachments=false 时退回 isFileInput 启发式：仅用于"必填文件入参缺图"的提示判定。
func isFillableInput(in wanwu.ChatflowInput, hasAttachments bool) bool {
	if chatflowReservedInputs[in.Name] {
		return false
	}
	if hasAttachments {
		return true
	}
	return isFileInput(in)
}

// buildChatflowParameters 组装 parameters 并检测 IM 场景填不了的必填入参。
// 入参来自 8999 schema（按 channelID+uuid 缓存）。返回 (parameters, tip, abort)：
//   - abort=true：存在必填入参 IM 填不了（文字型必填 / 文件型必填但没发文件），
//     parameters=nil，tip=回发提示，调用方应回发 tip 且不调后端。
//   - abort=false：正常。parameters 可能 nil（无附件/无可填入参），tip 为附件丢弃提示（可空）。
//   - schema 获取失败：无法判断必填，按无入参降级（纯文字不受影响，附件提示无法处理）。
func (h *Handler) buildChatflowParameters(ctx context.Context, ch *model.Channel, atts []*PendingAttachment) (map[string]any, string, bool) {
	inputs, err := h.getChatflowInputs(ctx, ch.ChannelID, ch.AppID, ch.OrgID, ch.UserID)
	if err != nil {
		if len(atts) > 0 {
			return nil, fmt.Sprintf("无法获取对话流入参配置，未能处理 %d 个附件", len(atts)), false
		}
		return nil, "", false
	}

	// 检测必填入参：存在 IM 场景填不了的必填入参时返回 (tip, abort)
	if tip, abort := checkChatflowRequiredInputs(inputs, len(atts)); abort {
		return nil, tip, true
	}

	// 无需中止：附件填进文件型入参
	if len(atts) == 0 {
		return nil, "", false
	}
	params, droppedTip := fillChatflowParameters(inputs, atts)
	return params, droppedTip, false
}

// checkChatflowRequiredInputs 检测画布必填入参中 IM 场景填不了的部分（纯函数，便于测试）。
// 排除系统保留名 USER_INPUT/CONVERSATION_NAME（用户输入走 query）。
// 注意：文字型必填入参不再在此 abort，改由 startChatflowInputPending 走交互式追问补全。
// 返回 (tip, abort)：
//   - 文件型必填入参但无附件（attCount=0）：→ abort=true，tip 提示先发文件（追问无法补文件）。
//   - 有附件时：附件会填进非保留 string 入参，不在此 abort（必填文件入参能被附件满足）。
//   - 否则 abort=false（文字型必填由调用方另走追问流程）。
func checkChatflowRequiredInputs(inputs []wanwu.ChatflowInput, attCount int) (string, bool) {
	if attCount > 0 {
		// 有附件：非保留 string 必填入参都能被附件填，不 abort
		return "", false
	}
	var requiredFile []string
	for _, in := range inputs {
		if !in.Required || chatflowReservedInputs[in.Name] || !isFillableInput(in, false) {
			continue
		}
		requiredFile = append(requiredFile, in.Name)
	}
	// 文件型必填入参但用户没发文件 → 提示先发文件
	if len(requiredFile) > 0 {
		tip := fmt.Sprintf("该对话流需要文件输入 %s，请先发送图片或文件", strings.Join(requiredFile, "、"))
		return tip, true
	}
	return "", false
}

// typeHint 返回入参类型的中文提示（用于追问文案）。
func typeHint(typ string) string {
	switch typ {
	case "integer":
		return "整数"
	case "number":
		return "数字"
	case "boolean":
		return "是/否"
	default: // string / 空 type
		return "文本"
	}
}

// startChatflowInputPending 检测到必填文字型入参未填时，存追问状态并回发首个追问。
// 返回 true 表示已进入追问流程（调用方应 return，不调后端）；false 表示无需追问（无必填文字入参）。
// 有附件时：非保留 string 必填入参都会被附件 URL 填充（isFillableInput(hasAttachments=true)），
// 不在此追问；仅无附件时才收集必填文字入参走交互式追问。
func (h *Handler) startChatflowInputPending(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage,
	conversationID, originalQuery string, atts []*PendingAttachment, inputs []wanwu.ChatflowInput) bool {
	hasAttachments := len(atts) > 0
	// 收集必填文字型入参（排除系统保留名 / 可被附件填的入参）
	var pending []ChatflowPendingInput
	for _, in := range inputs {
		if !in.Required || chatflowReservedInputs[in.Name] || isFillableInput(in, hasAttachments) {
			continue
		}
		pending = append(pending, ChatflowPendingInput{Name: in.Name, Type: in.Type})
	}
	if len(pending) == 0 {
		return false
	}
	// 存追问状态（保留原始 query + 已创建会话 + 附件，补全后复用）
	h.chatflowPendingInputs.Set(msg.ChannelID, msg.UserID, &ChatflowPendingInputState{
		UUID:           ch.AppID,
		ConversationID: conversationID,
		OriginalQuery:  originalQuery,
		PendingInputs:  pending,
		FilledParams:   make(map[string]any),
		Attachments:    atts,
	})
	// 回发首个追问提示
	tip := fmt.Sprintf("请提供 %s（%s）：", pending[0].Name, typeHint(pending[0].Type))
	if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); err != nil {
		log.Warnf("[Chatflow] send input-pending tip failed: channel=%s user=%s err=%v",
			msg.ChannelID, msg.UserID, err)
	}
	log.Infof("[Chatflow] start input pending: channel=%s user=%s pending=%v query=%q",
		msg.ChannelID, msg.UserID, pending, originalQuery)
	return true
}

// convertChatflowInput 把用户回复文字按 schema type 转换。转不过去返回 error（纯函数，便于测试）。
func convertChatflowInput(s, typ string) (any, error) {
	switch typ {
	case "integer":
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err != nil {
			return nil, fmt.Errorf("需要整数")
		}
		return n, nil
	case "number":
		f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return nil, fmt.Errorf("需要数字")
		}
		return f, nil
	case "boolean":
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "true", "是", "1", "yes":
			return true, nil
		case "false", "否", "0", "no":
			return false, nil
		default:
			return nil, fmt.Errorf("需要 是/否")
		}
	default: // string / 空 type
		return s, nil
	}
}

// handleChatflowInputReply 处理对话流追问状态下用户的回复：把回复按类型转换填入当前待补入参。
// 若还有未补入参继续追问下一个；全补完则删除追问状态，用原始 query + 完整 parameters 调后端。
// 由 HandlePlatformMessage 在 appType 分发前检测到追问状态时调用。
func (h *Handler) handleChatflowInputReply(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage, ps *ChatflowPendingInputState) error {
	current := ps.PendingInputs[0]
	reply := effectiveText(msg.Content, msg.Attachments) // 用户回复的文字

	// 按类型转换
	val, err := convertChatflowInput(reply, current.Type)
	if err != nil {
		// 转换失败：重问当前入参，提示格式错误，状态保留等用户重发
		tip := fmt.Sprintf("格式错误（%v），请重新提供 %s（%s）：", err, current.Name, typeHint(current.Type))
		if sendErr := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); sendErr != nil {
			log.Warnf("[Chatflow] send input-retry tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, sendErr)
		}
		return nil
	}
	ps.FilledParams[current.Name] = val
	log.Infof("[Chatflow] input filled: channel=%s user=%s name=%s val=%v, remaining=%d",
		msg.ChannelID, msg.UserID, current.Name, val, len(ps.PendingInputs)-1)

	// 还有未补入参：追问下一个（状态已在 cache 中，sync.Map 存指针，直接改字段）
	if len(ps.PendingInputs) > 1 {
		ps.PendingInputs = ps.PendingInputs[1:]
		next := ps.PendingInputs[0]
		tip := fmt.Sprintf("请提供 %s（%s）：", next.Name, typeHint(next.Type))
		if sendErr := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); sendErr != nil {
			log.Warnf("[Chatflow] send next-input tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, sendErr)
		}
		return nil
	}

	// 全部补完：删除追问状态，用原始 query + 完整 parameters 调后端
	h.chatflowPendingInputs.Delete(msg.ChannelID, msg.UserID)

	// 合并附件（追问前已上传的，补完后一起填进文件型入参）
	parameters := ps.FilledParams
	if len(ps.Attachments) > 0 {
		if inputs, ie := h.getChatflowInputs(ctx, ch.ChannelID, ch.AppID, ch.OrgID, ch.UserID); ie == nil && len(inputs) > 0 {
			if fileParams, _ := fillChatflowParameters(inputs, ps.Attachments); fileParams != nil {
				for k, v := range fileParams {
					parameters[k] = v
				}
			}
		}
	}

	chatReq := &wanwu.ChatflowChatRequest{
		UUID:           ps.UUID,
		ConversationID: ps.ConversationID,
		Query:          ps.OriginalQuery,
		Parameters:     parameters,
	}
	wanwuClient := wanwu.NewClient(h.cfg.BFF.ApiBaseUrl)
	resp, err := wanwuClient.ChatWithChatflow(ctx, ch.ApiKey, chatReq)
	if err != nil {
		// chat 调用失败：回发提示
		tip := "对话流调用失败，请稍后重试"
		if sendErr := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); sendErr != nil {
			log.Warnf("[Chatflow] send chat-fail tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, sendErr)
		}
		return fmt.Errorf("failed to call chatflow chat api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return h.handleChatflowSSEResponse(ctx, ch, msg, resp)
}

// fillChatflowParameters 把附件 presign URL 填进对话流 parameters（纯函数，便于测试）。
// 有附件时填进所有非保留 string 入参（isFillableInput(hasAttachments=true)，必填优先——
// inputs 已按 required 优先排序）。画布里 image/file 类型入参在 workflow schema 生成时被
// 降级成 type:string（文件子类型信号丢失），通道无法靠名字区分，故有附件时统一按顺序填进
// 非保留 string 入参（如 aaa/aaa2/aaa3 多图按顺序填）。返回 (parameters, droppedTip)：
//   - 无可填入参：parameters=nil, droppedTip 提示无法处理附件。
//   - 附件数 > 可填入参数：多余附件发 droppedTip 提示。
func fillChatflowParameters(inputs []wanwu.ChatflowInput, atts []*PendingAttachment) (map[string]any, string) {
	var fillable []wanwu.ChatflowInput
	for _, in := range inputs {
		if isFillableInput(in, true) {
			fillable = append(fillable, in)
		}
	}
	if len(fillable) == 0 {
		tip := fmt.Sprintf("该对话流无文件输入参数，无法处理 %d 个附件", len(atts))
		return nil, tip
	}

	parameters := make(map[string]any)
	n := len(atts)
	if n > len(fillable) {
		n = len(fillable)
	}
	for i := 0; i < n; i++ {
		parameters[fillable[i].Name] = atts[i].URL
	}

	var droppedTip string
	if len(atts) > len(fillable) {
		var dropped []string
		for _, a := range atts[len(fillable):] {
			dropped = append(dropped, a.FileName)
		}
		droppedTip = fmt.Sprintf("该对话流仅支持 %d 个文件输入，已忽略：%s",
			len(fillable), strings.Join(dropped, "、"))
	}
	return parameters, droppedTip
}

// uploadChatflowAttachments 把附件上传到对话流 minio（/chatflow/file/upload），返回待用附件列表。
// 与 uploadAttachments 的差异：走对话流专用上传接口，返回 presign URL（对话流画布期望的格式）。
// 无附件返回 nil。
func uploadChatflowAttachments(ctx context.Context, wanwuClient *wanwu.Client, apiKey, channelID, userID string, attachments []types.Attachment) ([]*PendingAttachment, error) {
	if len(attachments) == 0 {
		return nil, nil
	}
	out := make([]*PendingAttachment, 0, len(attachments))
	for _, att := range attachments {
		url, err := wanwuClient.UploadChatflowFile(ctx, apiKey, att.Name, att.MimeType, att.Data)
		if err != nil {
			return nil, fmt.Errorf("upload chatflow attachment %s failed: %w", att.Name, err)
		}
		out = append(out, &PendingAttachment{
			URL:      url,
			FileName: att.Name,
			MimeType: att.MimeType,
		})
		log.Infof("[Chatflow] uploaded attachment %s (%d bytes) -> %s for channel %s user %s",
			att.Name, len(att.Data), url, channelID, userID)
	}
	return out, nil
}

// handleChatflowSSEResponse 处理对话流 SSE 流式响应（Coze 风格）。
// 不能复用 handleAgentSSEResponse：对话流 SSE 是 id:/event:/data: 三行一组，
// 文本在 event=conversation.message.delta/completed 的 data.content，结束信号是 event=done（非 [DONE]）。
func (h *Handler) handleChatflowSSEResponse(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage, resp *http.Response) error {
	streamSender := h.manager.CreateStreamSender(ctx, msg.ChannelID, msg.UserID, msg.Extra)

	var fullContent strings.Builder
	reader := bufio.NewReader(resp.Body)
	chunkCount := 0

	log.Infof("[ChatflowSSE] channel=%s user=%s start streaming from chatflow %s (streamSender=%v)",
		ch.ChannelID, msg.UserID, ch.AppID, streamSender != nil)

	// SSE 块由空行分隔，每个块内可能含 id:/event:/data: 多行。
	// 累积当前块的 event/data，遇到空行时处理该块。
	// processBlock 处理一个完整 SSE 块，返回 stop=true 表示流应结束。
	var curEvent, curData string
	var streamError error // 后端把错误 JSON 当 SSE data 返回时记录，流结束后回发
	processBlock := func() (stop bool) {
		defer func() { curEvent, curData = "", "" }()
		// 临时诊断：打印每个 SSE 块的 event/data 摘要
		if curEvent != "" || curData != "" {
			log.Infof("[ChatflowSSE-DEBUG] channel=%s user=%s block event=%q data=%q",
				ch.ChannelID, msg.UserID, curEvent, truncate(curData, 200))
		}
		// done 事件：显式结束信号（部分后端会发 event:done），无论有无 data 都结束流。
		if curEvent == "done" {
			return true
		}
		// completed 事件：完整回复就绪 = 自然结束信号。
		// 实测后端在 completed 之后既不发 done 也不关流，读取循环会永久阻塞在 ReadString，
		// 导致非流式路径（微信 streamSender=nil）的回复永远发不出去。
		// 故把 completed 当作结束信号：delta 缺失时用其完整正文兜底，然后结束流。
		if curEvent == "conversation.message.completed" {
			if curData == "" {
				return true
			}
			var d struct {
				Role    string `json:"role"`
				Type    string `json:"type"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal([]byte(curData), &d); err != nil {
				log.Errorf("[ChatflowSSE] channel=%s user=%s failed to parse completed data: %v, raw: %s",
					ch.ChannelID, msg.UserID, err, curData)
				return true // 解析失败也结束（completed 已代表回复就绪）
			}
			// delta 缺失（后端只发 completed 没发 delta）时用完整正文兜底；
			// 已有 delta 累积时不重复写（completed 与 delta 内容重复）。
			if d.Type == "answer" && d.Content != "" && fullContent.Len() == 0 {
				if streamSender != nil {
					if err := streamSender.SendChunk(ctx, d.Content, false); err != nil {
						log.Errorf("[ChatflowSSE] channel=%s user=%s stream sender chunk (completed) failed, falling back: %v",
							ch.ChannelID, msg.UserID, err)
						closeStreamSender(streamSender, ctx, fmt.Errorf("stream chunk failed: %w", err))
						streamSender = nil
					}
				}
				fullContent.WriteString(d.Content)
				chunkCount++
			}
			return true
		}
		if curData == "" {
			return false
		}
		// delta 事件：逐字增量，流式发卡片 + 累积。
		if curEvent == "conversation.message.delta" {
			var d struct {
				Role    string `json:"role"`
				Type    string `json:"type"`
				Content string `json:"content"`
			}
			if err := json.Unmarshal([]byte(curData), &d); err != nil {
				log.Errorf("[ChatflowSSE] channel=%s user=%s failed to parse data: %v, raw: %s",
					ch.ChannelID, msg.UserID, err, curData)
				return false
			}
			if d.Type == "answer" && d.Content != "" {
				if streamSender != nil {
					if err := streamSender.SendChunk(ctx, d.Content, false); err != nil {
						log.Errorf("[ChatflowSSE] channel=%s user=%s stream sender chunk failed, falling back: %v",
							ch.ChannelID, msg.UserID, err)
						closeStreamSender(streamSender, ctx, fmt.Errorf("stream chunk failed: %w", err))
						streamSender = nil
					}
				}
				fullContent.WriteString(d.Content)
				chunkCount++
			}
			return false
		}
		// error 事件：后端执行失败（如视觉模型 500、必填参数缺失等）。
		// data 是 {"code":"720701013","msg":"..."}，code 可能是字符串或数字。
		// 记录错误待流结束后回发，但不结束流（后端可能后续仍会关流；EOF 时统一处理）。
		if curEvent == "error" {
			captureChatflowError(curData, ch, msg, &streamError)
			return false
		}
		// 其它事件（含 event 为空）：data 可能是后端返回的错误 JSON
		// （实测后端会把 {"code":720702002,"msg":"Missing required parameters..."} 当 SSE data
		// 直接返回，无 event 前缀。原逻辑因 curEvent!="delta" 直接跳过 → 错误被吞）。
		captureChatflowError(curData, ch, msg, &streamError)
		return false
	}

	for {
		select {
		case <-ctx.Done():
			log.Infof("[ChatflowSSE] channel=%s user=%s context cancelled after %d chunks", ch.ChannelID, msg.UserID, chunkCount)
			closeStreamSender(streamSender, ctx, ctx.Err())
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Infof("[ChatflowSSE-DEBUG] channel=%s user=%s EOF after %d chunks, lastLine=%q", ch.ChannelID, msg.UserID, chunkCount, line)
				_ = processBlock()
				break
			}
			closeStreamSender(streamSender, ctx, err)
			return fmt.Errorf("error reading chatflow SSE stream: %w", err)
		}
		// 临时诊断：打印每行原始 SSE（定位流卡住/无内容/格式不匹配）
		log.Infof("[ChatflowSSE-DEBUG] channel=%s user=%s raw=%q", ch.ChannelID, msg.UserID, strings.TrimRight(line, "\r\n"))
		line = strings.TrimRight(line, "\r\n")

		if line == "" {
			// 空行：当前 SSE 块结束。completed/done 返回 stop=true 时退出循环。
			if processBlock() {
				log.Infof("[ChatflowSSE] channel=%s user=%s stream ended by terminal event after %d chunks",
					ch.ChannelID, msg.UserID, chunkCount)
				break
			}
			continue
		}
		if strings.HasPrefix(line, "event:") {
			curEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			curData = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		}
		// id: 行忽略
	}

	replyContent := fullContent.String()

	// 流式路径：发最终 chunk 标记完成
	if streamSender != nil {
		if err := streamSender.SendChunk(ctx, "", true); err != nil {
			log.Errorf("[ChatflowSSE] channel=%s user=%s stream sender final chunk failed: %v", ch.ChannelID, msg.UserID, err)
			closeStreamSender(streamSender, ctx, fmt.Errorf("final chunk failed: %w", err))
			streamSender = nil
		} else {
			closeStreamSender(streamSender, ctx, nil)
			log.Infof("[ChatflowSSE] channel=%s user=%s stream completed via card, total %d chunks, %d chars",
				ch.ChannelID, msg.UserID, chunkCount, len(replyContent))
			// 流式已发卡片但后端返回了错误（且无正文）：补发错误提示，避免用户只看到半截
			if streamError != nil && replyContent == "" {
				_ = h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, streamError.Error(), msg.Extra)
			}
			return nil
		}
	}

	// 非流式路径：完整回复发送给平台
	if replyContent == "" {
		// 后端返回了错误（无正文）：回发错误提示，不再静默"没反应"
		if streamError != nil {
			log.Warnf("[ChatflowSSE] channel=%s user=%s empty reply with backend error: %v", ch.ChannelID, msg.UserID, streamError)
			if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, streamError.Error(), msg.Extra); err != nil {
				log.Warnf("[ChatflowSSE] channel=%s user=%s send backend-error tip failed: %v", ch.ChannelID, msg.UserID, err)
			}
			return nil
		}
		log.Warnf("[ChatflowSSE] channel=%s user=%s empty reply from chatflow after %d chunks", ch.ChannelID, msg.UserID, chunkCount)
		return nil
	}
	log.Infof("[ChatflowSSE] channel=%s user=%s stream completed, total %d chunks, reply length=%d, content: %s",
		ch.ChannelID, msg.UserID, chunkCount, len(replyContent), truncate(replyContent, 200))

	// 复用 agent 链路的正文后处理：内嵌图片剥离+单独下发、UTF-8 清洗、引用标注剥离
	textToSend := replyContent
	hasInlineImage := inlineImageRe.MatchString(replyContent)
	if hasInlineImage {
		textToSend = stripInlineImages(replyContent)
	}
	textToSend = stripInvalidUTF8(textToSend)
	textToSend = stripCitations(textToSend)
	textToSend = strings.TrimSpace(textToSend)
	if textToSend != "" {
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, textToSend, msg.Extra); err != nil {
			return fmt.Errorf("failed to send reply to platform: %w", err)
		}
		log.Infof("[ChatflowSSE] channel=%s user=%s reply text sent to platform successfully", msg.ChannelID, msg.UserID)
	}
	if hasInlineImage {
		h.sendInlineImages(ctx, msg, replyContent)
	}
	return nil
}

// captureChatflowError 从 SSE data 中解析后端错误 JSON 并记录到 streamError（仅首次有效）。
// 后端错误格式 {"code":"720701013","msg":"..."}，code 可能是字符串（event:error）或数字（无 event 前缀）。
// 用 json.RawMessage 兼容两种类型；msg 非空才记录（避免误吞正常事件）。
// 已存在 streamError 时不覆盖（保留首个错误）。
func captureChatflowError(curData string, ch *model.Channel, msg *types.PlatformMessage, streamError *error) {
	if curData == "" || streamError == nil {
		return
	}
	var errObj struct {
		Code json.RawMessage `json:"code"`
		Msg  string          `json:"msg"`
	}
	if json.Unmarshal([]byte(curData), &errObj) != nil {
		return
	}
	// code 为空或 "0" 视为非错误（正常事件 data 可能无 code）
	codeStr := strings.Trim(string(errObj.Code), `"`)
	if codeStr == "" || codeStr == "0" || errObj.Msg == "" {
		return
	}
	if *streamError == nil {
		*streamError = fmt.Errorf("对话流返回错误: %s", errObj.Msg)
	}
	log.Errorf("[ChatflowSSE] channel=%s user=%s backend error: code=%s msg=%s",
		ch.ChannelID, msg.UserID, codeStr, errObj.Msg)
}
