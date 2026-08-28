package agent_chat_builder

import (
	"encoding/json"
	"strconv"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/cloudwego/eino/schema"
)

type AgentStep int

const (
	AgentNoneProcessStep AgentStep = 0 //无需处理，过滤
	AgentStartStep       AgentStep = 1 //智能体开始
	AgentChatStep        AgentStep = 2 //智能体会话
	AgentStopStep        AgentStep = 3 //智能体结束
	AgentAllFinishStep   AgentStep = 4 //智能体全完成，透传内容
)

type MultiAgentMessageBuilder struct {
}

func NewMultiBuilder() *MultiAgentMessageBuilder {
	return &MultiAgentMessageBuilder{}
}
func (*MultiAgentMessageBuilder) MessageType() MessageType {
	return MultiAgentMessage
}
func (*MultiAgentMessageBuilder) FilterMessage(respContext *response.AgentChatRespContext, chatMessage *schema.Message) bool {
	if filterMessage(respContext, chatMessage) {
		return true
	}
	multiAgentStep := response.CreateMultiAgentStep(respContext, chatMessage)
	//智能体切换完成消息|切换回主智能体开始消息|supervisor 结束时会以exit结束（设置enio时传入），模型流式输出exit工具参数时过滤消息
	if multiAgentStep.TransferFinish() || multiAgentStep.MainTransferStart() || multiAgentStep.ExitStart() {
		return true
	}
	return false
}
func (*MultiAgentMessageBuilder) BuildContent(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message) ([]*response.AgentMessageContent, error) {
	step := buildAgentStep(req, chatMessage, respContext)
	switch step {
	case AgentNoneProcessStep: //无需处理
		return buildSkipMessage(), nil
	case AgentAllFinishStep: //直接返回内容
		return buildFinishMessage(respContext, chatMessage)
	case AgentChatStep: //智能体内容输出
		return buildChatMessage(req, respContext, chatMessage, step)
	default: //智能体开始/结束
		return buildAgentProcessContent(req, respContext, chatMessage, step)
	}
}

func buildFinishMessage(respContext *response.AgentChatRespContext, chatMessage *schema.Message) ([]*response.AgentMessageContent, error) {
	var retList []*response.AgentMessageContent
	if respContext.ThinkChatContext.Thinking {
		contents := respContext.ThinkChatContext.ThinkMessageByStep(chatMessage, response.ThinkFinish, respContext)
		if len(contents) > 0 {
			retList = append(retList, contents...)
		}
	}
	content := buildMessageContent([]string{chatMessage.Content}, nil)
	if len(content) > 0 {
		retList = append(retList, content...)
	}
	return retList, nil
}

func buildAgentProcessContent(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message, step AgentStep) ([]*response.AgentMessageContent, error) {
	var retList []*response.AgentMessageContent
	if step == AgentStartStep && respContext.ThinkChatContext.Thinking {
		contents := respContext.ThinkChatContext.ThinkMessageByStep(chatMessage, response.ThinkFinish, respContext)
		if len(contents) > 0 {
			retList = append(retList, contents...)
		}
	}
	//提取错误消息
	errMsg := extractErrMsg(chatMessage, step)
	if len(errMsg) == 0 && len(chatMessage.Content) > 0 {
		respContext.IncreaseOrder()
		contents := buildMessageContent([]string{chatMessage.Content}, buildSubAgentEvent(req, respContext, step))
		if len(contents) > 0 {
			retList = append(retList, contents...)
		}
		return retList, nil
	}

	var contentList []string
	subAgentEvent := buildSubAgentEvent(req, respContext, step)
	//错误消息重置消息状态为失败
	if len(errMsg) > 0 {
		subAgentEvent.ErrMessage = errMsg
		subAgentEvent.Status = response.EventFailStatus
		contentList = []string{errMsg}
	}

	contents := buildMessageContent(contentList, subAgentEvent)
	if len(contents) > 0 {
		retList = append(retList, contents...)
	}
	return retList, nil
}

func extractErrMsg(chatMessage *schema.Message, step AgentStep) string {
	if step == AgentStopStep && len(chatMessage.Content) > 0 {
		//错误消息处理
		errResp := response.AgentToolErrResp{}
		_ = json.Unmarshal([]byte(chatMessage.Content), &errResp)
		return errResp.ErrorMsg
	}
	return ""
}

// buildAgentStep 构建智能体步骤
func buildAgentStep(req *request.AgentChatContext, chatMessage *schema.Message, respContext *response.AgentChatRespContext) AgentStep {
	multiAgentStep := response.CreateMultiAgentStep(respContext, chatMessage)
	//智能体切换消息
	if multiAgentStep.TransferStart() {
		agentTransferStart(respContext, chatMessage)
	}
	//智能体切换中处理智能体名称
	if respContext.MultiAgentContext.AgentChangeStart {
		skip, agentStep := buildAgentNameAndStep(req, chatMessage, respContext)
		if !skip {
			return agentStep
		}
	}
	//子智能体结束
	if multiAgentStep.SubAgentFinish() {
		return AgentStopStep
	}
	//主智能体结束
	if multiAgentStep.ExitFinish() {
		return AgentAllFinishStep
	}
	return AgentChatStep
}

// buildChatMessage 构造智能体对话消息
func buildChatMessage(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message, step AgentStep) ([]*response.AgentMessageContent, error) {
	contentList, err := NewSingleBuilder().BuildContent(req, respContext, chatMessage)
	if err != nil {
		return nil, err
	}
	event := buildSubAgentEvent(req, respContext, step)

	for _, messageContent := range contentList {
		if event == nil {
			continue
		}
		if messageContent.SubEventData == nil {
			messageContent.SubEventData = buildSubEventData(respContext, event)
		} else {
			if len(messageContent.SubEventData.ParentId) == 0 {
				messageContent.SubEventData.ParentId = event.Id
			}
		}

		//子智能体的结束消息，不需要输出stop
		if event.Status == response.EventEndStatus {
			messageContent.NotStop = true
		}
	}

	return contentList, nil
}

func buildSubEventData(respContext *response.AgentChatRespContext, event *response.SubEventData) *response.SubEventData {
	if event.Status == response.EventProcessStatus && event.EventType == response.SubAgentEventType {
		eventData := event.Copy()
		eventData.ParentId = event.Id
		eventData.Id = event.Id + "-" + strconv.Itoa(respContext.Order)
		eventData.EventType = response.SkillTextEventType
		return eventData
	}
	return event
}

// agentTransferStart 智能体切换开始
func agentTransferStart(respContext *response.AgentChatRespContext, chatMessage *schema.Message) {
	respContext.MultiAgentContext.ChangeStart()
	respContext.ResetTool()
	respContext.AgentToolContext.DefaultToolId(chatMessage.ToolCalls[0].ID)
}

func buildSubAgentEvent(req *request.AgentChatContext, respContext *response.AgentChatRespContext, step AgentStep) *response.SubEventData {
	switch step {
	case AgentStartStep:
		//每切换一次智能体order + 1
		respContext.IncreaseOrder()
		respContext.MultiAgentContext.AgentOrder(respContext.Order)
		return response.BuildStartSubAgent(respContext)
	case AgentChatStep:
		return response.BuildProcessSubAgent(respContext)
	case AgentStopStep:
		subAgent := response.BuildEndSubAgent(respContext, util.NowSpanToHMS(respContext.MultiAgentContext.AgentStartTime))
		respContext.MultiAgentContext.ClearAgent()
		req.WriteAgent(respContext.MultiAgentContext.CurrentAgentId())
		return subAgent
	}
	return nil
}

// buildAgentNameAndStep 构建智能体名称和步骤
func buildAgentNameAndStep(req *request.AgentChatContext, chatMessage *schema.Message, respContext *response.AgentChatRespContext) (bool, AgentStep) {
	stepsMap, toolIdList := buildToolStep(chatMessage, respContext)
	if len(toolIdList) > 0 {
		//根据step循环构造输出的内容
		for _, toolId := range toolIdList {
			toolSteps := stepsMap[toolId]
			for _, step := range toolSteps {
				finish := writeAgentName(req, chatMessage, step, respContext)
				if finish {
					return false, AgentStartStep
				}
			}
		}
		return false, AgentNoneProcessStep
	}
	//走到这里可能有bad case了
	return true, AgentNoneProcessStep
}

// buildAgentStep 根据当前步骤构造需要输出的内容,构造<tool></tool>数据以及markdown格式
func writeAgentName(req *request.AgentChatContext, chatMessage *schema.Message, step response.ToolStep, respContext *response.AgentChatRespContext) bool {
	var finish bool
	switch step {
	case response.ToolParamStep:
		respContext.MultiAgentContext.WriteAgentName(chatMessage)
	case response.ToolParamFinishStep:
		respContext.MultiAgentContext.CreateAgent(req.SubAgentMap)
		//设置当前智能体id
		req.WriteAgent(respContext.MultiAgentContext.CurrentAgentId())
		//智能体参数输出完成
		respContext.MultiAgentContext.ChangeFinish()
		finish = true
	}
	return finish
}
