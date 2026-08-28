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
//query 分俩种：
//1. "query": "查询文本"
//2. "query": {
//   "text": "查询文本"
//   "image": "图片base64编码 或 url"
// },

type MultiModalRerankReq struct {
	Documents       []MultiDocument `json:"documents" validate:"required"` // 需要重排序的内容列表
	Model           string          `json:"model" validate:"required"`
	Query           interface{}     `json:"query" validate:"required"`  // 重排序的查询内容
	ReturnDocuments *bool           `json:"return_documents,omitempty"` // 是否返回排序前的文档。默认为true
	TopN            *int            `json:"top_n,omitempty"`            // 返回排序后的top_n个文档。默认返回全部文档。
	User            *string         `json:"user,omitempty"`             // 用户标识（兼容千帆)
	Instruction     *string         `json:"instruction,omitempty"`      // 指令内容（适配元景qwen_rerank）
	Fps             *float64        `json:"fps,omitempty"`              // 视频帧率（适配qwen rerank）
}

type MultiDocument struct {
	Text  *string `json:"text,omitempty"`
	Image *string `json:"image,omitempty"`
	Video *string `json:"video,omitempty"`
}

func (req *MultiModalRerankReq) Check() error {
	if req.TopN != nil && *req.TopN < 0 {
		return fmt.Errorf("top_n must greater than 0")
	}
	if req.ReturnDocuments == nil {
		defaultReturnDocuments := true
		req.ReturnDocuments = &defaultReturnDocuments
	}
	return nil
}

func (req *MultiModalRerankReq) Data() (map[string]interface{}, error) {
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

type MultiModalRerankResp struct {
	Results   []Result `json:"results" validate:"required,dive"`
	Model     string   `json:"model"`
	Object    *string  `json:"object,omitempty"`
	Usage     Usage    `json:"usage"`
	RequestId *string  `json:"request_id,omitempty"`
}

// --- request ---

type IMultiModalRerankReq interface {
	Data() map[string]interface{}
}

// multiModalRerankReq implementation of IMultiModalRerankReq
type multiModalRerankReq struct {
	data map[string]interface{}
}

func NewMultiModalRerankReq(data map[string]interface{}) IMultiModalRerankReq {
	return &multiModalRerankReq{data: data}
}

func (req *multiModalRerankReq) Data() map[string]interface{} {
	return req.data
}

// --- response ---

type IMultiModalRerankResp interface {
	String() string
	Data() (interface{}, bool)
	ConvertResp() (*MultiModalRerankResp, bool)
}

// multiModalRerankResp implementation of IMultiModalRerankResp
type multiModalRerankResp struct {
	raw string
}

func NewMultiModalRerankResp(raw string) IMultiModalRerankResp {
	return &multiModalRerankResp{raw: raw}
}

func (resp *multiModalRerankResp) String() string {
	return resp.raw
}

func (resp *multiModalRerankResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("multimodal-rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *multiModalRerankResp) ConvertResp() (*MultiModalRerankResp, bool) {
	var ret *MultiModalRerankResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("multimodal-rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	if err := util.Validate(ret); err != nil {
		log.Errorf("multimodal-rerank resp validate err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- multimodal-rerank ---

func MultiModalRerank(ctx context.Context, provider, apiKey, url string, req map[string]interface{}, headers ...Header) ([]byte, error) {
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
		return nil, fmt.Errorf("request %v %v multimodal-rerank err: %v", url, provider, err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v multimodal-rerank read response body failed: %v", url, provider, err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v multimodal-rerank http status %v msg: %v", url, provider, resp.StatusCode(), string(b))
	}
	return b, nil
}
