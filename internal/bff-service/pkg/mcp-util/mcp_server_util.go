package mcp_util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/server"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/getkin/kin-openapi/openapi3"
)

type McpTool struct {
	schema  string
	auth    *openapi3_util.Auth
	tool    *protocol.Tool
	handler server.ToolHandlerFunc
	strict  bool
}

func (tool *McpTool) Name() string { return tool.tool.Name }

func (tool *McpTool) Desc() string { return tool.tool.Description }

func (tool *McpTool) Schema() string { return tool.schema }

func (tool *McpTool) Auth() *openapi3_util.Auth { return tool.auth }

func (tool *McpTool) Update(ctx context.Context, name, desc string) (*McpTool, error) {
	doc, err := openapi3_util.LoadFromData(ctx, []byte(tool.schema))
	if err != nil {
		return nil, err
	}
	paths := doc.Paths
	doc.Paths = nil
	var exist bool
	for path, pathItem := range paths {
		for method, operation := range pathItem.Operations() {
			if operation.OperationID == tool.tool.Name {
				operation.OperationID = name
				operation.Description = desc
				doc.AddOperation(path, method, operation)
				exist = true
				break
			}
		}
		if exist {
			break
		}
	}
	if !exist {
		return nil, errors.New("mcp tool schema operation not found")
	}

	b, err := doc.MarshalJSON()
	if err != nil {
		return nil, err
	}
	newTool, err := openapi3_util.Doc2MCPProtocolTool(doc, name)
	if err != nil {
		return nil, err
	}
	return &McpTool{
		schema:  string(b),
		tool:    newTool,
		auth:    tool.auth,
		handler: genMcpToolHandler(doc, tool.auth, name, tool.strict),
		strict:  tool.strict,
	}, nil
}

func CreateMcpTool(ctx context.Context, schema string, auth *openapi3_util.Auth, operationID string, strict bool) (*McpTool, error) {
	doc, err := openapi3_util.LoadFromData(ctx, []byte(schema))
	if err != nil {
		return nil, err
	}
	b, err := openapi3_util.FilterDocOperations(doc, []string{operationID}).MarshalJSON()
	if err != nil {
		return nil, err
	}
	tool, err := openapi3_util.Doc2MCPProtocolTool(doc, operationID)
	if err != nil {
		return nil, err
	}
	return &McpTool{
		schema:  string(b),
		tool:    tool,
		auth:    auth,
		handler: genMcpToolHandler(doc, auth, operationID, strict),
		strict:  strict,
	}, nil
}

func CreateMcpTools(ctx context.Context, schema string, auth *openapi3_util.Auth, operationIDs []string, strict bool) ([]*McpTool, error) {
	var mcpTools []*McpTool
	for _, operationID := range operationIDs {
		mcpTool, err := CreateMcpTool(ctx, schema, auth, operationID, strict)
		if err != nil {
			return nil, err
		}
		mcpTools = append(mcpTools, mcpTool)
	}
	return mcpTools, nil
}

func genMcpToolHandler(doc *openapi3.T, auth *openapi3_util.Auth, operationID string, strict bool) server.ToolHandlerFunc {
	return func(ctx context.Context, req *protocol.CallToolRequest) (*protocol.CallToolResult, error) {

		// operation
		var operation *openapi3.Operation
		for _, pathItem := range doc.Paths {
			for _, op := range pathItem.Operations() {
				if op.OperationID == operationID {
					operation = op
					break
				}
			}
			if operation != nil {
				break
			}
		}
		if operation == nil {
			return nil, fmt.Errorf("operationId(%v) not found", operationID)
		}

		// params
		params := req.Arguments

		headerParams := make(map[string]string)
		pathParams := make(map[string]interface{})
		queryParams := make(map[string]interface{})
		bodyParams := make(map[string]interface{})

		for _, param := range operation.Parameters {
			if param.Value == nil {
				continue
			}
			field := param.Value.In + "-" + param.Value.Name
			switch param.Value.In {
			case "path":
				pathParams[param.Value.Name] = params[field]
			case "query":
				queryParams[param.Value.Name] = params[field]
			case "header":
				headerParams[param.Value.Name] = params[field].(string)
			}
		}
		if operation.RequestBody != nil && operation.RequestBody.Value != nil {
			for _, mediaType := range operation.RequestBody.Value.Content {
				if mediaType.Schema != nil && mediaType.Schema.Value != nil {
					for propName := range mediaType.Schema.Value.Properties {
						bodyParams[propName] = params[propName]
					}
				}
			}
		}

		// auth
		if auth != nil && auth.Type != "" && auth.Type != "none" && auth.Value != "" {
			switch auth.In {
			case "header":
				headerParams[auth.Name] = auth.Value
			case "query":
				queryParams[auth.Name] = auth.Value
			}
		}

		// http client
		client := openapi3_util.NewClientByDoc(doc)
		if strict {
			client = openapi3_util.NewStrictClientByDoc(doc)
		}

		// do request
		resp, err := client.DoRequestByOperationID(ctx, operationID, &openapi3_util.RequestParams{
			HeaderParams: headerParams,
			PathParams:   pathParams,
			QueryParams:  queryParams,
			BodyParams:   bodyParams,
		})
		if err != nil {
			return nil, err
		}

		// resp
		if respStr, ok := resp.(string); ok {
			return protocolCallToolResultText(respStr, false), nil
		}
		b, err := json.Marshal(resp)
		if err != nil {
			return protocolCallToolResultText(fmt.Sprintf("marshal resp err: %v", err), true), nil
		}
		return protocolCallToolResultText(string(b), false), nil
	}
}

func protocolCallToolResultText(text string, isError bool) *protocol.CallToolResult {
	return &protocol.CallToolResult{
		Content: []protocol.Content{
			&protocol.TextContent{
				Type: "text",
				Text: text,
			},
		},
		IsError: isError,
	}
}
