package mp_common

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"

	"github.com/UnicomAI/wanwu/pkg/log"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

// --- openapi request ---

type TextRerankReq struct {
	Documents       []string `json:"documents" validate:"required"` // 需要重排序的文本
	Model           string   `json:"model" validate:"required"`
	Query           string   `json:"query" validate:"required"`
	ReturnDocuments *bool    `json:"return_documents,omitempty"`
	TopN            *int     `json:"top_n,omitempty"`       // 返回排序后的top_n个文档。默认返回全部文档。
	User            *string  `json:"user,omitempty"`        // 用户标识（兼容千帆)
	Instruction     *string  `json:"instruction,omitempty"` // 指令内容（适配元景qwen_rerank）
}

func (req *TextRerankReq) Check() error {
	if req.TopN != nil && *req.TopN < 0 {
		return fmt.Errorf("top_n must greater than 0")
	}
	return nil
}

func (req *TextRerankReq) Data() (map[string]interface{}, error) {
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

// --- openapi response ---

type RerankResp struct {
	Results   []Result `json:"results" validate:"required,dive"`
	Model     string   `json:"model"`
	Object    *string  `json:"object,omitempty"`
	Usage     Usage    `json:"usage"`
	RequestId *string  `json:"request_id,omitempty"`
}

type Result struct {
	Index          int         `json:"index"`
	Document       interface{} `json:"document,omitempty"`
	RelevanceScore float64     `json:"relevance_score" validate:"required"`
}

type Document struct {
	Text string `json:"text"`
}

type Usage struct {
	PromptTokens     int  `json:"prompt_tokens,omitempty"`
	TotalTokens      int  `json:"total_tokens,omitempty"`
	CompletionTokens int  `json:"completion_tokens,omitempty"`
	ImageTokens      *int `json:"image_tokens,omitempty"`
	InputTokens      *int `json:"input_tokens,omitempty"`
}

// --- request ---

type ITextRerankReq interface {
	Data() map[string]interface{}
}

// rerankReq implementation of ITextRerankReq
type rerankReq struct {
	data map[string]interface{}
}

func NewRerankReq(data map[string]interface{}) ITextRerankReq {
	return &rerankReq{data: data}
}

func (req *rerankReq) Data() map[string]interface{} {
	return req.data
}

// --- response ---

type ITextRerankResp interface {
	String() string
	Data() (interface{}, bool)
	ConvertResp() (*RerankResp, bool)
}

// rerankResp implementation of ITextRerankResp
type rerankResp struct {
	raw string
}

func NewRerankResp(raw string) ITextRerankResp {
	return &rerankResp{raw: raw}
}

func (resp *rerankResp) String() string {
	return resp.raw
}

func (resp *rerankResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *rerankResp) ConvertResp() (*RerankResp, bool) {
	var ret *RerankResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	if err := util.Validate(ret); err != nil {
		log.Errorf("rerank resp validate err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- rerank ---

func Rerank(ctx context.Context, provider, apiKey, url string, req map[string]interface{}, headers ...Header) ([]byte, error) {
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
		SetBody(req).
		SetDoNotParseResponse(true)
	for _, header := range headers {
		request.SetHeader(header.Key, header.Value)
	}

	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("request %v %v rerank err: %v", url, provider, err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v rerank read response body failed: %v", url, provider, err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v rerank http status %v msg: %v", url, provider, resp.StatusCode(), string(b))
	}
	return b, nil
}
