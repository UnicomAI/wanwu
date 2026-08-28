package response

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	"github.com/cloudwego/eino/schema"
)

const (
	defaultAgentAvatar = "/v1/static/icon/agent-default-icon.png"
)

type MultiAgentNameInfo struct {
	AgentName string `json:"agent_name"`
}

type MultiAgentContext struct {
	MainAgentName    string //主智能体名称
	AgentChangeStart bool   //智能体切换开始
	AgentStartTime   int64
	AgentTempMessage strings.Builder
	CurrentAgent     *util.Stack[*AgentInfo] //当前智能体
	AgentExitStart   bool                    //退出工具开始
}

func NewMultiAgentContext(mainAgentName string) *MultiAgentContext {
	return &MultiAgentContext{
		MainAgentName: mainAgentName,
	}
}

func (m *MultiAgentContext) ChangeStart() {
	m.AgentChangeStart = true
	m.AgentStartTime = time.Now().UnixMilli()
	m.AgentTempMessage.Reset()
}

func (m *MultiAgentContext) ChangeFinish() {
	m.AgentChangeStart = false
	m.AgentTempMessage.Reset()
}

func (m *MultiAgentContext) WriteAgentName(chatMessage *schema.Message) {
	m.AgentTempMessage.WriteString(chatMessage.ToolCalls[0].Function.Arguments)
}

func (m *MultiAgentContext) PeekParentAgent() *AgentInfo {
	agent := m.CurrentAgent
	if agent == nil {
		return nil
	}
	peekParent, has := agent.PeekParent()
	if !has {
		return nil
	}
	return peekParent
}

func (m *MultiAgentContext) PeekAgent() *AgentInfo {
	agent := m.CurrentAgent
	if agent == nil {
		return nil
	}
	peek, has := agent.Peek()
	if !has {
		return nil
	}
	return peek
}

func (m *MultiAgentContext) CurrentAgentId() string {
	agent := m.PeekAgent()
	if agent == nil {
		return ""
	}
	return agent.Id
}

func (m *MultiAgentContext) CreateAgent(subAgentMap map[string]*request.AgentConfig) {
	agentName := buildAgentName(m.AgentTempMessage.String())
	if m.CurrentAgent == nil {
		m.CurrentAgent = util.NewStack[*AgentInfo]()
	}
	m.CurrentAgent.Push(CreateAgentInfo(agentName, buildAgentAvatar(agentName, subAgentMap)))
}

func (m *MultiAgentContext) AgentOrder(order int) {
	agent := m.PeekAgent()
	if agent == nil {
		return
	}
	agent.Order = order
}

func (m *MultiAgentContext) ClearAgent() {
	m.CurrentAgent.Pop()
}

// buildAgentName 构造智能体名称
func buildAgentName(tempMessage string) string {
	if len(tempMessage) == 0 {
		return ""
	}
	if !json.Valid([]byte(tempMessage)) {
		return ""
	}
	var agentInfo = &MultiAgentNameInfo{}
	_ = json.Unmarshal([]byte(tempMessage), agentInfo)
	return agentInfo.AgentName
}

// buildAgentAvatar 构造智能体头像
func buildAgentAvatar(agentName string, subAgentMap map[string]*request.AgentConfig) string {
	if len(subAgentMap) == 0 {
		return buildDefaultAvatar(agentName)
	}
	agentConfig := subAgentMap[agentName]
	if agentConfig == nil || len(agentConfig.AgentAvatar) == 0 {
		return buildDefaultAvatar(agentName)
	}
	return agentConfig.AgentAvatar
}

func buildDefaultAvatar(agentName string) string {
	if strings.HasPrefix(agentName, util.AgentSkillPrefix) {
		return defaultSkillAvatar
	}
	return defaultAgentAvatar
}

type MultiAgentStep struct {
	ChatMessage *schema.Message
	RespContext *AgentChatRespContext
}

func CreateMultiAgentStep(respContext *AgentChatRespContext, chatMessage *schema.Message) *MultiAgentStep {
	return &MultiAgentStep{
		ChatMessage: chatMessage,
		RespContext: respContext,
	}
}
func (m *MultiAgentStep) MainTransferStart() bool {
	if mainTransferStart(m) {
		m.RespContext.IncreaseOrder()
		return true
	}
	return false
}

func (m *MultiAgentStep) TransferStart() bool {
	chatMessage := m.ChatMessage
	if len(chatMessage.ToolCalls) == 0 {
		return false
	}
	toolCall := chatMessage.ToolCalls[0]
	functionName := toolCall.Function.Name
	return util.AgentStartLabel == functionName
}

func (m *MultiAgentStep) TransferFinish() bool {
	chatMessage := m.ChatMessage
	return chatMessage.Role == schema.Tool && chatMessage.ToolName == util.AgentStartLabel
}

func (m *MultiAgentStep) SubAgentFinish() bool {
	chatMessage := m.ChatMessage
	//子智能体结束
	return util.StopMessage(chatMessage) && m.RespContext.MultiAgentContext.PeekAgent() != nil
}

// ExitStart 多智能体结束会输出exitToolStart，是在创建时传进去的ExitTool
func (m *MultiAgentStep) ExitStart() bool {
	if exitStart(m) {
		m.RespContext.MultiAgentContext.AgentExitStart = true
		return true
	}
	return false
}

func (m *MultiAgentStep) ExitFinish() bool {
	chatMessage := m.ChatMessage
	if chatMessage.Role == schema.Tool && chatMessage.ToolName == util.MainAgentExitLabel {
		m.RespContext.MultiAgentContext.AgentExitStart = false
		return true
	}
	return false
}

func exitStart(m *MultiAgentStep) bool {
	if m.ExitFinish() {
		return false
	}
	chatMessage := m.ChatMessage
	if m.RespContext.MultiAgentContext.AgentExitStart {
		return true
	}
	if len(chatMessage.ToolCalls) > 0 {
		toolCall := chatMessage.ToolCalls[0]
		//因为不同模型输出tool不一样，如果同时出现exit 参数和返回都输出，则不认为exit 不用设置开始直接处理结束就行
		if toolCall.Function.Name == util.MainAgentExitLabel {
			return true
		}
	}
	return false
}

func mainTransferStart(m *MultiAgentStep) bool {
	if !m.TransferStart() {
		return false
	}
	agentName := m.ChatMessage.ToolCalls[0].Function.Arguments
	return agentName == m.RespContext.MultiAgentContext.MainAgentName
}
