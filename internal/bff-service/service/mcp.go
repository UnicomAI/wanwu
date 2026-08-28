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
	pkg_util "github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

type MCPBiz struct{}

func init() {
	InitBizService(&MCPBiz{})
}

func (*MCPBiz) BizType() string { return constant.BizModuleResourceMCP }

func (*MCPBiz) SearchBizOwner(ctx *gin.Context, bizId string) (userId, orgId string, err error) {
	// 优先按自定义MCP查询，查不到再按MCP服务查询（mcp/mcpserver 共用同一 bizType）
	resp, err := mcp.GetCustomMCP(ctx, &mcp_service.GetCustomMCPReq{
		McpId: bizId,
	})
	if err == nil && resp != nil && resp.Owner != nil {
		return resp.Owner.UserId, resp.Owner.OrgId, nil
	}
	mcpServerResp, err := mcp.GetMCPServer(ctx, &mcp_service.GetMCPServerReq{
		McpServerId: bizId,
	})
	if err != nil {
		return "", "", err
	}
	if mcpServerResp.Owner == nil {
		return "", "", nil
	}
	return mcpServerResp.Owner.UserId, mcpServerResp.Owner.OrgId, nil
}

func (*MCPBiz) SearchConversationLog(ctx *gin.Context, bizId, sourceFrom string) (*common.ConversationLog, error) {
	return nil, nil
}

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
	if req.MCPSquareID != "" {
		req.Transport = constant.MCPTransportSSE
		req.ApiAuth = &pkg_util.ApiAuthWebRequest{
			AuthType: pkg_util.AuthTypeNone,
		}
	}
	_, err := mcp.CreateCustomMCP(ctx.Request.Context(), &mcp_service.CreateCustomMCPReq{
		OrgId:         orgID,
		UserId:        userID,
		McpSquareId:   req.MCPSquareID,
		Name:          req.Name,
		Desc:          req.Desc,
		From:          req.From,
		SseUrl:        req.SSEURL,
		StreamableUrl: req.StreamableURL,
		Transport:     req.Transport,
		AvatarPath:    req.Avatar.Key,
		ApiAuth:       toApiAuthProto(req.ApiAuth),
		Headers:       req.Headers,
	})
	return err
}

func UpdateMCP(ctx *gin.Context, userID, orgID string, req request.MCPUpdate) error {
	existingMCP, err := mcp.GetCustomMCP(ctx.Request.Context(), &mcp_service.GetCustomMCPReq{McpId: req.MCPID})
	if err != nil {
		return err
	}
	if err := pkg_util.ValidateBriefUpdateAllowSpace(&req.Name, existingMCP.Info.Name, &req.Desc, existingMCP.Info.Desc, pkg_util.SubjectMCP); err != nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, err.Error())
	}
	_, err = mcp.UpdateCustomMCP(ctx.Request.Context(), &mcp_service.UpdateCustomMCPReq{
		OrgId:         orgID,
		UserId:        userID,
		McpId:         req.MCPID,
		Name:          req.Name,
		Desc:          req.Desc,
		From:          req.From,
		SseUrl:        req.SSEURL,
		StreamableUrl: req.StreamableURL,
		Transport:     req.Transport,
		AvatarPath:    req.Avatar.Key,
		ApiAuth:       toApiAuthProto(req.ApiAuth),
		Headers:       req.Headers,
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

func DeleteMCP(ctx *gin.Context, userID, orgID, mcpID string) error {
	// 删除智能体表AssistantMCP相关记录
	_, err := assistant.AssistantMCPDeleteByMCPId(ctx.Request.Context(), &assistant_service.AssistantMCPDeleteByMCPIdReq{
		McpId:   mcpID,
		McpType: constant.MCPTypeMCP,
		Identity: &assistant_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		return err
	}

	_, err = mcp.DeleteCustomMCP(ctx.Request.Context(), &mcp_service.DeleteCustomMCPReq{
		McpId:  mcpID,
		OrgId:  orgID,
		UserId: userID,
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
	// 获取自定义mcp列表
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
			// 兼容旧版
			MCPID:       mcpInfo.McpId,
			MCPSquareID: mcpInfo.Info.McpSquareId,
			Name:        mcpInfo.Info.Name,
			// 适用于智能体mcp下拉
			ToolId:   mcpInfo.McpId,
			ToolName: mcpInfo.Info.Name,
			ToolType: constant.MCPTypeMCP,
			// 共有字段
			Description:   mcpInfo.Info.Desc,
			ServerFrom:    mcpInfo.Info.From,
			ServerURL:     mcpInfo.SseUrl,
			StreamableURL: mcpInfo.StreamableUrl,
			Transport:     mcpInfo.Transport,
			Type:          constant.MCPTypeMCP,
			Avatar:        cacheMCPAvatar(ctx, mcpInfo.Info.AvatarPath, mcpInfo.AvatarPath),
			ApiAuth:       toApiAuthResponse(mcpInfo.GetApiAuth()),
			Headers:       mcpInfo.GetHeaders(),
		})
	}

	// 获取mcp server列表
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
			MCPID:         mcpServerInfo.McpServerId,
			MCPSquareID:   "",
			UniqueId:      bff_util.ConcatAssistantToolUniqueId(constant.AppTypeMCPServer, mcpServerInfo.McpServerId),
			Name:          mcpServerInfo.Name,
			Description:   mcpServerInfo.Desc,
			ServerFrom:    "mcp server",
			ServerURL:     mcpServerInfo.SseUrl,
			StreamableURL: mcpServerInfo.StreamableUrl,
			Transport:     mcpServerInfo.Transport,
			Type:          constant.MCPTypeMCPServer,
			// 适用于智能体mcp下拉
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

func GetMCPToolList(ctx *gin.Context, req request.MCPToolListReq) (*response.MCPToolList, error) {
	transportType := constant.MCPTransportSSE // 默认使用 sse
	serverUrl := req.ServerURL
	auth := req.ApiAuth
	headers := req.Headers
	if req.MCPID != "" {
		switch req.Type {
		case constant.MCPTypeMCP, "":
			mcpDetail, err := mcp.GetCustomMCP(ctx.Request.Context(), &mcp_service.GetCustomMCPReq{
				McpId: req.MCPID,
			})
			if err != nil {
				return nil, err
			}
			// 根据 transport 字段选择 URL
			switch mcpDetail.Transport {
			case constant.MCPTransportStreamable:
				serverUrl = mcpDetail.StreamableUrl
				transportType = constant.MCPTransportStreamable
			case constant.MCPTransportSSE:
				serverUrl = mcpDetail.SseUrl
				transportType = constant.MCPTransportSSE
			default:
				return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "transport empty")
			}
			// 获取鉴权信息和自定义请求头
			if mcpDetail.GetApiAuth() != nil {
				authResp := toApiAuthResponse(mcpDetail.GetApiAuth())
				auth = &authResp
			}
			headers = mcpDetail.GetHeaders()
		case constant.MCPTypeMCPServer:
			mcpServerDetail, err := mcp.GetMCPServer(ctx.Request.Context(), &mcp_service.GetMCPServerReq{
				McpServerId: req.MCPID,
			})
			if err != nil {
				return nil, err
			}
			// 根据 transport 字段选择 URL
			switch mcpServerDetail.Transport {
			case constant.MCPTransportStreamable:
				serverUrl = mcpServerDetail.StreamableUrl
				transportType = constant.MCPTransportStreamable
			case constant.MCPTransportSSE:
				serverUrl = mcpServerDetail.SseUrl
				transportType = constant.MCPTransportSSE
			default:
				return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "transport empty")
			}
		}
	} else if req.Transport != "" {
		// 如果传入了 transport 参数，使用传入的值
		transportType = req.Transport
	}
	if serverUrl == "" {
		return nil, grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "url empty")
	}

	tools, err := mcp_util.ListToolsWithAuth(ctx.Request.Context(), serverUrl, transportType, auth, headers)
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
			toolActions, err := openapi3_util.Schema2MCPProtocolTools(ctx.Request.Context(), []byte(tool.Schema))
			if err != nil {
				return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
			}
			actions = append(actions, toolActions...)
		}
	case constant.MCPTypeMCP:
		tools, err := GetMCPToolList(ctx, request.MCPToolListReq{
			MCPID: req.ToolId,
			Type:  req.ToolType,
		})
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

func GetMCPAvatar(ctx *gin.Context, avatarPath string) (*mcp_service.MCPAvatar, error) {
	return mcp.GetMCPAvatar(ctx.Request.Context(), &mcp_service.GetMCPAvatarReq{
		AvatarPath: avatarPath,
	})
}

// --- internal ---

func toMCPCustomDetail(ctx *gin.Context, mcpDetail *mcp_service.CustomMCPDetail) *response.MCPDetail {
	return &response.MCPDetail{
		MCPInfo: response.MCPInfo{
			MCPID:         mcpDetail.McpId,
			SSEURL:        mcpDetail.SseUrl,
			StreamableURL: mcpDetail.StreamableUrl,
			Transport:     mcpDetail.Transport,
			ApiAuth:       toApiAuthResponse(mcpDetail.GetApiAuth()),
			Headers:       mcpDetail.GetHeaders(),
			MCPSquareInfo: toMCPSquareInfo(ctx, mcpDetail.Info, mcpDetail.AvatarPath),
		},
		MCPSquareIntro: toMCPSquareIntro(mcpDetail.Intro),
	}
}

func toMCPCustomInfo(ctx *gin.Context, mcpInfo *mcp_service.CustomMCPInfo) response.MCPInfo {
	return response.MCPInfo{
		MCPID:         mcpInfo.McpId,
		Type:          constant.MCPTypeMCP,
		SSEURL:        mcpInfo.SseUrl,
		StreamableURL: mcpInfo.StreamableUrl,
		Transport:     mcpInfo.Transport,
		ApiAuth:       toApiAuthResponse(mcpInfo.GetApiAuth()),
		Headers:       mcpInfo.GetHeaders(),
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
		ret.Tools = append(ret.Tools, toToolAction(tool))
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
			Type:        protocol.PropertyType{protocol.DataType(v.Type)},
			Description: v.Description,
		}
	}
	return ret
}

func toApiAuthProto(auth *pkg_util.ApiAuthWebRequest) *common.ApiAuthWebRequest {
	return &common.ApiAuthWebRequest{
		AuthType:           auth.AuthType,
		ApiKeyHeaderPrefix: auth.ApiKeyHeaderPrefix,
		ApiKeyHeader:       auth.ApiKeyHeader,
		ApiKeyQueryParam:   auth.ApiKeyQueryParam,
		ApiKeyValue:        auth.ApiKeyValue,
	}
}

func toApiAuthResponse(auth *common.ApiAuthWebRequest) pkg_util.ApiAuthWebRequest {
	if auth == nil {
		return pkg_util.ApiAuthWebRequest{}
	}
	return pkg_util.ApiAuthWebRequest{
		AuthType:           auth.GetAuthType(),
		ApiKeyHeaderPrefix: auth.GetApiKeyHeaderPrefix(),
		ApiKeyHeader:       auth.GetApiKeyHeader(),
		ApiKeyQueryParam:   auth.GetApiKeyQueryParam(),
		ApiKeyValue:        auth.GetApiKeyValue(),
	}
}

// toApiAuthPtr 将 proto ApiAuthWebRequest 转换为指针类型的 ApiAuthWebRequest。
func toApiAuthPtr(auth *common.ApiAuthWebRequest) *pkg_util.ApiAuthWebRequest {
	if auth == nil {
		return nil
	}
	ret := toApiAuthResponse(auth)
	return &ret
}
