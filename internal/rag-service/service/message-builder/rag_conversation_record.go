package message_builder

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	minio_service "github.com/UnicomAI/wanwu/internal/rag-service/service/minio-service"
	"github.com/UnicomAI/wanwu/pkg/es"
	"github.com/UnicomAI/wanwu/pkg/log"
	minio_client "github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/redis"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	esTimeout          = 1 * time.Minute
	ragChunkCodeNormal = 0
)

// ragRecordChunk 只取落库需要的字段，其余原样忽略
type ragRecordChunk struct {
	Code       int                 `json:"code"`
	Message    string              `json:"message"`
	MsgType    RagMessageType      `json:"msg_type"`
	Data       *ragRecordChunkData `json:"data"`
	ErrMessage string              `json:"errMessage"`
}

type ragRecordChunkData struct {
	Output           string          `json:"output"`
	ReasoningContent string          `json:"reasoning_content"`
	SearchList       json.RawMessage `json:"searchList"`
}

// ragPhaseTimes 三个阶段的起止时刻，零值表示该阶段未发生。
type ragPhaseTimes struct {
	qaStart, qaEnd       time.Time // qa_start 帧 → qa_finish 帧
	kbStart, kbEnd       time.Time // knowledge_start 帧 → 首个非空 searchList 帧
	thinkStart, thinkEnd time.Time // 首个 reasoning_content 帧 → 首个 output 帧
}

// ragChatRecorder 累积一轮问答的流式输出
type ragChatRecorder struct {
	firstToken   time.Time // 首个带输出的帧到达时刻，零值表示整轮无输出
	phase        ragPhaseTimes
	response     strings.Builder
	reasoning    strings.Builder
	searchList   string
	qaSearchList string
	errMessage   string
	qaErrMessage string
}

// record 累积一行流式输出，解析失败只跳过该行，不影响对话主流程。
func (r *ragChatRecorder) record(line string) {
	defer util.PrintPanicStack()
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	if strings.HasPrefix(line, "error:") {
		r.errMessage = strings.TrimSpace(strings.TrimPrefix(line, "error:"))
		return
	}
	line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
	if line == "" {
		return
	}
	var chunk ragRecordChunk
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return
	}
	if chunk.Code != ragChunkCodeNormal {
		r.errMessage = chunk.Message
	}
	if chunk.MsgType == QAError {
		r.qaErrMessage = chunk.ErrMessage
	}
	if chunk.Data == nil {
		return
	}
	now := time.Now()
	// 首token时延按"用户看到的第一个字"计，qa_start/knowledge_start 这类空输出帧不算
	if r.firstToken.IsZero() && (chunk.Data.Output != "" || chunk.Data.ReasoningContent != "") {
		r.firstToken = now
	}
	searchList := nonEmptySearchList(chunk.Data.SearchList)
	r.markPhase(&chunk, searchList, now)
	// output/reasoning_content 都是增量，直接拼接
	r.response.WriteString(chunk.Data.Output)
	r.reasoning.WriteString(chunk.Data.ReasoningContent)
	if searchList != "" {
		if chunk.MsgType == QAFinish {
			r.qaSearchList = searchList
		} else {
			r.searchList = searchList
		}
	}
}

// nonEmptySearchList 返回去空白后的检索结果，空数组/null 一律当没有
func nonEmptySearchList(raw json.RawMessage) string {
	s := strings.TrimSpace(string(raw))
	if s == "" || s == "[]" || s == "null" {
		return ""
	}
	return s
}

// markPhase 按帧记录三个阶段的起止时刻
func (r *ragChatRecorder) markPhase(chunk *ragRecordChunk, searchList string, now time.Time) {
	switch chunk.MsgType {
	case QAStart:
		setOnce(&r.phase.qaStart, now)
	case QAFinish, QAError:
		// 命中、未命中且没配知识库、检索报错
		setOnce(&r.phase.qaEnd, now)
	case KnowledgeStart:
		// 未命中但配了知识库时 QAFinishBuilder 直接 Skip
		setOnce(&r.phase.qaEnd, now)
		setOnce(&r.phase.kbStart, now)
	}
	// 知识库检索以首个非空 searchList 收尾；qa_finish 帧的 searchList 属于问答库，不算
	if searchList != "" && chunk.MsgType != QAFinish {
		setOnce(&r.phase.kbEnd, now)
	}
	if chunk.Data.ReasoningContent != "" {
		setOnce(&r.phase.thinkStart, now)
	}
	if chunk.Data.Output != "" {
		// 首个正文帧即思考结束，与 bff 在此处发 REASONING_MESSAGE_END 同一时机
		setOnce(&r.phase.thinkEnd, now)
	}
}

func setOnce(dst *time.Time, now time.Time) {
	if dst.IsZero() {
		*dst = now
	}
}

// save 落一条问答明细到 ES；ES 不可用时静默降级，不影响对话
func (r *ragChatRecorder) save(originalCtx context.Context, ragContext *RagContext, chatErr error) {
	if es.Rag() == nil {
		return
	}
	if chatErr != nil && r.errMessage == "" {
		r.errMessage = chatErr.Error()
	}
	now := time.Now()
	// 客户端断流会 cancel 原 ctx，脱开重新计时，避免记录丢失
	ctx, cancel := context.WithTimeout(trace_util.DetachContext(originalCtx), esTimeout)
	defer cancel()

	if err := es.Rag().IndexDocument(ctx, es.RagChatHistoryIndex(now), r.buildDetail(ctx, ragContext, now.UnixMilli())); err != nil {
		log.Errorf("保存知识问答记录到ES失败，ragId: %s, msgId: %s, error: %v",
			ragContext.Req.GetRagId(), ragContext.MessageId, err)
	}
}

func (r *ragChatRecorder) buildDetail(ctx context.Context, ragContext *RagContext, nowMilli int64) *model.RagConversationDetail {
	req := ragContext.Req
	// 命中安全护栏时以 bff 暂存的内容为准：词表和替换都在 bff，这里累积的是替换前的
	// 模型原文、用户根本没看到过。sensitive 为 nil 表示这一轮没命中
	sensitive := redis.GetRagSensitiveConversation(req.GetConversationId(), ragContext.MessageId)
	errMessage := r.errMessage
	if sensitive != nil {
		errMessage = ""
	}
	return &model.RagConversationDetail{
		Id:               ragContext.MessageId,
		RagId:            req.GetRagId(),
		ConversationId:   req.GetConversationId(),
		Publish:          req.GetPublish(),
		Prompt:           req.GetQuestion(),
		Response:         r.buildResponse(sensitive),
		ReasoningContent: r.buildReasoningContent(sensitive),
		SearchList:       r.buildSearchList(),
		QaSearchList:     r.buildQaSearchList(),
		ErrMessage:       errMessage,
		QaErrMessage:     r.qaErrMessage,
		FileInfo:         buildRecordFileInfo(ctx, req.GetFileInfoList()),
		UserId:           req.GetIdentity().GetUserId(),
		OrgId:            req.GetIdentity().GetOrgId(),
		CreatedAt:        nowMilli,
		UpdatedAt:        nowMilli,
		Statistic:        r.buildStatistic(ragContext, nowMilli),
	}
}

// buildResponse 命中护栏时存用户实际看到的内容（已输出正文 + 护栏回复），而非模型原文
func (r *ragChatRecorder) buildResponse(sensitive *redis.RagSensitiveReply) string {
	if sensitive != nil {
		return sensitive.Response
	}
	return r.response.String()
}

// buildReasoningContent 命中护栏时同样只留用户看到过的那部分思考
func (r *ragChatRecorder) buildReasoningContent(sensitive *redis.RagSensitiveReply) *string {
	reasoning := r.reasoning.String()
	if sensitive != nil {
		reasoning = sensitive.Reasoning
	}
	return phaseValue(r.phase.thinkStart, reasoning)
}

func (r *ragChatRecorder) buildSearchList() *string {
	return phaseValue(r.phase.kbStart, r.searchList)
}

func (r *ragChatRecorder) buildQaSearchList() *string {
	return phaseValue(r.phase.qaStart, r.qaSearchList)
}

func phaseValue(start time.Time, value string) *string {
	if start.IsZero() {
		return nil
	}
	return &value
}

// buildStatistic 只算耗时，token 三个字段等下游 rag 返回 usage 后再填
func (r *ragChatRecorder) buildStatistic(ragContext *RagContext, nowMilli int64) model.RagStatistic {
	startMilli := ragContext.chatStartTime.UnixMilli()
	statistic := model.RagStatistic{
		StartTime:     startMilli,
		TotalCostTime: nowMilli - startMilli,
	}
	// 首token早于基准只可能是打点前就有内容帧，正常不会发生，兜底记0
	if latency := r.firstToken.UnixMilli() - startMilli; !r.firstToken.IsZero() && latency > 0 {
		statistic.FirstTokenLatency = latency
	}
	thinkEnd := r.phase.thinkEnd
	if r.phase.thinkStart.IsZero() {
		thinkEnd = time.Time{}
	}
	statistic.QaSearchListTimeCost = buildTimeCost(r.phase.qaStart, r.phase.qaEnd, nowMilli)
	statistic.SearchListTimeCost = buildTimeCost(r.phase.kbStart, r.phase.kbEnd, nowMilli)
	statistic.ReasoningTimeCost = buildTimeCost(r.phase.thinkStart, thinkEnd, nowMilli)
	return statistic
}

// buildTimeCost 把一段起止时刻转成落库结构。起点为零表示该阶段未发生，返回零值
func buildTimeCost(start, end time.Time, nowMilli int64) model.RagTimeCost {
	if start.IsZero() {
		return model.RagTimeCost{}
	}
	startMilli := start.UnixMilli()
	endMilli := nowMilli
	if !end.IsZero() {
		endMilli = end.UnixMilli()
	}
	if endMilli < startMilli {
		return model.RagTimeCost{}
	}
	return model.RagTimeCost{Start: startMilli, End: endMilli}
}

// buildRecordFileInfo 问答请求里的附件是临时存储（到期会被清），落历史前转永久存储。
// 转存失败就退回原地址，历史仍然记下来
func buildRecordFileInfo(ctx context.Context, fileInfoList []*rag_service.FileInfo) (retList []*model.RagFileInfo) {
	retList = make([]*model.RagFileInfo, 0, len(fileInfoList))
	defer util.PrintPanicStack()
	for _, fileInfo := range fileInfoList {
		if fileInfo == nil {
			continue
		}
		retList = append(retList, &model.RagFileInfo{
			FileName: fileInfo.FileName,
			FileSize: fileInfo.FileSize,
			FileUrl:  restoreFile(ctx, fileInfo.FileUrl),
		})
	}
	return retList
}

// restoreFile 把临时文件拷到永久目录，失败返回原地址
func restoreFile(ctx context.Context, fileUrl string) string {
	if fileUrl == "" || !minio_service.Enabled() {
		return fileUrl
	}
	newFileUrl, err := minio_service.CopyFile(ctx, fileUrl, minio_client.PermanentDir(), true)
	if err != nil || newFileUrl == "" {
		log.Errorf("知识问答附件转永久存储失败，fileUrl: %s, err: %v", fileUrl, err)
		return fileUrl
	}
	return newFileUrl
}
