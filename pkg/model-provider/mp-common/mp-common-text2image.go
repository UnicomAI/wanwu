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

type Text2ImageReq struct {
	Prompt         string `json:"prompt" validate:"required"` //提示词
	ResponseFormat string `json:"response_format,omitempty"`  //“url” or “b64_json”,默认b64_json
	ReportUrl      string `json:"report_url,omitempty"`       // 参考
	AdvancedOpt    string `json:"advanced_opt,omitempty"`     // 高级选项参数 json, {"height": 512, "width": 512, "num_images_per_prompt": 1, "style": "摄影"}
}

type AdvancedOptJson struct {
	Height             int    `json:"height,omitempty"`
	Width              int    `json:"width,omitempty"`
	NumImagesPerPrompt int    `json:"num_images_per_prompt,omitempty"`
	Style              string `json:"style,omitempty"`
	NegativePrompt     string `json:"negative_prompt,omitempty"`
}

func (req *Text2ImageReq) Check() error {
	return nil
}

func (req *Text2ImageReq) Data() (map[string]interface{}, error) {
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

type Text2ImageResp struct {
	Code    int      `json:"code"`
	Message string   `json:"msg"`
	Result  []string `json:"result"`
	Usage   T2IUsage `json:"usage"`
}

type T2IUsage struct {
	TotalPatches int `json:"total_patches"`
}

// --- request ---

type IText2ImageReq interface {
	Data() map[string]interface{}
}

// text2ImageReq implementation of IText2ImageReq
type text2ImageReq struct {
	data map[string]interface{}
}

func NewText2ImageReq(data map[string]interface{}) IText2ImageReq {
	return &text2ImageReq{data: data}
}

func (req *text2ImageReq) Data() map[string]interface{} {
	return req.data
}

// --- response ---

type IText2ImageResp interface {
	String() string
	Data() (interface{}, bool)
	ConvertResp() (*Text2ImageResp, bool)
}

// text2ImageResp implementation of IText2ImageResp
type text2ImageResp struct {
	raw string
}

func NewText2ImageResp(raw string) IText2ImageResp {
	return &text2ImageResp{raw: raw}
}

func (resp *text2ImageResp) String() string {
	return resp.raw
}

func (resp *text2ImageResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("text2Image resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *text2ImageResp) ConvertResp() (*Text2ImageResp, bool) {
	var ret *Text2ImageResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("text2Image resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	if err := util.Validate(ret); err != nil {
		log.Errorf("text2Image resp validate err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- text2Image ---

func Text2Image(ctx context.Context, provider, apiKey, url string, req map[string]interface{}, headers ...Header) ([]byte, error) {
	if apiKey != "" {
		headers = append(headers, Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}
	formdata := util.ConvertMapToString(req)

	request := trace_util.NewResty(ctx).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // 关闭证书校验
		SetTimeout(0).                                             // 关闭请求超时
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "multipart/form-data").
		SetHeader("Accept", "application/json").
		SetMultipartFormData(formdata).
		SetDoNotParseResponse(true)
	for _, header := range headers {
		request.SetHeader(header.Key, header.Value)
	}

	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("request %v %v text2Image err: %v", url, provider, err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v text2Image read response body failed: %v", url, provider, err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v text2Image http status %v msg: %v", url, provider, resp.StatusCode(), string(b))
	}
	return b, nil
}
