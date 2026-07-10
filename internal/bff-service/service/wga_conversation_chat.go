package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/api/proto/common"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	sse_connector "github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector"
	sse_model "github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/model"
	"github.com/UnicomAI/wanwu/pkg/sse-util/sse-connector/store"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/wga"
	wga_persistent "github.com/UnicomAI/wanwu/pkg/wga-persistent"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================================================
// WgaChatParams - 通用 WGA 对话参数
// ============================================================================

// WgaChatParams 通用 WGA 对话参数
// 支持不同业务场景通过传入不同的 WorkspaceStore 实现 workspace 与 threadID 的解耦
type WgaChatParams struct {
	// 必填参数
	UserID   string
	OrgID    string
	AgentID  string
	ThreadID string
	Messages []request.GeneralAgentConversationMessage

	// ClientID - 客户端标识，用于 SSE 断线重连的 session 隔离
	// 非空时启用断线重连能力，为空时行为与改造前完全一致（断连即停止后端执行）
	ClientID string

	// ModelConfig - 模型配置（可选）
	// 由调用方从自己的配置系统获取，或传 nil 使用默认配置
	// 通用智能体：从 wga_conversation_config 表获取
	// 新业务场景：从自己的配置系统获取，或传 nil
	ModelConfig *common.AppModelConfig

	// WorkspaceStore - 由调用方创建
	// - 通用智能体：NewGeneralAgentWorkspaceStore(threadID)，store 与 threadID 绑定
	// - 新业务跨 thread：自定义 Store（如共享 workspace），与请求 threadID 解耦
	// - 无 workspace：传 nil
	WorkspaceStore *wga_persistent.Store

	// WorkspaceReadOnly - 工作空间只读模式（控制文件写入）
	// - true: 只设置 InputDir，agent 执行产出不写回 workspace
	// - false: 同时设置 InputDir 和 OutputDir，agent 执行产出写回 workspace
	// - 仅当 WorkspaceStore 不为 nil 时生效，用户上传文件不受此限制
	WorkspaceReadOnly bool

	// SendWorkspaceEvent - 是否发送 AG-UI activity workspace event（控制事件通知，与 WorkspaceReadOnly 解耦）
	// - true: 在 RUN_FINISHED 时发送 workspace activity event 通知前端更新
	// - false: 不发送 workspace activity event（默认）
	// - 仅当 WorkspaceStore 不为 nil 时生效
	SendWorkspaceEvent bool

	// EnableHumanInTheLoop - 是否启用人机交互（HITL）
	// - true: 启用（默认）
	// - false: 禁用（如 OpenAPI 场景）
	EnableHumanInTheLoop bool
}

const (
	wgaConversationHistoryEventESIndexName = "wga_chat_history_event" // 通用智能体聊天历史ES索引
)

func wgaConversationHistoryEventESIndexNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "index_not_found_exception") && strings.Contains(err.Error(), wgaConversationHistoryEventESIndexName)
}

// ============================================================================
// WgaConversationChat - 通用 WGA 对话执行
// ============================================================================

// WgaConversationChat 通用的 WGA 对话执行（完整流程，包含 SSE 响应）
// 支持不同业务场景通过传入不同的 WorkspaceStore 实现 workspace 与 threadID 的解耦
func WgaConversationChat(ctx *gin.Context, params *WgaChatParams) error {
	// 参数检查
	if params.UserID == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "userId is required")
	}
	if params.OrgID == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "orgId is required")
	}
	if params.AgentID == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "agentId is required")
	}
	if params.ThreadID == "" {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "threadId is required")
	}
	if len(params.Messages) == 0 {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "messages is required")
	}

	// 过滤用户消息（最后一条必须为 user 消息）
	var userInputMessage *request.GeneralAgentConversationMessage
	for i := len(params.Messages) - 1; i >= 0; i-- {
		if params.Messages[i].Role == ag_ui_util.RoleUser {
			userInputMessage = &params.Messages[i]
			break
		}
	}
	if userInputMessage == nil {
		return grpc_util.ErrorStatus(err_code.Code_BFFGeneral, "no user message found in input messages")
	}

	// 获取上一次输出的目录信息（如果存在）
	var lastWorkspaceTotalSize int64
	var lastWorkspaceFileCount int
	if params.WorkspaceStore != nil {
		lastRunDir, err := GetWgaWorkspaceLastRunDir(params.WorkspaceStore)
		if err != nil {
			log.Errorf("[wga] thread %v failed to get last run dir: %v", params.ThreadID, err)
		}
		if lastRunDir != "" {
			wsInfo, err := GetWgaWorkspaceInfo(lastRunDir)
			if err != nil {
				log.Errorf("[wga] thread %v failed to get last workspace info: %v", params.ThreadID, err)
			} else {
				lastWorkspaceTotalSize = wsInfo.TotalSize
				lastWorkspaceFileCount = wsInfo.FileCount
			}
		}
	}

	runID := uuid.NewString()

	// 构建 WGA 选项
	opts, err := buildWgaRunOptions(ctx, params.UserID, params.OrgID, params.AgentID, params.ThreadID, runID, userInputMessage, params.ModelConfig, params.WorkspaceStore, params.WorkspaceReadOnly, params.EnableHumanInTheLoop)
	if err != nil {
		return err
	}

	// 初始化 SSE 链接保持器
	// ClientID 非空时启用断线重连能力，为空时 InvalidManager 使 Publish/Subscribe 成为 no-op，行为与改造前一致
	sseSession := &sse_model.Session{
		ConversationID: params.ThreadID,
		ClientID:       params.ClientID,
	}
	sseSessionManager := sse_connector.NewSSESessionValid(ctx.Request.Context(), sseSession, store.NewMemoryStore(), params.ClientID != "")
	sseSessionManager.AddExt(map[string]interface{}{
		"runId":    runID,
		"messages": params.Messages,
	})
	bgCtx := sseSessionManager.GetBgContext()

	// 运行 WGA（使用 bgCtx，不随客户端断连中断）
	_, iter, err := wga.Run(bgCtx, params.AgentID, opts...)
	if err != nil {
		_ = sse_connector.Close(sseSession)
		return err
	}

	// 转换为 AG-UI 事件流
	tr := ag_ui_util.NewEinoTranslator(params.ThreadID, runID)
	eventCh := tr.TranslateStream(bgCtx, iter)

	// AG-UI 事件流处理器
	processorConfig := &ag_ui_util.ProcessorConfig{
		ToolNameMapper: map[string]string{
			"transfer_to_agent": "正在交给专业智能体",
		},
		ExcludedAgentNames: []string{
			config.WgaCfg().AgentID,     // 排除主智能体，避免生成冗余的智能体折叠效果
			ag_ui_util.DefaultAgentName, // 排除default，避免生成冗余的切换智能体提示
		},
		// ResultFormatters: map[string]func(string) string{
		// 	"bocha_comprehensive_search": WgaFormatBochaWebSearchResult,
		// 	"bocha_web_search_only":      WgaFormatBochaWebSearchResult,
		// 	"bocha_search_structured":    WgaFormatBochaWebSearchResult,
		// 	"bocha_search_day":           WgaFormatBochaWebSearchResult,
		// 	"bocha_search_last_week":     WgaFormatBochaWebSearchResult,
		// 	"tavily_basic_search":        WgaFormatTavilySearchResult,
		// 	"tavily_deep_search":         WgaFormatTavilySearchResult,
		// 	"tavily_day_search":          WgaFormatTavilySearchResult,
		// 	"tavily_week_search":         WgaFormatTavilySearchResult,
		// 	"tavily_image_search":        WgaFormatTavilySearchResult,
		// 	"tavily_date_search":         WgaFormatTavilySearchResult,
		// },
	}

	processor := ag_ui_util.NewStreamProcessor(processorConfig)
	processedEventCh, historyEventCh := processor.Process(bgCtx, eventCh, map[string]interface{}{
		"threadId":       params.ThreadID,
		"runId":          runID,
		"messages":       []interface{}{userInputMessage},
		"state":          map[string]interface{}{},
		"tools":          []interface{}{},
		"context":        []interface{}{},
		"forwardedProps": map[string]interface{}{},
	})

	// 确定 workspace event 注入策略（与 WorkspaceReadOnly 解耦）
	var eventWorkspaceStore *wga_persistent.Store
	if params.WorkspaceStore != nil && params.SendWorkspaceEvent {
		eventWorkspaceStore = params.WorkspaceStore
	}

	// 刷新会话 updated_at，使会话列表按最近聊天排序
	if _, err := assistant.UpdateWgaConversationConfig(ctx.Request.Context(), &assistant_service.UpdateWgaConversationConfigReq{
		ThreadId: params.ThreadID,
		Identity: &assistant_service.Identity{UserId: params.UserID, OrgId: params.OrgID},
	}); err != nil {
		log.Warnf("[WgaConversationChat] touch conversation updated_at failed, threadId: %s, err: %v", params.ThreadID, err)
	}
	// 异步保存智能体返回的消息（detached context，即使 ClientID 为空，客户端断连也不影响 ES 写入）
	go saveWgaChatHistoryEvent(trace_util.DetachContext(bgCtx), historyEventCh, params.UserID, params.OrgID, params.ThreadID, runID,
		eventWorkspaceStore,
		lastWorkspaceTotalSize,
		lastWorkspaceFileCount,
	)

	// 事件分发：同时写入 SSE 客户端输出和 sse-connector store
	// fan-out goroutine 从 processedEventCh 读取，publish 到 store（不阻塞），非阻塞写入 SSE 输出 channel
	// workspace event 在 RunFinished 时注入（先于 RunFinished 本身），确保 store 和 SSE 输出中顺序一致
	outCh := make(chan string, 1024)
	go func() {
		defer util.PrintPanicStack()
		defer close(outCh)
		defer func() { _ = sse_connector.Close(sseSession) }()
		for evt := range processedEventCh {
			// inject workspace activity event when run finished, to trigger AG-UI to fetch workspace info and update workspace activity card
			if eventWorkspaceStore != nil && evt.Type() == aguievents.EventTypeRunFinished {
				if wsEvent, _ := BuildWgaWorkspaceEvent(eventWorkspaceStore, &WgaWorkspaceEventConfig{
					ThreadID:           params.ThreadID,
					RunID:              runID,
					LastWorkspaceSize:  lastWorkspaceTotalSize,
					LastWorkspaceCount: lastWorkspaceFileCount,
				}); wsEvent != nil {
					if wsData, wsErr := json.Marshal(wsEvent); wsErr == nil {
						_ = sseSessionManager.Publish(&sse_model.Message{Data: string(wsData)}, nil)
						select {
						case outCh <- string(wsData):
						default:
							// SSE 输出 channel 已满（客户端断连或处理慢），丢弃事件（store 已保存，可重连恢复）
						}
					}
				}
			}
			data, err := json.Marshal(evt)
			if err != nil {
				log.Warnf("[WgaConversationChat] threadId=%s, runId=%s, marshal sse event error: %v", params.ThreadID, runID, err)
			} else {
				_ = sseSessionManager.Publish(&sse_model.Message{Data: string(data)}, nil)
				select {
				case outCh <- string(data):
				default:
					// SSE 输出 channel 已满（客户端断连或处理慢），丢弃事件（store 已保存，可重连恢复）
				}
			}
		}
	}()

	// SSE 响应输出（使用 ctx.Request.Context，客户端断连时停止写入）
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	ctx.Stream(func(w io.Writer) bool {
		select {
		case line, ok := <-outCh:
			if !ok {
				log.Infof("[WgaConversationChat] threadId=%s, runId=%s, outputCh closed", params.ThreadID, runID)
				return false
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", line)
			return true
		case <-ctx.Request.Context().Done():
			return false
		}
	})
	return nil
}

func filterWgaHistoryMessages(ctx *gin.Context, userId, orgId, threadId string) ([]*schema.Message, error) {
	resp, err := assistant.SearchFromES(ctx.Request.Context(), &assistant_service.SearchFromESReq{
		IndexName: wgaConversationHistoryEventESIndexName,
		Conditions: map[string]string{
			"threadId": threadId,
			"userId":   userId,
			"orgId":    orgId,
		},
		SortOrder: "asc",
		PageNo:    1,
		PageSize:  1000,
	})
	if err != nil {
		if wgaConversationHistoryEventESIndexNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	var messages []adk.Message
	for _, docJson := range resp.DocJsonList {
		var doc map[string]interface{}
		if err := json.Unmarshal([]byte(docJson), &doc); err != nil {
			continue
		}

		if eventsStr, ok := doc["events"].(string); ok {
			var events []map[string]interface{}
			if err := json.Unmarshal([]byte(eventsStr), &events); err != nil {
				log.Errorf("[wga] thread %v unmarshal events err: %v", threadId, err)
				continue
			}
			for _, event := range events {
				eventType, ok := event["type"].(string)
				if !ok {
					continue
				}
				var message *schema.Message
				switch aguievents.EventType(eventType) {
				case aguievents.EventTypeRunStarted:
					if input := event["input"].(map[string]interface{}); input != nil {
						if msgs, ok := input["messages"].([]interface{}); ok {
							for _, msg := range msgs { // messages中只有一个用户消息
								if m, ok := msg.(map[string]interface{}); ok {
									message = convertWgaMessage(m["role"].(string), m["content"])
								}
							}
						}
					}

				case aguievents.EventTypeReasoningMessageContent:
					if delta, ok := event["delta"].(string); ok {
						message = &schema.Message{
							Role:    ag_ui_util.RoleAssistant,
							Content: "[思考过程] " + delta,
						}
					}
				case aguievents.EventTypeTextMessageContent:
					if delta, ok := event["delta"].(string); ok {
						message = &schema.Message{
							Role:    ag_ui_util.RoleAssistant,
							Content: delta,
						}
					}
				default:
					// do nothing
				}
				if message != nil {
					messages = append(messages, message)
				}
			}
		}
	}
	return messages, nil
}

func buildWgaRunOptions(ctx *gin.Context, userID, orgID, agentID, threadID, runID string, userInputMessage *request.GeneralAgentConversationMessage, modelConfig *common.AppModelConfig, workspaceStore *wga_persistent.Store, workspaceReadOnly, enableHumanInTheLoop bool) ([]wga_option.Option, error) {
	// 获取 WGA 配置
	wgaConfigResp, err := assistant.GetWgaConfig(ctx.Request.Context(), &assistant_service.GetWgaConfigReq{
		Identity: &assistant_service.Identity{
			UserId: userID,
			OrgId:  orgID,
		},
	})
	if err != nil {
		return nil, err
	}
	wgaConfig := wgaConfigResp.Config
	if wgaConfig == nil {
		wgaConfig = &assistant_service.WgaConfig{}
	}

	// 解析用户消息中的 @提及资源
	mentionResources := &wgaMentionResources{}
	if userInputMessage != nil {
		mentionNames := parseWgaResourceMentions(userInputMessage.GetTextContent())
		mentionResources = fetchWgaMentionResources(ctx, userID, orgID, mentionNames)
	}

	opts := []wga_option.Option{
		wga_option.WithRunSession(wga_option.RunSession{
			ThreadID: threadID,
			RunID:    runID,
		}),
		wga_option.WithEnableHumanInTheLoop(config.WgaCfg().HumanInTheLoop && enableHumanInTheLoop, true),
		wga_option.WithSystemMessageStrategy(wga_option.SystemMessageStrategyMerge),
	}
	// 校验并构建模型配置选项（由调用方提供 ModelConfig）
	if modelConfig != nil && modelConfig.ModelId != "" {
		if err := checkModelConfigFromProto(ctx, modelConfig); err != nil {
			return nil, err
		}
		modelOpt, err := buildWgaModelOption(ctx, modelConfig)
		if err != nil {
			return nil, err
		}
		opts = append(opts, modelOpt)
	}

	// 校验并构建工具配置选项
	if len(wgaConfig.ToolList) > 0 {
		toolOpts, err := buildWgaToolOptions(ctx, userID, orgID, wgaConfig.ToolList)
		if err != nil {
			return nil, err
		}
		if agentID != config.WgaCfg().AgentID {
			checkResult, err := wga.CheckToolOptions(ctx.Request.Context(), agentID, toolOpts...)
			if err != nil {
				return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("tool options check failed: %v", err))
			}
			for _, tc := range checkResult.ToolCategories {
				if !tc.Meet {
					return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("tool category (%s) condition (%s) not meet", tc.Category, tc.Condition))
				}
			}
		}
		opts = append(opts, toolOpts...)
	}

	// 校验并构建工作流配置选项（追加@提及的工作流）
	workflowList := append([]*assistant_service.WgaConfigWorkflow{}, wgaConfig.WorkflowList...)
	workflowList = append(workflowList, mentionResources.WorkflowList...)
	if len(workflowList) > 0 {
		// 去重
		seen := make(map[string]bool)
		dedupedList := make([]*assistant_service.WgaConfigWorkflow, 0, len(workflowList))
		for _, w := range workflowList {
			if !seen[w.WorkflowId] {
				seen[w.WorkflowId] = true
				dedupedList = append(dedupedList, w)
			}
		}
		if err := checkWgaWorkflowConfig(ctx, userID, orgID, dedupedList); err != nil {
			return nil, err
		}
		workflowOpts, err := buildWgaWorkflowOptions(ctx, userID, orgID, dedupedList)
		if err != nil {
			return nil, err
		}
		opts = append(opts, workflowOpts...)
	}

	// 校验并构建MCP配置选项（追加@提及的MCP）
	mcpList := append([]*assistant_service.WgaConfigMcp{}, wgaConfig.McpList...)
	mcpList = append(mcpList, mentionResources.McpList...)
	if len(mcpList) > 0 {
		// 去重
		seen := make(map[string]bool)
		dedupedList := make([]*assistant_service.WgaConfigMcp, 0, len(mcpList))
		for _, m := range mcpList {
			if !seen[m.McpId] {
				seen[m.McpId] = true
				dedupedList = append(dedupedList, m)
			}
		}
		if err := checkWgaMCPConfig(ctx, userID, orgID, dedupedList); err != nil {
			return nil, err
		}
		mcpOpts, err := buildWgaMCPOptions(ctx, userID, orgID, dedupedList)
		if err != nil {
			return nil, err
		}
		opts = append(opts, mcpOpts...)
	}

	// 校验并构建智能体配置选项（追加@提及的智能体）
	assistantList := append([]*assistant_service.WgaConfigAssistant{}, wgaConfig.AssistantList...)
	assistantList = append(assistantList, mentionResources.AssistantList...)
	if len(assistantList) > 0 {
		// 去重
		seen := make(map[string]bool)
		dedupedList := make([]*assistant_service.WgaConfigAssistant, 0, len(assistantList))
		for _, a := range assistantList {
			if !seen[a.AssistantId] {
				seen[a.AssistantId] = true
				dedupedList = append(dedupedList, a)
			}
		}
		if err := checkWgaAssistantConfig(ctx, userID, orgID, dedupedList); err != nil {
			return nil, err
		}
		assistantOpts, err := buildWgaAssistantOptions(ctx, userID, orgID, dedupedList)
		if err != nil {
			return nil, err
		}
		opts = append(opts, assistantOpts...)
	}

	// 校验并构建Skills配置选项（追加@提及的Skills）
	skillList := append([]*assistant_service.WgaConfigSkill{}, wgaConfig.SkillList...)
	skillList = append(skillList, mentionResources.SkillList...)
	if len(skillList) > 0 {
		// 去重（使用 skillId + skillType 作为唯一标识）
		seen := make(map[string]bool)
		dedupedList := make([]*assistant_service.WgaConfigSkill, 0, len(skillList))
		for _, s := range skillList {
			key := s.SkillId + ":" + s.SkillType
			if !seen[key] {
				seen[key] = true
				dedupedList = append(dedupedList, s)
			}
		}
		// 运行时过滤无效 skill
		validSkillList, err := checkWgaSkillConfig(ctx, userID, orgID, dedupedList)
		if err != nil {
			return nil, err
		}
		if len(validSkillList) > 0 {
			skillOpts, err := buildWgaSkillOptions(ctx, userID, orgID, threadID, runID, validSkillList)
			if err != nil {
				return nil, err
			}
			opts = append(opts, skillOpts...)
		}
	}
	builtinSkillMessage, err := buildBuiltinSkillMessage(ctx, userID, orgID)
	if err != nil {
		return nil, err
	}

	// 校验并构建Knowledge配置选项（追加@提及的Knowledge）
	knowledgeList := append([]*assistant_service.WgaConfigKnowledge{}, wgaConfig.KnowledgeList...)
	knowledgeList = append(knowledgeList, mentionResources.KnowledgeList...)
	if len(knowledgeList) > 0 {
		// 去重
		seen := make(map[string]bool)
		dedupedList := make([]*assistant_service.WgaConfigKnowledge, 0, len(knowledgeList))
		for _, k := range knowledgeList {
			if !seen[k.KnowledgeId] {
				seen[k.KnowledgeId] = true
				dedupedList = append(dedupedList, k)
			}
		}
		if err := checkWgaKnowledgeConfig(ctx, userID, orgID, dedupedList); err != nil {
			return nil, err
		}
		knowledgeOpts, err := buildWgaKnowledgeOptions(ctx, userID, orgID, threadID, runID, dedupedList)
		if err != nil {
			return nil, err
		}
		opts = append(opts, knowledgeOpts...)
	}

	// 校验并构建Ontology配置选项
	var ontologyOpts []wga_option.Option
	var ontologyMessage *schema.Message
	if config.Cfg().Ontology.Enable != 0 {
		if agentID == "DIP Agent" {
			ontologyOpts, ontologyMessage, err = buildWgaOntologyDIPMode(ctx, userID, orgID, threadID, runID, strings.TrimSpace(userInputMessage.GetTextContent()))
		} else {
			ontologyOpts, ontologyMessage, err = buildWgaOntologyNonDIPMode(ctx, userID, orgID, mentionResources.OntologyList, wgaConfig.OntologyKnowledgeList)
		}
	}
	if err != nil {
		return nil, err
	}
	opts = append(opts, ontologyOpts...)

	// Skill Preview Agent 模式，需要构建 skill 变量的 schema.Message
	var skillMessage *schema.Message
	if agentID == generalAgentSkillChatPreviewAgentID {
		resp, err := mcp.GetCustomSkillByPreviewID(ctx.Request.Context(), &mcp_service.GetCustomSkillByPreviewIDReq{
			PreviewThreadId: threadID,
		})
		if err != nil {
			return nil, err
		}
		if customSkill := resp.GetSkill(); customSkill != nil {
			variables, err := getCustomSkillVariables(ctx, customSkill.GetSkillId())
			if err != nil {
				return nil, err
			}
			if len(variables) > 0 {
				skillMessage = &schema.Message{
					Role:    schema.System,
					Content: buildWgaSkillVariablesMessage(customSkill, variables),
				}
			}
		}
	}

	// 历史消息 + @资源提示消息 + 当前用户消息
	messages, err := filterWgaHistoryMessages(ctx, userID, orgID, threadID)
	if err != nil {
		return nil, err
	}
	if mentionResources.hasResources() {
		messages = append(messages, buildWgaMentionResourcesMessage(mentionResources))
	}
	// 追加技能变量提示消息（仅 Skill Preview Agent 模式）
	if skillMessage != nil {
		messages = append(messages, skillMessage)
	}
	// 追加内置技能提示消息
	if builtinSkillMessage != nil {
		messages = append(messages, builtinSkillMessage)
	}
	// 追加Ontology提示消息
	if ontologyMessage != nil {
		messages = append(messages, ontologyMessage)
	}
	// 当前用户消息放在最后
	messages = append(messages, convertWgaMessage(userInputMessage.Role, userInputMessage.Content))
	opts = append(opts, wga_option.WithMessages(messages))

	// 持久化存储（最后一步，确保前面所有配置校验通过才创建工作空间目录）
	if workspaceStore != nil {
		dirs, err := PrepareWgaWorkspaceDirs(workspaceStore, runID, true)
		if err != nil {
			log.Errorf("[wga] thread %v prepare input output dir err: %v", threadID, err)
		} else {
			if agentID == generalAgentSkillChatNormalAgentID || agentID == generalAgentSkillChatImportAgentID || agentID == generalAgentSkillChatPreviewAgentID {
				if dirs, err = applyWgaWorkspaceDirSkillPolicy(dirs); err != nil {
					log.Errorf("[wga] thread %v apply workspace dir policy err: %v", threadID, err)
					return nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, err.Error())
				}
			}
			opts = append(opts, wga_option.WithInputDir(dirs.InputDir))
			if !workspaceReadOnly {
				opts = append(opts, wga_option.WithOutputDir(dirs.OutputDir))
			}
			if urls := userInputMessage.GetURLs(); len(urls) > 0 {
				if err := DownloadWgaWorkspaceURLs(ctx.Request.Context(), urls, dirs.OutputDir); err != nil {
					log.Errorf("[wga] thread %v download URLs %+v to workspace dir %v failed: %v", threadID, urls, dirs.OutputDir, err)
				}
			}
		}
	}

	return opts, nil
}

// saveWgaChatHistoryEvent 保存智能体返回的聊天历史到 ES
func saveWgaChatHistoryEvent(
	ctx context.Context,
	historyEventCh <-chan aguievents.Event,
	userId, orgId, threadId, runId string,
	workspaceStore *wga_persistent.Store,
	lastWorkspaceTotalSize int64,
	lastWorkspaceFileCount int) {
	defer util.PrintPanicStack()

	var events []aguievents.Event
	for event := range historyEventCh {
		// inject workspace activity event when run finished, to trigger AG-UI to fetch workspace info and update workspace activity card
		if workspaceStore != nil && event.Type() == aguievents.EventTypeRunFinished {
			if wsEvent, _ := BuildWgaWorkspaceEvent(workspaceStore, &WgaWorkspaceEventConfig{
				ThreadID:           threadId,
				RunID:              runId,
				LastWorkspaceSize:  lastWorkspaceTotalSize,
				LastWorkspaceCount: lastWorkspaceFileCount,
			}); wsEvent != nil {
				events = append(events, wsEvent)
			}
		}
		events = append(events, event)
	}
	b, _ := json.Marshal(events)

	doc := map[string]interface{}{
		"id":        util.GenUUID(),
		"threadId":  threadId,
		"runId":     runId,
		"userId":    userId,
		"orgId":     orgId,
		"createdAt": time.Now().UnixMilli(),
		"events":    string(b),
	}
	docJson, err := json.Marshal(doc)
	if err != nil {
		log.Warnf("[wga] marshal history doc failed: %v", err)
		return
	}

	_, err = assistant.SaveToES(ctx, &assistant_service.SaveToESReq{
		IndexName: wgaConversationHistoryEventESIndexName,
		DocJson:   string(docJson),
	})
	if err != nil {
		log.Warnf("[wga] save history to ES failed: %v", err)
	}
}

// buildWgaMentionResourcesMessage 构建 @ 资源提示消息
func buildWgaMentionResourcesMessage(resources *wgaMentionResources) *schema.Message {
	var lines []string
	lines = append(lines, "本次对话已加载用户 @ 提及的资源：")

	for _, item := range resources.McpItems {
		lines = append(lines, fmt.Sprintf("- @%s (mcp) %s", item.Name, item.Desc))
	}
	for _, item := range resources.WorkflowItems {
		lines = append(lines, fmt.Sprintf("- @%s (workflow) %s", item.Name, item.Desc))
	}
	for _, item := range resources.SkillItems {
		lines = append(lines, fmt.Sprintf("- @%s (skill) %s", item.Name, item.Desc))
	}
	for _, item := range resources.AssistantItems {
		lines = append(lines, fmt.Sprintf("- @%s (assistant) %s", item.Name, item.Desc))
	}
	for _, item := range resources.KnowledgeItems {
		lines = append(lines, fmt.Sprintf("- @%s (knowledge) %s", item.Name, item.Desc))
	}
	for _, item := range resources.OntologyItems {
		lines = append(lines, fmt.Sprintf("- @%s (ontology knowledge network) %s", item.Name, item.Desc))
	}

	return &schema.Message{
		Role:    schema.System,
		Content: strings.Join(lines, "\n"),
	}
}

func convertWgaMessage(role string, content interface{}) *schema.Message {
	switch v := content.(type) {
	case string:
		return &schema.Message{
			Role:    schema.RoleType(role),
			Content: v,
		}
	case []interface{}:
		var parts []schema.MessageInputPart
		for _, msg := range v {
			if m, ok := msg.(map[string]interface{}); ok {
				parts = append(parts, convertWgaMessageInputPart(m))
			}
		}
		return &schema.Message{
			Role:                  schema.RoleType(role),
			UserInputMultiContent: parts,
		}
	default:
		return &schema.Message{Role: schema.RoleType(role)}
	}
}

func convertWgaMessageInputPart(m map[string]interface{}) schema.MessageInputPart {
	part := schema.MessageInputPart{}

	typ, _ := m["type"].(string)
	switch typ {
	case "text":
		part.Type = schema.ChatMessagePartTypeText
		if text, ok := m["text"].(string); ok {
			part.Text = text
		}
	case "binary":
		mimeType, _ := m["mimeType"].(string)
		url, _ := m["url"].(string)
		switch {
		case strings.HasPrefix(mimeType, "image/"):
			part.Type = schema.ChatMessagePartTypeImageURL
			part.Image = &schema.MessageInputImage{
				MessagePartCommon: schema.MessagePartCommon{
					URL:      &url,
					MIMEType: mimeType,
				},
			}
		case strings.HasPrefix(mimeType, "audio/"):
			part.Type = schema.ChatMessagePartTypeAudioURL
			part.Audio = &schema.MessageInputAudio{
				MessagePartCommon: schema.MessagePartCommon{
					URL:      &url,
					MIMEType: mimeType,
				},
			}
		case strings.HasPrefix(mimeType, "video/"):
			part.Type = schema.ChatMessagePartTypeVideoURL
			part.Video = &schema.MessageInputVideo{
				MessagePartCommon: schema.MessagePartCommon{
					URL:      &url,
					MIMEType: mimeType,
				},
			}
		default:
			// github.com/cloudwego/eino-ext/libs/acl/openai/chat_model.go:485 v0.1.1
			// 该版本暂未支持 schema.ChatMessagePartTypeFileURL 处理，暂用 schema.ChatMessagePartTypeText 替代
			// part.Type = schema.ChatMessagePartTypeFileURL
			// part.File = &schema.MessageInputFile{
			// 	MessagePartCommon: schema.MessagePartCommon{
			// 		URL:      &url,
			// 		MIMEType: mimeType,
			// 	},
			// }
			part.Type = schema.ChatMessagePartTypeText
			if fileName, _ := m["fileName"].(string); fileName != "" {
				part.Text = fmt.Sprintf("[%s](%s)", fileName, url)
			} else {
				part.Text = url
			}
		}
	default:
		part.Type = schema.ChatMessagePartType(typ)
		if text, ok := m["text"].(string); ok {
			part.Text = text
		}
	}
	return part
}

// ============================================================================
// WGA 断线重连 - Service 层
// ============================================================================

// WgaConversationPending 查询是否有运行中的 WGA 会话
func WgaConversationPending(ctx *gin.Context, userId, orgId, clientId string, req request.WgaConversationPendingReq) (*response.WgaConversationPendingResp, error) {
	session := &sse_model.Session{
		ConversationID: req.ThreadID,
		ClientID:       clientId,
	}
	mgr := sse_connector.GetSession(session)
	if mgr == nil {
		return &response.WgaConversationPendingResp{
			ThreadID:               req.ThreadID,
			HasPendingConversation: false,
		}, nil
	}
	var messages []request.GeneralAgentConversationMessage
	if ext := mgr.GetExt(); ext != nil {
		msgs, _ := ext["messages"].([]request.GeneralAgentConversationMessage)
		for _, msg := range msgs {
			messages = append(messages, msg.WithMinioUrlReplaced())
		}
	}
	return &response.WgaConversationPendingResp{
		ThreadID:               req.ThreadID,
		HasPendingConversation: true,
		Messages:               messages,
	}, nil
}

// WgaConversationConnect WGA 流式问答断线重连
func WgaConversationConnect(ctx *gin.Context, userId, orgId, clientId string, req request.WgaConversationConnectReq) error {
	session := &sse_model.Session{
		ConversationID: req.ThreadID,
		ClientID:       clientId,
	}
	chatCh, err := sse_connector.Connect[string](ctx.Request.Context(), session, func(data *sse_model.Message) string {
		return data.Data.(string)
	})
	if err != nil {
		return err
	}

	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")

	ctx.Stream(func(w io.Writer) bool {
		select {
		case line, ok := <-chatCh:
			if !ok {
				return false
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", line)
			return true
		case <-ctx.Request.Context().Done():
			return false
		}
	})
	return nil
}

// WgaConversationCancel WGA 流式问答手动停止
func WgaConversationCancel(ctx *gin.Context, userId, orgId, clientId string, req request.WgaConversationCancelReq) error {
	session := &sse_model.Session{
		ConversationID: req.ThreadID,
		ClientID:       clientId,
	}
	return sse_connector.Close(session)
}
