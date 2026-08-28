package params_process

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/constant"
)

type MCPToolInfo struct {
	URL          string   `json:"url"`
	Transport    string   `json:"transport"`
	ToolNameList []string `json:"toolNameList"` // MCP工具方法列表,会根据此方法名的列表进行mcp方法的过滤，如果此列为空，则标识不进行过滤
}

type McpProcess struct {
}

func init() {
	AddServiceContainer(&McpProcess{})
}

func (k *McpProcess) ServiceType() ServiceType {
	return McpType
}

func (k *McpProcess) Prepare(agent *AgentInfo, prepareParams *AgentPrepareParams, clientInfo *ClientInfo, userQueryParams *UserQueryParams) error {
	mcpInfos, err := buildMcpInfos(agent, clientInfo)
	if err != nil {
		return fmt.Errorf("Assistant服务获取MCP信息失败，assistantId: %d, error: %v", agent.Assistant.ID, err)
	}
	if len(mcpInfos) == 0 {
		return nil
	}
	customMcpIdList, mcpServerIdList, mcpToolMap := buildMcpIdList(mcpInfos)
	if len(customMcpIdList) == 0 && len(mcpServerIdList) == 0 {
		return nil
	}
	mcpListResp, err1 := clientInfo.MCP.GetMCPByMCPIdList(userQueryParams.Ctx, &mcp_service.GetMCPByMCPIdListReq{
		McpIdList:       customMcpIdList,
		McpServerIdList: mcpServerIdList,
		Identity: &mcp_service.Identity{
			UserId: "",
			OrgId:  "",
		},
	})
	if err1 != nil {
		return fmt.Errorf("MCP服务获取MCP信息失败，assistantId: %d, error: %v", agent.Assistant.ID, err1)
	}
	prepareParams.CustomMcpList = mcpListResp.Infos
	prepareParams.McpServerList = mcpListResp.Servers
	prepareParams.McpToolMap = mcpToolMap
	return nil
}
func (k *McpProcess) Build(assistant *AgentInfo, prepareParams *AgentPrepareParams, agentChatParams *assistant_service.AgentDetail) error {
	var mcpMap = make(map[string]config.MCPToolInfo)
	buildCustomMcpList(prepareParams, mcpMap)
	buildMcpServerList(prepareParams, mcpMap)
	if len(mcpMap) > 0 {
		var mcpToolList []*assistant_service.MCPToolInfo
		for _, info := range mcpMap {
			mcpToolList = append(mcpToolList, &assistant_service.MCPToolInfo{
				Transport:    info.Transport,
				Url:          info.URL,
				ToolNameList: info.ToolNameList,
				Avatar:       info.Avatar,
				Headers:      info.Headers,
				ApiAuth:      info.ApiAuth,
			})
		}
		agentChatParams.ToolParams.McpToolList = append(agentChatParams.ToolParams.McpToolList, mcpToolList...)
	}
	return nil
}

// buildMcpInfos 获取MCP列表
func buildMcpInfos(agent *AgentInfo, clientInfo *ClientInfo) ([]*model.AssistantMCP, error) {
	if agent.Draft {
		mcpInfos, err := clientInfo.Cli.GetAssistantMCPList(context.Background(), agent.Assistant.ID)
		if err != nil {
			return nil, errors.New("GetAssistantMCPList error")
		}
		return mcpInfos, nil
	}
	var mcpInfos []*model.AssistantMCP
	if len(agent.AssistantSnapshot.AssistantMCPConfig) > 0 {
		err := json.Unmarshal([]byte(agent.AssistantSnapshot.AssistantMCPConfig), &mcpInfos)
		if err != nil {
			return nil, err
		}
	}
	return mcpInfos, nil
}

// buildMcpIdList 构建MCP列表
func buildMcpIdList(mcpInfos []*model.AssistantMCP) (customMcpIdList []string, mcpServerIdList []string, mcpToolMap map[string][]string) {
	// 遍历工具列表，处理每个有效工具
	for _, mcp := range mcpInfos {
		if !mcp.Enable {
			continue
		}
		if len(mcpToolMap) == 0 {
			mcpToolMap = make(map[string][]string)
		}
		switch mcp.MCPType {
		case constant.MCPTypeMCP:
			customMcpIdList = append(customMcpIdList, mcp.MCPId)
			fillMcpTooMap(mcpToolMap, mcp)
		case constant.MCPTypeMCPServer:
			mcpServerIdList = append(mcpServerIdList, mcp.MCPId)
			fillMcpTooMap(mcpToolMap, mcp)
		}
	}
	return customMcpIdList, mcpServerIdList, mcpToolMap
}

func fillMcpTooMap(mcpToolMap map[string][]string, mcp *model.AssistantMCP) {
	dataList, exist := mcpToolMap[mcp.MCPId]
	if !exist {
		mcpToolMap[mcp.MCPId] = []string{mcp.ActionName}
	} else {
		dataList = append(dataList, mcp.ActionName)
		mcpToolMap[mcp.MCPId] = dataList
	}
}

func buildCustomMcpList(prepareParams *AgentPrepareParams, mcpTools map[string]config.MCPToolInfo) {
	if len(prepareParams.CustomMcpList) > 0 {
		for _, mcpCustom := range prepareParams.CustomMcpList {
			toolList := prepareParams.McpToolMap[mcpCustom.McpId]
			url, transport := selectMcpUrlAndTransportByType(mcpCustom.SseUrl, mcpCustom.StreamableUrl, mcpCustom.Transport)
			mcpTools[mcpCustom.McpId] = config.MCPToolInfo{
				URL:          url,
				Transport:    transport,
				ToolNameList: toolList,
				Headers:      mcpCustom.Headers,
				ApiAuth:      mcpCustom.ApiAuth,
			}
		}
	}
}

func buildMcpServerList(prepareParams *AgentPrepareParams, mcpTools map[string]config.MCPToolInfo) {
	if len(prepareParams.McpServerList) > 0 {
		for _, mcpServer := range prepareParams.McpServerList {
			toolList := prepareParams.McpToolMap[mcpServer.McpServerId]
			url, transport := selectMcpUrlAndTransportByType(mcpServer.SseUrl, mcpServer.StreamableUrl, mcpServer.Transport)
			mcpTools[mcpServer.McpServerId] = config.MCPToolInfo{
				URL:          url,
				Transport:    transport,
				ToolNameList: toolList,
			}
		}
	}
}

// selectMcpUrlAndTransportByType 根据 transport 类型选择 URL
// transport: "sse" 或 "streamable"，如果为空则根据 URL 是否存在自动选择
func selectMcpUrlAndTransportByType(sseUrl, streamableUrl, transport string) (url, selectedTransport string) {
	switch transport {
	case constant.MCPTransportStreamable:
		return streamableUrl, constant.MCPTransportStreamable
	case constant.MCPTransportSSE:
		return sseUrl, constant.MCPTransportSSE
	default:
		// 如果 transport 为空，根据 URL 是否存在自动选择
		if streamableUrl != "" {
			return streamableUrl, constant.MCPTransportStreamable
		}
		return sseUrl, constant.MCPTransportSSE
	}
}
