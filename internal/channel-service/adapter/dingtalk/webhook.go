package dingtalk

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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
)

// WebhookClient 钉钉 Webhook 模式客户端
// 用于接收钉钉 HTTP 回调推送，通过 sessionWebhook 或 API 回复消息
type WebhookClient struct {
	appKey    string
	appSecret string
	channelID string // 渠道 ID

	accessToken     string
	tokenExpiry     time.Time
	accessTokenLock sync.RWMutex

	// 旧版 OAPI Token（GET /gettoken），用于媒体/文件上传接口。与新版 API Token（accessToken）是两套。
	oapiToken       string
	oapiTokenExpiry time.Time
	oapiTokenLock   sync.RWMutex

	// sessionWebhook 缓存
	sessionWebhooks map[string]*sessionWebhookInfo
	webhookMu       sync.RWMutex

	// 消息处理器
	messageHandler MessageHandler

	// HTTP 客户端
	httpClient *http.Client

	// 状态
	status string
}

type sessionWebhookInfo struct {
	webhookURL  string
	expiredTime time.Time
}

// WebhookConfig Webhook 客户端配置
type WebhookConfig struct {
	AppKey    string
	AppSecret string
	ChannelID string
}

// NewWebhookClient 创建 Webhook 客户端
func NewWebhookClient(config *WebhookConfig) *WebhookClient {
	return &WebhookClient{
		appKey:          config.AppKey,
		appSecret:       config.AppSecret,
		channelID:       config.ChannelID,
		sessionWebhooks: make(map[string]*sessionWebhookInfo),
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		status:          "offline",
	}
}

// OnMessage 注册消息处理器
func (c *WebhookClient) OnMessage(handler MessageHandler) {
	c.messageHandler = handler
}

// Start 启动（Webhook 模式不需要建立连接，标记为在线即可）
func (c *WebhookClient) Start(_ context.Context) error {
	c.status = "online"
	log.Infof("[DingTalk Webhook] Client started for channel %s", c.channelID)
	return nil
}

// Stop 停止
func (c *WebhookClient) Stop() {
	c.status = "offline"
	log.Infof("[DingTalk Webhook] Client stopped for channel %s", c.channelID)
}

// GetStatus 获取状态
func (c *WebhookClient) GetStatus() string {
	return c.status
}

// HandleWebhook 处理钉钉 Webhook 推送消息
// 此方法应在 HTTP callback 路由中调用
func (c *WebhookClient) HandleWebhook(ctx context.Context, body []byte, timestamp, sign string) error {
	// 1. 验证签名
	if err := c.verifySign(timestamp, sign); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	// 2. 解析消息
	var req WebhookRequest
	if err := json.Unmarshal(body, &req); err != nil {
		return fmt.Errorf("failed to parse webhook message: %w", err)
	}

	// 3. 转换为统一消息格式
	msg := c.convertToMessage(&req)
	msg.ChannelID = c.channelID
	msg.Receiver = req.ChatbotUserID

	// 4. 缓存 sessionWebhook 用于回复
	if req.SessionWebhook != "" && req.Sender != "" {
		c.webhookMu.Lock()
		c.sessionWebhooks[req.Sender] = &sessionWebhookInfo{
			webhookURL:  req.SessionWebhook,
			expiredTime: time.Now().Add(time.Duration(req.SessionWebhookExpiredTime) * time.Second),
		}
		c.webhookMu.Unlock()
	}

	// 5. 更新状态
	c.status = "online"

	// 6. 调用消息处理器
	if c.messageHandler != nil {
		go func() {
			// 若为图片/文件消息，先下载附件字节填入 Attachments（供 chat handler 上传给 WGA）
			if url, ok := msg.Raw["attachmentURL"].(string); ok && url != "" {
				name, _ := msg.Raw["attachmentName"].(string)
				if name == "" {
					name = "dingtalk-file"
				}
				data, dErr := downloadFromURL(ctx, url)
				if dErr != nil {
					log.Errorf("[DingTalk Webhook] Failed to download attachment: %v", dErr)
				} else {
					msg.Attachments = append(msg.Attachments, Attachment{
						Name:     name,
						MimeType: http.DetectContentType(data),
						Data:     data,
					})
					log.Infof("[DingTalk Webhook] Downloaded attachment, size=%d", len(data))
				}
			}
			if err := c.messageHandler(ctx, msg); err != nil {
				log.Errorf("[DingTalk Webhook] Message handler error: %v", err)
			}
		}()
	}

	return nil
}

// convertToMessage 转换 Webhook 请求为统一消息格式
func (c *WebhookClient) convertToMessage(req *WebhookRequest) *Message {
	msg := &Message{
		MsgType:        MessageType(req.Msgtype),
		Conversation:   req.ConversationID,
		Sender:         req.Sender,
		SenderNick:     req.SenderNick,
		Timestamp:      time.UnixMilli(req.CreateTime),
		IsInGroup:      req.ConversationType == "2",
		SessionWebhook: req.SessionWebhook,
		Raw:            make(map[string]interface{}),
	}

	// 解析 @ 用户
	for _, atUser := range req.AtUsers {
		msg.AtUserIDs = append(msg.AtUserIDs, atUser.DingtalkID)
	}

	// 根据消息类型设置内容
	switch req.Msgtype {
	case "text":
		msg.Content = req.Text.Content
	case "picture":
		// 图片：Content 留空，将下载 URL 存入 Raw，由 HandleWebhook 下载字节填入 Attachments
		msg.Content = ""
		msg.MsgType = MessageTypeImage
		msg.Raw["attachmentURL"] = req.Picture.PicURL
		msg.Raw["attachmentName"] = "dingtalk-image"
	case "voice":
		msg.Content = req.Voice.Content
		msg.MsgType = MessageTypeVoice
	case "file":
		// 文件：Content 留空，将下载 URL 存入 Raw，由 HandleWebhook 下载字节填入 Attachments
		msg.Content = ""
		msg.MsgType = MessageTypeFile
		msg.Raw["attachmentURL"] = req.File.FileURL
		msg.Raw["attachmentName"] = req.File.FileName
	case "richText":
		msg.Content = req.RichText.Content
		msg.MsgType = MessageTypeMarkdown
	}

	return msg
}

// verifySign 验证钉钉 Webhook 签名
func (c *WebhookClient) verifySign(timestamp, sign string) error {
	if c.appSecret == "" {
		return nil // 没有配置 secret 则跳过验证
	}

	// 构造签名字符串: timestamp + "\n" + appSecret
	stringToSign := timestamp + "\n" + c.appSecret

	// 计算 HMAC-SHA256
	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(stringToSign))
	expectedSign := base64.StdEncoding.EncodeToString(h.Sum(nil))

	if sign != expectedSign {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// GetAccessToken 获取 access token（自动缓存和刷新）
func (c *WebhookClient) GetAccessToken(ctx context.Context) (string, error) {
	c.accessTokenLock.RLock()
	if c.accessToken != "" && time.Now().Before(c.tokenExpiry) {
		token := c.accessToken
		c.accessTokenLock.RUnlock()
		return token, nil
	}
	c.accessTokenLock.RUnlock()

	apiURL := "https://api.dingtalk.com/v1.0/oauth2/accessToken"
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
	c.tokenExpiry = time.Now().Add(time.Duration(result.ExpiresIn-60) * time.Second)
	c.accessTokenLock.Unlock()

	return c.accessToken, nil
}

// GetOapiToken 获取旧版 OAPI Access Token（自动缓存和刷新），用于媒体/文件上传接口。
// 端点：GET https://oapi.dingtalk.com/gettoken?appkey=<appKey>&appsecret=<appSecret>
func (c *WebhookClient) GetOapiToken(ctx context.Context) (string, error) {
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
func (c *WebhookClient) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
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

// SendText 发送文本消息
// 优先使用 sessionWebhook 回复，否则使用 API 发送
func (c *WebhookClient) SendText(ctx context.Context, receiver string, content string, atUserIDs []string) error {
	// 优先使用 sessionWebhook
	if webhookURL := c.getSessionWebhook(receiver); webhookURL != "" {
		return c.sendViaWebhook(ctx, webhookURL, &TextMessage{
			MsgType: "text",
			Text: struct {
				Content string `json:"content"`
			}{Content: content},
		})
	}

	// 使用 API 发送
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
	reqBody := map[string]interface{}{
		"robotCode": c.appKey,
		"userIds":   []string{receiver},
		"msgKey":    "sampleText",
		"msgParam": map[string]string{
			"content": content,
		},
	}

	return c.sendAPI(ctx, token, apiURL, reqBody)
}

// SendGroupText 发送文本消息到群聊（通过 API）
func (c *WebhookClient) SendGroupText(ctx context.Context, conversationID string, content string, atUserIDs []string) error {
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := "https://api.dingtalk.com/v1.0/robot/groupMessages/send"
	reqBody := map[string]interface{}{
		"robotCode":          c.appKey,
		"openConversationId": conversationID,
		"msgKey":             "sampleText",
		"msgParam": map[string]string{
			"content": content,
		},
	}

	return c.sendAPI(ctx, token, apiURL, reqBody)
}

// SendMarkdown 发送 Markdown 消息
func (c *WebhookClient) SendMarkdown(ctx context.Context, receiver string, title, content string, atUserIDs []string) error {
	// 优先使用 sessionWebhook
	if webhookURL := c.getSessionWebhook(receiver); webhookURL != "" {
		return c.sendViaWebhook(ctx, webhookURL, &MarkdownMessage{
			MsgType: "markdown",
			Markdown: struct {
				Title string `json:"title"`
				Text  string `json:"text"`
			}{Title: title, Text: content},
		})
	}

	// 使用 API 发送
	token, err := c.GetAccessToken(ctx)
	if err != nil {
		return err
	}

	apiURL := "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
	reqBody := map[string]interface{}{
		"robotCode": c.appKey,
		"userIds":   []string{receiver},
		"msgKey":    "sampleMarkdown",
		"msgParam": map[string]string{
			"title": title,
			"text":  content,
		},
	}

	return c.sendAPI(ctx, token, apiURL, reqBody)
}

// SendLink 发送链接消息
func (c *WebhookClient) SendLink(ctx context.Context, receiver string, title, content, picURL, messageURL string) error {
	// 优先使用 sessionWebhook
	if webhookURL := c.getSessionWebhook(receiver); webhookURL != "" {
		return c.sendViaWebhook(ctx, webhookURL, &LinkMessage{
			MsgType: "link",
			Link: struct {
				Title      string `json:"title"`
				Text       string `json:"text"`
				MessageURL string `json:"messageUrl"`
				PicURL     string `json:"picUrl,omitempty"`
			}{Title: title, Text: content, PicURL: picURL, MessageURL: messageURL},
		})
	}

	return fmt.Errorf("link message via API not implemented")
}

// SendImage Webhook 模式不支持图片上传，降级为 Markdown
func (c *WebhookClient) SendImage(ctx context.Context, receiver string, imageData []byte, atUserIDs []string) error {
	return fmt.Errorf("image message not supported in webhook mode, use StreamClient instead")
}

// uploadMedia 上传媒体文件到钉钉（普通上传，≤20MB），用旧版 OAPI Token。返回 media_id（可能带或不带 `@`）。
func (c *WebhookClient) uploadMedia(ctx context.Context, mediaData []byte, filename, mediaType string) (string, error) {
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

	log.Infof("[DingTalk Webhook] Uploaded media: mediaId=%s, type=%s, filename=%s", result.MediaID, mediaType, filename)
	return result.MediaID, nil
}

// uploadMediaChunked 分块事务上传大文件（>20MB），返回 download_code（作为 mediaId）。三步：enable/chunk/submit。
func (c *WebhookClient) uploadMediaChunked(ctx context.Context, mediaData []byte, fileName string) (string, error) {
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
	log.Infof("[DingTalk Webhook] Chunk upload enabled: upload_id=%s, fileName=%s, size=%d", uploadID, fileName, fileSize)

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
		log.Debugf("[DingTalk Webhook] Chunk uploaded: %d/%d, size=%d", i+1, totalChunks, len(chunk))
	}

	// 3. 提交事务，拿 download_code
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
	log.Infof("[DingTalk Webhook] Chunk upload submitted: download_code=%s, fileName=%s", submitResult.DownloadCode, fileName)
	return submitResult.DownloadCode, nil
}

// uploadFile 统一文件上传入口：≤20MB 走普通上传，>20MB 走分块事务上传。
func (c *WebhookClient) uploadFile(ctx context.Context, mediaData []byte, fileName string) (string, error) {
	if len(mediaData) > dingTalkChunkUploadThreshold {
		log.Infof("[DingTalk Webhook] File %s (%d bytes) > 20MB, using chunked upload", fileName, len(mediaData))
		return c.uploadMediaChunked(ctx, mediaData, fileName)
	}
	return c.uploadMedia(ctx, mediaData, fileName, "file")
}

// SendFile 发送文件附件（由 DingTalkAdapter.SendFile 调用）。
// 单聊走 oToMessages/batchSend，群聊走 groupMessages/send，msgKey 固定 sampleFile，mediaId 带 `@`。
func (c *WebhookClient) SendFile(ctx context.Context, receiver, fileName, fileType string, mediaData []byte, isInGroup bool, conversationID string) error {
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
		"msgParam":  string(msgParam),
	}
	if isInGroup {
		apiURL = "https://api.dingtalk.com/v1.0/robot/groupMessages/send"
		reqBody["openConversationId"] = conversationID
		log.Infof("[DingTalk] SendFile via webhook groupMessages API, conversationID=%s, fileName=%s", conversationID, fileName)
	} else {
		apiURL = "https://api.dingtalk.com/v1.0/robot/oToMessages/batchSend"
		reqBody["userIds"] = []string{receiver}
		log.Infof("[DingTalk] SendFile via webhook oToMessages API, receiver=%s, fileName=%s", receiver, fileName)
	}

	return c.sendAPI(ctx, token, apiURL, reqBody)
}

// getSessionWebhook 获取缓存的 sessionWebhook URL
func (c *WebhookClient) getSessionWebhook(userID string) string {
	c.webhookMu.RLock()
	defer c.webhookMu.RUnlock()

	info, ok := c.sessionWebhooks[userID]
	if !ok || time.Now().After(info.expiredTime) {
		return ""
	}
	return info.webhookURL
}

// sendViaWebhook 通过 sessionWebhook 发送消息
func (c *WebhookClient) sendViaWebhook(ctx context.Context, webhookURL string, msg interface{}) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	log.Debugf("[DingTalk Webhook] sendViaWebhook: url=%s, body=%s", webhookURL, truncate(string(body), 500))

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("dingtalk webhook request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	log.Debugf("[DingTalk Webhook] sendViaWebhook response: status=%d, body=%s", resp.StatusCode, truncate(string(respBody), 500))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk webhook returned status %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	var result SendMessageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Warnf("[DingTalk Webhook] sendViaWebhook: failed to parse response as SendMessageResponse: %v, raw=%s", err, truncate(string(respBody), 200))
		return nil // HTTP 200 但 JSON 格式不是预期的，sessionWebhook 回复可能不返回 errcode
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("send message via webhook failed: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	return nil
}

// sendAPI 通过钉钉 API 发送消息
func (c *WebhookClient) sendAPI(ctx context.Context, token, apiURL string, reqBody interface{}) error {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	log.Debugf("[DingTalk Webhook] sendAPI: url=%s, body=%s", apiURL, truncate(string(body), 500))

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

	log.Debugf("[DingTalk Webhook] sendAPI response: status=%d, body=%s", resp.StatusCode, truncate(string(respBody), 500))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("dingtalk API returned status %d: %s", resp.StatusCode, truncate(string(respBody), 200))
	}

	var result SendMessageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Warnf("[DingTalk Webhook] sendAPI: failed to parse response as SendMessageResponse: %v, raw=%s", err, truncate(string(respBody), 200))
		return nil // HTTP 200 但 JSON 格式不是预期的
	}

	if result.ErrCode != 0 {
		return fmt.Errorf("send message via API failed: errcode=%d, errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	return nil
}

// SignURL 生成带签名的 Webhook URL（用于自定义机器人）
func SignURL(webhookURL, secret string) (string, error) {
	if secret == "" {
		return webhookURL, nil
	}

	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	stringToSign := timestamp + "\n" + secret

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	sign := url.QueryEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))

	if strings.Contains(webhookURL, "?") {
		return fmt.Sprintf("%s&timestamp=%s&sign=%s", webhookURL, timestamp, sign), nil
	}
	return fmt.Sprintf("%s?timestamp=%s&sign=%s", webhookURL, timestamp, sign), nil
}

// downloadFromURL 通过 HTTP GET 下载指定 URL 的文件字节（用于 webhook 图片/文件附件）。
func downloadFromURL(ctx context.Context, fileURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download status %d: %s", resp.StatusCode, string(b))
	}
	return io.ReadAll(resp.Body)
}
