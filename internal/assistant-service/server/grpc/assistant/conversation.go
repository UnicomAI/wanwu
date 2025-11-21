package assistant

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	net_url "net/url"
	"os"
	"strconv"
	"strings"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	knowledgebase_service "github.com/UnicomAI/wanwu/api/proto/knowledgebase-service"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/es"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	pkgUtil "github.com/UnicomAI/wanwu/pkg/util"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	metaTypeNumber = "number"
	metaTypeTime   = "time"
)

// ConversationCreate Create a conversation
func (s *Service) ConversationCreate(ctx context.Context, req *assistant_service.ConversationCreateReq) (*assistant_service.ConversationCreateResp, error) {
	// Assemble model parameters
	assistantID, err := pkgUtil.U32(req.AssistantId)
	if err != nil {
		return nil, err
	}

	conversation := &model.Conversation{
		AssistantId: assistantID,
		Title:       req.Prompt, // Use prompt as initial title
		UserId:      req.Identity.UserId,
		OrgId:       req.Identity.OrgId,
	}

	// Call the client method to create a conversation
	if status := s.cli.CreateConversation(ctx, conversation); status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	return &assistant_service.ConversationCreateResp{
		ConversationId: strconv.FormatUint(uint64(conversation.ID), 10),
	}, nil
}

// ConversationDelete Delete conversation
func (s *Service) ConversationDelete(ctx context.Context, req *assistant_service.ConversationDeleteReq) (*emptypb.Empty, error) {
	// Conversion ID
	conversationID, err := strconv.ParseUint(req.ConversationId, 10, 32)
	if err != nil {
		return nil, err
	}

	// Call the client method to delete the conversation
	if status := s.cli.DeleteConversation(ctx, uint32(conversationID)); status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	return &emptypb.Empty{}, nil
}

// GetConversationList conversation list
func (s *Service) GetConversationList(ctx context.Context, req *assistant_service.GetConversationListReq) (*assistant_service.GetConversationListResp, error) {
	// Calculate offset
	offset := (req.PageNo - 1) * req.PageSize

	// Call the client method to get the conversation list
	conversations, total, status := s.cli.GetConversationList(ctx, req.AssistantId, req.Identity.UserId, req.Identity.OrgId, offset, req.PageSize)
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}

	// Convert to responsive format
	var conversationInfos []*assistant_service.ConversationInfo
	for _, conversation := range conversations {
		conversationInfos = append(conversationInfos, &assistant_service.ConversationInfo{
			ConversationId: strconv.FormatUint(uint64(conversation.ID), 10),
			AssistantId:    strconv.FormatUint(uint64(conversation.AssistantId), 10),
			Title:          conversation.Title,
			CreatTime:      conversation.CreatedAt,
		})
	}

	return &assistant_service.GetConversationListResp{
		Data:     conversationInfos,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

// GetConversationDetailList Conversation details history list
func (s *Service) GetConversationDetailList(ctx context.Context, req *assistant_service.GetConversationDetailListReq) (*assistant_service.GetConversationDetailListResp, error) {
	// Calculate paging parameters
	from := (req.PageNo - 1) * req.PageSize
	size := int(req.PageSize)

	// Assemble query conditions
	fieldConditions := map[string]interface{}{
		"conversationId": req.ConversationId,
		"userId.keyword": req.Identity.UserId,
		"orgId.keyword":  req.Identity.OrgId,
	}

	// Query all conversation details index using wildcards
	indexPattern := "conversation_detail_infos_*"

	// Query data from ES
	documents, total, err := es.Assistant().SearchByFields(ctx, indexPattern, fieldConditions, int(from), size)
	if err != nil {
		log.Errorf("从ES查询对话详情失败，conversationId: %s, userId: %s, error: %v", req.ConversationId, req.Identity.UserId, err)
		return nil, fmt.Errorf("查询对话详情失败: %v", err)
	}

	// Convert query results to response format
	var conversationDetails []*assistant_service.ConversionDetailInfo
	for _, doc := range documents {
		var detail model.ConversationDetails
		if err := json.Unmarshal(doc, &detail); err != nil {
			log.Warnf("解析ES文档失败: %v", err)
			continue
		}

		conversationDetails = append(conversationDetails, &assistant_service.ConversionDetailInfo{
			Id:             detail.Id,
			AssistantId:    detail.AssistantId,
			ConversationId: detail.ConversationId,
			Prompt:         detail.Prompt,
			SysPrompt:      detail.SysPrompt,
			Response:       detail.Response,
			SearchList:     detail.SearchList,
			QaType:         detail.QaType,
			CreatedBy:      detail.UserId, // Mapping UserId using CreatedBy field
			CreatedAt:      detail.CreatedAt,
			UpdatedAt:      detail.UpdatedAt,
			RequestFiles:   transRequestFiles(detail.FileInfo),
			FileSize:       detail.FileSize,
			FileName:       detail.FileName,
		})
	}

	log.Infof("成功从ES查询对话详情，conversationId: %s, userId: %s, 总数: %d, 返回: %d",
		req.ConversationId, req.Identity.UserId, total, len(conversationDetails))

	return &assistant_service.GetConversationDetailListResp{
		Data:     conversationDetails,
		Total:    total,
		PageSize: req.PageSize,
		PageNo:   req.PageNo,
	}, nil
}

// AssistantConversionStream agent streaming conversation
func (s *Service) AssistantConversionStream(req *assistant_service.AssistantConversionStreamReq, stream assistant_service.AssistantService_AssistantConversionStreamServer) error {
	ctx := stream.Context()
	reqUserId := req.Identity.UserId
	log.Debugf("Assistant服务开始智能体流式对话，assistantId: %s, userId: %s, orgId: %s, conversationId: %s, fileInfo: %+v, trial: %v, prompt: %s",
		req.AssistantId, reqUserId, req.Identity.OrgId, req.ConversationId, req.FileInfo, req.Trial, req.Prompt)

	// Variables used to track the status of streaming responses
	var fullResponse strings.Builder
	var searchList string
	var hasReadFirstMessage bool
	var streamStarted bool
	var conversationSaved bool // Mark whether the conversation has been saved

	// Use defer to uniformly handle context cancellation situations
	defer func() {
		// The "Terminated" message is only saved if the context is manually canceled and the conversation has not been saved.
		if ctx.Err() != nil && !req.Trial && !conversationSaved {
			var terminationMessage string

			if !streamStarted {
				// Streaming response has not started yet, save basic termination message
				terminationMessage = "本次回答已被终止"
			} else if !hasReadFirstMessage || fullResponse.Len() == 0 {
				// Streaming response started but no valid content
				terminationMessage = "本次回答已被终止"
			} else {
				// There is already some response content, save the received content
				terminationMessage = fullResponse.String() + "\n本次回答已被终止"
			}

			saveConversation(ctx, req, terminationMessage, searchList)
			log.Infof("因上下文取消保存终止消息，assistantId: %s, conversationId: %s", req.AssistantId, req.ConversationId)
		}
	}()

	// Query agent information based on agent id
	assistantID, err := strconv.ParseUint(req.AssistantId, 10, 32)
	if err != nil {
		log.Errorf("Assistant服务智能体ID转换失败，assistantId: %s, error: %v", req.AssistantId, err)
		return err
	}

	assistant, status := s.cli.GetAssistant(ctx, uint32(assistantID), "", "")
	if status != nil {
		log.Errorf("Assistant服务获取智能体信息失败，assistantId: %s, error: %v", req.AssistantId, status)
		SSEError(stream, "智能体信息获取失败")
		saveConversation(ctx, req, "智能体信息获取失败", "")
		return errStatus(errs.Code_AssistantConversationErr, status)
	}

	log.Debugf("Assistant服务获取到智能体信息，assistantId: %s, 名称: %s, Scope: %d, userId: %s, orgId: %s",
		req.AssistantId, assistant.Name, assistant.Scope, assistant.UserId, assistant.OrgId)

	// Get Assistant configuration
	assistantConfig := config.Cfg().Assistant
	if assistantConfig.SseUrl == "" {
		log.Errorf("Assistant服务SSE URL配置为空，assistantId: %s", req.AssistantId)
		SSEError(stream, "智能体SSE URL配置错误")
		saveConversation(ctx, req, "智能体SSE URL配置错误", "")
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "SSE URL配置错误")
	}

	// Assemble the agent capability interface request body
	sseReq := &config.AgentSSERequest{
		Input:        req.Prompt,
		Stream:       true,
		AutoCitation: true,
	}

	if assistant.Instructions != "" {
		sseReq.SystemRole = assistant.Instructions
	}

	sseReq.UploadFileUrl = extractFileUrls(req.FileInfo)

	// Model parameter configuration
	_, err = s.setModelConfigParams(sseReq, assistant)
	if err != nil {
		SSEError(stream, "智能体模型配置解析失败")
		saveConversation(ctx, req, "智能体模型配置解析失败", "")
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "模型配置解析失败")
	}

	// Knowledge base parameter configuration
	if err = s.setKnowledgebaseParams(ctx, sseReq, req, assistant); err != nil {
		SSEError(stream, "智能体知识库配置解析失败")
		saveConversation(ctx, req, "智能体知识库配置解析失败", "")
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "知识库配置解析失败")
	}

	// plugin parameter configuration
	if err := s.setToolAndWorkflowParams(ctx, sseReq, req.AssistantId, req.Identity); err != nil {
		SSEError(stream, "智能体plugin配置错误")
		saveConversation(ctx, req, "智能体plugin配置错误", "")
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "plugin配置错误")
	}

	// MCP information parameter configuration
	if err = s.setMCPParams(ctx, sseReq, assistant); err != nil {
		SSEError(stream, "智能体MCP配置解析失败")
		saveConversation(ctx, req, "智能体MCP配置解析失败", "")
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "MCP配置解析失败")
	}

	// Historical chat record configuration
	if !req.Trial && req.ConversationId != "" {
		s.setHistoryParams(ctx, sseReq, req)
	}

	// The underlying agent capability interface request body
	var requestBody map[string]interface{}
	reqBytes, err := json.Marshal(sseReq)
	if err != nil {
		log.Errorf("Assistant服务序列化请求体失败，assistantId: %s, error: %v", req.AssistantId, err)
		SSEError(stream, "请求参数错误")
		saveConversation(ctx, req, "请求参数错误", "")
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "请求参数错误")
	}
	if err = json.Unmarshal(reqBytes, &requestBody); err != nil {
		log.Errorf("Assistant服务反序列化请求体到map失败，assistantId: %s, error: %v", req.AssistantId, err)
		SSEError(stream, "请求参数错误")
		saveConversation(ctx, req, "请求参数错误", "")
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "请求参数错误")
	}

	// Merge dynamic model parameters
	if sseReq.ModelParams != nil {
		requestBody = mergeMaps(requestBody, sseReq.ModelParams)
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		log.Errorf("Assistant服务序列化最终请求体失败，assistantId: %s, error: %v", req.AssistantId, err)
		SSEError(stream, "请求参数错误")
		saveConversation(ctx, req, "请求参数错误", "")
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "请求参数错误")
	}

	timeout := 300 * time.Second
	startTime := time.Now()
	id := uuid.New().String()

	// xuid is passed to RAG through the agent for use. It is required that xuid is consistent with the userId of the knowledge base creator. The userId of the current version of the agent is consistent with the userId of the knowledge base creator. This may need to be modified after the knowledge base is shared later.
	xuid := assistant.UserId

	log.Infof("Assistant服务开始调用HttpRequestLlmStream，uuid: %s, assistantId: %s, url: %s, userId: %s, timeout: %v, body: %s",
		id, req.AssistantId, assistantConfig.SseUrl, reqUserId, timeout, string(requestBodyBytes))
	sseResp, err := HttpRequestLlmStream(ctx, assistantConfig.SseUrl, reqUserId, xuid, bytes.NewReader(requestBodyBytes), timeout)
	if err != nil {
		log.Errorf("Assistant服务调用智能体能力接口失败，assistantId: %s, uuid: %s, error: %v", req.AssistantId, id, err)
		if ctx.Err() == nil { //non-context canceled
			SSEError(stream, "agent服务异常")
			saveConversation(ctx, req, "agent服务异常", "")
		}
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "agent服务异常")
	}
	defer sseResp.Body.Close()
	log.Infof("Assistant服务成功连接智能体能力接口，uuid: %s, assistantId: %s, statusCode: %d, time: %v毫秒", id, req.AssistantId, sseResp.StatusCode, time.Since(startTime).Milliseconds())

	// SSE request returns Code greater than 400
	if sseResp.StatusCode > http.StatusBadRequest {
		log.Errorf("Assistant服务智能体能力接口返回错误状态码，assistantId: %s, statusCode: %d", req.AssistantId, sseResp.StatusCode)
		SSEError(stream, "agent服务异常")
		saveConversation(ctx, req, "agent服务异常", "")
		return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "agent服务异常")
	}

	// Read the agent interface return and write the streaming response
	reader := bufio.NewReader(sseResp.Body)
	lineCount := 0
	streamStarted = true
	searchListExtracted := false
	for {
		// Check context
		if ctx.Err() != nil {
			log.Infof("Assistant服务检测到上下文取消，assistantId: %s", req.AssistantId)
			return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "智能体问答上下文异常")
		}
		line, err := reader.ReadBytes('\n')
		if err != nil && err == io.EOF { //End normally
			// Q&A debugging is not saved
			if !req.Trial {
				// Only save and mark as saved if the context has not been canceled
				if ctx.Err() == nil {
					saveConversation(ctx, req, fullResponse.String(), searchList)
					conversationSaved = true // Mark saved
				}
				// If the context is canceled, do not set conversationSaved and let the defer function handle the termination message.
			}
			log.Debugf("Assistant服务流式响应正常结束，assistantId: %s, 总处理行数: %d", req.AssistantId, lineCount)
			return nil
		}
		if err != nil && err == io.ErrUnexpectedEOF { //abnormal end
			// Real SSE read error, saving "interrupted" message
			log.Errorf("Assistant服务读取流式响应失败，assistantId: %s, error: %v, 已处理行数: %d", req.AssistantId, err, lineCount)
			if !req.Trial {
				errorMessage := "本次回答已中断"
				if hasReadFirstMessage && fullResponse.Len() > 0 {
					errorMessage = fullResponse.String() + "\n" + errorMessage
				}
				saveConversation(ctx, req, errorMessage, searchList)
				conversationSaved = true // Mark saved to avoid repeated saving in defer
				log.Debugf("Assistant服务保存了中断消息，assistantId: %s, errorMessage: %s", req.AssistantId, errorMessage)
			}
			SSEError(stream, "本次回答已中断")
			return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "本次回答已中断")
		}
		strLine := string(line)
		lineCount++
		if len(strLine) >= 5 && strLine[:5] == "data:" {
			jsonStrData := strLine[5:]
			// Parse streaming data and extract response field and search_list
			var streamData map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStrData), &streamData); err == nil {
				log.Debugf("Assistant服务解析流式数据，assistantId: %s, streamData: %+v", req.AssistantId, streamData)
				code, ok := extractCodeFromStreamData(streamData)
				if !ok {
					log.Errorf("Assistant服务无法提取code字段，assistantId: %s, streamData: %+v", req.AssistantId, streamData)
					continue
				}
				switch code {
				case 0:
					if response, ok := streamData["response"].(string); ok && response != "" {
						fullResponse.WriteString(response)
					}
					// Extract the first search_list
					if !searchListExtracted {
						if searchListData, ok := streamData["search_list"]; ok {
							searchListBytes, err := json.Marshal(searchListData)
							if err == nil {
								searchList = string(searchListBytes)
								searchListExtracted = true
								log.Debugf("Assistant服务提取到search_list，assistantId: %s, searchList: %s", req.AssistantId, searchList)
							}
						}
					}
				case 1:
					if message, ok := streamData["message"].(string); ok && message != "" {
						fullResponse.WriteString(message)
						if err := stream.Send(&assistant_service.AssistantConversionStreamResp{
							Content: "{\"code\":1,\"message\":\"" + "智能体无法回答：" + message + "\",\"finish\":1}",
						}); err != nil {
							log.Errorf("Assistant服务发送流式响应失败，assistantId: %s, error: %v", req.AssistantId, err)
							return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "assistant服务异常")
						}
						// The tag has read and returned the first valid message
						if !hasReadFirstMessage {
							hasReadFirstMessage = true
						}
						continue
					}
				}
			}
			if err := stream.Send(&assistant_service.AssistantConversionStreamResp{
				Content: jsonStrData,
			}); err != nil {
				log.Errorf("Assistant服务发送流式响应失败，assistantId: %s, error: %v", req.AssistantId, err)
				return grpc_util.ErrorStatusWithKey(errs.Code_AssistantConversationErr, "assistant_conversation", "assistant服务异常")
			}
			// The tag has read and returned the first valid message
			if !hasReadFirstMessage {
				hasReadFirstMessage = true
			}
		}
	}
}

// Set model configuration parameters
func (s *Service) setModelConfigParams(sseReq *config.AgentSSERequest, assistant *model.Assistant) (*common.AppModelConfig, error) {
	if assistant.ModelConfig == "" {
		log.Warnf("Assistant服务智能体模型配置为空，assistantId: %s", assistant.ID)
		return nil, nil
	}

	log.Debugf("Assistant服务解析模型配置，assistantId: %s, modelConfig: %s", assistant.ID, assistant.ModelConfig)
	modelConfig := &common.AppModelConfig{}
	if err := json.Unmarshal([]byte(assistant.ModelConfig), modelConfig); err != nil {
		return nil, fmt.Errorf("Assistant服务解析智能体模型配置失败，assistantId: %d, error: %v, modelConfigRaw: %s", assistant.ID, err, assistant.ModelConfig)
	}
	sseReq.ModelId = modelConfig.ModelId
	log.Debugf("Assistant服务成功解析智能体模型配置，assistantId: %s, provider: %s, model: %s, modelId: %s, modelType: %s",
		assistant.ID, modelConfig.Provider, modelConfig.Model, modelConfig.ModelId, modelConfig.ModelType)

	modelEndpoint := mp.ToModelEndpoint(modelConfig.ModelId, modelConfig.Model)
	log.Debugf("Assistant服务生成模型端点，assistantId: %s, modelEndpoint: %+v", assistant.ID, modelEndpoint)
	sseReq.Model = modelEndpoint["model"].(string)
	sseReq.ModelUrl = modelEndpoint["model_url"].(string)

	_, modelParams, _ := mp.ToModelParams(modelConfig.Provider, modelConfig.ModelType, modelConfig.Config)
	log.Debugf("Assistant服务生成模型参数，assistantId: %s, modelParams: %+v", assistant.ID, modelParams)
	if modelParams != nil {
		sseReq.ModelParams = modelParams
	}

	return modelConfig, nil
}

// Set knowledge base parameters
func (s *Service) setKnowledgebaseParams(ctx context.Context, sseReq *config.AgentSSERequest, req *assistant_service.AssistantConversionStreamReq, assistant *model.Assistant) error {
	knowledgeBaseConfig := &RAGKnowledgeBaseConfig{}
	if assistant.KnowledgebaseConfig == "" {
		return nil
	}

	if err := json.Unmarshal([]byte(assistant.KnowledgebaseConfig), knowledgeBaseConfig); err != nil {
		log.Errorf("Assistant服务解析智能体知识库配置失败，assistantId: %s, error: %v, knowledgebaseConfigRaw: %s", req.AssistantId, err, assistant.KnowledgebaseConfig)
		return err
	}
	log.Debugf("Assistant服务解析知识库成功，knowledgeBaseConfig: %+v", knowledgeBaseConfig)

	if len(knowledgeBaseConfig.KnowledgeBaseIds) > 0 {
		rerankEndpoint, err := buildRerank(req, knowledgeBaseConfig, assistant)
		if err != nil {
			return err
		}
		knowledgeInfoList, err := Knowledge.SelectKnowledgeDetailByIdList(ctx, &knowledgebase_service.KnowledgeDetailSelectListReq{
			KnowledgeIds: knowledgeBaseConfig.KnowledgeBaseIds,
		})
		if err != nil {
			log.Errorf("Assistant服务获取知识库详情失败, err: %v", err)
			return err
		}
		log.Infof("knowledgeInfoList = %+v", knowledgeInfoList)

		var knowNames []string
		for _, v := range knowledgeInfoList.List {
			knowNames = append(knowNames, v.Name)
		}

		params, err := buildMetaDataFilterParams(knowledgeBaseConfig.AppKnowledgeBaseList)
		if err != nil {
			log.Errorf("Assistant buildMetaDataFilterParams, err: %v", err)
			return err
		}
		sseReq.KnParams = &config.KnParams{
			KnowledgeBase:        knowNames,
			KnowledgeIdList:      knowledgeBaseConfig.KnowledgeBaseIds,
			RerankId:             rerankEndpoint["model_id"],
			Model:                rerankEndpoint["model"],
			ModelUrl:             rerankEndpoint["model_url"],
			RerankMod:            buildRerankMod(knowledgeBaseConfig.PriorityMatch),
			RetrieveMethod:       buildRetrieveMethod(knowledgeBaseConfig.MatchType),
			Weights:              buildWeight(knowledgeBaseConfig),
			MaxHistory:           knowledgeBaseConfig.MaxHistory,
			Threshold:            knowledgeBaseConfig.Threshold,
			TopK:                 knowledgeBaseConfig.TopK,
			RewriteQuery:         true,
			TermWeight:           buildTermWeight(knowledgeBaseConfig),
			MetaFilter:           len(params) > 0,
			MetaFilterConditions: params,
			UseGraph:             knowledgeBaseConfig.UseGraph,
		}
		sseReq.UseKnow = true
	}
	return nil
}

// Setup tools (custom tools, built-in tools, and workflows)
func (s *Service) setToolAndWorkflowParams(ctx context.Context, sseReq *config.AgentSSERequest, assistantId string, identity *assistant_service.Identity) error {
	toolPluginList, err := s.buildToolPluginListAlgParam(ctx, sseReq, assistantId, identity)
	if err != nil {
		return fmt.Errorf("智能体tool配置错误: %w", err)
	}

	workflowPluginList, err := s.buildWorkflowPluginListAlgParam(ctx, assistantId)
	if err != nil {
		return fmt.Errorf("智能体workflow配置错误: %w", err)
	}

	log.Debugf("智能体workflow配置，assistantId: %s, workflowPluginList: %s", assistantId, workflowPluginList)
	allPlugin := append(toolPluginList, workflowPluginList...)
	sseReq.PluginList = allPlugin
	log.Debugf("智能体tool_plugin_list，assistantId: %s, tool_plugin_list: %s", assistantId, allPlugin)
	return nil
}

// Set MCP parameters
func (s *Service) setMCPParams(ctx context.Context, sseReq *config.AgentSSERequest, assistant *model.Assistant) error {
	mcpInfos, err := s.cli.GetAssistantMCPList(ctx, assistant.ID)
	if err != nil {
		return fmt.Errorf("Assistant服务获取MCP信息失败，assistantId: %d, error: %v", assistant.ID, err)
	}
	mcpTools := make(map[string]config.MCPToolInfo)
	for _, mcp := range mcpInfos {
		if !mcp.Enable {
			continue
		}

		switch mcp.MCPType {
		case constant.MCPTypeMCP:
			mcpCustom, err := MCP.GetCustomMCP(ctx, &mcp_service.GetCustomMCPReq{
				McpId: mcp.MCPId,
			})
			if err != nil {
				log.Errorf("Assistant服务获取MCP Custom信息失败，assistantId: %d, error: %v", assistant.ID, err)
				continue
			}
			mcpTools[mcpCustom.Info.Name] = config.MCPToolInfo{
				URL:       mcpCustom.SseUrl,
				Transport: "sse",
			}
			sseReq.McpTools = mcpTools
			sseReq.ToolsName = append(sseReq.ToolsName, mcp.ActionName)
		case constant.MCPTypeMCPServer:
			mcpServer, err := MCP.GetMCPServer(ctx, &mcp_service.GetMCPServerReq{
				McpServerId: mcp.MCPId,
			})
			if err != nil {
				log.Errorf("Assistant服务获取MCP Server信息失败，assistantId: %d, error: %v", assistant.ID, err)
				continue
			}
			mcpTools[mcpServer.Name] = config.MCPToolInfo{
				URL:       mcpServer.SseUrl,
				Transport: "sse",
			}
			sseReq.McpTools = mcpTools
			sseReq.ToolsName = append(sseReq.ToolsName, mcp.ActionName)
		}
	}

	return nil
}

// Set history parameters
func (s *Service) setHistoryParams(ctx context.Context, sseReq *config.AgentSSERequest, req *assistant_service.AssistantConversionStreamReq) {
	fieldConditions := map[string]interface{}{
		"conversationId": req.ConversationId,
		"userId":         req.Identity.UserId,
		"orgId":          req.Identity.OrgId,
	}
	indexPattern := "conversation_detail_infos_*"

	documents, _, err := es.Assistant().SearchByFields(ctx, indexPattern, fieldConditions, 0, 1000)
	if err != nil {
		log.Warnf("Assistant服务查询历史聊天记录失败，conversationId: %s, userId: %s, error: %v", req.ConversationId, req.Identity.UserId, err)
		return
	}

	var historyList []config.AssistantConversionHistory
	for _, doc := range documents {
		var detail model.ConversationDetails
		if err := json.Unmarshal(doc, &detail); err != nil {
			log.Warnf("Assistant服务解析ES历史聊天记录失败: %v", err)
			continue
		}
		history := config.AssistantConversionHistory{
			Query:         detail.Prompt,
			UploadFileUrl: extractFileUrlsFromModel(detail.FileInfo),
			Response:      detail.Response,
		}
		historyList = append(historyList, history)
	}

	if len(historyList) > 0 {
		sseReq.History = historyList
		log.Debugf("Assistant服务添加历史聊天记录到请求参数，conversationId: %s, 历史记录数: %d", req.ConversationId, len(historyList))
	}
}

func buildRerank(req *assistant_service.AssistantConversionStreamReq, knowledgebaseConfig *RAGKnowledgeBaseConfig, assistant *model.Assistant) (map[string]interface{}, error) {
	var rerankEndpoint map[string]interface{}
	if knowledgebaseConfig.PriorityMatch != 1 {
		rerankConfig := &common.AppModelConfig{}
		if assistant.RerankConfig != "" {
			if err := json.Unmarshal([]byte(assistant.RerankConfig), rerankConfig); err != nil {
				log.Errorf("Assistant服务解析智能体rerank配置失败，assistantId: %s, error: %v, rerankConfigRaw: %s", req.AssistantId, err, assistant.RerankConfig)
				return nil, err
			}
			if rerankConfig.Model == "" || rerankConfig.ModelId == "" {
				log.Errorf("Assistant服务缺少rerank配置，assistantId: %s", req.AssistantId)
				return nil, fmt.Errorf("智能体缺少rerank配置")
			}
		}
		rerankEndpoint = mp.ToModelEndpoint(rerankConfig.ModelId, rerankConfig.Model)
	}
	return rerankEndpoint, nil
}

// Helper function for saving conversations using a separate context
func saveConversation(originalCtx context.Context, req *assistant_service.AssistantConversionStreamReq, response, searchList string) {
	// If the original context has been canceled, create a new independent context
	if originalCtx.Err() != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := saveConversationDetailToES(ctx, req, response, searchList); err != nil {
			log.Errorf("保存聊天记录到ES失败，assistantId: %s, conversationId: %s, error: %v",
				req.AssistantId, req.ConversationId, err)
		}
		return
	}

	// Continue to use the original context while it is not canceled
	if err := saveConversationDetailToES(originalCtx, req, response, searchList); err != nil {
		log.Errorf("保存聊天记录到ES失败，assistantId: %s, conversationId: %s, error: %v",
			req.AssistantId, req.ConversationId, err)
	}
}

// buildRetrieveMethod constructs the retrieval method
func buildRetrieveMethod(matchType string) string {
	switch matchType {
	case "vector":
		return "semantic_search"
	case "text":
		return "full_text_search"
	case "mix":
		return "hybrid_search"
	}
	return ""
}

// buildRerankMod constructs reranking mode
func buildRerankMod(priorityType int32) string {
	if priorityType == 1 {
		return "weighted_score"
	}
	return "rerank_model"
}

// buildTermWeight constructs keyword coefficients
func buildTermWeight(knowConfig *RAGKnowledgeBaseConfig) float32 {
	if knowConfig.TermWeightEnable {
		return knowConfig.TermWeight
	}
	return 0.0
}

// buildWeight constructs weight information
func buildWeight(knowConfig *RAGKnowledgeBaseConfig) *config.WeightParams {
	if knowConfig.PriorityMatch != 1 {
		return nil
	}
	return &config.WeightParams{
		VectorWeight: knowConfig.SemanticsPriority,
		TextWeight:   knowConfig.KeywordPriority,
	}
}

type AppKnowledgebaseConfig struct {
	Knowledgebases []AppKnowledgeBase     `json:"knowledgebases"`
	Config         AppKnowledgebaseParams `json:"config"`
}

type AppKnowledgeBase struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AppKnowledgebaseParams struct {
	MaxHistory int32   `json:"maxHistory"` // longest context
	Threshold  float32 `json:"threshold"`  // filter threshold
	TopK       int32   `json:"topK"`       // Number of knowledge items

	MatchType         string  `json:"matchType"`         //matchType: vector (vector search), text (text search), mix (mixed search: vector + text)
	PriorityMatch     int32   `json:"priorityMatch"`     // Weight matching. This is only set to 1 after selecting the weight setting in mixed search mode.
	SemanticsPriority float32 `json:"semanticsPriority"` // semantic weight
	KeywordPriority   float32 `json:"keywordPriority"`   // Keyword weight
}

// RAGKnowledgeBaseConfig knowledge base configuration structure
type RAGKnowledgeBaseConfig struct {
	KnowledgeBaseIds     []string                `json:"knowledgeBaseIds"`     // Knowledge base information
	MaxHistory           int32                   `json:"maxHistory"`           // longest context
	Threshold            float32                 `json:"threshold"`            // filter threshold
	TopK                 int32                   `json:"topK"`                 // topK
	MatchType            string                  `json:"matchType"`            // Search type: vector (vector search), text (text search), mix (mixed search)
	KeywordPriority      float32                 `json:"keywordPriority"`      // Keyword weight
	PriorityMatch        int32                   `json:"priorityMatch"`        // Weight matching, only valid in mixed search mode, 1 means enabled
	SemanticsPriority    float32                 `json:"semanticsPriority"`    // semantic weight
	TermWeight           float32                 `json:"termWeight"`           // Keyword coefficient, default is 1
	TermWeightEnable     bool                    `json:"termWeightEnable"`     // Keyword coefficient switch
	AppKnowledgeBaseList []*AppKnowledgeBaseInfo `json:"AppKnowledgeBaseList"` // Knowledge base metadata
	UseGraph             bool                    `json:"useGraph"`             // Knowledge graph switch
}

type AppKnowledgeBaseInfo struct {
	KnowledgeBaseId      string                `json:"knowledgeBaseId"`
	KnowledgeBaseName    string                `json:"knowledgeBaseName"`
	MetaDataFilterParams *MetaDataFilterParams `json:"metaDataFilterParams"`
}

type MetaDataFilterParams struct {
	FilterEnable     bool                `json:"filterEnable"`     // Metadata filter switch
	FilterLogicType  string              `json:"filterLogicType"`  // Metadata logical conditions: and/or
	MetaFilterParams []*MetaFilterParams `json:"metaFilterParams"` // Metadata filter parameter list
}

type MetaFilterParams struct {
	Condition string `json:"condition"`
	Key       string `json:"key"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}

func mergeMaps(map1, map2 map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range map1 {
		result[k] = v
	}
	for k, v := range map2 {
		result[k] = v // If the key is repeated, the value of map2 overwrites map1
	}
	return result
}

func (s *Service) buildWorkflowPluginListAlgParam(ctx context.Context, assistantId string) (pluginList []config.PluginListAlgRequest, err error) {
	workflows, status := s.cli.GetAssistantWorkflowsByAssistantID(ctx, pkgUtil.MustU32(assistantId))
	if status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	// workflow ids
	var workflowIDs []string
	for _, workflow := range workflows {
		if !workflow.Enable {
			continue
		}
		workflowIDs = append(workflowIDs, workflow.WorkflowId)
	}
	if len(workflowIDs) == 0 {
		return nil, nil
	}
	// workflow schemas
	url, _ := net_url.JoinPath(config.Cfg().Workflow.Endpoint, config.Cfg().Workflow.ListSchemaUri)
	reqBody, _ := json.Marshal(map[string]interface{}{
		"workflow_ids": workflowIDs,
	})
	result, err := http_client.Default().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       reqBody,
		Timeout:    time.Minute,
		MonitorKey: "workflow_schema",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var schemas []map[string]interface{}
	if err = json.Unmarshal(result, &schemas); err != nil {
		return nil, err
	}
	for _, schema := range schemas {
		schemaByte, err := json.Marshal(schema)
		if err != nil {
			return nil, err
		}
		//Verify schema
		if err := openapi3_util.ValidateSchema(ctx, schemaByte); err != nil {
			return nil, err
		}
		pluginList = append(pluginList, config.PluginListAlgRequest{APISchema: schema})
	}
	log.Infof("Assistant服务查询到workflow，assistantId: %s, workflowList: %v", assistantId, pluginList)
	return pluginList, nil
}

func (s *Service) buildToolPluginListAlgParam(ctx context.Context, sseReq *config.AgentSSERequest, assistantId string, identity *assistant_service.Identity) (pluginList []config.PluginListAlgRequest, err error) {
	// Convert assistantId
	assistantIdConv := pkgUtil.MustU32(assistantId)
	resp, status := s.cli.GetAssistantToolList(ctx, assistantIdConv)
	if status != nil {
		return pluginList, errStatus(errs.Code_AssistantConversationErr, status)
	}

	// Iterate through the list of tools, processing each valid tool
	for _, tool := range resp {
		if !tool.Enable {
			continue // Skip disabled tools
		}

		var rawSchema string            // original schema string
		var apiAuth *openapi3_util.Auth // API certification information

		// Get details and original schema based on tool type
		switch tool.ToolType {
		case constant.ToolTypeCustom:
			// Get custom tool details
			customTool, err := MCP.GetCustomToolInfo(ctx, &mcp_service.GetCustomToolInfoReq{
				CustomToolId: tool.ToolId,
			})
			if err != nil {
				log.Errorf("获取自定义工具信息失败，assistantId: %s, toolId: %s, err: %v", assistantId, tool.ToolId, err)
				continue
			}
			rawSchema = customTool.Schema

			// API certification for building custom tools
			if customTool.ApiAuth != nil {
				if apiAuth, err = util.ConvertApiAuthWebRequestProto(customTool.ApiAuth); err != nil {
					log.Errorf("转换自定义工具API失败，assistantId: %s, toolId: %s, err: %v", assistantId, tool.ToolId, err)
					continue
				}
			}
		case constant.ToolTypeBuiltIn:
			// If it is a Bocha search, special processing is performed and it is compatible with the old agent interface parameter transfer format.
			if tool.ToolId == "bochawebsearch" {
				// Get built-in tool details
				builtinTool, err := MCP.GetSquareTool(ctx, &mcp_service.GetSquareToolReq{
					ToolSquareId: tool.ToolId,
					Identity: &mcp_service.Identity{
						UserId: identity.UserId,
						OrgId:  identity.OrgId,
					},
				})
				if err != nil {
					log.Infof("获取内置工具信息失败，assistantId: %s, toolId: %s, err: %v", assistantId, tool.ToolId, err)
					continue
				}
				if builtinTool.BuiltInTools == nil || builtinTool.BuiltInTools.ApiAuth == nil {
					log.Errorf("获取bocha内置工具apiKey失败，assistantId: %s, toolId: %s", assistantId, tool.ToolId)
					continue
				}

				sseReq.SearchKey = builtinTool.BuiltInTools.ApiAuth.ApiKeyValue

				// Calculate SearchUrl: parse the schema to obtain the first server url and the unique path url
				doc, err := openapi3_util.LoadFromData(ctx, []byte(builtinTool.Schema))
				if err != nil {
					log.Errorf("解析内置工具Schema失败，assistantId: %s, toolId: %s, err: %v", assistantId, tool.ToolId, err)
					continue
				} else {
					if len(doc.Servers) > 0 {
						serverURL := doc.Servers[0].URL
						for path := range doc.Paths.Map() {
							sseReq.SearchUrl = serverURL + path
							break
						}
					}
				}
				if tool.ToolConfig != "" {
					var toolConfig map[string]interface{}
					if err := json.Unmarshal([]byte(tool.ToolConfig), &toolConfig); err != nil {
						log.Errorf("解析工具配置失败，assistantId: %s, toolId: %s, err: %v", assistantId, tool.ToolId, err)
						continue
					} else {
						if rerankId, ok := toolConfig["rerankId"]; ok {
							sseReq.SearchRerankId = rerankId
						}
					}
				} else {
					log.Errorf("bocha内置工具配置为空，缺少rerankId。assistantId: %s, toolId: %s", assistantId, tool.ToolId)
					continue
				}
				sseReq.UseSearch = true
				continue
			}
			// Get built-in tool details
			builtinTool, err := MCP.GetSquareTool(ctx, &mcp_service.GetSquareToolReq{
				ToolSquareId: tool.ToolId,
				Identity: &mcp_service.Identity{
					UserId: identity.UserId,
					OrgId:  identity.OrgId,
				},
			})
			if err != nil {
				log.Errorf("获取内置工具信息失败，assistantId: %s, toolId: %s, err: %v", assistantId, tool.ToolId, err)
				continue
			}
			rawSchema = builtinTool.Schema

			// Build API certification for built-in tools
			apiAuth, err = util.ConvertApiAuthWebRequestProto(builtinTool.BuiltInTools.ApiAuth)
			if err != nil {
				return nil, err
			}

		}

		// Process schema
		apiSchema, err := processSchema(ctx, rawSchema, tool.ActionName)
		if err != nil {
			return pluginList, err
		}

		pluginList = append(pluginList, config.PluginListAlgRequest{
			APISchema: apiSchema,
			APIAuth:   apiAuth,
		})
	}

	return pluginList, nil
}

func processSchema(ctx context.Context, rawSchema string, actionName string) (map[string]interface{}, error) {
	// Filter the specified operation_id in the schema
	filteredSchema, err := openapi3_util.FilterSchemaOperations(ctx, []byte(rawSchema), []string{actionName})
	if err != nil {
		return nil, err
	}

	// Verify schema format
	validatedSchema, err := openapi3_util.LoadFromData(ctx, filteredSchema)
	if err != nil {
		return nil, err
	}

	// Convert to map[string]interface{}
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

// SSEError Send SSE error response
func SSEError(stream assistant_service.AssistantService_AssistantConversionStreamServer, message string) {
	log.Errorf("SSE错误: %s", message)
	// Send error information via streaming response
	if stream != nil {
		errorResponse := fmt.Sprintf("error:%s", message)
		if err := stream.Send(&assistant_service.AssistantConversionStreamResp{
			Content: errorResponse,
		}); err != nil {
			log.Errorf("发送SSE错误响应失败: %v", err)
		} else {
			log.Infof("成功发送SSE错误响应: %s", message)
		}
	} else {
		log.Warnf("stream为nil，无法发送SSE错误响应: %s", message)
	}
}

func HttpRequestLlmStream(ctx context.Context, url, userId, xuid string, body io.Reader, timeout time.Duration) (*http.Response, error) {
	requestCtx, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		log.Errorf("HttpRequestLlmStream创建HTTP请求失败，url: %s, userId: %s, error: %v", url, userId, err)
		return nil, err
	}

	// Set request header
	requestCtx.Header.Set("Content-Type", "application/json")
	requestCtx.Header.Set("X-Uid", xuid)

	log.Debugf("HttpRequestLlmStream请求详情，url: %s, userId: %s, method: %s, headers: %+v",
		url, userId, requestCtx.Method, requestCtx.Header)

	// Create a client and send a request
	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	response, err := client.Do(requestCtx)
	if err != nil {
		log.Errorf("HttpRequestLlmStream发送HTTP请求失败，url: %s, userId: %s, error: %v", url, userId, err)
		return nil, err
	}

	log.Debugf("HttpRequestLlmStream收到HTTP响应，url: %s, userId: %s, statusCode: %d, responseHeaders: %+v",
		url, userId, response.StatusCode, response.Header)

	return response, err
}

// saveConversationDetailToES saves chat records to ES
func saveConversationDetailToES(ctx context.Context, req *assistant_service.AssistantConversionStreamReq, response, searchList string) error {
	// Generate an index name based on the current time in the format conversation_detail_infos_YYYYMM
	now := time.Now()
	indexName := fmt.Sprintf("conversation_detail_infos_%d%02d", now.Year(), now.Month())

	// Assembling ConversationDetails data
	nowMilli := now.UnixMilli()
	conversationDetail := &model.ConversationDetails{
		Id:             uuid.New().String(),
		AssistantId:    req.AssistantId,
		ConversationId: req.ConversationId,
		Prompt:         req.Prompt,
		FileInfo:       extractFileInfos(req.FileInfo),
		Response:       response,
		SearchList:     searchList,
		UserId:         req.Identity.UserId,
		OrgId:          req.Identity.OrgId,
		CreatedAt:      nowMilli,
		UpdatedAt:      nowMilli,
	}

	// Write to ES
	if err := es.Assistant().IndexDocument(ctx, indexName, conversationDetail); err != nil {
		return fmt.Errorf("写入ES失败: %v", err)
	}

	log.Infof("成功保存聊天记录到ES，索引: %s, assistantId: %s, conversationId: %s",
		indexName, req.AssistantId, req.ConversationId)
	return nil
}

// ConversationDeleteByAssistantId Delete a conversation based on agent ID
func (s *Service) ConversationDeleteByAssistantId(ctx context.Context, req *assistant_service.ConversationDeleteByAssistantIdReq) (*emptypb.Empty, error) {
	if status := s.cli.DeleteConversationByAssistantID(ctx, req.AssistantId, req.Identity.UserId, req.Identity.OrgId); status != nil {
		return nil, errStatus(errs.Code_AssistantConversationErr, status)
	}
	return &emptypb.Empty{}, nil
}

// extractCodeFromStreamData safely extracts code fields from streaming data
// After JSON parsing, the number type is float64 and needs to be safely converted to int.
func extractCodeFromStreamData(streamData map[string]interface{}) (int, bool) {
	codeVal, exists := streamData["code"]
	if !exists {
		return 0, false
	}

	switch v := codeVal.(type) {
	case float64:
		return int(v), true
	case int:
		return v, true
	default:
		return 0, false
	}
}

// extractFileInfos extracts all file information from proto FileInfo to model FileInfo
func extractFileInfos(fileInfos []*assistant_service.ConversionStreamFile) []model.FileInfo {
	if len(fileInfos) == 0 {
		return nil
	}
	var result []model.FileInfo
	for _, file := range fileInfos {
		if file != nil {
			result = append(result, model.FileInfo{
				FileName: file.FileName,
				FileSize: file.FileSize,
				FileUrl:  file.FileUrl,
			})
		}
	}
	return result
}

// extractFileUrls extracts all file URLs from proto FileInfo
func extractFileUrls(fileInfos []*assistant_service.ConversionStreamFile) []string {
	if len(fileInfos) == 0 {
		return nil
	}
	var fileUrls []string
	for _, file := range fileInfos {
		if file != nil && file.FileUrl != "" {
			fileUrls = append(fileUrls, file.FileUrl)
		}
	}
	return fileUrls
}

// extractFileUrlsFromModel extracts all file URLs from model FileInfo
func extractFileUrlsFromModel(fileInfos []model.FileInfo) []string {
	if len(fileInfos) == 0 {
		return nil
	}
	var fileUrls []string
	for _, file := range fileInfos {
		if file.FileUrl != "" {
			fileUrls = append(fileUrls, file.FileUrl)
		}
	}
	return fileUrls
}

// buildMetaDataFilterParams constructs metadata filter parameters
func buildMetaDataFilterParams(knowledgeInfos []*AppKnowledgeBaseInfo) ([]*config.MetadataFilterParam, error) {
	if len(knowledgeInfos) == 0 {
		return nil, nil
	}
	var ragMetaDataFilterParams []*config.MetadataFilterParam
	for _, k := range knowledgeInfos {
		if k.MetaDataFilterParams == nil || !k.MetaDataFilterParams.FilterEnable ||
			len(k.MetaDataFilterParams.MetaFilterParams) == 0 {
			continue
		}
		item, err := buildMetadataFilterItem(k.MetaDataFilterParams.MetaFilterParams)
		if err != nil {
			log.Errorf("buildMetaDataFilterParams error %v", err)
			return nil, err
		}
		ragMetaDataFilterParams = append(ragMetaDataFilterParams, &config.MetadataFilterParam{
			FilterKnowledgeName: k.KnowledgeBaseName,
			LogicalOperator:     k.MetaDataFilterParams.FilterLogicType,
			MetaList:            item,
		})
	}
	return ragMetaDataFilterParams, nil
}

func buildMetadataFilterItem(metaFilterParams []*MetaFilterParams) ([]*config.MetadataFilterItem, error) {
	var ragMetaDataFilterItem []*config.MetadataFilterItem
	for _, k := range metaFilterParams {
		data, err := buildValueData(k.Type, k.Value, k.Condition)
		if err != nil {
			log.Errorf("buildMetadataFilterItem error %v", err)
			return nil, err
		}
		ragMetaDataFilterItem = append(ragMetaDataFilterItem, &config.MetadataFilterItem{
			ComparisonOperator: k.Condition,
			MetaName:           k.Key,
			MetaType:           k.Type,
			Value:              data,
		})
	}
	return ragMetaDataFilterItem, nil
}

func buildValueData(valueType string, value string, condition string) (interface{}, error) {
	if condition == "empty" || condition == "not empty" {
		return nil, nil
	}
	switch valueType {
	case metaTypeNumber:
	case metaTypeTime:
		return strconv.ParseInt(value, 10, 64)
	}
	return value, nil
}

// transRequestFiles converts model.FileInfo to assistant_service.RequestFile, and replaces fileUrl with minio external download url
func transRequestFiles(files []model.FileInfo) []*assistant_service.RequestFile {
	if files == nil {
		return nil
	}

	downloadURL := os.Getenv("MINIO_DOWNLOAD_URL")
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")

	var result []*assistant_service.RequestFile
	for _, file := range files {
		// Replace fileUrl with minio external download url
		replacedUrl := strings.Replace(file.FileUrl, "http://"+minioEndpoint+"/", downloadURL, 1)

		result = append(result, &assistant_service.RequestFile{
			FileName: file.FileName,
			FileSize: file.FileSize,
			FileUrl:  replacedUrl,
		})
	}
	return result
}
