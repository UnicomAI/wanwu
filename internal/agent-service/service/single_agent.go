package service

import (
	"context"
	"encoding/json"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/grpc-consumer/consumer/assistant"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	agent_message_processor "github.com/UnicomAI/wanwu/internal/agent-service/service/agent-message-processor"
	agent_preprocessor "github.com/UnicomAI/wanwu/internal/agent-service/service/agent-preprocessor"
	agent_tool "github.com/UnicomAI/wanwu/internal/agent-service/service/agent-tool"
	local_agent "github.com/UnicomAI/wanwu/internal/agent-service/service/local-agent"
	service_model "github.com/UnicomAI/wanwu/internal/agent-service/service/service-model"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/gin-gonic/gin"
)

type SingleAgent struct {
	ChatContext     *request.AgentChatContext
	ChatModelAgent  *adk.ChatModelAgent
	Req             *request.AgentChatParams
	AgentPreprocess *agent_preprocessor.AgentPreprocess
}

// SingleAgentChat 单智能体问答
func SingleAgentChat(ctx *gin.Context, req *request.AgentChatReq) error {
	singleAgentDetail, err := searchSingleAgent(ctx, req)
	if err != nil {
		return err
	}
	return SingleAgentChatDirect(ctx, BuildAgentParams(req, singleAgentDetail, true))
}

// SingleAgentChatDirect 单智能体问答
func SingleAgentChatDirect(ctx *gin.Context, agentChatParams *request.AgentChatParams) error {
	agent, err := CreateSingleAgent(ctx, agentChatParams)
	if err != nil {
		return err
	}
	return agent.Chat(ctx)
}

// CreateSingleAgent 创建单智能体
func CreateSingleAgent(ctx *gin.Context, req *request.AgentChatParams) (*SingleAgent, error) {
	return BaseCreateSingleAgent(ctx, req, nil)
}

// BaseCreateSingleAgent 创建单智能体
func BaseCreateSingleAgent(ctx *gin.Context, req *request.AgentChatParams, agentInfo *request.AgentInfo) (*SingleAgent, error) {
	data, _ := json.Marshal(req)
	log.Infof("single agent chat req %s", string(data))
	chatInfo, err := buildAgentChatInfo(ctx, req)
	if err != nil {
		log.Errorf("failed to build chat info: %v", err)
		return nil, err
	}
	chatContext := &request.AgentChatContext{CurrentAgent: agentInfo}
	localAgentService := local_agent.CreateLocalAgentService(ctx, req, chatInfo, chatContext)
	//创建模型
	chatModel, err := localAgentService.CreateChatModel(ctx, req, chatInfo)
	if err != nil {
		log.Errorf("failed to create chat model: %v", err)
		return nil, err
	}
	//2.创建智能体
	agent, toolIDMap, err := createAgent(ctx, req, chatModel, chatInfo)
	if err != nil {
		log.Errorf("failed to create agent: %v", err)
		return nil, err
	}
	chatContext.ToolMap = toolIDMap
	return &SingleAgent{
		ChatModelAgent: agent,
		Req:            req,
		AgentPreprocess: &agent_preprocessor.AgentPreprocess{
			LocalAgentService: localAgentService,
			AgentChatInfo:     chatInfo,
			GinContext:        ctx,
			CallDetail:        true,
		},
		ChatContext: chatContext,
	}, nil
}

func (s *SingleAgent) Chat(ctx *gin.Context) error {
	//1.执行流式agent问答调用
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           s,
		EnableStreaming: s.Req.Stream,
	})
	iter := runner.Query(ctx, s.Req.Input)

	//2.处理结果
	_, err := agent_message_processor.AgentMessage(ctx, iter, &request.AgentChatContext{AgentChatReq: s.Req,
		KnowledgeHitData: s.ChatContext.KnowledgeHitData, ToolMap: s.ChatContext.ToolMap, Order: s.ChatContext.Order, CurrentAgent: s.ChatContext.CurrentAgent})
	return err
}

func (s *SingleAgent) Name(ctx context.Context) string {
	return s.ChatModelAgent.Name(ctx)
}
func (s *SingleAgent) Description(ctx context.Context) string {
	return s.ChatModelAgent.Description(ctx)
}

func (s *SingleAgent) Run(ctx context.Context, input *adk.AgentInput, options ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	log.Infof("[%s] single agent run", s.Req.AgentBaseParams.Name)
	//参数预处理
	process, chatContext := agent_preprocessor.AgentPreProcess(s.AgentPreprocess, input, s.Req, s.ChatContext.CurrentAgent)
	s.ChatContext.Order = chatContext.Order
	return s.ChatModelAgent.Run(ctx, process, options...)
}

func (s *SingleAgent) Resume(ctx context.Context, info *adk.ResumeInfo, opts ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	return s.ChatModelAgent.Resume(ctx, info, opts...)
}

func (s *SingleAgent) OnSetSubAgents(ctx context.Context, subAgents []adk.Agent) error {
	return s.ChatModelAgent.OnSetSubAgents(ctx, subAgents)
}

func (s *SingleAgent) OnSetAsSubAgent(ctx context.Context, parent adk.Agent) error {
	return s.ChatModelAgent.OnSetAsSubAgent(ctx, parent)
}

func (s *SingleAgent) OnDisallowTransferToParent(ctx context.Context) error {
	return s.ChatModelAgent.OnDisallowTransferToParent(ctx)
}

// buildAgentChatInfo 构建智能体信息
func buildAgentChatInfo(ctx *gin.Context, req *request.AgentChatParams) (*service_model.AgentChatInfo, error) {
	modelInfo, err := SearchModel(ctx, req.ModelParams.ModelId)
	if err != nil {
		return nil, err
	}
	var functionCall = modelInfo.Config.FunctionCalling != "noSupport"
	var vision = modelInfo.Config.VisionSupport == "support"
	imageFile, err := buildUploadFileType(req.UploadFile)
	if err != nil {
		return nil, err
	}
	return &service_model.AgentChatInfo{
		FunctionCalling: functionCall,
		VisionSupport:   vision,
		UploadUrl:       len(req.UploadFile) > 0,
		ImageUpload:     imageFile,
		ModelInfo:       modelInfo,
	}, nil
}

// buildUploadFileType 构建图片上传
func buildUploadFileType(uploadFile []string) (imageFile bool, err error) {
	if len(uploadFile) == 0 {
		return false, nil
	}
	for _, fileUrl := range uploadFile {
		fileName := util.ExtractFileNameFromURL(fileUrl)
		if util.ImageFile(fileName) {
			imageFile = true
			break
		}
	}
	return imageFile, nil
}

// searchSingleAgent 查询智能体详情
func searchSingleAgent(ctx *gin.Context, req *request.AgentChatReq) (*assistant_service.AssistantDetailResp, error) {
	return assistant.GetClient().GetAssistantDetailById(ctx, &assistant_service.GetAssistantDetailByIdReq{
		AssistantId:    req.AssistantId,
		ConversationId: req.ConversationId,
		Draft:          req.Draft,
		Identity: &assistant_service.Identity{
			UserId: req.UserId,
			OrgId:  req.OrgId,
		},
	})
}

// 创建对应智能体
func createAgent(ctx *gin.Context, req *request.AgentChatParams, chatModel model.ToolCallingChatModel, chatInfo *service_model.AgentChatInfo) (*adk.ChatModelAgent, map[string]*request.ToolConfig, error) {
	baseParams := req.AgentBaseParams
	toolsConfig, toolMap, err := agent_tool.BuildAgentToolsConfig(ctx, req, chatInfo)
	if err != nil {
		return nil, nil, err
	}
	var exit tool.BaseTool
	if req.MultiAgent {
		exit = &adk.ExitTool{}
	}
	////配置全局追踪
	//trace_util.EnioGlobalTracing()
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Model:       chatModel,
		Name:        baseParams.Name,
		Description: baseParams.Description,
		ToolsConfig: toolsConfig,
		Exit:        exit,
		Middlewares: []adk.AgentMiddleware{agent_tool.NewToolResultFilterMiddleware()},
	})
	return agent, toolMap, err
}
