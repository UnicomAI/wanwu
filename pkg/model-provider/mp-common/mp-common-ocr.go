package mp_common

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

// --- openapi request ---

type OcrReq struct {
	Files *multipart.FileHeader `form:"file" json:"file" validate:"required"`
}

func (req *OcrReq) Check() error {
	return nil
}

func (req *OcrReq) Data() (map[string]interface{}, error) {
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

type OcrResp struct {
	Code      int       `json:"code"`
	Message   string    `json:"message" validate:"required"`
	Version   string    `json:"version"`
	TimeStamp string    `json:"timestamp"`
	Id        string    `json:"id"`
	TimeCost  float64   `json:"time_cost"`
	OcrData   []OcrData `json:"data" validate:"required,dive"`
}
type OcrData struct {
	PageNum []int  `json:"page_num" validate:"required,min=1"`
	Type    string `json:"type" validate:"required"`
	Text    string `json:"text"`
	Length  int    `json:"length"`
}

// --- request ---

type IOcrReq interface {
	Data() *OcrReq
}

// ocrReq implementation of IOcrReq
type ocrReq struct {
	data *OcrReq
}

func NewOcrReq(data *OcrReq) IOcrReq {
	return &ocrReq{data: data}
}

func (req *ocrReq) Data() *OcrReq {
	return req.data
}

// --- response ---

type IOcrResp interface {
	String() string
	Data() (interface{}, bool)
	ConvertResp() (*OcrResp, bool)
}

// ocrResp implementation of IOcrResp
type ocrResp struct {
	raw string
}

func NewOcrResp(raw string) IOcrResp {
	return &ocrResp{raw: raw}
}

func (resp *ocrResp) String() string {
	return resp.raw
}

func (resp *ocrResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("ocr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *ocrResp) ConvertResp() (*OcrResp, bool) {
	var ret *OcrResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("ocr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	if err := util.Validate(ret); err != nil {
		log.Errorf("ocr resp validate err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- ocr ---

func Ocr(ctx *gin.Context, provider, apiKey, url string, req *OcrReq, headers ...Header) ([]byte, error) {
	if apiKey != "" {
		headers = append(headers, Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}
	file, err := req.Files.Open()
	if err != nil {
		return nil, fmt.Errorf("request %v %v ocr err: %v", url, provider, err)
	}
	defer file.Close()
	request := resty.New().
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // Turn off certificate verification
		SetTimeout(0).                                             // Close request timeout
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "multipart/form-data").
		SetHeader("Accept", "application/json").
		SetFileReader("file", req.Files.Filename, file).
		SetDoNotParseResponse(true)
	for _, header := range headers {
		request.SetHeader(header.Key, header.Value)
	}

	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("request %v %v ocr err: %v", url, provider, err)
	} else if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v ocr http status %v msg: %v", url, provider, resp.StatusCode(), resp.String())
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v ocr read response body err: %v", url, provider, err)
	}
	return b, nil
}
