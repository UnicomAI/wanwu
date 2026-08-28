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

type MultiModalEmbeddingReq struct {
	Model          string       `json:"model" validate:"required"`
	Input          []MultiInput `json:"input" validate:"required"`
	EncodingFormat string       `json:"encoding_format,omitempty"`
	Dimensions     *int         `json:"dimensions,omitempty"`
	Parameters     interface{}  `json:"parameters,omitempty"`
}
type MultiInput struct {
	Text  *string `json:"text,omitempty"`
	Image *string `json:"image,omitempty"`
	Audio *string `json:"audio,omitempty"`
	Video *string `json:"video,omitempty"`
}

func (req *MultiModalEmbeddingReq) Check() error { return nil }

func (req *MultiModalEmbeddingReq) Data() (map[string]interface{}, error) {
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

type MultiModalEmbeddingResp struct {
	Id      *string         `json:"id,omitempty"`
	Model   string          `json:"model" validate:"required"`
	Object  *string         `json:"object,omitempty"`
	Data    []EmbeddingData `json:"data" validate:"required,dive"`
	Usage   Usage           `json:"usage"`
	Created *int            `json:"created,omitempty"`
}

// --- request ---

type IMultiModalEmbeddingReq interface {
	Data() map[string]interface{}
}

// multiModalEmbeddingReq implementation of IMultiModalEmbeddingReq
type multiModalEmbeddingReq struct {
	data map[string]interface{}
}

func NewMultiModalEmbeddingReq(data map[string]interface{}) IMultiModalEmbeddingReq {
	return &multiModalEmbeddingReq{data: data}
}

func (req *multiModalEmbeddingReq) Data() map[string]interface{} {
	return req.data
}

// --- response ---

type IMultiModalEmbeddingResp interface {
	String() string
	Data() (map[string]interface{}, bool)
	ConvertResp() (*MultiModalEmbeddingResp, bool)
}

// multiModalEmbeddingResp implementation of IMultiModalEmbeddingResp
type multiModalEmbeddingResp struct {
	raw string
}

func NewMultiModalEmbeddingResp(raw string) IMultiModalEmbeddingResp {
	return &multiModalEmbeddingResp{raw: raw}
}

func (resp *multiModalEmbeddingResp) String() string {
	return resp.raw
}

func (resp *multiModalEmbeddingResp) Data() (map[string]interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("multiModalEmbedding resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *multiModalEmbeddingResp) ConvertResp() (*MultiModalEmbeddingResp, bool) {
	var ret *MultiModalEmbeddingResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("multiModalEmbedding resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	if err := util.Validate(ret); err != nil {
		log.Errorf("multiModalEmbedding resp validate err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- multiModalEmbedding ---

func MultiModalEmbeddings(ctx context.Context, provider, apiKey, url string, req map[string]interface{}, headers ...Header) ([]byte, error) {
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
		return nil, fmt.Errorf("request %v %v multimodal-embeddings err: %v", url, provider, err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v multimodal-embeddings read response body err: %v", url, provider, err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v multimodal-embeddings http status %v msg: %v", url, provider, resp.StatusCode(), string(b))
	}
	return b, nil
}
