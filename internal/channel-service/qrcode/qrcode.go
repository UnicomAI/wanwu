package qrcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/types"
	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/wechat"
	"github.com/UnicomAI/wanwu/internal/channel-service/client"
	"github.com/UnicomAI/wanwu/internal/channel-service/client/model"
	"github.com/UnicomAI/wanwu/internal/channel-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
)

// QRLoginManager 扫码登录管理器
type QRLoginManager struct {
	cfg config.Config
	cli client.IClient
}

// NewQRLoginManager 创建扫码登录管理器
func NewQRLoginManager(cfg config.Config, cli client.IClient) *QRLoginManager {
	return &QRLoginManager{
		cfg: cfg,
		cli: cli,
	}
}

// StartCleanup 启动过期扫码会话的定时清理协程。
// 每隔 cleanupInterval 扫描一次，删除 expire_at < now 的 qr_sessions 记录。
// 返回的 stop 函数用于优雅停止清理协程（传入的 ctx 取消同样会停止）。
// 建议在服务启动时调用，服务停止时调用返回的 stop。
func (m *QRLoginManager) StartCleanup(ctx context.Context, cleanupInterval time.Duration) (stop func()) {
	if cleanupInterval <= 0 {
		cleanupInterval = 10 * time.Minute
	}

	ticker := time.NewTicker(cleanupInterval)
	done := make(chan struct{})

	go func() {
		defer close(done)
		// 启动时立即清理一次
		m.cleanupExpiredSessions(ctx)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				m.cleanupExpiredSessions(ctx)
			}
		}
	}()

	return func() {
		ticker.Stop()
		<-done
	}
}

// cleanupExpiredSessions 删除所有已过期的扫码会话
func (m *QRLoginManager) cleanupExpiredSessions(ctx context.Context) {
	now := time.Now().Unix()
	deleted, err := m.cli.DeleteExpiredQRSessions(ctx, now)
	if err != nil {
		log.Errorf("[QRLogin] failed to cleanup expired sessions: %v", err)
		return
	}
	if deleted > 0 {
		log.Infof("[QRLogin] cleanup expired qr sessions: deleted %d records", deleted)
	}
}

// QRLoginResult 扫码登录结果
type QRLoginResult struct {
	SessionID  string
	QrUrl      string // 二维码 URL（所有平台通用，前端用此字段生成/展示二维码）
	ExpireAt   int64  // 过期时间 Unix 时间戳
	ExpireTime int64  // 过期时长（秒）
}

// CreateQRLogin 发起扫码登录
func (m *QRLoginManager) CreateQRLogin(ctx context.Context, channelType, userID, orgID string) (*QRLoginResult, error) {
	switch channelType {
	case "dingtalk":
		return m.createDingTalkQRLogin(ctx, userID, orgID)
	case "wechat":
		return m.createWechatQRLogin(ctx, userID, orgID)
	case "feishu":
		return m.createFeishuQRLogin(ctx, userID, orgID)
	default:
		return nil, fmt.Errorf("unsupported channel type: %s", channelType)
	}
}

// GetQRLoginStatus 查询扫码状态
func (m *QRLoginManager) GetQRLoginStatus(ctx context.Context, channelType, sessionID string) (string, map[string]string, error) {
	session, err := m.cli.GetQRSession(ctx, sessionID)
	if err != nil {
		return "", nil, fmt.Errorf("qr session not found: %v", err)
	}

	// 校验 channelType 与 session 中存储的类型一致
	if session.ChannelType != channelType {
		return "", nil, fmt.Errorf("channel type mismatch: session is %s but request is %s", session.ChannelType, channelType)
	}

	// 检查过期
	if time.Now().Unix() > session.ExpireAt {
		_ = m.cli.UpdateQRSession(ctx, sessionID, map[string]interface{}{"status": "expired"})
		return "expired", nil, nil
	}

	// 如果是等待中状态，尝试从 Redis 检查更新
	if session.Status == "waiting" || session.Status == "scanned" {
		cachedStatus := m.checkRedisStatus(channelType, sessionID)
		if cachedStatus != "" && cachedStatus != session.Status {
			_ = m.cli.UpdateQRSession(ctx, sessionID, map[string]interface{}{"status": cachedStatus})
			session.Status = cachedStatus
		}
	}

	// 解析凭据
	var credentials map[string]string
	if session.Status == "success" && session.Credentials != "" {
		_ = json.Unmarshal([]byte(session.Credentials), &credentials)
	}

	return session.Status, credentials, nil
}

// CancelQRLogin 取消扫码登录
func (m *QRLoginManager) CancelQRLogin(ctx context.Context, channelType, sessionID string) error {
	// 校验 channelType 与 session 中存储的类型一致
	session, err := m.cli.GetQRSession(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("qr session not found: %v", err)
	}
	if session.ChannelType != channelType {
		return fmt.Errorf("channel type mismatch: session is %s but request is %s", session.ChannelType, channelType)
	}
	return m.cli.DeleteQRSession(ctx, sessionID)
}

// --- 钉钉扫码登录（Device Flow） ---

const dingtalkBaseURL = "https://oapi.dingtalk.com"

// dingtalkInitResponse 钉钉 Device Flow Init 响应
type dingtalkInitResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
	Nonce   string `json:"nonce"`
}

// dingtalkBeginResponse 钉钉 Device Flow Begin 响应
type dingtalkBeginResponse struct {
	ErrCode                 int    `json:"errcode"`
	ErrMsg                  string `json:"errmsg"`
	DeviceCode              string `json:"device_code"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// dingtalkPollResponse 钉钉 Device Flow Poll 响应
type dingtalkPollResponse struct {
	ErrCode      int    `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
	Status       string `json:"status"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	FailReason   string `json:"fail_reason"`
}

// 钉钉 poll 可重试的 errcode（表示用户还未扫码或正在处理中）
var dingtalkRetryErrCodes = map[int]bool{
	40078: true, // authorization_pending - 等待用户扫码
	40079: true, // slow_down - 轮询过快
}

func (m *QRLoginManager) createDingTalkQRLogin(ctx context.Context, userID, orgID string) (*QRLoginResult, error) {
	sessionID := model.NewSessionID()

	// Step 1: Init — 初始化 Device Flow
	initBody, _ := json.Marshal(map[string]string{"source": "wanwu"})
	initResp, err := http.Post(dingtalkBaseURL+"/app/registration/init", "application/json", bytes.NewReader(initBody))
	if err != nil {
		return nil, fmt.Errorf("钉钉初始化失败: %w", err)
	}
	defer func() { _ = initResp.Body.Close() }()

	initData, err := io.ReadAll(initResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取钉钉初始化响应失败: %w", err)
	}

	var initResult dingtalkInitResponse
	if err := json.Unmarshal(initData, &initResult); err != nil {
		return nil, fmt.Errorf("解析钉钉初始化响应失败: %w", err)
	}
	if initResult.ErrCode != 0 || initResult.Nonce == "" {
		return nil, fmt.Errorf("钉钉初始化失败: errcode=%d, errmsg=%s", initResult.ErrCode, initResult.ErrMsg)
	}

	// Step 2: Begin — 获取二维码 URL 和 device_code
	beginBody, _ := json.Marshal(map[string]string{"nonce": initResult.Nonce})
	beginResp, err := http.Post(dingtalkBaseURL+"/app/registration/begin", "application/json", bytes.NewReader(beginBody))
	if err != nil {
		return nil, fmt.Errorf("获取钉钉二维码失败: %w", err)
	}
	defer func() { _ = beginResp.Body.Close() }()

	beginData, err := io.ReadAll(beginResp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取钉钉二维码响应失败: %w", err)
	}

	var beginResult dingtalkBeginResponse
	if err := json.Unmarshal(beginData, &beginResult); err != nil {
		return nil, fmt.Errorf("解析钉钉二维码响应失败: %w", err)
	}
	if beginResult.ErrCode != 0 || beginResult.DeviceCode == "" {
		return nil, fmt.Errorf("获取钉钉二维码失败: errcode=%d, errmsg=%s", beginResult.ErrCode, beginResult.ErrMsg)
	}

	qrURL := beginResult.VerificationURIComplete
	expireTime := int64(beginResult.ExpiresIn)
	if expireTime == 0 {
		expireTime = 600 // 默认 10 分钟
	}
	expireAt := time.Now().Unix() + expireTime

	// 保存扫码会话（含 device_code 用于轮询）
	session := &model.QRSession{
		SessionID:   sessionID,
		ChannelType: "dingtalk",
		Status:      "waiting",
		Credentials: "",
		DeviceCode:  beginResult.DeviceCode,
		UserID:      userID,
		OrgID:       orgID,
		ExpireAt:    expireAt,
	}
	if _, err := m.cli.CreateQRSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create qr session: %w", err)
	}

	// 在 Redis 中设置扫码状态缓存
	m.setRedisStatus("dingtalk", sessionID, "waiting", expireTime)

	// 异步轮询钉钉扫码状态
	go m.pollDingTalkQRLogin(sessionID, beginResult.DeviceCode, expireTime)

	return &QRLoginResult{
		SessionID:  sessionID,
		QrUrl:      qrURL,
		ExpireAt:   expireAt,
		ExpireTime: expireTime,
	}, nil
}

// pollDingTalkQRLogin 轮询钉钉扫码登录状态
func (m *QRLoginManager) pollDingTalkQRLogin(sessionID, deviceCode string, expiresIn int64) {
	interval := 2 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	timeout := time.After(time.Duration(expiresIn) * time.Second)

	for {
		select {
		case <-timeout:
			_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
				"status": "expired",
			})
			m.setRedisStatus("dingtalk", sessionID, "expired", 300)
			return
		case <-ticker.C:
			pollBody, _ := json.Marshal(map[string]string{"device_code": deviceCode})
			pollResp, err := http.Post(dingtalkBaseURL+"/app/registration/poll", "application/json", bytes.NewReader(pollBody))
			if err != nil {
				log.Errorf("poll dingtalk qr login failed: %v", err)
				continue
			}

			var pollData dingtalkPollResponse
			body, _ := io.ReadAll(pollResp.Body)
			_ = pollResp.Body.Close()
			if err := json.Unmarshal(body, &pollData); err != nil {
				log.Errorf("parse dingtalk poll response failed: %v", err)
				continue
			}

			// 钉钉 poll 在用户未扫码时返回非零 errcode（如 40078 authorization_pending）
			// 这是正常等待状态，不应当作错误，继续轮询即可
			if pollData.ErrCode != 0 {
				if dingtalkRetryErrCodes[pollData.ErrCode] {
					// authorization_pending / slow_down → 继续轮询
					continue
				}
				// 其他错误码才是真正的失败
				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status": "error",
				})
				m.setRedisStatus("dingtalk", sessionID, "error", 300)
				log.Errorf("dingtalk poll error: errcode=%d, errmsg=%s", pollData.ErrCode, pollData.ErrMsg)
				return
			}

			switch pollData.Status {
			case "SUCCESS":
				// 加密存储凭据
				credentials := map[string]string{
					"client_id":     pollData.ClientID,
					"client_secret": pollData.ClientSecret,
				}
				credJSON, _ := json.Marshal(credentials)

				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status":      "success",
					"credentials": string(credJSON),
				})
				m.setRedisStatus("dingtalk", sessionID, "success", 300)
				log.Infof("dingtalk qr login success for session %s", sessionID)
				return

			case "FAIL":
				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status": "error",
				})
				m.setRedisStatus("dingtalk", sessionID, "error", 300)
				log.Errorf("dingtalk qr login failed: %s", pollData.FailReason)
				return

			case "EXPIRED":
				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status": "expired",
				})
				m.setRedisStatus("dingtalk", sessionID, "expired", 300)
				return

			default:
				// waiting / scanned 等状态，继续轮询
				continue
			}
		}
	}
}

// HandleDingTalkCallback 处理钉钉扫码登录回调
// 钉钉 OAuth 回调时调用此方法更新 session 状态
func (m *QRLoginManager) HandleDingTalkCallback(ctx context.Context, sessionID string, appKey, appSecret string) error {
	// 加密存储凭据
	credentials := map[string]string{
		"appKey":    appKey,
		"appSecret": appSecret,
	}
	credJSON, _ := json.Marshal(credentials)

	// 更新 session 状态
	if err := m.cli.UpdateQRSession(ctx, sessionID, map[string]interface{}{
		"status":      "success",
		"credentials": string(credJSON),
	}); err != nil {
		return fmt.Errorf("failed to update qr session: %w", err)
	}

	// 更新 Redis 状态
	m.setRedisStatus("dingtalk", sessionID, "success", 300)

	log.Infof("dingtalk qr login success for session %s", sessionID)
	return nil
}

// --- 微信扫码登录（iLink/OpenClaw 协议） ---

func (m *QRLoginManager) createWechatQRLogin(ctx context.Context, userID, orgID string) (*QRLoginResult, error) {
	// 使用 WeChatAdapter 获取二维码
	baseURL := m.cfg.WeChat.BaseURL
	if baseURL == "" {
		baseURL = wechat.DefaultBaseURL
	}

	adapter := wechat.NewWeChatAdapter()
	_ = adapter.Connect(types.AdapterConfig{
		BaseUrl: baseURL,
	})

	qr, err := adapter.GetQRCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("获取微信二维码失败: %w", err)
	}

	sessionID := model.NewSessionID()
	expireTime := int64(180) // 3 分钟（微信 iLink 平台二维码有效期较短）
	expireAt := time.Now().Unix() + expireTime

	// 保存扫码会话（含 qrcode_id 用于轮询）
	session := &model.QRSession{
		SessionID:   sessionID,
		ChannelType: "wechat",
		Status:      "waiting",
		Credentials: "",
		DeviceCode:  qr.QRCodeID, // 复用 DeviceCode 字段存储 qrcode_id
		UserID:      userID,
		OrgID:       orgID,
		ExpireAt:    expireAt,
	}
	if _, err := m.cli.CreateQRSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create qr session: %w", err)
	}

	// 在 Redis 中设置扫码状态缓存
	m.setRedisStatus("wechat", sessionID, "waiting", expireTime)

	// 异步轮询微信扫码状态
	go m.pollWechatQRLogin(sessionID, qr.QRCodeID, baseURL, expireTime)

	return &QRLoginResult{
		SessionID:  sessionID,
		QrUrl:      qr.QRCodeURL,
		ExpireAt:   expireAt,
		ExpireTime: expireTime,
	}, nil
}

// pollWechatQRLogin 轮询微信扫码登录状态
func (m *QRLoginManager) pollWechatQRLogin(sessionID, qrcodeID, baseURL string, expiresIn int64) {
	adapter := wechat.NewWeChatAdapter()
	_ = adapter.Connect(types.AdapterConfig{
		BaseUrl: baseURL,
	})

	interval := 2 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	timeout := time.After(time.Duration(expiresIn) * time.Second)

	for {
		select {
		case <-timeout:
			_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
				"status": "expired",
			})
			m.setRedisStatus("wechat", sessionID, "expired", 300)
			return
		case <-ticker.C:
			status, err := adapter.CheckQRStatus(context.Background(), qrcodeID)
			if err != nil {
				log.Errorf("poll wechat qr login failed: %v", err)
				continue
			}

			// 兼容 OpenClaw 可能返回的状态值：confirmed / success / logged_in
			isConfirmed := status.Status == "confirmed" || status.Status == "success" || status.Status == "logged_in"

			if isConfirmed && status.BotToken != "" {
				// 如果微信 API 未返回 base_url，使用默认值
				baseURL := status.BaseURL
				if baseURL == "" {
					baseURL = wechat.DefaultBaseURL
				}

				// 加密存储凭据
				credentials := map[string]string{
					"token":     status.BotToken,
					"baseUrl":   baseURL,
					"accountId": status.ILinkBotID,
				}
				credJSON, _ := json.Marshal(credentials)

				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status":      "success",
					"credentials": string(credJSON),
				})
				m.setRedisStatus("wechat", sessionID, "success", 300)
				log.Infof("wechat qr login success for session %s", sessionID)
				return
			}

			if status.Status == "expired" {
				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status": "expired",
				})
				m.setRedisStatus("wechat", sessionID, "expired", 300)
				return
			}

			// waiting / scanned 等状态，继续轮询
		}
	}
}

// --- 飞书扫码登录（Device Flow） ---

const feishuBaseURL = "https://accounts.feishu.cn"

// feishuInitResponse 飞书 Device Flow Init 响应
type feishuInitResponse struct {
	SupportedAuthMethods []string `json:"supported_auth_methods"`
}

// feishuBeginResponse 飞书 Device Flow Begin 响应
type feishuBeginResponse struct {
	DeviceCode              string `json:"device_code"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	Error                   string `json:"error"`
	ErrorDescription        string `json:"error_description"`
}

// feishuPollResponse 飞书 Device Flow Poll 响应
type feishuPollResponse struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	UserInfo     struct {
		OpenID      string `json:"open_id"`
		TenantBrand string `json:"tenant_brand"`
	} `json:"user_info"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (m *QRLoginManager) createFeishuQRLogin(ctx context.Context, userID, orgID string) (*QRLoginResult, error) {
	// Step 1: Init — 初始化 Device Flow
	initBody := "action=init"
	initResp, err := http.Post(feishuBaseURL+"/oauth/v1/app/registration", "application/x-www-form-urlencoded", bytes.NewReader([]byte(initBody)))
	if err != nil {
		return nil, fmt.Errorf("飞书初始化失败: %w", err)
	}
	defer func() { _ = initResp.Body.Close() }()

	var initData feishuInitResponse
	body, _ := io.ReadAll(initResp.Body)
	if err := json.Unmarshal(body, &initData); err != nil {
		return nil, fmt.Errorf("解析飞书初始化响应失败: %w", err)
	}

	supported := false
	for _, method := range initData.SupportedAuthMethods {
		if method == "client_secret" {
			supported = true
			break
		}
	}
	if !supported {
		return nil, fmt.Errorf("飞书不支持 client_secret 认证方法")
	}

	// Step 2: Begin — 获取二维码 URL 和 device_code
	beginBody := "action=begin&archetype=PersonalAgent&auth_method=client_secret&request_user_info=open_id"
	beginResp, err := http.Post(feishuBaseURL+"/oauth/v1/app/registration", "application/x-www-form-urlencoded", bytes.NewReader([]byte(beginBody)))
	if err != nil {
		return nil, fmt.Errorf("获取飞书二维码失败: %w", err)
	}
	defer func() { _ = beginResp.Body.Close() }()

	var beginData feishuBeginResponse
	body, _ = io.ReadAll(beginResp.Body)
	if err := json.Unmarshal(body, &beginData); err != nil {
		return nil, fmt.Errorf("解析飞书二维码响应失败: %w", err)
	}

	if beginData.Error != "" {
		return nil, fmt.Errorf("飞书获取二维码失败: %s: %s", beginData.Error, beginData.ErrorDescription)
	}
	if beginData.DeviceCode == "" {
		return nil, fmt.Errorf("获取飞书二维码失败: device_code 为空")
	}

	qrURL := beginData.VerificationURIComplete
	if qrURL != "" {
		qrURL = qrURL + "&from=sdk&tp=sdk&source=wanwu"
	}

	sessionID := model.NewSessionID()
	expireTime := int64(beginData.ExpiresIn)
	if expireTime == 0 {
		expireTime = 600 // 默认 10 分钟
	}
	expireAt := time.Now().Unix() + expireTime

	// 保存扫码会话
	session := &model.QRSession{
		SessionID:   sessionID,
		ChannelType: "feishu",
		Status:      "waiting",
		Credentials: "",
		DeviceCode:  beginData.DeviceCode,
		UserID:      userID,
		OrgID:       orgID,
		ExpireAt:    expireAt,
	}
	if _, err := m.cli.CreateQRSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create qr session: %w", err)
	}

	// 在 Redis 中设置扫码状态缓存
	m.setRedisStatus("feishu", sessionID, "waiting", expireTime)

	// 异步轮询飞书扫码状态
	go m.pollFeishuQRLogin(sessionID, beginData.DeviceCode, expireTime)

	return &QRLoginResult{
		SessionID:  sessionID,
		QrUrl:      qrURL,
		ExpireAt:   expireAt,
		ExpireTime: expireTime,
	}, nil
}

// pollFeishuQRLogin 轮询飞书扫码登录状态
func (m *QRLoginManager) pollFeishuQRLogin(sessionID, deviceCode string, expiresIn int64) {
	interval := 5 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	timeout := time.After(time.Duration(expiresIn) * time.Second)

	for {
		select {
		case <-timeout:
			_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
				"status": "expired",
			})
			m.setRedisStatus("feishu", sessionID, "expired", 300)
			return
		case <-ticker.C:
			pollBody := fmt.Sprintf("action=poll&device_code=%s", deviceCode)
			pollResp, err := http.Post(feishuBaseURL+"/oauth/v1/app/registration", "application/x-www-form-urlencoded", bytes.NewReader([]byte(pollBody)))
			if err != nil {
				log.Errorf("poll feishu qr login failed: %v", err)
				continue
			}

			var pollData feishuPollResponse
			body, _ := io.ReadAll(pollResp.Body)
			_ = pollResp.Body.Close()
			if err := json.Unmarshal(body, &pollData); err != nil {
				log.Errorf("parse feishu poll response failed: %v", err)
				continue
			}

			// 成功获取凭据
			if pollData.ClientID != "" && pollData.ClientSecret != "" {
				credentials := map[string]string{
					"appId":     pollData.ClientID,
					"appSecret": pollData.ClientSecret,
				}
				credJSON, _ := json.Marshal(credentials)

				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status":      "success",
					"credentials": string(credJSON),
				})
				m.setRedisStatus("feishu", sessionID, "success", 300)
				log.Infof("feishu qr login success for session %s", sessionID)
				return
			}

			switch pollData.Error {
			case "authorization_pending":
				// 等待用户扫码
				continue
			case "slow_down":
				// 轮询过快，增加间隔
				interval = interval + 2*time.Second
				ticker.Reset(interval)
				continue
			case "access_denied":
				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status": "error",
				})
				m.setRedisStatus("feishu", sessionID, "error", 300)
				log.Errorf("feishu qr login denied for session %s", sessionID)
				return
			case "expired_token":
				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status": "expired",
				})
				m.setRedisStatus("feishu", sessionID, "expired", 300)
				return
			case "":
				// 无错误但也没有凭据，继续轮询
				continue
			default:
				_ = m.cli.UpdateQRSession(context.Background(), sessionID, map[string]interface{}{
					"status": "error",
				})
				m.setRedisStatus("feishu", sessionID, "error", 300)
				log.Errorf("feishu qr login error: %s: %s", pollData.Error, pollData.ErrorDescription)
				return
			}
		}
	}
}

// --- Redis 状态管理 ---

func (m *QRLoginManager) setRedisStatus(channelType, sessionID, status string, ttl int64) {
	if redis.OP() == nil || redis.OP().Cli() == nil {
		return
	}
	key := fmt.Sprintf("channel:qr:%s:%s", channelType, sessionID)
	redis.OP().Cli().Set(context.Background(), key, status, time.Duration(ttl)*time.Second)
}

func (m *QRLoginManager) checkRedisStatus(channelType, sessionID string) string {
	if redis.OP() == nil || redis.OP().Cli() == nil {
		return ""
	}
	key := fmt.Sprintf("channel:qr:%s:%s", channelType, sessionID)
	val, err := redis.OP().Cli().Get(context.Background(), key).Result()
	if err != nil {
		return ""
	}
	return val
}
