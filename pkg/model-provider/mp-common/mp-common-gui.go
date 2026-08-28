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

type GuiReq struct {
	Algo                    string   `json:"algo,omitempty"`                                     // 算法名称，默认gui_agent_v1
	Platform                string   `json:"platform" validate:"required"`                       // 平台信息，移动端填写Mobile, Windows端填写WIN，Mac端填写MAC
	CurrentScreenshotXml    string   `json:"current_screenshot_xml,omitempty"`                   // 屏幕布局导出的xml文件
	CurrentScreenshot       string   `json:"current_screenshot" validate:"required"`             // 当前屏幕截图，Base64编码的图像字符串
	CurrentScreenshotWidth  int      `json:"current_screenshot_width" validate:"required,gt=0"`  // 当前屏幕截图的宽度
	CurrentScreenshotHeight int      `json:"current_screenshot_height" validate:"required,gt=0"` // 当前屏幕截图的高度
	Task                    string   `json:"task" validate:"required"`                           //当前用户任务
	History                 []string `json:"history"`                                            //当前任务的历史返回结果，历次返回结果中的content字段
}

func (req *GuiReq) Check() error {
	if req.History == nil {
		req.History = make([]string, 0)
	}
	return nil
}

func (req *GuiReq) Data() (map[string]interface{}, error) {
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

type GuiResp struct {
	Code    int        `json:"code"`
	Message string     `json:"message" validate:"required"`
	Content GuiContent `json:"content" validate:"required"`
	Usage   Usage      `json:"usage" validate:"required"`
}

type GuiContent struct {
	Description string `json:"description" validate:"required"`
	Operation   string `json:"operation" validate:"required"`
	Action      string `json:"action" validate:"required"`
	Box         []int  `json:"box"`
	Value       string `json:"value"`
	Sensitivity string `json:"sensitivity"`
}

// --- request ---

type IGuiReq interface {
	Data() *GuiReq
}

// guiReq implementation of IGuiReq
type guiReq struct {
	data *GuiReq
}

func NewGuiReq(data *GuiReq) IGuiReq {
	return &guiReq{data: data}
}

func (req *guiReq) Data() *GuiReq {
	return req.data
}

// --- response ---

type IGuiResp interface {
	String() string
	Data() (interface{}, bool)
	ConvertResp() (*GuiResp, bool)
}

// guiResp implementation of IGuiResp
type guiResp struct {
	raw string
}

func NewGuiResp(raw string) IGuiResp {
	return &guiResp{raw: raw}
}

func (resp *guiResp) String() string {
	return resp.raw
}

func (resp *guiResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("gui resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *guiResp) ConvertResp() (*GuiResp, bool) {
	var ret *GuiResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("gui resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	log.Infof("gui resp: %v", resp.raw)
	if err := util.Validate(ret); err != nil {
		log.Errorf("gui resp validate err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- gui ---

func Gui(ctx context.Context, provider, apiKey, url string, req *GuiReq, headers ...Header) ([]byte, error) {
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
		return nil, fmt.Errorf("request %v %v gui err: %v", url, provider, err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v gui read response body err: %v", url, provider, err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v gui http status %v msg: %v", url, provider, resp.StatusCode(), string(b))
	}
	return b, nil
}
