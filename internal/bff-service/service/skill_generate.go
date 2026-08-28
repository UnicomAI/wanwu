package service

import (
	"context"
	"encoding/json"
	"fmt"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	mcp2skill "github.com/UnicomAI/wanwu/pkg/mcp2skill"
	openapi2skill "github.com/UnicomAI/wanwu/pkg/openapi2skill"
	pkg_util "github.com/UnicomAI/wanwu/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

// toMCP2SkillAuth converts pkg/util.ApiAuthWebRequest to mcp2skill.APIAuthConfig.
// This bridges the service layer (which uses pkg/util types from proto conversion)
// with the mcp2skill package (which uses a self-contained auth type to avoid
// pulling in heavy proto/log dependencies in the standalone binary).
func toMCP2SkillAuth(auth *pkg_util.ApiAuthWebRequest) *mcp2skill.APIAuthConfig {
	if auth == nil {
		return nil
	}
	return &mcp2skill.APIAuthConfig{
		AuthType:           auth.AuthType,
		ApiKeyHeaderPrefix: auth.ApiKeyHeaderPrefix,
		ApiKeyHeader:       auth.ApiKeyHeader,
		ApiKeyQueryParam:   auth.ApiKeyQueryParam,
		ApiKeyValue:        auth.ApiKeyValue,
	}
}

// GenerateSkillFromMCP generates skill files from an MCP server.
// It fetches MCP connection info by mcpID via gRPC, then uses mcp2skill
// to connect to the server, list tools, and write skill output to outputDir.
func GenerateSkillFromMCP(ctx *gin.Context, mcpID, outputDir string) error {
	mcpDetail, err := mcp.GetCustomMCP(ctx.Request.Context(), &mcp_service.GetCustomMCPReq{
		McpId: mcpID,
	})
	if err != nil {
		return fmt.Errorf("failed to get MCP detail: %w", err)
	}

	transport := mcpDetail.Transport
	if transport == "" {
		transport = "streamable"
	}

	cfg := &mcp2skill.MCPConfig{
		Name:          mcpDetail.Info.Name,
		Description:   mcpDetail.Info.Desc,
		StreamableUrl: mcpDetail.StreamableUrl,
		SseUrl:        mcpDetail.SseUrl,
		Transport:     transport,
		ApiAuth:       toMCP2SkillAuth(toApiAuthPtr(mcpDetail.GetApiAuth())),
		Headers:       mcpDetail.GetHeaders(),
	}

	return mcp2skill.ConvertFromMCPConfig(ctx.Request.Context(), cfg, outputDir)
}

// GenerateSkillFromCustomTool generates skill files from a custom tool's OpenAPI schema.
// It fetches the schema by customToolID via gRPC, overrides the schema's title with
// info.Name and prepends info.Description to the schema's description, then converts.
func GenerateSkillFromCustomTool(ctx *gin.Context, customToolID, outputDir string) error {
	info, err := mcp.GetCustomToolInfo(ctx.Request.Context(), &mcp_service.GetCustomToolInfoReq{
		CustomToolId: customToolID,
	})
	if err != nil {
		return fmt.Errorf("failed to get custom tool info: %w", err)
	}

	specData, err := overrideSchemaMeta([]byte(info.Schema), info.Name, info.Description)
	if err != nil {
		return fmt.Errorf("failed to override schema meta: %w", err)
	}

	return generateSkillFromOpenAPISchema(ctx.Request.Context(), specData, outputDir)
}

// GenerateSkillFromAgent generates skill files for the wanwu agent API.
// It takes the agent's numeric ID, resolves it to a UUID via gRPC, fetches name
// and description, embeds both the UUID and the metadata into the OpenAPI spec,
// then converts to skill.
// userID and orgID are extracted from the gin context.
func GenerateSkillFromAgent(ctx *gin.Context, id, outputDir string) error {
	userID := ctx.GetString(gin_util.USER_ID)
	orgID := ctx.GetHeader(gin_util.X_ORG_ID)

	uuid, name, desc, err := getAgentMetaByID(ctx, userID, orgID, id)
	if err != nil {
		return fmt.Errorf("failed to get agent meta: %w", err)
	}

	specData, err := renderWanwuOpenAPISpec(SkillCategoryAgent, uuid, name, desc)
	if err != nil {
		return fmt.Errorf("failed to render agent OpenAPI spec: %w", err)
	}
	return generateSkillFromOpenAPISchema(ctx.Request.Context(), specData, outputDir)
}

// GenerateSkillFromWorkflow generates skill files for the wanwu workflow API.
// It takes the workflow ID (which is the UUID), fetches name and description
// via HTTP, then converts to skill.
func GenerateSkillFromWorkflow(ctx *gin.Context, id, outputDir string) error {
	name, desc, err := getWorkflowMeta(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get workflow meta: %w", err)
	}

	specData, err := renderWanwuOpenAPISpec(SkillCategoryWorkflow, id, name, desc)
	if err != nil {
		return fmt.Errorf("failed to render workflow OpenAPI spec: %w", err)
	}
	return generateSkillFromOpenAPISchema(ctx.Request.Context(), specData, outputDir)
}

// GenerateSkillFromRAG generates skill files for the wanwu RAG API.
// It takes the RAG application ID (which is the ragId UUID), fetches name and
// description via gRPC, then converts to skill.
func GenerateSkillFromRAG(ctx *gin.Context, id, outputDir string) error {
	name, desc, err := getRAGMeta(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get rag meta: %w", err)
	}

	specData, err := renderWanwuOpenAPISpec(SkillCategoryRAG, id, name, desc)
	if err != nil {
		return fmt.Errorf("failed to render rag OpenAPI spec: %w", err)
	}
	return generateSkillFromOpenAPISchema(ctx.Request.Context(), specData, outputDir)
}

// generateSkillFromOpenAPISchema converts raw OpenAPI spec bytes into skill files.
func generateSkillFromOpenAPISchema(ctx context.Context, specData []byte, outputDir string) error {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(specData)
	if err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}
	if err := doc.Validate(ctx); err != nil {
		return fmt.Errorf("invalid OpenAPI spec: %w", err)
	}

	opts := openapi2skill.ConvertOptions{
		OutputDir: outputDir,
	}
	return openapi2skill.ConvertDoc(doc, opts)
}

// --- Metadata fetchers ---

// getAgentMetaByID resolves a numeric agent ID to UUID and fetches name/description.
func getAgentMetaByID(ctx *gin.Context, userID, orgID, id string) (uuid, name, desc string, err error) {
	resp, err := assistant.GetAssistantInfo(ctx.Request.Context(), &assistant_service.GetAssistantInfoReq{
		AssistantId: id,
		Identity: &assistant_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		return "", "", "", fmt.Errorf("get assistant info by id: %w", err)
	}

	uuid = resp.Uuid
	name = resp.AssistantBrief.GetName()
	desc = resp.AssistantBrief.GetDesc()
	if uuid == "" {
		return "", "", "", fmt.Errorf("assistant %s has no uuid", id)
	}
	if name == "" {
		name = uuid
	}
	return uuid, name, desc, nil
}

// getWorkflowMeta fetches the workflow's name and description by ID (UUID).
func getWorkflowMeta(ctx *gin.Context, id string) (string, string, error) {
	workflowData, err := ListWorkflowByIDs(ctx, "", []string{id})
	if err != nil {
		return "", "", fmt.Errorf("get workflow info: %w", err)
	}
	if len(workflowData.Workflows) == 0 {
		return id, "", nil // fallback: use id as name, empty desc
	}
	name := workflowData.Workflows[0].Name
	desc := workflowData.Workflows[0].Desc
	if name == "" {
		name = id
	}
	return name, desc, nil
}

// getRAGMeta fetches the RAG application's name and description by ID (ragId UUID).
func getRAGMeta(ctx *gin.Context, id string) (string, string, error) {
	resp, err := rag.GetRagByIds(ctx.Request.Context(), &rag_service.GetRagByIdsReq{
		RagIdList: []string{id},
	})
	if err != nil {
		return "", "", fmt.Errorf("get rag info: %w", err)
	}
	if len(resp.RagInfos) == 0 {
		return id, "", nil // fallback: use id as name, empty desc
	}
	name := resp.RagInfos[0].Name
	desc := resp.RagInfos[0].Desc
	if name == "" {
		name = id
	}
	return name, desc, nil
}

// overrideSchemaMeta replaces the OpenAPI spec's info.title with name and
// prepends description to info.description.
func overrideSchemaMeta(schemaData []byte, name, description string) ([]byte, error) {
	var spec map[string]json.RawMessage
	if err := json.Unmarshal(schemaData, &spec); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	infoRaw, ok := spec["info"]
	if !ok {
		return nil, fmt.Errorf("openapi spec missing info section")
	}

	var info map[string]json.RawMessage
	if err := json.Unmarshal(infoRaw, &info); err != nil {
		return nil, fmt.Errorf("invalid info section: %w", err)
	}

	// Override title with name.
	if name != "" {
		title, _ := json.Marshal(name)
		info["title"] = title
	}

	// Prepend description to existing description.
	if description != "" {
		var newDesc string
		if existingRaw, exists := info["description"]; exists {
			var existing string
			_ = json.Unmarshal(existingRaw, &existing)
			newDesc = description + ";" + existing
		} else {
			newDesc = description
		}
		descJSON, _ := json.Marshal(newDesc)
		info["description"] = descJSON
	}

	updatedInfo, _ := json.Marshal(info)
	spec["info"] = updatedInfo
	return json.Marshal(spec)
}
