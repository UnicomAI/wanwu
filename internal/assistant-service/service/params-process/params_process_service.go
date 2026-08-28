package params_process

import (
	"context"
	"fmt"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
)

type ServiceType string

const (
	KnowledgeType         ServiceType = "KnowledgeService"         //知识库服务
	PluginToolType        ServiceType = "PluginToolService"        //自定义工具服务
	WorkflowType          ServiceType = "WorkflowService"          //工作流服务
	McpType               ServiceType = "McpService"               //mcp服务
	ConversionHistoryType ServiceType = "ConversionHistoryService" //会话历史
	SkillType             ServiceType = "SkillService"             //技能服务
)

var serviceList []ProcessService
var serviceMap = make(map[ServiceType]ProcessService)

type UserQueryParams struct {
	ConversationId string `json:"conversationId"` //会话id,当此参数不为空，会构造会话历史
	QueryUserId    string `json:"queryUserId"`
	QueryOrgId     string `json:"queryOrgId"`
	Ctx            context.Context
}

type AgentInfo struct {
	Assistant         *model.Assistant
	AssistantSnapshot *model.AssistantSnapshot
	Draft             bool
}

type AgentChatParams struct {
	Input           string           `json:"input"`
	Stream          bool             `json:"stream"`
	UploadFile      []string         `json:"uploadFile"`
	AgentBaseParams *AgentBaseParams `json:"agentBaseParams"` // 智能体基础参数
	ModelParams     *ModelParams     `json:"modelParams"`     // 模型参数
	KnowledgeParams *KnowledgeParams `json:"knowledgeParams"` // 知识库参数，如果后续需要增加透传，理论上只需要修改此KnowledgeParams即可
	ToolParams      *ToolParams      `json:"toolParams"`      // 工具相关参数，mcp tool plugin tool
}

type AgentBaseParams struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Instruction string `json:"instruction"`
}

type ModelParams struct {
	ModelId          string                              `json:"modelId"`
	History          []config.AssistantConversionHistory `json:"history,omitempty"`
	MaxHistory       int                                 `json:"maxHistory"`
	Temperature      *float32                            `json:"temperature,omitempty"`      //温度
	TopP             *float32                            `json:"topP,omitempty"`             //topP
	FrequencyPenalty *float32                            `json:"frequencyPenalty,omitempty"` //频率惩罚
	PresencePenalty  *float32                            `json:"presence_penalty,omitempty"` //存在惩罚
	MaxTokens        *int                                `json:"max_tokens,omitempty"`       //模型输出最大token数，这个字段暂时不设置，因为模型可能触发接口调用不确定是否会超，先不传
}

type ToolParams struct {
	PluginToolList []config.PluginListAlgRequest `json:"pluginTool,omitempty"`
	McpToolList    []*MCPToolInfo                `json:"mcpToolList,omitempty"`
}

type APIAuth struct {
	Type  string `json:"type"`
	In    string `json:"in"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type AgentPrepareParams struct {
	KnowledgeList        []*knowledgebase_service.KnowledgeInfo
	AssistantToolMap     map[string][]string
	CustomToolList       []*mcp_service.GetCustomToolInfoResp
	SquareToolList       []*mcp_service.SquareToolDetail
	WorkflowList         []map[string]interface{}
	CustomMcpList        []*mcp_service.CustomMCPInfo
	McpServerList        []*mcp_service.MCPServerInfo
	McpToolMap           map[string][]string
	ConversionDetailList []*model.ConversationDetails
	builtinSkillList     []*SkillDetail
	CustomSkillList      []*CustomSkillDetail
	AcquiredSkillList    []*AcquiredSkillDetail
	Err                  error
}

type ClientInfo struct {
	Cli       client.IClient
	Knowledge knowledgebase_service.KnowledgeBaseServiceClient
	MCP       mcp_service.MCPServiceClient
}

type ProcessService interface {
	ServiceType() ServiceType
	Prepare(assistant *AgentInfo, prepareParams *AgentPrepareParams, clientInfo *ClientInfo, userQueryParams *UserQueryParams) error
	Build(assistant *AgentInfo, prepareParams *AgentPrepareParams, agentChatParams *assistant_service.AgentDetail) error
}

func AddServiceContainer(service ProcessService) {
	serviceList = append(serviceList, service)
	serviceMap[service.ServiceType()] = service
}

func PrepareParams(serviceType ServiceType, assistant *AgentInfo, prepareParams *AgentPrepareParams, clientInfo *ClientInfo, userQueryParams *UserQueryParams) error {
	service := serviceMap[serviceType]
	if service == nil {
		return fmt.Errorf("service not found: %s", serviceType)
	}
	return service.Prepare(assistant, prepareParams, clientInfo, userQueryParams)
}

func BuildParams(serviceType ServiceType, assistant *AgentInfo, prepareParams *AgentPrepareParams, agentParams *assistant_service.AgentDetail) error {
	service := serviceMap[serviceType]
	if service == nil {
		return fmt.Errorf("service not found: %s", serviceType)
	}
	return service.Build(assistant, prepareParams, agentParams)
}
