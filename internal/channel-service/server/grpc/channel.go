package grpc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	channel_service "github.com/UnicomAI/wanwu/api/proto/channel-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/channel-service/adapter"
	"github.com/UnicomAI/wanwu/internal/channel-service/adapter/types"
	"github.com/UnicomAI/wanwu/internal/channel-service/client"
	"github.com/UnicomAI/wanwu/internal/channel-service/client/model"
	"github.com/UnicomAI/wanwu/internal/channel-service/config"
	"github.com/UnicomAI/wanwu/internal/channel-service/qrcode"
	"github.com/UnicomAI/wanwu/internal/channel-service/wanwu"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ChannelService struct {
	channel_service.UnimplementedChannelServiceServer
	cfg         config.Config
	cli         client.IClient
	manager     *adapter.Manager
	qrMgr       *qrcode.QRLoginManager
	wanwuCli    *wanwu.Client
	convManager *wanwu.ConversationManager // 预置/查询 channel_conversations 映射（测试消息发送后落表）
}

func NewChannelService(cfg *config.Config, cli client.IClient, mgr *adapter.Manager) *ChannelService {
	return &ChannelService{
		cfg:         *cfg,
		cli:         cli,
		manager:     mgr,
		qrMgr:       qrcode.NewQRLoginManager(*cfg, cli),
		wanwuCli:    wanwu.NewClient(cfg.BFF.ApiBaseUrl),
		convManager: wanwu.NewConversationManager(cli),
	}
}

// --- 扫码登录 ---

// CreateQRLogin 发起扫码登录
func (s *ChannelService) CreateQRLogin(ctx context.Context, req *channel_service.CreateQRLoginReq) (*channel_service.CreateQRLoginResp, error) {
	result, err := s.qrMgr.CreateQRLogin(ctx, req.ChannelType, req.UserId, req.OrgId)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("qr login failed: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelNotFound, fmt.Sprintf("qr session not found: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("failed to cancel qr login: %v", err))
	}
	return &emptypb.Empty{}, nil
}

// CompleteQRLogin 完成扫码登录（扫码成功后创建通道）
func (s *ChannelService) CompleteQRLogin(ctx context.Context, req *channel_service.CompleteQRLoginReq) (*channel_service.Channel, error) {
	// 查询会话状态
	statusStr, credentials, err := s.qrMgr.GetQRLoginStatus(ctx, req.ChannelType, req.SessionId)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelNotFound, fmt.Sprintf("qr session not found: %v", err))
	}

	if statusStr != "success" {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("qr login not confirmed yet, current status: %s", statusStr))
	}

	if credentials == nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, "qr login credentials not found")
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, fmt.Sprintf("unsupported channel type: %s", channelType))
	}

	// 微信 token 作为 apiKey 明文存储
	var apiKey string
	if channelType == "wechat" && credentials["token"] != "" {
		apiKey = credentials["token"]
	}

	// 序列化 config
	configJSON, err := json.Marshal(configMap)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, fmt.Sprintf("invalid config: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("failed to create channel: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, "api_key is required when api_key_id is provided")
	}

	// 兼容前端传入的 client_id / client_secret 字段名
	normalizeConfig(req.ChannelType, req.Config)

	// 校验：通道配置必填字段
	if err := validateChannelConfig(req.ChannelType, req.Config); err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, fmt.Sprintf("invalid config: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, fmt.Sprintf("invalid config: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("failed to create channel: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("failed to list channels: %v", err))
	}

	resp := &channel_service.ListChannelsResp{
		List:  make([]*channel_service.Channel, 0, len(channels)),
		Total: int32(total),
	}
	for _, ch := range channels {
		resp.List = append(resp.List, s.channelToProtoWithConnectivity(ch))
	}
	return resp, nil
}

// GetChannel 获取通道详情
func (s *ChannelService) GetChannel(ctx context.Context, req *channel_service.GetChannelReq) (*channel_service.Channel, error) {
	ch, err := s.cli.GetChannel(ctx, req.ChannelId)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelNotFound, fmt.Sprintf("channel not found: %v", err))
	}
	return s.channelToProtoWithConnectivity(ch), nil
}

// UpdateChannel 更新通道
func (s *ChannelService) UpdateChannel(ctx context.Context, req *channel_service.UpdateChannelReq) (*channel_service.Channel, error) {
	// 校验：更新 API Key 绑定时必须同时提供完整值
	if req.ApiKeyId != "" && req.ApiKey == "" {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, "api_key is required when api_key_id is provided")
	}

	// 校验：通道配置必填字段（仅在更新 config 时校验）
	if len(req.Config) > 0 {
		// 获取当前通道信息以确定 channelType
		existing, err := s.cli.GetChannel(ctx, req.ChannelId)
		if err != nil {
			return nil, grpc_util.ErrorStatus(err_code.Code_ChannelNotFound, fmt.Sprintf("channel not found: %v", err))
		}
		// 兼容前端传入的 client_id / client_secret 字段名
		normalizeConfig(existing.ChannelType, req.Config)
		if err := validateChannelConfig(existing.ChannelType, req.Config); err != nil {
			return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, fmt.Sprintf("invalid config: %v", err))
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
			return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, fmt.Sprintf("invalid config: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("failed to update channel: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("failed to update channel status: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("failed to delete channel: %v", err))
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
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("failed to disconnect channel: %v", err))
	}

	// 断开适配器
	if err := s.manager.StopAdapter(req.ChannelId); err != nil {
		log.Errorf("failed to stop adapter for channel %s: %v", req.ChannelId, err)
	}

	return &emptypb.Empty{}, nil
}

// --- 通道连通性 / 测试消息 ---

// SendTestMessage 发送测试消息：给指定通道的收件人发一条消息，验证通道出站可用。
// 用途：①建通道后即时自检连通性 ②"建通道后给通道发消息"的首次投递。
// 成功发送后预置 channel_conversations 映射行（conversation_id 留空，首次真实对话才填），
// 使该收件人 user_id 进入会话表，便于后续推送 API 直接查表取收件人。
func (s *ChannelService) SendTestMessage(ctx context.Context, req *channel_service.SendTestMessageReq) (*channel_service.SendTestMessageResp, error) {
	if req.ChannelId == "" {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, "channel_id is required")
	}

	// 权限校验：操作者须为通道拥有者（同 CreateChannel/UpdateChannel 的归属校验口径）
	ch, err := s.cli.GetChannel(ctx, req.ChannelId)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelNotFound, fmt.Sprintf("channel not found: %v", err))
	}
	if req.OperatorId != "" && ch.UserID != "" && req.OperatorId != ch.UserID {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, "operator is not the channel owner")
	}

	// 适配器须已启动（通道启用且已连接）
	if _, ok := s.manager.GetAdapter(req.ChannelId); !ok {
		return &channel_service.SendTestMessageResp{
			Connectivity: s.toConnectivityProto(types.ChannelStatus{State: types.ChannelStateOffline, Detail: "adapter not started, please enable the channel", Checked: nowMilli()}),
			MessageSent:  false,
			Error:        "channel adapter not started",
		}, nil
	}

	// 确定收件人 userID：调用方提供则用之；未提供则自动从 channel_conversations
	// 查该通道最近互动过的 IM 用户（路径1：收件人来自历史互动入站落表）。
	// 注意：机器人不能给自己发（微信 @im.bot / 钉钉 appKey 均不可作收件人），
	// 故不使用 account_id 自测，必须取真实 IM 用户 ID。
	userID := req.UserId
	if userID == "" {
		convs, err := s.cli.ListConversationsByChannel(ctx, req.ChannelId, 1)
		if err != nil {
			return s.selfProbeOnly(ctx, req.ChannelId, "failed to query recipients: "+err.Error())
		}
		if len(convs) == 0 || convs[0].UserID == "" {
			// 无历史互动用户：仅做连通性自检，引导用户先给 bot 发一条消息（触发入站落表）
			return s.selfProbeOnly(ctx, req.ChannelId, "no recipient: send any message to the bot first, or provide userId explicitly")
		}
		userID = convs[0].UserID
	}

	content := req.Content
	if content == "" {
		content = "通道连通测试 ✅"
	}

	// 出站投递：单聊，extra 留空（推送场景不依赖 sessionWebhook/群聊 cid）
	if err := s.manager.SendMessage(ctx, req.ChannelId, userID, content, nil); err != nil {
		st := s.manager.ProbeChannel(ctx, req.ChannelId)
		return &channel_service.SendTestMessageResp{
			Connectivity: s.toConnectivityProto(st),
			MessageSent:  false,
			Error:        err.Error(),
		}, nil
	}

	// 发送成功：预置会话映射（conversation_id 留空，首次真实对话时由 chat handler 填充）。
	// 用通道 AppType 作为 app_type key（与 chat.go dispatch 一致：agent/wga/dip）。
	appTypeKey := ch.AppType
	if appTypeKey == "" {
		appTypeKey = "agent"
	}
	s.convManager.SetConversationID(ctx, req.ChannelId, userID, appTypeKey, "")

	st := s.manager.GetChannelStatus(req.ChannelId)
	// 发送成功即视为连通，即便被动 Status 还未刷新为 connected 也置 ok
	if st.State != types.ChannelStateConnected {
		st = types.ChannelStatus{State: types.ChannelStateConnected, Detail: "test message sent", Checked: nowMilli()}
	}
	return &channel_service.SendTestMessageResp{
		Connectivity: s.toConnectivityProto(st),
		MessageSent:  true,
	}, nil
}

// selfProbeOnly 不投递消息，仅做连通性主动探测后返回。
// 用于钉钉/飞书等无法自发自收的平台在未提供 userId 时的自检：能确认通道"连得通、凭据有效"，
// 但不验证真实投递。detail 携带引导用户进一步操作的信息。
func (s *ChannelService) selfProbeOnly(ctx context.Context, channelID, hint string) (*channel_service.SendTestMessageResp, error) {
	st := s.manager.ProbeChannel(ctx, channelID)
	return &channel_service.SendTestMessageResp{
		Connectivity: s.toConnectivityProto(st),
		MessageSent:  false,
		Error:        hint,
	}, nil
}

// SendMessage 供内部服务（经 bff callback）给指定通道发消息。
// 与 SendTestMessage 的区别：无 owner 权限校验、不预置会话映射（不调 SetConversationID）、
// 不返回连通性，仅做出站投递。会话行由入站流程（用户给 bot 发消息时 chat handler 落表）维护。
// msg_type 支持 text/markdown/file：
//   - text/默认：纯文本（content 必填）
//   - markdown：钉钉 md 卡片，微信降级纯文本（content 必填，title 留空自动生成）
//   - file：文件附件（file_url+file_name 必填，content 可空作附带文案）。file_url 为万悟 minio 下载地址，
//     channel-service 下载字节后经适配器 SendFile 投递。钉钉/微信支持，飞书未实现返回错误（不降级）。
//     content 非空时先发文案再发文件（两条独立投递）：文案失败文件不发；文案已发而文件失败时返回错误
//     （不回滚文案，错误信息可据此判断是否重试，重试会重复发文案）。
//
// 错误用 err_code 表达：Code_ChannelInvalidArg 入参缺失 / Code_ChannelNotConnected 适配器未启动或无收件人 /
// Code_ChannelRateLimited IM 平台频控（命中 types.ErrIMRateLimited，调用方可退避重试）/ Code_ChannelGeneral 其它失败。
func (s *ChannelService) SendMessage(ctx context.Context, req *channel_service.SendMessageReq) (*channel_service.SendMessageResp, error) {
	if req.ChannelId == "" {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, "channel_id is required")
	}
	// file 场景 content 可空（仅作附带文案），其余场景 content 必填
	if req.MsgType != "file" && req.Content == "" {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, "content is required for text/markdown message")
	}
	if req.MsgType == "file" && (req.FileUrl == "" || req.FileName == "") {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelInvalidArg, "file_url and file_name are required for file message")
	}

	// 适配器须已启动（通道启用且已连接）
	if _, ok := s.manager.GetAdapter(req.ChannelId); !ok {
		return nil, grpc_util.ErrorStatus(err_code.Code_ChannelNotConnected, "channel adapter not started, please enable the channel")
	}

	// 确定收件人：调用方提供则用之；未提供则自动从 channel_conversations 取该通道最近互动过的 IM 用户。
	userID := req.UserId
	if userID == "" {
		convs, err := s.cli.ListConversationsByChannel(ctx, req.ChannelId, 1)
		if err != nil {
			return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("failed to query recipients: %v", err))
		}
		if len(convs) == 0 || convs[0].UserID == "" {
			// 无历史互动用户：引导用户先给 bot 发一条消息（触发入站落表）
			return nil, grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, "no recipient: send any message to the bot first, or provide userId explicitly")
		}
		userID = convs[0].UserID
	}

	// 出站投递：按 msg_type 分流。
	var sendErr error
	switch req.MsgType {
	case "markdown":
		title := req.Title
		if title == "" {
			title = deriveMarkdownTitle(req.Content)
		}
		sendErr = s.manager.SendMarkdown(ctx, req.ChannelId, userID, title, req.Content)
	case "file":
		// content 非空时先发文案再发文件（与文档「附带文案」语义对齐）。
		// 两条独立投递：文案发送失败直接返回（文件不发）；文案已发而文件失败时如实返回错误，
		// 错误信息标注文案已发，便于调用方判断是否需要重试（重试会重复发文案，调用方酌情处理）。
		if req.Content != "" {
			if err := s.manager.SendMessage(ctx, req.ChannelId, userID, req.Content, nil); err != nil {
				return nil, s.wrapSendErr(err)
			}
		}
		sendErr = s.sendFileMessage(ctx, req.ChannelId, userID, req.FileUrl, req.FileName, req.FileMimeType)
	default: // text
		sendErr = s.manager.SendMessage(ctx, req.ChannelId, userID, req.Content, nil)
	}

	if sendErr != nil {
		return nil, s.wrapSendErr(sendErr)
	}

	return &channel_service.SendMessageResp{Ok: true, UserId: userID}, nil
}

// wrapSendErr 将适配器发送错误映射为对应的 gRPC err_code：
//   - ErrFileSendUnsupported（飞书等不支持发文件）：Code_ChannelGeneral，不降级发文本
//   - ErrIMRateLimited（微信 ret=-2 / 钉钉频控）：Code_ChannelRateLimited，调用方可退避重试
//   - 其它：Code_ChannelGeneral
func (s *ChannelService) wrapSendErr(err error) error {
	// 飞书等不支持发文件的平台：返回明确错误（不降级发文本，避免调用方误以为送达）
	if errors.Is(err, types.ErrFileSendUnsupported) {
		return grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("channel does not support file send: %v", err))
	}
	// 命中 IM 平台频控（微信 ret=-2 / 钉钉频控 errcode）单独返回 Code_ChannelRateLimited，
	// 调用方可据此退避重试（区别于 Code_ChannelGeneral 的永久性错误）。
	if errors.Is(err, types.ErrIMRateLimited) {
		return grpc_util.ErrorStatus(err_code.Code_ChannelRateLimited, fmt.Sprintf("im platform rate limited: %v", err))
	}
	return grpc_util.ErrorStatus(err_code.Code_ChannelGeneral, fmt.Sprintf("send message failed: %v", err))
}

// sendFileMessage 下载 file_url（万悟 minio 文件地址）字节后经适配器 SendFile 投递。
// fileMimeType 留空时按 file_name 扩展名推断（适配器内部亦按扩展名识别类型，故留空通常无碍）。
// extra 留空：内部单聊场景钉钉走 oToMessages、微信走 ilink，均不依赖 sessionWebhook/群聊 cid。
func (s *ChannelService) sendFileMessage(ctx context.Context, channelID, userID, fileURL, fileName, fileMimeType string) error {
	data, err := downloadFileBytes(ctx, fileURL)
	if err != nil {
		return fmt.Errorf("download file %s failed: %w", fileName, err)
	}
	mimeType := fileMimeType
	if mimeType == "" {
		mimeType = guessMimeTypeByExt(fileName)
	}
	return s.manager.SendFile(ctx, channelID, userID, fileName, mimeType, data, nil)
}

// downloadFileBytes HTTP GET 下载指定 URL 的文件字节（用于按 minio 下载地址取文件投递 IM）。
// 复用独立 http.Client 不设超时（文件可能较大），与 WGAWorkspaceDownload 的下载口径一致。
func downloadFileBytes(ctx context.Context, fileURL string) ([]byte, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	dlClient := &http.Client{}
	resp, err := dlClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, truncateBody(body))
	}
	return io.ReadAll(resp.Body)
}

// guessMimeTypeByExt 按文件名扩展名推断 MIME 类型，推断失败返回空串（交由适配器按扩展名处理）。
func guessMimeTypeByExt(fileName string) string {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return ""
	}
	if mt := mime.TypeByExtension(ext); mt != "" {
		return mt
	}
	return ""
}

// truncateBody 截断错误响应体，避免日志/错误信息过长。
func truncateBody(b []byte) string {
	const max = 200
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max])
}


// deriveMarkdownTitle 从 markdown 内容生成卡片标题：取第一行非空文本，
// 去掉行首 # 标记符号与前后空白，截断到 20 字（钉钉 sampleMarkdown 的 title 字段必填，
// 用于通知栏/会话列表预览）。
func deriveMarkdownTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 去掉行首的 # 标题标记（#、##、### ...）
		line = strings.TrimLeft(line, "#")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// 截断到 20 字符（rune 计数，避免中文截半）
		r := []rune(line)
		if len(r) > 20 {
			line = string(r[:20])
		}
		return line
	}
	return "消息通知"
}

// GetChannelConnectivity 查询通道实时连通状态（从适配器内存读，非 DB 快照）。
func (s *ChannelService) GetChannelConnectivity(ctx context.Context, req *channel_service.GetChannelConnectivityReq) (*channel_service.Connectivity, error) {
	st := s.manager.GetChannelStatus(req.ChannelId)
	return s.toConnectivityProto(st), nil
}

// toConnectivityProto 将 types.ChannelStatus 转换为 protobuf Connectivity。
func (s *ChannelService) toConnectivityProto(st types.ChannelStatus) *channel_service.Connectivity {
	return &channel_service.Connectivity{
		State:   string(st.State),
		Detail:  st.Detail,
		Ok:      st.State == types.ChannelStateConnected,
		Checked: st.Checked,
	}
}

// nowMilli 返回当前 Unix 毫秒时间戳。
func nowMilli() int64 { return time.Now().UnixMilli() }

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
	// 出口展示归一化：存储用后端规范名，对外展示用前端展示名，避免同一凭据两套 key 并存导致页面混乱。
	displayConfig(ch.ChannelType, configMap)

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

// channelToProtoWithConnectivity 在 modelToChannelProto 基础上补实时连通状态。
// connectivity 从适配器内存读（适配器未启动→offline），非 DB 快照，供前端展示真实状态。
func (s *ChannelService) channelToProtoWithConnectivity(ch *model.Channel) *channel_service.Channel {
	p := modelToChannelProto(ch)
	p.Connectivity = s.toConnectivityProto(s.manager.GetChannelStatus(ch.ChannelID))
	return p
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

// displayConfig 出口展示归一化（与 normalizeConfig 入口归一化互逆）：
// 把存储用的后端规范名转成前端展示名，并删除冗余 key，使页面只看到一套字段，不混乱。
// 仅做内存转换，不落库；存储仍保留后端规范名。
//
// 钉钉存储为 appKey/appSecret（可能并存冗余的 client_id/client_secret），
// 展示只输出 client_id/client_secret（钉钉开放平台标准命名，前端展示用）。
// 微信/飞书保持原样不动。
func displayConfig(channelType string, config map[string]string) {
	switch channelType {
	case "dingtalk":
		appKey := config["appKey"]
		if appKey == "" {
			appKey = config["client_id"]
		}
		appSecret := config["appSecret"]
		if appSecret == "" {
			appSecret = config["client_secret"]
		}
		// 清空后只保留展示名
		for k := range config {
			delete(config, k)
		}
		if appKey != "" {
			config["client_id"] = appKey
		}
		if appSecret != "" {
			config["client_secret"] = appSecret
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
