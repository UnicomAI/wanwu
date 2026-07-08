package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/dingtalk"
	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/types"
	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/wechat"
	"github.com/UnicomAI/wanwu/internal/channel-service/client"
	"github.com/UnicomAI/wanwu/internal/channel-service/client/model"
	"github.com/UnicomAI/wanwu/internal/channel-service/config"
	"github.com/UnicomAI/wanwu/pkg/log"
)

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
