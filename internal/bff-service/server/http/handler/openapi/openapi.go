package openapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/gin-gonic/gin"
)

//	@title		AI Agent Productivity Platform - Open API
//	@version	v0.0.1

//	@BasePath	/openapi/v1

// CreateAgentConversation
//
//	@Tags			openapi
//	@Summary		智能体创建对话OpenAPI [EN] @Summary Agent creation dialogue OpenAPI
//	@Description	智能体创建对话OpenAPI [EN] @Description Agent creates conversationOpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIAgentCreateConversationRequest	true	"请求参数" [EN] @Param data body request.OpenAPIAgentCreateConversationRequest true "Request Parameters"
//	@Success		400		{object}	response.Response{data=response.OpenAPIAgentCreateConversationResponse}
//	@Router			/agent/conversation [post]
func CreateAgentConversation(ctx *gin.Context) {
	var req request.OpenAPIAgentCreateConversationRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	appID := getAppID(ctx)

	resp, err := service.ConversationCreate(ctx, userID, orgID, request.ConversationCreateRequest{
		AssistantId: appID,
		Prompt:      req.Title,
	})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	gin_util.Response(ctx, response.OpenAPIAgentCreateConversationResponse{ConversationID: resp.ConversationId}, nil)
}

// ChatAgent
//
//	@Tags			openapi
//	@Summary		智能体对话OpenAPI [EN] @Summary Agent Dialogue OpenAPI
//	@Description	智能体对话OpenAPI [EN] @Description Agent Dialogue OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIAgentChatRequest	true	"请求参数" [EN] @Param data body request.OpenAPIAgentChatRequest true "Request Parameters"
//	@Success		200		{object}	response.OpenAPIAgentChatResponse
//	@Success		400		{object}	response.Response
//	@Router			/agent/chat [post]
func ChatAgent(ctx *gin.Context) {
	var req request.OpenAPIAgentChatRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	appID := getAppID(ctx)

	// 流式 [EN] streaming
	if req.Stream {
		if err := service.AssistantConversionStream(ctx, userID, orgID, request.ConversionStreamRequest{
			AssistantId:    appID,
			ConversationId: req.ConversationID,
			Prompt:         req.Query,
			FileInfo:       []request.ConversionStreamFile{},
			Trial:          false,
		}); err != nil {
			gin_util.Response(ctx, nil, err)
		}
		return
	}
	// 非流式 [EN] non-streaming
	chatCh, err := service.CallAssistantConversationStream(ctx, userID, orgID, request.ConversionStreamRequest{
		AssistantId:    appID,
		ConversationId: req.ConversationID,
		Prompt:         req.Query,
		FileInfo:       []request.ConversionStreamFile{},
		Trial:          false,
	})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	var output string
	resp := &response.OpenAPIAgentChatResponse{}
	for chat := range chatCh {
		// 注意这里智能体的原始流式返回没有data:前缀 [EN] Note that the raw streaming return of the agent here does not have the data: prefix
		if strings.TrimSpace(chat) == "" {
			continue
		}
		curr := &response.OpenAPIAgentChatResponse{}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(chat, "data:")), curr); err != nil {
			log.Errorf("[Agent] %v conversation %v user %v org %v unmarshal %v err: %v", appID, req.ConversationID, userID, orgID, err)
			continue
		}
		resp = curr
		output += curr.Response
	}
	resp.Response = output
	b, _ := json.Marshal(resp)
	status := http.StatusOK
	ctx.Set(gin_util.STATUS, status)
	ctx.Set(gin_util.RESULT, string(b))
	ctx.JSON(status, resp)
}

// ChatRag
//
//	@Tags			openapi
//	@Summary		文本问答OpenAPI [EN] @Summary Text Q&AOpenAPI
//	@Description	文本问答OpenAPI [EN] @Description Text Q&AOpenAPI
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OpenAPIRagChatRequest	true	"请求参数" [EN] @Param data body request.OpenAPIRagChatRequest true "Request Parameters"
//	@Success		200		{object}	response.OpenAPIRagChatResponse
//	@Success		400		{object}	response.Response
//	@Router			/rag/chat [post]
func ChatRag(ctx *gin.Context) {
	var req request.OpenAPIRagChatRequest
	if !gin_util.Bind(ctx, &req) {
		return
	}
	userID := getUserID(ctx)
	orgID := getOrgID(ctx)
	appID := getAppID(ctx)

	// 流式 [EN] streaming
	if req.Stream {
		if err := service.ChatRagStream(ctx, userID, orgID, request.ChatRagRequest{RagID: appID, Question: req.Query, History: req.History}); err != nil {
			gin_util.Response(ctx, nil, err)
		}
		return
	}
	// 非流式 [EN] non-streaming
	chatCh, err := service.CallRagChatStream(ctx, userID, orgID, request.ChatRagRequest{RagID: appID, Question: req.Query, History: req.History})
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	var output string
	resp := &response.OpenAPIRagChatResponse{}
	for chat := range chatCh {
		if !strings.HasPrefix(chat, "data:") || strings.HasPrefix(chat, strings.TrimSpace(sse_util.DONE_MSG)) {
			continue
		}
		curr := &response.OpenAPIRagChatResponse{}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(chat, "data:")), curr); err != nil {
			log.Errorf("[RAG] %v user %v org %v unmarshal %v err: %v", appID, userID, orgID, err)
			continue
		}
		resp = curr
		output += curr.Data.Output
	}
	resp.Data.Output = output
	b, _ := json.Marshal(resp)
	status := http.StatusOK
	ctx.Set(gin_util.STATUS, status)
	ctx.Set(gin_util.RESULT, string(b))
	ctx.JSON(status, resp)
}

// WorkflowRun
//
//	@Tags			openapi
//	@Summary		工作流OpenAPI [EN] @Summary WorkflowOpenAPI
//	@Description	工作流OpenAPI [EN] @Description WorkflowOpenAPI
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response
//	@Router			/workflow/run [post]
func WorkflowRun(ctx *gin.Context) {
	var body []byte
	if cb, ok := ctx.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	resp, err := service.OpenAPIWorkflowRun(ctx, getAppID(ctx), body)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	_, err = ctx.Writer.Write(resp)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Writer.Flush()
}

// WorkflowFileUpload
//
//	@Tags			openapi
//	@Summary		工作流OpenAPI文件上传 [EN] @Summary WorkflowOpenAPI file upload
//	@Description	工作流OpenAPI文件上传 [EN] @Description WorkflowOpenAPI file upload
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			file	formData	file	true	"文件" [EN] @Param file formData file true "file"
//	@Success		200		{object}	string
//	@Success		400		{object}	response.Response
//	@Router			/workflow/file/upload [post]
func WorkflowFileUpload(ctx *gin.Context) {
	resp, err := service.OpenAPIWorkflowFileUpload(ctx)
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
	ctx.String(http.StatusOK, resp)
}

// GetMCPServerSSE
//
//	@Tags			openapi
//	@Summary		获取MCPServer SSE [EN] @Summary Get MCPServer SSE
//	@Description	获取MCPServer SSE [EN] @Description Get MCPServer SSE
//	@Accept			json
//	@Produce		json
//	@Param			key	query		string	true	"key"
//	@Success		200	{object}	response.Response{}
//	@Router			/mcp/server/sse [get]
func GetMCPServerSSE(ctx *gin.Context) {
	err := service.GetMCPServerSSE(ctx, getAppID(ctx), ctx.Query("key"))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
}

// GetMCPServerMessage
//
//	@Tags			openapi
//	@Summary		获取MCPServer Message [EN] @Summary Get MCPServer Message
//	@Description	获取MCPServer Message [EN] @Description Get MCPServer Message
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string	true	"mcpServerId"
//	@Success		200	{object}	response.Response{}
//	@Router			/mcp/server/message [post]
func GetMCPServerMessage(ctx *gin.Context) {
	err := service.GetMCPServerMessage(ctx, getAppID(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
}

// GetMCPServerStreamable
//
//	@Tags			openapi
//	@Summary		获取MCPServer streamable 类型消息 [EN] @Summary Get MCPServer streamable type message
//	@Description	获取MCPServer streamable 类型消息 [EN] @Description Get MCPServer streamable type message
//	@Accept			json
//	@Produce		json
//	@Param			key	query		string	true	"key"
//	@Success		200	{object}	response.Response{}
//	@Router			/mcp/server/streamable [post]
func GetMCPServerStreamable(ctx *gin.Context) {
	err := service.GetMCPServerStreamable(ctx, getAppID(ctx))
	if err != nil {
		gin_util.Response(ctx, nil, err)
		return
	}
}

// --- internal ---

// 获取当前用户ID [EN] Get current user ID
func getUserID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.USER_ID)
}

// 获取当前组织ID [EN] Get the current organization ID
func getOrgID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.X_ORG_ID)
}

// 获取当前appID [EN] Get current appID
func getAppID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.APP_ID)
}
