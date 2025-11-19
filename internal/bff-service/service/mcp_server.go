package service

import (
	"bytes"
	"context"
	"io"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	mcp_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/mcp-util"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func StartMCPServer(ctx context.Context) error {
	mcpServerList, err := mcp.GetMCPServerList(ctx, &mcp_service.GetMCPServerListReq{
		Identity: &mcp_service.Identity{},
	})
	if err != nil {
		return err
	}
	for _, mcpServerInfo := range mcpServerList.List {
		mcpServerToolList, err := mcp.GetMCPServerToolList(ctx, &mcp_service.GetMCPServerToolListReq{
			McpServerId: mcpServerInfo.McpServerId,
		})
		if err != nil {
			return err
		}
		var mcpTools []*mcp_util.McpTool
		for _, tool := range mcpServerToolList.List {
			tools, err := mcp_util.CreateMcpTools(ctx, tool.Schema, util.ConvertApiAuthProto(tool.ApiAuth), []string{tool.Name})
			if err != nil {
				return err
			}
			mcpTools = append(mcpTools, tools...)
		}
		if err = mcp_util.StartMCPServer(ctx, mcpServerInfo.McpServerId); err != nil {
			return err
		}
		if err = mcp_util.RegisterMCPServerTools(mcpServerInfo.McpServerId, mcpTools); err != nil {
			return err
		}
	}
	return nil
}

func CreateMCPServer(ctx *gin.Context, userID, orgID string, req request.MCPServerCreateReq) (*response.MCPServerCreateResp, error) {
	resp, err := mcp.CreateMCPServer(ctx.Request.Context(), &mcp_service.CreateMCPServerReq{
		Name:       req.Name,
		Desc:       req.Desc,
		AvatarPath: req.Avatar.Key,
		Identity: &mcp_service.Identity{
			OrgId:  orgID,
			UserId: userID,
		},
	})
	if err != nil {
		return nil, err
	}
	_, err = app.GenApiKey(ctx.Request.Context(), &app_service.GenApiKeyReq{
		AppId:   resp.McpServerId,
		AppType: constant.AppTypeMCPServer,
		UserId:  userID,
		OrgId:   orgID,
	})
	if err != nil {
		return nil, err
	}
	err = mcp_util.StartMCPServer(ctx, resp.McpServerId)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_mcp_server_start_err", err.Error())
	}
	return &response.MCPServerCreateResp{
		MCPServerID: resp.McpServerId,
	}, err
}

func UpdateMCPServer(ctx *gin.Context, req request.MCPServerUpdateReq) error {
	_, err := mcp.UpdateMCPServer(ctx.Request.Context(), &mcp_service.UpdateMCPServerReq{
		McpServerId: req.MCPServerID,
		Name:        req.Name,
		Desc:        req.Desc,
		AvatarPath:  req.Avatar.Key,
	})
	if err != nil {
		return err
	}
	return nil
}

func GetMCPServerDetail(ctx *gin.Context, mcpServerId string) (*response.MCPServerDetail, error) {
	mcpServerInfo, err := mcp.GetMCPServer(ctx.Request.Context(), &mcp_service.GetMCPServerReq{
		McpServerId: mcpServerId,
	})
	if err != nil {
		return nil, err
	}
	mcpServerTools, err := mcp.GetMCPServerToolList(ctx.Request.Context(), &mcp_service.GetMCPServerToolListReq{
		McpServerId: mcpServerId,
	})
	if err != nil {
		return nil, err
	}
	return toMCPServerDetail(ctx, mcpServerInfo, mcpServerTools.List), nil
}

func DeleteMCPServer(ctx *gin.Context, mcpServerId string) error {
	// 删除智能体表AssistantMCPServer相关记录 [EN] Delete records related to AssistantMCPServer on the smart body table
	_, err := assistant.AssistantMCPDeleteByMCPId(ctx.Request.Context(), &assistant_service.AssistantMCPDeleteByMCPIdReq{
		McpId:   mcpServerId,
		McpType: constant.MCPTypeMCPServer,
	})
	if err != nil {
		return err
	}

	// 删除MCPServer相关 [EN] Delete MCPServer related
	_, err = app.DeleteApp(ctx.Request.Context(), &app_service.DeleteAppReq{
		AppId:   mcpServerId,
		AppType: constant.AppTypeMCPServer,
	})
	if err != nil {
		return err
	}
	_, err = mcp.DeleteMCPServer(ctx.Request.Context(), &mcp_service.DeleteMCPServerReq{
		McpServerId: mcpServerId,
	})
	if err != nil {
		return err
	}
	err = mcp_util.ShutDownMCPServer(ctx, mcpServerId)
	if err != nil {
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_mcp_server_shutdown_err", err.Error())
	}
	return nil
}

func GetMCPServerList(ctx *gin.Context, userID, orgID, name string) (*response.ListResult, error) {
	resp, err := mcp.GetMCPServerList(ctx.Request.Context(), &mcp_service.GetMCPServerListReq{
		Name: name,
		Identity: &mcp_service.Identity{
			OrgId:  orgID,
			UserId: userID,
		},
	})
	if err != nil {
		return nil, err
	}
	var list []response.MCPServerInfo
	for _, mcpServerInfo := range resp.List {
		list = append(list, toMCPServerInfo(ctx, mcpServerInfo))
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

func CreateMCPServerTool(ctx *gin.Context, req request.MCPServerToolCreateReq) error {
	var builder mcpServerToolBuilder
	switch req.Type {
	case constant.MCPServerToolTypeCustomTool:
		builder = &mcpServerCustomToolBuilder{
			customToolID: req.Id,
		}
	case constant.MCPServerToolTypeBuiltInTool:
		builder = &mcpServerBuiltInToolBuilder{
			toolSquareId: req.Id,
		}
	default:
		// TODO
	}

	return createMCPServerTool(ctx, req.MCPServerID, builder, []string{req.MethodName})
}

func UpdateMCPServerTool(ctx *gin.Context, req request.MCPServerToolUpdateReq) error {
	tool, err := mcp.GetMCPServerTool(ctx.Request.Context(), &mcp_service.GetMCPServerToolReq{
		McpServerToolId: req.MCPServerToolID,
	})
	if err != nil {
		return err
	}
	if tool.Name == req.MethodName && tool.Desc == req.Desc {
		return nil
	}

	mcpTool, err := mcp_util.CreateMcpTool(ctx.Request.Context(), tool.Schema, util.ConvertApiAuthProto(tool.ApiAuth), tool.Name)
	if err != nil {
		return err
	}
	mcpTool, err = mcpTool.Update(ctx.Request.Context(), req.MethodName, req.Desc)
	if err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}

	if _, err = mcp.UpdateMCPServerTool(ctx.Request.Context(), &mcp_service.UpdateMCPServerToolReq{
		McpServerToolId: req.MCPServerToolID,
		Name:            mcpTool.Name(),
		Desc:            mcpTool.Desc(),
		Schema:          mcpTool.Schema(),
	}); err != nil {
		return err
	}

	err = mcp_util.UnRegisterMCPServerTools(tool.McpServerId, []string{tool.Name})
	if err != nil {
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_mcp_server_unregister_tool_err", err.Error())
	}
	err = mcp_util.RegisterMCPServerTools(tool.McpServerId, []*mcp_util.McpTool{mcpTool})
	if err != nil {
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_mcp_server_register_tool_err", err.Error())
	}
	return err
}

func DeleteMCPServerTool(ctx *gin.Context, mcpServerToolId string) error {
	toolInfo, err := mcp.GetMCPServerTool(ctx.Request.Context(), &mcp_service.GetMCPServerToolReq{
		McpServerToolId: mcpServerToolId,
	})
	if err != nil {
		return err
	}
	_, err = mcp.DeleteMCPServerTool(ctx.Request.Context(), &mcp_service.DeleteMCPServerToolReq{
		McpServerToolId: mcpServerToolId,
	})
	if err != nil {
		return err
	}
	err = mcp_util.UnRegisterMCPServerTools(toolInfo.McpServerId, []string{toolInfo.Name})
	if err != nil {
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_mcp_server_unregister_tool_err", err.Error())
	}
	return nil
}

func CreateMCPServerOpenAPITool(ctx *gin.Context, userID, orgID string, req request.MCPServerOpenAPIToolCreate) error {
	return createMCPServerTool(ctx, req.MCPServerID, &mcpServerOpenapiSchemaBuilder{
		name:   req.Name,
		schema: req.Schema,
		auth:   req.ApiAuth,
	}, req.MethodNames)
}

func GetMCPServerSSE(ctx *gin.Context, mcpServerId string, key string) error {
	queryParams := ctx.Request.URL.Query()
	queryParams.Set("key", key)
	ctx.Request.URL.RawQuery = queryParams.Encode()
	if err := mcp_util.HandleSSE(mcpServerId, ctx.Writer, ctx.Request); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	return nil
}

func GetMCPServerMessage(ctx *gin.Context, mcpServerId string) error {
	var body []byte
	if cb, ok := ctx.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body != nil {
		// 调用前再次确保Body可用（防止中间件已读取） [EN] Make sure the Body is available again before calling (to prevent the middleware from reading it)
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}

	if err := mcp_util.HandleMessage(mcpServerId, ctx.Writer, ctx.Request); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	return nil
}

func GetMCPServerStreamable(ctx *gin.Context, mcpServerId string) error {
	var body []byte
	if cb, ok := ctx.Get(gin.BodyBytesKey); ok {
		if cbb, ok := cb.([]byte); ok {
			body = cbb
		}
	}
	if body != nil {
		// 调用前再次确保Body可用（防止中间件已读取） [EN] Make sure the Body is available again before calling (to prevent the middleware from reading it)
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	}
	if err := mcp_util.HandleStreamable(mcpServerId, ctx.Writer, ctx.Request); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	return nil
}

// --- internal ---

func createMCPServerTool(ctx *gin.Context, mcpServerID string, builder mcpServerToolBuilder, operationIDs []string) error {
	if !mcp_util.CheckMCPServerExist(mcpServerID) {
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_mcp_server_not_exist")
	}

	schema, auth, err := builder.GetOpenapiSchema(ctx)
	if err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}

	tools, err := mcp_util.CreateMcpTools(ctx.Request.Context(), schema, auth, operationIDs)
	if err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}

	var toolInfos []*mcp_service.MCPServerToolInfo
	for _, tool := range tools {
		toolInfos = append(toolInfos, &mcp_service.MCPServerToolInfo{
			McpServerId: mcpServerID,
			Type:        builder.MCPServerToolType(),
			AppToolId:   builder.AppID(),
			AppToolName: builder.AppName(),
			Name:        tool.Name(),
			Desc:        tool.Desc(),
			Schema:      tool.Schema(),
			ApiAuth: &common.ApiAuth{
				AuthType:  tool.Auth().Type,
				AuthIn:    tool.Auth().In,
				AuthName:  tool.Auth().Name,
				AuthValue: tool.Auth().Value,
			},
		})
	}
	if _, err = mcp.CreateMCPServerTool(ctx.Request.Context(), &mcp_service.CreateMCPServerToolReq{
		McpServerId:         mcpServerID,
		McpServiceToolInfos: toolInfos,
	}); err != nil {
		return err
	}

	if err = mcp_util.RegisterMCPServerTools(mcpServerID, tools); err != nil {
		return grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_mcp_server_register_tool_err", err.Error())
	}
	return err
}

func toMCPServerInfo(ctx *gin.Context, mcpServerInfo *mcp_service.MCPServerInfo) response.MCPServerInfo {
	return response.MCPServerInfo{
		MCPServerID: mcpServerInfo.McpServerId,
		Avatar:      cacheMCPServerAvatar(ctx, mcpServerInfo.AvatarPath),
		Name:        mcpServerInfo.Name,
		Desc:        mcpServerInfo.Desc,
		ToolNum:     mcpServerInfo.ToolNum,
	}
}

func toMCPServerDetail(ctx *gin.Context, mcpServerInfo *mcp_service.MCPServerInfo, mcpServerToolInfos []*mcp_service.MCPServerToolInfo) *response.MCPServerDetail {
	var mcpServerTools []response.MCPServerToolInfo
	for _, mcpServerToolInfo := range mcpServerToolInfos {
		mcpServerTools = append(mcpServerTools, response.MCPServerToolInfo{
			MCPServerToolID: mcpServerToolInfo.McpServerToolId,
			MethodName:      mcpServerToolInfo.Name,
			Type:            mcpServerToolInfo.Type,
			Id:              mcpServerToolInfo.AppToolId,
			Name:            mcpServerToolInfo.AppToolName,
			Desc:            mcpServerToolInfo.Desc,
		})
	}
	return &response.MCPServerDetail{
		MCPServerID:       mcpServerInfo.McpServerId,
		Avatar:            cacheMCPServerAvatar(ctx, mcpServerInfo.AvatarPath),
		Name:              mcpServerInfo.Name,
		Desc:              mcpServerInfo.Desc,
		SSEURL:            mcpServerInfo.SseUrl,
		SSEExample:        mcpServerInfo.SseExample,
		StreamableURL:     mcpServerInfo.StreamableUrl,
		StreamableExample: mcpServerInfo.StreamableExample,
		Tools:             mcpServerTools,
	}
}
