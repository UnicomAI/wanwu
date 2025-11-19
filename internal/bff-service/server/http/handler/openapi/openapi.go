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
//	@Summary Agent creation dialogue OpenAPI
//	@Description Agent creates conversationOpenAPI
//	@Accept			json
//	@Produce		json
//	@Param data body request.OpenAPIAgentCreateConversationRequest true "Request Parameters"
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
//	@Summary Agent Dialogue OpenAPI
//	@Description Agent Dialogue OpenAPI
//	@Accept			json
//	@Produce		json
//	@Param data body request.OpenAPIAgentChatRequest true "Request Parameters"
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

	// streaming
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
	// non-streaming
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
		// Note that the raw streaming return of the agent here does not have the data: prefix
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
//	@Summary Text Q&AOpenAPI
//	@Description Text Q&AOpenAPI
//	@Accept			json
//	@Produce		json
//	@Param data body request.OpenAPIRagChatRequest true "Request Parameters"
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

	// streaming
	if req.Stream {
		if err := service.ChatRagStream(ctx, userID, orgID, request.ChatRagRequest{RagID: appID, Question: req.Query, History: req.History}); err != nil {
			gin_util.Response(ctx, nil, err)
		}
		return
	}
	// non-streaming
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
//	@Summary WorkflowOpenAPI
//	@Description WorkflowOpenAPI
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
//	@Summary WorkflowOpenAPI file upload
//	@Description WorkflowOpenAPI file upload
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param file formData file true "file"
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
//	@Summary Get MCPServer SSE
//	@Description Get MCPServer SSE
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
//	@Summary Get MCPServer Message
//	@Description Get MCPServer Message
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
//	@Summary Get MCPServer streamable type message
//	@Description Get MCPServer streamable type message
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

// Get current user ID
func getUserID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.USER_ID)
}

// Get the current organization ID
func getOrgID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.X_ORG_ID)
}

// Get current appID
func getAppID(ctx *gin.Context) string {
	return ctx.GetString(gin_util.APP_ID)
}
