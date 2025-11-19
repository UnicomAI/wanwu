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

type EmbeddingReq struct {
	Model          string   `json:"model" validate:"required"`
	Input          []string `json:"input" validate:"required"`
	EncodingFormat *string  `json:"encoding_format,omitempty"`
}

func (req *EmbeddingReq) Check() error { return nil }

func (req *EmbeddingReq) Data() (map[string]interface{}, error) {
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

type EmbeddingResp struct {
	Id      *string         `json:"id,omitempty"`
	Model   string          `json:"model" validate:"required"`
	Object  *string         `json:"object,omitempty"`
	Data    []EmbeddingData `json:"data" validate:"required,dive"`
	Usage   Usage           `json:"usage"`
	Created *int            `json:"created,omitempty"`
}

type EmbeddingData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding" validate:"required,min=1"`
	Index     int       `json:"index"`
}

// --- request ---

type IEmbeddingReq interface {
	Data() map[string]interface{}
}

// embeddingReq implementation of IEmbeddingReq
type embeddingReq struct {
	data map[string]interface{}
}

func NewEmbeddingReq(data map[string]interface{}) IEmbeddingReq {
	return &embeddingReq{data: data}
}

func (req *embeddingReq) Data() map[string]interface{} {
	return req.data
}

// --- response ---

type IEmbeddingResp interface {
	String() string
	Data() (map[string]interface{}, bool)
	ConvertResp() (*EmbeddingResp, bool)
}

// embeddingResp implementation of IEmbeddingResp
type embeddingResp struct {
	raw string
}

func NewEmbeddingResp(raw string) IEmbeddingResp {
	return &embeddingResp{raw: raw}
}

func (resp *embeddingResp) String() string {
	return resp.raw
}

func (resp *embeddingResp) Data() (map[string]interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("embedding resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *embeddingResp) ConvertResp() (*EmbeddingResp, bool) {
	var ret *EmbeddingResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("embedding resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	if err := util.Validate(ret); err != nil {
		log.Errorf("embedding resp validate err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- embedding ---

func Embeddings(ctx context.Context, provider, apiKey, url string, req map[string]interface{}, headers ...Header) ([]byte, error) {
	if apiKey != "" {
		headers = append(headers, Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}

	request := resty.New().
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // 关闭证书校验 [EN] Turn off certificate verification
		SetTimeout(0).                                             // 关闭请求超时 [EN] Close request timeout
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
		return nil, fmt.Errorf("request %v %v embeddings err: %v", url, provider, err)
	} else if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v embeddings http status %v msg: %v", url, provider, resp.StatusCode(), resp.String())
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v embeddings read response body err: %v", url, provider, err)
	}
	return b, nil
}
