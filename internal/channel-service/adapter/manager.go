package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/dingtalk"
	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/types"
	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/wechat"
	"github.com/UnicomAI/wanwu/internal/channel-service/client"
	"github.com/UnicomAI/wanwu/internal/channel-service/client/model"
	"github.com/UnicomAI/wanwu/internal/channel-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// nowMilli 返回当前 Unix 毫秒时间戳，用于 ChannelStatus.Checked。
func nowMilli() int64 { return time.Now().UnixMilli() }

// Manager 适配器管理器
// 管理所有通道的平台适配器实例
type Manager struct {
	mu       sync.RWMutex
	adapters map[string]types.Adapter // key: channelID
	cfg      config.Config
	cli      client.IClient
	handler  types.MessageHandler
}

// NewManager 创建适配器管理器
func NewManager(cfg config.Config, cli client.IClient) *Manager {
	return &Manager{
		adapters: make(map[string]types.Adapter),
		cfg:      cfg,
		cli:      cli,
	}
}

// SetMessageHandler 设置全局消息处理函数
func (m *Manager) SetMessageHandler(handler types.MessageHandler) {
	m.handler = handler
}

// StartAll 启动所有已启用且已登录的通道适配器
func (m *Manager) StartAll(ctx context.Context) error {
	channels, err := m.cli.ListEnabledChannels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list enabled channels: %w", err)
	}

	for _, ch := range channels {
		if err := m.StartAdapter(ctx, ch); err != nil {
			log.Errorf("failed to start adapter for channel %s: %v", ch.ChannelID, err)
			continue
		}
		log.Infof("started adapter for channel %s (type=%s)", ch.ChannelID, ch.ChannelType)
	}
	return nil
}

// StartAdapter 启动指定通道的适配器
func (m *Manager) StartAdapter(ctx context.Context, ch *model.Channel) error {
	adapterConfig, err := m.buildAdapterConfig(ch)
	if err != nil {
		return fmt.Errorf("failed to build adapter config: %w", err)
	}

	adapterInst := m.createAdapter(ch.ChannelType)
	if adapterInst == nil {
		return fmt.Errorf("unsupported channel type: %s", ch.ChannelType)
	}

	// 注册消息回调
	if m.handler != nil {
		adapterInst.OnMessage(m.handler)
	}

	if err := adapterInst.Connect(*adapterConfig); err != nil {
		return fmt.Errorf("failed to connect adapter: %w", err)
	}

	// 获取账号信息并更新
	accountId, nickname, avatar, err := adapterInst.GetAccountInfo()
	if err == nil {
		updates := map[string]interface{}{}
		if accountId != "" {
			updates["account_id"] = accountId
		}
		if nickname != "" {
			updates["nickname"] = nickname
		}
		if avatar != "" {
			updates["avatar"] = avatar
		}
		if len(updates) > 0 {
			_, _ = m.cli.UpdateChannel(ctx, ch.ChannelID, updates)
		}
	}

	m.mu.Lock()
	m.adapters[ch.ChannelID] = adapterInst
	m.mu.Unlock()

	return nil
}

// StopAdapter 停止指定通道的适配器
func (m *Manager) StopAdapter(channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	adapterInst, ok := m.adapters[channelID]
	if !ok {
		return nil
	}

	if err := adapterInst.Disconnect(); err != nil {
		return fmt.Errorf("failed to disconnect adapter: %w", err)
	}

	delete(m.adapters, channelID)
	return nil
}

// StopAll 停止所有适配器
func (m *Manager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for channelID, adapterInst := range m.adapters {
		if err := adapterInst.Disconnect(); err != nil {
			log.Errorf("failed to disconnect adapter for channel %s: %v", channelID, err)
		}
		delete(m.adapters, channelID)
	}
}

// GetAdapter 获取指定通道的适配器
func (m *Manager) GetAdapter(channelID string) (types.Adapter, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	adapterInst, ok := m.adapters[channelID]
	return adapterInst, ok
}

// GetChannelStatus 返回通道当前连通状态（轻量：读适配器被动状态，不发起网络请求）。
// 供前端列表轮询 / 推送前快速自检。适配器未启动→offline。
func (m *Manager) GetChannelStatus(channelID string) types.ChannelStatus {
	m.mu.RLock()
	adapterInst, ok := m.adapters[channelID]
	m.mu.RUnlock()
	if !ok {
		return types.ChannelStatus{State: types.ChannelStateOffline, Detail: "adapter not started", Checked: nowMilli()}
	}
	return adapterInst.Status()
}

// ProbeChannel 主动探测通道"还在通"（发起网络请求：钉钉验出站 token、微信读登录态）。
// 供建通道后即时自检 / 低频心跳巡检。适配器未启动→offline；未实现 Prober→降级读 Status。
func (m *Manager) ProbeChannel(ctx context.Context, channelID string) types.ChannelStatus {
	m.mu.RLock()
	adapterInst, ok := m.adapters[channelID]
	m.mu.RUnlock()
	if !ok {
		return types.ChannelStatus{State: types.ChannelStateOffline, Detail: "adapter not started", Checked: nowMilli()}
	}
	if prober, ok := adapterInst.(types.Prober); ok {
		return prober.Probe(ctx)
	}
	return adapterInst.Status()
}

// StartHealthCheck 启动心跳巡检协程：周期性 Probe 所有已启动通道，状态变化时回写 DB。
// 用于发现"连接着但发不出 / token 静默过期"等半死状态并实时反映到 channels.status。
// 返回停止函数，随 ctx 退出或调用停止函数即结束。
func (m *Manager) StartHealthCheck(ctx context.Context, interval time.Duration) func() {
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-stop:
				return
			case <-ticker.C:
				m.healthCheckOnce(ctx)
			}
		}
	}()
	return func() { close(stop) }
}

// healthCheckOnce 执行一轮心跳巡检：对所有已启动通道 Probe，状态非 connected 时回写 DB。
func (m *Manager) healthCheckOnce(ctx context.Context) {
	m.mu.RLock()
	ids := make([]string, 0, len(m.adapters))
	for id := range m.adapters {
		ids = append(ids, id)
	}
	m.mu.RUnlock()

	for _, id := range ids {
		st := m.ProbeChannel(ctx, id)
		// 仅在异常（非 connected）时回写 DB，避免高频无谓写入。
		// connected 认为与建通道时的 loggedIn 一致，不回写。
		if st.State == types.ChannelStateConnected {
			continue
		}
		if _, err := m.cli.UpdateChannel(ctx, id, map[string]interface{}{"status": string(st.State)}); err != nil {
			log.Warnf("[HealthCheck] failed to update channel %s status to %s: %v", id, st.State, err)
		} else {
			log.Warnf("[HealthCheck] channel %s status -> %s (%s)", id, st.State, st.Detail)
		}
	}
}

// SendMessage 通过指定通道发送消息
func (m *Manager) SendMessage(ctx context.Context, channelID, userID, content string, extra map[string]string) error {
	m.mu.RLock()
	adapterInst, ok := m.adapters[channelID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("adapter not found for channel: %s", channelID)
	}
	return adapterInst.SendMessage(ctx, userID, content, extra)
}

// CreateStreamSender 通过指定通道创建流式回复发送器
// 返回 nil 表示该平台不支持流式回复，调用方应降级为 SendMessage
func (m *Manager) CreateStreamSender(ctx context.Context, channelID, userID string, extra map[string]string) types.StreamSender {
	m.mu.RLock()
	adapterInst, ok := m.adapters[channelID]
	m.mu.RUnlock()

	if !ok {
		return nil
	}
	return adapterInst.CreateStreamSender(ctx, userID, extra)
}

// SendFile 通过指定通道向平台用户发送文件附件。
// 平台适配器需实现 types.FileSender 接口；未实现时返回 ErrFileSendUnsupported，
// 调用方据此降级为文本提示。
func (m *Manager) SendFile(ctx context.Context, channelID, userID, fileName, mimeType string, data []byte, extra map[string]string) error {
	m.mu.RLock()
	adapterInst, ok := m.adapters[channelID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("adapter not found for channel: %s", channelID)
	}

	sender, ok := adapterInst.(types.FileSender)
	if !ok {
		return types.ErrFileSendUnsupported
	}
	return sender.SendFile(ctx, userID, fileName, mimeType, data, extra)
}

// SendMarkdown 通过指定通道发送 Markdown 消息。
// 适配器实现 MarkdownSender 时走平台 md 渲染（钉钉 sampleMarkdown 卡片）；
// 未实现时（微信个人号不支持 md 卡片）降级为 SendMessage 纯文本投递，content 原样发送。
// title 用于钉钉卡片标题/通知预览，调用方未提供时应由上层从 content 自动生成。
func (m *Manager) SendMarkdown(ctx context.Context, channelID, userID, title, content string) error {
	m.mu.RLock()
	adapterInst, ok := m.adapters[channelID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("adapter not found for channel: %s", channelID)
	}

	if sender, ok := adapterInst.(types.MarkdownSender); ok {
		return sender.SendMarkdownMessage(ctx, userID, title, content)
	}
	// 平台不支持 md：降级纯文本（content 原样，md 符号不渲染）
	return adapterInst.SendMessage(ctx, userID, content, nil)
}

// RestartAdapter 重启指定通道的适配器
func (m *Manager) RestartAdapter(ctx context.Context, ch *model.Channel) error {
	_ = m.StopAdapter(ch.ChannelID)
	return m.StartAdapter(ctx, ch)
}

// --- 内部方法 ---

// buildAdapterConfig 从通道模型构建适配器配置
func (m *Manager) buildAdapterConfig(ch *model.Channel) (*types.AdapterConfig, error) {
	// API Key 用于后续调用 wanwu OpenAPI，不传给适配器
	_ = ch.ApiKey

	config := &types.AdapterConfig{
		ChannelID:   ch.ChannelID,
		ChannelType: ch.ChannelType,
		Extra:       make(map[string]string),
	}

	// 解析平台配置
	var configMap map[string]string
	if ch.Config != "" {
		_ = json.Unmarshal([]byte(ch.Config), &configMap)
	}

	switch ch.ChannelType {
	case types.ChannelTypeDingTalk:
		config.AppKey = configMap["appKey"]
		config.AppSecret = configMap["appSecret"]
		config.ConnMode = configMap["connectionMode"]
		config.Extra["streamReply"] = configMap["streamReply"]
		config.Extra["cardTemplateId"] = configMap["cardTemplateId"]
	case types.ChannelTypeWeChat:
		config.Token = configMap["token"]
		config.BaseUrl = configMap["baseUrl"]
		config.AccountId = configMap["accountId"]
	case types.ChannelTypeFeiShu:
		config.AppId = configMap["appId"]
		config.AppSecret = configMap["appSecret"]
		config.EncryptKey = configMap["encryptKey"]
		config.VerifyToken = configMap["verificationToken"]
		config.ConnMode = configMap["connectionMode"]
		if config.ConnMode == "" {
			config.ConnMode = "websocket"
		}
	}

	return config, nil
}

// createAdapter 根据通道类型创建适配器
func (m *Manager) createAdapter(channelType string) types.Adapter {
	switch channelType {
	case types.ChannelTypeDingTalk:
		return dingtalk.NewDingTalkAdapter()
	case types.ChannelTypeWeChat:
		return wechat.NewWeChatAdapter()
	default:
		return nil
	}
}
