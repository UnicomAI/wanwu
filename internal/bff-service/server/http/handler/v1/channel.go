package v1

import (
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
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

// ListWanwuAgents
//
//	@Tags			channel
//	@Summary		获取万悟智能体列表
//	@Description	获取万悟智能体列表（含 UUID，供通道绑定使用）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.ListWanwuAgentsRequest	true	"万悟智能体列表查询参数"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.WanwuAgentResponse}}
//	@Router			/channel/agent [get]
func ListWanwuAgents(ctx *gin.Context) {
	var req request.ListWanwuAgentsRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.ListWanwuAgents(ctx, getUserID(ctx), getOrgID(ctx), req.AppType, req.Name)
	gin_util.Response(ctx, resp, err)
}

// ListWanwuApiKeys
//
//	@Tags			channel
//	@Summary		获取万悟 API Key 列表
//	@Description	获取万悟 API Key 列表（用于通道选择下拉）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.ListResult{list=[]response.WanwuApiKeyResponse}}
//	@Router			/channel/apikeys [get]
func ListWanwuApiKeys(ctx *gin.Context) {
	resp, err := service.ListWanwuApiKeys(ctx, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// ListWanwuModels
//
//	@Tags			channel
//	@Summary		获取万悟模型列表
//	@Description	获取万悟模型列表（用于 WGA 通道选择 modelUuid）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.ListModelsRequest	true	"万悟模型列表查询参数"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.OpenAPIModelListItem}}
//	@Router			/channel/models [get]
func ListWanwuModels(ctx *gin.Context) {
	var req request.ListModelsRequest
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.ListWanwuModels(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ListWanwuWGASubAgents
//
//	@Tags			channel
//	@Summary		获取 WGA 子智能体列表
//	@Description	获取 WGA 子智能体列表（用于 WGA 通道选择 agentId），首部置入“无”选项，过滤数字员工、Skill Chat Agent、Data Analysis Agent 等内部智能体
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.GetGeneralAgentSubListResp}
//	@Router			/channel/wga/sub-agents [get]
func ListWanwuWGASubAgents(ctx *gin.Context) {
	resp, err := service.GetGeneralAgentSubList(ctx)
	if err != nil {
		gin_util.Response(ctx, resp, err)
		return
	}
	filtered := make([]response.GeneralAgentInfo, 0, len(resp.WgaAgentList)+1)
	// 首部置入“无”选项：agentId 用固定哨兵 "null"，创建通道时选它即忽略子智能体、使用通用智能体
	// （agent_id 存 "null"，channel-service 调 WGA 时归一化为空串走 Supervisor 默认路由）。
	filtered = append(filtered, response.GeneralAgentInfo{
		AgentID:     "null",
		AgentName:   "无",
		Avatar:      request.Avatar{Key: "", Path: ""},
		Placeholder: "选择一款模型，和我对话吧",
	})
	for _, a := range resp.WgaAgentList {
		if a.AgentID == "DIP Agent" || a.AgentID == "Skill Chat Agent" || a.AgentID == "Data Analysis Agent" {
			continue
		}
		filtered = append(filtered, a)
	}
	resp.WgaAgentList = filtered
	gin_util.Response(ctx, resp, nil)
}

// ListWanwuDIPAgents
//
//	@Tags			channel
//	@Summary		获取数字员工列表
//	@Description	获取数字员工列表（用于 DIP 通道选择绑定的数字员工）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"数字员工名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.GeneralAgentOntologyEmployee}}
//	@Router			/channel/dip/employees [get]
func ListWanwuDIPAgents(ctx *gin.Context) {
	name := ctx.Query("name")
	resp, err := service.ListWanwuDIPAgents(ctx, getUserID(ctx), getOrgID(ctx), name)
	gin_util.Response(ctx, resp, err)
}

// --- 扫码登录 ---

// CreateQRLogin
//
//	@Tags			channel
//	@Summary		发起扫码登录
//	@Description	发起扫码登录，返回二维码地址与会话ID
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			channelType	path		string	true	"通道类型：wechat/dingtalk/feishu"
//	@Success		200			{object}	response.Response{data=response.QRLoginResponse}
//	@Router			/channel/qrcode/{channelType} [post]
func CreateQRLogin(ctx *gin.Context) {
	channelType := ctx.Param("channelType")
	if channelType == "" {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_type_required")
		return
	}
	if !isValidChannelType(channelType) {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_type_invalid")
		return
	}
	resp, err := service.CreateQRLogin(ctx, channelType, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetQRLoginStatus
//
//	@Tags			channel
//	@Summary		查询扫码状态
//	@Description	查询扫码状态
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			channelType	path		string	true	"通道类型：wechat/dingtalk/feishu"
//	@Param			sessionId	path		string	true	"扫码会话ID"
//	@Success		200			{object}	response.Response{data=response.QRLoginStatusResponse}
//	@Router			/channel/qrcode/{channelType}/status/{sessionId} [get]
func GetQRLoginStatus(ctx *gin.Context) {
	channelType := ctx.Param("channelType")
	sessionID := ctx.Param("sessionId")
	if channelType == "" || sessionID == "" {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_type_and_session_id_required")
		return
	}
	if !isValidChannelType(channelType) {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_type_invalid")
		return
	}
	resp, err := service.GetQRLoginStatus(ctx, channelType, sessionID)
	gin_util.Response(ctx, resp, err)
}

// CancelQRLogin
//
//	@Tags			channel
//	@Summary		取消扫码登录
//	@Description	取消扫码登录
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			channelType	path		string	true	"通道类型：wechat/dingtalk/feishu"
//	@Param			sessionId	path		string	true	"扫码会话ID"
//	@Success		200			{object}	response.Response
//	@Router			/channel/qrcode/{channelType}/{sessionId} [delete]
func CancelQRLogin(ctx *gin.Context) {
	channelType := ctx.Param("channelType")
	sessionID := ctx.Param("sessionId")
	if channelType == "" || sessionID == "" {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_type_and_session_id_required")
		return
	}
	if !isValidChannelType(channelType) {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_type_invalid")
		return
	}
	err := service.CancelQRLogin(ctx, channelType, sessionID)
	gin_util.Response(ctx, nil, err)
}

// CompleteQRLogin
//
//	@Tags			channel
//	@Summary		完成扫码登录
//	@Description	完成扫码登录（扫码成功后创建通道）
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			channelType	path		string	true	"通道类型：wechat/dingtalk/feishu"
//	@Param			sessionId	path		string	true	"扫码会话ID"
//	@Success		200			{object}	response.Response{data=response.ChannelResponse}
//	@Router			/channel/qrcode/{channelType}/complete/{sessionId} [post]
func CompleteQRLogin(ctx *gin.Context) {
	channelType := ctx.Param("channelType")
	sessionID := ctx.Param("sessionId")
	if channelType == "" || sessionID == "" {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_type_and_session_id_required")
		return
	}
	if !isValidChannelType(channelType) {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_type_invalid")
		return
	}
	resp, err := service.CompleteQRLogin(ctx, channelType, sessionID, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// --- 通道管理 ---

// CreateChannel
//
//	@Tags			channel
//	@Summary		创建通道
//	@Description	创建通道
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateChannelRequest	true	"创建通道请求参数"
//	@Success		200		{object}	response.Response{data=response.ChannelResponse}
//	@Router			/channel/channels [post]
func CreateChannel(ctx *gin.Context) {
	var req request.CreateChannelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isValidChannelType(req.ChannelType) {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_type_invalid")
		return
	}
	resp, err := service.CreateChannel(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// ListChannels
//
//	@Tags			channel
//	@Summary		获取通道列表
//	@Description	获取通道列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"通道名称"
//	@Param			pageNo		query		int		true	"页面编号，从1开始"
//	@Param			pageSize	query		int		true	"单页数量，从1开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.ChannelResponse}}
//	@Router			/channel/channels [get]
func ListChannels(ctx *gin.Context) {
	name := ctx.Query("name")
	resp, err := service.ListChannels(ctx, getUserID(ctx), getOrgID(ctx), name, getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetChannel
//
//	@Tags			channel
//	@Summary		获取通道详情
//	@Description	获取通道详情
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"通道ID"
//	@Success		200	{object}	response.Response{data=response.ChannelResponse}
//	@Router			/channel/channels/{id} [get]
func GetChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_id_required")
		return
	}
	resp, err := service.GetChannel(ctx, id)
	gin_util.Response(ctx, resp, err)
}

// UpdateChannel
//
//	@Tags			channel
//	@Summary		更新通道
//	@Description	更新通道
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string							true	"通道ID"
//	@Param			data	body		request.UpdateChannelRequest	true	"更新通道请求参数"
//	@Success		200		{object}	response.Response{data=response.ChannelResponse}
//	@Router			/channel/channels/{id} [put]
func UpdateChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_id_required")
		return
	}
	var req request.UpdateChannelRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.UpdateChannel(ctx, id, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// UpdateChannelStatus
//
//	@Tags			channel
//	@Summary		启用/停用通道
//	@Description	启用/停用通道
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"通道ID"
//	@Param			data	body		request.UpdateChannelStatusRequest	true	"启用/停用通道请求参数"
//	@Success		200		{object}	response.Response{data=response.ChannelResponse}
//	@Router			/channel/channels/{id}/status [post]
func UpdateChannelStatus(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_id_required")
		return
	}
	var req request.UpdateChannelStatusRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.UpdateChannelStatus(ctx, id, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DeleteChannel
//
//	@Tags			channel
//	@Summary		删除通道
//	@Description	删除通道
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"通道ID"
//	@Success		200	{object}	response.Response
//	@Router			/channel/channels/{id} [delete]
func DeleteChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_id_required")
		return
	}
	err := service.DeleteChannel(ctx, id, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, nil, err)
}

// DisconnectChannel
//
//	@Tags			channel
//	@Summary		断开通道
//	@Description	断开通道
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"通道ID"
//	@Success		200	{object}	response.Response{data=response.DisconnectChannelResponse}
//	@Router			/channel/channels/{id}/disconnect [post]
func DisconnectChannel(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		gin_util.ResponseErrCodeKey(ctx, err_code.Code_BFFInvalidArg, "channel_id_required")
		return
	}
	resp, err := service.DisconnectChannel(ctx, id)
	gin_util.Response(ctx, resp, err)
}
