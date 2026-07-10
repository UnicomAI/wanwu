package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	channel_service "github.com/UnicomAI/wanwu/api/proto/channel-service"
	"github.com/UnicomAI/wanwu/internal/channel-service/adapter"
	"github.com/UnicomAI/wanwu/internal/channel-service/client"
	"github.com/UnicomAI/wanwu/internal/channel-service/client/model"
	"github.com/UnicomAI/wanwu/internal/channel-service/config"
	"github.com/UnicomAI/wanwu/internal/channel-service/qrcode"
	"github.com/UnicomAI/wanwu/internal/channel-service/wanwu"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ChannelService struct {
	channel_service.UnimplementedChannelServiceServer
	cfg      config.Config
	cli      client.IClient
	manager  *adapter.Manager
	qrMgr    *qrcode.QRLoginManager
	wanwuCli *wanwu.Client
}

func NewChannelService(cfg *config.Config, cli client.IClient, mgr *adapter.Manager) *ChannelService {
	return &ChannelService{
		cfg:      *cfg,
		cli:      cli,
		manager:  mgr,
		qrMgr:    qrcode.NewQRLoginManager(*cfg, cli),
		wanwuCli: wanwu.NewClient(cfg.BFF.ApiBaseUrl),
	}
}

// --- 扫码登录 ---

// CreateQRLogin 发起扫码登录
func (s *ChannelService) CreateQRLogin(ctx context.Context, req *channel_service.CreateQRLoginReq) (*channel_service.CreateQRLoginResp, error) {
	result, err := s.qrMgr.CreateQRLogin(ctx, req.ChannelType, req.UserId, req.OrgId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "qr login failed: %v", err)
	}

	return &channel_service.CreateQRLoginResp{
		SessionId:  result.SessionID,
		QrUrl:      result.QrUrl,
		ExpireAt:   result.ExpireAt,
		ExpireTime: result.ExpireTime,
	}, nil
}

// GetQRLoginStatus 查询扫码状态
func (s *ChannelService) GetQRLoginStatus(ctx context.Context, req *channel_service.GetQRLoginStatusReq) (*channel_service.QRLoginStatus, error) {
	statusStr, credentials, err := s.qrMgr.GetQRLoginStatus(ctx, req.ChannelType, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "qr session not found: %v", err)
	}

	resp := &channel_service.QRLoginStatus{
		Status: statusStr,
		Error:  "",
	}

	if statusStr == "success" && credentials != nil {
		resp.Credentials = credentials
	}

	return resp, nil
}

// CancelQRLogin 取消扫码登录
func (s *ChannelService) CancelQRLogin(ctx context.Context, req *channel_service.CancelQRLoginReq) (*emptypb.Empty, error) {
	err := s.qrMgr.CancelQRLogin(ctx, req.ChannelType, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to cancel qr login: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// CompleteQRLogin 完成扫码登录（扫码成功后创建通道）
func (s *ChannelService) CompleteQRLogin(ctx context.Context, req *channel_service.CompleteQRLoginReq) (*channel_service.Channel, error) {
	// 查询会话状态
	statusStr, credentials, err := s.qrMgr.GetQRLoginStatus(ctx, req.ChannelType, req.SessionId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "qr session not found: %v", err)
	}

	if statusStr != "success" {
		return nil, status.Errorf(codes.FailedPrecondition, "qr login not confirmed yet, current status: %s", statusStr)
	}

	if credentials == nil {
		return nil, status.Errorf(codes.Internal, "qr login credentials not found")
	}

	// 使用 session 中存储的 channelType（已在 GetQRLoginStatus 中校验与 req.ChannelType 一致）
	channelType := req.ChannelType
	var name string
	var configMap map[string]string
	var accountID string

	switch channelType {
	case "wechat":
		name = fmt.Sprintf("微信通道 %s", time.Now().Format("2006-01-02"))
		baseUrl := credentials["baseUrl"]
		if baseUrl == "" {
			baseUrl = "https://ilinkai.weixin.qq.com"
		}
		configMap = map[string]string{
			"token":   credentials["token"],
			"baseUrl": baseUrl,
		}
		accountID = credentials["accountId"]

	case "dingtalk":
		name = fmt.Sprintf("钉钉通道 %s", time.Now().Format("2006-01-02"))
		configMap = map[string]string{
			"appKey":    credentials["client_id"],
			"appSecret": credentials["client_secret"],
		}
		accountID = credentials["client_id"]

	case "feishu":
		name = fmt.Sprintf("飞书通道 %s", time.Now().Format("2006-01-02"))
		configMap = map[string]string{
			"appId":     credentials["appId"],
			"appSecret": credentials["appSecret"],
		}
		accountID = credentials["appId"]

	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported channel type: %s", channelType)
	}

	// 微信 token 作为 apiKey 明文存储
	var apiKey string
	if channelType == "wechat" && credentials["token"] != "" {
		apiKey = credentials["token"]
	}

	// 序列化 config
	configJSON, err := json.Marshal(configMap)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid config: %v", err)
	}

	// 确定初始状态
	channelStatus := "loggedIn"
	if channelType == "wechat" {
		if credentials["token"] == "" {
			channelStatus = "offline"
		}
	}

	channel := &model.Channel{
		ChannelID:   model.NewChannelID(),
		Name:        name,
		ChannelType: channelType,
		Status:      channelStatus,
		Enabled:     true,
		AppType:     "agent",
		ApiKey:      apiKey,
		Config:      string(configJSON),
		AccountId:   accountID,
		UserID:      req.UserId,
		OrgID:       req.OrgId,
	}

	created, err := s.cli.CreateChannel(ctx, channel)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create channel: %v", err)
	}

	// 创建成功后启动适配器
	if created.Enabled && created.Status == "loggedIn" {
		go func() {
			if err := s.manager.StartAdapter(context.Background(), created); err != nil {
				log.Errorf("failed to start adapter for channel %s: %v", created.ChannelID, err)
			}
		}()
	}

	return modelToChannelProto(created), nil
}

// --- 通道管理 ---

// CreateChannel 创建通道
func (s *ChannelService) CreateChannel(ctx context.Context, req *channel_service.CreateChannelReq) (*channel_service.Channel, error) {
	// 校验：绑定 API Key 时必须同时提供完整值
	if req.ApiKeyId != "" && req.ApiKey == "" {
		return nil, status.Errorf(codes.InvalidArgument, "api_key is required when api_key_id is provided")
	}

	// 兼容前端传入的 client_id / client_secret 字段名
	normalizeConfig(req.ChannelType, req.Config)

	// 校验：通道配置必填字段
	if err := validateChannelConfig(req.ChannelType, req.Config); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid config: %v", err)
	}

	// 微信通道 baseUrl 为空时设置默认值
	if req.ChannelType == "wechat" {
		if req.Config["baseUrl"] == "" {
			if req.Config == nil {
				req.Config = make(map[string]string)
			}
			req.Config["baseUrl"] = "https://ilinkai.weixin.qq.com"
		}
	}

	// 序列化 config
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid config: %v", err)
	}

	// 确定 appType 默认值
	appType := req.AppType
	if appType == "" {
		appType = "agent"
	}

	// 确定初始状态
	channelStatus := "loggedIn"
	// 微信创建时 token 为空，状态为 offline
	if req.ChannelType == "wechat" {
		if req.Config["token"] == "" {
			channelStatus = "offline"
		}
	}

	// 确定 accountId
	accountId := ""
	switch req.ChannelType {
	case "dingtalk":
		accountId = req.Config["appKey"]
	case "wechat":
		accountId = req.Config["accountId"]
	case "feishu":
		accountId = req.Config["appId"]
	}

	channel := &model.Channel{
		ChannelID:   model.NewChannelID(),
		Name:        req.Name,
		ChannelType: req.ChannelType,
		Status:      channelStatus,
		Enabled:     true,
		AppType:     appType,
		AppID:       req.AppId,
		AppName:     req.AppName,
		ApiKeyID:    req.ApiKeyId,
		ApiKeyName:  "", // 后续通过代理接口同步
		ApiKey:      req.ApiKey,
		ModelUuid:   req.ModelUuid,
		AgentId:     req.AgentId,
		Config:      string(configJSON),
		AccountId:   accountId,
		UserID:      req.UserId,
		OrgID:       req.OrgId,
	}

	created, err := s.cli.CreateChannel(ctx, channel)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create channel: %v", err)
	}

	// 创建成功后启动适配器
	if created.Enabled && created.Status == "loggedIn" {
		go func() {
			if err := s.manager.StartAdapter(context.Background(), created); err != nil {
				log.Errorf("failed to start adapter for channel %s: %v", created.ChannelID, err)
			}
		}()
	}

	return modelToChannelProto(created), nil
}

// ListChannels 获取通道列表
func (s *ChannelService) ListChannels(ctx context.Context, req *channel_service.ListChannelsReq) (*channel_service.ListChannelsResp, error) {
	pageNo := req.PageNo
	if pageNo <= 0 {
		pageNo = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	channels, total, err := s.cli.ListChannels(ctx, req.UserId, req.OrgId, req.Name, pageNo, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list channels: %v", err)
	}

	resp := &channel_service.ListChannelsResp{
		List:  make([]*channel_service.Channel, 0, len(channels)),
		Total: int32(total),
	}
	for _, ch := range channels {
		resp.List = append(resp.List, modelToChannelProto(ch))
	}
	return resp, nil
}

// GetChannel 获取通道详情
func (s *ChannelService) GetChannel(ctx context.Context, req *channel_service.GetChannelReq) (*channel_service.Channel, error) {
	ch, err := s.cli.GetChannel(ctx, req.ChannelId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "channel not found: %v", err)
	}
	return modelToChannelProto(ch), nil
}

// UpdateChannel 更新通道
func (s *ChannelService) UpdateChannel(ctx context.Context, req *channel_service.UpdateChannelReq) (*channel_service.Channel, error) {
	// 校验：更新 API Key 绑定时必须同时提供完整值
	if req.ApiKeyId != "" && req.ApiKey == "" {
		return nil, status.Errorf(codes.InvalidArgument, "api_key is required when api_key_id is provided")
	}

	// 校验：通道配置必填字段（仅在更新 config 时校验）
	if len(req.Config) > 0 {
		// 获取当前通道信息以确定 channelType
		existing, err := s.cli.GetChannel(ctx, req.ChannelId)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "channel not found: %v", err)
		}
		// 兼容前端传入的 client_id / client_secret 字段名
		normalizeConfig(existing.ChannelType, req.Config)
		if err := validateChannelConfig(existing.ChannelType, req.Config); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid config: %v", err)
		}
	}

	updates := make(map[string]interface{})

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.AppId != nil {
		// optional app_id：传了即写入（含空串，用于 WGA 选「无」子智能体时清空 app_id）；未传则保留旧值
		updates["app_id"] = req.GetAppId()
	}
	if req.AppName != nil {
		// optional app_name：传了即写入（含空串，用于清空）；未传则保留旧值
		updates["app_name"] = req.GetAppName()
	}
	if req.ApiKeyId != "" {
		updates["api_key_id"] = req.ApiKeyId
	}
	if req.ApiKey != "" {
		updates["api_key"] = req.ApiKey
	}
	if req.ModelUuid != "" {
		updates["model_uuid"] = req.ModelUuid
	}
	if req.AgentId != nil {
		// optional agent_id：传了即写入（含空串，用于 WGA 清空 agentId 切回默认 Supervisor）；未传则保留旧值
		updates["agent_id"] = req.GetAgentId()
	}
	if len(req.Config) > 0 {
		configJSON, err := json.Marshal(req.Config)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid config: %v", err)
		}
		updates["config"] = string(configJSON)
		// 更新 accountId
		if appKey, ok := req.Config["appKey"]; ok {
			updates["account_id"] = appKey
		}
	}

	if len(updates) == 0 {
		return s.GetChannel(ctx, &channel_service.GetChannelReq{ChannelId: req.ChannelId})
	}

	ch, err := s.cli.UpdateChannel(ctx, req.ChannelId, updates)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update channel: %v", err)
	}

	// 更新后按需重启适配器（仅当连接相关字段变更时才重启）
	needRestart := len(req.Config) > 0 || req.ApiKey != ""

	if needRestart {
		go func() {
			if err := s.manager.RestartAdapter(context.Background(), ch); err != nil {
				log.Errorf("failed to restart adapter for channel %s: %v", ch.ChannelID, err)
			}
		}()
	}

	return modelToChannelProto(ch), nil
}

// UpdateChannelStatus 启用/停用通道
func (s *ChannelService) UpdateChannelStatus(ctx context.Context, req *channel_service.UpdateChannelStatusReq) (*channel_service.Channel, error) {
	ch, err := s.cli.UpdateChannel(ctx, req.ChannelId, map[string]interface{}{
		"enabled": req.Enabled,
		"status":  s.statusForEnabled(req.Enabled),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update channel status: %v", err)
	}

	// 启用时重连适配器，停用时断开
	go func() {
		if req.Enabled {
			if err := s.manager.StartAdapter(context.Background(), ch); err != nil {
				log.Errorf("failed to start adapter for channel %s: %v", ch.ChannelID, err)
			}
		} else {
			if err := s.manager.StopAdapter(ch.ChannelID); err != nil {
				log.Errorf("failed to stop adapter for channel %s: %v", ch.ChannelID, err)
			}
		}
	}()

	return modelToChannelProto(ch), nil
}

// DeleteChannel 删除通道
func (s *ChannelService) DeleteChannel(ctx context.Context, req *channel_service.DeleteChannelReq) (*emptypb.Empty, error) {
	// 先停止适配器
	if err := s.manager.StopAdapter(req.ChannelId); err != nil {
		log.Errorf("failed to stop adapter for channel %s: %v", req.ChannelId, err)
	}

	if err := s.cli.DeleteChannel(ctx, req.ChannelId); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete channel: %v", err)
	}

	// 级联清理该通道下的会话映射（threadId/conversationId），仅删通道时清理
	if err := s.cli.DeleteConversationsByChannel(ctx, req.ChannelId); err != nil {
		log.Errorf("failed to cleanup conversations for channel %s: %v", req.ChannelId, err)
	}
	return &emptypb.Empty{}, nil
}

// DisconnectChannel 断开通道连接
func (s *ChannelService) DisconnectChannel(ctx context.Context, req *channel_service.DisconnectChannelReq) (*emptypb.Empty, error) {
	// 更新状态为 offline
	_, err := s.cli.UpdateChannel(ctx, req.ChannelId, map[string]interface{}{
		"status": "offline",
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to disconnect channel: %v", err)
	}

	// 断开适配器
	if err := s.manager.StopAdapter(req.ChannelId); err != nil {
		log.Errorf("failed to stop adapter for channel %s: %v", req.ChannelId, err)
	}

	return &emptypb.Empty{}, nil
}

// --- 内部方法 ---

func (s *ChannelService) statusForEnabled(enabled bool) string {
	if enabled {
		return "loggedIn"
	}
	return "offline"
}

// modelToChannelProto 将数据库模型转换为 protobuf 响应
func modelToChannelProto(ch *model.Channel) *channel_service.Channel {
	// 解析 config
	configMap := make(map[string]string)
	if ch.Config != "" {
		_ = json.Unmarshal([]byte(ch.Config), &configMap)
	}

	hasApiKey := ch.ApiKey != ""

	return &channel_service.Channel{
		Id:          ch.ChannelID,
		Name:        ch.Name,
		ChannelType: ch.ChannelType,
		Status:      ch.Status,
		AccountId:   ch.AccountId,
		Nickname:    ch.Nickname,
		Avatar:      ch.Avatar,
		Enabled:     ch.Enabled,
		AppType:     ch.AppType,
		AppId:       ch.AppID,
		AppName:     ch.AppName,
		ApiKeyId:    ch.ApiKeyID,
		ApiKeyName:  ch.ApiKeyName,
		HasApiKey:   hasApiKey,
		ModelUuid:   ch.ModelUuid,
		AgentId:     ch.AgentId,
		Config:      configMap,
		CreatedAt:   util.Time2Str(ch.CreatedAt),
		UpdatedAt:   util.Time2Str(ch.UpdatedAt),
		UserId:      ch.UserID,
		OrgId:       ch.OrgID,
	}
}

// normalizeConfig 兼容前端传入的 client_id / client_secret 字段名，
// 统一映射为后端期望的 appKey / appSecret。
func normalizeConfig(channelType string, config map[string]string) {
	switch channelType {
	case "dingtalk":
		if v, ok := config["client_id"]; ok && config["appKey"] == "" {
			config["appKey"] = v
		}
		if v, ok := config["client_secret"]; ok && config["appSecret"] == "" {
			config["appSecret"] = v
		}
	case "feishu":
		if v, ok := config["client_id"]; ok && config["appId"] == "" {
			config["appId"] = v
		}
		if v, ok := config["client_secret"]; ok && config["appSecret"] == "" {
			config["appSecret"] = v
		}
	}
}

// validateChannelConfig 校验通道配置必填字段
func validateChannelConfig(channelType string, config map[string]string) error {
	if config == nil {
		config = make(map[string]string)
	}
	switch channelType {
	case "dingtalk":
		if config["appKey"] == "" {
			return fmt.Errorf("dingtalk channel requires appKey in config")
		}
		if config["appSecret"] == "" {
			return fmt.Errorf("dingtalk channel requires appSecret in config")
		}
	case "wechat":
		if config["token"] == "" {
			return fmt.Errorf("wechat channel requires token in config")
		}
	case "feishu":
		if config["appId"] == "" {
			return fmt.Errorf("feishu channel requires appId in config")
		}
		if config["appSecret"] == "" {
			return fmt.Errorf("feishu channel requires appSecret in config")
		}
	}
	return nil
}
