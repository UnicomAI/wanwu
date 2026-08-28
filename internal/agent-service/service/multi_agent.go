package service

import (
	"context"
	"errors"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/grpc-consumer/consumer/assistant"
	agent_message_processor "github.com/UnicomAI/wanwu/internal/agent-service/service/agent-message-processor"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/gin-gonic/gin"
)

type MultiAgent struct {
	MultiAgent       adk.Agent
	AgentChatContext *request.AgentChatContext
}

// MultiAgentChat 多智能体问答
func MultiAgentChat(ctx *gin.Context, req *request.MultiAgentChatParams) error {
	multiAgentConfig, err := searchMultiAgentConfig(ctx, req)
	if err != nil {
		return err
	}
	agent, err := CreateSupervisorMultiAgent(ctx, req, multiAgentConfig)
	if err != nil {
		return err
	}
	return agent.Chat(ctx)
}

func CreateSupervisorMultiAgent(ctx *gin.Context, multiAgentChatParams *request.MultiAgentChatParams, multiAgentConfig *assistant_service.MultiAssistantDetailResp) (*MultiAgent, error) {
	var multiAgentChatReq = BuildMultiAgentParams(multiAgentChatParams, multiAgentConfig)
	agentChatParams := buildAgentChatParams(multiAgentChatReq)
	currentAgent := &request.AgentInfo{}
	agentChatContext := &request.AgentChatContext{AgentChatReq: agentChatParams, CurrentAgent: currentAgent}
	//构造supervisor,也是一个单智能体
	sv, err := CreateSingleAgent(ctx, agentChatParams)
	if err != nil {
		return nil, err
	}
	//构造子智能体
	multiSubAgent, subAgentMap, err := buildMultiSubAgent(ctx, multiAgentChatReq, currentAgent)
	if err != nil {
		return nil, err
	}
	agentChatContext.SubAgentMap = subAgentMap
	//构造多智能体
	agent, err := supervisor.New(ctx, &supervisor.Config{
		Supervisor: sv,
		SubAgents:  multiSubAgent,
	})
	if err != nil {
		return nil, err
	}
	return &MultiAgent{
		MultiAgent:       agent,
		AgentChatContext: agentChatContext,
	}, nil
}

func (s *MultiAgent) Chat(ctx *gin.Context) error {
	//1.执行流式agent问答调用
	req := s.AgentChatContext.AgentChatReq
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           s,
		EnableStreaming: req.Stream,
	})
	iter := runner.Query(ctx, req.Input)
	//2.处理结果
	err := agent_message_processor.MultiAgentMessage(ctx, iter, s.AgentChatContext)
	return err
}

func (s *MultiAgent) Name(ctx context.Context) string {
	return s.MultiAgent.Name(ctx)
}
func (s *MultiAgent) Description(ctx context.Context) string {
	return s.MultiAgent.Description(ctx)
}

func (s *MultiAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	log.Infof("[%s] multi agent run", s.AgentChatContext.AgentChatReq.AgentBaseParams.Name)
	return s.MultiAgent.Run(ctx, input, options...)
}

// searchMultiAgentConfig 查询多智能体配置信息
func searchMultiAgentConfig(ctx *gin.Context, req *request.MultiAgentChatParams) (*assistant_service.MultiAssistantDetailResp, error) {
	multiAgent, err := assistant.GetClient().GetMultiAssistantById(ctx, &assistant_service.GetMultiAssistantByIdReq{
		AssistantId:    req.MultiAgentId,
		ConversationId: req.ConversationId,
		Draft:          req.Draft,
		Identity: &assistant_service.Identity{
			UserId: req.UserId,
			OrgId:  req.OrgId,
		},
		FilterSubEnable: true,
	})
	if err != nil {
		log.Errorf("failed to get multi assistant by id: %v", err)
		return nil, errors.New("failed to get multi assistant")
	}
	return multiAgent, nil
}

func buildAgentChatParams(multiAgentChatReq *request.MultiAgentChatReq) *request.AgentChatParams {
	var subAgentInfoList []*request.SubAgentInfo
	agentList := multiAgentChatReq.AgentList
	if len(agentList) > 0 {
		for _, subAgent := range agentList {
			subAgentInfoList = append(subAgentInfoList, &request.SubAgentInfo{
				Name:        subAgent.AgentBaseParams.Name,
				Description: subAgent.AgentBaseParams.Description,
			})
		}
	}
	baseParams := multiAgentChatReq.AgentChatBaseParams
	return &request.AgentChatParams{
		DetailId:         multiAgentChatReq.DetailId,
		MultiAgent:       true,
		SubAgentInfoList: subAgentInfoList,
		Input:            multiAgentChatReq.Input,
		UploadFile:       multiAgentChatReq.UploadFile,
		Stream:           multiAgentChatReq.Stream,
		AgentChatBaseParams: request.AgentChatBaseParams{
			ModelParams:     baseParams.ModelParams,
			AgentBaseParams: baseParams.AgentBaseParams,
		},
	}
}

func buildMultiSubAgent(ctx *gin.Context, multiAgentChatReq *request.MultiAgentChatReq, agentInfo *request.AgentInfo) ([]adk.Agent, map[string]*request.AgentConfig, error) {
	var subAgents []adk.Agent
	var subAgentMap = make(map[string]*request.AgentConfig)
	for _, agentParams := range multiAgentChatReq.AgentList {
		subAgent, err := BaseCreateSingleAgent(ctx, &request.AgentChatParams{
			AgentChatBaseParams: *agentParams,
			Stream:              multiAgentChatReq.Stream,
			UploadFile:          multiAgentChatReq.UploadFile,
			DetailId:            multiAgentChatReq.DetailId,
			MultiSubAgent:       true,
		}, agentInfo)
		if err != nil {
			return nil, nil, err
		}
		subAgents = append(subAgents, subAgent)
		baseParams := agentParams.AgentBaseParams
		subAgentMap[baseParams.Name] = &request.AgentConfig{
			AgentId:          baseParams.AgentId,
			AgentName:        baseParams.Name,
			AgentAvatar:      baseParams.Avatar,
			AgentChatContext: subAgent.ChatContext, //如果同一智能体不出现并发没问题，如果可能触发同一智能体并发调用，就会有bug
		}
	}
	//处理skill包装成子智能体
	agentChatBaseParams := multiAgentChatReq.AgentChatBaseParams
	if agentChatBaseParams.ToolParams != nil && len(agentChatBaseParams.ToolParams.SkillToolList) != 0 {
		for _, skill := range agentChatBaseParams.ToolParams.SkillToolList {
			subAgent, err := skillToSingleAgent(ctx, multiAgentChatReq, skill)
			if err != nil {
				return nil, nil, err
			}
			subAgents = append(subAgents, subAgent)
			baseParams := subAgent.ChatContext.AgentChatReq.AgentBaseParams
			subAgentMap[baseParams.Name] = &request.AgentConfig{
				AgentId:          baseParams.AgentId,
				AgentName:        baseParams.Name,
				AgentAvatar:      baseParams.Avatar,
				AgentChatContext: subAgent.ChatContext, //如果同一智能体不出现并发没问题，如果可能触发同一智能体并发调用，就会有bug
			}
		}
	}
	return subAgents, subAgentMap, nil
}

// skillToSingleAgent 将一个skill包装成单智能体
func skillToSingleAgent(ctx *gin.Context, multiAgentChatReq *request.MultiAgentChatReq, skill *request.SkillToolInfo) (*SingleAgent, error) {
	agentChatBaseParams := multiAgentChatReq.AgentChatBaseParams
	return CreateSingleAgent(ctx, &request.AgentChatParams{
		AgentChatBaseParams: request.AgentChatBaseParams{
			AgentBaseParams: &request.AgentBaseParams{
				Name:        skill.Name,
				Description: skill.Desc,
			},
			ModelParams: agentChatBaseParams.ModelParams,
			ToolParams:  agentChatBaseParams.ToolParams,
		},
		Stream:     multiAgentChatReq.Stream,
		UploadFile: multiAgentChatReq.UploadFile,
		DetailId:   multiAgentChatReq.DetailId,
	})
}
