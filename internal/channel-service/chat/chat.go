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
	"unicode/utf8"

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
	cfg             config.Config
	cli             client.IClient
	manager         adapterManager
	convManager        *wanwu.ConversationManager
	artifactMgr        *wanwu.ArtifactManager
	questionMgr        *QuestionManager
	attachmentCache    *AttachmentCache
	chatflowInputCache *ChatflowInputCache
	chatflowPendingInputs *ChatflowPendingInputCache
}

// adapterManager 适配器管理接口（避免循环依赖）
type adapterManager interface {
	GetAdapter(channelID string) (types.Adapter, bool)
	SendMessage(ctx context.Context, channelID, userID, content string, extra map[string]string) error
	SendMarkdown(ctx context.Context, channelID, userID, title, content string) error
	CreateStreamSender(ctx context.Context, channelID, userID string, extra map[string]string) types.StreamSender
	SendFile(ctx context.Context, channelID, userID, fileName, mimeType string, data []byte, extra map[string]string) error
}

// NewHandler 创建消息处理器
func NewHandler(cfg config.Config, cli client.IClient, manager adapterManager) *Handler {
	return &Handler{
		cfg:             cfg,
		cli:             cli,
		manager:         manager,
		convManager:        wanwu.NewConversationManager(cli),
		artifactMgr:        wanwu.NewArtifactManager(cli),
		questionMgr:        NewQuestionManager(cfg.BFF.ApiBaseUrl),
		attachmentCache:    NewAttachmentCache(),
		chatflowInputCache: NewChatflowInputCache(),
		chatflowPendingInputs: NewChatflowPendingInputCache(),
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

	// 4.5 优先处理对话流待补入参：chatflow 通道且该用户有追问状态时，
	// 本次消息当作入参回复（按 schema 类型转换填入 parameters），不发给对话流后端。
	// 追问状态仅在 chatflow 链路写入，不会误命中其他通道。
	if ps, ok := h.chatflowPendingInputs.Get(msg.ChannelID, msg.UserID); ok {
		return h.handleChatflowInputReply(ctx, ch, msg, ps)
	}

	// 5. 按 appType 分发
	switch ch.AppType {
	case "wga":
		return h.handleWGAMessage(ctx, ch, msg)
	case "dip":
		return h.handleDIPMessage(ctx, ch, msg)
	case "chatflow":
		return h.handleChatflowMessage(ctx, ch, msg)
	default: // "agent"
		return h.handleAgentMessage(ctx, ch, msg)
	}
}

// handleAgentMessage 处理普通智能体消息
// 支持"先发附件、再发文字"的分两步操作（与 WGA 链路同构，差异：agent 走 openapi file_info 单文件，
// WGA 走多模态 content 数组多文件）：
//   - 纯附件消息（无有效文字指令）：上传 minio → 存待用附件缓存 → 回提示 → 不调 agent。
//   - 有文字指令：drain 暂存附件 + 本条附件，取首个填 file_info 发给智能体；多附件仅用首个、其余丢弃并提示。
func (h *Handler) handleAgentMessage(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage) error {
	apiKey := ch.ApiKey
	wanwuClient := wanwu.NewClient(h.cfg.BFF.ApiBaseUrl)

	// 上传本条附件到 minio（纯附件消息也要上传，才能存 URL 进待用缓存）
	currentAtts, err := uploadAttachments(ctx, wanwuClient, apiKey, msg.ChannelID, msg.UserID, msg.Attachments)
	if err != nil {
		return fmt.Errorf("failed to upload attachments for channel %s user %s: %w", msg.ChannelID, msg.UserID, err)
	}

	// 提取有效文字指令：排除空白、微信占位符、等于附件文件名的 Content
	text := effectiveText(msg.Content, msg.Attachments)

	if text == "" {
		// 纯附件消息：存入待用缓存 + 回提示，不调 agent
		for _, a := range currentAtts {
			h.attachmentCache.Append(msg.ChannelID, msg.UserID, a)
		}
		tip := fmt.Sprintf("已收到 %d 个文件，请说明要做什么（%d 分钟内有效）",
			len(currentAtts), int(pendingAttachmentTTL.Minutes()))
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); err != nil {
			log.Warnf("[Agent] send attachment-received tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, err)
		}
		return nil
	}

	// 取出暂存附件，与本条附件一起取首个填 file_info
	pending := h.attachmentCache.Drain(msg.ChannelID, msg.UserID)
	fileInfo, droppedTip := buildAgentFileInfo(pending, currentAtts)
	if len(pending) > 0 {
		log.Infof("[Agent] drained %d pending attachments for channel %s user %s",
			len(pending), msg.ChannelID, msg.UserID)
	}
	// 多附件被丢弃：在 SSE 开始前作为独立消息发提示，与流式回复分离
	if droppedTip != "" {
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, droppedTip, msg.Extra); err != nil {
			log.Warnf("[Agent] send dropped-attachment tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, err)
		}
	}

	// 获取或创建万悟会话 ID（同一用户同一通道复用同一会话，保持上下文记忆）
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
		Query:          text,
		Stream:         true,
		FileInfo:       fileInfo,
	}

	resp, err := wanwuClient.ChatWithAgent(ctx, apiKey, chatReq)
	if err != nil {
		return fmt.Errorf("failed to call wanwu chat api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// 处理 SSE 流式响应
	return h.handleAgentSSEResponse(ctx, ch, msg, resp)
}

// buildAgentFileInfo 把暂存附件 + 本条附件取首个构造 openapi file_info（bff 硬限 1 个文件）。
// 返回 (fileInfo, droppedTip)：
//   - 无附件：fileInfo=nil, droppedTip=""。
//   - 1 个附件：fileInfo 含 1 项，droppedTip=""。
//   - >1 个附件：fileInfo 仅含首个，droppedTip 提示其余被忽略（列出文件名）。
//
// 注意：file_info.FileName 填的是上传响应 fileId（PendingAttachment.FileId），非原始文件名——
// 对齐文档「file_info.fileName 对应上传响应 fileId」。原始文件名仅用于 droppedTip 提示。
func buildAgentFileInfo(pending, current []*PendingAttachment) ([]wanwu.AgentFileInfo, string) {
	all := make([]*PendingAttachment, 0, len(pending)+len(current))
	all = append(all, pending...)
	all = append(all, current...)
	if len(all) == 0 {
		return nil, ""
	}
	first := all[0]
	fileInfo := []wanwu.AgentFileInfo{{
		FileName: first.FileId,
		FileSize: first.FileSize,
		FileUrl:  first.URL,
	}}
	if len(all) == 1 {
		return fileInfo, ""
	}
	// 多附件：其余丢弃，拼提示
	dropped := make([]string, 0, len(all)-1)
	for _, a := range all[1:] {
		dropped = append(dropped, a.FileName)
	}
	tip := fmt.Sprintf("智能体对话仅支持 1 个文件，已用「%s」，其余 %d 个忽略：%s",
		first.FileName, len(dropped), strings.Join(dropped, "、"))
	return fileInfo, tip
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
// 支持"先发附件、再发文字"的分两步操作：
//
//   - 纯附件消息（无有效文字指令）：上传 minio → 存待用附件缓存 → 回提示 → 不调 WGA。
//
//   - 有文字指令：把暂存附件 + 本条附件 + 文字拼进同一条 WGA 消息发出，清空缓存。
//
//   - appTypeKey: 会话隔离 key（"wga" / "dip"）
//
//   - agentID: 传给 WGA 的 agentId（wga 用 ch.AgentId 直连子智能体；dip 固定 "DIP Agent"）
//
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

	// 上传本条附件到 minio（纯附件消息也要上传，才能存 URL 进待用缓存）
	currentAtts, err := uploadAttachments(ctx, wanwuClient, apiKey, msg.ChannelID, msg.UserID, msg.Attachments)
	if err != nil {
		return fmt.Errorf("failed to upload attachments for channel %s user %s: %w", msg.ChannelID, msg.UserID, err)
	}

	// 提取有效文字指令：排除空白、微信占位符（[图片]/[语音] 等）、等于附件文件名的 Content
	text := effectiveText(msg.Content, msg.Attachments)

	if text == "" {
		// 纯附件消息：存入待用缓存 + 回提示，不调 WGA
		for _, a := range currentAtts {
			h.attachmentCache.Append(msg.ChannelID, msg.UserID, a)
		}
		tip := fmt.Sprintf("已收到 %d 个文件，请说明要做什么（%d 分钟内有效）",
			len(currentAtts), int(pendingAttachmentTTL.Minutes()))
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, tip, msg.Extra); err != nil {
			log.Warnf("[WGA] send attachment-received tip failed: channel=%s user=%s err=%v",
				msg.ChannelID, msg.UserID, err)
		}
		return nil
	}

	// DIP：消息文本前缀 "@员工名 "，使 BFF 解析出执行者（仅在原文不以 @ 开头时改写，避免重复）
	// 员工名取自 ch.AppName（员工 id 存 ch.AgentId）。仅在有文字指令时改写。
	if rewriteWithEmployee && ch.AppName != "" && !strings.HasPrefix(strings.TrimSpace(text), "@") {
		text = "@" + ch.AppName + " " + text
	}

	// 取出暂存附件，与本条附件一起拼进同一条 WGA 消息
	pending := h.attachmentCache.Drain(msg.ChannelID, msg.UserID)
	content := buildWGAContentFromPending(text, pending, currentAtts)
	if len(pending) > 0 {
		log.Infof("[WGA] drained %d pending attachments for channel %s user %s",
			len(pending), msg.ChannelID, msg.UserID)
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

// placeholderContents 微信附件消息 Content 会填的占位文本，不作为文字指令。
// 微信文件消息 Content 填文件名（由 effectiveText 比对附件名排除），图片/语音/视频填这些占位符。
var placeholderContents = map[string]bool{
	"[图片]":   true,
	"[语音]":   true,
	"[视频]":   true,
	"[未知消息]": true,
}

// effectiveText 提取消息中的有效文字指令；无有效指令返回空串。
// 排除：空白、微信占位符、等于某个附件文件名的 Content（微信文件消息 Content=文件名，当附件名不当指令）。
func effectiveText(content string, attachments []types.Attachment) string {
	t := strings.TrimSpace(content)
	if t == "" || placeholderContents[t] {
		return ""
	}
	for _, a := range attachments {
		if t == strings.TrimSpace(a.Name) {
			return ""
		}
	}
	return t
}

// uploadAttachments 把附件上传到万悟 minio（/file/upload/direct），返回待用附件列表（含 minio URL）。
// 无附件返回 nil。供 doWGAChat 在纯附件分支存缓存、在有指令分支拼装 content 共用。
func uploadAttachments(ctx context.Context, wanwuClient *wanwu.Client, apiKey, channelID, userID string, attachments []types.Attachment) ([]*PendingAttachment, error) {
	if len(attachments) == 0 {
		return nil, nil
	}
	out := make([]*PendingAttachment, 0, len(attachments))
	for _, att := range attachments {
		uf, err := wanwuClient.UploadFile(ctx, apiKey, att.Name, att.MimeType, att.Data)
		if err != nil {
			return nil, fmt.Errorf("upload attachment %s failed: %w", att.Name, err)
		}
		out = append(out, &PendingAttachment{
			URL:      uf.FilePath,
			FileName: att.Name,
			MimeType: att.MimeType,
			FileId:   uf.FileId,
			FileSize: uf.FileSize,
		})
		log.Infof("[WGA] uploaded attachment %s (%d bytes) -> %s for channel %s user %s",
			att.Name, len(att.Data), uf.FilePath, channelID, userID)
	}
	return out, nil
}

// buildWGAContentFromPending 把文字指令与附件（暂存 + 本条）拼成 WGA 消息内容。
// 无附件时返回纯文本 string；有附件时返回多模态 []WGAMessageContentPart（先 text，再所有 binary）。
func buildWGAContentFromPending(text string, pending, current []*PendingAttachment) interface{} {
	if len(pending) == 0 && len(current) == 0 {
		return text
	}
	parts := make([]wanwu.WGAMessageContentPart, 0, len(pending)+len(current)+1)
	if text != "" {
		parts = append(parts, wanwu.WGAMessageContentPart{Type: "text", Text: text})
	}
	for _, a := range pending {
		parts = append(parts, wanwu.WGAMessageContentPart{
			Type:     "binary",
			MimeType: a.MimeType,
			URL:      a.URL,
			FileName: a.FileName,
		})
	}
	for _, a := range current {
		parts = append(parts, wanwu.WGAMessageContentPart{
			Type:     "binary",
			MimeType: a.MimeType,
			URL:      a.URL,
			FileName: a.FileName,
		})
	}
	return parts
}

// handleAgentSSEResponse 处理普通智能体 SSE 流式响应
func (h *Handler) handleAgentSSEResponse(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage, resp *http.Response) error {
	// 尝试创建流式发送器（支持流式卡片的平台会返回非 nil）
	streamSender := h.manager.CreateStreamSender(ctx, msg.ChannelID, msg.UserID, msg.Extra)

	var fullContent strings.Builder
	reader := bufio.NewReader(resp.Body)
	chunkCount := 0

	// 钉钉走 markdown 卡片整段下发（SendMarkdown，渲染 md），微信走纯文本一次性下发（SendMessage）。
	// 钉钉非流式路径：循环内按段落边界增量下发（攒够一段发一条卡片），缓解长回复干等；微信循环内只累积。
	isWeChat := ch.ChannelType == "wechat"

	// dingSt 仅钉钉非流式路径用（streamSender==nil && !isWeChat）：跟踪流式分段下发的累积状态。
	var dingSt dingStreamState

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
			fullContent.WriteString(sseData.Response)
			chunkCount++
			// 流式路径：逐 chunk 更新卡片
			if streamSender != nil {
				if err := streamSender.SendChunk(ctx, sseData.Response, false); err != nil {
					log.Errorf("[AgentSSE] channel=%s user=%s stream sender chunk failed, falling back to non-streaming: %v",
						ch.ChannelID, msg.UserID, err)
					// 流式发送失败，收尾卡片（置 failed）后降级为非流式
					closeStreamSender(streamSender, ctx, fmt.Errorf("stream chunk failed: %w", err))
					streamSender = nil
				}
			} else if !isWeChat {
				// 钉钉非流式：攒够一段（段落边界切）就增量下发，缓解生成期间干等。
				// 失败只 Warn，不中止生成（继续累积，循环后 isFinal 兜底重发剩余，含已失败段）。
				if err := h.flushDingTalkSegments(ctx, msg, fullContent.String(), &dingSt, false); err != nil {
					log.Warnf("[AgentSSE] channel=%s user=%s incremental segment send failed (will retry at end): %v",
						ch.ChannelID, msg.UserID, err)
				}
			}
			// 微信：循环内不下发（攒到最后 SendMessage 一次性，原逻辑）
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

	// ===== 非流式路径：循环结束后统一下发（streamSender == nil）=====
	// 钉钉：流式分段下发已在循环内增量发掉攒够的段，此处发剩余未发的尾巴段（isFinal 强制发全部剩余）
	// + 图片。微信：SendMessage 纯文本一次性下发（循环内只累积）。
	// 图片判断从 replyContent 算（无论是否已增量发，原图 URL 不变）。
	log.Infof("[AgentSSE] channel=%s user=%s stream completed, total %d chunks, reply length=%d, content: %s",
		ch.ChannelID, msg.UserID, chunkCount, len(replyContent), truncate(replyContent, 200))

	hasImg := inlineImageRe.MatchString(replyContent)

	if !isWeChat {
		// 钉钉：发剩余未下发的尾巴段（isFinal 强制发全部剩余，含循环内失败重发的段），再发图片。
		// flushDingTalkSegments 内部已对每段做 stripInvalidUTF8/stripInlineImages/stripCitations/TrimSpace，
		// 故此处不再需要循环外的 textToSend 后处理链。
		if dingSt.sentBytes < len(replyContent) {
			if err := h.flushDingTalkSegments(ctx, msg, replyContent, &dingSt, true); err != nil {
				return fmt.Errorf("failed to send final markdown segment: %w", err)
			}
		}
		if hasImg {
			h.sendInlineImages(ctx, msg, replyContent)
		}
		return nil
	}

	// 微信：纯文本一次性下发（原逻辑不变，保留 textToSend 后处理链）
	textToSend := replyContent
	if hasImg {
		textToSend = stripInlineImages(replyContent)
	}
	// 先清洗非法 UTF-8 字节（上游 LLM 偶发坏字节，被解码成 U+FFFD，IM 端显示 ��），再做引用剥离更稳。
	textToSend = stripInvalidUTF8(textToSend)
	// 知识库问答正文里的【x^】引用标注仅供网页端渲染来源脚注，IM 端不渲染会显示成字面乱码，发前剥离
	// （连带收敛剥离后的多余空白）。图片 URL 仍随下方 sendInlineImages 单独下发。
	textToSend = stripCitations(textToSend)
	textToSend = strings.TrimSpace(textToSend)
	if textToSend != "" {
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, textToSend, msg.Extra); err != nil {
			return fmt.Errorf("failed to send reply to platform: %w", err)
		}
		log.Infof("[AgentSSE] channel=%s user=%s reply text sent to platform successfully", msg.ChannelID, msg.UserID)
	}
	if hasImg {
		h.sendInlineImages(ctx, msg, replyContent)
	}

	return nil
}

// sendInlineImages 处理普通智能体回复正文里内嵌的 markdown 图片：下载每个图片 URL 并作为图片消息下发。
// 正文里的图片语法从 text 中剥离（避免当文字重复显示），返回剥离后的纯文本。
// 下载/发送失败只记日志，不影响文本下发（图片不丢失——URL 仍在万悟网页端可见）。
// 微信 png/jpg 走 image_item（SendFile 内部判定），AES+CDN 中转，复用 sendFileWithRetry 的 ret=-2 退避。
func (h *Handler) sendInlineImages(ctx context.Context, msg *types.PlatformMessage, text string) string {
	urls := inlineImageRe.FindAllStringSubmatch(text, -1)
	if len(urls) == 0 {
		return text
	}
	log.Infof("[AgentSSE] channel=%s user=%s found %d inline image(s) in reply, downloading", msg.ChannelID, msg.UserID, len(urls))

	for i, m := range urls {
		imgURL := m[1]
		data, err := downloadImage(ctx, imgURL)
		if err != nil {
			log.Warnf("[AgentSSE] channel=%s user=%s download inline image %d failed: %v (url=%s)",
				msg.ChannelID, msg.UserID, i+1, err, truncate(imgURL, 120))
			continue
		}
		fileName := inlineImageFileName(imgURL, i)
		mime := imageMimeTypeByExt(fileName)
		if err := h.sendFileWithRetry(ctx, msg, fileName, mime, data); err != nil {
			if errors.Is(err, types.ErrFileSendUnsupported) {
				// 平台不支持发文件（如飞书），图片无法下发，记日志即可（文本仍会发）
				log.Infof("[AgentSSE] channel=%s user=%s inline image %d not sent: platform unsupported",
					msg.ChannelID, msg.UserID, i+1)
				return stripInlineImages(text)
			}
			log.Warnf("[AgentSSE] channel=%s user=%s send inline image %d (%s) failed: %v",
				msg.ChannelID, msg.UserID, i+1, fileName, err)
			continue
		}
		log.Infof("[AgentSSE] channel=%s user=%s sent inline image %d (%s, %d bytes)",
			msg.ChannelID, msg.UserID, i+1, fileName, len(data))
		// 多图片间隔，避免密集推送撞微信 ret=-2 频控（同 WGA 工作区文件下发）
		if i < len(urls)-1 {
			select {
			case <-ctx.Done():
				return stripInlineImages(text)
			case <-time.After(workspaceFileSendGap):
			}
		}
	}
	return stripInlineImages(text)
}

// stripInlineImages 从文本中去掉 markdown 图片语法 ![alt](url)，避免微信 text_item 当文字显示。
func stripInlineImages(text string) string {
	return inlineImageRe.ReplaceAllString(text, "")
}

// deriveMarkdownTitle 从 markdown 内容生成钉钉卡片标题：取第一行非空文本，
// 去掉行首 # 标记与前后空白，截断 20 字（rune 计数，避免中文截半）。空内容兜底 "消息通知"。
// 钉钉 sampleMarkdown 的 title 字段必填，用于通知栏/会话列表预览。
func deriveMarkdownTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimSpace(strings.TrimLeft(line, "#"))
		if line == "" {
			continue
		}
		r := []rune(line)
		if len(r) > 20 {
			line = string(r[:20])
		}
		return line
	}
	return "消息通知"
}

// dingMarkdownMaxBytes 钉钉 sampleMarkdown 单条 msgParam body 上限 20000 bytes。
// content 经 JSON 转义会膨胀（\n→\\n、"→\" 等，中文不变），按 18000 留余量，
// 避免超限被钉钉 400 拒收（invalidParameter.msgParam.tooLong）。
const dingMarkdownMaxBytes = 18000

// dingMarkdownSegGap 钉钉多条 markdown卡片间的发送间隔，降低撞频控（130101/4001003）概率。
const dingMarkdownSegGap = 300 * time.Millisecond

// dingStreamSegmentBytes 钉钉非流式路径流式分段下发的「攒够阈值」。
// SSE 生成期间，未下发部分攒够该字节量就切出一条完整段落下发（不等全部生成完），缓解长回复干等。
// 必须 < dingMarkdownMaxBytes：若等于/超过，5000 字回复（~15000 bytes）永远攒不够一段，
// 会一路攒到 isFinal 一次性发，干等问题不解决。3000 bytes ≈ 1000 中文字，5000 字回复约 5 条卡片、
// 首条约 6 秒到达（按 167 字/秒估），兼顾实时性与刷屏。单段仍以 dingMarkdownMaxBytes 为上限防 20000 超限。
const dingStreamSegmentBytes = 3000

// dingStreamState 钉钉流式分段下发的累积状态。
// sentBytes 标记 fullContent 中已成功下发的前缀长度（失败段不前移，循环后 isFinal 兜底重发）；
// segCount 已下发段数，用于续号 title；baseTitle 缓存首段计算的标题（内容增长 title 不变，避免重复计算）。
type dingStreamState struct {
	sentBytes int
	segCount  int
	baseTitle string
	hasTitle  bool
}

// planDingStreamCuts 为钉钉流式分段下发计算切分点（pending 内的字节偏移，递增，最后一个 <= len(pending)）。
// 纯函数，不下发，便于表驱动测试。pending 为未下发部分（fullContent[sentBytes:]）。
//   - 非最终（!isFinal）：仅在攒够阈值（len(pending) >= streamBytes）时切，每段 ≤ maxBytes。先在
//     pending[:min(len,maxBytes)] 找最后一个 \n\n 切（段落边界，干净）；找不到且 pending >= 2*maxBytes
//     （极端长段落）按行（\n）兜底切；仍找不到则不切（留到 isFinal），避免切断正在生成的标题/行。
//     非最终不发没有 \n\n 收尾的尾巴段（可能还在生成），故只返回最多 1 个切分点。
//   - 最终（isFinal）：从 pending 头部逐段切到尾部，发出全部剩余（含尾巴）。每段 ≤ maxBytes：
//     段落边界优先 → 行兜底 → 字节硬切（UTF-8 边界回退）。
//
// 返回 nil 表示本轮不下发（短回复攒不够、或非最终未找到边界）。
func planDingStreamCuts(pending string, streamBytes, maxBytes int, isFinal bool) []int {
	if pending == "" {
		return nil
	}
	if !isFinal {
		// 非最终：攒够阈值才切，避免碎卡
		if len(pending) < streamBytes {
			return nil
		}
		// 段落边界优先：在前 maxBytes 范围内找最后一个 \n\n
		if cut := lastParagraphBoundary(pending, maxBytes); cut > 0 {
			return []int{cut}
		}
		// 极端长段落（pending >= 2*maxBytes 仍无 \n\n）：按行兜底切
		if len(pending) >= 2*maxBytes {
			if cut := lastLineBoundary(pending, maxBytes); cut > 0 {
				return []int{cut}
			}
		}
		// 单行超长且无边界（maxBytes 内无 \n）：不切，留到下个 chunk 或 isFinal，避免切断正在生成的内容
		return nil
	}
	// 最终：从头部逐段切到尾部，发全部剩余（含尾巴），每段 ≤ maxBytes
	var cuts []int
	pos := 0
	for pos < len(pending) {
		remain := pending[pos:]
		// 先尝试段落边界（切点 <= maxBytes）
		if cut := lastParagraphBoundary(remain, maxBytes); cut > 0 {
			pos += cut
			cuts = append(cuts, pos)
			continue
		}
		// 行兜底（单段落超 maxBytes）
		if cut := lastLineBoundary(remain, maxBytes); cut > 0 {
			pos += cut
			cuts = append(cuts, pos)
			continue
		}
		// 字节硬切：单行超 maxBytes。取 maxBytes 回退到 UTF-8 字符边界；不足 maxBytes 则取剩余全部（尾巴）
		end := maxBytes
		if end > len(remain) {
			end = len(remain)
		}
		for end < len(remain) && end > 0 && !utf8.RuneStart(remain[end]) {
			end--
		}
		if end == 0 {
			end = maxBytes // maxBytes 小于单字符，强制切避免死循环
			if end > len(remain) {
				end = len(remain)
			}
		}
		pos += end
		cuts = append(cuts, pos)
	}
	return cuts
}

// lastParagraphBoundary 返回 s[:min(len(s),limit)] 内最后一个 \n\n 的切分点（\n\n 之后的位置）。
// 找不到返回 0。切分点 <= limit。
func lastParagraphBoundary(s string, limit int) int {
	end := len(s)
	if end > limit {
		end = limit
	}
	idx := strings.LastIndex(s[:end], "\n\n")
	if idx < 0 {
		return 0
	}
	return idx + 2 // 跳过 \n\n，下一段从此处开始
}

// lastLineBoundary 返回 s[:min(len(s),limit)] 内最后一个 \n 的切分点（\n 之后的位置）。
// 找不到返回 0。
func lastLineBoundary(s string, limit int) int {
	end := len(s)
	if end > limit {
		end = limit
	}
	idx := strings.LastIndexByte(s[:end], '\n')
	if idx <= 0 {
		return 0
	}
	return idx + 1
}

// flushDingTalkSegments 从 fullContent[st.sentBytes:] 切出可下发的完整段并 SendMarkdown 下发。
// 钉钉非流式路径（!isWeChat && streamSender==nil）的流式分段下发：SSE 生成期间攒够一段就发一条 markdown
// 卡片，不等全部生成完，缓解长回复干等。
//
// 切分由 planDingStreamCuts 计算（段落边界优先，干净）；每个下发的段先 stripInvalidUTF8（字符级安全）
// + stripInlineImages + stripCitations（段在 \n\n/\n 边界切出，标记完整包含、不跨边界，安全）+ TrimSpace，
// 空段跳过。isFinal=false 时只发完整段（有边界收尾），不发可能还在生成的尾巴；isFinal=true 强制发全部剩余。
//
// sentBytes 按下发原文前移，但仅在该段 SendMarkdown 成功后才前移 + segCount++ —— 失败段不前移，
// 循环后 isFinal 会兜底重发（含已失败段）。返回发送中遇到的第一个 error（频控/网络），nil 表示正常。
func (h *Handler) flushDingTalkSegments(ctx context.Context, msg *types.PlatformMessage, fullContent string, st *dingStreamState, isFinal bool) error {
	pending := fullContent[st.sentBytes:]
	cuts := planDingStreamCuts(pending, dingStreamSegmentBytes, dingMarkdownMaxBytes, isFinal)
	if len(cuts) == 0 {
		return nil
	}
	// 首段计算并缓存标题（内容增长 title 不变）
	if !st.hasTitle {
		st.baseTitle = deriveMarkdownTitle(fullContent)
		st.hasTitle = true
	}
	sentThisRound := 0 // 本轮已成功下发字节数（用于段间间隔判断）
	for _, cut := range cuts {
		rawSeg := pending[sentThisRound:cut]
		sentThisRound = cut
		seg := stripInvalidUTF8(rawSeg)
		seg = stripInlineImages(seg)
		seg = stripCitations(seg)
		seg = strings.TrimSpace(seg)
		if seg == "" {
			// 空段（如纯引用/纯图片段 strip 后为空）：不下发，但仍前移 sentBytes 跳过，避免卡死
			st.sentBytes += len(rawSeg)
			continue
		}
		title := st.baseTitle
		if st.segCount > 0 {
			title = fmt.Sprintf("%s(续%d)", st.baseTitle, st.segCount+1)
		}
		if err := h.manager.SendMarkdown(ctx, msg.ChannelID, msg.UserID, title, seg); err != nil {
			return err
		}
		st.sentBytes += len(rawSeg)
		st.segCount++
		log.Infof("[AgentSSE] channel=%s user=%s incremental markdown segment %d sent (final=%v), %d chars, sentBytes=%d/%d",
			msg.ChannelID, msg.UserID, st.segCount, isFinal, len(seg), st.sentBytes, len(fullContent))
		// 段间短间隔（最后一段除外），降低撞钉钉频控（130101/4001003）概率
		if cut != cuts[len(cuts)-1] {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(dingMarkdownSegGap):
			}
		}
	}
	return nil
}

// sendWGADingTalkMarkdown 钉钉 WGA 正文段以 markdown 卡片下发（渲染 md），与普通 agent 路径一致。
// WGA 逐段实时下发，每段一张 md 卡片：title 取段首行（deriveMarkdownTitle，去#截20字，空兜底"消息通知"），
// content 为段原文（md 语法由钉钉 sampleMarkdown 渲染）。失败返回 error，调用方 Warn 不中止生成
// （与原 SendMessage 行为对齐——过程段丢失不阻断后续段/产物下发）。
func (h *Handler) sendWGADingTalkMarkdown(ctx context.Context, msg *types.PlatformMessage, segment string) error {
	title := deriveMarkdownTitle(segment)
	return h.manager.SendMarkdown(ctx, msg.ChannelID, msg.UserID, title, segment)
}

// downloadImage HTTP GET 下载图片字节。URL 为 minio 带签名直链（签名全在 query，无需鉴权 header），
// 用 URL 原始 host 下载（签名绑 host，不能替换）。channel-service 容器实测可达该 host。
func downloadImage(ctx context.Context, imgURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	return data, nil
}

// inlineImageFileName 从图片 URL 推断文件名（去掉 query 取最后路径段）；无扩展名时按 png 兜底。
// 用于发图时给 IM 一个带后缀的文件名，以便判定 mime 走 image_item。
func inlineImageFileName(imgURL string, idx int) string {
	name := imgURL
	if i := strings.IndexAny(name, "?#"); i >= 0 {
		name = name[:i]
	}
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if name == "" || !strings.Contains(name, ".") {
		name = fmt.Sprintf("image_%d.png", idx+1)
	}
	return name
}

// imageMimeTypeByExt 从文件名后缀推断图片 MIME（.png/.jpg/.jpeg/.gif/.webp），
// 非图片或未知返回 application/octet-stream（走 file_item）。微信 SendFile 据此判定 image_item vs file_item。
func imageMimeTypeByExt(fileName string) string {
	lower := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(lower, ".gif"):
		return "image/gif"
	case strings.HasSuffix(lower, ".webp"):
		return "image/webp"
	}
	return "application/octet-stream"
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
	// activityLines 收集各子智能体 finished 的进度行（🤖 子智能体: Xxx Agent (耗时)），
	// 不再逐个即时下发——一次分析常有 3-5 个子智能体，逐条发易撞微信 ret=-2 配额（约 9 条）。
	// 改为 RUN_FINISHED 后合并成一条发送，省 N-1 条配额，仍保留各子智能体耗时可见性。
	var activityLines []string
	// fullText 累积本次 SSE 流全部正文（跨段），供 sendWorkspaceFiles 做文件名主干兜底匹配
	// （PPT Agent 等正文只写标题不带扩展名，靠主干子串命中产物）。
	var fullText strings.Builder
	// deferredText + realtimeSegs：微信配额治理。微信 ilink sendmessage 配额约 9 条（撞后卡死，
	// 仅用户入站可解锁，见 wechat-ilink-ret2-quota），而 WGA 正文逐段下发（每段一条）会吃光配额，
	// 导致排在队尾的产物文件全 ret=-2、用户零投递。故微信路径正文只实时发前 N 段
	// （wechatRealtimeTextSegments），第 N+1 段起累积到 deferredText，RUN_FINISHED 后合并成一条发出，
	// 把配额留给产物文件。钉钉/飞书走流式卡片不占多条配额，仍逐段实时（isWeChat 为 false 时此机制不启用）。
	var deferredText strings.Builder
	realtimeSegs := 0
	isWeChat := ch.ChannelType == "wechat"
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
					fullText.WriteString(event.Delta)
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
					break
				}
				// 微信配额治理：微信只实时发前 N 段，第 N+1 段起累积到 deferredText，结束合并发。
				// 钉钉/飞书（isWeChat=false）始终逐段实时。
				if isWeChat && realtimeSegs >= wechatRealtimeTextSegments {
					deferredText.WriteString(segment)
					if !strings.HasSuffix(segment, "\n") {
						deferredText.WriteByte('\n')
					}
					continue
				}
				// 钉钉：markdown 卡片渲染 md（与普通 agent 路径一致，治 WGA 正文 md 语法裸露）；
				// 微信/飞书：纯文本 SendMessage。
				if ch.ChannelType == types.ChannelTypeDingTalk {
					if err := h.sendWGADingTalkMarkdown(ctx, msg, segment); err != nil {
						log.Warnf("[WGA-SSE] channel=%s user=%s send markdown segment failed: %v",
							ch.ChannelID, msg.UserID, err)
					} else {
						log.Infof("[WGA-SSE] channel=%s user=%s sent markdown segment (%d chars): %s",
							ch.ChannelID, msg.UserID, len(segment), truncate(segment, 50))
					}
					// 段间短间隔，降低撞钉钉频控（130101/4001003）概率
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(dingMarkdownSegGap):
					}
				} else if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, segment, msg.Extra); err != nil {
					log.Warnf("[WGA-SSE] channel=%s user=%s send text segment failed: %v",
						ch.ChannelID, msg.UserID, err)
				} else {
					log.Infof("[WGA-SSE] channel=%s user=%s sent text segment (%d chars): %s",
						ch.ChannelID, msg.UserID, len(segment), truncate(segment, 50))
				}
				if isWeChat {
					realtimeSegs++
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
				// 子智能体结束（sub_agent finished）：收集进度行，RUN_FINISHED 后合并下发
				// （逐个即时发易撞微信 ret=-2 配额）。question/workspace 快照不返回 completed。
				completed, _ := agg.handleEvent(wgaEv)
				if completed != nil && completed.kind == fragActivity {
					activityLines = append(activityLines, renderActivityLine(completed))
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
		if isWeChat {
			// 微信路径：残留段也并入 deferredText 合并发，不单独占一条配额。
			deferredText.WriteString(segment)
			if !strings.HasSuffix(segment, "\n") {
				deferredText.WriteByte('\n')
			}
		} else if ch.ChannelType == types.ChannelTypeDingTalk {
			// 钉钉：markdown 卡片渲染 md（与循环内逐段一致）
			if err := h.sendWGADingTalkMarkdown(ctx, msg, segment); err != nil {
				log.Warnf("[WGA-SSE] channel=%s user=%s send trailing markdown segment failed: %v",
					ch.ChannelID, msg.UserID, err)
			} else {
				log.Infof("[WGA-SSE] channel=%s user=%s sent trailing markdown segment (%d chars)",
					ch.ChannelID, msg.UserID, len(segment))
			}
		} else if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, segment, msg.Extra); err != nil {
			log.Warnf("[WGA-SSE] channel=%s user=%s send trailing text segment failed: %v",
				ch.ChannelID, msg.UserID, err)
		} else {
			log.Infof("[WGA-SSE] channel=%s user=%s sent trailing text segment (%d chars)",
				ch.ChannelID, msg.UserID, len(segment))
		}
		textBuf.Reset()
	}

	// 微信配额治理：把累积的第 N+1 段起的正文合并成一条发出（adapter 层自动分块，但仍尽量少占配额）。
	// 放在产物下发前：合并正文先到，产物紧随。非微信路径 deferredText 为空，跳过。
	if merged := strings.TrimRight(deferredText.String(), "\n"); merged != "" {
		if err := h.manager.SendMessage(ctx, msg.ChannelID, msg.UserID, merged, msg.Extra); err != nil {
			log.Warnf("[WGA-SSE] channel=%s user=%s send deferred text failed: %v",
				ch.ChannelID, msg.UserID, err)
		} else {
			log.Infof("[WGA-SSE] channel=%s user=%s sent deferred text (%d chars, %d realtime segs)",
				ch.ChannelID, msg.UserID, len(merged), realtimeSegs)
		}
	}

	// 合并下发收集到的子智能体进度行（逐个发易撞微信 ret=-2 配额，故合并成一条）。
	// 放在产物下发前：进度汇总先到，产物紧随，且合并只占 1 条配额而非 N 条。
	if len(activityLines) > 0 {
		sendProgress(strings.Join(activityLines, "\n"))
	}

	// 收尾聚合器（把未关闭的 activity 挂回顶层）；诊断未完成的过程 fragment（只记日志不下发，
	// 未完成的思考/工具调用内容不完整，下发会误导用户）。
	agg.finalize()
	for _, f := range agg.unfinishedToolCalls() {
		log.Warnf("[WGA-SSE] channel=%s user=%s unfinished tool_call: %s (id=%s), no RESULT received",
			ch.ChannelID, msg.UserID, f.toolCallName, f.toolCallID)
	}

	// 提取本次 run 实际产生的产物文件名（write 工具写过的文件，文件名来自 write args）。
	// 作为回发工作区文件的强信号，优先于正文文件名/stem 兜底，避免正文子串误命中历史文件。
	// write 产物补"智能体生成但不点名文件名"的漏发（文件名来自 write args，不依赖正文/ls）。
	producedFiles := extractProducedFiles(agg.topFragments)
	if len(producedFiles) > 0 {
		log.Infof("[WGA-SSE] channel=%s user=%s extracted produced files (write): %v",
			ch.ChannelID, msg.UserID, producedFiles)
	}

	// 把本次 run 的 write 产物持久化到 thread 级累积清单（跨 run "把报告发来"时复用）。
	// producedFiles 含 write + ls 产物，整批落库去重；累积清单只增不删，重启不丢。
	h.artifactMgr.AppendArtifacts(ctx, ch.ChannelID, msg.UserID, threadID, producedFiles)
	// 加载 thread 累积清单，作为跨 run 回发信号传入 sendWorkspaceFiles。
	accumulated := h.artifactMgr.ListArtifacts(ctx, ch.ChannelID, msg.UserID, threadID)

	if err := h.sendWorkspaceFiles(ctx, ch, msg, threadID, runID, mentionedFiles, fullText.String(), producedFiles, accumulated); err != nil {
		log.Warnf("[WGA-SSE] channel=%s user=%s send workspace files failed: %v",
			ch.ChannelID, msg.UserID, err)
	}
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

// inlineImageRe 匹配正文里的 markdown 图片语法 ![alt](url)。
// 智能体（普通 agent）知识库问答时会把图片以 markdown 嵌进 response 正文，URL 为 minio 带签名直链
// （如 http://172.25.67.233:8081/minio/download/.../xxx.png?X-Amz-Signature=...）。微信 text_item
// 不渲染 markdown，原样发出会显示成乱码文字且 URL 是内网地址用户打不开。故解析出图片单独下载下发，
// 并从正文去掉图片语法避免当文字重复显示。url 用非贪婪 [^)]+ 捕获到第一个 ) 前（minio 签名 URL 不含 )）。
var inlineImageRe = regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)`)

// citationRe 匹配知识库问答正文里的引用标注【x^】（agent-service 检索召回时给每条参考片段编号，
// 系统 prompt 要求模型在句末回引对应编号，如 【1^】【3^】）。该标记仅供万悟网页端渲染来源脚注，
// 微信 text_item 不渲染，原样下发显示成字面 【1^】 类似乱码，故发微信前剥离。
var citationRe = regexp.MustCompile(`【\s*\d+\s*\^】`)

// trailingSpaceRe 匹配行尾多余空白（含全角空格），剥离引用标注后裁掉句末残留空格。
var trailingSpaceRe = regexp.MustCompile(`[ \t　]+$`)

// multiBlankLineRe 匹配 3 个及以上连续换行（中间可有空白），压缩为 2 个换行，避免正文空洞。
var multiBlankLineRe = regexp.MustCompile(`\n[ \t]*(?:\n[ \t]*){2,}`)

// stripCitations 从正文中去掉知识库引用标注【x^】，并收敛剥离后残留的多余空白：
// 句末的标注剥离后留下尾随空格，行尾多余空格被裁掉，连续 3+ 空行压成 2 行，避免微信显示空行空洞。
func stripCitations(text string) string {
	text = citationRe.ReplaceAllString(text, "")
	text = trailingSpaceRe.ReplaceAllString(text, "")
	text = multiBlankLineRe.ReplaceAllString(text, "\n\n")
	return text
}

// stripInvalidUTF8 清洗正文里的 U+FFFD 替换字符。
// 上游 agent/BFF 的 LLM 流式输出偶发吐出非法 UTF-8 字节，channel-service 用 json.Unmarshal 解码时
// Go 会把非法字节替换成 U+FFFD（合法字符），微信端显示成 �� 问号块。strings.ToValidUTF8 此时无效
// （U+FFFD 已是合法字符），故直接把 U+FFFD 字符删掉：其余正常中文不受影响，仅个别坏字丢失。
// 根因在上游，此处为发 IM 前的兜底净化，避免微信端显示醒目乱码。
func stripInvalidUTF8(text string) string {
	return strings.ReplaceAll(text, "�", "")
}

// 微信 ilink sendmessage 在短时间密集推送时会返回 ret=-2（频控），留 2.5s 间隔降低撞频控概率。
const workspaceFileSendGap = 2500 * time.Millisecond

// wechatRealtimeTextSegments 微信 WGA 路径正文实时下发的段数上限。
// 微信 sendmessage 配额约 9 条（撞后卡死），WGA 正文逐段下发会吃光配额导致产物文件 ret=-2 零投递。
// 故只实时发前 N 段，第 N+1 段起累积合并成一条发（见 handleWGASSEResponse deferredText）。
// 配额算账（N=3）：3 实时段 + 1 合并段 + 1 进度汇总 + 1~3 文件 ≈ 6~8 条，留余量给文件。
// 钉钉/飞书走流式卡片不占多条配额，此上限不生效。
const wechatRealtimeTextSegments = 3

// sendFileWithRetry 发送工作区文件，按错误类型决定是否重试：
//   - ErrIMRateLimited（平台频控，如微信 ret=-2）：最多短退避重试 1 次（8s）；
//     微信 ret=-2 是配额耗尽型，撞满后卡死、仅用户入站消息可解锁，长时间退避（5s→15s→30s）
//     期间配额不会自发恢复，纯属空等，故不再 3 轮 50s 退避。保留 1 次 8s 短退避兜瞬时恢复
//     （恰好用户发入站解锁的场景），仍失败则立即返回让调用方走降级文本提示。
//   - ErrFileSendUnsupported：平台不支持，不重试，直接返回让调用方降级文本提示；
//   - 其他错误：不重试（多为永久性失败，退避无意义），直接返回。
func (h *Handler) sendFileWithRetry(ctx context.Context, msg *types.PlatformMessage, name, mime string, data []byte) error {
	err := h.manager.SendFile(ctx, msg.ChannelID, msg.UserID, name, mime, data, msg.Extra)
	if err == nil {
		return nil
	}
	if errors.Is(err, types.ErrFileSendUnsupported) {
		return err // 平台不支持，不重试
	}
	if !errors.Is(err, types.ErrIMRateLimited) {
		return err // 非频控的永久性错误，不重试
	}

	// 频控（微信 ret=-2 配额耗尽）：仅 1 次 8s 短退避兜瞬时恢复，不再长退避空等。
	log.Warnf("[WGA-WS] channel=%s user=%s send file %s rate-limited, retry 1/1 after 8s: %v",
		msg.ChannelID, msg.UserID, name, err)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(8 * time.Second):
	}
	err = h.manager.SendFile(ctx, msg.ChannelID, msg.UserID, name, mime, data, msg.Extra)
	if err == nil {
		return nil
	}
	if errors.Is(err, types.ErrFileSendUnsupported) {
		return err
	}
	return err // 仍频控（持续卡死，需用户入站解锁）或其他错误，放弃，走降级
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

// sendWorkspaceFiles 在 WGA 对话结束后，把产物下载并回发 IM。
// 工作区是 thread 级累积（含历史 run 的文件，无修改时间/runId 字段），无法靠快照 diff 区分本次新增
// （PPT Agent 用固定路径覆盖写，diff 永远判"无新增"会漏发）。按"本次产物 / 历史回发"两类匹配，命中即发：
//
//	本次产物（默认只发本次生成的）：
//	 1. 强信号 producedFiles：本次 run 智能体用 write 工具写过的文件名（basename 精确匹配）。
//	 2. 主路径 mentionedFiles：本次 SSE 智能体正文里带扩展名的文件名（basename 精确匹配）。
//	 3. 根目录高价值文档兜底：produced>0 且智能体正文提到文档主题词——覆盖 write 写脚本间接产出 .pptx 的场景。
//	    信号用智能体正文（本次产物的内容信号），produced>0 排除纯聊天。
//
//	历史回发（除非用户本轮明确点名，否则不发历史文件）：
//	 4. 累积清单 accumulated：仅当本次无强信号、且用户本轮 msg.Content 点名清单中文件主题/stem 时回发。
//	    覆盖"跨 run 把MaaS报告发来"。信号用用户本轮消息，不用智能体正文。
//	 5. 全工作区 stem：仅当本次无强信号、且用户本轮 msg.Content 点名文件 stem 时回发。
//	    信号用用户本轮消息——智能体分析正文里的多字词（如"日本"）不代表用户在要历史 日本.pptx。
//
// 纯聊天（用户未点名、无 produced）→ 历史回发不触发，本次产物也为空 → 不发任何文件，产物仍在网页端。
//
// 钉钉/微信实现 FileSender 真实发文件；飞书不实现，降级文本提示。任何失败只记日志，不影响已发文本。
func (h *Handler) sendWorkspaceFiles(ctx context.Context, ch *model.Channel, msg *types.PlatformMessage, threadID, runID string, mentionedFiles []string, fullText string, producedFiles []string, accumulated []string) error {
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
	// 跳过 node_modules 子树：PPT/前端 Agent 会装大量依赖，node_modules 下的 README/LICENSE/CHANGELOG
	// 等绝非用户产物，全收进 allFiles 既浪费又会污染 final artifacts 诊断日志；按 basename 匹配虽一般
	// 命不中，但累积清单（历史 write 过 README.md 等）可能误命中。从收集阶段就排除最干净。
	allFiles := make([]wgaFileItem, 0, ws.FileCount)
	var walk func(nodes []*wanwu.WGAFileNode, dir string)
	walk = func(nodes []*wanwu.WGAFileNode, dir string) {
		for _, n := range nodes {
			if n == nil {
				continue
			}
			// 跳过 node_modules 目录及其整个子树
			if n.Type == "dir" && strings.EqualFold(n.Name, "node_modules") {
				continue
			}
			curPath := joinWGAPath(dir, n.Name)
			if n.Type == "file" {
				// 兜底：路径中段含 node_modules 的也跳过（防御非标准层级）
				if !strings.Contains(curPath, "node_modules/") {
					allFiles = append(allFiles, wgaFileItem{name: n.Name, path: curPath, mime: n.MimeType, size: n.Size})
				}
			}
			if len(n.Children) > 0 {
				walk(n.Children, curPath)
			}
		}
	}
	walk(ws.Files, "")

	if len(allFiles) == 0 {
		log.Infof("[WGA-WS] channel=%s user=%s workspace has no file nodes", msg.ChannelID, msg.UserID)
		return nil
	}

	// 回发产物分两类：工作区是 thread 级历史累积（无修改时间、无 runId），全发会误推历史文件
	// 并触发 IM 频控，故按"本次产物 / 历史回发"分别精确匹配，命中即发：
	//  本次产物（默认只发本次生成的，信号用智能体行为/正文）：
	//   1. 强信号 producedFiles：本次 run 智能体用 write 工具写过的产物文件名（basename 精确匹配）。
	//      最可靠——直接反映本次 run 创建/修改的文件，不会把正文里偶然出现的多字词（如"日本"）
	//      误匹配成历史文件 日本.pptx。write 产物补"智能体生成但正文不点名"的漏发。
	//   2. 主路径 mentionedFiles：智能体正文出现带扩展名的完整文件名（basename 精确匹配）。
	//  历史回发（除非用户本轮明确点名，否则不发历史；信号用用户本轮 msg.Content，非智能体正文）：
	//   3. 累积清单 accumulated：仅当本次无强信号、且用户本轮点名清单中文件主题/stem 时回发。
	//      覆盖"跨 run 把MaaS报告发来"。不再无条件全发清单——纯聊天时用户不点名，自然不命中。
	//   4. 全工作区 stem：仅当本次无强信号、且用户本轮点名文件 stem 时回发。信号用用户本轮消息——
	//      智能体分析正文里的多字词不代表用户在要历史文件。
	//   5. 根目录高价值文档兜底：produced>0 且智能体正文提到文档主题词——本次产物性质，
	//      覆盖 write 写脚本间接产出 .pptx 的场景；produced>0 排除纯聊天。
	// 纯聊天（用户未点名、无 produced）→ 历史回发不触发，本次产物也为空 → 不发任何文件。
	producedSet := make(map[string]struct{}, len(producedFiles))
	for _, p := range producedFiles {
		producedSet[p] = struct{}{}
	}
	mentionedSet := make(map[string]struct{}, len(mentionedFiles))
	for _, m := range mentionedFiles {
		mentionedSet[m] = struct{}{}
	}
	accumulatedSet := make(map[string]struct{}, len(accumulated))
	for _, a := range accumulated {
		accumulatedSet[a] = struct{}{}
	}
	// 每个文件名只回发一份：工作区是递归目录树，同一文件名可能在不同目录出现多份
	// （如 output/西施.pptx 与子目录副本），按名字去重避免同一产物重复发送。
	matched := make(map[string]struct{})
	var files []wgaFileItem
	// produced 命中暂存：同 stem 不同扩展名（如 报告.md + 报告.html）只留高价值格式一份，
	// 避免同内容两格式重复下发、挤占 IM 配额。循环后按 stem 去重再并入 files。
	var producedHit []wgaFileItem
	// 第一轮：强信号 + 主路径（basename 精确匹配）
	for _, f := range allFiles {
		if !isFinalArtifact(f.name, f.mime) {
			continue
		}
		if _, dup := matched[f.name]; dup {
			continue
		}
		// 强信号 A：本次 write 工具写过的产物（直接写）——暂存，稍后按 stem 去重
		if _, ok := producedSet[f.name]; ok {
			producedHit = append(producedHit, f)
			matched[f.name] = struct{}{}
			continue
		}
		// 主路径：正文明确点名了带扩展名的文件名
		if _, ok := mentionedSet[f.name]; ok {
			files = append(files, f)
			matched[f.name] = struct{}{}
			continue
		}
	}
	// produced 按 stem 去重：同 stem 只留高价值格式一份（.html>.pdf>.docx>.pptx>.xlsx>.csv>.md>其他）。
	// 本次 write 的产物里常有"报告.md + 报告.html"同内容两格式，全发既重复又挤占配额。
	files = append(files, dedupeProducedByStem(producedHit)...)
	// 用户本轮请求文本：用于"历史回发"判定（②累积清单 / ③全工作区stem）。
	// 关键设计：历史文件回发只看"用户本轮是否明确点名"，不看智能体正文——
	// 智能体分析数据时正文提到"日本"是分析内容，不代表用户在要历史 日本.pptx，
	// 用智能体正文当信号会误发历史。本次产物（①⑤）仍用智能体 fullText，性质不同。
	// 去除空白后做子串匹配，避免换行/多空格导致"把MaaS报告发来"命不中。
	userText := stripWhitespace(msg.Content)

	// 第二轮：历史回发——累积清单里、用户本轮 Content 明确点名的才发。
	// 覆盖"用户跨 run 说把MaaS报告发来"：用户本轮点名文件名主题/主干，命中累积清单里的历史产物。
	// 不再无条件全发清单——纯聊天("你能做什么")时 Content 不含任何文件名，自然不命中，不发历史。
	// 命中口径与⑤轮对齐：文件名 topic 首段（"MaaS平台…"→"MaaS"）或完整 stem 子串出现在 Content。
	if len(files) == 0 && len(accumulatedSet) > 0 && userText != "" {
		for _, f := range allFiles {
			if !isFinalArtifact(f.name, f.mime) {
				continue
			}
			if _, dup := matched[f.name]; dup {
				continue
			}
			if _, ok := accumulatedSet[f.name]; !ok {
				continue
			}
			if userMentionsFile(userText, f.name) {
				files = append(files, f)
				matched[f.name] = struct{}{}
			}
		}
	}
	// 第三轮：历史回发——全工作区里、用户本轮 Content 点名 stem 的才发。
	// 信号源从智能体正文换成用户 Content：避免智能体分析正文里的多字词（如"日本"）把历史
	// 日本.pptx 误拖出来发。本次产物的 PPT 场景由①⑤(produced)覆盖，不依赖本路径。
	if len(files) == 0 && userText != "" {
		for _, f := range allFiles {
			if !isFinalArtifact(f.name, f.mime) {
				continue
			}
			if _, dup := matched[f.name]; dup {
				continue
			}
			if userMentionsFile(userText, f.name) {
				files = append(files, f)
				matched[f.name] = struct{}{}
			}
		}
	}
	// 第五轮：根目录高价值文档兜底（收紧版——正文 stem 匹配）。
	// PPT/文档类 Agent 的最终产物（.pptx/.docx 等）常由 write 写的脚本间接生成（node 执行脚本产出），
	// 既非 write 工具直接写（→强信号抓不到），正文也不一定点名（→主路径/stem 抓不到），前 4 轮可能全 miss。
	// 这类 Agent 习惯把最终文档放在工作区根目录（path 不含 /），与 node_modules/、slides/output/ 等
	// 噪音目录天然隔离。故在前 4 轮全 miss 且本轮确有 write 活动（produced>0，排除纯聊天）时，
	// 在根目录高价值文档中，只发本轮正文 stem 能匹配上的那一个（如正文含"刘备"→只发刘备 PPT）。
	//
	// 为什么必须叠正文 stem：根目录会累积多个 run 的历史 PPT（刘备/北京/李信…），workspace 无时间戳
	// 无法区分"本次新增"。若全发会把历史 PPT 一起误推（曾发生：让生成刘备 PPT 却连发北京、李信）。
	// 叠加正文 stem 把候选缩到本轮主题那一个，避免误发历史。代价：正文不点名主题时放弃兜底（宁漏不误），
	// 产物仍在万悟工作区网页端可见。
	if len(files) == 0 && len(producedFiles) > 0 && fullText != "" {
		var rootDocCandidates []string
		for _, f := range allFiles {
			if !isFinalArtifact(f.name, f.mime) {
				continue
			}
			if _, dup := matched[f.name]; dup {
				continue
			}
			// 仅根目录文件（path 不含路径分隔符），排除 node_modules/、output/ 等子目录噪音
			if strings.Contains(f.path, "/") {
				continue
			}
			if !isHighValueDocument(f.name) {
				continue
			}
			rootDocCandidates = append(rootDocCandidates, f.name)
			// 正文必须提到该文档的主题词（文件名首段，去扩展名后按 -/_/—— 拆分取首段）。
			// PPT Agent 正文一般只提主题（如"刘备"）不提完整文件名（"刘备-蜀汉昭烈帝.pptx"），
			// 故用首段匹配而非完整 stem。首段需 stemAcceptable（中文≥2、拉丁≥3）防短词误命中。
			topic := fileTopicToken(f.name)
			if topic != "" && strings.Contains(fullText, topic) {
				files = append(files, f)
				matched[f.name] = struct{}{}
			}
		}
		// 诊断：前 4 轮全 miss 时，记录根目录高价值文档候选 + 正文摘要，用于排查 PPT 漏发/误发。
		if len(files) == 0 {
			log.Infof("[WGA-WS] channel=%s user=%s root-doc fallback miss: candidates=%v, fullText=%s",
				msg.ChannelID, msg.UserID, rootDocCandidates, truncate(fullText, 200))
		}
	}
	log.Infof("[WGA-WS] channel=%s user=%s matched %d file(s) in workspace (produced=%d, mentioned=%d, accumulated=%d, fallbackStem=%v)",
		msg.ChannelID, msg.UserID, len(files), len(producedFiles), len(mentionedFiles), len(accumulated), len(files) == 0)
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
	sentNames := make([]string, 0, len(files)) // 实际发送成功的文件名（汇总只列已发的，避免列了没发出的误导用户）
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
			// 频控（微信 ret=-2 配额耗尽，撞后卡死、仅用户入站可解锁）：后续文件必失败，
			// 继续发只会逐个空等 8s 重试 + 刷失败日志。立即中止，走降级文本提示。
			if errors.Is(err, types.ErrIMRateLimited) {
				log.Warnf("[WGA-WS] channel=%s user=%s rate-limited, abort remaining %d file(s)",
					msg.ChannelID, msg.UserID, len(files)-i-1)
				break
			}
			continue
		}
		sent++
		sentNames = append(sentNames, f.name)
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
	// 只列实际发送成功的（sentNames）——ret=-2 频控 break 时后续文件未发，列全部会误导用户以为都收到了。
	if len(sentNames) > 0 {
		tip := "✅ 已生成：" + strings.Join(sentNames, "、")
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
// 误推给用户并触发 IM 频控。现改为：主路径回发正文明确点名的带扩展名文件名；兜底回发正文里
// 出现文件名主干（去扩展名）的产物（见 sendWorkspaceFiles）——主干命中比关键词匹配精确得多。

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
		"xml", "svg", "map", "lock", "log", "tmp", "bak", "swp",
		// 文本/配置元数据类：多为 node_modules 噪音（LICENSE/CHANGELOG/.editorconfig 等），
		// 过滤掉避免误发。注意 .md/.markdown 不在此列——调研报告等用户产物常以 .md 形式存在，
		// node_modules 的 README.md 噪音由 allFiles 收集时跳过 node_modules 路径排除（见 walk）。
		"txt", "rst", "textile",
		"editorconfig", "npmignore", "jekyll-metadata", "gitignore", "gitattributes",
		"license", "licence", "changes", "changelog", "history", "sponsors",
		"npmrc", "yarnrc":
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

// highValueDocExts 是用户最终文档产物的高价值扩展名白名单。
// PPT/文档类 Agent 的最终产物通常是这些格式（pptxgenjs 出 .pptx、python-docx 出 .docx 等），
// 与 .md/.txt 等中间文本区分开，作为根目录兜底下发的判定依据。
var highValueDocExts = map[string]struct{}{
	"pptx": {}, "ppt": {},
	"docx": {}, "doc": {},
	"xlsx": {}, "xls": {},
	"pdf": {},
	"csv": {},
}

// isHighValueDocument 判定文件是否为高价值最终文档（.pptx/.docx/.xlsx/.pdf/.csv 等）。
// 用于根目录兜底：只在产物是这类用户文档时下发，避免把根目录偶然的脚本/配置当产物误发。
func isHighValueDocument(name string) bool {
	ext := strings.ToLower(strings.TrimPrefix(filepathExt(name), "."))
	_, ok := highValueDocExts[ext]
	return ok
}

// producedExtPriority produced 产物同 stem 去重时的扩展名优先级（值大者优先保留）。
// 智能体常同时输出"报告.md + 报告.html"同内容两格式，用户可直接打开的（html/pdf/docx）优于源文件（md）。
// html 居首（浏览器直接看，且调研报告 Agent 习惯出 html）；md 居末（源文件，需渲染）。
var producedExtPriority = map[string]int{
	"html": 100, "htm": 99,
	"pdf":  90,
	"docx": 80, "doc": 79,
	"pptx": 70, "ppt": 69,
	"xlsx": 60, "xls": 59,
	"csv": 50,
	"md":  40,
	"txt": 30,
}

// producedExtScore 取文件扩展名的优先级分，未列入的扩展名给 1（最低，仍可保留为同 stem 唯一者）。
func producedExtScore(name string) int {
	ext := strings.ToLower(strings.TrimPrefix(filepathExt(name), "."))
	if p, ok := producedExtPriority[ext]; ok {
		return p
	}
	return 1
}

// dedupeProducedByStem 把 produced 命中的文件按 stem（去扩展名）分组，每组只留扩展名优先级最高的一份。
// 避免同内容两格式（报告.md + 报告.html）重复下发、挤占 IM 配额。不同 stem 互不影响，全部保留。
func dedupeProducedByStem(items []wgaFileItem) []wgaFileItem {
	if len(items) <= 1 {
		return items
	}
	best := make(map[string]wgaFileItem, len(items))
	for _, f := range items {
		stem := strings.TrimSuffix(f.name, filepathExt(f.name))
		cur, ok := best[stem]
		if !ok || producedExtScore(f.name) > producedExtScore(cur.name) {
			best[stem] = f
		}
	}
	out := make([]wgaFileItem, 0, len(best))
	for _, f := range best {
		out = append(out, f)
	}
	return out
}

// fileTopicToken 取文件名的主题词（去扩展名后按分隔符拆分取首段）。
// PPT/文档 Agent 常用"主题-副标题.pptx"命名（如 刘备-蜀汉昭烈帝.pptx、北京-千年古都-现代之城.pptx），
// 正文一般只提主题"刘备"不提完整文件名。取首段作为正文匹配的 token，缩窄根目录兜底候选到本轮主题。
// 首段需 stemAcceptable（中文≥2、拉丁/数字≥3），过短则返回空（不参与匹配，避免 a.pptx 误命中）。
func fileTopicToken(name string) string {
	stem := strings.TrimSuffix(name, filepathExt(name))
	// 按 - _ —— 空格 拆分，取第一段
	for _, sep := range []string{"——", "-", "_", " "} {
		if i := strings.Index(stem, sep); i > 0 {
			stem = stem[:i]
			break
		}
	}
	stem = strings.TrimSpace(stem)
	if !stemAcceptable(stem) {
		return ""
	}
	return stem
}

// stripWhitespace 删除字符串中所有空白字符（含换行/制表/全角空格），用于把用户消息
// 折叠成无空白串做子串匹配——用户"把 MaaS 报告 发来"含空格/换行，折叠后才与文件名主题命中。
func stripWhitespace(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == '\v' || r == '\f' || r == '　' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// userMentionsFile 判定用户本轮消息文本（已 stripWhitespace）是否点名了文件 fileName。
// 用于"历史回发"判定（②累积清单 / ③全工作区stem）：用户本轮明确要历史文件才回发，
// 智能体正文不参与此判定。命中口径与⑤轮对齐：
//   - 文件名 topic 首段（去扩展名后按 -/_/—— 拆分取首段，如"MaaS平台微服务及端口调研报告.md"→"MaaS平台微服务及端口调研报告"
//     再取首段"MaaS平台微服务及端口调研报告"——实际按分隔符拆，见 fileTopicToken）子串出现在用户文本；
//   - 或完整 stem（去扩展名，如"MaaS平台微服务及端口调研报告"）子串出现。
//
// topic/stem 均需 stemAcceptable（中文≥2、拉丁≥3），过短不参与匹配，避免短词误命中历史。
func userMentionsFile(userText, fileName string) bool {
	if userText == "" || fileName == "" {
		return false
	}
	stem := strings.TrimSuffix(fileName, filepathExt(fileName))
	if stemAcceptable(stem) && strings.Contains(userText, stem) {
		return true
	}
	if topic := fileTopicToken(fileName); topic != "" && strings.Contains(userText, topic) {
		return true
	}
	return false
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

// stemAcceptable 判定文件名主干（去扩展名）是否足够特异，可在正文里做子串兜底匹配。
// 中文字符信息密度高，两字（如 台湾/老舍/美元）已足够特异，阈值放低到 2；
// 拉丁/数字字符信息密度低，保持 ≥3，避免 a.pdf→"a"、go 这类短主干在正文误命中历史文件。
// 判定依据：主干中是否含 CJK 字符（含中日韩统一表意及扩展区）。
func stemAcceptable(stem string) bool {
	runes := []rune(stem)
	if len(runes) == 0 {
		return false
	}
	for _, r := range runes {
		if isCJKRune(r) {
			return len(runes) >= 2
		}
	}
	return len(runes) >= 3
}

// isCJKRune 判定 rune 是否为 CJK 字符（中日韩统一表意文字主区及扩展A区，
// 覆盖常见中文文件名用字）。扩展B及以上区罕见用于文件名，未纳入。
func isCJKRune(r rune) bool {
	return (r >= 0x4E00 && r <= 0x9FFF) || // CJK 统一表意文字主区
		(r >= 0x3400 && r <= 0x4DBF) || // CJK 扩展A
		(r >= 0xF900 && r <= 0xFAFF) // CJK 兼容表意
}
