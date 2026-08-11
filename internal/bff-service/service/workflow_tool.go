package service

import (
	"context"
	"fmt"
	net_url "net/url"
	"strings"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

func GetWorkflowToolSelect(ctx *gin.Context, userId, orgId, toolType, name string) (*response.ListResult, error) {
	switch toolType {
	case constant.ToolTypeBuiltIn:
		list, err := getBuiltInToolSelect4Workflow(ctx, userId, orgId, name)
		if err != nil {
			return nil, err
		}
		return &response.ListResult{List: list, Total: int64(len(list))}, nil
	case constant.ToolTypeCustom:
		list, err := getCustomToolSelect4Workflow(ctx, userId, orgId, name)
		if err != nil {
			return nil, err
		}
		return &response.ListResult{List: list, Total: int64(len(list))}, nil
	case "":
		builtInList, err := getBuiltInToolSelect4Workflow(ctx, userId, orgId, name)
		if err != nil {
			return nil, err
		}
		customList, err := getCustomToolSelect4Workflow(ctx, userId, orgId, name)
		if err != nil {
			return nil, err
		}
		list := append(builtInList, customList...)
		return &response.ListResult{List: list, Total: int64(len(list))}, nil
	}
	return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_tool_select", "unsupported tool type")
}

func getBuiltInToolSelect4Workflow(ctx *gin.Context, userId, orgId, name string) ([]response.ToolSelect4Workflow, error) {
	resp, err := mcp.GetSquareToolList(ctx.Request.Context(), &mcp_service.GetSquareToolListReq{
		Name: name,
	})
	if err != nil {
		return nil, err
	}
	var list []response.ToolSelect4Workflow
	for _, item := range resp.Infos {
		detail, err := GetToolSquareDetail(ctx, userId, orgId, item.ToolSquareId)
		if err != nil {
			return nil, err
		}
		url, _ := net_url.JoinPath(config.Cfg().Server.ApiBaseUrl, detail.Avatar.Path)
		list = append(list, response.ToolSelect4Workflow{
			ToolID:   item.ToolSquareId,
			ToolName: item.Name,
			ToolType: constant.ToolTypeBuiltIn,
			IconUrl:  url,
			ApiKey:   detail.APIKey,
			Desc:     detail.Desc,
			Actions:  builtInToolActions4Workflow(detail.Tools),
		})
	}
	return list, nil
}

func getCustomToolSelect4Workflow(ctx *gin.Context, userId, orgId, name string) ([]response.ToolSelect4Workflow, error) {
	resp, err := GetCustomToolList(ctx, userId, orgId, name)
	if err != nil {
		return nil, err
	}
	var list []response.ToolSelect4Workflow
	for _, item := range resp.List.([]response.CustomToolInfo) {
		detail, err := GetCustomTool(ctx, userId, orgId, item.CustomToolId)
		if err != nil {
			log.Errorf("skip custom tool %s(%s) for workflow select: %v", item.Name, item.CustomToolId, err)
			continue
		}
		url, _ := net_url.JoinPath(config.Cfg().Server.ApiBaseUrl, detail.Avatar.Path)
		list = append(list, response.ToolSelect4Workflow{
			ToolID:   item.CustomToolId,
			ToolName: item.Name,
			ToolType: constant.ToolTypeCustom,
			IconUrl:  url,
			ApiKey:   "",
			Desc:     detail.Description,
			Actions:  customToolActions4Workflow(detail.ApiList),
		})
	}
	return list, nil
}

func GetWorkflowToolDetail(ctx *gin.Context, userId, orgId, toolId, toolType, operationId string) (*response.ToolDetail4Workflow, error) {
	var schema, iconUrl string
	switch toolType {
	case constant.ToolTypeBuiltIn:
		resp, err := mcp.GetSquareTool(ctx.Request.Context(), &mcp_service.GetSquareToolReq{
			ToolSquareId: toolId,
			Identity: &mcp_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if err != nil {
			return nil, err
		}
		schema = resp.Schema
		iconUrl, _ = net_url.JoinPath(config.Cfg().Server.ApiBaseUrl, cacheToolAvatar(ctx, constant.ToolTypeBuiltIn, resp.Info.AvatarPath).Path)
	case constant.ToolTypeCustom:
		resp, err := mcp.GetCustomToolInfo(ctx.Request.Context(), &mcp_service.GetCustomToolInfoReq{
			CustomToolId: toolId,
			Identity: &mcp_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if err != nil {
			return nil, err
		}
		schema = resp.Schema
		iconUrl, _ = net_url.JoinPath(config.Cfg().Server.ApiBaseUrl, cacheToolAvatar(ctx, constant.ToolTypeCustom, resp.AvatarPath).Path)
	}

	inputs, outputs, err := openapiSchema2ToolActionInputsAndOutputs4Workflow(ctx.Request.Context(), schema, operationId)
	if err != nil {
		return nil, err
	}
	return &response.ToolDetail4Workflow{
		ActionID:   operationId,
		ActionName: operationId,
		IconUrl:    iconUrl,
		Inputs:     inputs,
		Outputs:    outputs,
	}, nil
}

// --- internal ---

func builtInToolActions4Workflow(actions []*protocol.Tool) []response.ToolAction4Workflow {
	var ret []response.ToolAction4Workflow
	for _, action := range actions {
		ret = append(ret, response.ToolAction4Workflow{
			ActionName: action.Name,
			ActionID:   action.Name,
			Desc:       action.Description,
		})
	}
	return ret
}

func customToolActions4Workflow(actions []response.CustomToolActionInfo) []response.ToolAction4Workflow {
	var ret []response.ToolAction4Workflow
	for _, action := range actions {
		ret = append(ret, response.ToolAction4Workflow{
			ActionName: action.Name,
			ActionID:   action.Name,
			Desc:       action.Desc,
		})
	}
	return ret
}

func openapiSchema2ToolActionInputsAndOutputs4Workflow(ctx context.Context, schema, operationID string) ([]interface{}, []interface{}, error) {
	doc, err := openapi3_util.LoadFromData(ctx, []byte(schema))
	if err != nil {
		return nil, nil, err
	}
	var exist bool
	var inputs, outputs []interface{}
	for _, pathItem := range doc.Paths {
		for _, operation := range pathItem.Operations() {
			if operation.OperationID != operationID {
				continue
			}
			exist = true
			inputs = openapiOperation2ToolActionInputs4Workflow(operation)
			outputs = openapiOperation2ToolActionOutputs4Workflow(operation)
			break
		}
	}
	if !exist {
		return nil, nil, fmt.Errorf("opentionID(%v) not found", operationID)
	}
	return inputs, outputs, nil
}

func openapiOperation2ToolActionInputs4Workflow(operation *openapi3.Operation) []interface{} {
	inputs := []interface{}{}

	// 解析路径参数、查询参数、header 参数等
	if operation.Parameters != nil {
		inputs = openapiParameters2ToolActionParams4Workflow(operation.Parameters)
	}

	// 解析请求体
	if operation.RequestBody != nil && operation.RequestBody.Value != nil {
		for _, mediaType := range operation.RequestBody.Value.Content {
			if mediaType.Schema != nil && mediaType.Schema.Value != nil {
				inputs = append(inputs, openapiSchemaProperties2ToolActionParams4Workflow(mediaType.Schema.Value.Properties, mediaType.Schema.Value.Required)...)
			}
		}
	}
	return inputs
}

func openapiOperation2ToolActionOutputs4Workflow(operation *openapi3.Operation) []interface{} {
	outputs := []interface{}{}
	if operation.Responses == nil {
		return outputs
	}

	var responseRef *openapi3.ResponseRef
	// 优先查找 200 响应
	for statusCode, currResponseRef := range operation.Responses {
		if strings.HasPrefix(statusCode, "2") {
			responseRef = currResponseRef
			break
		}
	}
	// 如果没有2开头的响应，使用第一个可用的响应
	if responseRef == nil {
		for _, currResponseRef := range operation.Responses {
			responseRef = currResponseRef
			break
		}
	}

	if responseRef != nil && responseRef.Value != nil {
		for _, mediaType := range responseRef.Value.Content {
			if mediaType.Schema != nil && mediaType.Schema.Value != nil {
				outputs = append(outputs, openapiSchemaProperties2ToolActionParams4Workflow(mediaType.Schema.Value.Properties, mediaType.Schema.Value.Required)...)
				break
			}
		}
	}
	return outputs
}

func openapiParameters2ToolActionParams4Workflow(parameters openapi3.Parameters) []interface{} {
	rets := []interface{}{}
	if parameters == nil {
		return rets
	}

	for _, param := range parameters {
		if param.Value == nil {
			continue
		}
		var ret interface{}
		propType := openapiParameterType4Workflow(param.Value)
		switch propType {
		case "list":
			prop := response.ToolActionParamWithTypeList4Workflow{
				Type:        propType,
				Name:        param.Value.In + "-" + param.Value.Name,
				Description: param.Value.Description,
				Required:    param.Value.Required,
				Schema:      response.ToolActionParamInTypeList4Workflow{},
			}
			if param.Value.Schema != nil && param.Value.Schema.Value != nil && param.Value.Schema.Value.Items != nil {
				prop.Schema.Type = openaiSchemaType4Workflow(param.Value.Schema.Value.Items.Value)
				prop.Schema.Children = []interface{}{}
			}
			ret = prop
		default:
			prop := response.ToolActionParamWithoutTypeList4Workflow{
				Type:        propType,
				Name:        param.Value.In + "-" + param.Value.Name,
				Description: param.Value.Description,
				Required:    param.Value.Required,
				Children:    []interface{}{},
			}
			if param.Value.Schema != nil && param.Value.Schema.Value != nil {
				prop.Children = openapiSchemaProperties2ToolActionParams4Workflow(param.Value.Schema.Value.Properties, param.Value.Schema.Value.Required)
			}
			ret = prop
		}
		rets = append(rets, ret)
	}
	return rets
}

func openapiSchemaProperties2ToolActionParams4Workflow(properties openapi3.Schemas, required []string) []interface{} {
	rets := []interface{}{}
	if properties == nil {
		return rets
	}

	for propName, propSchema := range properties {
		if propSchema.Value == nil {
			continue
		}
		var ret interface{}
		propType := openaiSchemaType4Workflow(propSchema.Value)
		switch propType {
		case "list":
			prop := response.ToolActionParamWithTypeList4Workflow{
				Type:        propType,
				Name:        propName,
				Description: propSchema.Value.Description,
				Required:    util.Exist(required, propName),
				Schema:      response.ToolActionParamInTypeList4Workflow{},
			}
			if propSchema.Value.Items != nil && propSchema.Value.Items.Value != nil {
				prop.Schema.Type = openaiSchemaType4Workflow(propSchema.Value.Items.Value)
				prop.Schema.Children = openapiSchemaProperties2ToolActionParams4Workflow(propSchema.Value.Items.Value.Properties, propSchema.Value.Required)
			}
			ret = prop
		default:
			prop := response.ToolActionParamWithoutTypeList4Workflow{
				Type:        propType,
				Name:        propName,
				Description: propSchema.Value.Description,
				Required:    util.Exist(required, propName),
				Children:    []interface{}{},
			}
			prop.Children = openapiSchemaProperties2ToolActionParams4Workflow(propSchema.Value.Properties, propSchema.Value.Required)
			ret = prop
		}
		rets = append(rets, ret)
	}
	return rets
}

// openapiParameterType4Workflow 获取参数类型
func openapiParameterType4Workflow(param *openapi3.Parameter) string {
	if param.Schema != nil && param.Schema.Value != nil {
		return openaiSchemaType4Workflow(param.Schema.Value)
	}
	return "string"
}

// openaiSchemaType4Workflow 获取 schema 的类型
func openaiSchemaType4Workflow(schema *openapi3.Schema) string {
	if schema.Type != "" {
		// 检查类型切片中的具体类型
		switch schema.Type {
		case openapi3.TypeObject:
			return "object"
		case openapi3.TypeArray:
			return "list"
		case openapi3.TypeString:
			return "string"
		case openapi3.TypeNumber:
			return "float"
		case openapi3.TypeInteger:
			return "integer"
		case openapi3.TypeBoolean:
			return "boolean"
		default:
			return "null"
		}
	}

	if len(schema.AnyOf) > 0 {
		return "anyOf"
	}
	if len(schema.AllOf) > 0 {
		return "allOf"
	}
	if len(schema.OneOf) > 0 {
		return "oneOf"
	}

	if schema.Format != "" {
		return schema.Format
	}

	return "string"
}
