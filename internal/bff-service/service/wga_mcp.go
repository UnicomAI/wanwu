package service

import (
	"fmt"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/gin-gonic/gin"
)

// --- internal wga mcp ---

// checkWgaMCPConfig 校验wga MCP配置（用于更新配置）
func checkWgaMCPConfig(ctx *gin.Context, userId, orgId string, mcpList []*assistant_service.WgaConfigMcp) error {
	if len(mcpList) == 0 {
		return nil
	}

	var mcpCustomIds, mcpServerIds []string
	for _, m := range mcpList {
		switch m.McpType {
		case constant.MCPTypeMCP:
			mcpCustomIds = append(mcpCustomIds, m.McpId)
		case constant.MCPTypeMCPServer:
			mcpServerIds = append(mcpServerIds, m.McpId)
		default:
			return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("invalid mcp type: %s", m.McpType))
		}
	}

	validIds, mcpTypes, err := getValidMcpIds(ctx, mcpCustomIds, mcpServerIds)
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "mcp not found")
	}

	for _, m := range mcpList {
		// 校验 MCP 是否存在
		if !validIds[m.McpId] {
			return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("mcp not found: %s", m.McpId))
		}
		// 校验 McpType 与 ID 是否匹配
		if actualType, ok := mcpTypes[m.McpId]; !ok || actualType != m.McpType {
			return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("mcp type mismatch: %s (expected %s, got %s)", m.McpId, m.McpType, actualType))
		}
	}
	return nil
}

func buildWgaMCPOptions(ctx *gin.Context, userId, orgId string, mcpList []*assistant_service.WgaConfigMcp) ([]wga_option.Option, error) {
	if len(mcpList) == 0 {
		return nil, nil
	}

	var mcpCustomIds, mcpServerIds []string
	for _, m := range mcpList {
		switch m.McpType {
		case constant.MCPTypeMCP:
			mcpCustomIds = append(mcpCustomIds, m.McpId)
		case constant.MCPTypeMCPServer:
			mcpServerIds = append(mcpServerIds, m.McpId)
		default:
			return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("invalid mcp type: %s", m.McpType))
		}
	}

	mcpResp, err := mcp.GetMCPByMCPIdList(ctx.Request.Context(), &mcp_service.GetMCPByMCPIdListReq{
		McpIdList:       mcpCustomIds,
		McpServerIdList: mcpServerIds,
	})
	if err != nil {
		return nil, err
	}

	var opts []wga_option.Option
	for _, item := range mcpResp.Infos {
		opts = append(opts, wga_option.WithMCP(wga_option.MCP{
			Name:        item.Info.GetName(),
			URL:         util.IfElse(item.Transport == constant.MCPTransportStreamable, item.StreamableUrl, item.SseUrl),
			Description: item.Info.GetDesc(),
			ApiAuth:     toApiAuthPtr(item.GetApiAuth()),
			Headers:     item.GetHeaders(),
		}))
	}
	for _, item := range mcpResp.Servers {
		opts = append(opts, wga_option.WithMCP(wga_option.MCP{
			Name:        item.Name,
			URL:         util.IfElse(item.Transport == constant.MCPTransportStreamable, item.StreamableUrl, item.SseUrl),
			Description: item.Desc,
		}))
	}
	return opts, nil
}

// getValidMcpIds 批量获取有效的MCP ID映射
// 返回: validIds - 有效ID映射, mcpTypes - ID对应的类型映射(mcp/mcpserver), error
func getValidMcpIds(ctx *gin.Context, mcpCustomIds, mcpServerIds []string) (map[string]bool, map[string]string, error) {
	if len(mcpCustomIds) == 0 && len(mcpServerIds) == 0 {
		return make(map[string]bool), make(map[string]string), nil
	}
	mcpResp, err := mcp.GetMCPByMCPIdList(ctx.Request.Context(), &mcp_service.GetMCPByMCPIdListReq{
		McpIdList:       mcpCustomIds,
		McpServerIdList: mcpServerIds,
	})
	if err != nil {
		return nil, nil, err
	}
	validIds := make(map[string]bool)
	mcpTypes := make(map[string]string)
	for _, item := range mcpResp.Infos {
		validIds[item.McpId] = true
		mcpTypes[item.McpId] = constant.MCPTypeMCP
	}
	for _, item := range mcpResp.Servers {
		validIds[item.McpServerId] = true
		mcpTypes[item.McpServerId] = constant.MCPTypeMCPServer
	}
	return validIds, mcpTypes, nil
}
