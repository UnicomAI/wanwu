package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// GetGeneralAgentSubList
//
//	@Tags			wga
//	@Summary		获取通用智能体子智能体列表
//	@Description	获取通用智能体支持的子智能体列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.GetGeneralAgentSubListResp}
//	@Router			/general/agent/sub/list [get]
func GetGeneralAgentSubList(ctx *gin.Context) {
	resp, err := service.GetGeneralAgentSubList(ctx)
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentUploadLimit
//
//	@Tags			wga
//	@Summary		获取通用智能体上传文件格式限制
//	@Description	获取通用智能体所支持的上传文件格式及大小限制
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.GeneralAgentUploadLimitResp}
//	@Router			/general/agent/upload/limit [get]
func GetGeneralAgentUploadLimit(ctx *gin.Context) {
	resp, err := service.GetGeneralAgentUploadLimit(ctx, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// UpdateGeneralAgentConfig
//
//	@Tags			wga
//	@Summary		修改通用智能体配置
//	@Description	更新通用智能体工具配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateGeneralAgentConfigReq	true	"更新通用智能体配置请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/general/agent/config [put]
func UpdateGeneralAgentConfig(ctx *gin.Context) {
	var req request.UpdateGeneralAgentConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateGeneralAgentConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// GetGeneralAgentConfig
//
//	@Tags			wga
//	@Summary		获取通用智能体配置
//	@Description	获取通用智能体配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.GetGeneralAgentConfigResp}
//	@Router			/general/agent/config [get]
func GetGeneralAgentConfig(ctx *gin.Context) {
	resp, err := service.GetGeneralAgentConfig(ctx, getUserID(ctx), getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentAssistantSelect
//
//	@Tags			wga
//	@Summary		通用智能体智能体选择，只返回单智能体
//	@Description	获取通用智能体智能体选择，只返回单智能体
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"智能体名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.ExplorationAppInfo}}
//	@Router			/general/agent/assistant/select [get]
func GetGeneralAgentAssistantSelect(ctx *gin.Context) {
	req := request.GetExplorationAppListRequest{
		Name: ctx.Query("name"),
	}
	resp, err := service.GetAssistantSelectSingle(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentToolSelect
//
//	@Tags			wga
//	@Summary		通用智能体工具选择
//	@Description	获取通用智能体工具选择，用于用户选择工具进行对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			agentId	query		string	true	"智能体ID，默认为空"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.GetGeneralAgentToolSelectResp}}
//	@Router			/general/agent/tool/select [get]
func GetGeneralAgentToolSelect(ctx *gin.Context) {
	resp, err := service.GetGeneralAgentToolSelect(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("agentId"))
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentToolInfo
//
//	@Tags			wga
//	@Summary		通用智能体工具详情
//	@Description	获取通用智能体工具详情，用于工具调用
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			toolId		query		string	true	"工具ID"
//	@Param			toolType	query		string	true	"工具类型"
//	@Success		200			{object}	response.Response{data=response.GeneralAgentToolInfoResp}
//	@Router			/general/agent/tool/info [get]
func GetGeneralAgentToolInfo(ctx *gin.Context) {
	resp, err := service.GetGeneralAgentToolInfo(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("toolId"), ctx.Query("toolType"))
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentMCPSelect
//
//	@Tags			wga
//	@Summary		通用智能体MCP下拉接口列表
//	@Description	获取通用智能体MCP下拉接口列表，用于用户选择MCP进行对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"MCP名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.MCPSelect}}
//	@Router			/general/agent/mcp/select [get]
func GetGeneralAgentMCPSelect(ctx *gin.Context) {
	resp, err := service.GetMCPSelect(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentWorkflowSelect
//
//	@Tags			wga
//	@Summary		通用智能体工作流列表
//	@Description	获取通用智能体工作流列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"workflow名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.ExplorationAppInfo}}
//	@Router			/general/agent/workflow/select [get]
func GetGeneralAgentWorkflowSelect(ctx *gin.Context) {
	req := request.GetExplorationAppListRequest{
		Name: ctx.Query("name"),
	}
	resp, err := service.GetWorkflowSelect(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentSkillSelect
//
//	@Tags			wga
//	@Summary		获取通用智能体skill列表
//	@Description	获取通用智能体skill列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"skill名称"
//	@Success		200		{object}	response.Response{data=response.ListResult{list=[]response.PublishedSkillInfo}}
//	@Router			/general/agent/skill/select [get]
func GetGeneralAgentSkillSelect(ctx *gin.Context) {
	resp, err := service.GetCustomSkillList(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentKnowledgeSelect
//
//	@Tags			wga
//	@Summary		通用智能体知识库下拉接口列表
//	@Description	获取通用智能体知识库下拉接口列表
//	@Security		JWT
//	@Accept			json
//	@Param			data	body	request.KnowledgeSelectReq	true	"查询知识库列表"
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.KnowledgeListResp}
//	@Router			/general/agent/knowledge/select [post]
func GetGeneralAgentKnowledgeSelect(ctx *gin.Context) {
	userId, orgId := getUserID(ctx), getOrgID(ctx)
	var req request.KnowledgeSelectReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.SelectKnowledgeList(ctx, userId, orgId, &req)
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentOntologyEmployeeSelect
//
//	@Tags			wga
//	@Summary		通用智能体本体数字员工下拉接口列表
//	@Description	获取通用智能体本体数字员工下拉接口列表
//	@Security		JWT
//	@Accept			json
//	@Param			name	query	string	false	"数字员工名称"
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.KnowledgeListResp}
//	@Router			/general/agent/ontology/employee/select [post]
func GetGeneralAgentOntologyEmployeeSelect(ctx *gin.Context) {
	resp, err := service.GetGeneralAgentOntologyEmployeeSelect(ctx, getUserID(ctx), getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentResourceSelect
//
//	@Tags			wga
//	@Summary		通用智能体资源选择列表
//	@Description	获取通用智能体资源选择列表，包括MCP、Skill、Workflow、Assistant、Ontology五种类型
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.GeneralAgentResourceSelectList}
//	@Router			/general/agent/resource/select [get]
func GetGeneralAgentResourceSelect(ctx *gin.Context) {
	resp, err := service.GetGeneralAgentResourceSelect(ctx, getUserID(ctx), getOrgID(ctx), "")
	gin_util.Response(ctx, resp, err)
}

// CreateGeneralAgentConversation
//
//	@Tags			wga
//	@Summary		创建通用智能体对话
//	@Description	创建通用智能体对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateGeneralAgentConversationReq	true	"创建通用智能体对话请求参数"
//	@Success		200		{object}	response.Response{data=response.CreateGeneralAgentConversationResp}
//	@Router			/general/agent/conversation [post]
func CreateGeneralAgentConversation(ctx *gin.Context) {
	var req request.CreateGeneralAgentConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateGeneralAgentConversation(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// CreateGeneralAgentSkillConversation
//
//	@Tags			wga
//	@Summary		Create skill conversation
//	@Description	Create a dedicated conversation for skill creation.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CreateGeneralAgentSkillConversationReq	true	"create skill conversation request"
//	@Success		200		{object}	response.Response{data=response.CreateGeneralAgentSkillConversationResp}
//	@Router			/general/agent/skill/conversation [post]
func CreateGeneralAgentSkillConversation(ctx *gin.Context) {
	var req request.CreateGeneralAgentSkillConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CreateGeneralAgentSkillConversation(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// ImportGeneralAgentSkillConversation
//
//	@Tags			wga
//	@Summary		Import skill conversation
//	@Description	Import a skill zip into the skill conversation workspace.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ImportGeneralAgentSkillConversationReq	true	"import skill conversation request"
//	@Success		200		{object}	response.Response{data=response.ImportGeneralAgentSkillConversationResp}
//	@Router			/general/agent/skill/import/conversation [post]
func ImportGeneralAgentSkillConversation(ctx *gin.Context) {
	var req request.ImportGeneralAgentSkillConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.ImportGeneralAgentSkillConversation(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// ConvertGeneralAgentSkillConversation
//
//	@Tags			wga
//	@Summary		Convert resource to skill conversation
//	@Description	Convert MCP/tool/agent/workflow/RAG into a skill conversation workspace.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.ConvertGeneralAgentSkillConversationReq	true	"convert skill conversation request"
//	@Success		200		{object}	response.Response{data=response.ConvertGeneralAgentSkillConversationResp}
//	@Router			/general/agent/skill/convert/conversation [post]
func ConvertGeneralAgentSkillConversation(ctx *gin.Context) {
	var req request.ConvertGeneralAgentSkillConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.ConvertGeneralAgentSkillConversation(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// RefreshGeneralAgentSkillConversation
//
//	@Tags			wga
//	@Summary		Refresh skill conversation
//	@Description	Create a new WGA conversation with empty model config and bind it to an existing custom skill.
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.RefreshGeneralAgentSkillConversationReq	true	"refresh skill conversation request"
//	@Success		200		{object}	response.Response{data=response.RefreshGeneralAgentSkillConversationResp}
//	@Router			/general/agent/skill/refresh/conversation [post]
func RefreshGeneralAgentSkillConversation(ctx *gin.Context) {
	var req request.RefreshGeneralAgentSkillConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.RefreshGeneralAgentSkillConversation(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// DeleteGeneralAgentConversation
//
//	@Tags			wga
//	@Summary		删除通用智能体对话
//	@Description	删除通用智能体对话
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteGeneralAgentConversationReq	true	"删除通用智能体对话请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/general/agent/conversation [delete]
func DeleteGeneralAgentConversation(ctx *gin.Context) {
	var req request.DeleteGeneralAgentConversationReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.DeleteGeneralAgentConversation(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// GetGeneralAgentConversationList
//
//	@Tags			wga
//	@Summary		通用智能体对话列表
//	@Description	获取通用智能体对话历史列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"页码，默认1"
//	@Param			pageSize	query		int		false	"每页数量，默认20"
//	@Param			searchText	query		string	false	"检索词"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.GeneralAgentConversationInfo}}
//	@Router			/general/agent/conversation/list [get]
func GetGeneralAgentConversationList(ctx *gin.Context) {
	var req request.GetGeneralAgentConversationListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetGeneralAgentConversationList(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentConversationDetail
//
//	@Tags			wga
//	@Summary		通用智能体对话详情
//	@Description	获取指定会话的对话详情，包括对话标题、创建时间等信息
//	@Security		JWT
//	@Produce		json
//	@Param			threadId	query		string	true	"会话ID"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.GeneralAgentConversationDetailInfo}}
//	@Router			/general/agent/conversation/detail [get]
func GetGeneralAgentConversationDetail(ctx *gin.Context) {
	var req request.GetGeneralAgentConversationDetailReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetGeneralAgentConversationDetail(ctx, getUserID(ctx), getOrgID(ctx), req.ThreadID)
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentSkillPreviewConversationDetail
//
//	@Tags			wga
//	@Summary		获取 Skill preview 对话详情
//	@Description	用于回显 Skill preview 模式的历史消息
//	@Security		JWT
//	@Produce		json
//	@Param			previewId	query		string	true	"创建或导入接口返回的预览对话 ID"
//	@Success		200			{object}	response.Response{data=response.ListResult{list=[]response.GeneralAgentConversationDetailInfo}}
//	@Router			/general/agent/skill/preview/conversation/detail [get]
func GetGeneralAgentSkillPreviewConversationDetail(ctx *gin.Context) {
	var req request.GetGeneralAgentSkillPreviewConversationDetailReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetGeneralAgentSkillPreviewConversationDetail(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GetGeneralAgentConversationConfig
//
//	@Tags			wga
//	@Summary		通用智能体对话配置
//	@Description	获取指定会话的对话配置信息
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			threadId	query		string	true	"会话ID"
//	@Success		200			{object}	response.Response{data=response.GetGeneralAgentConversationConfigResp}
//	@Router			/general/agent/conversation/config [get]
func GetGeneralAgentConversationConfig(ctx *gin.Context) {
	var req request.GetGeneralAgentConversationConfigReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GetGeneralAgentConversationConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// UpdateGeneralAgentConversationConfig
//
//	@Tags			wga
//	@Summary		修改通用智能体对话配置
//	@Description	修改通用智能体对话配置
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UpdateGeneralAgentConversationConfigReq	true	"修改通用智能体对话配置请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/general/agent/conversation/config [put]
func UpdateGeneralAgentConversationConfig(ctx *gin.Context) {
	var req request.UpdateGeneralAgentConversationConfigReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.UpdateGeneralAgentConversationConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, nil, err)
}

// CheckGeneralAgentConversationConfig
//
//	@Tags			wga
//	@Summary		通用智能体配置检查接口
//	@Description	通用智能体配置检查接口
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GeneralAgentConfigCheckRequest	true	"通用智能体配置检查请求参数"
//	@Success		200		{object}	response.Response{data=response.GeneralAgentConfigCheckResponse}
//	@Router			/general/agent/conversation/config/check [post]
func CheckGeneralAgentConversationConfig(ctx *gin.Context) {
	var req request.GeneralAgentConfigCheckRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CheckGeneralAgentConversationConfig(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GeneralAgentWorkspaceDownload
//
//	@Tags			wga
//	@Summary		通用智能体workspace下载
//	@Description	通用智能体workspace下载接口
//	@Security		JWT
//	@Accept			json
//	@Produce		application/octet-stream
//	@Param			data	query	request.GeneralAgentWorkspaceDownloadReq	true	"workspace下载请求参数"
//	@Success		200		{file}	stream
//	@Router			/general/agent/conversation/workspace/download [get]
func GeneralAgentWorkspaceDownload(ctx *gin.Context) {
	var req request.GeneralAgentWorkspaceDownloadReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	fileName, data, err := service.GeneralAgentWorkspaceDownload(ctx, getUserID(ctx), getOrgID(ctx), req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.ResponseAttachment(ctx, fileName, data)
}

// GeneralAgentWorkspacePreview
//
//	@Tags			wga
//	@Summary		通用智能体workspace预览
//	@Description	通用智能体workspace预览接口，查看所给path的文件内容，返回文件内容用于前端预览
//	@Security		JWT
//	@Accept			json
//	@Produce		*/*
//	@Param			data	query	request.GeneralAgentWorkspacePreviewReq	true	"workspace预览请求参数"
//	@Success		200		{file}	stream
//	@Router			/general/agent/conversation/workspace/preview [get]
func GeneralAgentWorkspacePreview(ctx *gin.Context) {
	var req request.GeneralAgentWorkspacePreviewReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	fileName, data, contentType, err := service.GeneralAgentWorkspacePreview(ctx, getUserID(ctx), getOrgID(ctx), req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.ResponseInline(ctx, fileName, contentType, data)
}

// GeneralAgentWorkspaceInfo
//
//	@Tags			wga
//	@Summary		通用智能体workspace目录树
//	@Description	通用智能体workspace目录树接口，查看所给path的层级目录，返回目录结构与文件名等信息，类似于linux的tree命令
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	query		request.GeneralAgentWorkspaceReq	true	"workspace目录树请求参数"
//	@Success		200		{object}	response.Response{data=response.GeneralAgentWorkspaceResp}
//	@Router			/general/agent/conversation/workspace [get]
func GeneralAgentWorkspaceInfo(ctx *gin.Context) {
	var req request.GeneralAgentWorkspaceReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.GeneralAgentWorkspaceInfo(ctx, getUserID(ctx), getOrgID(ctx), req)
	gin_util.Response(ctx, resp, err)
}

// GeneralAgentConversationChat
//
//	@Tags			wga
//	@Summary		通用智能体对话流
//	@Description	通用智能体对话流，用于实时接收用户输入和获取智能体回复，SSE流式返回
//	@Security		JWT
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			data	body		request.GeneralAgentConversationChatReq	true	"通用智能体对话流请求参数"
//	@Success		200		{object}	string									"SSE流式返回"
//	@Router			/general/agent/conversation/chat [post]
func GeneralAgentConversationChat(ctx *gin.Context) {
	var req request.GeneralAgentConversationChatReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.GeneralAgentConversationChat(ctx, getUserID(ctx), getOrgID(ctx), getClientID(ctx), req, true)
	if err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// GeneralAgentSkillConversationChat
//
//	@Tags			wga
//	@Summary		Skill对话流
//	@Description	Skill对话流，用于创建自定义技能，SSE流式返回
//	@Security		JWT
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			data	body		request.GeneralAgentSkillConversationChatReq	true	"Skill对话请求参数"
//	@Success		200		{object}	string											"SSE流式返回"
//	@Router			/general/agent/skill/conversation/chat [post]
func GeneralAgentSkillConversationChat(ctx *gin.Context) {
	var req request.GeneralAgentSkillConversationChatReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.GeneralAgentSkillConversationChat(ctx, getUserID(ctx), getOrgID(ctx), getClientID(ctx), req)
	if err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// GeneralAgentConversationPending
//
//	@Tags			wga
//	@Summary		查询通用智能体运行中会话
//	@Description	查询通用智能体是否有运行中的会话，用于断线重连判断
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			threadId	query		string	true	"对话ID"
//	@Success		200			{object}	response.Response{data=response.WgaConversationPendingResp}
//	@Router			/general/agent/conversation/pending [get]
func GeneralAgentConversationPending(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.WgaConversationPendingReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.WgaConversationPending(ctx, userId, orgId, clientId, req)
	gin_util.Response(ctx, resp, err)
}

// GeneralAgentSkillConversationPending
//
//	@Tags			wga
//	@Summary		查询通用智能体(Skill)运行中会话
//	@Description	查询通用智能体(Skill)是否有运行中的会话，用于断线重连判断
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GeneralAgentSkillConversationPendingReq	true	"Skill对话请求参数"
//	@Success		200		{object}	response.Response{data=response.WgaConversationPendingResp}
//	@Router			/general/agent/skill/conversation/pending [get]
func GeneralAgentSkillConversationPending(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.GeneralAgentSkillConversationPendingReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.WgaConversationPending(ctx, userId, orgId, clientId, request.WgaConversationPendingReq{ThreadID: req.ChatThreadID()})
	gin_util.Response(ctx, resp, err)
}

// GeneralAgentConversationConnect
//
//	@Tags			wga
//	@Summary		通用智能体流式问答断线重连
//	@Description	通用智能体流式问答断线重连，续接之前未完成的SSE流
//	@Security		JWT
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			data	body		request.WgaConversationConnectReq	true	"断线重连请求参数"
//	@Success		200		{object}	string								"SSE流式返回"
//	@Router			/general/agent/conversation/connect [post]
func GeneralAgentConversationConnect(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.WgaConversationConnectReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.WgaConversationConnect(ctx, userId, orgId, clientId, req); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// GeneralAgentSkillConversationConnect
//
//	@Tags			wga
//	@Summary		通用智能体(Skill)流式问答断线重连
//	@Description	通用智能体(Skill)流式问答断线重连，续接之前未完成的SSE流
//	@Security		JWT
//	@Accept			json
//	@Produce		text/event-stream
//	@Param			data	body		request.GeneralAgentSkillConversationConnectReq	true	"Skill对话请求参数"
//	@Success		200		{object}	string											"SSE流式返回"
//	@Router			/general/agent/skill/conversation/connect [post]
func GeneralAgentSkillConversationConnect(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.GeneralAgentSkillConversationConnectReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if err := service.WgaConversationConnect(ctx, userId, orgId, clientId, request.WgaConversationConnectReq{ThreadID: req.ChatThreadID()}); err != nil {
		gin_util.Response(ctx, nil, err)
	}
}

// GeneralAgentConversationCancel
//
//	@Tags			wga
//	@Summary		通用智能体流式问答手动停止
//	@Description	通用智能体流式问答手动停止，区分网络断开和主动取消
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.WgaConversationCancelReq	true	"手动停止请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/general/agent/conversation/cancel [post]
func GeneralAgentConversationCancel(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.WgaConversationCancelReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.WgaConversationCancel(ctx, userId, orgId, clientId, req)
	gin_util.Response(ctx, nil, err)
}

// GeneralAgentSkillConversationCancel
//
//	@Tags			wga
//	@Summary		通用智能体(Skill)流式问答手动停止
//	@Description	通用智能体(Skill)流式问答手动停止，区分网络断开和主动取消
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GeneralAgentSkillConversationCancelReq	true	"Skill对话请求参数"
//	@Success		200		{object}	response.Response
//	@Router			/general/agent/skill/conversation/cancel [post]
func GeneralAgentSkillConversationCancel(ctx *gin.Context) {
	userId, orgId, clientId := getUserID(ctx), getOrgID(ctx), getClientID(ctx)
	var req request.GeneralAgentSkillConversationCancelReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.WgaConversationCancel(ctx, userId, orgId, clientId, request.WgaConversationCancelReq{ThreadID: req.ChatThreadID()})
	gin_util.Response(ctx, nil, err)
}

// GeneralAgentReplyQuestion
//
//	@Tags			wga
//	@Summary		回答问题
//	@Description	回答智能体提出的问题（Human-in-the-Loop），解除 AI 阻塞等待
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GeneralAgentReplyQuestionReq	true	"回答问题请求参数"
//	@Success		200		{object}	response.Response						"成功响应"
//	@Failure		400		{object}	response.Response						"参数错误"
//	@Failure		500		{object}	response.Response						"服务器错误"
//	@Router			/general/agent/question/reply [post]
func GeneralAgentReplyQuestion(ctx *gin.Context) {
	var req request.GeneralAgentReplyQuestionReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.GeneralAgentReplyQuestion(ctx, req.RunID, req.QuestionID, req.Answers)
	gin_util.Response(ctx, nil, err)
}

// GeneralAgentRejectQuestion
//
//	@Tags			wga
//	@Summary		拒绝问题
//	@Description	取消智能体提出的问题（Human-in-the-Loop），AI 将收到 RejectedError
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.GeneralAgentRejectQuestionReq	true	"拒绝问题请求参数"
//	@Success		200		{object}	response.Response						"成功响应"
//	@Failure		400		{object}	response.Response						"参数错误"
//	@Failure		500		{object}	response.Response						"服务器错误"
//	@Router			/general/agent/question/reject [post]
func GeneralAgentRejectQuestion(ctx *gin.Context) {
	var req request.GeneralAgentRejectQuestionReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.GeneralAgentRejectQuestion(ctx, req.RunID, req.QuestionID)
	gin_util.Response(ctx, nil, err)
}
