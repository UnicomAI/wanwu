package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
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

// loadWGAChannelForUser 加载通道并校验：存在、已绑定 apiKey、属于该用户。
// 用于 WGA 工作区/上传接口的鉴权与越权防护。
func (s *ChannelService) loadWGAChannelForUser(ctx context.Context, channelID, userID string) (*model.Channel, error) {
	ch, err := s.cli.GetChannel(ctx, channelID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "channel not found: %v", err)
	}
	if !ch.HasApiKey() {
		return nil, status.Errorf(codes.FailedPrecondition, "channel %s has no api key bound", channelID)
	}
	// 越权防护：仅通道属主可操作其会话
	if userID != "" && ch.UserID != "" && ch.UserID != userID {
		return nil, status.Errorf(codes.PermissionDenied, "channel %s does not belong to user %s", channelID, userID)
	}
	return ch, nil
}

// verifyWGAThreadOwner 校验 threadId 属于该 channel+user 的 WGA 会话，防止越权下载他人工作区。
func (s *ChannelService) verifyWGAThreadOwner(ctx context.Context, channelID, userID, threadID string) error {
	conv, err := s.cli.GetConversation(ctx, channelID, userID, "wga")
	if err != nil || conv == nil || conv.ConversationID != threadID {
		return status.Errorf(codes.PermissionDenied, "thread %s does not belong to channel %s user %s", threadID, channelID, userID)
	}
	return nil
}

// --- 通用智能体（WGA）工作区与文件 ---

// GetWGAWorkspace 获取 WGA 工作区目录树
func (s *ChannelService) GetWGAWorkspace(ctx context.Context, req *channel_service.GetWGAWorkspaceReq) (*channel_service.GetWGAWorkspaceResp, error) {
	ch, err := s.loadWGAChannelForUser(ctx, req.ChannelId, req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.verifyWGAThreadOwner(ctx, req.ChannelId, req.UserId, req.ThreadId); err != nil {
		return nil, err
	}
	ws, err := s.wanwuCli.WGAWorkspace(ctx, ch.ApiKey, req.ThreadId, req.RunId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get wga workspace: %v", err)
	}
	return wgaWorkspaceToProto(ws), nil
}

// DownloadWGAWorkspace 下载 WGA 工作区文件（path 为空则下载整个工作区 ZIP）
func (s *ChannelService) DownloadWGAWorkspace(ctx context.Context, req *channel_service.DownloadWGAWorkspaceReq) (*channel_service.DownloadWGAWorkspaceResp, error) {
	ch, err := s.loadWGAChannelForUser(ctx, req.ChannelId, req.UserId)
	if err != nil {
		return nil, err
	}
	if err := s.verifyWGAThreadOwner(ctx, req.ChannelId, req.UserId, req.ThreadId); err != nil {
		return nil, err
	}
	resp, err := s.wanwuCli.WGAWorkspaceDownload(ctx, ch.ApiKey, req.ThreadId, req.RunId, req.Path)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to download wga workspace: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to read wga workspace download: %v", err)
	}
	// 从 Content-Disposition 提取文件名，回退用 path 或默认名
	fileName := extractFileName(resp.Header.Get("Content-Disposition"), req.Path)
	return &channel_service.DownloadWGAWorkspaceResp{FileName: fileName, Data: data}, nil
}

// UploadWGAFile 上传文件到万悟 minio，返回 filePath（供 WGA 多模态对话 binary.url 使用）
func (s *ChannelService) UploadWGAFile(ctx context.Context, req *channel_service.UploadWGAFileReq) (*channel_service.UploadWGAFileResp, error) {
	ch, err := s.loadWGAChannelForUser(ctx, req.ChannelId, req.UserId)
	if err != nil {
		return nil, err
	}
	if len(req.Data) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "file data is empty")
	}
	uf, err := s.wanwuCli.UploadFile(ctx, ch.ApiKey, req.FileName, req.MimeType, req.Data)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upload wga file: %v", err)
	}
	return &channel_service.UploadWGAFileResp{
		FileName: uf.FileName,
		FileId:   uf.FileId,
		FilePath: uf.FilePath,
		FileSize: uf.FileSize,
	}, nil
}

// extractFileName 从 Content-Disposition 头提取文件名，提取失败时按 path 推断或返回默认名。
func extractFileName(contentDisposition, path string) string {
	if contentDisposition != "" {
		if idx := strings.Index(contentDisposition, "filename="); idx >= 0 {
			name := strings.Trim(contentDisposition[idx+len("filename="):], `" `)
			if name != "" {
				return name
			}
		}
	}
	if path != "" {
		if idx := strings.LastIndex(path, "/"); idx >= 0 {
			return path[idx+1:]
		}
		return path
	}
	return "workspace.zip"
}

// wgaWorkspaceToProto 将 wanwu.WGAWorkspace 转换为 protobuf 响应
func wgaWorkspaceToProto(ws *wanwu.WGAWorkspace) *channel_service.GetWGAWorkspaceResp {
	if ws == nil {
		return &channel_service.GetWGAWorkspaceResp{}
	}
	return &channel_service.GetWGAWorkspaceResp{
		ThreadId:  ws.ThreadID,
		RunId:     ws.RunID,
		FileCount: ws.FileCount,
		TotalSize: ws.TotalSize,
		IsDisplay: ws.IsDisplay,
		Path:      ws.Path,
		Files:     wgaFileNodesToProto(ws.Files),
	}
}

func wgaFileNodesToProto(nodes []*wanwu.WGAFileNode) []*channel_service.WGAFileNode {
	if len(nodes) == 0 {
		return nil
	}
	out := make([]*channel_service.WGAFileNode, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, &channel_service.WGAFileNode{
			Name:     n.Name,
			Type:     n.Type,
			Size:     n.Size,
			MimeType: n.MimeType,
			Children: wgaFileNodesToProto(n.Children),
		})
	}
	return out
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
