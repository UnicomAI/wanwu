package response

import (
	"strings"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	agent_util "github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/cloudwego/eino/schema"
)

type AgentEventType int
type SubEventStatus int

const (
	MainAgentEventType    = 0 //单智能体事件/多智能体主智能体
	SubAgentEventType     = 1 //子智能体事件
	KnowledgeEventType    = 2 //知识库事件
	ToolEventType         = 3 //工具事件
	SkillEventType        = 4 //技能事件
	SubAgentToolEventType = 5 //子智能体工具事件
	ThinkingEventType     = 6 //智能体思考事件

	SkillTextEventType = 20 //技能内容事件

	EventStartStatus   SubEventStatus = 1 //开始事件
	EventProcessStatus SubEventStatus = 2 //输出中
	EventEndStatus     SubEventStatus = 3 //结束事件
	EventFailStatus    SubEventStatus = 4 //子智能体失败
)

type AgentMessageContent struct {
	ContentList  []string
	SubEventData *SubEventData
	NotStop      bool
}

func (t *AgentMessageContent) Empty() bool {
	return len(t.ContentList) == 0 && t.SubEventData == nil
}

type SubEventData struct {
	Status     SubEventStatus `json:"status"`
	Id         string         `json:"id"`
	EventType  int            `json:"eventType"`
	Name       string         `json:"name"`
	Profile    string         `json:"profile"`
	TimeCost   string         `json:"timeCost"`
	ParentId   string         `json:"parentId"`
	Order      int            `json:"order"`
	ErrMessage string         `json:"errMessage"` //错误信息
}

func (s *SubEventData) Copy() *SubEventData {
	return &SubEventData{
		Status:    s.Status,
		Id:        s.Id,
		ParentId:  s.ParentId,
		Name:      s.Name,
		Profile:   s.Profile,
		TimeCost:  s.TimeCost,
		Order:     s.Order,
		EventType: s.EventType,
	}
}

func BuildEventTypeByTool(agentTool *AgentTool) int {
	var eventType int
	if agentTool.ToolName == agent_util.AgentSearchKnowledgeName {
		eventType = KnowledgeEventType
	} else if strings.HasPrefix(agentTool.ToolName, agent_util.AgentSkillPrefix) {
		eventType = SkillEventType
	} else {
		eventType = ToolEventType
	}
	return eventType
}

func BuildStartSubAgent(respContext *AgentChatRespContext) *SubEventData {
	subAgent := StartSubAgent(respContext.MultiAgentContext.PeekAgent(), respContext.Order, SubAgentEventType)
	parent := respContext.MultiAgentContext.PeekParentAgent()
	if parent != nil {
		subAgent.ParentId = parent.Id
	}
	return subAgent
}

func BuildProcessSubAgent(respContext *AgentChatRespContext) *SubEventData {
	return ProcessSubAgent(respContext.MultiAgentContext.PeekAgent(), respContext.Order, SubAgentEventType)
}

func BuildEndSubAgent(respContext *AgentChatRespContext, timeCost string) *SubEventData {
	return EndSubAgent(respContext.MultiAgentContext.PeekAgent(), timeCost, respContext.Order, SubAgentEventType)
}

func BuildStartTool(agentTool *AgentTool) *SubEventData {
	return StartSubAgent(&AgentInfo{
		Avatar: agentTool.Avatar,
		Id:     agentTool.ToolId,
		Name:   agentTool.ToolName,
	}, agentTool.Order, agentTool.ToolType)
}

func BuildProcessTool(agentTool *AgentTool) *SubEventData {
	return ProcessSubAgent(&AgentInfo{
		Avatar: agentTool.Avatar,
		Id:     agentTool.ToolId,
		Name:   agentTool.ToolName,
	}, agentTool.Order, agentTool.ToolType)
}

func BuildEndTool(agentTool *AgentTool) *SubEventData {
	return EndSubAgent(&AgentInfo{
		Avatar: agentTool.Avatar,
		Id:     agentTool.ToolId,
		Name:   agentTool.ToolName,
	}, util.NowSpanToHMS(agentTool.StartTime), agentTool.Order, agentTool.ToolType)
}

func StartSubAgent(agentInfo *AgentInfo, order int, eventType int) *SubEventData {
	return &SubEventData{
		Status:    EventStartStatus,
		Id:        agentInfo.Id,
		Name:      agentInfo.Name,
		Profile:   agentInfo.Avatar,
		Order:     order,
		EventType: eventType,
	}
}

func ProcessSubAgent(agentInfo *AgentInfo, order int, eventType int) *SubEventData {
	if agentInfo == nil || len(agentInfo.Id) == 0 || len(agentInfo.Name) == 0 {
		return nil
	}
	return &SubEventData{
		Status:    EventProcessStatus,
		Id:        agentInfo.Id,
		Name:      agentInfo.Name,
		Profile:   agentInfo.Avatar,
		Order:     order,
		EventType: eventType,
	}
}

func EndSubAgent(agentInfo *AgentInfo, timeCost string, order int, eventType int) *SubEventData {
	return &SubEventData{
		Status:    EventEndStatus,
		Id:        agentInfo.Id,
		Name:      agentInfo.Name,
		Profile:   agentInfo.Avatar,
		TimeCost:  timeCost,
		Order:     order,
		EventType: eventType,
	}
}

func buildSubAgentEventInfo(respContext *request.AgentChatContext, chatMessage *schema.Message, subAgentEventData *SubEventData, order int) ([]string, error) {
	var outputList = make([]string, 0)
	var agentChatResp = &AgentChatResp{
		Code:       agentSuccessCode,
		DetailId:   respContext.AgentChatReq.DetailId,
		Message:    "success",
		Response:   "",
		Order:      order,
		EventType:  buildEventType(subAgentEventData),
		EventData:  subAgentEventData,
		SearchList: buildSubAgentSearchList(subAgentEventData, respContext),
		Finish:     buildFinish(chatMessage, true),
		Usage:      buildUsage(chatMessage),
	}
	respString, err := buildRespString(agentChatResp)
	if err != nil {
		return nil, err
	}
	outputList = append(outputList, respString)
	return outputList, nil
}

// buildEventType 事件类型构造
func buildEventType(subEvent *SubEventData) AgentEventType {
	if subEvent == nil {
		return MainAgentEventType
	}
	if subEvent.EventType > 0 {
		return AgentEventType(subEvent.EventType)
	}
	return SubAgentEventType
}
