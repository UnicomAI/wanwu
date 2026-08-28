package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	model_service "github.com/UnicomAI/wanwu/api/proto/model-service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	mp "github.com/UnicomAI/wanwu/pkg/model-provider"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	// additionalPrompt 始终追加到推荐系统提示词之后，约束输出格式：
	// 第一行有且仅有类型标识 ANSWER/REJECT，从第二行起输出正文，问题以换行分隔。
	additionalPrompt = "\n输出格式（必须严格遵守，第一行有且仅有类型标识）：\n- 正常推荐：第一行输出 ANSWER，从第二行起每行一个问题（共3行），问题结尾不加问号、不加序号、不输出思考过程；\n- 拒绝推荐：第一行输出 REJECT，第二行输出一句拒绝说明。\n示例1：\nANSWER\n这种植物需要每天浇水吗\n它的生长期一般是多久\n室内养植需要注意阳光吗\n示例2：\nREJECT\n当前对话涉及暴力伤害类内容，无法推荐相关问题"
	// systemPrompt 是默认推荐系统提示词（用户未配置时使用）。核心是把模型限定为"追问推荐生成器"，
	// 并明确禁止它回答对话里的问题——否则通用问答模型会直接把用户最后的问题又答一遍。
	systemPrompt = "你是一个\"追问推荐\"生成器，任务是根据给定的对话，预测用户接下来最可能继续提出的问题。特别注意：你不是在与用户对话，绝对不要回答或回应对话中的任何问题，只输出推荐的后续问题。要求：1. 给出3个问题，彼此有区分度、与最近一轮话题紧密相关、可适当延伸；2. 不要重复对话中已问过的问题；3. 每个问题单独一行。安全：若对话涉及政治敏感、违法违规、暴力伤害、违反公序良俗类内容，则拒绝推荐。"
)

type RecommendLLMResp struct {
	ID                string                    `json:"id"`                               // 唯一标识
	Object            string                    `json:"object"`                           // 固定为 "chat.completion"
	Created           int                       `json:"created"`                          // 时间戳（秒）
	Model             string                    `json:"model" validate:"required"`        // 使用的模型
	Choices           []RecommendRespChoice     `json:"choices" validate:"required,dive"` // 生成结果列表
	Usage             mp_common.OpenAIRespUsage `json:"usage"`                            // token 使用统计
	ServiceTier       *string                   `json:"service_tier"`                     // （火山）指定是否使用TPM保障包。生效对象为购买了保障包推理接入点
	SystemFingerprint *string                   `json:"system_fingerprint"`
	Code              *int                      `json:"code,omitempty"`
	ImgId             *string                   `json:"img_id,omitempty"` // 视觉模型返回图片id
}

type RecommendRespChoice struct {
	Index        int                  `json:"index"`             // 选项索引
	Message      *mp_common.OpenAIMsg `json:"message,omitempty"` // 非流式生成的消息
	Delta        *mp_common.OpenAIMsg `json:"delta,omitempty"`   // 流式生成的消息
	FinishReason string               `json:"finish_reason"`     // 停止原因
	Logprobs     interface{}          `json:"logprobs"`
	ContentType  string               `json:"contentType"` // "answer": 正常推荐 "tips": 拒绝推荐
}

func AgentRecommendChatCompletions(ctx *gin.Context, modelID string, req *mp_common.LLMReq) {
	detachedCtx := trace_util.DetachContext(ctx.Request.Context())
	// 推荐属于锦上添花能力，流开始前的任何失败都不打扰用户（前端见 4xx 走 FatalError：仅置 loading=false，
	// 不弹窗、不渲染、不重试）。失败原因写入 ERROR 日志，并随 4xx JSON body 返回，前端读取 body 后可在 F12 排查。
	iLLM, llmReq, modelInfo, startTime, err := setupRecommendLLMReq(ctx, modelID, req)
	if err != nil {
		recommendFail(ctx, "%v", err)
		return
	}
	requestBody := MarshalStatisticBody(req)
	_, sseCh, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReq)
	if err != nil {
		go func() {
			defer util.PrintPanicStack()
			recordModelStatisticV2Failure(detachedCtx, modelInfo, false, requestBody, err)
		}()
		recommendFail(ctx, "model %v chat completions err: %v", modelInfo.ModelId, err)
		return
	}
	streamRecommend(ctx, detachedCtx, modelInfo, sseCh, startTime, requestBody)
}

// setupRecommendLLMReq 完成推荐模型的获取、校验、配置与请求构造（GetModel→ToModelConfig→关 thinking→NewReq），
// 返回可直接 ChatCompletions 的 iLLM、llmReq、modelInfo 与起始时间。
// 供流式(AgentRecommendChatCompletions)与收集(collectRecommendLLMAnswer)复用，调用方各自处理 ChatCompletions 之后的差异。
func setupRecommendLLMReq(ctx *gin.Context, modelID string, req *mp_common.LLMReq) (mp.ILLM, mp_common.ILLMReq, *model_service.ModelInfo, time.Time, error) {
	modelInfo, err := model.GetModel(ctx.Request.Context(), &model_service.GetModelReq{ModelId: modelID})
	if err != nil {
		return nil, nil, nil, time.Time{}, fmt.Errorf("model %v get model err: %v", modelID, err)
	}
	if !modelInfo.IsActive {
		return nil, nil, nil, time.Time{}, fmt.Errorf("model %v inactive", modelInfo.ModelId)
	}
	if req != nil && req.Model != modelInfo.Model {
		return nil, nil, nil, time.Time{}, fmt.Errorf("model %v chat completions err: model mismatch (req=%v)", modelInfo.ModelId, req.Model)
	}
	// llm config
	llm, err := mp.ToModelConfig(modelInfo.Provider, modelInfo.ModelType, modelInfo.ProviderConfig)
	if err != nil {
		return nil, nil, nil, time.Time{}, fmt.Errorf("model %v to model config err: %v", modelInfo.ModelId, err)
	}
	// 推荐不需要模型输出思考过程，模型支持深度思考时强制关闭 thinking
	if req != nil && modelSupportsThinking(llm) {
		enableThinking := false
		req.EnableThinking = &enableThinking
	}
	iLLM, ok := llm.(mp.ILLM)
	if !ok {
		return nil, nil, nil, time.Time{}, fmt.Errorf("model %v chat completions err: invalid provider", modelInfo.ModelId)
	}
	startTime := time.Now()
	llmReq, err := iLLM.NewReq(req)
	if err != nil {
		return nil, nil, nil, time.Time{}, fmt.Errorf("model %v chat completions NewReq err: %v", modelInfo.ModelId, err)
	}
	return iLLM, llmReq, modelInfo, startTime, nil
}

// recommendFail 处理推荐流开始前的失败：记录 ERROR 日志，并返回 4xx JSON 错误。
// 前端 onopen 见 4xx 会走 FatalError 分支（仅置 loading=false，不弹窗、不渲染、不重试），对用户仍静默。
func recommendFail(ctx *gin.Context, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log.Errorf("[Recommend] %s", msg)
	gin_util.Response(ctx, nil, grpc_util.ErrorStatus(err_code.Code_BFFGeneral, msg))
}

// streamRecommend 消费 LLM 的 SSE 流，按推荐协议剥离标识后下发，并记录统计。
// detachedCtx 用于请求结束后异步上报统计，避免被 gin 请求上下文取消影响。
func streamRecommend(ctx *gin.Context, detachedCtx context.Context, modelInfo *model_service.ModelInfo, sseCh <-chan mp_common.ILLMResp, startTime time.Time, requestBody string) {
	// stream
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Content-Type", "text/event-stream; charset=utf-8")
	var (
		answer            string
		firstTokenTime    time.Time
		firstTokenLatency int
		promptTokens      int
		completionTokens  int
		totalTokens       int
		finishReason      string
	)
	p := &recommendStreamProcessor{}
	for sseResp := range sseCh {
		data, ok := sseResp.ConvertResp()
		dataStr := ""
		write := true
		if ok && data != nil {
			// token 用量与首字延迟只取决于模型输出，与是否下发前端无关：
			// 即便整段被护栏 suppressed，模型也已消耗 token，必须如实统计。
			if firstTokenTime.IsZero() {
				firstTokenTime = time.Now()
				firstTokenLatency = int(time.Since(startTime).Milliseconds())
			}
			promptTokens = data.Usage.PromptTokens
			completionTokens = data.Usage.CompletionTokens
			totalTokens = data.Usage.TotalTokens
			if len(data.Choices) > 0 && data.Choices[0].FinishReason != "" {
				finishReason = data.Choices[0].FinishReason
			}

			if len(data.Choices) > 0 && data.Choices[0].Delta != nil {
				answer = answer + data.Choices[0].Delta.Content
				var payload string
				if payload, write = p.process(data); write {
					dataStr = fmt.Sprintf("data: %v\n", payload)
				}
			}
		} else {
			// 流式过程中，大模型sse返回的这一行是空行，即sseResp.String()==""；前端正常展示，也需要这个空行
			dataStr = fmt.Sprintf("%v\n", sseResp.String())
		}
		if !write {
			continue // 仅跳过下发，统计已在上方采集
		}
		if _, err := ctx.Writer.Write([]byte(dataStr)); err != nil {
			log.Errorf("model %v chat completions sse err: %v", modelInfo.ModelId, err)
		}
		ctx.Writer.Flush()
	}
	ctx.Set(gin_util.STATUS, http.StatusOK)
	ctx.Set(gin_util.RESULT, answer)
	if p.suppressed {
		log.Warnf("[Recommend] model %v output missing ANSWER/REJECT tag, suppressed (not shown to frontend)", modelInfo.ModelId)
	}
	go func() {
		defer util.PrintPanicStack()
		recordModelStatisticV2(detachedCtx, modelInfo,
			promptTokens, completionTokens, totalTokens, 0, firstTokenLatency, true,
			http.StatusOK, requestBody, answer, finishReason, "")
	}()
}

// modelSupportsThinking 通过模型 provider 配置判断是否支持深度思考（thinkingSupport=support）。
func modelSupportsThinking(llm interface{}) bool {
	jsonBytes, err := json.Marshal(llm)
	if err != nil {
		return false
	}
	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return false
	}
	ts, _ := result["thinkingSupport"].(string)
	return ts == "support"
}

const (
	recommendAnswerTag = "ANSWER" // 正常推荐类型标识（独占首行）
	recommendRejectTag = "REJECT" // 拒绝推荐类型标识（独占首行）
)

// recommendStreamProcessor 解析推荐模型的流式输出。
// 约定：模型先输出可选的 <think>…</think> 思考段，随后第一行只放类型标识
// （ANSWER/REJECT），从第二行起为正文（问题以换行分隔）。处理分三个阶段：
// 剥离 think → 解析首行判定类型（只判一次） → 正文逐帧透传。
// 护栏：若模型未按约定输出 ANSWER/REJECT 标识，则判定整段输出不可信（suppressed），
// 全程静默不下发——避免把模型的对话回复/思考过程误当作推荐问题展示给前端。
type recommendStreamProcessor struct {
	typeDone   bool   // 已完成类型判定，进入正文透传
	errorFlag  bool   // 命中拒绝推荐（contentType=tips）
	suppressed bool   // 未识别到类型标识，整段不下发
	headBuf    string // 类型判定前的累计缓冲（含可能的 think 与首行）
	bodyBuf    string // 正文阶段的累计缓冲，凑齐整行后清洗下发
}

// process 处理单个 SSE 分片，返回要下发的 data 负载以及是否需要下发。
func (p *recommendStreamProcessor) process(data *mp_common.LLMResp) (string, bool) {
	if len(data.Choices) == 0 || data.Choices[0].Delta == nil {
		return "", false
	}
	delta := data.Choices[0].Delta
	// 深度思考内容直接丢弃（reasoning_content 通道）
	if delta.ReasoningContent != nil && *delta.ReasoningContent != "" {
		return "", false
	}
	finished := data.Choices[0].FinishReason != ""

	// 阶段一/二：剥 think + 判类型，未判定完成前一律不下发
	if !p.typeDone {
		body, done := p.consumeHead(delta.Content, finished)
		if !done {
			return "", false
		}
		p.typeDone = true
		delta.Content = body
	}

	// 护栏：未识别到 ANSWER/REJECT 标识，整段视为不可信，静默不下发
	if p.suppressed {
		return "", false
	}

	// 拒绝推荐结束时，将 stop 映射为 accidentStop，前端据此区分
	if p.errorFlag && data.Choices[0].FinishReason == "stop" {
		data.Choices[0].FinishReason = "accidentStop"
	}

	// 正文按行缓冲清洗：去掉模型可能擅自添加的行首序号/项目符号与行尾问号。
	delta.Content = p.consumeBody(delta.Content, finished)
	// 本帧没有凑齐完整行就不下发；但结束帧即使正文为空也要发（携带 finish_reason/usage）
	if delta.Content == "" && !finished {
		return "", false
	}
	dataByte, _ := json.Marshal(buildRecommendResp(p.errorFlag, data))
	return string(dataByte), true
}

// recommendLineMarkerRe 匹配行首的序号/项目符号：如 "1." "1、" "2)" "3：" "- " "• "。
// 要求数字后必须跟分隔符，避免误伤以数字开头的正常问句（如 "2024年发生了什么"）。
var recommendLineMarkerRe = regexp.MustCompile(`^\s*(?:\d+\s*[.、)）:：]|[-*•·])\s*`)

// cleanRecommendLine 清洗单行推荐问题：去行首序号/项目符号、去行尾问号、去首尾空白。
func cleanRecommendLine(line string) string {
	line = strings.TrimSpace(line)
	line = recommendLineMarkerRe.ReplaceAllString(line, "")
	line = strings.TrimRight(line, "?？")
	return strings.TrimSpace(line)
}

// consumeBody 在正文阶段缓冲内容，凑齐完整行（以 \n 分隔）后逐行清洗再下发，
// 行间仍用 \n 连接以兼容前端按 \n 切分。返回本次可下发的清洗后内容，无完整行且未结束时返回 ""。
func (p *recommendStreamProcessor) consumeBody(content string, finished bool) string {
	p.bodyBuf += content
	var out strings.Builder
	for {
		idx := strings.IndexByte(p.bodyBuf, '\n')
		if idx < 0 {
			break
		}
		line := p.bodyBuf[:idx]
		p.bodyBuf = p.bodyBuf[idx+1:]
		if cleaned := cleanRecommendLine(line); cleaned != "" {
			out.WriteString(cleaned)
			out.WriteString("\n")
		}
	}
	if finished {
		if cleaned := cleanRecommendLine(p.bodyBuf); cleaned != "" {
			out.WriteString(cleaned)
		}
		p.bodyBuf = ""
	}
	return out.String()
}

// consumeHead 在类型判定前累计内容，剥离可选的 <think>…</think>，再以首行解析类型标识。
// 返回 (类型行之后的正文, 是否已完成判定)；未完成时调用方应跳过该帧。
// 护栏：模型未按约定输出 ANSWER/REJECT 类型行时，置 suppressed 标记整段不下发。
func (p *recommendStreamProcessor) consumeHead(content string, finished bool) (string, bool) {
	p.headBuf += content
	// 仅去掉前置空白，便于识别 <think> 与类型标识；正文内部及结尾的换行必须保留，
	// 否则会丢掉问题之间的分隔符。
	buf := strings.TrimLeft(p.headBuf, " \t\r\n")

	// 剥离前置 <think>…</think>
	if strings.HasPrefix(buf, "<think>") {
		end := strings.Index(buf, "</think>")
		if end < 0 {
			if finished { // 思考段未闭合就结束，未拿到任何类型标识，整段不下发
				p.headBuf = ""
				p.suppressed = true
				return "", true
			}
			p.headBuf = buf
			return "", false // 继续等待 </think>
		}
		buf = strings.TrimLeft(buf[end+len("</think>"):], " \t\r\n")
	}
	p.headBuf = buf

	idx := strings.IndexByte(buf, '\n')
	if idx < 0 {
		// 尚无换行：可能类型标识还没输完，继续缓冲
		probe := strings.TrimSpace(buf)
		if !finished && isRecommendTagPrefix(probe) {
			return "", false
		}
		// 流已结束或确定不是类型标识前缀
		p.headBuf = ""
		if strings.EqualFold(probe, recommendRejectTag) {
			p.errorFlag = true
			return "", true
		}
		if strings.EqualFold(probe, recommendAnswerTag) {
			return "", true
		}
		p.suppressed = true // 无类型标识，整段不下发
		return "", true
	}

	firstLine := strings.TrimSpace(buf[:idx])
	rest := buf[idx+1:]
	p.headBuf = ""
	switch {
	case strings.EqualFold(firstLine, recommendRejectTag):
		p.errorFlag = true
		return rest, true
	case strings.EqualFold(firstLine, recommendAnswerTag):
		return rest, true
	default:
		// 首行不是 ANSWER/REJECT：模型未遵循格式，整段不下发
		p.suppressed = true
		return "", true
	}
}

// isRecommendTagPrefix 判断 s（忽略大小写）是否可能是 ANSWER/REJECT 标识的前缀。
// 空串视为前缀，以便首帧继续缓冲。
func isRecommendTagPrefix(s string) bool {
	if s == "" {
		return true
	}
	up := strings.ToUpper(s)
	return strings.HasPrefix(recommendAnswerTag, up) || strings.HasPrefix(recommendRejectTag, up)
}

func buildRecommendResp(errorFlag bool, data *mp_common.LLMResp) *RecommendLLMResp {
	// contentType 区分正常推荐与拒绝推荐，前端据此渲染（answer 可点击 / tips 提示样式）
	contentType := "answer"
	if errorFlag {
		contentType = "tips"
	}
	return &RecommendLLMResp{
		ID:      data.ID,
		Object:  data.Object,
		Created: data.Created,
		Model:   data.Model,
		Choices: []RecommendRespChoice{{
			Index:        data.Choices[0].Index,
			Message:      data.Choices[0].Message,
			Delta:        data.Choices[0].Delta,
			FinishReason: data.Choices[0].FinishReason,
			Logprobs:     data.Choices[0].Logprobs,
			ContentType:  contentType,
		}},
		Usage:             data.Usage,
		ServiceTier:       data.ServiceTier,
		SystemFingerprint: data.SystemFingerprint,
		Code:              data.Code,
		ImgId:             data.ImgId,
	}
}

// collectRecommendLLMAnswer 调用推荐模型并把输出收集成完整文本返回（不向客户端流式写出），
// 供"对话接口内嵌追问"复用。与 AgentRecommendChatCompletions 的区别是只收集、不 writer.Write。
func collectRecommendLLMAnswer(ctx *gin.Context, modelID string, req *mp_common.LLMReq) (string, error) {
	iLLM, llmReq, modelInfo, _, err := setupRecommendLLMReq(ctx, modelID, req)
	if err != nil {
		return "", err
	}
	resp, sseCh, err := iLLM.ChatCompletions(ctx.Request.Context(), llmReq)
	if err != nil {
		return "", err
	}
	// 非流式：直接取完整消息
	if !llmReq.Stream() {
		if data, ok := resp.ConvertResp(); ok && len(data.Choices) > 0 && data.Choices[0].Message != nil {
			return data.Choices[0].Message.Content, nil
		}
		return "", grpc_util.ErrorStatus(err_code.Code_BFFGeneral, fmt.Sprintf("model %v recommend err: invalid resp", modelInfo.ModelId))
	}
	// 流式：累加 delta
	var answer string
	for sseResp := range sseCh {
		if data, ok := sseResp.ConvertResp(); ok && data != nil && len(data.Choices) > 0 && data.Choices[0].Delta != nil {
			answer += data.Choices[0].Delta.Content
		}
	}
	return answer, nil
}

// parseRecommendQuestions 解析推荐模型输出（非流式整段），与流式 recommendStreamProcessor 同协议：
// 剥离前置 <think>…</think> → 首行须为类型标识 ANSWER（REJECT/无标识均返回空）→ 其余行为问题，
// 逐行用 cleanRecommendLine 清洗（去行首序号/项目符号、行尾问号）。
func parseRecommendQuestions(answer string) []string {
	if i := strings.LastIndex(answer, "</think>"); i >= 0 {
		answer = answer[i+len("</think>"):]
	}
	answer = strings.TrimLeft(answer, " \t\r\n")
	idx := strings.IndexByte(answer, '\n')
	if idx < 0 {
		return nil // 只有类型行、没有正文
	}
	if !strings.EqualFold(strings.TrimSpace(answer[:idx]), recommendAnswerTag) {
		return nil // REJECT 或未按格式输出，不返回追问
	}
	var questions []string
	for _, line := range strings.Split(answer[idx+1:], "\n") {
		if q := cleanRecommendLine(line); q != "" {
			questions = append(questions, q)
		}
	}
	return questions
}
