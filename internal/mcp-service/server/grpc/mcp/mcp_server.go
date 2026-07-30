package mcp

import (
	"context"
	"net/url"
	"strings"

	app_service "github.com/UnicomAI/wanwu/api/proto/app-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/config"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	exampleTemplate = "{\n  \"mcpServers\": {\n    \"server-name\": {\n      \"url\": \"{{url}}\"\n    }\n  }\n}"
)

func (s *Service) CreateMCPServer(ctx context.Context, req *mcp_service.CreateMCPServerReq) (*mcp_service.CreateMCPServerResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCreateMCPServerErr, toErrStatus("mcp_create_mcp_server_err", "identity is empty"))
	}
	var mcpServer *model.MCPServer
	mcpServerId := util.GenUUID()
	mcpServer = &model.MCPServer{
		MCPServerID: mcpServerId,
		Name:        req.Name,
		Description: req.Desc,
		AvatarPath:  req.AvatarPath,
		UserID:      req.Identity.UserId,
		OrgID:       req.Identity.OrgId,
	}
	err := s.cli.CreateMCPServer(ctx, mcpServer)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCreateMCPServerErr, err)
	}
	return &mcp_service.CreateMCPServerResp{McpServerId: mcpServerId}, nil
}

func (s *Service) UpdateMCPServer(ctx context.Context, req *mcp_service.UpdateMCPServerReq) (*emptypb.Empty, error) {
	mcpServer := &model.MCPServer{
		MCPServerID: req.McpServerId,
		Name:        req.Name,
		Description: req.Desc,
		AvatarPath:  req.AvatarPath,
	}
	err := s.cli.UpdateMCPServer(ctx, mcpServer)
	if err != nil {
		return nil, errStatus(errs.Code_MCPUpdateMCPServerErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetMCPServer(ctx context.Context, req *mcp_service.GetMCPServerReq) (*mcp_service.MCPServerInfo, error) {
	info, err := s.cli.GetMCPServer(ctx, req.McpServerId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetMCPServerInfoErr, err)
	}
	sseUrl, sseExample, streamableUrl, streamableExample := getMCPServerExample(ctx, req.McpServerId)
	return &mcp_service.MCPServerInfo{
		Name:              info.Name,
		McpServerId:       info.MCPServerID,
		Desc:              info.Description,
		AvatarPath:        info.AvatarPath,
		SseUrl:            sseUrl,
		SseExample:        sseExample,
		StreamableUrl:     streamableUrl,
		StreamableExample: streamableExample,
		Transport:         constant.MCPTransportSSE,
		Owner: &mcp_service.Identity{
			UserId: info.UserID,
			OrgId:  info.OrgID,
		},
		UpdatedAt: info.UpdatedAt,
		CreatedAt: info.CreatedAt,
	}, nil
}

func (s *Service) DeleteMCPServer(ctx context.Context, req *mcp_service.DeleteMCPServerReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPDeleteMCPServerErr, toErrStatus("mcp_delete_mcp_server_err", "identity is empty"))
	}
	userId := req.Identity.UserId
	orgId := req.Identity.OrgId
	err := s.cli.DeleteMCPServer(ctx, req.McpServerId, userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPDeleteMCPServerErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetMCPServerList(ctx context.Context, req *mcp_service.GetMCPServerListReq) (*mcp_service.GetMCPServerListResp, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPGetMCPServerListErr, toErrStatus("mcp_get_mcp_server_list_err", "identity is empty"))
	}
	id := req.GetIdentity()
	infos, err := s.cli.ListMCPServers(ctx, id.GetOrgId(), id.GetUserId(), req.Name)
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetMCPServerListErr, err)
	}
	list := make([]*mcp_service.MCPServerInfo, 0)
	for _, info := range infos {
		toolNum, err := s.cli.CountMCPServerTools(ctx, info.MCPServerID)
		if err != nil {
			return nil, errStatus(errs.Code_MCPGetMCPServerListErr, err)
		}
		sseUrl, sseExample, streamableUrl, streamableExample := getMCPServerExample(ctx, info.MCPServerID)
		list = append(list, &mcp_service.MCPServerInfo{
			McpServerId:       info.MCPServerID,
			Name:              info.Name,
			Desc:              info.Description,
			AvatarPath:        info.AvatarPath,
			ToolNum:           toolNum,
			SseUrl:            sseUrl,
			SseExample:        sseExample,
			StreamableUrl:     streamableUrl,
			StreamableExample: streamableExample,
			Transport:         constant.MCPTransportSSE,
		})
	}
	return &mcp_service.GetMCPServerListResp{
		List: list,
	}, nil
}

// GetAdminMCPServerPageList 管理员中心跨组织查询MCP服务列表（SQL分页）
func (s *Service) GetAdminMCPServerPageList(ctx context.Context, req *mcp_service.GetAdminMCPServerPageListReq) (*mcp_service.GetAdminMCPServerPageListResp, error) {
	infos, total, err := s.cli.ListMCPServersAdmin(ctx, req.OrgIdList, req.UserIdList, req.Name, int(req.PageNo), int(req.PageSize))
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetMCPServerListErr, err)
	}
	list := make([]*mcp_service.AdminMCPServerItem, 0, len(infos))
	for _, info := range infos {
		list = append(list, &mcp_service.AdminMCPServerItem{
			McpServerId: info.MCPServerID,
			Name:        info.Name,
			Desc:        info.Description,
			AvatarPath:  info.AvatarPath,
			Owner: &mcp_service.Identity{
				UserId: info.UserID,
				OrgId:  info.OrgID,
			},
			UpdatedAt: info.UpdatedAt,
		})
	}
	return &mcp_service.GetAdminMCPServerPageListResp{
		List:  list,
		Total: total,
	}, nil
}

func (s *Service) GetMCPServerTool(ctx context.Context, req *mcp_service.GetMCPServerToolReq) (*mcp_service.MCPServerToolInfo, error) {
	info, err := s.cli.GetMCPServerTool(ctx, req.McpServerToolId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetMCPServerToolInfoErr, err)
	}
	return &mcp_service.MCPServerToolInfo{
		McpServerToolId: info.MCPServerToolId,
		McpServerId:     info.McpServerId,
		Name:            info.Name,
		Desc:            info.Description,
		Type:            info.Type,
		AppToolId:       info.AppToolId,
		AppToolName:     info.AppToolName,
		Schema:          info.Schema,
		ApiAuth: &common.ApiAuth{
			AuthType:  info.AuthType,
			AuthIn:    info.AuthIn,
			AuthName:  info.AuthName,
			AuthValue: info.AuthValue,
		},
	}, nil
}

func (s *Service) CreateMCPServerTool(ctx context.Context, req *mcp_service.CreateMCPServerToolReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPCreateMCPServerToolErr, toErrStatus("mcp_create_mcp_server_tool_err", "identity is empty"))
	}
	userId := req.Identity.UserId
	orgId := req.Identity.OrgId
	var mcpServerTools []*model.MCPServerTool
	for _, info := range req.McpServiceToolInfos {
		mcpServerTools = append(mcpServerTools, &model.MCPServerTool{
			MCPServerToolId: util.GenUUID(),
			McpServerId:     info.McpServerId,
			Name:            info.Name,
			Description:     info.Desc,
			Type:            info.Type,
			AppToolId:       info.AppToolId,
			AppToolName:     info.AppToolName,
			Schema:          info.Schema,
			AuthType:        info.ApiAuth.AuthType,
			AuthIn:          info.ApiAuth.AuthIn,
			AuthName:        info.ApiAuth.AuthName,
			AuthValue:       info.ApiAuth.AuthValue,
			UserID:          userId,
			OrgID:           orgId,
		})
	}
	err := s.cli.CreateMCPServerTool(ctx, mcpServerTools)
	if err != nil {
		return nil, errStatus(errs.Code_MCPCreateMCPServerToolErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) UpdateMCPServerTool(ctx context.Context, req *mcp_service.UpdateMCPServerToolReq) (*emptypb.Empty, error) {
	mcpServerTool := &model.MCPServerTool{
		MCPServerToolId: req.McpServerToolId,
		Name:            req.Name,
		Description:     req.Desc,
		Schema:          req.Schema,
	}
	err := s.cli.UpdateMCPServerTool(ctx, mcpServerTool)
	if err != nil {
		return nil, errStatus(errs.Code_MCPUpdateMCPServerToolErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteMCPServerTool(ctx context.Context, req *mcp_service.DeleteMCPServerToolReq) (*emptypb.Empty, error) {
	if req.GetIdentity() == nil {
		return nil, errStatus(errs.Code_MCPDeleteMCPServerToolErr, toErrStatus("mcp_delete_mcp_server_tool_err", "identity is empty"))
	}
	userId := req.Identity.UserId
	orgId := req.Identity.OrgId
	err := s.cli.DeleteMCPServerTool(ctx, req.McpServerToolId, userId, orgId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPDeleteMCPServerToolErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetMCPServerToolList(ctx context.Context, req *mcp_service.GetMCPServerToolListReq) (*mcp_service.GetMCPServerToolListResp, error) {
	infos, err := s.cli.ListMCPServerTools(ctx, req.McpServerId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetMCPServerToolListErr, err)
	}
	list := make([]*mcp_service.MCPServerToolInfo, 0)
	for _, info := range infos {
		list = append(list, &mcp_service.MCPServerToolInfo{
			McpServerToolId: info.MCPServerToolId,
			McpServerId:     info.McpServerId,
			Name:            info.Name,
			Desc:            info.Description,
			Type:            info.Type,
			AppToolId:       info.AppToolId,
			AppToolName:     info.AppToolName,
			Schema:          info.Schema,
			ApiAuth: &common.ApiAuth{
				AuthType:  info.AuthType,
				AuthIn:    info.AuthIn,
				AuthName:  info.AuthName,
				AuthValue: info.AuthValue,
			},
		})
	}
	return &mcp_service.GetMCPServerToolListResp{
		List: list,
	}, nil
}

// internal
func getMCPServerExample(ctx context.Context, mcpServerId string) (string, string, string, string) {
	appKey := "API KEY"
	appKeys, err := app.GetAppKeyList(ctx, &app_service.GetAppKeyListReq{
		AppType: constant.AppTypeMCPServer,
		AppId:   mcpServerId,
	})
	if err == nil && appKeys.Total >= 1 {
		appKey = appKeys.Info[0].AppKey
	}

	query := url.Values{}
	query.Add("key", appKey)
	sseUrl, _ := url.JoinPath(config.Cfg().Server.ApiBaseUrl, "/openapi/v1/mcp/server/sse")
	sseUrl = sseUrl + "?" + query.Encode()
	sseExample := strings.ReplaceAll(exampleTemplate, "{{url}}", sseUrl)
	streamableUrl, _ := url.JoinPath(config.Cfg().Server.ApiBaseUrl, "/openapi/v1/mcp/server/streamable")
	streamableUrl = streamableUrl + "?" + query.Encode()
	streamableExample := strings.ReplaceAll(exampleTemplate, "{{url}}", streamableUrl)
	return sseUrl, sseExample, streamableUrl, streamableExample
}
