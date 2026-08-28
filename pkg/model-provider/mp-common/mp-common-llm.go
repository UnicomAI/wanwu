package mp_common

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/go-resty/resty/v2"
)

type MsgRole string

const (
	MsgRoleSystem    MsgRole = "system"
	MsgRoleUser      MsgRole = "user"
	MsgRoleAssistant MsgRole = "assistant"
	MsgRoleFunction  MsgRole = "tool"
)

const (
	TagChat                string = "文本生成"
	TagTextEmbedding       string = "文本向量化"
	TagMultiModalEmbedding string = "多模态向量化"
	TagTextRerank          string = "文本重排序"
	TagMultiModalRerank    string = "多模态重排序"
	TagGui                 string = "GUI"
	TagOcr                 string = "OCR"
	TagPdfParser           string = "PDF文档解析"
	TagSyncAsr             string = "短语音识别"
	TagText2Image          string = "文生图"
	TagVisionSupport       string = "图文问答"
	TagToolCall            string = "工具调用"
	TagThinkingSupport     string = "深度思考开关"
	TagScopeTypePrivate    string = "个人"
	TagScopeTypePublic     string = "全局公开"
	TagScopeTypeOrg        string = "组织公开"
	TagSourceTypeLocal     string = "本地"
)

type Tag struct {
	Text string `json:"text"`
}

func GetTagsByFunctionCall(fcType string) []Tag {
	var tags []Tag
	if FCType(fcType) == FCTypeToolCall {
		tags = append(tags, Tag{
			Text: TagToolCall,
		})
	}
	return tags
}

func GetTagsByVisionSupport(visionType string) []Tag {
	var tags []Tag
	if VSType(visionType) == VSTypeSupport {
		tags = append(tags, Tag{
			Text: TagVisionSupport,
		})
	}
	return tags
}

func GetTagsByContentSize(size *int) []Tag {
	var tags []Tag
	if size != nil && *size > 0 {
		kValue := *size / 1024
		// 格式化为"XK"字符串并添加到tags列表
		tags = append(tags, Tag{
			Text: fmt.Sprintf("%dK", kValue),
		})
	}
	return tags
}

type ToolType string

const (
	ToolTypeFunction ToolType = "function"
)

type FCType string

const (
	FCTypeFunctionCall FCType = "functionCall"
	FCTypeNoSupport    FCType = "noSupport"
	FCTypeToolCall     FCType = "toolCall"
)

type VSType string

const (
	VSTypeSupport   VSType = "support"
	VSTypeNoSupport VSType = "noSupport"
)

type ThinkingType string

const (
	ThinkingTypeSupport   ThinkingType = "support"
	ThinkingTypeNoSupport ThinkingType = "noSupport"
)

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// --- openapi request ---

type LLMReq struct {
	// general
	Model          string                `json:"model" validate:"required"`
	Messages       []OpenAIReqMsg        `json:"messages" validate:"required"`
	Stream         *bool                 `json:"stream,omitempty"`
	MaxTokens      *int                  `json:"max_tokens,omitempty"`
	Stop           *string               `json:"stop,omitempty"`
	ResponseFormat *OpenAIResponseFormat `json:"response_format,omitempty"`
	Temperature    *float64              `json:"temperature,omitempty"`
	Tools          []OpenAITool          `json:"tools,omitempty"`

	// custom
	Thinking            *Thinking              `json:"thinking,omitempty"`              // 控制模型是否开启深度思考模式。
	EnableThinking      *bool                  `json:"enable_thinking,omitempty"`       // 控制模型是否开启深度思考模式。
	ChatTemplateKwargs  map[string]interface{} `json:"chat_template_kwargs,omitempty"`  // 控制模型是否开启深度思考模式（元景）
	ReasoningEffort     *string                `json:"reasoning_effort,omitempty"`      // 控制模型的推理强度（Ds）Possible values: [high, max]
	MaxCompletionTokens *int                   `json:"max_completion_tokens,omitempty"` // 控制模型输出的最大长度[0,64k]
	LogitBias           map[string]int         `json:"logit_bias,omitempty"`            // 调整指定 token 在模型输出内容中出现的概率
	ToolChoice          interface{}            `json:"tool_choice,omitempty"`           // 强制指定工具调用的策略
	TopP                *float64               `json:"top_p,omitempty"`
	TopK                *int                   `json:"top_k,omitempty"`
	MinP                *float64               `json:"min_p,omitempty"`
	ParallelToolCalls   *bool                  `json:"parallel_tool_calls,omitempty"` // 是否开启并行工具调用
	StreamOptions       *StreamOptions         `json:"stream_options,omitempty"`      // 当启用流式输出时，可通过将本参数设置为{"include_usage": true}，在输出的最后一行显示所使用的Token数。

	PresencePenalty   *float64 `json:"presence_penalty,omitempty"`   // 控制模型生成文本时的内容重复度
	FrequencyPenalty  *float64 `json:"frequency_penalty,omitempty"`  // 频率惩罚系数
	RepetitionPenalty *float64 `json:"repetition_penalty,omitempty"` // 模型生成时连续序列中的重复度

	Seed           *int  `json:"seed,omitempty"`         // 种子
	Logprobs       *bool `json:"logprobs,omitempty"`     // 是否返回输出 Token 的对数概率
	TopLogprobs    *int  `json:"top_logprobs,omitempty"` // 指定在每一步生成时，返回模型最大概率的候选 Token 个数
	N              *int  `json:"n,omitempty"`
	ThinkingBudget *int  `json:"thinking_budget,omitempty"` // 思考过程的最大长度，只在enable_thinking为true时生效

	WebSearch *WebSearch `json:"web_search,omitempty"` //搜索增强
	User      *string    `json:"user,omitempty"`       // 用户标识（兼容千帆)
	// Yuanjing
	DoSample  *bool      `json:"do_sample,omitempty"`
	ExtraBody *ExtraBody `json:"extra_body,omitempty"` // 扩展参数
}

type OpenAIReqMsg struct {
	Role             MsgRole       `json:"role"` // "system" | "user" | "assistant" | "function(已弃用)"
	Content          interface{}   `json:"content"`
	ToolCallId       *string       `json:"tool_call_id,omitempty"`
	ReasoningContent *string       `json:"reasoning_content,omitempty"`
	Reasoning        *string       `json:"reasoning,omitempty"`
	Name             *string       `json:"name,omitempty"`
	FunctionCall     *FunctionCall `json:"function_call,omitempty"`
	ToolCalls        []*ToolCall   `json:"tool_calls,omitempty"`
}

type ExtraBody struct {
	ApiOption string `json:"api_option"` // 选择指定功能。 1）math：拍照答题；2）ocr：多模态OCR；3）general：通用场景。   默认会根据prompt进行意图判断
}

func (req *LLMReq) Data() (map[string]interface{}, error) {
	m := make(map[string]interface{})
	b, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

type StreamOptions struct {
	IncludeUsage      *bool `json:"include_usage,omitempty"`
	ChunkIncludeUsage *bool `json:"chunk_include_usage,omitempty"`
}

type WebSearch struct {
	Enable         *bool `json:"enable,omitempty"`
	EnableCitation *bool `json:"enable_citation,omitempty"`
	EnableTrace    *bool `json:"enable_trace,omitempty"`
	EnableStatus   *bool `json:"enable_status,omitempty"`
}

type OpenAIMsg struct {
	Role             MsgRole       `json:"role"` // "system" | "user" | "assistant" | "function(已弃用)"
	Content          string        `json:"content"`
	ToolCallId       *string       `json:"tool_call_id,omitempty"`
	ReasoningContent *string       `json:"reasoning_content,omitempty"`
	Reasoning        *string       `json:"reasoning,omitempty"`
	Name             *string       `json:"name,omitempty"`
	FunctionCall     *FunctionCall `json:"function_call,omitempty"`
	ToolCalls        []*ToolCall   `json:"tool_calls,omitempty"`
}

type Thinking struct {
	Type string `json:"type" default:"enabled"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     ToolType     `json:"type"`
	Function FunctionCall `json:"function"`
	Index    *int         `json:"index,omitempty"`
}

type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type OpenAIResponseFormat struct {
	Type string `json:"type"` // "text" | "json"
}

type OpenAITool struct {
	Type     ToolType        `json:"type" validate:"required"`
	Function *OpenAIFunction `json:"function" validate:"required"`
}

type OpenAIFunction struct {
	Name        string      `json:"name" validate:"required"`
	Description string      `json:"description,omitempty"`
	Parameters  interface{} `json:"parameters,omitempty"`
}

func (req *LLMReq) Check() error { return nil }

// --- openapi response ---

type LLMResp struct {
	ID                string             `json:"id"`                               // 唯一标识
	Object            string             `json:"object"`                           // 固定为 "chat.completion"
	Created           int                `json:"created"`                          // 时间戳（秒）
	Model             string             `json:"model" validate:"required"`        // 使用的模型
	Choices           []OpenAIRespChoice `json:"choices" validate:"required,dive"` // 生成结果列表
	Usage             OpenAIRespUsage    `json:"usage"`                            // token 使用统计
	ServiceTier       *string            `json:"service_tier"`                     // （火山）指定是否使用TPM保障包。生效对象为购买了保障包推理接入点
	SystemFingerprint *string            `json:"system_fingerprint"`
	Code              *int               `json:"code,omitempty"`
	ImgId             *string            `json:"img_id,omitempty"` // 视觉模型返回图片id
}

// OpenAIRespUsage 结构体表示 token 消耗
type OpenAIRespUsage struct {
	CompletionTokens int `json:"completion_tokens"` // 输出 token 数
	PromptTokens     int `json:"prompt_tokens"`     // 输入 token 数
	TotalTokens      int `json:"total_tokens"`      // 总 token 数
}

// OpenAIRespChoice 结构体表示单个生成选项
type OpenAIRespChoice struct {
	Index        int         `json:"index"`             // 选项索引
	Message      *OpenAIMsg  `json:"message,omitempty"` // 非流式生成的消息
	Delta        *OpenAIMsg  `json:"delta,omitempty"`   // 流式生成的消息
	FinishReason string      `json:"finish_reason"`     // 停止原因
	Logprobs     interface{} `json:"logprobs"`
}

type OpenAIRespChoiceMsg struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

// --- request ---

type ILLMReq interface {
	Stream() bool
	Data() map[string]interface{}
	OpenAIReq() (*LLMReq, bool)
}

// llmReq implementation of ILLMReq
type llmReq struct {
	data map[string]interface{}
}

func NewLLMReq(data map[string]interface{}) ILLMReq {
	return &llmReq{data: data}
}

func (req *llmReq) Data() map[string]interface{} {
	return req.data
}

func (req *llmReq) Stream() bool {
	if req.data == nil {
		return false
	}
	v, ok := req.data["stream"]
	if !ok {
		return false
	}
	stream, _ := v.(bool)
	return stream
}

func (req *llmReq) OpenAIReq() (*LLMReq, bool) {
	if req == nil {
		return nil, false
	}
	b, err := json.Marshal(req.data)
	if err != nil {
		log.Errorf("LLMReq to LLMReq marshal err: %v", err)
		return nil, false
	}
	ret := &LLMReq{}
	if err = json.Unmarshal(b, ret); err != nil {
		log.Errorf("LLMReq to LLMReq unmarshal err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- response ---

type ILLMResp interface {
	String() string
	Raw() string
	ConvertResp() (*LLMResp, bool)
}

// llmResp implementation of ILLMResp
type llmResp struct {
	stream     bool
	raw        string   // 原始 数据
	resp       *LLMResp // 缓存 unmarshal 结果
	respStr    string   // 缓存 marshal 结果
	inThinking bool     // 流式思维链状态
	prefixSep  string   // 流式数据前缀分隔符，保持原始格式
}

func NewLLMResp(stream bool, raw string) ILLMResp {
	return &llmResp{stream: stream, raw: raw}
}

func (resp *llmResp) Raw() string {
	return resp.raw
}

func (resp *llmResp) String() string {
	if resp.respStr != "" {
		return resp.respStr
	}
	return resp.raw
}

func (resp *llmResp) ConvertResp() (*LLMResp, bool) {
	if resp.stream {
		if resp.raw == "data: [DONE]" || !strings.HasPrefix(resp.raw, "data:") {
			return nil, false
		}
	}

	if resp.resp != nil {
		return resp.resp, true
	}

	raw := resp.raw
	if resp.stream {
		// 提取原始前缀分隔符格式，保持 data: 或 data: 的一致性
		afterPrefix := strings.TrimPrefix(resp.raw, "data:")
		if len(afterPrefix) > 0 && afterPrefix[0] == ' ' {
			resp.prefixSep = "data: "
			raw = afterPrefix[1:] // 去掉空格
		} else {
			resp.prefixSep = "data:"
			raw = afterPrefix
		}
	}

	ret := &LLMResp{}
	if err := json.Unmarshal([]byte(raw), ret); err != nil {
		log.Errorf("llm resp (%v) convert to openai resp err: %v", raw, err)
		return nil, false
	}

	if err := util.Validate(ret); err != nil {
		log.Errorf("llm resp validate err: %v", err)
		return nil, false
	}

	if resp.stream {
		if len(ret.Choices) > 0 && ret.Choices[0].Delta != nil {
			delta := ret.Choices[0].Delta
			if delta.Role == "" {
				delta.Role = MsgRoleAssistant
			}
			resp.inThinking, _ = extractThinkingFromDelta(delta, resp.inThinking)
		}
	} else {
		extractThinkingFromResp(ret)
	}

	if newData, err := json.Marshal(ret); err == nil {
		prefix := ""
		if resp.stream {
			prefix = resp.prefixSep
		}
		resp.respStr = prefix + string(newData) + "\n"
	}

	resp.resp = ret
	return ret, true
}

// --- ChatCompletions ---

func ChatCompletions(ctx context.Context, provider, apiKey, url string, req ILLMReq, respConverter func(bool, string) ILLMResp, headers ...Header) (ILLMResp, <-chan ILLMResp, error) {
	if req.Stream() {
		ret, err := chatCompletionsStream(ctx, provider, apiKey, url, req, respConverter, headers...)
		return nil, ret, err
	}
	ret, err := chatCompletionsUnary(ctx, provider, apiKey, url, req, respConverter, headers...)
	return ret, nil, err
}

func chatCompletionsUnary(ctx context.Context, provider, apiKey, url string, req ILLMReq, respConverter func(bool, string) ILLMResp, headers ...Header) (ILLMResp, error) {
	if req.Stream() {
		return nil, fmt.Errorf("request %v %v chat completions unary but stream", url, provider)
	}

	if apiKey != "" {
		headers = append(headers, Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}

	request := trace_util.NewResty(ctx).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // 关闭证书校验
		SetTimeout(0).                                             // 关闭请求超时
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetBody(req.Data()).
		SetDoNotParseResponse(true)
	for _, header := range headers {
		request.SetHeader(header.Key, header.Value)
	}
	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("request %v %v chat completions unary err: %v", url, provider, err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v chat completions unary read response body err: %v", url, provider, err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v chat completions unary http status %v msg: %v", url, provider, resp.StatusCode(), string(b))
	}
	respData := respConverter(false, string(b))
	respData.ConvertResp()
	return respData, nil
}

func chatCompletionsStream(ctx context.Context, provider, apiKey, url string, req ILLMReq, respConverter func(bool, string) ILLMResp, headers ...Header) (<-chan ILLMResp, error) {
	if !req.Stream() {
		return nil, fmt.Errorf("request %v %v chat completions stream but unary", url, provider)
	}

	if apiKey != "" {
		headers = append(headers, Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}

	ret := make(chan ILLMResp, 1024)
	// 创建错误通道（缓冲1个，防止goroutine阻塞），主函数接收异步错误
	errChan := make(chan error, 1)

	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		var resp *resty.Response
		var err error

		request := trace_util.NewResty(ctx).
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // 关闭证书校验
			R().
			SetContext(ctx).
			SetHeader("Content-Type", "application/json").
			SetHeader("Accept", "application/json").
			SetBody(req.Data()).
			SetDoNotParseResponse(true)
		for _, header := range headers {
			request.SetHeader(header.Key, header.Value)
		}
		resp, err = request.Post(url)
		if err != nil {
			wrappedErr := fmt.Errorf("chat completions stream post request failed | provider: %s | url: %s | error: %v", provider, url, err)
			log.Errorf("%v", wrappedErr.Error())
			errChan <- wrappedErr
			return
		}
		defer func() {
			if resp != nil && resp.RawResponse != nil {
				_ = resp.RawResponse.Body.Close()
			}
		}()

		if resp.StatusCode() >= 300 {
			b, err := io.ReadAll(resp.RawResponse.Body)
			if err != nil {
				wrappedErr := fmt.Errorf("chat completions stream read response body failed | provider: %s | url: %s: %w", provider, url, err)
				log.Errorf("%v", wrappedErr)
				errChan <- wrappedErr
				return
			}
			wrappedErr := fmt.Errorf("chat completions stream request failed | provider: %s | url: %s | status: %d | message: %s", provider, url, resp.StatusCode(), string(b))
			log.Errorf("%v", wrappedErr.Error())
			errChan <- wrappedErr
			return
		}

		close(errChan)

		var inThinking bool
		scan := util.NewScanner(resp.RawResponse.Body)
		for scan.Scan() {
			sseData := scan.Text()
			sseResp := respConverter(true, sseData)

			if r, ok := sseResp.(*llmResp); ok {
				r.inThinking = inThinking
			}

			sseResp.ConvertResp()

			if r, ok := sseResp.(*llmResp); ok {
				inThinking = r.inThinking
			}

			select {
			case ret <- sseResp:
			case <-ctx.Done():
				log.Warnf("chat completions stream ctx canceled | provider: %s | url: %s", provider, url)
				return
			}
		}
		// 检查流读取过程中的错误（如网络中断、数据损坏）
		if scanErr := scan.Err(); scanErr != nil {
			log.Errorf("request %v %v chat completions stream scan err: %v", url, provider, scanErr)
			ret <- respConverter(false, scanErr.Error())
		}
	}()
	// 主函数等待错误通道的消息
	select {
	case sseErr, ok := <-errChan:
		if ok { // 通道未关闭，说明有错误
			return nil, sseErr
		}
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return ret, nil
}

const (
	thinkingStartTag = "<think>"
	thinkingEndTag   = "</think>"
)

func extractThinkingFromContent(content string) (reasoning, cleanContent string) {
	startIdx := strings.Index(content, thinkingStartTag)
	endIdx := strings.Index(content, thinkingEndTag)

	if startIdx == -1 {
		return "", content
	}

	reasoning = content[startIdx+len(thinkingStartTag):]
	if endIdx != -1 {
		reasoning = content[startIdx+len(thinkingStartTag) : endIdx]
		cleanContent = content[endIdx+len(thinkingEndTag):]
	} else {
		cleanContent = ""
	}

	return reasoning, cleanContent
}

func extractThinkingFromResp(resp *LLMResp) {
	if resp == nil || len(resp.Choices) == 0 {
		return
	}
	msg := resp.Choices[0].Message
	if msg == nil {
		return
	}
	// 优先级：reasoning_content > reasoning > 从content提取
	if msg.ReasoningContent == nil || *msg.ReasoningContent == "" {
		// 检查 reasoning 字段
		if msg.Reasoning != nil && *msg.Reasoning != "" {
			msg.ReasoningContent = msg.Reasoning
			msg.Reasoning = nil
		} else if msg.Content != "" {
			reasoning, cleanContent := extractThinkingFromContent(msg.Content)
			if reasoning != "" {
				msg.ReasoningContent = &reasoning
				msg.Content = cleanContent
			}
		}
	}
}

func extractThinkingFromDelta(delta *OpenAIMsg, inThinking bool) (bool, *string) {
	if delta == nil {
		return inThinking, nil
	}

	// 如果 reasoning 字段有值，直接转存到 reasoning_content
	if delta.Reasoning != nil && *delta.Reasoning != "" {
		delta.ReasoningContent = delta.Reasoning
		delta.Reasoning = nil
		return inThinking, nil
	}

	if delta.Content == "" {
		return inThinking, nil
	}

	content := delta.Content
	hasStartTag := strings.Contains(content, thinkingStartTag)
	hasEndTag := strings.Contains(content, thinkingEndTag)

	if hasStartTag {
		inThinking = true
	}

	if inThinking {
		startIdx := strings.Index(content, thinkingStartTag)
		endIdx := strings.Index(content, thinkingEndTag)

		if hasStartTag && hasEndTag {
			reasoning := content[startIdx+len(thinkingStartTag) : endIdx]
			delta.Content = content[endIdx+len(thinkingEndTag):]
			delta.ReasoningContent = &reasoning
			inThinking = false
		} else if hasStartTag {
			reasoning := content[startIdx+len(thinkingStartTag):]
			delta.Content = ""
			if reasoning != "" {
				delta.ReasoningContent = &reasoning
			} else {
				emptyStr := ""
				delta.ReasoningContent = &emptyStr
			}
		} else if hasEndTag {
			reasoning := content[:endIdx]
			delta.Content = content[endIdx+len(thinkingEndTag):]
			delta.ReasoningContent = &reasoning
			inThinking = false
		} else {
			delta.ReasoningContent = &content
			delta.Content = ""
		}
	}

	return inThinking, nil
}
