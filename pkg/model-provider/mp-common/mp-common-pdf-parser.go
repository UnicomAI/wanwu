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

type PdfParserReq struct {
	Files    *multipart.FileHeader `form:"file" json:"file" validate:"required"`
	FileName string                `form:"file_name" json:"file_name" validate:"required"`
}

func (req *PdfParserReq) Check() error {
	return nil
}

func (req *PdfParserReq) Data() (map[string]interface{}, error) {
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

type PdfParserResp struct {
	Code    string `json:"code" validate:"required"`
	Content string `json:"content" validate:"required"`
	Message string `json:"message" validate:"required"`
	Status  string `json:"status"`
	TraceId string `json:"trace_id"`
}

// --- request ---

type IPdfParserReq interface {
	Data() *PdfParserReq
}

// pdfParserReq implementation of IPdfParserReq
type pdfParserReq struct {
	data *PdfParserReq
}

func NewPdfParserReq(data *PdfParserReq) IPdfParserReq {
	return &pdfParserReq{data: data}
}

func (req *pdfParserReq) Data() *PdfParserReq {
	return req.data
}

// --- response ---

type IPdfParserResp interface {
	String() string
	Data() (interface{}, bool)
	ConvertResp() (*PdfParserResp, bool)
}

// pdfParserResp implementation of IPdfParserResp
type pdfParserResp struct {
	raw string
}

func NewPdfParserResp(raw string) IPdfParserResp {
	return &pdfParserResp{raw: raw}
}

func (resp *pdfParserResp) String() string {
	return resp.raw
}

func (resp *pdfParserResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("pdfParser resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *pdfParserResp) ConvertResp() (*PdfParserResp, bool) {
	var ret *PdfParserResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("pdfParser resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	if err := util.Validate(ret); err != nil {
		log.Errorf("pdfParser resp (%v) validate err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

// --- pdfParser ---

func PdfParser(ctx *gin.Context, provider, apiKey, url string, req *PdfParserReq, headers ...Header) ([]byte, error) {
	if apiKey != "" {
		headers = append(headers, Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}
	file, err := req.Files.Open()
	if err != nil {
		return nil, fmt.Errorf("request %v %v pdfParser err: %v", url, provider, err)
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
		SetMultipartFormData(map[string]string{
			"file_name": req.FileName, // 👈 Add file_name field here
		}).
		SetDoNotParseResponse(true)
	for _, header := range headers {
		request.SetHeader(header.Key, header.Value)
	}
	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("request %v %v pdfParser err: %v", url, provider, err)
	}
	bodyBytes, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %v", err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v pdfParser http status %v msg: %v", url, provider, resp.StatusCode(), string(bodyBytes))
	}

	return bodyBytes, nil
}
