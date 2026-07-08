package v1

import (
	"io"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// validChannelTypes 允许的通道类型
var validChannelTypes = map[string]bool{
	"wechat":   true,
	"dingtalk": true,
	"feishu":   true,
}

// isValidChannelType 验证通道类型是否合法
func isValidChannelType(channelType string) bool {
	return validChannelTypes[channelType]
}

// --- 万悟平台代理 ---

// ListWanwuAgents 获取万悟智能体列表（含 UUID，供通道绑定使用）
func ListWanwuAgents(ctx *gin.Context) {
	var req request.ListWanwuAgentsRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.ListWanwuAgents(ctx, getUserID(ctx), getOrgID(ctx), req.AppType, req.Name)
	gin_util.Response(ctx, resp, err)
}

// ListWanwuApiKeys 获取万悟 API Key 列表（用于通道选择下拉）
func ListWanwuApiKeys(ctx *gin.Context) {
	resp, err := service.ListWanwuApiKeys(ctx, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// ListWanwuModels 获取万悟模型列表（用于 WGA 通道选择 modelUuid）
func ListWanwuModels(ctx *gin.Context) {
	var req request.ListModelsRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.ListWanwuModels(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ListWanwuWGASubAgents 获取 WGA 子智能体列表（用于 WGA 通道选择 agentId）
// 复用 service.GetGeneralAgentSubList，数据源为 WGA 配置 sub_agents。
// 过滤掉不应对外开放的子智能体：
//   - 数字员工（DIP Agent）：已单独配置通道（走 /channel/dip/employees），不返回给通道侧。
//   - Skill Chat Agent（创建Skill）：内部智能体，走独立的 skill 创建入口，不在通道选择列表展示。
func ListWanwuWGASubAgents(ctx *gin.Context) {
	resp, err := service.GetGeneralAgentSubList(ctx)
	if err != nil {
		gin_util.Response(ctx, resp, err)
		return
	}
	filtered := make([]response.GeneralAgentInfo, 0, len(resp.WgaAgentList))
	for _, a := range resp.WgaAgentList {
		if a.AgentID == "DIP Agent" || a.AgentID == "Skill Chat Agent" {
			continue
		}
		filtered = append(filtered, a)
	}
	resp.WgaAgentList = filtered
	gin_util.Response(ctx, resp, nil)
}

// ListWanwuDIPAgents 获取数字员工列表（用于 DIP 通道选择绑定的数字员工）
// 复用 service.ListWanwuDIPAgents → GetGeneralAgentOntologyEmployeeSelect。
func ListWanwuDIPAgents(ctx *gin.Context) {
	name := ctx.Query("name")
	resp, err := service.ListWanwuDIPAgents(ctx, getUserID(ctx), getOrgID(ctx), name)
	gin_util.Response(ctx, resp, err)
}

// --- 扫码登录 ---

// CreateQRLogin 发起扫码登录
func CreateQRLogin(ctx *gin.Context) {
	channelType := ctx.Param("channelType")
	if channelType == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_type_required")
		return
	}
	if !isValidChannelType(channelType) {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_type_invalid")
		return
	}
	resp, err := service.CreateQRLogin(ctx, channelType, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetQRLoginStatus 查询扫码状态
func GetQRLoginStatus(ctx *gin.Context) {
	channelType := ctx.Param("channelType")
	sessionID := ctx.Param("sessionId")
	if channelType == "" || sessionID == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_type_and_session_id_required")
		return
	}
	if !isValidChannelType(channelType) {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_type_invalid")
		return
	}
	resp, err := service.GetQRLoginStatus(ctx, channelType, sessionID)
	gin_util.Response(ctx, resp, err)
}

// CancelQRLogin 取消扫码登录
func CancelQRLogin(ctx *gin.Context) {
	channelType := ctx.Param("channelType")
	sessionID := ctx.Param("sessionId")
	if channelType == "" || sessionID == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_type_and_session_id_required")
		return
	}
	if !isValidChannelType(channelType) {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_type_invalid")
		return
	}
	err := service.CancelQRLogin(ctx, channelType, sessionID)
	gin_util.Response(ctx, nil, err)
}

// CompleteQRLogin 完成扫码登录（扫码成功后创建通道）
func CompleteQRLogin(ctx *gin.Context) {
	channelType := ctx.Param("channelType")
	sessionID := ctx.Param("sessionId")
	if channelType == "" || sessionID == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_type_and_session_id_required")
		return
	}
	if !isValidChannelType(channelType) {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_type_invalid")
		return
	}
	resp, err := service.CompleteQRLogin(ctx, channelType, sessionID, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// --- 通道管理 ---

// CreateChannel 创建通道
func CreateChannel(ctx *gin.Context) {
	var req request.CreateChannelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isValidChannelType(req.ChannelType) {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_type_invalid")
		return
	}
	resp, err := service.CreateChannel(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// ListChannels 获取通道列表
func ListChannels(ctx *gin.Context) {
	name := ctx.Query("name")
	resp, err := service.ListChannels(ctx, getUserID(ctx), getOrgID(ctx), name, getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetChannel 获取通道详情
func GetChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_id_required")
		return
	}
	resp, err := service.GetChannel(ctx, id)
	gin_util.Response(ctx, resp, err)
}

// UpdateChannel 更新通道
func UpdateChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_id_required")
		return
	}
	var req request.UpdateChannelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.UpdateChannel(ctx, id, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// UpdateChannelStatus 启用/停用通道
func UpdateChannelStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_id_required")
		return
	}
	var req request.UpdateChannelStatusRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.UpdateChannelStatus(ctx, id, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DeleteChannel 删除通道
func DeleteChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_id_required")
		return
	}
	err := service.DeleteChannel(ctx, id, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, nil, err)
}

// DisconnectChannel 断开通道
func DisconnectChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_id_required")
		return
	}
	resp, err := service.DisconnectChannel(ctx, id)
	gin_util.Response(ctx, resp, err)
}

// --- WGA（通用智能体）工作区 / 上传 ---

// GetWGAWorkspace WGA 工作区目录树
func GetWGAWorkspace(ctx *gin.Context) {
	channelID := ctx.Query("channelId")
	threadID := ctx.Query("threadId")
	if channelID == "" || threadID == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_id_and_thread_id_required")
		return
	}
	resp, err := service.GetWGAWorkspace(ctx, channelID, getUserID(ctx), threadID, ctx.Query("runId"))
	gin_util.Response(ctx, resp, err)
}

// DownloadWGAWorkspace WGA 工作区文件下载（不传 path 下载整个工作区 ZIP）
func DownloadWGAWorkspace(ctx *gin.Context) {
	channelID := ctx.Query("channelId")
	threadID := ctx.Query("threadId")
	if channelID == "" || threadID == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_id_and_thread_id_required")
		return
	}
	fileName, data, err := service.DownloadWGAWorkspace(ctx, channelID, getUserID(ctx), threadID, ctx.Query("runId"), ctx.Query("path"))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	ctx.Header("Content-Disposition", "attachment; filename*=utf-8''"+url.QueryEscape(fileName))
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Access-Control-Expose-Headers", "Content-Disposition")
	ctx.Data(http.StatusOK, "application/octet-stream", data)
}

// UploadWGAFile 通过通道上传文件给 WGA，返回 filePath 供多模态 binary.url 引用
func UploadWGAFile(ctx *gin.Context) {
	channelID := ctx.Query("channelId")
	if channelID == "" {
		gin_util.ResponseErrCodeKey(ctx, 110001, "channel_id_required")
		return
	}
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		gin_util.ResponseErrCodeKey(ctx, 110001, "file_required")
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		gin_util.ResponseErr(ctx, err)
		return
	}
	defer func() { _ = file.Close() }()
	data, err := io.ReadAll(file)
	if err != nil {
		gin_util.ResponseErr(ctx, err)
		return
	}
	mimeType := fileHeader.Header.Get("Content-Type")
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(fileHeader.Filename))
	}
	resp, err := service.UploadWGAFile(ctx, channelID, getUserID(ctx), fileHeader.Filename, mimeType, data)
	gin_util.Response(ctx, resp, err)
}
