package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/UnicomAI/wanwu/pkg/db"

	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/config"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetSquareTool(ctx context.Context, req *mcp_service.GetSquareToolReq) (*mcp_service.SquareToolDetail, error) {
	toolCfg, exist := config.Cfg().Tool(req.ToolSquareId)
	if !exist {
		return nil, errStatus(errs.Code_MCPGetSquareToolErr, toErrStatus("mcp_get_square_tool_err", "toolSquareId not exist"))
	}

	if req.Identity == nil {
		return nil, errStatus(errs.Code_MCPGetSquareToolErr, toErrStatus("mcp_get_square_tool_err", "identity is empty"))
	}

	apiAuth := &common.ApiAuthWebRequest{}
	if toolCfg.NeedApiKeyInput {
		info, _ := s.cli.GetBuiltinTool(ctx, &model.BuiltinTool{
			ToolSquareId: req.ToolSquareId,
			OrgID:        req.Identity.OrgId,
			UserID:       req.Identity.UserId,
		})
		if info != nil {
			apiAuthJson := info.AuthJSON
			if err := json.Unmarshal([]byte(apiAuthJson), apiAuth); err != nil {
				return nil, errStatus(errs.Code_MCPGetSquareToolErr, toErrStatus("mcp_get_square_tool_err", err.Error()))
			}
			apiAuth = toBuiltinToolApiAuth(toolCfg, apiAuth.ApiKeyValue)
		}
	}
	return buildSquareToolDetail(toolCfg, apiAuth), nil
}

func (s *Service) GetSquareToolList(ctx context.Context, req *mcp_service.GetSquareToolListReq) (*mcp_service.SquareToolList, error) {
	var toolSquareInfo []*mcp_service.ToolSquareInfo
	for _, toolCfg := range config.Cfg().Tools {
		if req.Name != "" && !strings.Contains(toolCfg.Name, req.Name) {
			continue
		}
		toolSquareInfo = append(toolSquareInfo, buildSquareToolInfo(*toolCfg))
	}
	return &mcp_service.SquareToolList{Infos: toolSquareInfo}, nil
}

func (s *Service) UpsertBuiltinToolAPIKey(ctx context.Context, req *mcp_service.UpsertBuiltinToolAPIKeyReq) (*emptypb.Empty, error) {
	if req.Identity == nil {
		return nil, errStatus(errs.Code_MCPUpdateBuiltinToolErr, toErrStatus("mcp_update_builtin_tool_err", "identity is empty"))
	}
	builtToolCfg, exist := config.Cfg().Tool(req.ToolSquareId)
	if !exist {
		return nil, errStatus(errs.Code_MCPUpdateBuiltinToolErr, toErrStatus("mcp_update_builtin_tool_err", "toolSquareId not exist"))
	}
	apiAuth := toBuiltinToolApiAuth(builtToolCfg, req.ApiKey)
	apiAuthBytes, _ := json.Marshal(apiAuth)

	info, _ := s.cli.GetBuiltinTool(ctx, &model.BuiltinTool{
		ToolSquareId: req.ToolSquareId,
		OrgID:        req.Identity.OrgId,
		UserID:       req.Identity.UserId,
	})
	if info != nil {
		// update
		info.AuthJSON = db.LongText(apiAuthBytes)
		if err := s.cli.UpdateBuiltinTool(ctx, info); err != nil {
			return nil, errStatus(errs.Code_MCPUpdateBuiltinToolErr, err)
		}
		return &emptypb.Empty{}, nil
	} else {
		// create
		if err := s.cli.CreateBuiltinTool(ctx, &model.BuiltinTool{
			ToolSquareId: req.ToolSquareId,
			AuthJSON:     db.LongText(apiAuthBytes),
			UserID:       req.Identity.UserId,
			OrgID:        req.Identity.OrgId,
		}); err != nil {
			return nil, errStatus(errs.Code_MCPUpdateBuiltinToolErr, err)
		}
	}
	return &emptypb.Empty{}, nil
}

// --- internal ---

func buildSquareToolInfo(toolCfg config.ToolConfig) *mcp_service.ToolSquareInfo {
	return &mcp_service.ToolSquareInfo{
		ToolSquareId: toolCfg.ToolSquareId,
		AvatarPath:   toolCfg.AvatarPath,
		Name:         toolCfg.Name,
		Desc:         toolCfg.Desc,
		Tags:         toolCfg.Tags,
	}
}

func buildSquareToolDetail(toolCfg config.ToolConfig, apiAuth *common.ApiAuthWebRequest) *mcp_service.SquareToolDetail {
	return &mcp_service.SquareToolDetail{
		Info: buildSquareToolInfo(toolCfg),
		BuiltInTools: &mcp_service.BuiltInTools{
			NeedApiKeyInput: toolCfg.NeedApiKeyInput,
			ApiAuth:         apiAuth,
			Detail:          toolCfg.Detail,
			ActionSum:       int32(len(toolCfg.Tools)),
			Tools:           convertBuiltInTools(toolCfg.Tools),
		},
		Schema: toolCfg.Schema,
	}
}

func convertBuiltInTools(tools []*protocol.Tool) []*common.ToolAction {
	result := make([]*common.ToolAction, 0, len(tools))
	for _, tool := range tools {
		result = append(result, &common.ToolAction{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: convertBuiltInInputSchema(&tool.InputSchema),
		})
	}
	return result
}

func convertBuiltInInputSchema(schema *protocol.InputSchema) *common.ToolActionInputSchema {
	if schema == nil {
		return nil
	}

	properties := make(map[string]*common.ToolActionInputSchemaValue)

	for field, prop := range schema.Properties {
		properties[field] = &common.ToolActionInputSchemaValue{
			Type:        string(prop.Type),
			Description: prop.Description,
		}
	}

	return &common.ToolActionInputSchema{
		Type:       string(schema.Type),
		Required:   schema.Required,
		Properties: properties,
	}
}

func toBuiltinToolApiAuth(builtinToolCfg config.ToolConfig, apiKey string) *common.ApiAuthWebRequest {
	return &common.ApiAuthWebRequest{
		AuthType:           builtinToolCfg.AuthType,
		ApiKeyHeaderPrefix: builtinToolCfg.ApiKeyHeaderPrefix,
		ApiKeyHeader:       builtinToolCfg.ApiKeyHeader,
		ApiKeyQueryParam:   builtinToolCfg.ApiKeyQueryParam,
		ApiKeyValue:        apiKey,
	}
}
