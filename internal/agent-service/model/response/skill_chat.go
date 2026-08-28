package response

import (
	"strings"

	agent_util "github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	"github.com/cloudwego/eino/schema"
)

type SkillStep int

const (
	SkillNoneStep           SkillStep = 0 //非技能消息
	SkillTransferStartStep  SkillStep = 1 //技能转换开始也是技能开始
	SkillTransferringStep   SkillStep = 2 //技能转换过程中可能是流式
	SkillTransferFinishStep SkillStep = 3 //技能转换完成
	SkillResultStep         SkillStep = 4 //技能结果
)

type SkillChatContext struct {
	SkillStart bool
	SkillId    string
}

func (s *SkillChatContext) SkillTransferStart(skillId string) {
	s.SkillStart = true
	s.SkillId = skillId
}
func (s *SkillChatContext) SkillTransferFinish() {
	s.SkillStart = false
	s.SkillId = ""
}

func NewSkillChatContext() *SkillChatContext {
	return &SkillChatContext{}
}

type SkillChatStep struct {
	ChatMessage *schema.Message
	RespContext *AgentChatRespContext
}

func CreateSkillChatStep(respContext *AgentChatRespContext, chatMessage *schema.Message) *SkillChatStep {
	return &SkillChatStep{
		ChatMessage: chatMessage,
		RespContext: respContext,
	}
}

func (s *SkillChatStep) SkillNeedProcess() bool {
	step := s.BuildSkillStep()
	return step != SkillNoneStep
}

func (s *SkillChatStep) BuildSkillStep() SkillStep {
	if s.skillTransferStart() {
		return SkillTransferStartStep
	}
	//技能转换完成这个必须放在skillTransferring 之前判断，因为转换中且finish消息才是完成
	if s.skillTransferFinish() {
		return SkillTransferFinishStep
	}
	if s.skillTransferring() {
		return SkillTransferringStep
	}
	if s.skillResultMessage() {
		return SkillResultStep
	}
	return SkillNoneStep
}

func (s *SkillChatStep) skillTransferStart() bool {
	chatMessage := s.ChatMessage
	if len(chatMessage.ToolCalls) > 0 {
		call := chatMessage.ToolCalls[0]
		if strings.HasPrefix(call.Function.Name, agent_util.AgentSkillPrefix) {
			return true
		}
	}
	return false
}

func (s *SkillChatStep) skillTransferring() bool {
	return s.RespContext.SkillChatContext.SkillStart
}

func (s *SkillChatStep) skillTransferFinish() bool {
	if s.skillTransferring() {
		chatMessage := s.ChatMessage
		return chatMessage.ResponseMeta != nil && chatMessage.ResponseMeta.FinishReason == "tool_calls"
	}
	return false
}

func (s *SkillChatStep) skillResultMessage() bool {
	chatMessage := s.ChatMessage
	return chatMessage.Role == schema.Tool && len(chatMessage.Content) > 0 && strings.HasPrefix(chatMessage.ToolName, agent_util.AgentSkillPrefix)
}
