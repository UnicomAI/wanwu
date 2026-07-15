package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	ag_ui_util "github.com/UnicomAI/wanwu/pkg/ag-ui-util"
	"github.com/UnicomAI/wanwu/pkg/constant"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	aguievents "github.com/ag-ui-protocol/ag-ui/sdks/community/go/pkg/core/events"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- rag-service 业务返回码（chunk.Code）---
// 0/1：正常（0=成功、1=流式中间帧）；非 0/1 一律视为错误。
// 7：业务失败（如模型调用失败、检索失败等），对应 RagErrCodeUpstream；
//
//	其他非 0/1 的 code 归类为 RagErrCodeUnknown 兜底。
const (
	ragChunkCodeBusinessError = 7
)

// ragChatStreamParams 记录流式请求的过程参数（首 token 延迟、错误标志等）
type ragChatStreamParams struct {
	ctx               *gin.Context
	startTime         time.Time
	firstTokenLatency int64
	hasRecorded       bool
	hasErr            bool
}

// ragChunkData 对应 rag-service / rag-wanwu 返回的每条 SSE JSON 结构
type ragChunkData struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	MsgID   string         `json:"msg_id"`
	MsgType string         `json:"msg_type"`
	Data    *ragChunkInner `json:"data"`
	Finish  int            `json:"finish"`
}

// ragChunkInner 对应 data 字段，SearchList 保持 json.RawMessage 以便原样透传
type ragChunkInner struct {
	Output           string          `json:"output"`
	ReasoningContent string          `json:"reasoning_content"`
	SearchList       json.RawMessage `json:"searchList"`
}

// ChatRagStream RAG 私域问答，流式返回 AG-UI 协议事件
func ChatRagStream(ctx *gin.Context, userId, orgId string, req request.ChatRagRequest, needLatestPublished bool, source string) (err error) {
	streamParams := &ragChatStreamParams{ctx: ctx, startTime: time.Now()}
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	defer func() {
		if source != constant.AppStatisticSourceDraft {
			go func() {
				defer util.PrintPanicStack()
				RecordAppStatistic(detachedCtx, userId, orgId, req.RagID, constant.AppTypeRag, !streamParams.hasErr, true, streamParams.firstTokenLatency, 0, source)
			}()
		}
	}()

	chatCh, kbNameMap, err := CallRagChatStream(ctx, userId, orgId, req, needLatestPublished)
	if err != nil {
		streamParams.hasErr = true
		return err
	}

	// AG-UI 协议要求 threadId/runId 每次 run 唯一；RAG 当前无持久化会话概念，
	// 两者均使用 uuid（若后续引入 conversationID，可以把 threadID 换成它）
	runID := uuid.NewString()
	threadID := uuid.NewString()

	eventCh := convertRagStream2AGUIEvents(ctx.Request.Context(), chatCh, threadID, runID, streamParams, kbNameMap)
	outputCh := ag_ui_util.EventsToJSONChannel(ctx.Request.Context(), eventCh)

	//流式返回结果
	return sse_util.NewSSEWriter(ctx, fmt.Sprintf("[RAG-Stream] %v user %v org %v", req.RagID, userId, orgId), "").
		WriteStream(outputCh, streamParams, buildRagChatResp(), nil)
}

func buildRagChatResp() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		return fmt.Sprintf("data: %s\n\n", lineText), false, nil
	}
}

// parseChunkLine 解析一行 SSE 文本为 ragChunkData；不合法或空行返回 ok=false。
// 额外识别 rag-service 的裸 `error:` 前缀行（见 rag_manage_sevice.go 里
// requestRagStreamChat 的错误返回格式），合成一个 business-error chunk
// 交给 handleChunk 走统一的 RUN_ERROR 路径。
func parseChunkLine(line string) (ragChunkData, bool) {
	line = strings.TrimPrefix(line, "data:")
	line = strings.TrimSpace(line)
	if line == "" {
		return ragChunkData{}, false
	}
	if strings.HasPrefix(line, "error:") {
		return ragChunkData{
			Code:    ragChunkCodeBusinessError,
			Message: strings.TrimSpace(strings.TrimPrefix(line, "error:")),
		}, true
	}
	var chunk ragChunkData
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return ragChunkData{}, false
	}
	return chunk, true
}

// convertRagStream2AGUIEvents 将 RAG 原始 SSE channel 转换为 AG-UI 事件 channel。
// 详细转换规则见 ragStreamConverter 文档注释。
func convertRagStream2AGUIEvents(
	ctx context.Context,
	chatCh <-chan string,
	threadID, runID string,
	streamParams *ragChatStreamParams,
	kbNameMap map[string]string,
) <-chan aguievents.Event {
	out := make(chan aguievents.Event, 64)
	// NewBaseState 返回值类型，方法接收者是 *BaseState；取地址避免方法调用时拷贝状态
	state := ag_ui_util.NewBaseState(threadID, runID)
	c := &ragStreamConverter{
		ctx:          ctx,
		out:          out,
		state:        &state,
		runID:        runID,
		streamParams: streamParams,
		kbNameMap:    kbNameMap,
	}

	go func() {
		defer util.PrintPanicStack()
		defer close(out)

		// 首条：RUN_STARTED
		if !c.emit(c.state.EnsureRunStarted()...) {
			return
		}

		for {
			select {
			case <-ctx.Done():
				return
			case line, ok := <-chatCh:
				if !ok {
					// 上游 channel 关闭。分两种情况：
					//  - 已 finalize（正常 RUN_FINISHED 或已发过 RUN_ERROR）：直接退出
					//  - 未 finalize（上游异常断流，例如 rag-wanwu 遇到模型不可用
					//    但只关流没发错误帧）：兜底发 RUN_ERROR，避免前端卡 loading
					if !c.hasFinalized {
						c.streamParams.hasErr = true
						c.finalizeError(RagErrCodeUnknown, "upstream stream closed without finish")
					}
					return
				}
				chunk, ok := parseChunkLine(line)
				if !ok {
					continue
				}
				if c.handleChunk(chunk) {
					return
				}
			}
		}
	}()

	return out
}

// ChatRagStreamLegacy 旧版 RAG 流式接口（原样透传 rag-service SSE JSON）。
//
// 历史背景：新分支把 web 端 RAG 流式响应迁移到 AG-UI 协议（ChatRagStream），
// 但 /openapi/rag/chat 是对外暴露给第三方的 OpenAPI，已有外部集成方按旧格式
// 解析 `data: {"code":0,"msg_id":...,"data":{"output":"...","searchList":[...]},"finish":0|1}`，
// 不能随 web 一起改协议。故保留这份旧实现专供 openapi 使用：
//   - web / 草稿预览  → ChatRagStream（AG-UI 事件流）
//   - openapi         → ChatRagStreamLegacy（原始 SSE JSON 透传）
//
// 两者共用同一个底层 CallRagChatStream，仅输出层不同。
func ChatRagStreamLegacy(ctx *gin.Context, userId, orgId string, req request.ChatRagRequest, needLatestPublished bool, source string) (err error) {
	streamParams := &ragChatStreamParams{ctx: ctx, startTime: time.Now()}
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	defer func() {
		if source != constant.AppStatisticSourceDraft {
			go func() {
				defer util.PrintPanicStack()
				RecordAppStatistic(detachedCtx, userId, orgId, req.RagID, constant.AppTypeRag, !streamParams.hasErr, true, streamParams.firstTokenLatency, 0, source)
			}()
		}
	}()

	// openapi 不需要 kbNameMap（旧格式没有 user_kb_name 字段），忽略第二个返回值
	chatCh, _, err := CallRagChatStream(ctx, userId, orgId, req, needLatestPublished)
	if err != nil {
		streamParams.hasErr = true
		return err
	}
	// 旧版行处理器：带 data: 前缀的原样透传，error: 开头的转成 {code:-1,...}
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[RAG] %v user %v org %v", req.RagID, userId, orgId), sse_util.DONE_MSG).
		WriteStream(chatCh, streamParams, buildRagChatRespLineProcessorLegacy(), nil)
	return nil
}

// buildRagChatRespLineProcessorLegacy 旧版 RAG 行处理器（仅 ChatRagStreamLegacy 使用）。
// 用于保证 openapi 输出格式向后兼容
func buildRagChatRespLineProcessorLegacy() func(sse_util.SSEWriterClient[string], string, interface{}) (string, bool, error) {
	return func(c sse_util.SSEWriterClient[string], lineText string, params interface{}) (string, bool, error) {
		if p, ok := params.(*ragChatStreamParams); ok {
			if !p.hasRecorded {
				p.firstTokenLatency = time.Since(p.startTime).Milliseconds()
				p.hasRecorded = true
				if p.ctx != nil {
					p.ctx.Set(gin_util.FIRST_RESP_LATENCY, p.firstTokenLatency)
				}
			}
		}
		if strings.HasPrefix(lineText, "error:") {
			if p, ok := params.(*ragChatStreamParams); ok {
				p.hasErr = true
			}
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		return lineText + "\n\n", false, nil
	}
}

// CallRagChatStream 调用 Rag 对话，返回经敏感词处理后的原始 SSE 字符串 channel。
// 第二个返回值 kbNameMap 是 rag 内部 kb_name → 用户可见知识库名的映射，
// 供上层在透传 searchList 前为每个引用段落补填 user_kb_name。
func CallRagChatStream(ctx *gin.Context, userId, orgId string, req request.ChatRagRequest, needLatestPublished bool) (<-chan string, map[string]string, error) {
	ragInfo, kbNameMap, err := buildRagInfo(ctx, userId, orgId, req, needLatestPublished)
	if err != nil {
		return nil, nil, err
	}
	sensitiveConfig := ragInfo.SensitiveConfig
	//创建敏感词校验器
	sensitiveChecker := CreateSensitiveChecker(sensitiveConfig.GetTableIds(), &ragSensitiveService{}, sensitiveConfig.Enable)
	//任务执行器
	streamExecutor := ragStream(ctx, userId, orgId, req, needLatestPublished)
	//带敏感词校验的任务执行
	retCh, err := sensitiveChecker.Check(ctx, req.Question, streamExecutor)
	if err != nil {
		return nil, nil, err
	}
	return retCh, kbNameMap, nil
}

// rag流式会话
func ragStream(ctx *gin.Context, userId, orgId string, req request.ChatRagRequest, needLatestPublished bool) func() (ch <-chan string, callback func(string, string), err error) {
	return func() (ch <-chan string, callback func(string, string), err error) {
		stream, err := rag.ChatRag(ctx.Request.Context(), buildRagStreamParams(userId, orgId, req, needLatestPublished))
		if err != nil {
			return nil, nil, err
		}

		// 读取 gRPC 流内容到 channel
		SSEReader := &sse_util.SSEReader[rag_service.ChatRagResp]{
			BusinessKey:    "chat_rag",
			StreamReceiver: sse_util.NewGrpcStreamReceiver(stream),
		}
		rawCh, err := SSEReader.ReadStreamWithBuilder(ctx, func(resp *rag_service.ChatRagResp) string {
			return resp.Content
		})
		if err != nil {
			return nil, nil, err
		}
		return rawCh, nil, nil
	}
}

func buildRagInfo(ctx *gin.Context, userId, orgId string, req request.ChatRagRequest, needLatestPublished bool) (*rag_service.RagInfo, map[string]string, error) {
	// 根据 ragID 获取敏感词配置
	ragInfo, err := rag.GetRagDetail(ctx.Request.Context(), &rag_service.RagDetailReq{
		RagId:   req.RagID,
		Publish: util.IfElse(needLatestPublished, int32(1), int32(0)),
	})
	if err != nil {
		return nil, nil, err
	}
	// 构造 kb_name → user_kb_name 映射（失败时仅退化为空 map，不中断对话）
	kbNameMap := buildRagKbNameMap(ctx, userId, ragInfo)
	return ragInfo, kbNameMap, nil
}

// buildRagStreamParams 构造 rag 流式问答参数
func buildRagStreamParams(userId, orgId string, req request.ChatRagRequest, needLatestPublished bool) *rag_service.ChatRagReq {
	var ragHistory []*rag_service.HistoryItem
	if len(req.History) > 0 {
		for _, history := range req.History {
			ragHistory = append(ragHistory, &rag_service.HistoryItem{
				Query:       history.Query,
				Response:    history.Response,
				NeedHistory: history.NeedHistory,
			})
		}
	}
	return &rag_service.ChatRagReq{
		RagId:    req.RagID,
		Question: req.Question,
		History:  ragHistory,
		Identity: &rag_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
		Publish:      util.IfElse(needLatestPublished, int32(1), int32(0)),
		FileInfoList: buildRagFileInfoList(req.FileInfo),
	}
}

// buildRagKbNameMap 根据 RAG 应用的知识库/问答库绑定，查出每个知识库的用户可见名，
// 返回 rag 内部 kb_name（即 KnowledgeInfo.RagName）→ 用户可见名（KnowledgeInfo.Name）的映射。
// 任何上游错误都降级为空 map，不阻断对话主流程。
func buildRagKbNameMap(ctx *gin.Context, userId string, ragInfo *rag_service.RagInfo) map[string]string {
	if ragInfo == nil {
		return map[string]string{}
	}
	idSet := make(map[string]struct{})
	if kbCfg := ragInfo.GetKnowledgeBaseConfig(); kbCfg != nil {
		for _, per := range kbCfg.GetPerKnowledgeConfigs() {
			if id := per.GetKnowledgeId(); id != "" {
				idSet[id] = struct{}{}
			}
		}
	}
	if qaCfg := ragInfo.GetQAknowledgeBaseConfig(); qaCfg != nil {
		for _, per := range qaCfg.GetPerKnowledgeConfigs() {
			if id := per.GetKnowledgeId(); id != "" {
				idSet[id] = struct{}{}
			}
		}
	}
	if len(idSet) == 0 {
		return map[string]string{}
	}
	idList := make([]string, 0, len(idSet))
	for id := range idSet {
		idList = append(idList, id)
	}
	list, err := selectKnowledgeListByIdList(ctx, &request.KnowledgeBatchSelectReq{
		UserId:          userId,
		KnowledgeIdList: idList,
	})
	if err != nil || list == nil {
		return map[string]string{}
	}
	nameMap := make(map[string]string, len(list.KnowledgeList))
	for _, kb := range list.KnowledgeList {
		if kb == nil || kb.RagName == "" {
			continue
		}
		nameMap[kb.RagName] = kb.Name
	}
	return nameMap
}

// enrichSearchListWithUserKbName 在透传上游 searchList 前，为每个引用段落补填 user_kb_name 字段。
// 实现策略：解析为松散的 []map[string]interface{}，按每项的 kb_name 查 nameMap 写入 user_kb_name。
// 保持其他字段原样透传；解析失败时回退为原始 RawMessage，保证至少不破坏下游渲染。
func enrichSearchListWithUserKbName(raw json.RawMessage, nameMap map[string]string) []map[string]interface{} {
	items := []map[string]interface{}{}
	if len(raw) == 0 {
		return items
	}
	if err := json.Unmarshal(raw, &items); err != nil {
		return items
	}
	for _, item := range items {
		if item == nil {
			continue
		}
		// 已有 user_kb_name（上游未来可能直接返回）则不覆盖
		if existing, ok := item["user_kb_name"].(string); ok && existing != "" {
			continue
		}
		kbName, _ := item["kb_name"].(string)
		if kbName == "" {
			// QA 条目没有 kb_name，用 QABase 作 key
			kbName, _ = item["QABase"].(string)
		}
		if kbName == "" {
			continue
		}
		if display, ok := nameMap[kbName]; ok && display != "" {
			item["user_kb_name"] = display
		} else {
			// 兜底：即便映射缺失也让前端能显示一个名字，而不是空白
			item["user_kb_name"] = kbName
		}
	}
	// 老版本 rag-wanwu（如 64 服务器）顶层不返回 score，只在 rerank_info[0].score 里有；
	// 为让前端 Score 徽章始终能显示，顶层缺失时从 rerank_info 抽上来。
	// TODO: 64 服务器 rag-wanwu 后续会更新，届时确认顶层已有 score 后删除此段。
	// for _, item := range items {
	// 	if item == nil {
	// 		continue
	// 	}
	// 	if _, ok := item["score"].(float64); ok {
	// 		continue
	// 	}
	// 	rerank, ok := item["rerank_info"].([]interface{})
	// 	if !ok || len(rerank) == 0 {
	// 		continue
	// 	}
	// 	first, ok := rerank[0].(map[string]interface{})
	// 	if !ok {
	// 		continue
	// 	}
	// 	if s, ok := first["score"].(float64); ok {
	// 		item["score"] = s
	// 	}
	// }
	return items
}

func buildRagFileInfoList(fileInfoList []request.ConversionStreamFile) []*rag_service.FileInfo {
	retList := make([]*rag_service.FileInfo, 0)
	if len(fileInfoList) > 0 {
		for _, fileInfo := range fileInfoList {
			retList = append(retList, &rag_service.FileInfo{
				FileName: fileInfo.FileName,
				FileSize: fileInfo.FileSize,
				FileUrl:  fileInfo.FileUrl,
			})
		}
	}
	return retList
}

// --- ragSensitiveService: 实现 sensitiveService 接口，供 ProcessSensitiveWords 使用 ---

type ragSensitiveService struct{}

func (s *ragSensitiveService) serviceType() string {
	return constant.AppTypeRag
}

func (s *ragSensitiveService) parseContent(raw string) (id, content string) {
	// 1. 清理数据前缀
	raw = strings.TrimPrefix(raw, "data:")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	// 2. 解析 JSON
	resp := struct {
		MsgID string `json:"msg_id"`
		Data  struct {
			Output string `json:"output"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", ""
	}
	// 3. 返回 content
	return resp.MsgID, resp.Data.Output
}

func (s *ragSensitiveService) buildSensitiveResp(id string, content string) []string {
	resp := map[string]interface{}{
		"code":    0,
		"message": "success",
		"msg_id":  id,
		"data": map[string]interface{}{
			"output":     content,
			"searchList": []interface{}{},
		},
		"history": []interface{}{},
		"finish":  1,
	}
	marshal, _ := json.Marshal(resp)
	return []string{"data: " + string(marshal)}
}
