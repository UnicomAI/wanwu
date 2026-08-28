package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

// 所有工具的创建/更新时间统一使用 2026-01-01 00:00:00 UTC（纳秒）
const toolBoxFixedTimeNs int64 = 1767225600_000_000_000

// 工具创建/更新用户：内置工具固定为 system，自定义工具固定为 user
const (
	toolBoxBuiltinUser = "system"
	toolBoxCustomUser  = "user"
)

// toolBoxSource 一份工具箱的 schema 与时间/作者/鉴权元数据
type toolBoxSource struct {
	schema       string
	createTimeNs int64
	updateTimeNs int64
	createUser   string
	updateUser   string
	apiKey       string
	apiAuth      response.ToolBoxAPIAuth
}

// GetToolBoxDetail 工具箱明细查询：按 box_id + box_type 解析 schema 摊平成 tools[]
func GetToolBoxDetail(ctx *gin.Context, userID, orgID string, req *request.ToolBoxDetailReq) (*response.ToolBoxDetail, error) {
	page, pageSize := req.Page, req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 100
	}

	if req.BoxID == "" || (req.BoxType != constant.ToolTypeBuiltIn && req.BoxType != constant.ToolTypeCustom) {
		return emptyToolBoxDetail(req.BoxID, page, pageSize), nil
	}

	src, err := fetchToolBoxSource(ctx, userID, orgID, req.BoxID, req.BoxType)
	if err != nil {
		return nil, err
	}
	tools, err := parseSchema2ToolBoxItems(ctx.Request.Context(), src)
	if err != nil {
		return nil, err
	}
	if req.ToolID != "" {
		filtered := make([]response.ToolBoxToolItem, 0, 1)
		for _, t := range tools {
			if t.ToolID == req.ToolID {
				filtered = append(filtered, t)
			}
		}
		tools = filtered
	}
	return &response.ToolBoxDetail{
		Total:      len(tools),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: 1,
		HasNext:    false,
		HasPrev:    false,
		BoxID:      req.BoxID,
		APIKey:     src.apiKey,
		APIAuth:    src.apiAuth,
		Tools:      tools,
	}, nil
}

func emptyToolBoxDetail(boxID string, page, pageSize int) *response.ToolBoxDetail {
	return &response.ToolBoxDetail{
		Page:     page,
		PageSize: pageSize,
		BoxID:    boxID,
		Tools:    []response.ToolBoxToolItem{},
	}
}

func fetchToolBoxSource(ctx *gin.Context, userID, orgID, boxID, boxType string) (toolBoxSource, error) {
	switch boxType {
	case constant.ToolTypeBuiltIn:
		resp, err := mcp.GetSquareTool(ctx.Request.Context(), &mcp_service.GetSquareToolReq{
			ToolSquareId: boxID,
			Identity: &mcp_service.Identity{
				UserId: userID,
				OrgId:  orgID,
			},
		})
		if err != nil {
			return toolBoxSource{}, err
		}
		auth := toToolBoxAPIAuth(resp.GetBuiltInTools().GetApiAuth())
		return toolBoxSource{
			schema:       resp.Schema,
			createTimeNs: toolBoxFixedTimeNs,
			updateTimeNs: toolBoxFixedTimeNs,
			createUser:   toolBoxBuiltinUser,
			updateUser:   toolBoxBuiltinUser,
			apiKey:       auth.APIKeyValue,
			apiAuth:      auth,
		}, nil
	case constant.ToolTypeCustom:
		resp, err := mcp.GetCustomToolInfo(ctx.Request.Context(), &mcp_service.GetCustomToolInfoReq{
			CustomToolId: boxID,
		})
		if err != nil {
			return toolBoxSource{}, err
		}
		auth := toToolBoxAPIAuth(resp.GetApiAuth())
		return toolBoxSource{
			schema:       resp.Schema,
			createTimeNs: toolBoxFixedTimeNs,
			updateTimeNs: toolBoxFixedTimeNs,
			createUser:   toolBoxCustomUser,
			updateUser:   toolBoxCustomUser,
			apiKey:       auth.APIKeyValue,
			apiAuth:      auth,
		}, nil
	}
	return toolBoxSource{}, nil
}

// --- internal ---

// toolBoxDocContext OpenAPI 文档级别共享上下文（同一 box 内多个 action 复用）
type toolBoxDocContext struct {
	version    string
	serverURL  string
	components any
}

// parseSchema2ToolBoxItems 将 OpenAPI schema 摊平成 tools[]
//
// openapi3 库做严格 schema 验证 + 取顶层字段；raw map 保留原始 $ref 供对接方使用。
func parseSchema2ToolBoxItems(ctx context.Context, src toolBoxSource) ([]response.ToolBoxToolItem, error) {
	if strings.TrimSpace(src.schema) == "" {
		return []response.ToolBoxToolItem{}, nil
	}
	doc, err := openapi3_util.LoadFromData(ctx, []byte(src.schema))
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("parse openapi schema: %v", err))
	}
	var raw map[string]any
	if err := yaml.Unmarshal([]byte(src.schema), &raw); err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("parse openapi schema raw: %v", err))
	}
	rawPaths, _ := raw["paths"].(map[string]any)

	serverURL := ""
	if doc != nil && len(doc.Servers) > 0 {
		serverURL = doc.Servers[0].URL
	}
	docCtx := toolBoxDocContext{
		version:    doc.Info.Version,
		serverURL:  serverURL,
		components: raw["components"],
	}

	pathKeys := make([]string, 0, len(doc.Paths))
	for p := range doc.Paths {
		pathKeys = append(pathKeys, p)
	}
	sort.Strings(pathKeys)

	tools := make([]response.ToolBoxToolItem, 0)
	usedIDs := map[string]int{}
	for _, path := range pathKeys {
		pathItem := doc.Paths[path]
		if pathItem == nil {
			continue
		}
		rawPathItem, _ := rawPaths[path].(map[string]any)
		tools = append(tools, buildToolBoxItemsForPath(pathItem, rawPathItem, path, docCtx, src, usedIDs)...)
	}
	return tools, nil
}

// buildToolBoxItemsForPath 把单个 path 下的所有 HTTP method 摊平成 tools[] 元素。
// method 来自 pathItem.Operations()（大写）；raw schema 里 method key 是小写，需要做一次大小写映射。
func buildToolBoxItemsForPath(pathItem *openapi3.PathItem, rawPathItem map[string]any, path string,
	doc toolBoxDocContext, src toolBoxSource, usedIDs map[string]int) []response.ToolBoxToolItem {
	ops := pathItem.Operations()
	methods := make([]string, 0, len(ops))
	for m := range ops {
		methods = append(methods, m)
	}
	sort.Strings(methods) // 排序保证返回稳定
	out := make([]response.ToolBoxToolItem, 0, len(methods))
	for _, method := range methods {
		rawOp, _ := rawPathItem[strings.ToLower(method)].(map[string]any)
		if rawOp == nil {
			continue
		}
		out = append(out, buildToolBoxItem(rawOp, path, method, doc, src, usedIDs))
	}
	return out
}

func buildToolBoxItem(op map[string]any, path, method string, doc toolBoxDocContext, src toolBoxSource, usedIDs map[string]int) response.ToolBoxToolItem {
	// 取 operationId，空则用 method_path 兜底
	operationID, _ := op["operationId"].(string)
	if operationID == "" {
		operationID = fmt.Sprintf("%s_%s", strings.ToLower(method), strings.Trim(path, "/"))
		operationID = strings.NewReplacer("/", "_", "{", "", "}", "").Replace(operationID)
	}
	// 同一 box 内若 operationId 重复，追加 _2 / _3 后缀
	toolID := operationID
	if n, ok := usedIDs[operationID]; ok {
		toolID = fmt.Sprintf("%s_%d", operationID, n+1)
		usedIDs[operationID] = n + 1
	} else {
		usedIDs[operationID] = 1
	}

	desc, _ := op["description"].(string)
	summary, _ := op["summary"].(string)
	description := desc
	if description == "" {
		description = summary
	}
	params, _ := op["parameters"].([]any)
	if params == nil {
		params = []any{}
	}
	tagsRaw, _ := op["tags"].([]any)
	tags := make([]string, 0, len(tagsRaw))
	for _, v := range tagsRaw {
		if s, ok := v.(string); ok {
			tags = append(tags, s)
		}
	}

	return response.ToolBoxToolItem{
		ToolID:       toolID,
		Name:         operationID,
		Description:  description,
		Status:       "enabled",
		MetadataType: "openapi",
		Metadata: response.ToolBoxMetadata{
			Version:     doc.version,
			Summary:     operationID,
			Description: description,
			ServerURL:   doc.serverURL,
			Path:        path,
			Method:      method, // 来自 pathItem.Operations()，已经是大写
			CreateTime:  src.createTimeNs,
			UpdateTime:  src.updateTimeNs,
			CreateUser:  src.createUser,
			UpdateUser:  src.updateUser,
			APISpec: response.ToolBoxAPISpec{
				Parameters:   params,
				RequestBody:  op["requestBody"],
				Responses:    toolBoxOpenapiResponsesToArray(op["responses"]),
				Components:   doc.components,
				Callbacks:    op["callbacks"],
				Security:     op["security"],
				Tags:         tags,
				ExternalDocs: op["externalDocs"],
			},
		},
		UseRule:          "",
		GlobalParameters: response.ToolBoxGlobalParams{},
		CreateTime:       src.createTimeNs,
		UpdateTime:       src.updateTimeNs,
		CreateUser:       src.createUser,
		UpdateUser:       src.updateUser,
		ExtendInfo:       map[string]any{},
		ResourceObject:   "tool",
	}
}

// toToolBoxAPIAuth 把下游 common.ApiAuthWebRequest 转成对外的 snake_case 结构
func toToolBoxAPIAuth(a *common.ApiAuthWebRequest) response.ToolBoxAPIAuth {
	if a == nil {
		return response.ToolBoxAPIAuth{}
	}
	return response.ToolBoxAPIAuth{
		AuthType:           a.AuthType,
		APIKeyHeaderPrefix: a.ApiKeyHeaderPrefix,
		APIKeyHeader:       a.ApiKeyHeader,
		APIKeyQueryParam:   a.ApiKeyQueryParam,
		APIKeyValue:        a.ApiKeyValue,
	}
}

func toolBoxOpenapiResponsesToArray(raw any) []response.ToolBoxResponseItem {
	m, _ := raw.(map[string]any)
	out := make([]response.ToolBoxResponseItem, 0, len(m))
	if m == nil {
		return out
	}
	codes := make([]string, 0, len(m))
	for code := range m {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	for _, code := range codes {
		body, _ := m[code].(map[string]any)
		if body == nil {
			continue
		}
		desc, _ := body["description"].(string)
		out = append(out, response.ToolBoxResponseItem{
			StatusCode:  code,
			Description: desc,
			Content:     body["content"],
		})
	}
	return out
}
