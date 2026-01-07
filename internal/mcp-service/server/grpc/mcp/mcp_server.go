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
	"github.com/UnicomAI/wanwu/pkg/db"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	exampleTemplate = "{\n  \"mcpServers\": {\n    \"server-name\": {\n      \"url\": \"{{url}}\"\n    }\n  }\n}"
)

func (s *Service) CreateMCPServer(ctx context.Context, req *mcp_service.CreateMCPServerReq) (*mcp_service.CreateMCPServerResp, error) {
	var mcpServer *model.MCPServer
	mcpServerId := util.GenUUID()
	mcpServer = &model.MCPServer{
		MCPServerID: mcpServerId,
		Name:        req.Name,
		Description: db.LongText(req.Desc),
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
		Description: db.LongText(req.Desc),
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
		Desc:              string(info.Description),
		AvatarPath:        info.AvatarPath,
		SseUrl:            sseUrl,
		SseExample:        sseExample,
		StreamableUrl:     streamableUrl,
		StreamableExample: streamableExample,
	}, nil
}

func (s *Service) DeleteMCPServer(ctx context.Context, req *mcp_service.DeleteMCPServerReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteMCPServer(ctx, req.McpServerId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPDeleteMCPServerErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) GetMCPServerList(ctx context.Context, req *mcp_service.GetMCPServerListReq) (*mcp_service.GetMCPServerListResp, error) {
	infos, err := s.cli.ListMCPServers(ctx, req.Identity.OrgId, req.Identity.UserId, req.Name)
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
			Desc:              string(info.Description),
			AvatarPath:        info.AvatarPath,
			ToolNum:           toolNum,
			SseUrl:            sseUrl,
			SseExample:        sseExample,
			StreamableUrl:     streamableUrl,
			StreamableExample: streamableExample,
		})
	}
	return &mcp_service.GetMCPServerListResp{
		List: list,
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
		Desc:            string(info.Description),
		Type:            info.Type,
		AppToolId:       info.AppToolId,
		AppToolName:     info.AppToolName,
		Schema:          string(info.Schema),
		ApiAuth: &common.ApiAuth{
			AuthType:  info.AuthType,
			AuthIn:    info.AuthIn,
			AuthName:  info.AuthName,
			AuthValue: info.AuthValue,
		},
	}, nil
}

func (s *Service) CreateMCPServerTool(ctx context.Context, req *mcp_service.CreateMCPServerToolReq) (*emptypb.Empty, error) {
	var mcpServerTools []*model.MCPServerTool
	for _, info := range req.McpServiceToolInfos {
		mcpServerTools = append(mcpServerTools, &model.MCPServerTool{
			MCPServerToolId: util.GenUUID(),
			McpServerId:     info.McpServerId,
			Name:            info.Name,
			Description:     db.LongText(info.Desc),
			Type:            info.Type,
			AppToolId:       info.AppToolId,
			AppToolName:     info.AppToolName,
			Schema:          db.LongText(info.Schema),
			AuthType:        info.ApiAuth.AuthType,
			AuthIn:          info.ApiAuth.AuthIn,
			AuthName:        info.ApiAuth.AuthName,
			AuthValue:       info.ApiAuth.AuthValue,
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
		Description:     db.LongText(req.Desc),
		Schema:          db.LongText(req.Schema),
	}
	err := s.cli.UpdateMCPServerTool(ctx, mcpServerTool)
	if err != nil {
		return nil, errStatus(errs.Code_MCPUpdateMCPServerToolErr, err)
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteMCPServerTool(ctx context.Context, req *mcp_service.DeleteMCPServerToolReq) (*emptypb.Empty, error) {
	err := s.cli.DeleteMCPServerTool(ctx, req.McpServerToolId)
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
			Desc:            string(info.Description),
			Type:            info.Type,
			AppToolId:       info.AppToolId,
			AppToolName:     info.AppToolName,
			Schema:          string(info.Schema),
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
