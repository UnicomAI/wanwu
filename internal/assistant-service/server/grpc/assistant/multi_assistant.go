package assistant

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/internal/assistant-service/service"
	params_process "github.com/UnicomAI/wanwu/internal/assistant-service/service/params-process"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"google.golang.org/grpc"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetMultiAssistantById(ctx context.Context, req *assistant_service.GetMultiAssistantByIdReq) (*assistant_service.MultiAssistantDetailResp, error) {
	agent, agentSnapshot, subAgents, err := s.cli.GetMultiAssistant(ctx, req.AssistantId, req.Identity.GetUserId(), req.Identity.GetOrgId(), req.Draft, req.Version, req.FilterSubEnable)
	if err != nil {
		return nil, err
	}
	multiAgentParams, err := buildMultiAgentParams(ctx, s.cli, agent, agentSnapshot, req)
	if err != nil {
		return nil, err
	}
	var subParamsList []*assistant_service.AgentDetail
	for _, subAgent := range subAgents {
		params, err := buildSubAgentParams(ctx, s.cli, agentSnapshot, subAgent)
		if err != nil {
			return nil, err
		}
		subParamsList = append(subParamsList, params)
	}
	return &assistant_service.MultiAssistantDetailResp{
		MultiAgent: multiAgentParams,
		SubAgents:  subParamsList,
	}, nil
}

func (s *Service) MultiAssistantConversionStream(req *assistant_service.MultiAssistantConversionStreamReq, stream grpc.ServerStreamingServer[assistant_service.AssistantConversionStreamResp]) error {
	// 刷新会话 updated_at，使会话列表按最近聊天排序
	if convID := util.MustU32(req.ConversationId); convID > 0 {
		if status := s.cli.UpdateConversation(stream.Context(), &model.Conversation{ID: convID}); status != nil {
			log.Errorf("[Conversation] touch conversation updated_at failed, id: %s, err: %v", req.ConversationId, status)
		}
	}
	//会话处理
	conversationProcessor := &service.ConversationProcessor{
		SSEWriter: sse_util.NewGrpcSSEWriter(stream, "MultiAssistantConversionStreamNew", nil),
	}
	err := conversationProcessor.Process(stream.Context(), buildMultiConversationParams(req), buildMultiAgentSendRequest(req))
	if err != nil {
		log.Errorf("Assistant服务处理智能体流式对话失败，assistantId: %s, error: %v", req.AssistantId, err)
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "agent服务异常")
	}
	return nil
}

func buildMultiAgentParams(ctx context.Context, cli client.IClient, agent *model.Assistant, snapshot *model.AssistantSnapshot, req *assistant_service.GetMultiAssistantByIdReq) (*assistant_service.AgentDetail, error) {
	clientInfo := &params_process.ClientInfo{
		Cli:       cli,
		Knowledge: Knowledge,
		MCP:       MCP,
	}
	//传入了 ConversationId就会尝试构造历史数据
	userQueryParams := &params_process.UserQueryParams{
		ConversationId: req.ConversationId,
		QueryOrgId:     req.Identity.OrgId,
		QueryUserId:    req.Identity.UserId,
		Ctx:            ctx,
	}
	return service.NewAgentChatParamsBuilder(&params_process.AgentInfo{
		Assistant:         agent,
		AssistantSnapshot: snapshot,
		Draft:             snapshot == nil,
	}, userQueryParams, clientInfo).
		AgentBaseParams().
		ModelParams().
		Build()
}

// buildSubAgentParams 构建子智能体参数，由于子智能体只能选发布后的智能体所以结构体是snapshot
func buildSubAgentParams(ctx context.Context, cli client.IClient, multiAgentSnapshot *model.AssistantSnapshot, agentSnapshot *model.AssistantSnapshot) (*assistant_service.AgentDetail, error) {
	clientInfo := &params_process.ClientInfo{
		Cli:       cli,
		Knowledge: Knowledge,
		MCP:       MCP,
	}
	var assistant = &model.Assistant{}
	if err := jsonToStruct(agentSnapshot.AssistantInfo, assistant); err != nil {
		log.Errorf("转换智能体信息失败，assistantId: %d, error: %v", agentSnapshot.AssistantID, err)
		return nil, errors.New("build agent info err")
	}
	//兼容多智能体,知识库状态隐藏功能
	if multiAgentSnapshot != nil && multiAgentSnapshot.SnapshotExtra != "" {
		agentSnapshot.SnapshotExtra = multiAgentSnapshot.SnapshotExtra
	}
	return service.NewAgentChatParamsBuilder(&params_process.AgentInfo{
		Assistant:         assistant,
		AssistantSnapshot: agentSnapshot,
		Draft:             agentSnapshot == nil,
	}, &params_process.UserQueryParams{Ctx: ctx}, clientInfo).
		AgentBaseParams().
		ModelParams().
		KnowledgeParams().
		ToolParams().
		SkillParams().
		Build()
}

func buildMultiConversationParams(req *assistant_service.MultiAssistantConversionStreamReq) *service.ConversationParams {
	return &service.ConversationParams{
		AssistantId:    req.AssistantId,
		ConversationId: req.ConversationId,
		FileInfo:       extractFileInfos(req.FileInfo),
		OrgId:          req.Identity.OrgId,
		Query:          req.Prompt,
		UserId:         req.Identity.UserId,
		DetailId:       req.DetailId,
	}
}

// buildMultiAgentSendRequest 构建底层智能体能力接口请求体
func buildMultiAgentSendRequest(req *assistant_service.MultiAssistantConversionStreamReq) func(ctx context.Context) (string, *http.Response, context.CancelFunc, error) {
	var conversationID string
	// 历史聊天记录配置
	if req.ConversationId != "" {
		conversationID = req.ConversationId
	}
	// 底层智能体能力接口请求体
	chatReq := service.BuildMultiAgentChatReq(&service.AgentUserInputParams{
		Input:          req.Prompt,
		Stream:         true,
		UploadFile:     extractFileUrls(req.FileInfo),
		ConversationId: conversationID,
		UserId:         req.Identity.UserId,
		OrgId:          req.Identity.OrgId,
		Draft:          req.Draft,
		DetailId:       req.DetailId,
	}, util.MustU32(req.AssistantId))
	var monitorKey = "multi_agent_chat_service"

	return func(ctx context.Context) (string, *http.Response, context.CancelFunc, error) {
		paramsBytes, err := json.Marshal(chatReq)
		if err != nil {
			return monitorKey, nil, nil, err
		}
		// 获取Assistant配置
		assistantConfig := config.Cfg().Assistant
		if assistantConfig.MultiAgentChatUrl == "" {
			return monitorKey, nil, nil, errors.New("多智能体会话URL配置错误")
		}
		params := &http_client.HttpRequestParams{
			Body:       paramsBytes,
			Timeout:    15 * time.Minute,
			Url:        assistantConfig.MultiAgentChatUrl,
			MonitorKey: monitorKey,
			LogLevel:   http_client.LogAll,
		}
		ctx, cancel := context.WithTimeout(ctx, params.Timeout)
		result, err := http_client.Default().PostJsonOriResp(ctx, params)
		if err == nil {
			err = readHttpErr(result)
		}
		return monitorKey, result, cancel, err
	}
}

func (s *Service) MultiAgentCreate(ctx context.Context, req *assistant_service.MultiAgentCreateReq) (*empty.Empty, error) {
	// 获取已发布的子智能体详情
	subAgent, err := s.cli.GetAssistantSnapshot(ctx, util.MustU32(req.AgentId), "")
	if err != nil {
		return nil, errStatus(errs.Code_AssistantErr, err)
	}

	snapshotAssistant := &model.Assistant{}
	if err := jsonToStruct(subAgent.AssistantInfo, &snapshotAssistant); err != nil {
		return nil, errStatus(errs.Code_AssistantErr, toErrStatus("assistant_snapshot", err.Error()))
	}

	// 判断是否已有重复子智能体
	_, err = s.cli.FetchMultiAssistantRelationFirst(ctx, util.MustU32(req.AssistantId), subAgent.AssistantID)
	if err == nil {
		return nil, errStatus(errs.Code_AssistantMultiAgentErr, &errs.Status{
			TextKey: "assistant_multi_agent_repeat",
			Args:    nil,
		})
	}

	// 组装multiAgent参数
	newMultiAgent := &model.MultiAgentRelation{
		MultiAgentId: util.MustU32(req.AssistantId),
		AgentId:      subAgent.AssistantID,
		Description:  snapshotAssistant.Desc,
		Enable:       true,
		UserId:       req.Identity.UserId,
		OrgId:        req.Identity.OrgId,
	}

	// 调用client方法创建多智能体
	if status := s.cli.CreateMultiAssistantRelation(ctx, newMultiAgent); status != nil {
		return nil, errStatus(errs.Code_AssistantMultiAgentErr, status)
	}
	return &empty.Empty{}, nil
}

func (s *Service) MultiAgentDelete(ctx context.Context, req *assistant_service.MultiAgentCreateReq) (*empty.Empty, error) {
	if status := s.cli.DeleteMultiAssistantRelation(ctx, util.MustU32(req.AssistantId), util.MustU32(req.AgentId)); status != nil {
		return nil, errStatus(errs.Code_AssistantMultiAgentErr, status)
	}
	return &empty.Empty{}, nil
}

func (s *Service) MultiAgentEnableSwitch(ctx context.Context, req *assistant_service.MultiAgentEnableSwitchReq) (*empty.Empty, error) {
	// 获取多智能体详情
	multiAgent, err := s.cli.FetchMultiAssistantRelationFirst(ctx, util.MustU32(req.AssistantId), util.MustU32(req.AgentId))
	if err != nil {
		return nil, errStatus(errs.Code_AssistantMultiAgentErr, err)
	}
	multiAgent.Enable = req.Enable
	if status := s.cli.UpdateMultiAssistantRelation(ctx, multiAgent); status != nil {
		return nil, errStatus(errs.Code_AssistantMultiAgentErr, status)
	}
	return &empty.Empty{}, nil
}

func (s *Service) MultiAgentConfigUpdate(ctx context.Context, req *assistant_service.MultiAgentConfigUpdateReq) (*empty.Empty, error) {
	// 获取多智能体详情
	multiAgent, err := s.cli.FetchMultiAssistantRelationFirst(ctx, util.MustU32(req.AssistantId), util.MustU32(req.AgentId))
	if err != nil {
		return nil, errStatus(errs.Code_AssistantMultiAgentErr, err)
	}
	multiAgent.Description = req.Desc
	if status := s.cli.UpdateMultiAssistantRelation(ctx, multiAgent); status != nil {
		return nil, errStatus(errs.Code_AssistantMultiAgentErr, status)
	}
	return &empty.Empty{}, nil
}
