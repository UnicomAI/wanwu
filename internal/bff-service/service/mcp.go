package service

import (
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	mcp_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/mcp-util"
	bff_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/gin-gonic/gin"
)

func GetMCPSquareDetail(ctx *gin.Context, userID, orgID, mcpSquareID string) (*response.MCPSquareDetail, error) {
	mcpSquare, err := mcp.GetSquareMCP(ctx.Request.Context(), &mcp_service.GetSquareMCPReq{
		OrgId:       orgID,
		UserId:      userID,
		McpSquareId: mcpSquareID,
	})
	if err != nil {
		return nil, err
	}
	return toMCPSquareDetail(ctx, mcpSquare), nil
}

func GetMCPSquareList(ctx *gin.Context, userID, orgID, category, name string) (*response.ListResult, error) {
	resp, err := mcp.GetSquareMCPList(ctx.Request.Context(), &mcp_service.GetSquareMCPListReq{
		OrgId:    orgID,
		UserId:   userID,
		Category: category,
		Name:     name,
	})
	if err != nil {
		return nil, err
	}
	var list []response.MCPSquareInfo
	for _, mcpSquare := range resp.Infos {
		list = append(list, toMCPSquareInfo(ctx, mcpSquare, ""))
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

func CreateMCP(ctx *gin.Context, userID, orgID string, req request.MCPCreate) error {
	_, err := mcp.CreateCustomMCP(ctx.Request.Context(), &mcp_service.CreateCustomMCPReq{
		OrgId:       orgID,
		UserId:      userID,
		McpSquareId: req.MCPSquareID,
		Name:        req.Name,
		Desc:        req.Desc,
		From:        req.From,
		SseUrl:      req.SSEURL,
		AvatarPath:  req.Avatar.Key,
	})
	return err
}

func UpdateMCP(ctx *gin.Context, userID, orgID string, req request.MCPUpdate) error {
	_, err := mcp.UpdateCustomMCP(ctx.Request.Context(), &mcp_service.UpdateCustomMCPReq{
		OrgId:      orgID,
		UserId:     userID,
		McpId:      req.MCPID,
		Name:       req.Name,
		Desc:       req.Desc,
		From:       req.From,
		SseUrl:     req.SSEURL,
		AvatarPath: req.Avatar.Key,
	})
	return err
}

func GetMCP(ctx *gin.Context, mcpID string) (*response.MCPDetail, error) {
	mcpDetail, err := mcp.GetCustomMCP(ctx.Request.Context(), &mcp_service.GetCustomMCPReq{
		McpId: mcpID,
	})
	if err != nil {
		return nil, err
	}
	return toMCPCustomDetail(ctx, mcpDetail), nil
}

func DeleteMCP(ctx *gin.Context, mcpID string) error {
	// Delete records related to AssistantMCP on the smart body surface
	_, err := assistant.AssistantMCPDeleteByMCPId(ctx.Request.Context(), &assistant_service.AssistantMCPDeleteByMCPIdReq{
		McpId:   mcpID,
		McpType: constant.MCPTypeMCP,
	})
	if err != nil {
		return err
	}

	_, err = mcp.DeleteCustomMCP(ctx.Request.Context(), &mcp_service.DeleteCustomMCPReq{
		McpId: mcpID,
	})
	return err
}

func GetMCPList(ctx *gin.Context, userID, orgID, name string) (*response.ListResult, error) {
	resp, err := mcp.GetCustomMCPList(ctx.Request.Context(), &mcp_service.GetCustomMCPListReq{
		OrgId:  orgID,
		UserId: userID,
		Name:   name,
	})
	if err != nil {
		return nil, err
	}
	var list []response.MCPInfo
	for _, mcpInfo := range resp.Infos {
		list = append(list, toMCPCustomInfo(ctx, mcpInfo))
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

func GetMCPSelect(ctx *gin.Context, userID, orgID string, name string) (*response.ListResult, error) {
	// Get custom mcp list
	resp, err := mcp.GetCustomMCPList(ctx.Request.Context(), &mcp_service.GetCustomMCPListReq{
		OrgId:  orgID,
		UserId: userID,
		Name:   name,
	})
	if err != nil {
		return nil, err
	}
	var list []response.MCPSelect
	for _, mcpInfo := range resp.Infos {
		list = append(list, response.MCPSelect{
			UniqueId: bff_util.ConcatAssistantToolUniqueId("mcp", mcpInfo.McpId),
			// Compatible with older versions
			MCPID:       mcpInfo.McpId,
			MCPSquareID: mcpInfo.Info.McpSquareId,
			Name:        mcpInfo.Info.Name,
			// Suitable for agent mcp dropdown
			ToolId:   mcpInfo.McpId,
			ToolName: mcpInfo.Info.Name,
			ToolType: constant.MCPTypeMCP,
			// Common fields
			Description: mcpInfo.Info.Desc,
			ServerFrom:  mcpInfo.Info.From,
			ServerURL:   mcpInfo.SseUrl,
			Type:        constant.MCPTypeMCP,
			Avatar:      cacheMCPAvatar(ctx, mcpInfo.Info.AvatarPath, mcpInfo.AvatarPath),
		})
	}

	// Get mcp server list
	mcpServerList, err := mcp.GetMCPServerList(ctx.Request.Context(), &mcp_service.GetMCPServerListReq{
		Name: name,
		Identity: &mcp_service.Identity{
			OrgId:  orgID,
			UserId: userID,
		},
	})
	if err != nil {
		return nil, err
	}
	for _, mcpServerInfo := range mcpServerList.List {
		list = append(list, response.MCPSelect{
			MCPID:       mcpServerInfo.McpServerId,
			MCPSquareID: "",
			UniqueId:    bff_util.ConcatAssistantToolUniqueId(constant.AppTypeMCPServer, mcpServerInfo.McpServerId),
			Name:        mcpServerInfo.Name,
			Description: mcpServerInfo.Desc,
			ServerFrom:  "mcp server",
			ServerURL:   mcpServerInfo.SseUrl,
			Type:        constant.MCPTypeMCPServer,
			// Suitable for agent mcp dropdown
			ToolId:   mcpServerInfo.McpServerId,
			ToolName: mcpServerInfo.Name,
			ToolType: constant.MCPTypeMCPServer,
			Avatar:   cacheMCPServerAvatar(ctx, mcpServerInfo.AvatarPath),
		})
	}
	return &response.ListResult{
		List:  list,
		Total: int64(len(list)),
	}, nil
}

func GetMCPToolList(ctx *gin.Context, mcpID, sseUrl string) (*response.MCPToolList, error) {
	if mcpID != "" {
		mcpDetail, err := mcp.GetCustomMCP(ctx.Request.Context(), &mcp_service.GetCustomMCPReq{
			McpId: mcpID,
		})
		if err != nil {
			return nil, err
		}
		if mcpDetail.SseUrl != "" {
			sseUrl = mcpDetail.SseUrl
		}
	}
	if sseUrl == "" {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "sseUrl empty")
	}

	tools, err := mcp_util.ListTools(ctx.Request.Context(), sseUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
	}
	return &response.MCPToolList{Tools: tools}, nil
}

func GetMCPActionList(ctx *gin.Context, userID, orgID string, req request.MCPActionListReq) (*response.MCPActionList, error) {
	var actions []*protocol.Tool
	switch req.ToolType {
	case constant.MCPTypeMCPServer:
		mcpServerList, err := mcp.GetMCPServerToolList(ctx.Request.Context(), &mcp_service.GetMCPServerToolListReq{
			McpServerId: req.ToolId,
		})
		if err != nil {
			return nil, err
		}
		for _, tool := range mcpServerList.List {
			toolActions, err := openapi3_util.Schema2ProtocolTools(ctx.Request.Context(), []byte(tool.Schema))
			if err != nil {
				return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
			}
			actions = append(actions, toolActions...)
		}
	case constant.MCPTypeMCP:
		tools, err := GetMCPToolList(ctx, req.ToolId, "")
		if err != nil {
			return nil, err
		}
		actions = tools.Tools
	default:
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "invalid toolType")
	}
	return &response.MCPActionList{
		Actions: actions,
	}, nil
}

// --- internal ---

func toMCPCustomDetail(ctx *gin.Context, mcpDetail *mcp_service.CustomMCPDetail) *response.MCPDetail {
	return &response.MCPDetail{
		MCPInfo: response.MCPInfo{
			MCPID:         mcpDetail.McpId,
			SSEURL:        mcpDetail.SseUrl,
			MCPSquareInfo: toMCPSquareInfo(ctx, mcpDetail.Info, mcpDetail.AvatarPath),
		},
		MCPSquareIntro: toMCPSquareIntro(mcpDetail.Intro),
	}
}

func toMCPCustomInfo(ctx *gin.Context, mcpInfo *mcp_service.CustomMCPInfo) response.MCPInfo {
	return response.MCPInfo{
		MCPID:         mcpInfo.McpId,
		SSEURL:        mcpInfo.SseUrl,
		MCPSquareInfo: toMCPSquareInfo(ctx, mcpInfo.Info, mcpInfo.AvatarPath),
	}
}

func toMCPSquareDetail(ctx *gin.Context, mcpSquare *mcp_service.SquareMCPDetail) *response.MCPSquareDetail {
	ret := &response.MCPSquareDetail{
		MCPSquareInfo:  toMCPSquareInfo(ctx, mcpSquare.Info, ""),
		MCPSquareIntro: toMCPSquareIntro(mcpSquare.Intro),
		MCPActions: response.MCPActions{
			SSEURL:    mcpSquare.Tool.SseUrl,
			HasCustom: mcpSquare.Tool.HasCustom,
		},
	}
	for _, tool := range mcpSquare.Tool.Tools {
		ret.MCPActions.Tools = append(ret.MCPActions.Tools, toToolAction(tool))
	}
	return ret
}

func toMCPSquareInfo(ctx *gin.Context, mcpSquareInfo *mcp_service.SquareMCPInfo, customAvatarPath string) response.MCPSquareInfo {
	return response.MCPSquareInfo{
		MCPSquareID: mcpSquareInfo.McpSquareId,
		Avatar:      cacheMCPAvatar(ctx, mcpSquareInfo.AvatarPath, customAvatarPath),
		Name:        mcpSquareInfo.Name,
		Desc:        mcpSquareInfo.Desc,
		From:        mcpSquareInfo.From,
		Category:    mcpSquareInfo.Category,
	}
}

func toMCPSquareIntro(mcpSquareIntro *mcp_service.SquareMCPIntro) response.MCPSquareIntro {
	if mcpSquareIntro == nil {
		return response.MCPSquareIntro{}
	}
	return response.MCPSquareIntro{
		Summary:  mcpSquareIntro.Summary,
		Feature:  mcpSquareIntro.Feature,
		Scenario: mcpSquareIntro.Scenario,
		Manual:   mcpSquareIntro.Manual,
		Detail:   mcpSquareIntro.Detail,
	}
}

func toToolAction(tool *common.ToolAction) *protocol.Tool {
	ret := &protocol.Tool{
		Name:        tool.Name,
		Description: tool.Description,
		InputSchema: protocol.InputSchema{
			Type:       protocol.InputSchemaType(tool.InputSchema.GetType()),
			Required:   tool.InputSchema.GetRequired(),
			Properties: make(map[string]*protocol.Property),
		},
	}
	for k, v := range tool.InputSchema.GetProperties() {
		ret.InputSchema.Properties[k] = &protocol.Property{
			Type:        protocol.DataType(v.Type),
			Description: v.Description,
		}
	}
	return ret
}
