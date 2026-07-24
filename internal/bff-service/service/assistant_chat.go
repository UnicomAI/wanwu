package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/redis"
	safe_go_util "github.com/UnicomAI/wanwu/pkg/safe-go-util"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	sse_connector "github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector"
	sse_model "github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/model"
	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/session"
	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/store"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"google.golang.org/grpc"
)

const (
	agentEventFailStatus = 4 //事件失败
)

type agentChatStreamParams struct {
	ctx               *gin.Context
	startTime         time.Time
	firstTokenLatency int64
	hasRecorded       bool
	hasErr            bool
}

func AssistantConversionStream(ctx *gin.Context, userId, orgId, clientID string, req request.ConversionStreamRequest, needLatestPublished bool, source string) (err error) {
	// 1. CallAssistantConversationStream
	streamParams := &agentChatStreamParams{ctx: ctx, startTime: time.Now()}
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	defer func() {
		if source != constant.AppStatisticSourceDraft {
			go func() {
				defer util.PrintPanicStack()
				RecordAppStatistic(detachedCtx, userId, orgId, req.AssistantId, constant.AppTypeAgent, !streamParams.hasErr, true, streamParams.firstTokenLatency, 0, source)
			}()
		}
	}()

	chatCh, err := CallAssistantConversationStream(ctx, userId, orgId, clientID, req, needLatestPublished)
	if err != nil {
		streamParams.hasErr = true
		return err
	}
	// 2. 流式返回结果
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[Agent] %v conversation %v user %v org %v recv", req.AssistantId, req.ConversationId, userId, orgId), sse_util.DONE_MSG).
		WriteStream(chatCh, streamParams, buildAgentChatRespLineProcessor(), nil)
	return nil
}

func GetPendingConversation(ctx *gin.Context, userId, orgId, clientID string, req request.PendingConversionRequest) (*response.PendingConversationResp, error) {
	conversationID, err := getConversationID(ctx, userId, orgId, req)
	if err != nil {
		return nil, err
	}
	session := sse_connector.GetSession(&sse_model.Session{ConversationID: conversationID, ClientID: clientID})
	if session == nil {
		return &response.PendingConversationResp{
			ConversationId:         conversationID,
			HasPendingConversation: false,
		}, nil
	}
	ext := session.GetExt()
	var prompt string
	promptData := ext["prompt"]
	if promptData != nil {
		prompt1, ok := promptData.(string)
		if ok {
			prompt = prompt1
		}
	}
	fileInfoData := ext["fileInfo"]
	var requestFiles []response.AssistantRequestFile
	if fileInfoData != nil {
		files, ok := fileInfoData.([]request.ConversionStreamFile)
		if ok {
			if len(files) > 0 {
				for _, file := range files {
					requestFiles = append(requestFiles, response.AssistantRequestFile{
						FileName: file.FileName,
						FileSize: file.FileSize,
						FileUrl:  file.FileUrl,
					})
				}
			}
		}
	}

	return &response.PendingConversationResp{
		ConversationId:         conversationID,
		HasPendingConversation: true,
		Prompt:                 prompt,
		RequestFiles:           requestFiles,
	}, nil
}

func getConversationID(ctx *gin.Context, userId, orgId string, req request.PendingConversionRequest) (string, error) {
	var conversationID = req.ConversationId
	if req.Draft {
		// 获取 conversation_id
		conversationIdResp, err := assistant.GetConversationIdByAssistantId(ctx.Request.Context(), &assistant_service.GetConversationIdByAssistantIdReq{
			AssistantId:      req.AssistantId,
			ConversationType: constant.ConversationTypeDraft,
			Identity: &assistant_service.Identity{
				UserId: userId,
				OrgId:  orgId,
			},
		})

		if err != nil {
			// 草稿对话尚未创建：删除请求幂等成功，不向调用方抛 5xx。其它错误原样上抛。
			if isRecordNotFoundErr(err) {
				return "", nil
			}
			return "", err
		}
		if conversationIdResp == nil {
			return "", nil
		}
		conversationID = conversationIdResp.ConversationId
	}
	return conversationID, nil
}
func AssistantConversionStreamConnect(ctx *gin.Context, userId, orgId, clientID string, req request.ConversionStreamConnectRequest) error {
	// 1. CallAssistantConversationStream
	streamParams := &agentChatStreamParams{ctx: ctx, startTime: time.Now()}
	chatCh, err := sse_connector.Connect[string](ctx, &sse_model.Session{ConversationID: req.ConversationId, ClientID: clientID}, func(data *sse_model.Message) string {
		return data.Data.(string)
	})
	if err != nil {
		return err
	}
	// 2. 流式返回结果
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[Agent] %v conversation %v user %v org %v recv", req.AssistantId, req.ConversationId, userId, orgId), sse_util.DONE_MSG).
		WriteStream(chatCh, streamParams, buildAgentChatRespLineProcessor(), nil)
	return nil
}

func AssistantConversionStreamCancel(ctx *gin.Context, userId, orgId, clientID string, req request.ConversionStreamCancelRequest) error {
	conversationID, err := getConversationID(ctx, userId, orgId, req.PendingConversionRequest)
	if err != nil {
		return err
	}
	return sse_connector.Close(&sse_model.Session{ConversationID: conversationID, ClientID: clientID})
}

func CallAssistantConversationStream(ctx *gin.Context, userId, orgId, clientId string, req request.ConversionStreamRequest, needLatestPublished bool) (<-chan string, error) {
	// 根据agentID获取敏感词配置
	agentInfo, err := searchAssistantInfo(ctx, userId, orgId, req.AssistantId, needLatestPublished)
	if err != nil {
		return nil, err
	}
	//创建敏感词校验器（SafetyConfig 可能为 nil，如 openapi 创建的智能体未配置安全，用空安全 getter 避免解引用崩溃）
	sensitiveChecker := CreateSensitiveChecker(agentSafetyList(agentInfo.SafetyConfig), &agentSensitiveService{}, agentInfo.SafetyConfig.GetEnable())
	//任务执行器
	streamExecutor := conversationStream(ctx, userId, orgId, clientId, agentInfo, req, needLatestPublished)
	//带敏感词校验的任务执行
	return sensitiveChecker.Check(ctx, req.Prompt, streamExecutor)
}

// AssistantQuestionRecommend 智能体问题推荐
func AssistantQuestionRecommend(ctx *gin.Context, userId, orgId string, req *request.QuestionRecommendRequest) {
	//查询智能体服务
	agentInfo, err := searchAssistantInfo(ctx, userId, orgId, req.AssistantId, !req.Trial)
	if err != nil {
		log.Errorf("[Agent] %v conversation %v user %v org %v get assistant info err: %v", req.AssistantId, req.ConversationId, userId, orgId, err)
		gin_util.Response(ctx, nil, nil)
		return
	}
	// 检验参数
	if err = checkRecommendParam(agentInfo); err != nil {
		log.Errorf("[Agent] %v conversation %v user %v org %v check param err: %v", req.AssistantId, req.ConversationId, userId, orgId, err)
		gin_util.Response(ctx, nil, nil)
		return
	}
	data := mp_common.LLMReq{}
	// 构造参数
	if req.Trial {
		data = buildTrialRecommendParams(agentInfo, true, req.Query)
	} else {
		data, err = buildPublishRecommendParams(ctx, userId, orgId, true, req, agentInfo)
		if err != nil {
			log.Errorf("[Agent] %v conversation %v user %v org %v build publish recommend params err: %v", req.AssistantId, req.ConversationId, userId, orgId, err)
			gin_util.Response(ctx, nil, nil)
			return
		}
	}
	// 后续流式响应由 AgentRecommendChatCompletions 内部直接写入 ctx
	AgentRecommendChatCompletions(ctx, agentInfo.RecommendConfig.ModelConfig.ModelId, &data)
}

// GenAgentRecommendQuestions 在对话结束后生成追问问题列表（非流式收集），供 openapi 对话接口内嵌返回。
// 未开启追问/未配置推荐模型/模型拒绝推荐/解析失败时返回 nil，不影响对话本身。
func GenAgentRecommendQuestions(ctx *gin.Context, userId, orgId, assistantId, conversationId, query string, needLatestPublished bool) []string {
	agentInfo, err := searchAssistantInfo(ctx, userId, orgId, assistantId, needLatestPublished)
	if err != nil {
		log.Errorf("[Agent] %v recommend get assistant info err: %v", assistantId, err)
		return nil
	}
	if agentInfo.RecommendConfig == nil {
		return nil
	}
	// checkRecommendParam 会校验追问开关/推荐模型，并按需补默认 systemPrompt
	if err := checkRecommendParam(agentInfo); err != nil {
		return nil
	}
	req := &request.QuestionRecommendRequest{Query: query, AssistantId: assistantId, ConversationId: conversationId}
	data, err := buildPublishRecommendParams(ctx, userId, orgId, false, req, agentInfo)
	if err != nil {
		log.Errorf("[Agent] %v recommend build params err: %v", assistantId, err)
		return nil
	}
	answer, err := collectRecommendLLMAnswer(ctx, agentInfo.RecommendConfig.ModelConfig.ModelId, &data)
	if err != nil {
		log.Errorf("[Agent] %v recommend llm err: %v", assistantId, err)
		return nil
	}
	return parseRecommendQuestions(answer)
}

func buildPublishRecommendParams(ctx *gin.Context, userId string, orgId string, streamValue bool, req *request.QuestionRecommendRequest, agentInfo *assistant_service.AssistantInfo) (mp_common.LLMReq, error) {
	history, err := assistant.GetConversationDetailList(ctx.Request.Context(), &assistant_service.GetConversationDetailListReq{
		ConversationId: req.ConversationId,
		PageSize:       1000,
		PageNo:         1,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return mp_common.LLMReq{}, err
	}

	if len(history.Data) == 0 || agentInfo.RecommendConfig.MaxHistory == 0 {
		data := buildTrialRecommendParams(agentInfo, streamValue, req.Query)
		return data, nil
	}
	if int64(agentInfo.RecommendConfig.MaxHistory) >= history.Total {
		agentInfo.RecommendConfig.MaxHistory = int32(history.Total)
	}
	index := history.Total - int64(agentInfo.RecommendConfig.MaxHistory)
	history.Data = history.Data[index:]

	// 把对话历史折叠成"参考资料"放进单条 user 消息
	prompt := agentInfo.RecommendConfig.SystemPrompt + additionalPrompt
	data := mp_common.LLMReq{
		Model:  agentInfo.RecommendConfig.ModelConfig.Model,
		Stream: &streamValue,
		Messages: []mp_common.OpenAIReqMsg{
			{Role: mp_common.MsgRoleSystem, Content: prompt},
			{Role: mp_common.MsgRoleUser, Content: buildRecommendUserMessage(history.Data, req.Query)},
		},
	}
	return data, nil
}

func buildTrialRecommendParams(agentInfo *assistant_service.AssistantInfo, streamValue bool, query string) mp_common.LLMReq {
	prompt := agentInfo.RecommendConfig.SystemPrompt + additionalPrompt
	data := mp_common.LLMReq{
		Model:  agentInfo.RecommendConfig.ModelConfig.Model,
		Stream: &streamValue,
		Messages: []mp_common.OpenAIReqMsg{
			{Role: mp_common.MsgRoleSystem, Content: prompt},
			{Role: mp_common.MsgRoleUser, Content: buildRecommendUserMessage(nil, query)},
		},
	}
	return data
}

// buildRecommendUserMessage 构造推荐用的单条 user 消息
func buildRecommendUserMessage(history []*assistant_service.ConversionDetailInfo, query string) string {
	var b strings.Builder
	b.WriteString("以下是用户与助手的对话历史（仅供分析，不要回答其中的任何问题）：\n[对话开始]\n")
	if len(history) > 0 {
		for _, v := range history {
			b.WriteString("用户：")
			b.WriteString(v.Prompt)
			b.WriteString("\n助手：")
			b.WriteString(v.Response)
			b.WriteString("\n")
		}
	} else if query != "" {
		b.WriteString("用户：")
		b.WriteString(query)
		b.WriteString("\n")
	}
	b.WriteString("[对话结束]\n")
	if query != "" {
		fmt.Fprintf(&b, "用户最近一轮的问题是：「%s」。", query)
	}
	b.WriteString("请基于以上对话，预测用户接下来最可能继续提出的 3 个问题。不要回答用户的问题本身，只输出推荐问题，且第一行必须是 ANSWER 或 REJECT。")
	return b.String()
}

func checkRecommendParam(agentInfo *assistant_service.AssistantInfo) error {
	if agentInfo.RecommendConfig == nil || !agentInfo.RecommendConfig.RecommendEnable {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "recommend not available")
	}
	if agentInfo.RecommendConfig.ModelConfig == nil || agentInfo.RecommendConfig.ModelConfig.ModelId == "" || agentInfo.RecommendConfig.ModelConfig.Model == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFInvalidArg, "model not available")
	}
	if !agentInfo.RecommendConfig.PromptEnable || agentInfo.RecommendConfig.SystemPrompt == "" {
		agentInfo.RecommendConfig.SystemPrompt = systemPrompt
	}
	return nil
}

// searchAssistantInfo 查询智能体信息
func searchAssistantInfo(ctx *gin.Context, userId, orgId, assistantId string, publish bool) (*assistant_service.AssistantInfo, error) {
	var agentInfo *assistant_service.AssistantInfo
	var err error
	if publish {
		agentInfo, err = assistant.AssistantSnapshotInfo(ctx.Request.Context(), &assistant_service.AssistantSnapshotInfoReq{
			AssistantId: assistantId,
		})
	} else {
		agentInfo, err = assistant.GetAssistantInfo(ctx.Request.Context(), &assistant_service.GetAssistantInfoReq{
			AssistantId: assistantId,
			Identity: &assistant_service.Identity{ //草稿只能看自己的
				UserId: userId,
				OrgId:  orgId,
			},
		})
	}
	if err != nil {
		return nil, err
	}
	return agentInfo, nil
}

// transFileInfo 转换文件信息从请求模型到protobuf模型
func transFileInfo(fileInfo []request.ConversionStreamFile) []*assistant_service.ConversionStreamFile {
	if len(fileInfo) == 0 {
		return nil
	}
	result := make([]*assistant_service.ConversionStreamFile, 0, len(fileInfo))
	for _, file := range fileInfo {
		result = append(result, &assistant_service.ConversionStreamFile{
			FileName: file.FileName,
			FileSize: file.FileSize,
			FileUrl:  file.FileUrl,
		})
	}
	return result
}

// buildAgentChatRespLineProcessor 构造agent对话结果行处理器
func buildAgentChatRespLineProcessor() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		if p, ok := params.(*agentChatStreamParams); ok {
			if !p.hasRecorded {
				p.firstTokenLatency = time.Since(p.startTime).Milliseconds()
				p.hasRecorded = true
				if p.ctx != nil {
					p.ctx.Set(gin_util.FIRST_RESP_LATENCY, p.firstTokenLatency)
				}
			}
		}
		if strings.HasPrefix(lineText, "error:") {
			if p, ok := params.(*agentChatStreamParams); ok {
				p.hasErr = true
			}
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		return "data:" + lineText + "\n\n", false, nil
	}
}

// --- agent sensitive ---

type agentSensitiveService struct {
	currentOrder     int
	currentEventType int
	currentEventData *agentEventData
	currentDetailId  string
}

type agentEventData struct {
	Status    int    `json:"status"`
	Id        string `json:"id"`
	EventType int    `json:"eventType"`
	Name      string `json:"name"`
	Profile   string `json:"profile"`
	TimeCost  string `json:"timeCost"`
	ParentId  string `json:"parentId"`
	Order     int    `json:"order"`
}

func (s *agentSensitiveService) serviceType() string {
	return constant.AppTypeAgent
}

// parseContent implements ChatService.
func (s *agentSensitiveService) parseContent(raw string) (id, content string) {
	raw = strings.TrimPrefix(raw, "data:")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	resp := struct {
		MsgID     string          `json:"msg_id"`
		DetailId  string          `json:"detailId"`
		Response  string          `json:"response"`
		EventType int             `json:"eventType"`
		Order     int             `json:"order"`
		EventData *agentEventData `json:"eventData"`
	}{}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", ""
	}
	s.currentOrder = resp.Order
	s.currentEventType = resp.EventType
	s.currentEventData = resp.EventData
	s.currentDetailId = resp.DetailId
	return resp.MsgID, resp.Response
}

// buildSensitiveResp implements ChatService.
func (s *agentSensitiveService) buildSensitiveResp(id string, content string) []string {
	data := s.currentEventData
	if data != nil {
		data.Status = agentEventFailStatus
	}
	resp := map[string]interface{}{
		"code":              0,
		"message":           "success",
		"response":          content,
		"detailId":          s.currentDetailId,
		"eventType":         s.currentEventType,
		"order":             s.currentOrder,
		"eventData":         data,
		"gen_file_url_list": []interface{}{},
		"history":           []interface{}{},
		"finish":            1,
		"usage": map[string]interface{}{
			"prompt_tokens":     0,
			"completion_tokens": 0,
			"total_tokens":      0,
		},
		"search_list": []interface{}{},
	}
	marshal, _ := json.Marshal(resp)
	return []string{"data: " + string(marshal)}
}

func buildMultiAssistantConversionStreamReq(req *assistant_service.AssistantConversionStreamReq) *assistant_service.MultiAssistantConversionStreamReq {
	return &assistant_service.MultiAssistantConversionStreamReq{
		AssistantId:    req.AssistantId,
		ConversationId: req.ConversationId,
		FileInfo:       req.FileInfo,
		Prompt:         req.Prompt,
		SystemPrompt:   req.SystemPrompt,
		Identity:       req.Identity,
		Draft:          req.Draft,
		DetailId:       req.DetailId,
	}
}

// sseCompactProcessor 构造sse消息合并处理器
func sseCompactProcessor() func(currentMsg *sse_model.Message, lastMsg *sse_model.Message) (bool, *sse_model.Message) {
	return func(currentMsg *sse_model.Message, lastMsg *sse_model.Message) (bool, *sse_model.Message) {
		// 判断是否需要合并
		noneProcess, lastMsgData, currentMsgData := noneCompactMessage(currentMsg, lastMsg)
		if noneProcess {
			return true, currentMsg
		}
		//开始合并
		compact := lastMsgData.Compact(currentMsgData)
		if compact != nil { //合并成功
			resp, err := response.MarshalAgentResp(compact)
			if err != nil {
				log.Errorf("marshal agent resp error %v", err)
				return true, currentMsg
			}
			lastMsg.Data = resp
			return false, lastMsg
		}
		return true, currentMsg
	}
}

// noneCompactMessage 判断是否需要合并
func noneCompactMessage(currentMsg *sse_model.Message, lastMsg *sse_model.Message) (bool, *response.AgentChatResp, *response.AgentChatResp) {
	lastMsgData, err1 := response.UnmarshalAgentResp(lastMsg.Data.(string))
	if err1 != nil {
		log.Errorf("unmarshal agent resp %s error %v", lastMsg.Data.(string), err1)
		return true, nil, nil
	}
	currentMsgData, err2 := response.UnmarshalAgentResp(currentMsg.Data.(string))
	if err2 != nil {
		log.Errorf("unmarshal agent resp %s error %v", currentMsg.Data.(string), err2)
		return true, nil, nil
	}
	// 非成功状态码直接返回
	if currentMsgData.Code != 0 {
		return true, nil, nil
	}
	if currentMsgData.Finish != 0 {
		return true, nil, nil
	}
	return false, lastMsgData, currentMsgData
}

// 智能体流式会话
func conversationStream(ctx *gin.Context, userId, orgId, clientId string, agentInfo *assistant_service.AssistantInfo, req request.ConversionStreamRequest, needLatestPublished bool) func() (ch <-chan string, callback func(string, string), err error) {
	return func() (ch <-chan string, callback func(string, string), err error) {
		// 构建参数
		agentReq, sessionManager := buildAssistantChatParams(ctx, userId, orgId, clientId, req, needLatestPublished)
		bgCtx := sessionManager.GetBgContext()
		//执行调用
		var stream grpc.ServerStreamingClient[assistant_service.AssistantConversionStreamResp]
		if agentInfo.Category == constant.AgentCategoryMulti {
			stream, err = assistant.MultiAssistantConversionStream(bgCtx, buildMultiAssistantConversionStreamReq(agentReq))
		} else {
			stream, err = assistant.AssistantConversionStream(bgCtx, agentReq)
		}
		if err != nil {
			return nil, nil, err
		}
		//流式数据接收
		rawCh := safe_go_util.SafeChannelReceiveByIterCloser(bgCtx, assistantIteratorReader(agentReq, sessionManager, stream), sessionCloser(agentReq, sessionManager))
		callback = func(messageId string, sensitiveMsg string) {
			//敏感词存入redis
			redis.StoreSensitiveConversation(agentReq.ConversationId, agentReq.DetailId, sensitiveMsg)
			//触发sse cancel
			_ = sessionManager.Cancel()
		}
		return rawCh, callback, nil
	}
}

// assistantIteratorReader enio 返回的智能体数据处理器
func assistantIteratorReader(req *assistant_service.AssistantConversionStreamReq, sseSessionManager *session.Manager, stream grpc.ServerStreamingClient[assistant_service.AssistantConversionStreamResp]) *safe_go_util.IteratorReader[*assistant_service.AssistantConversionStreamResp, string] {
	//event读取器
	var reader = func(ctx context.Context) safe_go_util.IteratorReaderResponse[*assistant_service.AssistantConversionStreamResp, string] {
		event, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				log.Errorf("[Agent] %v conversation %v user %v org %v recv err: %v", req.AssistantId, req.ConversationId, req.Identity.UserId, req.Identity.OrgId, err)
			}
			return safe_go_util.IteratorResponseStop[*assistant_service.AssistantConversionStreamResp, string]()
		}
		return safe_go_util.IteratorReaderResponse[*assistant_service.AssistantConversionStreamResp, string]{Data: event}
	}
	// event数据处理器
	var processor = func(ctx context.Context, data *assistant_service.AssistantConversionStreamResp, rawCh chan string) ([]string, *safe_go_util.IteratorError[string]) {
		_ = sseSessionManager.Publish(&sse_model.Message{Data: data.Content}, sseCompactProcessor())
		select {
		case rawCh <- data.Content:
		default:
			//log.Debugf("[Agent] %v conversation %v user %v org %v recv chan full", req.AssistantId, req.ConversationId, req.Identity.UserId, req.Identity.OrgId)
		}
		return nil, nil
	}
	return &safe_go_util.IteratorReader[*assistant_service.AssistantConversionStreamResp, string]{
		Reader:    reader,
		Processor: processor,
	}
}

func sessionCloser(req *assistant_service.AssistantConversionStreamReq, sessionManager *session.Manager) func(ctx context.Context) {
	return func(ctx context.Context) {
		log.Infof("[Agent] %v conversation %v user %v org %v session finish", req.AssistantId, req.ConversationId, req.Identity.UserId, req.Identity.OrgId)
		if err1 := sse_connector.Close(sessionManager.GetSession()); err1 != nil {
			log.Errorf("[Agent] %v conversation %v user %v org %v session finish err: %v", req.AssistantId, req.ConversationId, req.Identity.UserId, req.Identity.OrgId, err1)
		}
	}
}

// buildAssistantChatParams 构造智能体会话参数
func buildAssistantChatParams(ctx *gin.Context, userId, orgId, clientId string, req request.ConversionStreamRequest, needLatestPublished bool) (*assistant_service.AssistantConversionStreamReq, *session.Manager) {
	agentReq := &assistant_service.AssistantConversionStreamReq{
		DetailId:       uuid.New().String(),
		AssistantId:    req.AssistantId,
		ConversationId: req.ConversationId,
		FileInfo:       transFileInfo(req.FileInfo),
		Prompt:         req.Prompt,
		SystemPrompt:   req.SystemPrompt,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
		Draft: !needLatestPublished,
	}

	//初始化sse 链接保持器
	sseSessionManager := sse_connector.NewSSESessionValid(ctx, &sse_model.Session{ConversationID: req.ConversationId, ClientID: clientId}, store.NewMemoryStore(), req.SseHold)
	// 添加扩展信息
	sseSessionManager.AddExt(map[string]interface{}{"prompt": req.Prompt, "fileInfo": req.FileInfo})
	return agentReq, sseSessionManager
}

func agentSafetyList(safetyConfig *assistant_service.AssistantSafetyConfig) []string {
	var ids []string
	for _, idx := range safetyConfig.GetSensitiveTable() {
		ids = append(ids, idx.TableId)
	}
	return ids
}
