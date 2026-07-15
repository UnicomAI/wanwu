package dingtalk

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// DingTalkOpenAPIEndpoint 钉钉开放平台 API 地址
	DingTalkOpenAPIEndpoint = "https://api.dingtalk.com"
	// OpenConnectionAPI 开启 Stream 连接的 API
	OpenConnectionAPI = DingTalkOpenAPIEndpoint + "/v1.0/gateway/connections/open"
	// ChatbotMessageTopic 机器人消息 Topic
	ChatbotMessageTopic = "/v1.0/im/bot/messages/get"
)

// StreamClient 钉钉 Stream 模式客户端
// 使用 WebSocket 长连接接收钉钉消息推送，无需公网回调 URL
// 参考 dingtalk-stream 协议：https://open.dingtalk.com/document/orgapp/stream-mode-protocol
type StreamClient struct {
	appKey    string
	appSecret string
	channelID string // 渠道 ID，用于消息路由

	accessToken     string
	tokenExpiry     time.Time
	accessTokenLock sync.RWMutex

	// 旧版 OAPI Token（GET /gettoken），用于媒体/文件上传接口（oapi.dingtalk.com/media/upload、
	// /file/upload/*）。与上面的新版 API Token（/v1.0/oauth2/accessToken，用于发送消息）是两套，不能混用。
	oapiToken       string
	oapiTokenExpiry time.Time
	oapiTokenLock   sync.RWMutex

	// WebSocket 连接
	conn      *websocket.Conn
	connLock  sync.Mutex
	connected bool
	stopChan  chan struct{}

	// 消息处理器
	messageHandler MessageHandler

	// HTTP 客户端
	httpClient *http.Client

	// 状态
	status string

	// 消息去重：记录最近处理过的消息，防止重复处理
	// 钉钉 Stream 协议在机器人未及时回复时会重发消息，且重发时 messageId 可能不同
	// 因此同时按 messageId 和 senderID+content 两个维度去重
	seenMsgIDs    sync.Map // key: messageId, value: time.Time
	seenMsgDigest sync.Map // key: "senderID:content", value: time.Time — 内容去重（兜底 messageId 变化的重发）
}

// StreamConfig Stream 客户端配置
type StreamConfig struct {
	AppKey    string
	AppSecret string
	ChannelID string // 可选，渠道 ID
}

// NewStreamClient 创建 Stream 客户端
func NewStreamClient(config *StreamConfig) *StreamClient {
	return &StreamClient{
		appKey:     config.AppKey,
		appSecret:  config.AppSecret,
		channelID:  config.ChannelID,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		status:     "offline",
		stopChan:   make(chan struct{}),
	}
}

// OnMessage 注册消息处理器
func (c *StreamClient) OnMessage(handler MessageHandler) {
	c.messageHandler = handler
}

// Start 启动 Stream 连接（非阻塞）
func (c *StreamClient) Start(ctx context.Context) error {
	go c.run(ctx)
	return nil
}

// run 主循环：连接 → 处理消息 → 断线重连
func (c *StreamClient) run(ctx context.Context) {
	var authFailCount int
	const maxAuthFailRetries = 3

	for {
		select {
		case <-ctx.Done():
			log.Infof("[DingTalk Stream] Context cancelled, stopping...")
			return
		case <-c.stopChan:
			log.Infof("[DingTalk Stream] Stop signal received, stopping...")
			return
		default:
			err := c.connect(ctx)
			if err != nil {
				// 鉴权失败（401/authFailed）不应无限重试，避免刷屏日志
				if isAuthError(err) {
					authFailCount++
					if authFailCount >= maxAuthFailRetries {
						log.Errorf("[DingTalk Stream] Authentication failed %d times, stopping reconnect. "+
							"Please check appKey/appSecret configuration. err: %v", authFailCount, err)
						c.connLock.Lock()
						c.status = "auth_failed"
						c.connLock.Unlock()
						return
					}
					backoff := time.Duration(authFailCount*10) * time.Second
					log.Errorf("[DingTalk Stream] Authentication failed (%d/%d), retrying in %v... err: %v",
						authFailCount, maxAuthFailRetries, backoff, err)
					time.Sleep(backoff)
					continue
				}

				// 非鉴权错误，正常重试并重置鉴权计数
				authFailCount = 0
				log.Errorf("[DingTalk Stream] Connection error: %v, reconnecting in 5s...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			// 连接成功，重置鉴权失败计数
			authFailCount = 0

			// 连接成功，处理消息
			c.handleMessages(ctx)
			log.Infof("[DingTalk Stream] Connection closed, reconnecting in 3s...")
			time.Sleep(3 * time.Second)
		}
	}
}

// isAuthError 判断是否为鉴权类错误（不可通过重试恢复）
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "status=401") ||
		strings.Contains(msg, "authFailed") ||
		strings.Contains(msg, "invalidClient") ||
		strings.Contains(msg, "invalid_client")
}

// connect 建立 WebSocket 连接
func (c *StreamClient) connect(ctx context.Context) error {
	// 1. 调用开放平台 API 获取连接信息
	connection, err := c.openConnection()
	if err != nil {
		return fmt.Errorf("open connection failed: %w", err)
	}

	if connection == nil || connection.Endpoint == "" || connection.Ticket == "" {
		return fmt.Errorf("invalid connection response: endpoint or ticket is empty")
	}

	// 2. 建立 WebSocket 连接
	wsURL := fmt.Sprintf("%s?ticket=%s", connection.Endpoint, connection.Ticket)

	dialer := websocket.Dialer{
		HandshakeTimeout: 30 * time.Second,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: false},
	}

	conn, _, err := dialer.DialContext(ctx, wsURL, nil)
	if err != nil {
		return fmt.Errorf("websocket dial failed: %w", err)
	}

	c.connLock.Lock()
	c.conn = conn
	c.connected = true
	c.status = "online"
	c.connLock.Unlock()

	log.Infof("[DingTalk Stream] Connected to %s", connection.Endpoint)

	// 3. 启动心跳
	go c.keepalive(ctx)

	return nil
}

// openConnection 调用开放平台 API 获取 WebSocket 连接信息
func (c *StreamClient) openConnection() (*ConnectionResponse, error) {
	reqBody := map[string]interface{}{
		"clientId":     c.appKey,
		"clientSecret": c.appSecret,
		"subscriptions": []map[string]string{
			{"type": "CALLBACK", "topic": ChatbotMessageTopic},
		},
	}

	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", OpenConnectionAPI, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open connection failed: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result ConnectionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse connection response failed: %w", err)
	}

	return &result, nil
}

// handleMessages 处理 WebSocket 消息循环
func (c *StreamClient) handleMessages(ctx context.Context) {
	defer func() {
		c.connLock.Lock()
		if c.conn != nil {
			_ = c.conn.Close()
			c.conn = nil
		}
		c.connected = false
		c.status = "offline"
		c.connLock.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		default:
			c.connLock.Lock()
			conn := c.conn
			c.connLock.Unlock()

			if conn == nil {
				return
			}

			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Errorf("[DingTalk Stream] Read message error: %v", err)
				return
			}

			go c.processMessage(ctx, message)
		}
	}
}

// processMessage 处理单条消息
func (c *StreamClient) processMessage(ctx context.Context, message []byte) {
	var msg StreamMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Errorf("[DingTalk Stream] Parse message failed: %v, body=%s", err, string(message))
		return
	}

	switch msg.Type {
	case "SYSTEM":
		c.handleSystemMessage(ctx, message)
	case "CALLBACK":
		c.handleCallbackMessage(ctx, message)
	default:
		log.Warnf("[DingTalk Stream] Unknown message type: %s", msg.Type)
	}
}

// handleSystemMessage 处理系统消息（需要回复 ACK）
func (c *StreamClient) handleSystemMessage(ctx context.Context, message []byte) {
	var sysMsg SystemMessage
	if err := json.Unmarshal(message, &sysMsg); err != nil {
		log.Errorf("[DingTalk Stream] Parse system message failed: %v", err)
		return
	}

	log.Infof("[DingTalk Stream] System message: topic=%s", sysMsg.Headers.Topic)

	// 发送 ACK
	ack := AckMessage{
		Code:    200,
		Message: "OK",
		Headers: AckHeaders{
			MessageID:   sysMsg.Headers.MessageID,
			ContentType: "application/json",
		},
	}

	c.sendAck(ack)

	// 处理断连消息
	if sysMsg.Headers.Topic == "disconnect" {
		log.Warnf("[DingTalk Stream] Received disconnect message, will reconnect...")
	}
}

// handleCallbackMessage 处理回调消息（机器人消息）
func (c *StreamClient) handleCallbackMessage(ctx context.Context, message []byte) {
	var callbackMsg struct {
		Type    string          `json:"type"`
		Headers CallbackHeaders `json:"headers"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(message, &callbackMsg); err != nil {
		log.Errorf("[DingTalk Stream] Parse callback message failed: %v", err)
		return
	}

	log.Infof("[DingTalk Stream] Received callback: topic=%s, messageId=%s",
		callbackMsg.Headers.Topic, callbackMsg.Headers.MessageID)

	// 只处理机器人消息
	if callbackMsg.Headers.Topic != ChatbotMessageTopic {
		log.Warnf("[DingTalk Stream] Unknown callback topic: %s", callbackMsg.Headers.Topic)
		return
	}

	// 解析 data（可能是 JSON 字符串或对象）
	var dataMap map[string]interface{}
	if len(callbackMsg.Data) > 0 {
		if callbackMsg.Data[0] == '"' {
			// data 是 JSON 字符串，需要先解包
			var dataStr string
			if err := json.Unmarshal(callbackMsg.Data, &dataStr); err != nil {
				log.Errorf("[DingTalk Stream] Parse data string failed: %v", err)
			} else {
				if err := json.Unmarshal([]byte(dataStr), &dataMap); err != nil {
					log.Errorf("[DingTalk Stream] Parse data content failed: %v, content: %s", err, dataStr)
				}
			}
		} else {
			if err := json.Unmarshal(callbackMsg.Data, &dataMap); err != nil {
				log.Errorf("[DingTalk Stream] Parse data object failed: %v", err)
			}
		}
	}

	// 发送 ACK
	ack := AckMessage{
		Code:    200,
		Message: "OK",
		Headers: AckHeaders{
			MessageID:   callbackMsg.Headers.MessageID,
			ContentType: "application/json",
		},
		Data: dataMap,
	}
	c.sendAck(ack)

	// 转换为统一消息格式并调用处理器
	if c.messageHandler != nil && dataMap != nil {
		msg := c.convertToMessage(dataMap)
		// 设置 ChannelID - 优先使用注册时设置的 channelID
		msg.ChannelID = c.channelID

		// 消息去重（双层）：钉钉 Stream 协议在机器人未及时回复时会重发消息，
		// 且 Stream 层 messageId（callbackMsg.Headers.MessageId）重发时会变化，不能用于去重；
		// IM 层 msgId（data.msgId，即 msg.MessageID）重发时稳定不变，作为主去重键。
		isDuplicate := false

		// 第一层：按 IM 层 msgId 去重（重发时稳定不变）；msgId 缺失时回退到 Stream 层 messageId
		msgID := msg.MessageID
		if msgID == "" {
			msgID = callbackMsg.Headers.MessageID
		}
		if msgID != "" {
			if _, seen := c.seenMsgIDs.Load(msgID); seen {
				log.Warnf("[DingTalk Stream] Duplicate message detected (by messageId), skipping: messageId=%s, sender=%s, content=%s",
					msgID, msg.Sender, truncate(msg.Content, 50))
				isDuplicate = true
			}
		}

		// 第二层：按 senderID+content 去重（兜底 msgId 缺失或变化的重发）
		if !isDuplicate && msg.Sender != "" && msg.Content != "" {
			digestKey := msg.Sender + ":" + msg.Content
			if lastTime, seen := c.seenMsgDigest.Load(digestKey); seen {
				if time.Since(lastTime.(time.Time)) < 120*time.Second {
					log.Warnf("[DingTalk Stream] Duplicate message detected (by content), skipping: messageId=%s, sender=%s, content=%s",
						msgID, msg.Sender, truncate(msg.Content, 50))
					isDuplicate = true
				}
			}
		}

		if isDuplicate {
			return
		}

		// 记录去重
		if msgID != "" {
			c.seenMsgIDs.Store(msgID, time.Now())
		}
		if msg.Sender != "" && msg.Content != "" {
			c.seenMsgDigest.Store(msg.Sender+":"+msg.Content, time.Now())
		}
		go c.cleanupSeenMsgIDs()

		log.Infof("[DingTalk Stream] Received message from %s, content=%s, channelID=%s, sessionWebhook=%s, messageId=%s",
			msg.Sender, truncate(msg.Content, 100), msg.ChannelID, truncate(msg.SessionWebhook, 50), msg.MessageID)

		// 立即通过 sessionWebhook 发送"思考中"占位回复，防止钉钉超时重发
		if msg.SessionWebhook != "" {
			if err := c.ReplyWithWebhook(ctx, msg.SessionWebhook, "思考中...", nil); err != nil {
				log.Warnf("[DingTalk Stream] Failed to send thinking placeholder: %v", err)
			} else {
				log.Infof("[DingTalk Stream] Sent thinking placeholder via sessionWebhook for messageId=%s", msg.MessageID)
			}
		}

		go func() {
			// 若为图片/文件消息，先下载附件字节填入 Attachments（供 chat handler 上传给 WGA）
			if dc, ok := msg.Raw["downloadCode"].(string); ok && dc != "" {
				data, dErr := c.DownloadFile(ctx, dc)
				if dErr != nil {
					log.Errorf("[DingTalk Stream] Failed to download attachment for messageId=%s: %v", msg.MessageID, dErr)
				} else {
					msg.Attachments = append(msg.Attachments, Attachment{
						Name:     "dingtalk-file",
						MimeType: http.DetectContentType(data),
						Data:     data,
					})
					log.Infof("[DingTalk Stream] Downloaded attachment for messageId=%s, size=%d", msg.MessageID, len(data))
				}
			}
			if err := c.messageHandler(ctx, msg); err != nil {
				log.Errorf("[DingTalk Stream] Message handler error: %v", err)
			}
		}()
	}
}

// convertToMessage 转换钉钉消息为统一格式
func (c *StreamClient) convertToMessage(data map[string]interface{}) *Message {
	msg := &Message{
		Raw:       data,
		Timestamp: time.Now(),
	}

	// 提取消息类型
	if msgType, ok := data["msgtype"].(string); ok {
		msg.MsgType = MessageType(msgType)
	}

	// 从 text 字段提取内容
	if textData, ok := data["text"].(map[string]interface{}); ok {
		if content, ok := textData["content"].(string); ok {
			msg.Content = content
		}
	}

	// 如果还没有内容，尝试从 content 字段提取
	if msg.Content == "" {
		if content, ok := data["content"].(map[string]interface{}); ok {
			if text, ok := content["text"].(string); ok {
				msg.Content = text
			}
			// 处理富文本
			if richText, ok := content["richText"].([]interface{}); ok {
				var texts []string
				for _, item := range richText {
					if m, ok := item.(map[string]interface{}); ok {
						if text, ok := m["text"].(string); ok && text != "\n" {
							texts = append(texts, text)
						}
					}
				}
				if len(texts) > 0 {
					msg.Content = texts[0]
					msg.MsgType = MessageTypeMarkdown
				}
			}
			// 处理图片/文件：钉钉图片和文件消息均带 downloadCode，Content 留空，
			// 实际文件字节在 handleCallbackMessage 中通过 DownloadFile 下载后填入 msg.Attachments。
			if downloadCode, ok := content["downloadCode"].(string); ok {
				msg.Content = ""
				msg.MsgType = MessageTypeImage
				if msg.Raw == nil {
					msg.Raw = make(map[string]interface{})
				}
				msg.Raw["downloadCode"] = downloadCode
			}
		}
	}

	// 提取消息 ID（用于去重）
	if msgID, ok := data["msgId"].(string); ok {
		msg.MessageID = msgID
	}

	// 提取发送者信息
	if senderID, ok := data["senderId"].(string); ok {
		msg.Sender = senderID
	}
	if senderNick, ok := data["senderNick"].(string); ok {
		msg.SenderNick = senderNick
	}
	// 优先使用 staffId 作为发送者 ID
	if senderStaffID, ok := data["senderStaffId"].(string); ok && senderStaffID != "" {
		msg.Sender = senderStaffID
	}

	// 提取会话信息
	if conversationID, ok := data["conversationId"].(string); ok {
		msg.Conversation = conversationID
	}
	if conversationType, ok := data["conversationType"].(string); ok {
		msg.IsInGroup = conversationType == "2"
	}

	// 提取机器人信息
	if chatbotUserID, ok := data["chatbotUserId"].(string); ok {
		msg.Receiver = chatbotUserID
	}

	// 提取 sessionWebhook
	if sessionWebhook, ok := data["sessionWebhook"].(string); ok {
		msg.SessionWebhook = sessionWebhook
	}

	// 提取 @ 用户
	if atUsers, ok := data["atUsers"].([]interface{}); ok {
		for _, u := range atUsers {
			if m, ok := u.(map[string]interface{}); ok {
				if dingtalkID, ok := m["dingtalkId"].(string); ok {
					msg.AtUserIDs = append(msg.AtUserIDs, dingtalkID)
				}
				if staffID, ok := m["staffId"].(string); ok {
					msg.AtUserIDs = append(msg.AtUserIDs, staffID)
				}
			}
		}
	}

	// 提取创建时间
	if createAt, ok := data["createAt"].(float64); ok {
		msg.Timestamp = time.UnixMilli(int64(createAt))
	}

	return msg
}

// sendAck 发送 ACK 消息
func (c *StreamClient) sendAck(ack AckMessage) {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	if c.conn == nil {
		return
	}

	data, _ := json.Marshal(ack)
	if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Errorf("[DingTalk Stream] Send ACK failed: %v", err)
	}
}

// keepalive 心跳
func (c *StreamClient) keepalive(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.connLock.Lock()
			conn := c.conn
			c.connLock.Unlock()

			if conn != nil {
				if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
					log.Errorf("[DingTalk Stream] Ping failed: %v", err)
					return
				}
			}
		}
	}
}

// Stop 停止客户端
func (c *StreamClient) Stop() {
	close(c.stopChan)

	c.connLock.Lock()
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	c.connected = false
	c.status = "offline"
	c.connLock.Unlock()
}

// IsConnected 检查连接状态
func (c *StreamClient) IsConnected() bool {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	return c.connected
}

// GetStatus 获取状态
func (c *StreamClient) GetStatus() string {
	return c.status
}

// GetAccessToken 获取 Access Token（自动缓存和刷新）
func (c *StreamClient) GetAccessToken(ctx context.Context) (string, error) {
	c.accessTokenLock.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		token := c.accessToken
		c.accessTokenLock.RUnlock()
		return token, nil
	}
	c.accessTokenLock.RUnlock()

	// 请求新的 token
	apiURL := DingTalkOpenAPIEndpoint + "/v1.0/oauth2/accessToken"
	reqBody := map[string]string{
		"appKey":    c.appKey,
		"appSecret": c.appSecret,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("get access token failed: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	c.accessTokenLock.Lock()
	c.accessToken = result.AccessToken
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	c.accessTokenLock.Unlock()

	return c.accessToken, nil
}

// GetOapiToken 获取旧版 OAPI Access Token（自动缓存和刷新）。
// 用于媒体/文件上传接口：oapi.dingtalk.com/media/upload、/file/upload/transaction/*。
// 与 GetAccessToken（新版 API Token，发消息用）是两套 token，不可混用。
// 端点：GET https://oapi.dingtalk.com/gettoken?appkey=<appKey>&appsecret=<appSecret>
func (c *StreamClient) GetOapiToken(ctx context.Context) (string, error) {
	c.oapiTokenLock.RLock()
	if c.oapiToken != "" && time.Now().Before(c.oapiTokenExpiry) {
		token := c.oapiToken
		c.oapiTokenLock.RUnlock()
		return token, nil
	}
	c.oapiTokenLock.RUnlock()

	apiURL := "https://oapi.dingtalk.com/gettoken?appkey=" + c.appKey + "&appsecret=" + c.appSecret
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var result OapiTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("get oapi token failed: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	c.oapiTokenLock.Lock()
	c.oapiToken = result.AccessToken
	c.oapiTokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	c.oapiTokenLock.Unlock()

	return c.oapiToken, nil
}

// DownloadFile 下载机器人接收到的图片/文件。
// 用 downloadCode 调用钉钉 /v1.0/robot/messageFiles/download 获取下载 URL，再 GET 拉取文件字节。
// robotCode 为机器人标识，钉钉机器人场景下使用 appKey。
func (c *StreamClient) DownloadFile(ctx context.Context, downloadCode string) ([]byte, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("get access token failed: %w", err)
	}

	// 1. 用 downloadCode 换取下载 URL
	apiURL := DingTalkOpenAPIEndpoint + "/v1.0/robot/messageFiles/download"
	reqBody := map[string]string{
		"robotCode":    c.appKey,
		"downloadCode": downloadCode,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create download request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call download api failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var dlResult struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		URL     string `json:"downloadUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&dlResult); err != nil {
		return nil, fmt.Errorf("decode download result failed: %w", err)
	}
	if dlResult.ErrCode != 0 || dlResult.URL == "" {
		return nil, fmt.Errorf("download api failed: errcode=%d, errmsg=%s", dlResult.ErrCode, dlResult.ErrMsg)
	}

	// 2. GET 下载 URL 拉取文件字节
	dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, dlResult.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create file download request failed: %w", err)
	}
	dlResp, err := c.httpClient.Do(dlReq)
	if err != nil {
		return nil, fmt.Errorf("download file failed: %w", err)
	}
	defer func() { _ = dlResp.Body.Close() }()
	if dlResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(dlResp.Body)
		return nil, fmt.Errorf("download file status %d: %s", dlResp.StatusCode, string(b))
	}
	data, err := io.ReadAll(dlResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read file data failed: %w", err)
	}
	return data, nil
}

// GetUserInfo 获取用户信息
func (c *StreamClient) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return nil, err
	}

	apiURL := fmt.Sprintf("https://api.dingtalk.com/v1.0/contact/users/%s", userID)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-acs-dingtalk-access-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		ErrCode int       `json:"errcode"`
		ErrMsg  string    `json:"errmsg"`
		User    *UserInfo `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("get user info failed: %s", result.ErrMsg)
	}

	return result.User, nil
}

// SendText 发送文本消息（通过 API，单聊）
func (c *StreamClient) SendText(ctx context.Context, receiver string, content string, atUserIDs []string) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
	msgParam, err := json.Marshal(map[string]string{
		"content": content,
	})
	if err != nil {
		return fmt.Errorf("marshal msgParam failed: %w", err)
	}
	reqBody := map[string]interface{}{
		"robotCode": c.appKey,
		"userIds":   []string{receiver},
		"msgKey":    "sampleText",
		"msgParam":  string(msgParam),
	}

	return c.sendAPI(ctx, token, apiURL, reqBody)
}

// SendGroupText 发送文本消息到群聊（通过 API）
func (c *StreamClient) SendGroupText(ctx context.Context, conversationID string, content string, atUserIDs []string) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := "https://api.dingtalk.com/v1.0/robot/groupMessages/send"
	groupMsgParam, err := json.Marshal(map[string]string{
		"content": content,
	})
	if err != nil {
		return fmt.Errorf("marshal msgParam failed: %w", err)
	}
	reqBody := map[string]interface{}{
		"robotCode":          c.appKey,
		"openConversationId": conversationID,
		"msgKey":             "sampleText",
		"msgParam":           string(groupMsgParam),
	}

	return c.sendAPI(ctx, token, apiURL, reqBody)
}

// SendMarkdown 发送 Markdown 消息（通过 API）
func (c *StreamClient) SendMarkdown(ctx context.Context, receiver string, title, content string, atUserIDs []string) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
	msgParam, err := json.Marshal(map[string]string{
		"title": title,
		"text":  content,
	})
	if err != nil {
		return fmt.Errorf("marshal msgParam failed: %w", err)
	}
	reqBody := map[string]interface{}{
		"robotCode": c.appKey,
		"userIds":   []string{receiver},
		"msgKey":    "sampleMarkdown",
		"msgParam":  string(msgParam),
	}

	return c.sendAPI(ctx, token, apiURL, reqBody)
}

// SendLink 发送链接消息
func (c *StreamClient) SendLink(ctx context.Context, receiver string, title, content, picURL, messageURL string) error {
	return fmt.Errorf("link message not implemented for stream mode")
}

// SendImage 发送图片消息
func (c *StreamClient) SendImage(ctx context.Context, receiver string, imageData []byte, atUserIDs []string) error {
	// 1. 上传图片到钉钉
	mediaID, err := c.uploadMedia(ctx, imageData, "image.png", "image")
	if err != nil {
		return fmt.Errorf("upload image failed: %w", err)
	}

	// 2. 发送图片消息
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
	msgParam, err := json.Marshal(map[string]string{
		"photoURL": mediaID,
	})
	if err != nil {
		return fmt.Errorf("marshal msgParam failed: %w", err)
	}
	reqBody := map[string]interface{}{
		"robotCode": c.appKey,
		"userIds":   []string{receiver},
		"msgKey":    "sampleImageMsg",
		"msgParam":  string(msgParam),
	}

	return c.sendAPI(ctx, token, apiURL, reqBody)
}

// dingTalkChunkUploadThreshold 普通上传与分块上传的分界（20MB，钉钉普通上传上限）。
const dingTalkChunkUploadThreshold = 20 * 1024 * 1024

// chunkSizeFor 按文件总大小动态决定分块大小（参考钉钉实现：默认 5MB，>50MB 用 6MB，>100MB 用 8MB；
// 下限 100KB，上限 8MB）。
func chunkSizeFor(fileSize int) int {
	switch {
	case fileSize > 100*1024*1024:
		return 8 * 1024 * 1024
	case fileSize > 50*1024*1024:
		return 6 * 1024 * 1024
	default:
		return 5 * 1024 * 1024
	}
}

// uploadMediaChunked 分块事务上传大文件（>20MB）到钉钉，返回 download_code（作为后续发送的 mediaId）。
// 三步：1) enable 开启事务拿 upload_id；2) 顺序逐块 chunk 上传；3) submit 提交事务拿 download_code。
// 用旧版 OAPI Token（/gettoken）作 access_token query。
func (c *StreamClient) uploadMediaChunked(ctx context.Context, mediaData []byte, fileName string) (string, error) {
	token, err := c.GetOapiToken(ctx)
	if err != nil {
		return "", fmt.Errorf("get oapi token failed: %w", err)
	}
	fileSize := len(mediaData)

	// 1. 开启上传事务
	enableURL := "https://oapi.dingtalk.com/file/upload/transaction/enable?access_token=" + token
	enableBody := &bytes.Buffer{}
	enableWriter := multipart.NewWriter(enableBody)
	if err := enableWriter.WriteField("file_name", fileName); err != nil {
		return "", fmt.Errorf("write file_name field failed: %w", err)
	}
	if err := enableWriter.WriteField("file_size", strconv.Itoa(fileSize)); err != nil {
		return "", fmt.Errorf("write file_size field failed: %w", err)
	}
	if err := enableWriter.Close(); err != nil {
		return "", fmt.Errorf("close enable writer failed: %w", err)
	}
	enableReq, err := http.NewRequestWithContext(ctx, "POST", enableURL, enableBody)
	if err != nil {
		return "", fmt.Errorf("create enable request failed: %w", err)
	}
	enableReq.Header.Set("Content-Type", enableWriter.FormDataContentType())

	enableResp, err := c.httpClient.Do(enableReq)
	if err != nil {
		return "", fmt.Errorf("enable upload transaction failed: %w", err)
	}
	enableRespBody, _ := io.ReadAll(enableResp.Body)
	_ = enableResp.Body.Close()
	var enableResult struct {
		ErrCode  int    `json:"errcode"`
		ErrMsg   string `json:"errmsg"`
		UploadID string `json:"upload_id"`
	}
	if err := json.Unmarshal(enableRespBody, &enableResult); err != nil {
		return "", fmt.Errorf("parse enable response failed: %w, body=%s", err, truncate(string(enableRespBody), 200))
	}
	if enableResult.ErrCode != 0 || enableResult.UploadID == "" {
		return "", fmt.Errorf("enable upload transaction failed: errcode=%d, errmsg=%s", enableResult.ErrCode, enableResult.ErrMsg)
	}
	uploadID := enableResult.UploadID
	log.Infof("[DingTalk Stream] Chunk upload enabled: upload_id=%s, fileName=%s, size=%d", uploadID, fileName, fileSize)

	// 2. 顺序逐块上传
	chunkSize := chunkSizeFor(fileSize)
	totalChunks := (fileSize + chunkSize - 1) / chunkSize
	chunkURL := "https://oapi.dingtalk.com/file/upload/chunk?access_token=" + token
	for i := 0; i < totalChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > fileSize {
			end = fileSize
		}
		chunk := mediaData[start:end]

		chunkBody := &bytes.Buffer{}
		chunkWriter := multipart.NewWriter(chunkBody)
		if err := chunkWriter.WriteField("upload_id", uploadID); err != nil {
			return "", fmt.Errorf("write upload_id field failed: %w", err)
		}
		if err := chunkWriter.WriteField("chunk_number", strconv.Itoa(i+1)); err != nil {
			return "", fmt.Errorf("write chunk_number field failed: %w", err)
		}
		if err := chunkWriter.WriteField("total_chunks", strconv.Itoa(totalChunks)); err != nil {
			return "", fmt.Errorf("write total_chunks field failed: %w", err)
		}
		part, err := chunkWriter.CreateFormFile("file", fileName)
		if err != nil {
			return "", fmt.Errorf("create chunk form file failed: %w", err)
		}
		if _, err := part.Write(chunk); err != nil {
			return "", fmt.Errorf("write chunk data failed: %w", err)
		}
		if err := chunkWriter.Close(); err != nil {
			return "", fmt.Errorf("close chunk writer failed: %w", err)
		}

		chunkReq, err := http.NewRequestWithContext(ctx, "POST", chunkURL, chunkBody)
		if err != nil {
			return "", fmt.Errorf("create chunk request failed: %w", err)
		}
		chunkReq.Header.Set("Content-Type", chunkWriter.FormDataContentType())

		chunkResp, err := c.httpClient.Do(chunkReq)
		if err != nil {
			return "", fmt.Errorf("upload chunk %d failed: %w", i+1, err)
		}
		chunkRespBody, _ := io.ReadAll(chunkResp.Body)
		_ = chunkResp.Body.Close()
		var chunkResult struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		if err := json.Unmarshal(chunkRespBody, &chunkResult); err != nil {
			return "", fmt.Errorf("parse chunk %d response failed: %w, body=%s", i+1, err, truncate(string(chunkRespBody), 200))
		}
		if chunkResult.ErrCode != 0 {
			return "", fmt.Errorf("upload chunk %d failed: errcode=%d, errmsg=%s", i+1, chunkResult.ErrCode, chunkResult.ErrMsg)
		}
		log.Debugf("[DingTalk Stream] Chunk uploaded: %d/%d, size=%d", i+1, totalChunks, len(chunk))
	}

	// 3. 提交事务，拿 download_code（即后续发送的 mediaId）
	submitURL := "https://oapi.dingtalk.com/file/upload/transaction/submit?access_token=" + token +
		"&upload_id=" + uploadID + "&file_name=" + url.QueryEscape(fileName)
	submitReq, err := http.NewRequestWithContext(ctx, "GET", submitURL, nil)
	if err != nil {
		return "", fmt.Errorf("create submit request failed: %w", err)
	}
	submitResp, err := c.httpClient.Do(submitReq)
	if err != nil {
		return "", fmt.Errorf("submit upload transaction failed: %w", err)
	}
	submitRespBody, _ := io.ReadAll(submitResp.Body)
	_ = submitResp.Body.Close()
	var submitResult struct {
		ErrCode      int    `json:"errcode"`
		ErrMsg       string `json:"errmsg"`
		DownloadCode string `json:"download_code"`
	}
	if err := json.Unmarshal(submitRespBody, &submitResult); err != nil {
		return "", fmt.Errorf("parse submit response failed: %w, body=%s", err, truncate(string(submitRespBody), 200))
	}
	if submitResult.ErrCode != 0 || submitResult.DownloadCode == "" {
		return "", fmt.Errorf("submit upload transaction failed: errcode=%d, errmsg=%s", submitResult.ErrCode, submitResult.ErrMsg)
	}
	log.Infof("[DingTalk Stream] Chunk upload submitted: download_code=%s, fileName=%s", submitResult.DownloadCode, fileName)
	return submitResult.DownloadCode, nil
}

// uploadFile 统一文件上传入口：≤20MB 走普通上传（返回 media_id），>20MB 走分块事务上传（返回 download_code）。
// 两者统一作为「mediaId 字符串」返回，由调用方在发送前补 `@` 前缀。
func (c *StreamClient) uploadFile(ctx context.Context, mediaData []byte, fileName string) (string, error) {
	if len(mediaData) > dingTalkChunkUploadThreshold {
		log.Infof("[DingTalk Stream] File %s (%d bytes) > 20MB, using chunked upload", fileName, len(mediaData))
		return c.uploadMediaChunked(ctx, mediaData, fileName)
	}
	return c.uploadMedia(ctx, mediaData, fileName, "file")
}

// SendFile 发送文件附件（实现 types.FileSender 的底层能力，由 DingTalkAdapter.SendFile 调用）。
// 流程：uploadFile 拿 mediaId（普通/分块自动选择）→ 补 `@` 前缀 → 发 sampleFile 消息。
// 单聊走 oToMessages/batchSend，群聊走 groupMessages/send。msgKey 固定 sampleFile，
// msgParam 为字符串化的 JSON（mediaId 必须带 `@`）。
func (c *StreamClient) SendFile(ctx context.Context, receiver, fileName, fileType string, mediaData []byte, isInGroup bool, conversationID string) error {
	mediaID, err := c.uploadFile(ctx, mediaData, fileName)
	if err != nil {
		return fmt.Errorf("upload file failed: %w", err)
	}
	if !strings.HasPrefix(mediaID, "@") {
		mediaID = "@" + mediaID
	}

	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	msgParam, err := json.Marshal(map[string]string{
		"mediaId":  mediaID,
		"fileName": fileName,
		"fileType": fileType,
	})
	if err != nil {
		return fmt.Errorf("marshal msgParam failed: %w", err)
	}

	var apiURL string
	reqBody := map[string]interface{}{
		"robotCode": c.appKey,
		"msgKey":    "sampleFile",
		"msgParam":  string(msgParam), // 必须是字符串化的 JSON，非对象
	}
	if isInGroup {
		apiURL = "https://api.dingtalk.com/v1.0/robot/groupMessages/send"
		reqBody["openConversationId"] = conversationID
		log.Infof("[DingTalk] SendFile via groupMessages API, conversationID=%s, fileName=%s", conversationID, fileName)
	} else {
		apiURL = "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
		reqBody["userIds"] = []string{receiver}
		log.Infof("[DingTalk] SendFile via oToMessages API, receiver=%s, fileName=%s", receiver, fileName)
	}

	return c.sendAPI(ctx, token, apiURL, reqBody)
}

// ReplyWithWebhook 使用 sessionWebhook 回复文本消息
func (c *StreamClient) ReplyWithWebhook(ctx context.Context, webhookURL, content string, atUserIDs []string) error {
	if webhookURL == "" {
		return fmt.Errorf("sessionWebhook is empty")
	}

	reqBody := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
	if len(atUserIDs) > 0 {
		reqBody["at"] = map[string]interface{}{
			"atUserIds": atUserIDs,
		}
	}

	return c.sendHTTPRequest(ctx, "POST", webhookURL, reqBody, "")
}

// ReplyMarkdownWithWebhook 使用 sessionWebhook 回复 Markdown 消息
func (c *StreamClient) ReplyMarkdownWithWebhook(ctx context.Context, webhookURL, title, content string, atUserIDs []string) error {
	if webhookURL == "" {
		return fmt.Errorf("sessionWebhook is empty")
	}

	reqBody := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"title": title,
			"text":  content,
		},
	}
	if len(atUserIDs) > 0 {
		reqBody["at"] = map[string]interface{}{
			"atUserIds": atUserIDs,
		}
	}

	return c.sendHTTPRequest(ctx, "POST", webhookURL, reqBody, "")
}

// uploadMedia 上传媒体文件到钉钉（普通上传，≤20MB）。
// 用旧版 OAPI Token（/gettoken）作 access_token query；返回的 media_id 可能带也可能不带 `@`，
// 由调用方在发送前统一补 `@` 前缀。
func (c *StreamClient) uploadMedia(ctx context.Context, mediaData []byte, filename, mediaType string) (string, error) {
	token, err := c.GetOapiToken(ctx)
	if err != nil {
		return "", fmt.Errorf("get oapi token failed: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("type", mediaType); err != nil {
		return "", fmt.Errorf("write type field failed: %w", err)
	}

	part, err := writer.CreateFormFile("media", filename)
	if err != nil {
		return "", fmt.Errorf("create form file failed: %w", err)
	}
	if _, err := part.Write(mediaData); err != nil {
		return "", fmt.Errorf("write media data failed: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer failed: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://oapi.dingtalk.com/media/upload?access_token="+token, body)
	if err != nil {
		return "", fmt.Errorf("create upload request failed: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload media failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read upload response failed: %w", err)
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		MediaID string `json:"media_id"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse upload response failed: %w", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("upload media failed: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	log.Infof("[DingTalk Stream] Uploaded media: mediaId=%s, type=%s, filename=%s", result.MediaID, mediaType, filename)
	return result.MediaID, nil
}

// --- 卡片 API 方法 ---

// CreateAndDeliverCard 创建并投放 AI 卡片
// 参考：https://open.dingtalk.com/document/orgapp/create-and-deliver-cards
func (c *StreamClient) CreateAndDeliverCard(
	ctx context.Context,
	cardTemplateID string,
	cardData map[string]string,
	senderID string,
	conversationID string,
	isInGroup bool,
) (string, error) {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("get access token failed: %w", err)
	}

	// 生成唯一的 cardInstanceID
	cardInstanceID := generateCardInstanceID(senderID, conversationID)

	// 构建创建卡片请求
	createReq := CreateAndDeliverCardRequest{
		CardTemplateID: cardTemplateID,
		OutTrackID:     cardInstanceID,
		CardData: CardData{
			CardParamMap: cardData,
		},
		CallbackType:          "STREAM",
		ImGroupOpenSpaceModel: &CardSpaceModel{SupportForward: true},
		ImRobotOpenSpaceModel: &CardSpaceModel{SupportForward: true},
		UserIDType:            1,
	}

	// 根据群聊/单聊设置投放目标
	if isInGroup {
		createReq.OpenSpaceID = fmt.Sprintf("dtv1.card//IM_GROUP.%s", conversationID)
		createReq.ImGroupOpenDeliverModel = &DeliverModel{
			RobotCode: c.appKey,
		}
	} else {
		createReq.OpenSpaceID = fmt.Sprintf("dtv1.card//IM_ROBOT.%s", senderID)
		createReq.ImRobotOpenDeliverModel = &DeliverModel{
			SpaceType: "IM_ROBOT",
		}
	}

	apiURL := DingTalkOpenAPIEndpoint + "/v1.0/card/instances/createAndDeliver"
	body, _ := json.Marshal(createReq)

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create card request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("create card request error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("create card failed: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	log.Infof("[DingTalk Stream] Created and delivered card: cardInstanceID=%s", cardInstanceID)
	return cardInstanceID, nil
}

// StreamingCard 流式更新卡片内容
// 参考：https://open.dingtalk.com/document/orgapp/interactive-card-streaming-output
func (c *StreamClient) StreamingCard(
	ctx context.Context,
	cardInstanceID string,
	contentKey string,
	content string,
	isFull bool,
	isFinalize bool,
	isError bool,
) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token failed: %w", err)
	}

	streamingReq := StreamingCardRequest{
		OutTrackID: cardInstanceID,
		GUID:       uuid.New().String(),
		Key:        contentKey,
		Content:    content,
		IsFull:     isFull,
		IsFinalize: isFinalize,
		IsError:    isError,
	}

	apiURL := DingTalkOpenAPIEndpoint + "/v1.0/card/streaming"
	body, _ := json.Marshal(streamingReq)

	req, err := http.NewRequestWithContext(ctx, "PUT", apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("streaming card request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("streaming card request error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("streaming card failed: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	return nil
}

// FinishCard 完成卡片（更新状态为 FINISHED）
func (c *StreamClient) FinishCard(ctx context.Context, cardInstanceID string, cardData map[string]string) error {
	cardData["flowStatus"] = string(AICardStatusFinished)
	return c.updateCard(ctx, cardInstanceID, cardData)
}

// FailCard 标记卡片失败
func (c *StreamClient) FailCard(ctx context.Context, cardInstanceID string, cardData map[string]string) error {
	cardData["flowStatus"] = string(AICardStatusFailed)
	return c.updateCard(ctx, cardInstanceID, cardData)
}

// SetCardInputing 将卡片状态置为 INPUTING（输入中）。
// 钉钉官方 AI 卡片流式流程：createAndDeliver(PROCESSING) → put_card_data(INPUTING) → streaming(...)。
// 首次 streaming 前需先置 INPUTING，否则可能返回未知错误。
func (c *StreamClient) SetCardInputing(ctx context.Context, cardInstanceID string) error {
	cardData := map[string]string{"flowStatus": string(AICardStatusInputing)}
	return c.updateCard(ctx, cardInstanceID, cardData)
}

// updateCard 更新卡片内容（内部方法）
func (c *StreamClient) updateCard(ctx context.Context, cardInstanceID string, cardData map[string]string) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token failed: %w", err)
	}

	updateReq := UpdateCardRequest{
		OutTrackID: cardInstanceID,
		CardData: CardData{
			CardParamMap: cardData,
		},
	}

	apiURL := DingTalkOpenAPIEndpoint + "/v1.0/card/instances"
	body, _ := json.Marshal(updateReq)

	req, err := http.NewRequestWithContext(ctx, "PUT", apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("update card request failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("update card request error: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update card failed: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	return nil
}

// sendAPI 通过钉钉 API 发送消息
func (c *StreamClient) sendAPI(ctx context.Context, token, apiURL string, reqBody interface{}) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	log.Debugf("[DingTalk Stream] sendAPI: url=%s, body=%s", apiURL, truncate(string(body), 500))

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-acs-dingtalk-access-token", token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dingtalk API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	log.Debugf("[DingTalk Stream] sendAPI response: status=%d, body=%s", resp.StatusCode, truncate(string(respBody), 500))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk API returned status %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	var result SendMessageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Warnf("[DingTalk Stream] sendAPI: failed to parse response as SendMessageResponse: %v, raw=%s", err, truncate(string(respBody), 200))
		return nil // HTTP 200 但 JSON 格式不是预期的，可能是某些 API 不返回 errcode
	}

	if result.ErrCode != 0 {
		return classifyDingTalkSendErr(result.ErrCode, result.ErrMsg, "send message failed")
	}

	return nil
}

// sendHTTPRequest 发送通用 HTTP 请求
func (c *StreamClient) sendHTTPRequest(ctx context.Context, method, url string, reqBody interface{}, token string) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	log.Debugf("[DingTalk Stream] sendHTTPRequest: method=%s, url=%s, body=%s", method, url, truncate(string(body), 500))

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("x-acs-dingtalk-access-token", token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dingtalk HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	log.Debugf("[DingTalk Stream] sendHTTPRequest response: status=%d, body=%s", resp.StatusCode, truncate(string(respBody), 500))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk HTTP request returned status %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	var result SendMessageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Warnf("[DingTalk Stream] sendHTTPRequest: failed to parse response as SendMessageResponse: %v, raw=%s", err, truncate(string(respBody), 200))
		return nil // HTTP 200 但 JSON 格式不是预期的，sessionWebhook 回复可能不返回 errcode
	}

	if result.ErrCode != 0 {
		return classifyDingTalkSendErr(result.ErrCode, result.ErrMsg, "send message via webhook failed")
	}

	return nil
}

// --- Stream 消息类型定义 ---

// ConnectionResponse 连接响应
type ConnectionResponse struct {
	Endpoint string `json:"endpoint"`
	Ticket   string `json:"ticket"`
}

// StreamMessage Stream 消息基础结构
type StreamMessage struct {
	Type string `json:"type"`
}

// SystemMessage 系统消息
type SystemMessage struct {
	Type    string                 `json:"type"`
	Headers SystemMessageHeaders   `json:"headers"`
	Data    map[string]interface{} `json:"data"`
}

// SystemMessageHeaders 系统消息头
type SystemMessageHeaders struct {
	Topic     string `json:"topic"`
	MessageID string `json:"messageId"`
	Timestamp int64  `json:"timestamp"`
}

// CallbackHeaders 回调消息头
type CallbackHeaders struct {
	Topic       string `json:"topic"`
	MessageID   string `json:"messageId"`
	CreateTime  int64  `json:"createTime"`
	EventType   string `json:"eventType"`
	ContentType string `json:"contentType"`
	BornTime    int64  `json:"bornTime"`
	Producer    string `json:"producer"`
	GrantType   string `json:"grantType"`
	UnionId     string `json:"unionId"`
}

// AckMessage ACK 消息
type AckMessage struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Headers AckHeaders             `json:"headers"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// AckHeaders ACK 消息头
type AckHeaders struct {
	MessageID   string `json:"messageId"`
	ContentType string `json:"contentType"`
}

// cleanupSeenMsgIDs 清理过期的消息去重记录（5 分钟前的）
// 使用 sync.Map 的 Range 方法遍历删除，避免内存泄漏
func (c *StreamClient) cleanupSeenMsgIDs() {
	expiry := time.Now().Add(-5 * time.Minute)
	c.seenMsgIDs.Range(func(key, value interface{}) bool {
		if t, ok := value.(time.Time); ok && t.Before(expiry) {
			c.seenMsgIDs.Delete(key)
		}
		return true
	})
	c.seenMsgDigest.Range(func(key, value interface{}) bool {
		if t, ok := value.(time.Time); ok && t.Before(expiry) {
			c.seenMsgDigest.Delete(key)
		}
		return true
	})

}

// truncate 截断字符串
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// generateCardInstanceID 生成唯一的卡片实例 ID
// 参考 dingtalk_stream SDK 的 CardReplier.gen_card_id
func generateCardInstanceID(senderID, conversationID string) string {
	factor := fmt.Sprintf("%s_%s_%s", senderID, conversationID, uuid.New().String())
	h := sha256.Sum256([]byte(factor))
	return fmt.Sprintf("%x", h)
}
