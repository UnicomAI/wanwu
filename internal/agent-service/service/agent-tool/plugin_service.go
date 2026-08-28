package agent_tool

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	agent_http_client "github.com/UnicomAI/wanwu/internal/agent-service/pkg/http"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/getkin/kin-openapi/openapi3"
)

const (
	chatDocToolID = "chatdoc"
	callbackWanWu = "callback-wanwu"
)

// openAPITool 实现了 tool.InvokableTool 接口
type openAPITool struct {
	info    *schema.ToolInfo
	handler func(ctx context.Context, arguments string) (string, error)
}

// Info 返回工具的元信息
func (t *openAPITool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	marshal, _ := json.Marshal(t.info)
	log.Infof("openAPITool %v", string(marshal))
	return t.info, nil
}

// InvokableRun 执行工具
func (t *openAPITool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	if t.handler == nil {
		return "", fmt.Errorf("tool handler is not set")
	}
	return t.handler(ctx, argumentsInJSON)
}

func GetToolsFromOpenAPISchema(ctx context.Context, pluginToolList []*request.PluginToolInfo, changeToolName bool) ([]tool.BaseTool, bool, map[string]*request.ToolConfig, error) {
	if len(pluginToolList) == 0 {
		return nil, false, nil, nil
	}
	var allTools []tool.BaseTool
	var toolIDMap = make(map[string]*request.ToolConfig)
	var hasChatDoc = false

	for _, wrapper := range pluginToolList {
		if wrapper.APISchema == nil {
			continue
		}

		loader := openapi3.NewLoader()
		loader.IsExternalRefsAllowed = true

		if err := wrapper.APISchema.Validate(ctx, openapi3.EnableExamplesValidation()); err != nil {
			log.Errorf("Warning: OpenAPI schema validation failed: %v", err)
		}

		info := wrapper.APISchema.Info
		var apiTitle = ""
		if info != nil {
			apiTitle = info.Title
		}

		for path, pathItem := range wrapper.APISchema.Paths {
			operations := map[string]*openapi3.Operation{
				"get":    pathItem.Get,
				"post":   pathItem.Post,
				"put":    pathItem.Put,
				"delete": pathItem.Delete,
				"patch":  pathItem.Patch,
			}

			for method, operation := range operations {
				if operation == nil {
					continue
				}
				serverURL := ""
				if len(wrapper.APISchema.Servers) > 0 {
					serverURL = wrapper.APISchema.Servers[0].URL
				}

				//内置，chatDoc 单独处理
				if chatDocToolID == operation.OperationID && strings.Contains(serverURL, callbackWanWu) {
					hasChatDoc = true
					continue
				}
				einoTool := openapi3_util.Operation2EinoTool(operation)
				if einoTool.Name == "" {
					einoTool.Name = fmt.Sprintf("%s_%s", method, path)
				}
				//处理模型functionName 不能包含中文等特殊字符的问题
				var toolID = einoTool.Name
				var toolName = einoTool.Name
				if changeToolName {
					toolID = util.MD5(einoTool.Name)
					einoTool.Name = toolID
				}
				toolIDMap[toolID] = &request.ToolConfig{
					ToolID:   toolID,
					ToolName: toolName,
					Avatar:   wrapper.ToolAvatar,
				}

				if len(apiTitle) > 0 {
					einoTool.Desc = fmt.Sprintf("%s,%s", apiTitle, einoTool.Desc)
				}

				contentType := getRequestContentType(operation)
				handler := createHTTPHandler(serverURL, path, method, wrapper.APIAuth, contentType, toolName)

				tools := &openAPITool{
					info:    einoTool,
					handler: handler,
				}

				allTools = append(allTools, tools)
			}
		}
	}

	return allTools, hasChatDoc, toolIDMap, nil
}

func GetEnioToolsFromOpenAPISchema(ctx context.Context, pluginTool *request.PluginToolInfo) ([]*schema.ToolInfo, error) {
	var allTools []*schema.ToolInfo

	if pluginTool.APISchema == nil {
		log.Errorf("GetEnioToolsFromOpenAPISchema is nil")
		return nil, nil
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	if err := pluginTool.APISchema.Validate(ctx, openapi3.EnableExamplesValidation()); err != nil {
		log.Errorf("Warning: OpenAPI schema validation failed: %v", err)
	}

	for _, pathItem := range pluginTool.APISchema.Paths {
		operations := map[string]*openapi3.Operation{
			"get":    pathItem.Get,
			"post":   pathItem.Post,
			"put":    pathItem.Put,
			"delete": pathItem.Delete,
			"patch":  pathItem.Patch,
		}

		for _, operation := range operations {
			if operation == nil {
				continue
			}

			einoTool := openapi3_util.Operation2EinoTool(operation)
			allTools = append(allTools, einoTool)
		}
	}

	return allTools, nil
}

func getRequestContentType(operation *openapi3.Operation) string {
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		for contentType := range operation.RequestBody.Value.Content {
			return contentType
		}
	}
	return "application/json"
}

func createHTTPHandler(serverURL, path, method string, auth *openapi3_util.Auth, contentType, toolName string) func(ctx context.Context, arguments string) (string, error) {
	return func(ctx context.Context, arguments string) (string, error) {
		start := time.Now().UnixMilli()
		requestURL := serverURL + path

		var body io.Reader
		var actualContentType string

		// 解析 URL 以便添加查询参数(包括认证参数)
		parsedURL, err := url.Parse(requestURL)
		if err != nil {
			return "", fmt.Errorf("failed to parse URL: %w", err)
		}

		// 获取现有的查询参数
		queryValues := parsedURL.Query()

		// 如果认证方式是 query 参数,先添加认证参数
		if auth != nil && auth.Type == "apiKey" && auth.In == "query" {
			queryValues.Set(auth.Name, auth.Value)
		}

		// 处理 GET 请求的查询参数
		if method == "get" && arguments != "" {
			var params map[string]interface{}
			if err := json.Unmarshal([]byte(arguments), &params); err != nil {
				return "", fmt.Errorf("failed to parse arguments: %w", err)
			}

			// 添加业务查询参数
			for key, value := range params {
				if queryValues.Has(key) {
					log.Infof("requestURL %s query parameter %s already exists, overwriting with value %v", requestURL, key, value)
					continue
				}
				queryValues.Set(key, fmt.Sprintf("%v", value))
			}
		}

		// 更新 URL 的查询参数
		parsedURL.RawQuery = queryValues.Encode()
		requestURL = parsedURL.String()

		if method == "post" || method == "put" || method == "patch" {
			if contentType == "multipart/form-data" {
				var params map[string]interface{}
				if err := json.Unmarshal([]byte(arguments), &params); err != nil {
					return "", fmt.Errorf("failed to parse arguments: %w", err)
				}

				bodyBuf := &bytes.Buffer{}
				writer := multipart.NewWriter(bodyBuf)

				for key, value := range params {
					var valueStr string
					switch v := value.(type) {
					case string:
						valueStr = v
					default:
						valueBytes, _ := json.Marshal(v)
						valueStr = string(valueBytes)
					}

					if err := writer.WriteField(key, valueStr); err != nil {
						return "", fmt.Errorf("failed to write field %s: %w", key, err)
					}
				}

				if err := writer.Close(); err != nil {
					return "", fmt.Errorf("failed to close multipart writer: %w", err)
				}

				body = bodyBuf
				actualContentType = writer.FormDataContentType()
			} else {
				body = bytes.NewBufferString(arguments)
				actualContentType = "application/json"
			}
		}

		methodUpper := http.MethodPost
		switch method {
		case "get":
			methodUpper = http.MethodGet
		case "post":
			methodUpper = http.MethodPost
		case "put":
			methodUpper = http.MethodPut
		case "delete":
			methodUpper = http.MethodDelete
		case "patch":
			methodUpper = http.MethodPatch
		}

		req, err := http.NewRequestWithContext(ctx, methodUpper, requestURL, body)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		if auth != nil && auth.Type == "apiKey" && auth.In == "header" {
			req.Header.Set(auth.Name, auth.Value)
		}
		if body != nil {
			req.Header.Set("Content-Type", actualContentType)
		}

		resp, err := agent_http_client.GetClient().Client.Do(req)
		respBody, err := buildResult(resp, err)
		http_client.LogHttpRequest(ctx, "request_tool_call_"+toolName, method, requestURL, arguments, respBody, err, start)
		if err != nil {
			return response.ToolErrResp(err)
		}

		return respBody, nil
	}
}

func buildResult(resp *http.Response, err error) (string, error) {
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err1 := Body.Close()
		if err1 != nil {
			log.Infof("failed to close response body: %v", err1)
		}
	}(resp.Body) // 确保关闭响应体

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBody))
	}
	return string(respBody), nil
}
