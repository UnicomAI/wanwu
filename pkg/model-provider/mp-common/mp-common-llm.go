package mp_common

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
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
	TagChat          string = "CHAT"
	TagEmbedding     string = "Embedding"
	TagRerank        string = "Rerank"
	TagGui           string = "GUI"
	TagOcr           string = "OCR"
	TagPdfParser     string = "文档解析"
	TagVisionSupport string = "图文问答"
	TagToolCall      string = "工具调用"
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

func GetTagsByContentSize(size *int) []Tag {
	var tags []Tag
	if size != nil && *size > 0 {
		kValue := *size / 1024
		// Format to "XK" string and add to tags list
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
	Thinking            *Thinking      `json:"thinking,omitempty"` // Controls whether the model turns on deep thinking mode.
	EnableThinking      *bool          `json:"enable_thinking,omitempty"`
	MaxCompletionTokens *int           `json:"max_completion_tokens,omitempty"` // Controls the maximum length of model output [0,64k]
	LogitBias           map[string]int `json:"logit_bias,omitempty"`            // Adjust the probability of the specified token appearing in the model output content
	ToolChoice          interface{}    `json:"tool_choice,omitempty"`           // Force the policy to be called by the specified tool
	TopP                *float64       `json:"top_p,omitempty"`
	TopK                *int           `json:"top_k,omitempty"`
	MinP                *float64       `json:"min_p,omitempty"`
	ParallelToolCalls   *bool          `json:"parallel_tool_calls,omitempty"` // Whether to enable parallel tool calls
	StreamOptions       *StreamOptions `json:"stream_options,omitempty"`      //When streaming output is enabled, the number of tokens used can be displayed in the last line of the output by setting this parameter to {"include_usage": true}.

	PresencePenalty   *float64 `json:"presence_penalty,omitempty"`   // Control the content duplication when the model generates text
	FrequencyPenalty  *float64 `json:"frequency_penalty,omitempty"`  // frequency penalty coefficient
	RepetitionPenalty *float64 `json:"repetition_penalty,omitempty"` // The degree of repetition in the continuous sequence when the model is generated

	Seed           *int  `json:"seed,omitempty"`         // seed
	Logprobs       *bool `json:"logprobs,omitempty"`     // Whether to return the logarithmic probability of the output Token
	TopLogprobs    *int  `json:"top_logprobs,omitempty"` // Specify the number of candidate Tokens with the highest probability of returning to the model at each step of generation.
	N              *int  `json:"n,omitempty"`
	ThinkingBudget *int  `json:"thinking_budget,omitempty"` // The maximum length of the thinking process, only takes effect when enable_thinking is true

	WebSearch *WebSearch `json:"web_search,omitempty"` //Search enhancement
	User      *string    `json:"user,omitempty"`
	// Yuanjing
	DoSample  *bool      `json:"do_sample,omitempty"`
	ExtraBody *ExtraBody `json:"extra_body,omitempty"` // Extended parameters
}

type OpenAIReqMsg struct {
	Role             MsgRole       `json:"role"` // "system" | "user" | "assistant" | "function(deprecated)"
	Content          interface{}   `json:"content"`
	ToolCallId       *string       `json:"tool_call_id,omitempty"`
	ReasoningContent *string       `json:"reasoning_content,omitempty"`
	Name             *string       `json:"name,omitempty"`
	FunctionCall     *FunctionCall `json:"function_call,omitempty"`
	ToolCalls        []*ToolCall   `json:"tool_calls,omitempty"`
}

type ExtraBody struct {
	ApiOption string `json:"api_option"` // Select the specified function. 1) math: take photos to answer questions; 2) ocr: multi-modal OCR; 3) general: general scenarios.   By default, the intent will be judged based on the prompt.
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
	Role             MsgRole       `json:"role"` // "system" | "user" | "assistant" | "function(deprecated)"
	Content          string        `json:"content"`
	ToolCallId       *string       `json:"tool_call_id,omitempty"`
	ReasoningContent *string       `json:"reasoning_content,omitempty"`
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
	Name        string                    `json:"name" validate:"required"`
	Description string                    `json:"description,omitempty"`
	Parameters  *OpenAIFunctionParameters `json:"parameters,omitempty"`
}

type OpenAIFunctionParameters struct {
	Type       string                                      `json:"type"`
	Properties map[string]OpenAIFunctionParametersProperty `json:"properties"`
	Required   []string                                    `json:"required"`
}
type OpenAIFunctionParametersProperty struct {
	Description string `json:"description"`
	Type        string `json:"type"`
}

func (req *LLMReq) Check() error { return nil }

// --- openapi response ---

type LLMResp struct {
	ID                string             `json:"id"`                               // unique identifier
	Object            string             `json:"object"`                           // Fixed to "chat.completion"
	Created           int                `json:"created"`                          // Timestamp (seconds)
	Model             string             `json:"model" validate:"required"`        // Model used
	Choices           []OpenAIRespChoice `json:"choices" validate:"required,dive"` // Generate result list
	Usage             OpenAIRespUsage    `json:"usage"`                            // Token usage statistics
	ServiceTier       *string            `json:"service_tier"`                     // (Volcano) Specifies whether to use TPM assurance package. The effective object is the inference access point that purchased the guarantee package.
	SystemFingerprint *string            `json:"system_fingerprint"`
	Code              *int               `json:"code,omitempty"`
	ImgId             *string            `json:"img_id,omitempty"` // The visual model returns the image id
}

// The OpenAIRespUsage structure represents token consumption
type OpenAIRespUsage struct {
	CompletionTokens int `json:"completion_tokens"` // Output token number
	PromptTokens     int `json:"prompt_tokens"`     // Enter the number of tokens
	TotalTokens      int `json:"total_tokens"`      // Total number of tokens
}

// The OpenAIRespChoice structure represents a single build option
type OpenAIRespChoice struct {
	Index        int         `json:"index"`             // Option index
	Message      *OpenAIMsg  `json:"message,omitempty"` // Non-streaming generated messages
	Delta        *OpenAIMsg  `json:"delta,omitempty"`   // Streaming generated messages
	FinishReason string      `json:"finish_reason"`     // Stop reason
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
	Data() (map[string]interface{}, bool)
	ConvertResp() (*LLMResp, bool)
}

// llmResp implementation of ILLMResp
type llmResp struct {
	stream bool
	raw    string
}

func NewLLMResp(stream bool, raw string) ILLMResp {
	return &llmResp{stream: stream, raw: raw}
}

func (resp *llmResp) String() string {
	return resp.raw
}

func (resp *llmResp) Data() (map[string]interface{}, bool) {
	if resp.stream {
		if resp.raw == "data: [DONE]" || !strings.HasPrefix(resp.raw, "data:") {
			return nil, false
		}
	}

	raw := resp.raw
	if resp.stream {
		raw = strings.TrimPrefix(resp.raw, "data:")
	}

	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(raw), &ret); err != nil {
		log.Errorf("llm stream resp (%v) convert to data err: %v", raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *llmResp) ConvertResp() (*LLMResp, bool) {
	if resp.stream {
		if resp.raw == "data: [DONE]" || !strings.HasPrefix(resp.raw, "data:") {
			return nil, false
		}
	}

	raw := resp.raw
	if resp.stream {
		raw = strings.TrimPrefix(resp.raw, "data:")
	}

	ret := &LLMResp{}
	if err := json.Unmarshal([]byte(raw), ret); err != nil {
		log.Errorf("llm stream resp (%v) convert to openai resp err: %v", raw, err)
		return nil, false
	}

	if err := util.Validate(ret); err != nil {
		log.Errorf("llm resp validate err: %v", err)
		return nil, false
	}
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

	request := resty.New().
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // Turn off certificate verification
		SetTimeout(0).                                             // Close request timeout
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
	} else if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v chat completions unary http status %v msg: %v", url, provider, resp.StatusCode(), resp.String())
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v chat completions unary read response body err: %v", url, provider, err)
	}
	return respConverter(false, string(b)), nil
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
	go func() {
		defer util.PrintPanicStack()
		defer close(ret)
		request := resty.New().
			SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // Turn off certificate verification
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
			log.Errorf("request %v %v chat completions stream err: %v", url, provider, err)
			return
		} else if resp.StatusCode() >= 300 {
			log.Errorf("request %v %v chat completions stream http status %v msg: %v", url, provider, resp.StatusCode(), resp.String())
			return
		}
		scan := bufio.NewScanner(resp.RawResponse.Body)
		for scan.Scan() {
			ret <- respConverter(true, scan.Text())
		}
	}()
	return ret, nil
}
