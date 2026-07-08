package wechat

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/types"
	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	// DefaultBaseURL 微信 openclaw 默认地址
	DefaultBaseURL         = "https://ilinkai.weixin.qq.com"
	SessionExpiredErrCode  = -14
	MaxTextChunkSize       = 2000
	DefaultLongPollTimeout = 40
)

// WeChatAdapter 微信平台适配器
// 使用 openclaw 协议通过 Long Poll 接收消息
type WeChatAdapter struct {
	mu        sync.RWMutex
	config    types.AdapterConfig
	connected bool
	handler   types.MessageHandler

	// openclaw 客户端状态
	baseURL   string
	token     string
	accountID string
	status    string

	// 轮询控制
	polling  bool
	pollBuf  string
	stopChan chan struct{}

	// 消息去重：记录最近处理过的消息指纹
	seenMsgs map[string]struct{}

	// HTTP 客户端
	httpClient *http.Client
}

// NewWeChatAdapter 创建微信适配器
func NewWeChatAdapter() *WeChatAdapter {
	return &WeChatAdapter{
		status:     "offline",
		httpClient: &http.Client{Timeout: 60 * time.Second},
		stopChan:   make(chan struct{}),
		seenMsgs:   make(map[string]struct{}),
	}
}

// Connect 连接到微信平台
func (w *WeChatAdapter) Connect(config types.AdapterConfig) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.config = config
	w.baseURL = config.BaseUrl
	if w.baseURL == "" {
		w.baseURL = DefaultBaseURL
	}
	w.token = config.Token
	w.accountID = config.AccountId

	// 如果有 token，验证会话并开始轮询
	if w.token != "" {
		w.status = "logged_in"
		w.connected = true
		w.polling = true
		go w.pollLoop(context.Background())
		log.Infof("[WeChat] Adapter connected for channel %s (with token, polling started)", config.ChannelID)
		return nil
	}

	// 没有 token，设置为等待登录状态
	w.status = "waiting_login"
	w.connected = true
	log.Infof("[WeChat] Adapter connected for channel %s (waiting for login)", config.ChannelID)
	return nil
}

// Disconnect 断开微信连接
func (w *WeChatAdapter) Disconnect() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.polling = false
	w.status = "offline"
	w.connected = false

	select {
	case <-w.stopChan:
		// already closed
	default:
		close(w.stopChan)
	}

	log.Infof("[WeChat] Adapter disconnected for channel %s", w.config.ChannelID)
	return nil
}

// IsConnected 检查是否已连接
func (w *WeChatAdapter) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connected
}

// GetAccountInfo 获取微信账号信息
func (w *WeChatAdapter) GetAccountInfo() (accountId, nickname, avatar string, err error) {
	return w.accountID, "", "", nil
}

// SendMessage 向微信用户发送文本消息
func (w *WeChatAdapter) SendMessage(ctx context.Context, userID, content string, extra map[string]string) error {
	w.mu.RLock()
	token := w.token
	baseURL := w.baseURL
	w.mu.RUnlock()

	if token == "" {
		return fmt.Errorf("wechat not logged in")
	}

	// 提取 contextToken
	contextToken := ""
	if extra != nil {
		contextToken = extra["contextToken"]
	}

	// 分块发送（微信消息长度限制）
	chunks := chunkText(content, MaxTextChunkSize)
	for i, chunk := range chunks {
		if err := w.sendTextMessage(ctx, baseURL, token, userID, chunk, contextToken); err != nil {
			return fmt.Errorf("send chunk %d failed: %w", i, err)
		}
	}

	return nil
}

// OnMessage 注册消息回调
func (w *WeChatAdapter) OnMessage(handler types.MessageHandler) {
	w.handler = handler
}

// CreateStreamSender 创建流式回复发送器
// 微信不支持流式卡片，返回 nil 降级为非流式 SendMessage
func (w *WeChatAdapter) CreateStreamSender(ctx context.Context, userID string, extra map[string]string) types.StreamSender {
	return nil
}

// SetLoginInfo 设置登录信息（扫码登录成功后调用）
func (w *WeChatAdapter) SetLoginInfo(token, baseURL, accountID string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.token = token
	if baseURL != "" {
		w.baseURL = baseURL
	}
	w.accountID = accountID
	w.status = "logged_in"

	// 开始轮询
	if !w.polling {
		w.polling = true
		w.stopChan = make(chan struct{})
		go w.pollLoop(context.Background())
	}

	log.Infof("[WeChat] Login info set for account %s", accountID)
}

// GetStatus 获取状态
func (w *WeChatAdapter) GetStatus() string {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.status
}

// GetQRCode 获取登录二维码
func (w *WeChatAdapter) GetQRCode(ctx context.Context) (*QRCodeResult, error) {
	w.mu.RLock()
	baseURL := w.baseURL
	w.mu.RUnlock()

	apiURL := baseURL + "/ilink/bot/get_bot_qrcode?bot_type=3"

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("iLink-App-ClientVersion", "1")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get QR code: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get QR code failed: status=%d, body=%s", resp.StatusCode, string(body))
	}

	var result struct {
		QRCode           string `json:"qrcode"`
		QRCodeImgContent string `json:"qrcode_img_content"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse QR code response failed: %w", err)
	}

	return &QRCodeResult{
		QRCodeID:  result.QRCode,
		QRCodeURL: result.QRCodeImgContent,
	}, nil
}

// CheckQRStatus 检查二维码状态
func (w *WeChatAdapter) CheckQRStatus(ctx context.Context, qrcodeID string) (*QRStatusResult, error) {
	w.mu.RLock()
	baseURL := w.baseURL
	w.mu.RUnlock()

	apiURL := fmt.Sprintf("%s/ilink/bot/get_qrcode_status?qrcode=%s", baseURL, qrcodeID)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("iLink-App-ClientVersion", "1")

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("check QR status failed: status=%d", resp.StatusCode)
	}

	var result struct {
		Status      string `json:"status"`
		BotToken    string `json:"bot_token"`
		BaseURL     string `json:"base_url"`
		ILinkBotID  string `json:"ilink_bot_id"`
		ILinkUserID string `json:"ilink_user_id"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &QRStatusResult{
		Status:      result.Status,
		BotToken:    result.BotToken,
		BaseURL:     result.BaseURL,
		ILinkBotID:  result.ILinkBotID,
		ILinkUserID: result.ILinkUserID,
	}, nil
}

// --- 内部方法 ---

// pollLoop 消息轮询
func (w *WeChatAdapter) pollLoop(ctx context.Context) {
	log.Infof("[WeChat] Starting poll loop for account %s", w.accountID)
	backoff := 1 * time.Second
	maxBackoff := 30 * time.Second

	for w.polling {
		resp, err := w.getUpdates(ctx, w.pollBuf, DefaultLongPollTimeout*time.Second)
		if err != nil {
			log.Errorf("[WeChat] getUpdates error: %v", err)
			time.Sleep(backoff)
			if backoff*2 > maxBackoff {
				backoff = maxBackoff
			} else {
				backoff = backoff * 2
			}
			continue
		}

		if resp.ErrCode == SessionExpiredErrCode {
			log.Errorf("[WeChat] Session expired for account %s", w.accountID)
			w.mu.Lock()
			w.token = ""
			w.status = "session_expired"
			w.polling = false
			w.mu.Unlock()
			return
		}

		backoff = 1 * time.Second
		w.pollBuf = resp.GetUpdatesBuf

		for _, msg := range resp.Msgs {
			// 消息去重：用 from_user_id + context_token + create_time_ms 组合指纹
			msgFingerprint := fmt.Sprintf("%s:%s:%d", msg.FromUserID, msg.ContextToken, msg.CreateTimeMs)
			w.mu.Lock()
			if _, seen := w.seenMsgs[msgFingerprint]; seen {
				w.mu.Unlock()
				continue
			}
			w.seenMsgs[msgFingerprint] = struct{}{}
			// 限制去重缓存大小，避免内存泄漏
			if len(w.seenMsgs) > 1000 {
				w.seenMsgs = make(map[string]struct{})
			}
			w.mu.Unlock()

			platformMsg := w.convertMessage(&msg)
			if w.handler != nil {
				go func() {
					if err := w.handler(ctx, platformMsg); err != nil {
						log.Errorf("[WeChat] Message handler error: %v", err)
					}
				}()
			}
		}
	}

	log.Infof("[WeChat] Poll loop stopped for account %s", w.accountID)
}

// getUpdates 获取更新
func (w *WeChatAdapter) getUpdates(ctx context.Context, pollBuf string, timeout time.Duration) (*GetUpdatesResponse, error) {
	w.mu.RLock()
	token := w.token
	baseURL := w.baseURL
	w.mu.RUnlock()

	reqBody := map[string]interface{}{
		"base_info":       buildWeChatBaseInfo(),
		"get_updates_buf": pollBuf,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/ilink/bot/getupdates", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	w.setHeaders(httpReq, token)

	client := &http.Client{Timeout: timeout + 5*time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)

	var result GetUpdatesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// sendTextMessage 发送文本消息
func (w *WeChatAdapter) sendTextMessage(ctx context.Context, baseURL, token, toUserID, content, contextToken string) error {
	req := &SendRequest{
		BaseInfo: buildWeChatBaseInfo(),
		Msg: SendMsg{
			ToUserID:     toUserID,
			ClientID:     generateClientID(),
			MessageType:  2, // BOT
			MessageState: 2, // FINISH
			ContextToken: contextToken,
			ItemList: []MessageItem{{
				Type:     1, // Text
				TextItem: &TextItem{Text: content},
			}},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/ilink/bot/sendmessage", bytes.NewReader(body))
	if err != nil {
		return err
	}
	w.setHeaders(httpReq, token)

	resp, err := w.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send message failed: %s", string(respBody))
	}

	var sendResp SendResponse
	if err := json.Unmarshal(respBody, &sendResp); err == nil {
		if sendResp.Ret != 0 {
			return fmt.Errorf("send message failed: ret=%d, errmsg=%s", sendResp.Ret, sendResp.ErrMsg)
		}
	}

	return nil
}

// wechatILinkAppID iLink-App-Id，固定值（对齐 openclaw-weixin package.json 的 ilink_appid）。
const wechatILinkAppID = "bot"

// wechatILinkClientVersion iLink-App-ClientVersion，version "2.4.3" 编码为 0x00MMNNPP。
// (major<<16)|(minor<<8)|patch = (2<<16)|(4<<8)|3 = 132099。对齐 openclaw-weixin buildClientVersion。
const wechatILinkClientVersion = "132099"

// wechatChannelVersion base_info.channel_version，对齐 openclaw-weixin 插件版本。
const wechatChannelVersion = "2.4.3"

// setHeaders 设置请求头（对齐 openclaw-weixin buildHeaders）。
// X-WECHAT-UIN 每次请求随机生成（uint32→十进制字符串→base64）。
func (w *WeChatAdapter) setHeaders(req *http.Request, token string) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("AuthorizationType", "ilink_bot_token")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("iLink-App-Id", wechatILinkAppID)
	req.Header.Set("iLink-App-ClientVersion", wechatILinkClientVersion)
	req.Header.Set("X-WECHAT-UIN", randomWechatUIN())
}

// randomWeChatUIN 生成 X-WECHAT-UIN 头：随机 uint32 → 十进制字符串 → base64。
func randomWechatUIN() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		// 极少失败，回退固定值避免请求被拒
		return base64.StdEncoding.EncodeToString([]byte("0"))
	}
	uint32Val := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return base64.StdEncoding.EncodeToString([]byte(strconv.FormatUint(uint64(uint32Val), 10)))
}

// wechatBaseInfo 每个请求体都需携带的 base_info（对齐 openclaw-weixin buildBaseInfo）。
// channel_version 对齐插件版本，bot_agent 为默认 UA。
type wechatBaseInfo struct {
	ChannelVersion string `json:"channel_version"`
	BotAgent       string `json:"bot_agent"`
}

func buildWeChatBaseInfo() wechatBaseInfo {
	return wechatBaseInfo{
		ChannelVersion: wechatChannelVersion,
		BotAgent:       "OpenClaw",
	}
}

// convertMessage 转换 openclaw 消息为平台消息
func (w *WeChatAdapter) convertMessage(msg *WeixinMessage) *types.PlatformMessage {
	result := &types.PlatformMessage{
		ChannelID:   w.config.ChannelID,
		UserID:      msg.FromUserID,
		ChannelType: types.ChannelTypeWeChat,
		MsgType:     "text",
		Extra:       make(map[string]string),
	}

	// 提取文本内容
	for _, item := range msg.ItemList {
		switch item.Type {
		case 1: // Text
			if item.TextItem != nil {
				result.Content = item.TextItem.Text
			}
		case 2: // Image
			result.Content = "[图片]"
			result.MsgType = "image"
		case 3: // Voice
			if item.VoiceItem != nil && item.VoiceItem.Text != "" {
				result.Content = item.VoiceItem.Text
			} else {
				result.Content = "[语音]"
			}
			result.MsgType = "voice"
		case 4: // File
			if item.FileItem != nil {
				result.Content = item.FileItem.FileName
			}
			result.MsgType = "file"
		case 5: // Video
			result.Content = "[视频]"
			result.MsgType = "video"
		}
	}

	if result.Content == "" {
		result.Content = "[未知消息]"
	}

	// 保存上下文 token（用于回复同一会话）
	result.Extra["contextToken"] = msg.ContextToken
	result.Extra["fromUserId"] = msg.FromUserID

	return result
}

// chunkText 分块文本
func chunkText(text string, maxSize int) []string {
	if len(text) <= maxSize {
		return []string{text}
	}

	var chunks []string
	for len(text) > 0 {
		if len(text) <= maxSize {
			chunks = append(chunks, text)
			break
		}
		chunks = append(chunks, text[:maxSize])
		text = text[maxSize:]
	}
	return chunks
}

func generateClientID() string {
	return fmt.Sprintf("wanwu-%d", time.Now().UnixNano())
}

// --- openclaw 协议类型定义 ---

// QRCodeResult 二维码结果
type QRCodeResult struct {
	QRCodeID  string // 二维码 ID
	QRCodeURL string // 二维码图片 URL
	Base64    string // base64 编码的二维码图片
}

// QRStatusResult 二维码扫描状态
type QRStatusResult struct {
	Status      string // waiting/scanned/confirmed/expired
	BotToken    string
	BaseURL     string
	ILinkBotID  string
	ILinkUserID string
}

// WeixinMessage openclaw 微信消息格式
type WeixinMessage struct {
	FromUserID   string        `json:"from_user_id"`
	ToUserID     string        `json:"to_user_id"`
	MessageType  int           `json:"msg_type"`
	MessageState int           `json:"msg_state"`
	ContextToken string        `json:"context_token"`
	CreateTimeMs int64         `json:"create_time_ms"`
	ItemList     []MessageItem `json:"item_list"`
}

// MessageItem 消息项
type MessageItem struct {
	Type      int        `json:"type"`
	TextItem  *TextItem  `json:"text_item,omitempty"`
	ImageItem *ImageItem `json:"image_item,omitempty"`
	VoiceItem *VoiceItem `json:"voice_item,omitempty"`
	FileItem  *FileItem  `json:"file_item,omitempty"`
	VideoItem *VideoItem `json:"video_item,omitempty"`
}

// TextItem 文本项
type TextItem struct {
	Text string `json:"text"`
}

// ImageItem 图片项
type ImageItem struct {
	Media  *CDNMedia `json:"media,omitempty"`
	AESKey string    `json:"aes_key,omitempty"`
}

// VoiceItem 语音项
type VoiceItem struct {
	Media    *CDNMedia `json:"media,omitempty"`
	Playtime int       `json:"playtime,omitempty"`
	Text     string    `json:"text,omitempty"`
}

// FileItem 文件项
type FileItem struct {
	Media    *CDNMedia `json:"media,omitempty"`
	FileName string    `json:"file_name,omitempty"`
	Len      string    `json:"len,omitempty"`
	MD5      string    `json:"md5,omitempty"`
}

// VideoItem 视频项
type VideoItem struct {
	Media *CDNMedia `json:"media,omitempty"`
}

// CDNMedia CDN 媒体信息
type CDNMedia struct {
	MediaID string `json:"media_id"`
	AESKey  string `json:"aes_key"`
	URL     string `json:"url,omitempty"`
}

// SendRequest 发送消息请求
type SendRequest struct {
	BaseInfo wechatBaseInfo `json:"base_info"`
	Msg      SendMsg        `json:"msg"`
}

// SendMsg 发送消息
type SendMsg struct {
	FromUserID   string        `json:"from_user_id,omitempty"`
	ToUserID     string        `json:"to_user_id"`
	ClientID     string        `json:"client_id"`
	MessageType  int           `json:"msg_type"`
	MessageState int           `json:"msg_state"`
	ContextToken string        `json:"context_token,omitempty"`
	ItemList     []MessageItem `json:"item_list"`
}

// SendResponse 发送消息响应
type SendResponse struct {
	Ret    int    `json:"ret"`
	ErrMsg string `json:"errmsg,omitempty"`
}

// GetUpdatesResponse 获取更新响应
type GetUpdatesResponse struct {
	Ret           int             `json:"ret"`
	ErrCode       int             `json:"errcode"`
	ErrMsg        string          `json:"errmsg"`
	GetUpdatesBuf string          `json:"get_updates_buf"`
	Msgs          []WeixinMessage `json:"msgs"`
}

// --- 文件发送（openclaw iLink Bot 3 步流程）---
//
// 微信发文件/图片走 iLink Bot API：
//  1. getuploadurl：申请 CDN 上传地址（携带 AES 密钥、MD5、加密后大小等）
//  2. AES-128-ECB 加密文件 → PUT/POST 到 CDN，响应头 x-encrypted-param 为下载凭证
//  3. sendmessage：携带 CDN 下载凭证 + AES 密钥发送 file_item(type=4) / image_item(type=2)
//
// 参考 openclaw-weixin 协议。media_type：1=IMAGE, 2=VIDEO, 3=FILE。

const (
	ilinkMediaImage = 1
	ilinkMediaVideo = 2
	ilinkMediaFile  = 3

	ilinkItemImage = 2
	ilinkItemFile  = 4

	ilinkEncryptType = 1 // AES-128-ECB

	// wechatCDNBaseURL 微信 CDN 基址（对齐 openclaw-weixin CDN_BASE_URL）。
	// getuploadurl 未返回 upload_full_url 时，用 upload_param + 此基址拼接上传 URL：
	//   {cdnBaseUrl}/upload?encrypted_query_param={upload_param}&filekey={filekey}
	wechatCDNBaseURL = "https://novac2c.cdn.weixin.qq.com/c2c"
)

// fileUploadURLReq getuploadurl 请求体
type fileUploadURLReq struct {
	BaseInfo    wechatBaseInfo `json:"base_info"`
	Filekey     string         `json:"filekey"`
	MediaType   int            `json:"media_type"`
	ToUserID    string         `json:"to_user_id"`
	RawSize     int64          `json:"rawsize"`
	RawFileMD5  string         `json:"rawfilemd5"`
	FileSize    int64          `json:"filesize"` // AES 加密后大小
	NoNeedThumb bool           `json:"no_need_thumb"`
	AESKey      string         `json:"aeskey"` // 16 字节 hex
}

// fileUploadURLResp getuploadurl 响应
// 服务端二选一返回上传地址：upload_full_url（完整 URL，优先用）或 upload_param（需客户端拼接）。
type fileUploadURLResp struct {
	Ret           int    `json:"ret"`
	ErrMsg        string `json:"errmsg,omitempty"`
	UploadFullURL string `json:"upload_full_url"`
	UploadParam   string `json:"upload_param"` // 服务端只给参数时用，配合 wechatCDNBaseURL 拼接
}

// sendCDNMedia 发送方向的 CDN 媒体信息（字段与接收方向 CDNMedia 不同）
type sendCDNMedia struct {
	EncryptQueryParam string `json:"encrypt_query_param"`
	AESKey            string `json:"aes_key"` // base64 编码的 AES 密钥
	EncryptType       int    `json:"encrypt_type"`
}

// sendFileItem 发送方向的文件项
type sendFileItem struct {
	Media    *sendCDNMedia `json:"media,omitempty"`
	FileName string        `json:"file_name,omitempty"`
	Len      string        `json:"len,omitempty"` // 原始文件大小（字符串）
}

// sendImageItem 发送方向的图片项
type sendImageItem struct {
	Media   *sendCDNMedia `json:"media,omitempty"`
	MidSize int64         `json:"mid_size,omitempty"` // 加密后文件大小
}

// SendFile 向微信用户发送文件/图片附件（实现 types.FileSender）。
// 图片（png/jpg）走 image_item，其余走 file_item。均经 AES-128-ECB 加密 + CDN 中转。
func (w *WeChatAdapter) SendFile(ctx context.Context, userID, fileName, mimeType string, data []byte, extra map[string]string) error {
	w.mu.RLock()
	token := w.token
	baseURL := w.baseURL
	w.mu.RUnlock()

	if token == "" {
		return fmt.Errorf("wechat not logged in")
	}
	if len(data) == 0 {
		return fmt.Errorf("empty file data")
	}

	contextToken := ""
	if extra != nil {
		contextToken = extra["contextToken"]
	}

	// 1. 生成 filekey + AES 密钥（各 16 字节随机）
	aesKey := make([]byte, 16)
	if _, err := rand.Read(aesKey); err != nil {
		return fmt.Errorf("generate aes key failed: %w", err)
	}
	filekeyBytes := make([]byte, 16)
	if _, err := rand.Read(filekeyBytes); err != nil {
		return fmt.Errorf("generate filekey failed: %w", err)
	}
	filekey := hex.EncodeToString(filekeyBytes)

	// 2. AES-128-ECB 加密文件
	encData, err := aesECBEncrypt(aesKey, data)
	if err != nil {
		return fmt.Errorf("aes encrypt failed: %w", err)
	}

	// 3. 原始文件 MD5
	md5sum := md5.Sum(data)
	rawFileMD5 := hex.EncodeToString(md5sum[:])

	// 4. 判定媒体类型：图片走 image，其余走 file
	isImage := isImageFile(fileName, mimeType)
	mediaType := ilinkMediaFile
	if isImage {
		mediaType = ilinkMediaImage
	}

	// 5. getuploadurl
	uploadURL, err := w.getUploadURL(ctx, baseURL, token, fileUploadURLReq{
		BaseInfo:    buildWeChatBaseInfo(),
		Filekey:     filekey,
		MediaType:   mediaType,
		ToUserID:    userID,
		RawSize:     int64(len(data)),
		RawFileMD5:  rawFileMD5,
		FileSize:    int64(len(encData)),
		NoNeedThumb: true,
		AESKey:      hex.EncodeToString(aesKey),
	})
	if err != nil {
		return fmt.Errorf("get upload url failed: %w", err)
	}

	// 6. 上传加密后数据到 CDN，拿下载凭证
	encryptQueryParam, err := w.uploadToCDN(ctx, uploadURL, encData)
	if err != nil {
		return fmt.Errorf("upload to cdn failed: %w", err)
	}

	// 7. sendmessage 携带 CDN 凭证。
	// aes_key 编码对齐 openclaw-weixin：base64( hex字符串 的 UTF-8 字节 )，
	// 即先 hex.EncodeToString(aesKey) 得 32 字符 hex，再对其字节做 base64（非对原始 16 字节做 base64）。
	aesKeyHex := hex.EncodeToString(aesKey)
	media := &sendCDNMedia{
		EncryptQueryParam: encryptQueryParam,
		AESKey:            base64.StdEncoding.EncodeToString([]byte(aesKeyHex)),
		EncryptType:       ilinkEncryptType,
	}
	return w.sendFileMessage(ctx, baseURL, token, userID, contextToken, fileName, isImage, media, int64(len(data)), int64(len(encData)))
}

// sendFileMessage 发送携带 CDN 媒体的消息（图片/文件）。
func (w *WeChatAdapter) sendFileMessage(ctx context.Context, baseURL, token, toUserID, contextToken, fileName string,
	isImage bool, media *sendCDNMedia, rawSize, encSize int64) error {

	type sendMsgItem struct {
		Type      int            `json:"type"`
		ImageItem *sendImageItem `json:"image_item,omitempty"`
		FileItem  *sendFileItem  `json:"file_item,omitempty"`
	}
	type sendMsg struct {
		ToUserID     string        `json:"to_user_id"`
		ClientID     string        `json:"client_id"`
		MessageType  int           `json:"msg_type"`
		MessageState int           `json:"msg_state"`
		ContextToken string        `json:"context_token,omitempty"`
		ItemList     []sendMsgItem `json:"item_list"`
	}
	type sendReq struct {
		BaseInfo wechatBaseInfo `json:"base_info"`
		Msg      sendMsg        `json:"msg"`
	}

	item := sendMsgItem{Type: ilinkItemFile}
	if isImage {
		item.Type = ilinkItemImage
		item.ImageItem = &sendImageItem{Media: media, MidSize: encSize}
	} else {
		item.FileItem = &sendFileItem{Media: media, FileName: fileName, Len: fmt.Sprintf("%d", rawSize)}
	}

	req := sendReq{
		BaseInfo: buildWeChatBaseInfo(),
		Msg: sendMsg{
			ToUserID:     toUserID,
			ClientID:     generateClientID(),
			MessageType:  2, // BOT
			MessageState: 2, // FINISH
			ContextToken: contextToken,
			ItemList:     []sendMsgItem{item},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/ilink/bot/sendmessage", bytes.NewReader(body))
	if err != nil {
		return err
	}
	w.setHeaders(httpReq, token)

	resp, err := w.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send file message failed: %s", string(respBody))
	}

	var sendResp SendResponse
	if err := json.Unmarshal(respBody, &sendResp); err == nil {
		if sendResp.Ret != 0 {
			return fmt.Errorf("send file message failed: ret=%d, errmsg=%s", sendResp.Ret, sendResp.ErrMsg)
		}
	}
	log.Infof("[WeChat] SendFile ok: user=%s, fileName=%s, isImage=%v, rawSize=%d", toUserID, fileName, isImage, rawSize)
	return nil
}

// getUploadURL 调用 getuploadurl 拿 CDN 上传地址。
func (w *WeChatAdapter) getUploadURL(ctx context.Context, baseURL, token string, reqBody fileUploadURLReq) (string, error) {
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/ilink/bot/getuploadurl", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	w.setHeaders(httpReq, token)

	resp, err := w.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getuploadurl http %d: %s", resp.StatusCode, string(respBody))
	}

	var ur fileUploadURLResp
	if err := json.Unmarshal(respBody, &ur); err != nil {
		return "", fmt.Errorf("parse getuploadurl resp failed: %w, raw=%s", err, truncateWechat(string(respBody)))
	}

	// 服务端二选一返回上传地址：
	//  1. upload_full_url（完整 URL，优先用）
	//  2. upload_param（仅参数，需客户端拼接：{cdnBaseUrl}/upload?encrypted_query_param={param}&filekey={filekey}）
	// 对齐 openclaw-weixin cdn-url.ts buildCdnUploadUrl。
	if full := strings.TrimSpace(ur.UploadFullURL); full != "" {
		return full, nil
	}
	if param := strings.TrimSpace(ur.UploadParam); param != "" {
		return fmt.Sprintf("%s/upload?encrypted_query_param=%s&filekey=%s",
			wechatCDNBaseURL, url.QueryEscape(param), url.QueryEscape(reqBody.Filekey)), nil
	}
	return "", fmt.Errorf("getuploadurl returned no upload URL (need upload_full_url or upload_param), ret=%d, errmsg=%s, raw=%s",
		ur.Ret, ur.ErrMsg, truncateWechat(string(respBody)))
}

// uploadToCDN 上传加密后的文件字节到 CDN，返回下载凭证（x-encrypted-param 响应头）。
func (w *WeChatAdapter) uploadToCDN(ctx context.Context, uploadURL string, encData []byte) (string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewReader(encData))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/octet-stream")

	resp, err := w.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("cdn upload http %d: %s", resp.StatusCode, truncateWechat(string(body)))
	}

	// 下载凭证在响应头 x-encrypted-param（部分实现也可能放在响应体，二者择一）
	encryptQueryParam := resp.Header.Get("x-encrypted-param")
	if encryptQueryParam == "" {
		// 兜底：尝试从响应体解析
		body, _ := io.ReadAll(resp.Body)
		var b struct {
			EncryptQueryParam string `json:"encrypt_query_param"`
			XEncryptedParam   string `json:"x-encrypted-param"`
		}
		if json.Unmarshal(body, &b) == nil {
			encryptQueryParam = b.EncryptQueryParam
			if encryptQueryParam == "" {
				encryptQueryParam = b.XEncryptedParam
			}
		}
	}
	if encryptQueryParam == "" {
		return "", fmt.Errorf("cdn upload ok but empty encrypt_query_param")
	}
	return encryptQueryParam, nil
}

// aesECBEncrypt AES-128-ECB 加密（PKCS7 padding）。Go 标准库未直接提供 ECB，逐块加密。
func aesECBEncrypt(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	// PKCS7 padding
	padding := bs - len(plaintext)%bs
	padded := make([]byte, len(plaintext)+padding)
	copy(padded, plaintext)
	for i := len(plaintext); i < len(padded); i++ {
		padded[i] = byte(padding)
	}
	ciphertext := make([]byte, len(padded))
	for start := 0; start < len(padded); start += bs {
		block.Encrypt(ciphertext[start:start+bs], padded[start:start+bs])
	}
	return ciphertext, nil
}

// isImageFile 判定是否图片（按扩展名 + mimeType）。
func isImageFile(fileName, mimeType string) bool {
	ext := strings.ToLower(strings.TrimPrefix(path.Ext(fileName), "."))
	switch ext {
	case "png", "jpg", "jpeg", "gif", "bmp", "webp":
		return true
	}
	return strings.HasPrefix(strings.ToLower(mimeType), "image/")
}

// truncateWechat 截断日志字符串。
func truncateWechat(s string) string {
	const max = 300
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
