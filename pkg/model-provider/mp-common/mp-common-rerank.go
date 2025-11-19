package mp_common

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/go-resty/resty/v2"
)

// --- openapi request ---

type RerankReq struct {
	Documents       []string `json:"documents" validate:"required"`
	Model           string   `json:"model" validate:"required"`
	Query           string   `json:"query" validate:"required"`
	ReturnDocuments *bool    `json:"return_documents,omitempty"`
	TopN            *int     `json:"top_n,omitempty"`
}

func (req *RerankReq) Check() error {
	if req.TopN != nil && *req.TopN < 0 {
		return fmt.Errorf("top_n must greater than 0")
	}
	return nil
}

func (req *RerankReq) Data() (map[string]interface{}, error) {
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
	Index          int       `json:"index"`
	Document       *Document `json:"document,omitempty"`
	RelevanceScore float64   `json:"relevance_score" validate:"required"`
}

type Document struct {
	Text string `json:"text"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

// --- request ---

type IRerankReq interface {
	Data() map[string]interface{}
}

// rerankReq implementation of IRerankReq
type rerankReq struct {
	data map[string]interface{}
}

func NewRerankReq(data map[string]interface{}) IRerankReq {
	return &rerankReq{data: data}
}

func (req *rerankReq) Data() map[string]interface{} {
	return req.data
}

// --- response ---

type IRerankResp interface {
	String() string
	Data() (interface{}, bool)
	ConvertResp() (*RerankResp, bool)
}

// rerankResp implementation of IRerankResp
type rerankResp struct {
	raw string
}

func NewRerankResp(raw string) IRerankResp {
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

	request := resty.New().
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // Turn off certificate verification
		SetTimeout(0).                                             // Close request timeout
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
	} else if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v rerank http status %v msg: %v", url, provider, resp.StatusCode(), resp.String())
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v rerank read response body err: %v", url, provider, err)
	}
	return b, nil
}
