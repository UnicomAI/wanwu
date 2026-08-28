package service

import (
	"fmt"

	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/gin-gonic/gin"
)

func GetGeneralAgentToolSelect(ctx *gin.Context, userId, orgId, agentId string) (*response.ListResult, error) {
	toolResp, err := mcp.GetToolSelect(ctx.Request.Context(), &mcp_service.GetToolSelectReq{
		Identity: &mcp_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	toolNameToInfo := make(map[string]*mcp_service.GetToolItem)
	for _, item := range toolResp.List {
		if item.ToolType == constant.ToolTypeBuiltIn {
			toolNameToInfo[item.ToolName] = item
		}
	}

	// 获取全量工具列表
	toolCategories, err := wga.GetAgentToolCategories(config.WgaCfg().AgentID)
	if err != nil {
		return nil, err
	}
	// 对全量工具列表进行条件覆盖，默认不限制工具选择
	for _, toolCategory := range toolCategories {
		toolCategory.Condition = "none"
	}
	// 如果agentId不为空，则根据agentId获取工具选择条件进行覆盖，限制工具选择
	if agentId != "" {
		agentToolCategories, err := wga.GetAgentToolCategories(agentId)
		if err != nil {
			return nil, err
		}
		for _, toolCategory := range toolCategories {
			for _, agentToolCategory := range agentToolCategories {
				if toolCategory.Category == agentToolCategory.Category {
					toolCategory.Condition = agentToolCategory.Condition
					break
				}
			}
		}
	}

	result := make([]response.GetGeneralAgentToolSelectResp, 0, len(toolCategories))
	for _, tc := range toolCategories {
		categoryResp := response.GetGeneralAgentToolSelectResp{
			Category:  gin_util.I18nKey(ctx, string(tc.Category)),
			Condition: string(tc.Condition),
			ToolList:  []response.ToolInfo{},
		}

		for _, t := range tc.Tools {
			if item, ok := toolNameToInfo[t.Doc.Info.Title]; ok {
				categoryResp.ToolList = append(categoryResp.ToolList, response.ToolInfo{
					ToolId:          item.ToolId,
					ToolName:        item.ToolName,
					ToolType:        item.ToolType,
					Desc:            item.Desc,
					NeedApiKeyInput: item.NeedApiKeyInput,
					APIKey:          item.ApiKey,
					Avatar:          cacheToolAvatar(ctx, constant.ToolTypeBuiltIn, item.AvatarPath),
				})
			}
		}

		result = append(result, categoryResp)
	}

	return &response.ListResult{
		List:  result,
		Total: int64(len(result)),
	}, nil

}

func GetGeneralAgentToolInfo(ctx *gin.Context, userId, orgId, toolId, toolType string) (*response.GeneralAgentToolInfoResp, error) {
	resp, err := mcp.GetSquareTool(ctx.Request.Context(), &mcp_service.GetSquareToolReq{
		ToolSquareId: toolId,
		Identity: &mcp_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, fmt.Sprintf("tool not found: %s", toolId))
	}

	var actions []*protocol.Tool
	if resp.BuiltInTools != nil {
		for _, tool := range resp.BuiltInTools.Tools {
			actions = append(actions, toToolAction(tool))
		}
	}

	return &response.GeneralAgentToolInfoResp{
		Actions: actions,
		ToolInfo: response.ToolInfo{
			ToolId:          resp.Info.ToolSquareId,
			ToolName:        resp.Info.Name,
			ToolType:        constant.ToolTypeBuiltIn,
			Desc:            resp.Info.Desc,
			NeedApiKeyInput: resp.BuiltInTools.NeedApiKeyInput,
			APIKey:          resp.BuiltInTools.ApiAuth.ApiKeyValue,
			Avatar:          cacheToolAvatar(ctx, constant.ToolTypeBuiltIn, resp.Info.AvatarPath),
		},
	}, nil
}

// --- internal wga tool ---

// buildWgaToolOptions 构建工具配置选项（复用逻辑）
func buildWgaToolOptions(ctx *gin.Context, userId, orgId string, toolList []*assistant_service.WgaConfigTool) ([]wga_option.Option, error) {
	var opts []wga_option.Option
	for _, tool := range toolList {
		switch tool.ToolType {
		case constant.ToolTypeBuiltIn:
			toolResp, err := mcp.GetSquareTool(ctx.Request.Context(), &mcp_service.GetSquareToolReq{
				ToolSquareId: tool.ToolId,
				Identity: &mcp_service.Identity{
					UserId: userId,
					OrgId:  orgId,
				},
			})
			if err != nil {
				// 工具不存在时跳过，不阻断运行
				log.Warnf("[wga] tool %s not found, skip: %v", tool.ToolId, err)
				continue
			}
			toolDetail := toToolSquareDetail(ctx, toolResp)

			authType := toolDetail.ApiAuth.AuthType
			if authType == "" {
				authType = util.AuthTypeNone
			}
			apiAuth := &util.ApiAuthWebRequest{
				AuthType:           authType,
				ApiKeyHeaderPrefix: toolDetail.ApiAuth.ApiKeyHeaderPrefix,
				ApiKeyHeader:       toolDetail.ApiAuth.ApiKeyHeader,
				ApiKeyQueryParam:   toolDetail.ApiAuth.ApiKeyQueryParam,
				ApiKeyValue:        toolDetail.ApiAuth.ApiKeyValue,
			}

			opts = append(opts, wga_option.WithToolConfig(wga_option.ToolConfig{
				Title:   toolDetail.Name,
				APIAuth: apiAuth,
			}))
		}
	}
	return opts, nil
}

// getValidToolIds 批量获取有效的Tool ID映射
func getValidToolIds(ctx *gin.Context, userId, orgId string, toolIds []string) (map[string]bool, error) {
	if len(toolIds) == 0 {
		return make(map[string]bool), nil
	}
	validIds := make(map[string]bool)
	for _, toolId := range toolIds {
		_, err := mcp.GetSquareTool(ctx.Request.Context(), &mcp_service.GetSquareToolReq{
			ToolSquareId: toolId,
			Identity: &mcp_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})
		if err == nil {
			validIds[toolId] = true
		}
	}
	return validIds, nil
}
