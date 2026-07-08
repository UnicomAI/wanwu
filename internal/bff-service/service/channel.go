package service

import (
	"context"
	"fmt"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	channel_service "github.com/UnicomAI/wanwu/api/proto/channel-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/gin-gonic/gin"
)

// ListWanwuDIPAgents 获取数字员工列表（用于 DIP 通道选择绑定的数字员工）。
// 数据源复用 GetGeneralAgentOntologyEmployeeSelect（「知识网络构建专员」固定排第一），
// 返回 ListResult 以对齐前端 getAppSelect 期望的 {list, total} 结构。
func ListWanwuDIPAgents(ctx *gin.Context, userID, orgID, name string) (*response.ListResult, error) {
	employees, err := GetGeneralAgentOntologyEmployeeSelect(ctx, userID, orgID, name)
	if err != nil {
		return nil, err
	}
	return &response.ListResult{
		List:  employees,
		Total: int64(len(employees)),
	}, nil
}


func ListWanwuApiKeys(ctx *gin.Context, userID, orgID string) (*response.ListResult, error) {
	keys, err := app.ListApiKeys(ctx.Request.Context(), &app_service.ListApiKeysReq{
		PageNo:   1,
		PageSize: 1000, // 通道选择下拉场景，一次性获取全部
		UserIds:  []string{userID},
		OrgIds:   []string{orgID},
	})
	if err != nil {
		return nil, err
	}
	result := make([]*response.WanwuApiKeyResponse, 0, len(keys.Items))
	for _, key := range keys.Items {
		result = append(result, &response.WanwuApiKeyResponse{
			KeyID: key.KeyId,
			Key:   key.Key,
			Name:  key.Name,
			Desc:  key.Desc,
		})
	}
	return &response.ListResult{
		List:  result,
		Total: int64(len(result)),
	}, nil
}

// ListWanwuModels 获取万悟模型列表（JWT 认证，供 WGA 通道选择 modelUuid 使用）。
// 复用 ListModels 逻辑，裁剪为精简结构（uuid/displayName/provider/modelType/model），
// 与 OpenAPI /model/list 返回一致；默认仅返回已启用模型。
func ListWanwuModels(ctx *gin.Context, userID, orgID string, req *request.ListModelsRequest) (*response.ListResult, error) {
	full, err := ListModels(ctx, userID, orgID, req)
	if err != nil {
		return nil, err
	}
	if full == nil {
		return &response.ListResult{List: []response.OpenAPIModelListItem{}, Total: 0}, nil
	}
	models, ok := full.List.([]*response.ModelInfo)
	if !ok {
		log.Warnf("[Channel][ListWanwuModels] unexpected list type: %T, total=%d", full.List, full.Total)
		return &response.ListResult{List: []response.OpenAPIModelListItem{}, Total: full.Total}, nil
	}
	items := make([]*response.OpenAPIModelListItem, 0, len(models))
	for _, m := range models {
		if m == nil {
			continue
		}
		items = append(items, &response.OpenAPIModelListItem{
			UUID:        m.Uuid,
			DisplayName: m.DisplayName,
			Provider:    m.Provider,
			ModelType:   m.ModelType,
			Model:       m.Model,
			ScopeType:   m.ScopeType,
		})
	}
	return &response.ListResult{List: items, Total: full.Total}, nil
}

// ListWanwuAgents 获取万悟智能体列表（包含 UUID，供通道绑定使用）
func ListWanwuAgents(ctx *gin.Context, userID, orgID, appType, name string) (*response.ListResult, error) {
	// SearchType=all 返回本人+他人公开+本组织内应用，下方按「本人创建且已发布」过滤，
	// 与 OpenAPI 权限校验 CheckOpenAPIAccess 放行范围对齐，避免选到不可调用的应用。
	expResp, err := app.GetExplorationAppList(ctx.Request.Context(), &app_service.GetExplorationAppListReq{
		Name:       name,
		AppType:    appType,
		SearchType: "all",
		UserId:     userID,
		OrgId:      orgID,
	})
	if err != nil {
		return nil, err
	}
	if len(expResp.Infos) == 0 {
		return &response.ListResult{
			List:  []*response.WanwuAgentResponse{},
			Total: 0,
		}, nil
	}

	// 2. 收集 agent 类型的 appId，批量查询 UUID 和详情
	// 仅保留本人创建且已发布的应用：OpenAPI 只放行本人创建的应用（CheckOpenAPIAccess），
	// 未发布应用会被以 app not published 拒绝，故两者都不下拉。
	var agentAppIds []string
	for _, info := range expResp.Infos {
		if info.UserId != userID || info.OrgId != orgID || info.PublishType == "" {
			continue
		}
		if info.AppType == "agent" {
			agentAppIds = append(agentAppIds, info.AppId)
		}
	}

	// 批量查：appId -> (uuid, name, desc, avatarPath) 映射
	type agentDetail struct {
		uuid       string
		name       string
		desc       string
		avatarPath string
	}
	agentMap := make(map[string]*agentDetail) // appId -> detail
	if len(agentAppIds) > 0 {
		agentList, err := assistant.GetAssistantByIds(ctx.Request.Context(), &assistant_service.GetAssistantByIdsReq{
			AssistantIdList: agentAppIds,
			Identity: &assistant_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		if err == nil && agentList != nil {
			for _, a := range agentList.AssistantInfos {
				if a != nil && a.Info != nil {
					agentMap[a.Info.AppId] = &agentDetail{
						uuid:       a.Uuid,
						name:       a.Info.Name,
						desc:       a.Info.Desc,
						avatarPath: a.Info.AvatarPath,
					}
				}
			}
		}
	}

	// 3. 组装响应
	result := make([]*response.WanwuAgentResponse, 0, len(expResp.Infos))
	for _, info := range expResp.Infos {
		// 仅本人创建且已发布（与上方收集逻辑一致）
		if info.UserId != userID || info.OrgId != orgID || info.PublishType == "" {
			continue
		}
		agent := &response.WanwuAgentResponse{
			AppId:   info.AppId,
			AppType: info.AppType,
		}
		if detail, ok := agentMap[info.AppId]; ok {
			agent.AppId = detail.uuid // 使用 UUID 作为 AppId，创建通道时直接填入
			agent.Name = detail.name
			agent.Desc = detail.desc
			if detail.avatarPath != "" {
				agent.Avatar = CacheAvatar(detail.avatarPath).Path
			} else {
				agent.Avatar = config.Cfg().DefaultIcon.AgentIcon
			}
		}
		result = append(result, agent)
	}
	return &response.ListResult{
		List:  result,
		Total: int64(len(result)),
	}, nil
}

// --- 扫码登录 ---

// CreateQRLogin 发起扫码登录
func CreateQRLogin(ctx *gin.Context, channelType, userID, orgID string) (*response.QRLoginResponse, error) {
	resp, err := channel.CreateQRLogin(ctx.Request.Context(), &channel_service.CreateQRLoginReq{
		ChannelType: channelType,
		UserId:      userID,
		OrgId:       orgID,
	})
	if err != nil {
		return nil, err
	}
	return &response.QRLoginResponse{
		SessionID:  resp.SessionId,
		QrUrl:      resp.QrUrl,
		ExpireAt:   resp.ExpireAt,
		ExpireTime: resp.ExpireTime,
	}, nil
}

// GetQRLoginStatus 查询扫码状态
func GetQRLoginStatus(ctx *gin.Context, channelType, sessionID string) (*response.QRLoginStatusResponse, error) {
	resp, err := channel.GetQRLoginStatus(ctx.Request.Context(), &channel_service.GetQRLoginStatusReq{
		ChannelType: channelType,
		SessionId:   sessionID,
	})
	if err != nil {
		return nil, err
	}

	result := &response.QRLoginStatusResponse{
		Status:      resp.Status,
		Credentials: resp.Credentials,
		Error:       resp.Error,
	}

	// 从 credentials 中提取 baseUrl 作为顶层字段，方便前端直接获取
	if resp.Credentials != nil {
		if baseUrl, ok := resp.Credentials["baseUrl"]; ok {
			result.BaseUrl = baseUrl
		}
	}

	return result, nil
}

// CancelQRLogin 取消扫码登录
func CancelQRLogin(ctx *gin.Context, channelType, sessionID string) error {
	_, err := channel.CancelQRLogin(ctx.Request.Context(), &channel_service.CancelQRLoginReq{
		ChannelType: channelType,
		SessionId:   sessionID,
	})
	return err
}

// CompleteQRLogin 完成扫码登录（扫码成功后创建通道）
func CompleteQRLogin(ctx *gin.Context, channelType, sessionID, userID, orgID string) (*response.ChannelResponse, error) {
	resp, err := channel.CompleteQRLogin(ctx.Request.Context(), &channel_service.CompleteQRLoginReq{
		ChannelType: channelType,
		SessionId:   sessionID,
		UserId:      userID,
		OrgId:       orgID,
	})
	if err != nil {
		return nil, err
	}
	return protoToChannelResponse(resp), nil
}

// --- 通道管理 ---

// CreateChannel 创建通道
func CreateChannel(ctx *gin.Context, userID, orgID string, req request.CreateChannelRequest) (*response.ChannelResponse, error) {
	// 当前端传了 apiKeyId 但没传 apiKey 时，通过 app-service 查询完整 key 值
	apiKey := req.ApiKey
	if req.ApiKeyId != "" && apiKey == "" {
		key, err := resolveApiKeyByID(ctx, userID, orgID, req.ApiKeyId)
		if err != nil {
			return nil, err
		}
		apiKey = key
	}

	// 根据 appId 查询智能体名称
	appName := ""
	if req.AppID != "" {
		name, nameErr := resolveAppName(ctx, req.AppID, userID, orgID)
		if nameErr != nil {
			log.Warnf("failed to resolve app name for %s: %v", req.AppID, nameErr)
		}
		appName = name
	}

	resp, err := channel.CreateChannel(ctx.Request.Context(), &channel_service.CreateChannelReq{
		Name:        req.Name,
		ChannelType: req.ChannelType,
		AppType:     req.AppType,
		AppId:       req.AppID,
		AppName:     appName,
		ApiKeyId:    req.ApiKeyId,
		ApiKey:      apiKey,
		ModelUuid:   req.ModelUuid,
		AgentId:     req.AgentId,
		Config:      req.Config,
		UserId:      userID,
		OrgId:       orgID,
	})
	if err != nil {
		return nil, err
	}
	return protoToChannelResponse(resp), nil
}

// ListChannels 获取通道列表（分页）
func ListChannels(ctx *gin.Context, userID, orgID, name string, pageNo, pageSize int32) (*response.PageResult, error) {
	resp, err := channel.ListChannels(ctx.Request.Context(), &channel_service.ListChannelsReq{
		UserId:   userID,
		OrgId:    orgID,
		Name:     name,
		PageNo:   pageNo,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}

	channels := make([]*response.ChannelResponse, 0, len(resp.List))
	for _, ch := range resp.List {
		channels = append(channels, protoToChannelResponse(ch))
	}
	return &response.PageResult{
		List:     channels,
		Total:    int64(resp.Total),
		PageNo:   int(pageNo),
		PageSize: int(pageSize),
	}, nil
}

// GetChannel 获取通道详情
func GetChannel(ctx *gin.Context, channelID string) (*response.ChannelResponse, error) {
	resp, err := channel.GetChannel(ctx.Request.Context(), &channel_service.GetChannelReq{
		ChannelId: channelID,
	})
	if err != nil {
		return nil, err
	}
	return protoToChannelResponse(resp), nil
}

// UpdateChannel 更新通道
func UpdateChannel(ctx *gin.Context, channelID, userID, orgID string, req request.UpdateChannelRequest) (*response.ChannelResponse, error) {
	// 当前端传了 apiKeyId 但没传 apiKey 时，通过 app-service 查询完整 key 值
	apiKey := req.ApiKey
	if req.ApiKeyId != "" && apiKey == "" {
		key, err := resolveApiKeyByID(ctx, userID, orgID, req.ApiKeyId)
		if err != nil {
			return nil, err
		}
		apiKey = key
	}

	// 根据 appId 查询智能体名称
	appName := ""
	if req.AppID != "" {
		name, nameErr := resolveAppName(ctx, req.AppID, userID, orgID)
		if nameErr != nil {
			log.Warnf("failed to resolve app name for %s: %v", req.AppID, nameErr)
		}
		appName = name
	}

	resp, err := channel.UpdateChannel(ctx.Request.Context(), &channel_service.UpdateChannelReq{
		ChannelId: channelID,
		Name:      req.Name,
		AppId:     req.AppID,
		AppName:   appName,
		ApiKeyId:  req.ApiKeyId,
		ApiKey:    apiKey,
		ModelUuid: req.ModelUuid,
		AgentId:   req.AgentId,
		Config:    req.Config,
		UserId:    userID,
		OrgId:     orgID,
	})
	if err != nil {
		return nil, err
	}
	return protoToChannelResponse(resp), nil
}

// UpdateChannelStatus 启用/停用通道
func UpdateChannelStatus(ctx *gin.Context, channelID, userID, orgID string, req request.UpdateChannelStatusRequest) (*response.ChannelResponse, error) {
	resp, err := channel.UpdateChannelStatus(ctx.Request.Context(), &channel_service.UpdateChannelStatusReq{
		ChannelId: channelID,
		Enabled:   req.Enabled,
		UserId:    userID,
		OrgId:     orgID,
	})
	if err != nil {
		return nil, err
	}
	return protoToChannelResponse(resp), nil
}

// DeleteChannel 删除通道
func DeleteChannel(ctx *gin.Context, channelID, userID, orgID string) error {
	_, err := channel.DeleteChannel(ctx.Request.Context(), &channel_service.DeleteChannelReq{
		ChannelId: channelID,
		UserId:    userID,
		OrgId:     orgID,
	})
	return err
}

// DisconnectChannel 断开通道
func DisconnectChannel(ctx *gin.Context, channelID string) (*response.DisconnectChannelResponse, error) {
	_, err := channel.DisconnectChannel(ctx.Request.Context(), &channel_service.DisconnectChannelReq{
		ChannelId: channelID,
	})
	if err != nil {
		return nil, err
	}
	return &response.DisconnectChannelResponse{Message: "通道已断开"}, nil
}

// --- WGA（通用智能体）工作区 / 上传 ---

// GetWGAWorkspace 获取 WGA 工作区目录树（按 threadId + runId 隔离）。
// channel-service 侧校验通道归属与 threadId 归属，防越权。
func GetWGAWorkspace(ctx *gin.Context, channelID, userID, threadID, runID string) (*response.WGAWorkspaceResponse, error) {
	resp, err := channel.GetWGAWorkspace(ctx.Request.Context(), &channel_service.GetWGAWorkspaceReq{
		ChannelId: channelID,
		ThreadId:  threadID,
		RunId:     runID,
		UserId:    userID,
	})
	if err != nil {
		return nil, err
	}
	return wgaWorkspaceProtoToResp(resp), nil
}

// DownloadWGAWorkspace 下载 WGA 工作区文件（path 为空则下载整个工作区 ZIP）。
// 返回文件名 + 二进制字节，由 handler 写 octet-stream 响应。
func DownloadWGAWorkspace(ctx *gin.Context, channelID, userID, threadID, runID, path string) (string, []byte, error) {
	resp, err := channel.DownloadWGAWorkspace(ctx.Request.Context(), &channel_service.DownloadWGAWorkspaceReq{
		ChannelId: channelID,
		ThreadId:  threadID,
		RunId:     runID,
		Path:      path,
		UserId:    userID,
	})
	if err != nil {
		return "", nil, err
	}
	return resp.FileName, resp.Data, nil
}

// UploadWGAFile 上传文件到万悟 minio，返回 filePath 供 WGA 多模态对话 binary.url 引用。
func UploadWGAFile(ctx *gin.Context, channelID, userID, fileName, mimeType string, data []byte) (*response.WGAUploadFileResponse, error) {
	resp, err := channel.UploadWGAFile(ctx.Request.Context(), &channel_service.UploadWGAFileReq{
		ChannelId: channelID,
		FileName:  fileName,
		MimeType:  mimeType,
		Data:      data,
		UserId:    userID,
	})
	if err != nil {
		return nil, err
	}
	return &response.WGAUploadFileResponse{
		FileName: resp.FileName,
		FileId:   resp.FileId,
		FilePath: resp.FilePath,
		FileSize: resp.FileSize,
	}, nil
}

// wgaWorkspaceProtoToResp 将 proto 工作区响应转换为 bff 响应
func wgaWorkspaceProtoToResp(resp *channel_service.GetWGAWorkspaceResp) *response.WGAWorkspaceResponse {
	if resp == nil {
		return &response.WGAWorkspaceResponse{}
	}
	return &response.WGAWorkspaceResponse{
		ThreadID:  resp.ThreadId,
		RunID:     resp.RunId,
		FileCount: resp.FileCount,
		TotalSize: resp.TotalSize,
		IsDisplay: resp.IsDisplay,
		Path:      resp.Path,
		Files:     wgaFileNodesProtoToResp(resp.Files),
	}
}

func wgaFileNodesProtoToResp(nodes []*channel_service.WGAFileNode) []*response.WGAFileNode {
	if len(nodes) == 0 {
		return nil
	}
	out := make([]*response.WGAFileNode, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, &response.WGAFileNode{
			Name:     n.Name,
			Type:     n.Type,
			Size:     n.Size,
			MimeType: n.MimeType,
			Children: wgaFileNodesProtoToResp(n.Children),
		})
	}
	return out
}

// --- 内部方法 ---

// resolveApiKeyByID 通过 apiKeyId 查询完整的 apiKey 值
func resolveApiKeyByID(ctx *gin.Context, userID, orgID, apiKeyID string) (string, error) {
	keys, err := app.ListApiKeys(ctx.Request.Context(), &app_service.ListApiKeysReq{
		PageNo:   1,
		PageSize: 1000,
		UserIds:  []string{userID},
		OrgIds:   []string{orgID},
	})
	if err != nil {
		return "", fmt.Errorf("failed to query api key by id: %w", err)
	}
	for _, key := range keys.Items {
		if key.KeyId == apiKeyID {
			if !key.Status {
				return "", fmt.Errorf("api key %s is disabled", apiKeyID)
			}
			return key.Key, nil
		}
	}
	return "", fmt.Errorf("api key %s not found", apiKeyID)
}

// resolveAppName 根据 appId（UUID）查询智能体名称
func resolveAppName(ctx *gin.Context, appID, userID, orgID string) (string, error) {
	// 先通过 UUID 获取内部 appId
	internalID, err := assistant.GetAssistantIdByUuid(ctx.Request.Context(), &assistant_service.GetAssistantIdByUuidReq{
		Uuid: appID,
	})
	if err != nil {
		log.Errorf("failed to resolve app id from uuid %s: %v", appID, err)
		return "", fmt.Errorf("failed to resolve app id from uuid %s: %w", appID, err)
	}
	if internalID == nil || internalID.AssistantId == "" {
		log.Warnf("assistant not found for uuid %s: empty response", appID)
		return "", nil
	}

	// 通过内部 appId 获取智能体信息（必须传 Identity，否则 assistant-service 会 panic）
	info, err := assistant.GetAssistantInfo(ctx.Request.Context(), &assistant_service.GetAssistantInfoReq{
		AssistantId: internalID.AssistantId,
		Identity: &assistant_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		log.Errorf("failed to get assistant info for %s (uuid=%s): %v", internalID.AssistantId, appID, err)
		return "", fmt.Errorf("failed to get assistant info for %s: %w", internalID.AssistantId, err)
	}
	if info != nil && info.AssistantBrief != nil {
		return info.AssistantBrief.Name, nil
	}
	log.Warnf("assistant info empty for %s (uuid=%s)", internalID.AssistantId, appID)
	return "", nil
}

func protoToChannelResponse(ch *channel_service.Channel) *response.ChannelResponse {
	return &response.ChannelResponse{
		ID:          ch.Id,
		Name:        ch.Name,
		ChannelType: ch.ChannelType,
		Status:      ch.Status,
		AccountId:   ch.AccountId,
		Nickname:    ch.Nickname,
		Avatar:      ch.Avatar,
		Enabled:     ch.Enabled,
		AppType:     ch.AppType,
		AppId:       ch.AppId,
		AppName:     ch.AppName,
		ApiKeyId:    ch.ApiKeyId,
		ApiKeyName:  ch.ApiKeyName,
		HasApiKey:   ch.HasApiKey,
		ModelUuid:   ch.ModelUuid,
		AgentId:     ch.AgentId,
		Config:      ch.Config,
		CreatedAt:   ch.CreatedAt,
		UpdatedAt:   ch.UpdatedAt,
	}
}

// CheckChannelService 检查 channel-service 连接是否可用（用于健康检查）
func CheckChannelService(ctx context.Context) bool {
	// TODO: 添加健康检查逻辑
	return channel != nil
}
