package chat

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/types"
	"github.com/UnicomAI/wanwu/internal/channel-service/client"
	"github.com/UnicomAI/wanwu/internal/channel-service/client/model"
	"github.com/UnicomAI/wanwu/internal/channel-service/config"
	"github.com/UnicomAI/wanwu/internal/channel-service/wanwu"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// Handler 消息对话处理器
// 接收平台消息 → 查找通道配置 → 获取 API Key → 调用万悟 OpenAPI → 流式回传
type Handler struct {
	cfg         config.Config
	cli         client.IClient
	manager     adapterManager
	convManager *wanwu.ConversationManager
	questionMgr *QuestionManager
}

// adapterManager 适配器管理接口（避免循环依赖）
type adapterManager interface {
	GetAdapter(channelID string) (types.Adapter, bool)
	SendMessage(ctx context.Context, channelID, userID, content string, extra map[string]string) error
	CreateStreamSender(ctx context.Context, channelID, userID string, extra map[string]string) types.StreamSender
	SendFile(ctx context.Context, channelID, userID, fileName, mimeType string, data []byte, extra map[string]string) error
}

// NewHandler 创建消息处理器
func NewHandler(cfg config.Config, cli client.IClient, manager adapterManager) *Handler {
	return &Handler{
		cfg:         cfg,
		cli:         cli,
		manager:     manager,
		convManager: wanwu.NewConversationManager(cli),
		questionMgr: NewQuestionManager(cfg.BFF.ApiBaseUrl),
	}
}

// HandlePlatformMessage 处理来自平台的消息
func (h *Handler) HandlePlatformMessage(ctx context.Context, msg *types.PlatformMessage) error {
	log.Infof("received platform message: channel=%s, user=%s, type=%s, content=%s",
		msg.ChannelID, msg.UserID, msg.MsgType, truncate(msg.Content, 100))

	// 1. 查找通道配置
	ch, err := h.cli.GetChannel(ctx, msg.ChannelID)
	if err != nil {
		return fmt.Errorf("channel not found %s: %w", msg.ChannelID, err)
	}

	// 2. 检查通道状态
	if !ch.Enabled || ch.Status != "loggedIn" {
		return fmt.Errorf("channel %s is not active (enabled=%v, status=%s)", ch.ChannelID, ch.Enabled, ch.Status)
	}

	// 3. 获取 API Key
	if !ch.HasApiKey() {
		if ch.ApiKeyID != "" {
			return fmt.Errorf("channel %s has api_key_id (%s) but api_key value is empty, please rebind the API Key", ch.ChannelID, ch.ApiKeyID)
		}
		return fmt.Errorf("channel %s has no api key bound, please bind an API Key in channel settings", ch.ChannelID)
	}

	// 4. 优先处理 pending question：若该用户当前有待回答的 WGA question，
	// 本次消息不发给智能体，而是解析为 question 回复（序号 / 取消）。
	// 仅 wga 通道会产生 question，但 manager 按 channelID+userID 存取，不会误命中 agent 通道。
	if pq, ok := h.questionMgr.Get(msg.ChannelID, msg.UserID); ok {
		return h.handleQuestionReply(ctx, ch, msg, pq)
	}

	// 5. 按 appType 分发
	switch ch.AppType {
	case "wga":
		return h.handleWGAMessage(ctx, ch, msg)
	case "dip":
		return h.handleDIPMessage(ctx, ch, msg)
	default: // "agent"
		return h.handleAgentMessage(ctx, ch, msg)
	}
}

// handleAgentMessage 处理普通智能体消息
func (h *Handler) handleAgentMessage(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage) error {
	apiKey := ch.ApiKey

	// 获取或创建万悟会话 ID（同一用户同一通道复用同一会话，保持上下文记忆）
	wanwuClient := wanwu.NewClient(h.cfg.BFF.ApiBaseUrl)
	conversationID, ok := h.convManager.GetConversationID(ctx, msg.ChannelID, msg.UserID, "agent")
	if !ok {
		// 首次对话，创建会话
		convResp, err := wanwuClient.CreateConversation(ctx, apiKey, &wanwu.CreateConversationRequest{
			UUID:  ch.AppID,
			Title: truncate(msg.Content, 50),
		})
		if err != nil {
			log.Warnf("failed to create conversation for channel %s user %s: %v, will chat without conversation_id", msg.ChannelID, msg.UserID, err)
			// 创建会话失败时不阻断对话，不传 conversation_id 让 BFF 自动处理
		} else {
			conversationID = convResp.ConversationID
			h.convManager.SetConversationID(ctx, msg.ChannelID, msg.UserID, "agent", conversationID)
			log.Infof("created conversation %s for channel %s user %s", conversationID, msg.ChannelID, msg.UserID)
		}
	}

	// 调用万悟 OpenAPI 智能体对话
	chatReq := &wanwu.ChatRequest{
		UUID:           ch.AppID,
		ConversationID: conversationID,
		Query:          msg.Content,
		Stream:         true,
	}

	resp, err := wanwuClient.ChatWithAgent(ctx, apiKey, chatReq)
	if err != nil {
		return fmt.Errorf("failed to call wanwu chat api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 处理 SSE 流式响应
	return h.handleAgentSSEResponse(ctx, ch, msg, resp)
}

// handleWGAMessage 处理通用智能体（WGA）消息
func (h *Handler) handleWGAMessage(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage) error {
	return h.doWGAChat(ctx, ch, msg, "wga", wgaAgentID(ch.AgentId), false)
}

// wgaAgentID 把"通用智能体"哨兵 "null" 归一化为空串（WGA 端留空走 Supervisor 默认路由），
// 其余子智能体 id 原样返回。哨兵 "null" 由 bff 在选「无」子智能体时存入 channels.agent_id。
func wgaAgentID(id string) string {
	if id == "null" {
		return ""
	}
	return id
}

// handleDIPMessage 处理数字员工（DIP Agent）消息。
// DIP 模式要求 agentId 固定为 "DIP Agent"，且消息以 "@员工名称 " 开头（BFF buildWgaOntologyDIPMode
// 据此解析执行者）。通道绑定的员工名称存在 ch.AppName（员工 id 存 ch.AgentId），这里改写消息前缀后走 WGA 链路。
// 会话用独立 key "dip"，与 wga 会话隔离。
func (h *Handler) handleDIPMessage(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage) error {
	if ch.AppName == "" {
		return fmt.Errorf("channel %s is dip type but has no digital employee name (app_name) configured", ch.ChannelID)
	}
	return h.doWGAChat(ctx, ch, msg, "dip", "DIP Agent", true)
}

// doWGAChat WGA/DIP 共用的对话处理：建会话 → 构造消息 → 调 WGA 对话接口 → 处理 SSE。
//   - appTypeKey: 会话隔离 key（"wga" / "dip"）
//   - agentID: 传给 WGA 的 agentId（wga 用 ch.AgentId 直连子智能体；dip 固定 "DIP Agent"）
//   - rewriteWithEmployee: dip 场景给消息文本前缀 "@员工名 "（员工名取自 ch.AppName，仅当原文不以 @ 开头）
func (h *Handler) doWGAChat(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage, appTypeKey, agentID string, rewriteWithEmployee bool) error {
	apiKey := ch.ApiKey

	// 检查 modelUuid
	if ch.ModelUuid == "" {
		return fmt.Errorf("channel %s is %s type but has no model_uuid configured", ch.ChannelID, ch.AppType)
	}

	// 获取或创建 WGA 会话（threadId）
	wanwuClient := wanwu.NewClient(h.cfg.BFF.ApiBaseUrl)
	threadID, ok := h.convManager.GetConversationID(ctx, msg.ChannelID, msg.UserID, appTypeKey)
	if !ok {
		// 首次对话，创建 WGA 会话
		convResp, err := wanwuClient.CreateWGAConversation(ctx, apiKey, &wanwu.WGACreateConversationRequest{
			Title:     truncate(msg.Content, 50),
			ModelUuid: ch.ModelUuid,
		})
		if err != nil {
			return fmt.Errorf("failed to create wga conversation for channel %s user %s: %w", msg.ChannelID, msg.UserID, err)
		}
		threadID = convResp.ThreadID
		h.convManager.SetConversationID(ctx, msg.ChannelID, msg.UserID, appTypeKey, threadID)
		log.Infof("created wga conversation %s for channel %s user %s (appType=%s)", threadID, msg.ChannelID, msg.UserID, appTypeKey)
	}

	// DIP：消息文本前缀 "@员工名 "，使 BFF 解析出执行者（仅在原文不以 @ 开头时改写，避免重复）
	// 员工名取自 ch.AppName（员工 id 存 ch.AgentId）
	if rewriteWithEmployee && ch.AppName != "" && msg.Content != "" && !strings.HasPrefix(strings.TrimSpace(msg.Content), "@") {
		origContent := msg.Content
		msg.Content = "@" + ch.AppName + " " + msg.Content
		defer func() { msg.Content = origContent }() // 构造完毕后恢复，避免污染后续流程
	}

	// 构造消息内容：有附件时走多模态（先上传文件到 minio 拿 filePath，再拼 binary 引用）
	content, err := h.buildWGAContent(ctx, wanwuClient, apiKey, msg)
	if err != nil {
		return fmt.Errorf("failed to build wga message content for channel %s user %s: %w", msg.ChannelID, msg.UserID, err)
	}

	// 调用 WGA 对话接口
	chatReq := &wanwu.WGAChatRequest{
		ThreadID: threadID,
		Messages: []wanwu.WGAMessage{
			{
				Role:    "user",
				Content: content,
			},
		},
		ModelUuid: ch.ModelUuid,
		AgentId:   agentID, // wga: 直连通道绑定的子智能体（留空走 Supervisor）；dip: 固定 "DIP Agent"
	}

	resp, err := wanwuClient.ChatWithWGA(ctx, apiKey, chatReq)
	if err != nil {
		log.Errorf("[WGA] chat api call failed: channel=%s user=%s threadId=%s agentId=%s modelUuid=%s err=%v",
			ch.ChannelID, msg.UserID, threadID, agentID, ch.ModelUuid, err)
		return fmt.Errorf("failed to call wga chat api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	log.Infof("[WGA] chat api responded: channel=%s user=%s threadId=%s agentId=%s status=%d",
		ch.ChannelID, msg.UserID, threadID, agentID, resp.StatusCode)

	// 处理 WGA AG-UI SSE 流式响应
	return h.handleWGASSEResponse(ctx, ch, msg, resp, threadID)
}

// buildWGAContent 构造 WGA 消息内容。
// 无附件时返回纯文本 string；有附件时返回多模态数组：
// 先把附件上传到万悟 minio（/file/upload/direct）拿 filePath，再拼成 binary 引用，
// 供 WGA 识别文件内容。文本部分取 msg.Content（附件消息可能为空）。
func (h *Handler) buildWGAContent(ctx context.Context, wanwuClient *wanwu.Client, apiKey string, msg *types.PlatformMessage) (interface{}, error) {
	if len(msg.Attachments) == 0 {
		return msg.Content, nil
	}

	parts := make([]wanwu.WGAMessageContentPart, 0, len(msg.Attachments)+1)
	if msg.Content != "" {
		parts = append(parts, wanwu.WGAMessageContentPart{Type: "text", Text: msg.Content})
	}
	for _, att := range msg.Attachments {
		uf, err := wanwuClient.UploadFile(ctx, apiKey, att.Name, att.MimeType, att.Data)
		if err != nil {
			return nil, fmt.Errorf("upload attachment %s failed: %w", att.Name, err)
		}
		parts = append(parts, wanwu.WGAMessageContentPart{
			Type:     "binary",
			MimeType: att.MimeType,
			URL:      uf.FilePath,
			FileName: att.Name,
		})
		log.Infof("[WGA] uploaded attachment %s (%d bytes) -> %s for channel %s user %s",
			att.Name, len(att.Data), uf.FilePath, msg.ChannelID, msg.UserID)
	}
	return parts, nil
}

// handleAgentSSEResponse 处理普通智能体 SSE 流式响应
func (h *Handler) handleAgentSSEResponse(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage, resp *http.Response) error {
	// 尝试创建流式发送器（支持流式卡片的平台会返回非 nil）
	streamSender := h.manager.CreateStreamSender(ctx, msg.ChannelID, msg.UserID, msg.Extra)

	var fullContent strings.Builder
	reader := bufio.NewReader(resp.Body)
	chunkCount := 0

	log.Infof("[AgentSSE] channel=%s user=%s start streaming from agent %s (streamSender=%v)",
		ch.ChannelID, msg.UserID, ch.AppID, streamSender != nil)

	for {
		select {
		case <-ctx.Done():
			log.Infof("[AgentSSE] channel=%s user=%s context cancelled after %d chunks", ch.ChannelID, msg.UserID, chunkCount)
			closeStreamSender(streamSender, ctx, ctx.Err())
			return ctx.Err()
		default:
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			closeStreamSender(streamSender, ctx, err)
			return fmt.Errorf("error reading SSE stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 解析 SSE 数据行
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)

		if data == "[DONE]" {
			log.Infof("[AgentSSE] channel=%s user=%s received [DONE] signal after %d chunks", ch.ChannelID, msg.UserID, chunkCount)
			break
		}

		// 跳过空数据行
		if data == "" {
			continue
		}

		// 解析 SSE 数据
		// BFF OpenAPI agent chat 返回格式: data:{"response":"...","msg_id":"...","eventType":...}
		var sseData struct {
			Response string `json:"response"`
		}
		if err := json.Unmarshal([]byte(data), &sseData); err != nil {
			log.Errorf("[AgentSSE] channel=%s user=%s failed to parse SSE data: %v, raw: %s", ch.ChannelID, msg.UserID, err, data)
			continue
		}

		if sseData.Response != "" {
			// 流式路径：逐 chunk 更新卡片
			if streamSender != nil {
				if err := streamSender.SendChunk(ctx, sseData.Response, false); err != nil {
					log.Errorf("[AgentSSE] channel=%s user=%s stream sender chunk failed, falling back to non-streaming: %v",
						ch.ChannelID, msg.UserID, err)
					// 流式发送失败，收尾卡片（置 failed）后降级为非流式
					closeStreamSender(streamSender, ctx, fmt.Errorf("stream chunk failed: %w", err))
					streamSender = nil
				}
			}
			fullContent.WriteString(sseData.Response)
			chunkCount++
			// log.Debugf("[AgentSSE] channel=%s user=%s chunk #%d: %q", ch.ChannelID, msg.UserID, chunkCount, truncate(sseData.Response, 100))
		}
	}

	replyContent := fullContent.String()

	// 流式路径：发送最终 chunk 标记完成
	if streamSender != nil {
		if err := streamSender.SendChunk(ctx, "", true); err != nil {
			log.Errorf("[AgentSSE] channel=%s user=%s stream sender final chunk failed: %v", ch.ChannelID, msg.UserID, err)
			// 最终 chunk 失败，收尾卡片（置 failed）后降级为非流式
			closeStreamSender(streamSender, ctx, fmt.Errorf("final chunk failed: %w", err))
			streamSender = nil
		} else {
			// 正常完成：SendChunk(isFinal) 已将卡片置为 finished，Close 此时为幂等 no-op
			closeStreamSender(streamSender, ctx, nil)
			log.Infof("[AgentSSE] channel=%s user=%s stream completed via card, total %d chunks, %d chars",
				ch.ChannelID, msg.UserID, chunkCount, len(replyContent))
			return nil
		}
	}

	// 非流式路径：将完整回复发送给平台用户
	if replyContent == "" {
		log.Warnf("[AgentSSE] channel=%s user=%s empty reply from wanwu agent after %d chunks", ch.ChannelID, msg.UserID, chunkCount)
		return nil
	}

	log.Infof("[AgentSSE] channel=%s user=%s stream completed, total %d chunks, reply length=%d, content: %s",
		ch.ChannelID, msg.UserID, chunkCount, len(replyContent), truncate(replyContent, 200))

	if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, replyContent, msg.Extra); err != nil {
		return fmt.Errorf("failed to send reply to platform: %w", err)
	}

	log.Infof("[AgentSSE] channel=%s user=%s reply sent to platform successfully", ch.ChannelID, msg.UserID)
	return nil
}

// handleWGASSEResponse 处理 WGA AG-UI SSE 流式响应
// AG-UI 事件格式：
//
//	data: {"type":"TEXT_MESSAGE_START","messageId":"msg-1","role":"assistant"}
//	data: {"type":"TEXT_MESSAGE_CONTENT","messageId":"msg-1","delta":"你好"}
//	data: {"type":"TEXT_MESSAGE_END","messageId":"msg-1"}
//	data: {"type":"RUN_FINISHED","threadId":"xxx","runId":"run-1"}
func (h *Handler) handleWGASSEResponse(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage, resp *http.Response, threadID string) error {
	// 过程逐段下发：每个 TEXT_MESSAGE 段（START…END）结束时，把该段完整内容作为一条
	// 独立消息 SendMessage 给通道（钉钉/微信），让用户实时看到生成过程，不再用流式卡片。
	// textBuf 累积当前段的 delta，TEXT_MESSAGE_END 时一次性发出（不逐 delta，避免消息数爆炸）。
	var textBuf strings.Builder

	// sendProgress 把一条过程里程碑消息即时发给通道（只发 transfer/子智能体 finished 这类
	// 关键节点，常规工具调用不在此下发，避免过程刷屏撞 IM 频控）。
	sendProgress := func(text string) {
		if strings.TrimSpace(text) == "" {
			return
		}
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, text, msg.Extra); err != nil {
			log.Warnf("[WGA-SSE] channel=%s user=%s send progress failed: %v",
				ch.ChannelID, msg.UserID, err)
		} else {
			log.Infof("[WGA-SSE] channel=%s user=%s sent progress: %s",
				ch.ChannelID, msg.UserID, truncate(text, 50))
		}
	}

	// 事件聚合器：仍收集 fragment（保留扩展余地），但本路径不再用卡片渲染。
	agg := newWgaAggregator()
	reader := bufio.NewReader(resp.Body)
	var runID string // RUN_FINISHED 解析出的 runId，用于下载工作区产物
	// mentionedFiles 本次 SSE 流中智能体在正文里提到过的产物文件名（如 武则天.pptx），
	// RUN_FINISHED 后据此去工作区精确匹配并回发，不依赖快照 diff（diff 对固定路径覆盖写不可靠）。
	var mentionedFiles []string
	// questionCancelCh 在收到 ACTIVITY_SNAPSHOT(question,pending) 后被赋值；
	// 用户超时未答或放弃时被 close，通知本循环退出（避免 WGA 不再推事件时永久阻塞）。
	var questionCancelCh chan struct{}

	log.Infof("[WGA-SSE] channel=%s user=%s start streaming from WGA model %s",
		ch.ChannelID, msg.UserID, ch.ModelUuid)

	// 把阻塞的 ReadString 放进独立 goroutine，通过 lineCh 喂给主循环，
	// 这样主循环的 select 能同时响应 questionCancelCh/ctx.Done，不会卡死在读取上。
	type readResult struct {
		line string
		err  error
	}
	lineCh := make(chan readResult)
	go func() {
		for {
			line, err := reader.ReadString('\n')
			select {
			case lineCh <- readResult{line: line, err: err}:
			case <-ctx.Done():
				return
			}
			if err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-questionCancelCh:
			// pending question 超时或被放弃，结束本次 SSE 读取
			log.Infof("[WGA-SSE] channel=%s user=%s question cancelled/timed out, exit stream loop",
				ch.ChannelID, msg.UserID)
			return nil
		case r := <-lineCh:
			if r.err != nil {
				if r.err == io.EOF {
					// EOF 时 line 可能含最后一行数据，继续处理
					if strings.TrimSpace(r.line) == "" {
						goto wgaDone
					}
					// fallthrough 处理最后一行后退出
				} else {
					return fmt.Errorf("error reading WGA SSE stream: %w", r.err)
				}
			}
			line := r.line

			line = strings.TrimSpace(line)
			if line == "" {
				if r.err == io.EOF {
					goto wgaDone
				}
				continue
			}

			// 解析 SSE 数据行
			if !strings.HasPrefix(line, "data:") {
				if r.err == io.EOF {
					goto wgaDone
				}
				continue
			}

			data := strings.TrimPrefix(line, "data:")
			data = strings.TrimSpace(data)

			if data == "[DONE]" {
				goto wgaDone
			}

			// 跳过空数据行
			if data == "" {
				if r.err == io.EOF {
					goto wgaDone
				}
				continue
			}

			// 打印 WGA 返回的原始 SSE data 行（排查上游无响应/事件丢失等问题）
			// log.Debugf("[WGA-SSE-RAW] channel=%s user=%s data: %s", ch.ChannelID, msg.UserID, data)

			// 解析 AG-UI 事件（按 WGA 对话流协议，字段随事件类型不同）
			var event struct {
				Type         string          `json:"type"`
				Delta        string          `json:"delta"`
				RunId        string          `json:"runId"`
				ThreadId     string          `json:"threadId"`
				Message      string          `json:"message"`      // RUN_ERROR 错误信息
				MessageId    string          `json:"messageId"`    // TEXT/REASONING 消息 ID
				Timestamp    int64           `json:"timestamp"`    // 事件时间戳（ms）
				ToolCallName string          `json:"toolCallName"` // TOOL_CALL_START 工具名
				ToolCallId   string          `json:"toolCallId"`   // TOOL_CALL_* 工具调用 ID
				ActivityType string          `json:"activityType"` // ACTIVITY_SNAPSHOT 活动类型
				Content      json.RawMessage `json:"content"`      // TOOL_CALL_RESULT / ACTIVITY_SNAPSHOT 内容（结构不定）
			}
			if err := json.Unmarshal([]byte(data), &event); err != nil {
				log.Errorf("failed to parse WGA SSE data: %v, raw: %s", err, data)
				if r.err == io.EOF {
					goto wgaDone
				}
				continue
			}

			// 喂给聚合器的中间事件（type/delta/toolCall*/messageId/timestamp/activityType/content）
			wgaEv := &wgaEvent{
				eventType:    event.Type,
				delta:        event.Delta,
				toolCallID:   event.ToolCallId,
				toolCallName: event.ToolCallName,
				messageId:    event.MessageId,
				timestamp:    event.Timestamp,
				activityType: event.ActivityType,
				content:      event.Content,
			}

			switch event.Type {
			case "TEXT_MESSAGE_START":
				log.Debugf("[WGA-SSE] channel=%s user=%s %s", ch.ChannelID, msg.UserID, event.Type)
				agg.handleEvent(wgaEv)
				textBuf.Reset() // 新段开始，清空缓冲
			case "TEXT_MESSAGE_CONTENT":
				if event.Delta != "" {
					agg.handleEvent(wgaEv)
					textBuf.WriteString(event.Delta)
				}
			case "TEXT_MESSAGE_END":
				// 一段正文结束：逐段下发（正文照常实时发，过程类才只发里程碑）。
				// 同时从正文里提取智能体提到的产物文件名（如 武则天.pptx），供 RUN_FINISHED 后回发。
				agg.handleEvent(wgaEv)
				segment := textBuf.String()
				textBuf.Reset()
				if mentioned := extractMentionedFiles(segment); len(mentioned) > 0 {
					mentionedFiles = append(mentionedFiles, mentioned...)
				}
				if strings.TrimSpace(segment) == "" {
					log.Debugf("[WGA-SSE] channel=%s user=%s TEXT_MESSAGE_END empty segment, skip", ch.ChannelID, msg.UserID)
				} else if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, segment, msg.Extra); err != nil {
					log.Warnf("[WGA-SSE] channel=%s user=%s send text segment failed: %v",
						ch.ChannelID, msg.UserID, err)
				} else {
					log.Infof("[WGA-SSE] channel=%s user=%s sent text segment (%d chars): %s",
						ch.ChannelID, msg.UserID, len(segment), truncate(segment, 50))
				}
			case "RUN_FINISHED":
				// 对话结束，捕获 runId（下载工作区产物需要），跳出循环
				runID = event.RunId
				log.Infof("[WGA-SSE] channel=%s user=%s RUN_FINISHED: threadId=%s, runId=%s",
					ch.ChannelID, msg.UserID, event.ThreadId, runID)
				goto wgaDone
			case "RUN_ERROR":
				// 运行出错，WGA 不会再发 RUN_FINISHED，必须主动结束流，否则会一直阻塞等待
				log.Errorf("[WGA-SSE] channel=%s user=%s RUN_ERROR: threadId=%s, runId=%s, message=%s",
					ch.ChannelID, msg.UserID, event.ThreadId, event.RunId, event.Message)
				goto wgaDone
			case "RUN_STARTED":
				log.Infof("[WGA-SSE] channel=%s user=%s RUN_STARTED: threadId=%s, runId=%s",
					ch.ChannelID, msg.UserID, event.ThreadId, event.RunId)
			case "TOOL_CALL_START":
				log.Infof("[WGA-SSE] channel=%s user=%s TOOL_CALL_START: tool=%s, toolCallId=%s",
					ch.ChannelID, msg.UserID, event.ToolCallName, event.ToolCallId)
				agg.handleEvent(wgaEv)
			case "TOOL_CALL_ARGS":
				agg.handleEvent(wgaEv)
			case "TOOL_CALL_END":
				log.Infof("[WGA-SSE] channel=%s user=%s TOOL_CALL_END: toolCallId=%s",
					ch.ChannelID, msg.UserID, event.ToolCallId)
			case "TOOL_CALL_RESULT":
				log.Debugf("[WGA-SSE] channel=%s user=%s TOOL_CALL_RESULT: toolCallId=%s, content=%s",
					ch.ChannelID, msg.UserID, event.ToolCallId, truncate(string(event.Content), 200))
				// 工具调用结束（收到结果）：仅关键里程碑（如 Supervisor 委派 transfer）即时下发，
				// 常规工具（glob/read/skill/todowrite/bash 等）不下发，避免过程刷屏。
				completed, _ := agg.handleEvent(wgaEv)
				if completed != nil && completed.kind == fragToolCall && isMilestoneToolCall(completed) {
					sendProgress(renderToolCallLine(completed))
				}
			case "ACTIVITY_SNAPSHOT":
				// PPT Agent 等子智能体的进度快照（sub_agent started/finished、workspace 文件更新等）
				log.Infof("[WGA-SSE] channel=%s user=%s ACTIVITY_SNAPSHOT: activityType=%s, content=%s",
					ch.ChannelID, msg.UserID, event.ActivityType, truncate(string(event.Content), 200))
				// question（人机交互）：智能体提问，需用户回答后才继续。
				// 把选项拼成文本发出去，等用户回复序号后调 question/reply。
				if event.ActivityType == "question" {
					questionCancelCh = h.handleWGAQuestion(ctx, ch, msg, event.Content, questionCancelCh)
				}
				// 子智能体结束（sub_agent finished）：即时下发里程碑（不再缓冲合并）。
				// question/workspace 的快照不返回 completed，不会与此处下发冲突。
				completed, _ := agg.handleEvent(wgaEv)
				if completed != nil && completed.kind == fragActivity {
					sendProgress(renderActivityLine(completed))
				}
			case "REASONING_MESSAGE_START", "REASONING_MESSAGE_CONTENT":
				// 推理消息：仅喂聚合器（思考过程不下发到 IM，避免刷屏）
				// log.Debugf("[WGA-SSE] channel=%s user=%s %s", ch.ChannelID, msg.UserID, event.Type)
				agg.handleEvent(wgaEv)
			case "REASONING_MESSAGE_END":
				// 思考段结束：不下发（思考过程不再推到 IM）
				agg.handleEvent(wgaEv)
			default:
				log.Debugf("[WGA-SSE] channel=%s user=%s unhandled event: %s, raw=%s",
					ch.ChannelID, msg.UserID, event.Type, truncate(data, 300))
			}
		}
	}

wgaDone:

	// 各 TEXT_MESSAGE 段已在 END 时逐条发给通道；此处仅下发工作区产物（文件）。
	// 若末段未收到 END（流被中断），把残留 textBuf 兜底发出，避免丢最后一句。
	if segment := textBuf.String(); strings.TrimSpace(segment) != "" {
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, segment, msg.Extra); err != nil {
			log.Warnf("[WGA-SSE] channel=%s user=%s send trailing text segment failed: %v",
				ch.ChannelID, msg.UserID, err)
		} else {
			log.Infof("[WGA-SSE] channel=%s user=%s sent trailing text segment (%d chars)",
				ch.ChannelID, msg.UserID, len(segment))
		}
		textBuf.Reset()
	}

	// 收尾聚合器（把未关闭的 activity 挂回顶层）；诊断未完成的过程 fragment（只记日志不下发，
	// 未完成的思考/工具调用内容不完整，下发会误导用户）。
	agg.finalize()
	for _, f := range agg.unfinishedToolCalls() {
		log.Warnf("[WGA-SSE] channel=%s user=%s unfinished tool_call: %s (id=%s), no RESULT received",
			ch.ChannelID, msg.UserID, f.toolCallName, f.toolCallID)
	}

	_ = h.sendWorkspaceFiles(ctx, ch, msg, threadID, runID, mentionedFiles)
	return nil
}

// handleWGAQuestion 处理 SSE 收到的 WGA question（人机交互）事件。
// 把问题选项拼成「请回复序号」文本发到钉钉（独立消息），并把 pending question 存入 manager，
// 等用户回复序号后由 handleQuestionReply 调 question/reply。
// 返回该 pending 的 CancelCh，赋给 SSE 循环用于超时/放弃时退出。
// 已有 pending 时先复用其 CancelCh（同 user 不会并发两条 question，正常只会有一个）。
func (h *Handler) handleWGAQuestion(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage,
	rawContent json.RawMessage, prevCancelCh chan struct{}) chan struct{} {

	var content wanwu.WGAQuestionContent
	if err := json.Unmarshal(rawContent, &content); err != nil {
		log.Errorf("[WGA-SSE] channel=%s user=%s failed to parse question content: %v, raw=%s",
			ch.ChannelID, msg.UserID, err, truncate(string(rawContent), 300))
		return prevCancelCh
	}
	if content.QuestionID == "" || content.RunID == "" {
		log.Warnf("[WGA-SSE] channel=%s user=%s question missing questionId/runId, skip: %+v",
			ch.ChannelID, msg.UserID, content)
		return prevCancelCh
	}
	// 非 pending（answered/rejected）的快照不处理：通常伴随后续事件，无需发问。
	if content.Status != "" && content.Status != "pending" {
		log.Infof("[WGA-SSE] channel=%s user=%s question status=%s, skip",
			ch.ChannelID, msg.UserID, content.Status)
		return prevCancelCh
	}

	text := formatQuestionText(content.Questions)
	if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, text, msg.Extra); err != nil {
		log.Errorf("[WGA-SSE] channel=%s user=%s send question text to dingtalk failed: %v",
			ch.ChannelID, msg.UserID, err)
	}

	// 复用 prevCancelCh：避免丢弃已赋值给 SSE 循环的旧 channel（否则旧 channel 永不会被 close）。
	cancelCh := prevCancelCh
	h.questionMgr.Set(msg.ChannelID, msg.UserID, &PendingQuestion{
		QuestionID: content.QuestionID,
		RunID:      content.RunID,
		ThreadID:   content.ThreadID,
		ApiKey:     ch.ApiKey,
		Questions:  content.Questions,
		CancelCh:   cancelCh, // Set 内部确保非 nil
	})
	log.Infof("[WGA-SSE] channel=%s user=%s pending question stored: questionId=%s, runId=%s, %d question(s)",
		ch.ChannelID, msg.UserID, content.QuestionID, content.RunID, len(content.Questions))
	return cancelCh
}

// handleQuestionReply 处理用户对 pending question 的回复消息。
// - "取消" → 调 question/reject，删除 pending（close CancelCh 让 SSE 退出），发「已取消」。
// - 序号 → 解析为 answers 调 question/reply；成功后 Complete（不 close CancelCh，SSE 继续读后续事件）。
// - 格式错 → 发格式提示，保留 pending 等用户重发。
func (h *Handler) handleQuestionReply(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage, pq *PendingQuestion) error {
	wanwuClient := wanwu.NewClient(h.cfg.BFF.ApiBaseUrl)
	content := strings.TrimSpace(msg.Content)

	// 取消：放弃该 question
	if content == "取消" || content == "cancel" {
		if err := wanwuClient.RejectQuestion(ctx, pq.ApiKey, pq.RunID, pq.QuestionID); err != nil {
			log.Errorf("[Question] channel=%s user=%s reject failed: %v", msg.ChannelID, msg.UserID, err)
			// reject 失败仍删除 pending 并让 SSE 退出，避免永久卡住
		}
		h.questionMgr.Delete(msg.ChannelID, msg.UserID) // close CancelCh → SSE goroutine 退出
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, "已取消，本次生成已放弃。", msg.Extra); err != nil {
			log.Warnf("[Question] channel=%s user=%s send cancel notice failed: %v",
				msg.ChannelID, msg.UserID, err)
		}
		log.Infof("[Question] channel=%s user=%s question rejected by user", msg.ChannelID, msg.UserID)
		return nil
	}

	// 解析序号 → answers
	answers, perr := parseQuestionReply(content, pq.Questions)
	if perr != nil {
		tip := formatReplyError(perr, pq.Questions)
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); err != nil {
			log.Warnf("[Question] channel=%s user=%s send parse error tip failed: %v",
				msg.ChannelID, msg.UserID, err)
		}
		// 保留 pending，等用户按正确格式重发
		return nil
	}

	// 调 question/reply
	if err := wanwuClient.ReplyQuestion(ctx, pq.ApiKey, &wanwu.WGAQuestionReplyRequest{
		RunID:      pq.RunID,
		QuestionID: pq.QuestionID,
		Answers:    answers,
	}); err != nil {
		log.Errorf("[Question] channel=%s user=%s reply failed: %v", msg.ChannelID, msg.UserID, err)
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID,
			"提交选择失败，请重试，或回复「取消」放弃。", msg.Extra); err != nil {
			log.Warnf("[Question] channel=%s user=%s send reply-fail tip failed: %v",
				msg.ChannelID, msg.UserID, err)
		}
		// 保留 pending 让用户重试
		return nil
	}

	// 成功：从 store 移除（不 close CancelCh），SSE goroutine 继续读 WGA 推来的后续事件。
	h.questionMgr.Complete(msg.ChannelID, msg.UserID)
	log.Infof("[Question] channel=%s user=%s question replied ok: questionId=%s, runId=%s",
		msg.ChannelID, msg.UserID, pq.QuestionID, pq.RunID)
	return nil
}

// formatQuestionText 把 WGA question 拼成发给钉钉的「请回复序号」文本。
// 每个问题列出带序号的选项；末尾给出格式示例（空格分问题，逗号分多选）。
func formatQuestionText(questions []wanwu.WGAQuestion) string {
	var b strings.Builder
	b.WriteString("智能体需要你确认：\n\n")
	for i, q := range questions {
		header := q.Header
		if header == "" {
			header = q.Question
		}
		mark := "单选"
		if q.Multiple {
			mark = "多选"
		}
		fmt.Fprintf(&b, "【%d. %s】(%s)\n", i+1, header, mark)
		if q.Question != "" && q.Question != header {
			fmt.Fprintf(&b, "%s\n", q.Question)
		}
		for j, opt := range q.Options {
			if opt.Description != "" {
				fmt.Fprintf(&b, "  %d. %s（%s）\n", j+1, opt.Label, opt.Description)
			} else {
				fmt.Fprintf(&b, "  %d. %s\n", j+1, opt.Label)
			}
		}
		if q.Custom {
			b.WriteString("  (支持自定义输入)\n")
		}
		b.WriteByte('\n')
	}
	b.WriteString(formatReplyExample(questions))
	return b.String()
}

// formatReplyExample 生成格式示例，如「请回复序号：1 1,3 2（空格分问题，多选用逗号）」。
func formatReplyExample(questions []wanwu.WGAQuestion) string {
	parts := make([]string, 0, len(questions))
	for i, q := range questions {
		if q.Multiple {
			// 多选示例取前两个选项序号
			if len(q.Options) >= 2 {
				parts = append(parts, fmt.Sprintf("%d,%d", 1, 2))
			} else {
				parts = append(parts, fmt.Sprintf("%d", i+1))
			}
		} else {
			parts = append(parts, "1")
		}
	}
	return "请回复序号：" + strings.Join(parts, " ") + "（空格分问题，多选用逗号，回复「取消」放弃）"
}

// formatReplyError 解析失败时给出可读提示，并附上格式示例。
func formatReplyError(err error, questions []wanwu.WGAQuestion) string {
	return fmt.Sprintf("回复格式有误：%v。\n%s", err, formatReplyExample(questions))
}

// parseQuestionReply 把用户回复解析为 answers 二维数组。
// 格式：空格分问题组，每组内逗号/顿号分多选序号，如 "1 1,3 2"。
// 容错：中英文逗号、顿号、多余空格；越界/非数字报错。
// 多选问题给出单个序号也合法（只选一个）。
func parseQuestionReply(content string, questions []wanwu.WGAQuestion) ([][]string, error) {
	// 归一化分隔符：中文逗号、顿号 → 英文逗号
	normalized := strings.NewReplacer("，", ",", "、", ",", "\t", " ").Replace(content)
	groups := strings.Fields(normalized) // 按空格切分

	if len(groups) != len(questions) {
		return nil, fmt.Errorf("需要回复 %d 个问题的序号，但收到 %d 组（用空格分隔每个问题）",
			len(questions), len(groups))
	}

	answers := make([][]string, len(questions))
	for i, g := range groups {
		q := questions[i]
		tokens := strings.Split(g, ",")
		idxs := make([]int, 0, len(tokens))
		for _, tok := range tokens {
			tok = strings.TrimSpace(tok)
			if tok == "" {
				continue
			}
			n, err := strconv.Atoi(tok)
			if err != nil {
				return nil, fmt.Errorf("第 %d 个问题：%q 不是有效序号", i+1, tok)
			}
			if n < 1 || n > len(q.Options) {
				return nil, fmt.Errorf("第 %d 个问题：序号 %d 越界（可选 1~%d）",
					i+1, n, len(q.Options))
			}
			idxs = append(idxs, n)
		}
		if len(idxs) == 0 {
			return nil, fmt.Errorf("第 %d 个问题未选择任何序号", i+1)
		}
		// 非多选只允许选一个
		if !q.Multiple && len(idxs) > 1 {
			return nil, fmt.Errorf("第 %d 个问题是单选，但选了 %d 个序号", i+1, len(idxs))
		}
		labels := make([]string, 0, len(idxs))
		for _, n := range idxs {
			labels = append(labels, q.Options[n-1].Label)
		}
		answers[i] = labels
	}
	return answers, nil
}

// closeStreamSender 收尾流式发送器（置卡片为 finished/failed），忽略 nil 与 Close 自身的错误。
// 用于所有离开流式路径的出口（成功/失败/取消），避免卡片卡在 processing 状态。
func closeStreamSender(s types.StreamSender, ctx context.Context, err error) {
	if s == nil {
		return
	}
	if closeErr := s.Close(ctx, err); closeErr != nil {
		log.Warnf("close stream sender failed: %v (original err: %v)", closeErr, err)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// maxWorkspaceFileSize 单文件大小上限（100MB）。
// 钉钉支持分块事务上传（>20MB 自动走 enable/chunk/submit），微信整文件加密上传，两者均能发大文件；
// 此上限仅作防呆，避免单文件过大拖垮 IM 上传。
const maxWorkspaceFileSize = 100 * 1024 * 1024

// mentionedFileRe 匹配智能体正文里提到的产物文件名（文档/图片/压缩包/网页等最终产物扩展名）。
// 用于 RUN_FINISHED 后去工作区精确匹配回发。预编译避免每次调用重复编译。
// 含 html：网页生成场景下 .html 是用户要的最终产物（css/js 不在此列，由 isFinalArtifact 拦截）。
var mentionedFileRe = regexp.MustCompile(`[\w一-龥.\-]+\.(?:pptx?|docx?|xlsx?|pdf|png|jpe?g|gif|zip|rar|md|txt|csv|html?)`)

// workspaceFileSendGap 工作区产物文件之间的下发间隔（及失败兜底提示前的等待）。
// 微信 ilink sendmessage 在短时间密集推送时会返回 ret=-2（频控），留 1.5s 间隔降低撞频控概率。
const workspaceFileSendGap = 1500 * time.Millisecond

// workspaceFileSendRetry 文件下发失败的重试次数（不含首次）。
// 针对 IM 平台瞬时频控（如微信 ret=-2）：间隔后重试，仍失败则跳过。
const workspaceFileSendRetry = 2

// sendFileWithRetry 发送工作区文件，失败时间隔重试（应对 IM 平台瞬时频控，如微信 ret=-2）。
// ErrFileSendUnsupported 不重试（平台根本不支持，直接返回让调用方降级）。
func (h *Handler) sendFileWithRetry(ctx context.Context, msg *types.PlatformMessage, name, mime string, data []byte) error {
	var lastErr error
	for attempt := 0; attempt <= workspaceFileSendRetry; attempt++ {
		err := h.manager.SendFile(ctx, msg.ChannelID, msg.UserID, name, mime, data, msg.Extra)
		if err == nil {
			return nil
		}
		if errors.Is(err, types.ErrFileSendUnsupported) {
			return err // 平台不支持，不重试
		}
		lastErr = err
		log.Warnf("[WGA-WS] channel=%s user=%s send file %s attempt %d/%d failed: %v",
			msg.ChannelID, msg.UserID, name, attempt+1, workspaceFileSendRetry+1, err)
		if attempt < workspaceFileSendRetry {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(workspaceFileSendGap):
			}
		}
	}
	return lastErr
}

// wgaFileNode 是工作区目录树里一个文件节点的本地表示（含工作区内完整相对路径）。
type wgaFileItem struct {
	name string // 文件名（发 IM 用）
	path string // 工作区内完整相对路径（下载 API 用）
	mime string
	size int64
}

// joinWGAPath 拼接 WGA 工作区内相对路径（以 / 分隔）。dir 为空时返回 name，避免前导 /。
func joinWGAPath(dir, name string) string {
	if dir == "" {
		return name
	}
	return dir + "/" + name
}

// extractMentionedFiles 从智能体正文段里提取产物文件名（如 武则天.pptx、report.docx）。
// 匹配常见文档/图片/压缩包扩展名；返回去重后的文件名列表（仅文件名，不含路径）。
// 用于 RUN_FINISHED 后去工作区精确匹配回发，不依赖快照 diff（diff 对固定路径覆盖写不可靠）。
func extractMentionedFiles(text string) []string {
	matches := mentionedFileRe.FindAllString(text, -1)
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		// 过滤明显非文件名的误匹配（如纯扩展名 "pptx" 单独出现），要求含至少一个非扩展名字符
		name := strings.TrimSpace(m)
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// sendWorkspaceFiles 在 WGA 对话结束后，把智能体本次生成的产物下载并回发 IM。
// 工作区是 thread 级累积（含历史 run 的文件，无修改时间字段），无法靠快照 diff 区分本次新增
// （PPT Agent 用固定路径覆盖写，diff 永远判"无新增"会漏发）。改为按 mentionedFiles（本次 SSE 流中
// 智能体正文里提到的产物文件名）去工作区精确匹配回发；mentionedFiles 为空时降级用白名单+
// 用户请求关键词匹配兜底。钉钉/微信实现 FileSender 真实发文件；飞书不实现，降级文本提示。
// 任何失败只记日志，不影响已发文本回复。
func (h *Handler) sendWorkspaceFiles(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage, threadID, runID string, mentionedFiles []string) error {
	if threadID == "" || runID == "" {
		log.Infof("[WGA-WS] channel=%s user=%s skip workspace files: threadID or runID empty", msg.ChannelID, msg.UserID)
		return nil
	}

	wanwuClient := wanwu.NewClient(h.cfg.BFF.ApiBaseUrl)
	ws, err := wanwuClient.WGAWorkspace(ctx, ch.ApiKey, threadID, runID)
	if err != nil {
		log.Warnf("[WGA-WS] channel=%s user=%s failed to get workspace: %v", msg.ChannelID, msg.UserID, err)
		return nil
	}
	if !ws.IsDisplay || ws.FileCount == 0 {
		log.Infof("[WGA-WS] channel=%s user=%s no workspace files to send (isDisplay=%v, fileCount=%d)",
			msg.ChannelID, msg.UserID, ws.IsDisplay, ws.FileCount)
		return nil
	}

	// 递归收集所有 type=="file" 节点。
	// 目录树 API 节点只有 name（不含路径前缀），下载 API 的 path 需工作区内完整相对路径，
	// 因此递归时拼接父目录前缀（如 output/slide-08.js）。
	allFiles := make([]wgaFileItem, 0, ws.FileCount)
	var walk func(nodes []*wanwu.WGAFileNode, dir string)
	walk = func(nodes []*wanwu.WGAFileNode, dir string) {
		for _, n := range nodes {
			if n == nil {
				continue
			}
			if n.Type == "file" {
				allFiles = append(allFiles, wgaFileItem{name: n.Name, path: joinWGAPath(dir, n.Name), mime: n.MimeType, size: n.Size})
			}
			if len(n.Children) > 0 {
				walk(n.Children, joinWGAPath(dir, n.Name))
			}
		}
	}
	walk(ws.Files, "")

	if len(allFiles) == 0 {
		log.Infof("[WGA-WS] channel=%s user=%s workspace has no file nodes", msg.ChannelID, msg.UserID)
		return nil
	}

	// 仅回发智能体本次正文明确点名的产物文件。工作区是 thread 级历史累积（无修改时间、无本次
	// 新增信号），任何"按用户请求关键词匹配 / 全发候选"的兜底都会误推历史文件并触发 IM 频控；
	// 故未点名文件名即视为纯聊天/无本次产物，静默不发。产物不丢失，仍在万悟工作区网页端。
	var files []wgaFileItem
	if len(mentionedFiles) > 0 {
		mentionedSet := make(map[string]struct{}, len(mentionedFiles))
		for _, m := range mentionedFiles {
			mentionedSet[m] = struct{}{}
		}
		// 每个 mentioned 文件名只回发一份：工作区是递归目录树，同一文件名可能在不同目录
		// 出现多份（如 output/西施.pptx 与子目录副本），按名字去重避免同一产物重复发送。
		matched := make(map[string]struct{}, len(mentionedFiles))
		for _, f := range allFiles {
			if _, ok := mentionedSet[f.name]; !ok {
				continue
			}
			if _, dup := matched[f.name]; dup {
				continue
			}
			if isFinalArtifact(f.name, f.mime) {
				files = append(files, f)
				matched[f.name] = struct{}{}
			}
		}
		log.Infof("[WGA-WS] channel=%s user=%s matched %d/%d mentioned file(s) in workspace (mentioned=%v)",
			msg.ChannelID, msg.UserID, len(files), len(mentionedFiles), mentionedFiles)
	}
	if len(files) == 0 {
		// 智能体本次正文未提到任何产物文件名：视为纯聊天/无本次产物。
		// 工作区是 thread 级历史累积（无修改时间、无本次新增信号），回发会把历史文件
		// （README/LICENSE/CHANGELOG 等）误推给用户并触发 IM 频控。产物不丢失，仍在万悟工作区网页端。
		// [诊断] 打印工作区最终产物清单，用于排查"正文未点名但确有产物"的漏发场景（如 PPT Agent）。
		var artifacts []string
		for _, f := range allFiles {
			if isFinalArtifact(f.name, f.mime) {
				artifacts = append(artifacts, f.path)
			}
		}
		log.Infof("[WGA-WS] channel=%s user=%s skip workspace files: no mentioned artifact this run (workspace=%d, mentioned=%v); final artifacts in workspace: %v",
			msg.ChannelID, msg.UserID, len(allFiles), mentionedFiles, artifacts)
		return nil
	}

	log.Infof("[WGA-WS] channel=%s user=%s sending %d workspace file(s) (workspace total=%d), threadId=%s, runId=%s",
		msg.ChannelID, msg.UserID, len(files), len(allFiles), threadID, runID)

	sent := 0
	for i, f := range files {
		// 大文件跳过（100MB 上限，兼顾 IM 平台上传限制）
		if f.size > maxWorkspaceFileSize {
			log.Warnf("[WGA-WS] channel=%s user=%s skip file %s: size %d > %d",
				msg.ChannelID, msg.UserID, f.name, f.size, maxWorkspaceFileSize)
			continue
		}

		resp, err := wanwuClient.WGAWorkspaceDownload(ctx, ch.ApiKey, threadID, runID, f.path)
		if err != nil {
			log.Warnf("[WGA-WS] channel=%s user=%s download file %s failed: %v",
				msg.ChannelID, msg.UserID, f.name, err)
			continue
		}
		data, readErr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if readErr != nil {
			log.Warnf("[WGA-WS] channel=%s user=%s read file %s failed: %v",
				msg.ChannelID, msg.UserID, f.name, readErr)
			continue
		}

		if err := h.sendFileWithRetry(ctx, msg, f.name, f.mime, data); err != nil {
			if errors.Is(err, types.ErrFileSendUnsupported) {
				// 当前平台不支持发文件（如飞书），降级为文本提示：列出最终产物文件名，指引到工作区下载。
				h.sendWorkspaceFallbackTip(ctx, msg, files)
				log.Infof("[WGA-WS] channel=%s user=%s platform does not support file send, sent text tip",
					msg.ChannelID, msg.UserID)
				return nil
			}
			log.Warnf("[WGA-WS] channel=%s user=%s send file %s failed: %v",
				msg.ChannelID, msg.UserID, f.name, err)
			continue
		}
		sent++
		log.Infof("[WGA-WS] channel=%s user=%s sent file %s (%d bytes)",
			msg.ChannelID, msg.UserID, f.name, len(data))

		// 多文件场景下文件之间留一点间隔，避免短时间内连续推送触发 IM 平台频控
		// （微信 ilink sendmessage 在密集推送时会返回 ret=-2）。仅在有后续文件时 sleep。
		if i < len(files)-1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(workspaceFileSendGap):
			}
		}
	}

	if sent == 0 {
		// 支持发文件但全部发送/下载失败，间隔后提示用户到工作区下载（避免与失败请求连发再次撞频控）
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(workspaceFileSendGap):
		}
		h.sendWorkspaceFallbackTip(ctx, msg, files)
		return nil
	}

	// 至少发送成功一个文件：发一条 ✅ 生成汇总（关键里程碑），列出本次产物文件名。
	var names []string
	for _, f := range files {
		if f.size > maxWorkspaceFileSize {
			continue
		}
		names = append(names, f.name)
	}
	if len(names) > 0 {
		tip := "✅ 已生成：" + strings.Join(names, "、")
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); err != nil {
			log.Warnf("[WGA-WS] channel=%s user=%s send generated-summary failed: %v",
				msg.ChannelID, msg.UserID, err)
		}
	}
	return nil
}

// sendWorkspaceFallbackTip 发送降级文本提示：列出最终产物文件名，指引到工作区下载。
func (h *Handler) sendWorkspaceFallbackTip(ctx context.Context, msg *types.PlatformMessage, files []wgaFileItem) {
	var names []string
	for _, f := range files {
		names = append(names, f.name)
	}
	tip := fmt.Sprintf("已为你生成：%s\n请到万悟工作区网页端下载", strings.Join(names, "、"))
	if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); err != nil {
		log.Warnf("[WGA-WS] channel=%s user=%s send fallback tip failed: %v",
			msg.ChannelID, msg.UserID, err)
	}
}

// selectWorkspaceFiles / artifactNameScore 已移除：
// 原先"未点名文件名时按用户请求关键词匹配、全部不命中则全发候选"的兜底会把工作区历史文件
// 误推给用户并触发 IM 频控。改为仅回发正文明确点名的产物文件，未点名即静默不发。

// isFinalArtifact 判定是否最终产物（用户可直接打开的文档/图片/压缩包/网页）。
// 过滤中间产物：脚本（js/ts/py/go…）、配置（json/yaml）、样式（css/scss）、日志等。
// html/htm 算最终产物——网页生成场景下 .html 是用户要的产物，应回发；其依赖的 css/js 仍过滤。
func isFinalArtifact(name, mime string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepathExt(name), "."))
	switch ext {
	case "", "js", "mjs", "ts", "jsx", "tsx", "py", "rb", "go", "rs", "java",
		"c", "h", "cpp", "cc", "hpp", "cs", "php", "sh", "bash", "zsh",
		"json", "yaml", "yml", "toml", "ini", "cfg", "conf",
		"css", "scss", "less", "vue", "svelte",
		"xml", "svg", "map", "lock", "log", "tmp", "bak", "swp":
		return false
	}
	// mimeType 兜底：明确是代码/文本配置类的也过滤
	switch {
	case strings.HasPrefix(mime, "application/javascript"),
		strings.HasPrefix(mime, "text/javascript"),
		strings.HasPrefix(mime, "application/json"),
		strings.HasPrefix(mime, "application/x-yaml"),
		strings.HasPrefix(mime, "text/yaml"):
		return false
	}
	return true
}

// filepathExt 返回文件扩展名（含点），避免在 chat.go 引入 path/filepath 包。
func filepathExt(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i:]
		}
		if name[i] == '/' || name[i] == '\\' {
			break
		}
	}
	return ""
}
