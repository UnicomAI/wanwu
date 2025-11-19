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
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gin-gonic/gin"
)

func GetWorkflowToolSelect(ctx *gin.Context, userId, orgId, toolType, name string) (*response.ListResult, error) {
	switch toolType {
	case constant.ToolTypeBuiltIn:
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
		return &response.ListResult{
			List:  list,
			Total: int64(len(list)),
		}, nil
	case constant.ToolTypeCustom:
		resp, err := GetCustomToolList(ctx, userId, orgId, name)
		if err != nil {
			return nil, err
		}
		var list []response.ToolSelect4Workflow
		for _, item := range resp.List.([]response.CustomToolInfo) {
			detail, err := GetCustomTool(ctx, userId, orgId, item.CustomToolId)
			if err != nil {
				return nil, err
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
		return &response.ListResult{
			List:  list,
			Total: int64(len(list)),
		}, nil
	}
	return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_workflow_tool_select", "unsupported tool type")
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
	for _, pathItem := range doc.Paths.Map() {
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

	// Parse path parameters, query parameters, header parameters, etc.
	if operation.Parameters != nil {
		inputs = openapiParameters2ToolActionParams4Workflow(operation.Parameters)
	}

	// Parse request body
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
	// Find 200 responses first
	for statusCode, currResponseRef := range operation.Responses.Map() {
		if strings.HasPrefix(statusCode, "2") {
			responseRef = currResponseRef
			break
		}
	}
	// If there is no response starting with 2, the first available response is used
	if responseRef == nil {
		for _, currResponseRef := range operation.Responses.Map() {
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

// openapiParameterType4Workflow gets parameter type
func openapiParameterType4Workflow(param *openapi3.Parameter) string {
	if param.Schema != nil && param.Schema.Value != nil {
		return openaiSchemaType4Workflow(param.Schema.Value)
	}
	return "string"
}

// openaiSchemaType4Workflow gets the type of schema
func openaiSchemaType4Workflow(schema *openapi3.Schema) string {
	if schema.Type != nil {
		// Check for concrete types in type slices
		if schema.Type.Is("object") {
			return "object"
		} else if schema.Type.Is("array") {
			return "list"
		} else if schema.Type.Is("string") {
			return "string"
		} else if schema.Type.Is("number") {
			return "float"
		} else if schema.Type.Is("integer") {
			return "integer"
		} else if schema.Type.Is("boolean") {
			return "boolean"
		}

		// If there are multiple types, return the first one
		if len(*schema.Type) > 0 {
			return string((*schema.Type)[0])
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
