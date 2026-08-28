package params_process

import (
	"context"
	"encoding/json"
	"errors"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type PluginToolProcess struct {
}

func init() {
	AddServiceContainer(&PluginToolProcess{})
}

func (k *PluginToolProcess) ServiceType() ServiceType {
	return PluginToolType
}

func (k *PluginToolProcess) Prepare(agent *AgentInfo, prepareParams *AgentPrepareParams, clientInfo *ClientInfo, userQueryParams *UserQueryParams) error {
	resp, err := buildAssistantTools(agent, clientInfo)
	if err != nil {
		return errors.New("GetAssistantToolList error")
	}
	if len(resp) == 0 {
		return nil
	}
	var customToolIdList, builtInToolIdList, assistantToolMap = buildToolIdList(resp)
	if len(customToolIdList) == 0 && len(builtInToolIdList) == 0 {
		return nil
	}
	// 获取工具详情
	toolResp, err := clientInfo.MCP.GetToolDetailByIdList(userQueryParams.Ctx, &mcp_service.GetToolByToolIdListReq{
		BuiltInToolIdList: builtInToolIdList,
		CustomToolIdList:  customToolIdList,
		Identity: &mcp_service.Identity{
			UserId: "",
			OrgId:  "",
		},
	})
	if err != nil {
		return err
	}
	prepareParams.SquareToolList = toolResp.ToolSquareInfoList
	prepareParams.CustomToolList = toolResp.CustomList
	prepareParams.AssistantToolMap = assistantToolMap
	return nil
}
func (k *PluginToolProcess) Build(assistant *AgentInfo, prepareParams *AgentPrepareParams, agentChatParams *assistant_service.AgentDetail) error {
	customList, err := buildCustomToolPluginList(assistant.Assistant, prepareParams)
	if len(customList) > 0 {
		agentChatParams.ToolParams.PluginToolList = append(agentChatParams.ToolParams.PluginToolList, customList...)
	}
	if err != nil {
		return err
	}
	squareList, err := buildToolSquarePluginList(assistant.Assistant, prepareParams)
	if len(squareList) > 0 {
		agentChatParams.ToolParams.PluginToolList = append(agentChatParams.ToolParams.PluginToolList, squareList...)
	}
	return err
}

func buildAssistantTools(agent *AgentInfo, clientInfo *ClientInfo) ([]*model.AssistantTool, error) {
	if agent.Draft {
		list, status := clientInfo.Cli.GetAssistantToolList(context.Background(), agent.Assistant.ID)
		if status != nil {
			return nil, errors.New("GetAssistantToolList error")
		}
		return list, nil
	}
	var toolList []*model.AssistantTool
	if agent.AssistantSnapshot.AssistantToolConfig != "" {
		if err := json.Unmarshal([]byte(agent.AssistantSnapshot.AssistantToolConfig), &toolList); err != nil {
			return nil, errors.New("GetAssistantSnapshotToolList error")
		}
	}
	return toolList, nil
}

// buildToolIdList构建工具id列表
func buildToolIdList(resp []*model.AssistantTool) (customToolIdList []string, builtInToolIdList []string, assistantToolMap map[string][]string) {
	// 遍历工具列表，处理每个有效工具
	for _, tool := range resp {
		if !tool.Enable {
			continue // 跳过禁用的工具
		}
		if len(assistantToolMap) == 0 {
			assistantToolMap = make(map[string][]string)
		}
		// 根据工具类型获取详情和原始schema
		switch tool.ToolType {
		case constant.ToolTypeCustom:
			customToolIdList = append(customToolIdList, tool.ToolId)
			fillToolMap(assistantToolMap, tool)
		case constant.ToolTypeBuiltIn:
			builtInToolIdList = append(builtInToolIdList, tool.ToolId)
			fillToolMap(assistantToolMap, tool)
		}
	}
	return customToolIdList, builtInToolIdList, assistantToolMap
}

func fillToolMap(assistantToolMap map[string][]string, tool *model.AssistantTool) {
	dataList, exist := assistantToolMap[tool.ToolId]
	if !exist {
		assistantToolMap[tool.ToolId] = []string{tool.ActionName}
	} else {
		dataList = append(dataList, tool.ActionName)
		assistantToolMap[tool.ToolId] = dataList
	}
}

func buildCustomToolPluginList(assistant *model.Assistant, prepareParams *AgentPrepareParams) ([]*assistant_service.PluginToolInfo, error) {
	var pluginList []*assistant_service.PluginToolInfo
	if len(prepareParams.CustomToolList) > 0 {
		for _, customTool := range prepareParams.CustomToolList {
			apiAuth, rawSchema := buildCustomToolInfo(assistant, customTool)
			// 处理schema
			toolActionList, exist := prepareParams.AssistantToolMap[customTool.CustomToolId]
			if !exist {
				log.Infof("assistantId: %d, toolId: %s not exist", assistant.ID, customTool.CustomToolId)
				continue
			}
			for _, actionName := range toolActionList {
				apiSchema, err := processSchema(context.Background(), rawSchema, actionName)
				if err != nil {
					return pluginList, err
				}
				pluginList, err = buildPluginList(pluginList, apiSchema, apiAuth, "", actionName)
				if err != nil {
					return pluginList, err
				}
			}

		}
	}
	return pluginList, nil
}

func buildPluginList(pluginList []*assistant_service.PluginToolInfo, apiSchema map[string]interface{}, apiAuth *openapi3_util.Auth, avatar, actionName string) ([]*assistant_service.PluginToolInfo, error) {
	request := config.PluginListAlgRequest{
		APISchema: apiSchema,
		APIAuth:   apiAuth,
	}
	marshal, err := json.Marshal(request)
	if err != nil {
		return pluginList, err
	}
	pluginList = append(pluginList, &assistant_service.PluginToolInfo{
		PluginTool: string(marshal),
		Avatar:     avatar,
		Name:       actionName,
	})
	return pluginList, nil
}

func buildCustomToolInfo(assistant *model.Assistant, customTool *mcp_service.GetCustomToolInfoResp) (*openapi3_util.Auth, string) {
	// 构建自定义工具的API认证
	if customTool.ApiAuth != nil {
		apiAuth, err := util.ConvertApiAuthWebRequestProto(customTool.ApiAuth)
		if err != nil {
			log.Errorf("转换自定义工具API失败，assistantId: %d, toolId: %s, err: %v", assistant.ID, customTool.CustomToolId, err)
			return nil, customTool.Schema
		}
		return apiAuth, customTool.Schema
	}
	return nil, customTool.Schema
}

func buildToolSquarePluginList(assistant *model.Assistant, prepareParams *AgentPrepareParams) ([]*assistant_service.PluginToolInfo, error) {
	var pluginList []*assistant_service.PluginToolInfo
	if len(prepareParams.SquareToolList) > 0 {
		for _, squareTool := range prepareParams.SquareToolList {
			apiAuth, rawSchema := buildToolSquareInfo(assistant, squareTool)
			// 处理schema
			toolActionList, exist := prepareParams.AssistantToolMap[squareTool.Info.ToolSquareId]
			if !exist {
				log.Infof("assistantId: %d, toolId: %s not exist", assistant.ID, squareTool.Info.ToolSquareId)
				continue
			}
			for _, actionName := range toolActionList {
				apiSchema, err := processSchema(context.Background(), rawSchema, actionName)
				if err != nil {
					return pluginList, err
				}
				pluginList, err = buildPluginList(pluginList, apiSchema, apiAuth, "", actionName)
				if err != nil {
					return pluginList, err
				}
			}

		}
	}
	return pluginList, nil
}

func buildToolSquareInfo(assistant *model.Assistant, builtinTool *mcp_service.SquareToolDetail) (*openapi3_util.Auth, string) {
	// 构建内置工具的API认证
	apiAuth, err := util.ConvertApiAuthWebRequestProto(builtinTool.BuiltInTools.ApiAuth)
	if err != nil {
		log.Errorf("转换内置工具API失败，assistantId: %d, toolId: %s, err: %v", assistant.ID, builtinTool.Info.ToolSquareId, err)
		return nil, builtinTool.Schema
	}
	return apiAuth, builtinTool.Schema
}

func processSchema(ctx context.Context, rawSchema string, actionName string) (map[string]interface{}, error) {
	// 过滤schema中的指定operation_id
	filteredSchema, err := openapi3_util.FilterSchemaOperations(ctx, []byte(rawSchema), []string{actionName})
	if err != nil {
		return nil, err
	}

	// 校验schema格式
	validatedSchema, err := openapi3_util.LoadFromData(ctx, filteredSchema)
	if err != nil {
		return nil, err
	}

	// 转换为map[string]interface{}
	schemaBytes, err := json.Marshal(validatedSchema)
	if err != nil {
		return nil, err
	}

	var apiSchema map[string]interface{}
	if err := json.Unmarshal(schemaBytes, &apiSchema); err != nil {
		return nil, err
	}

	return apiSchema, nil
}
