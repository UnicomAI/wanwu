package mcp

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/mcp-service/client/model"
	"github.com/UnicomAI/wanwu/internal/mcp-service/config"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/util"
)

func (s *Service) GetToolByIdList(ctx context.Context, req *mcp_service.GetToolByToolIdListReq) (*mcp_service.GetToolByToolIdListResp, error) {
	// 内置工具
	var toolSquareInfo []*mcp_service.ToolSquareInfo
	for _, toolCfg := range config.Cfg().Tools {
		for _, builtInTool := range req.BuiltInToolIdList {
			if builtInTool != "" && !strings.Contains(toolCfg.Name, builtInTool) {
				toolSquareInfo = append(toolSquareInfo, buildSquareToolInfo(*toolCfg))
			}
		}
	}
	// 自定义工具
	var ids []uint32
	for _, idStr := range req.CustomToolIdList {
		id := util.MustU32(idStr)
		ids = append(ids, id)
	}

	infos, err := s.cli.ListCustomToolsByCustomToolIDs(ctx, ids)
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetCustomToolListErr, err)
	}
	list := make([]*mcp_service.GetCustomToolItem, 0)
	for _, info := range infos {
		list = append(list, &mcp_service.GetCustomToolItem{
			CustomToolId: util.Int2Str(info.ID),
			Name:         info.Name,
			Description:  string(info.Description),
			AvatarPath:   info.AvatarPath,
		})
	}
	return &mcp_service.GetToolByToolIdListResp{
		List:               list,
		ToolSquareInfoList: toolSquareInfo,
	}, nil
}

func (s *Service) GetToolSelect(ctx context.Context, req *mcp_service.GetToolSelectReq) (*mcp_service.GetToolListResp, error) {
	if req.Identity == nil {
		return nil, errStatus(errs.Code_MCPGetCustomToolListErr, toErrStatus("mcp_get_custom_tool_list_err", "identity is empty"))
	}
	// search custom tools
	infos, err := s.cli.ListCustomTools(ctx, req.Identity.OrgId, req.Identity.UserId, req.Name)
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetCustomToolListErr, err)
	}
	list := make([]*mcp_service.GetToolItem, 0)
	for _, info := range infos {
		needApiKeyInput := true
		if info.Type == model.ApiAuthNone {
			needApiKeyInput = false
		}

		// apikey
		apiAuth := &common.ApiAuthWebRequest{}
		if err := json.Unmarshal([]byte(info.AuthJSON), apiAuth); err != nil {
			return nil, errStatus(errs.Code_MCPGetCustomToolListErr, toErrStatus("mcp_get_custom_tool_list_err", err.Error()))
		}

		list = append(list, &mcp_service.GetToolItem{
			ToolId:          util.Int2Str(info.ID),
			ToolName:        info.Name,
			Desc:            string(info.Description),
			ApiKey:          apiAuth.ApiKeyValue,
			ToolType:        constant.ToolTypeCustom,
			NeedApiKeyInput: needApiKeyInput,
			AvatarPath:      info.AvatarPath,
		})
	}
	// search builtin tools
	// 先把该用户所有已配置apikey内置工具查出来，构造成map<ToolSquareId, *model.CustomTool>，然后通过ToolSquareId查询apikey
	builtinTools, err := s.cli.ListBuiltinTools(ctx, req.Identity.OrgId, req.Identity.UserId)
	if err != nil {
		return nil, errStatus(errs.Code_MCPGetCustomToolListErr, err)
	}
	builtinToolMap := make(map[string]*model.BuiltinTool)
	for _, tool := range builtinTools {
		builtinToolMap[tool.ToolSquareId] = tool
	}

	for _, toolCfg := range config.Cfg().Tools {
		if req.Name != "" && !strings.Contains(toolCfg.Name, req.Name) {
			continue
		}
		toolTab := &mcp_service.GetToolItem{
			ToolId:          toolCfg.ToolSquareId,
			ToolName:        toolCfg.Name,
			Desc:            toolCfg.Desc,
			ToolType:        constant.ToolTypeBuiltIn,
			NeedApiKeyInput: toolCfg.NeedApiKeyInput,
			AvatarPath:      toolCfg.AvatarPath,
		}
		// 从map中查询内置工具
		if tool, ok := builtinToolMap[toolCfg.ToolSquareId]; ok {
			apiAuth := &common.ApiAuthWebRequest{}
			if err := json.Unmarshal([]byte(tool.AuthJSON), apiAuth); err != nil {
				return nil, errStatus(errs.Code_MCPGetCustomToolListErr, toErrStatus("mcp_get_custom_tool_list_err", err.Error()))
			}
			toolTab.ApiKey = apiAuth.ApiKeyValue

		}
		list = append(list, toolTab)
	}
	return &mcp_service.GetToolListResp{
		List:  list,
		Total: int32(len(list)),
	}, nil
}
