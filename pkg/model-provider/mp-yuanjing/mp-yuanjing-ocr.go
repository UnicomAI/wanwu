package mp_yuanjing

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

type Ocr struct {
	ApiKey           string   `json:"apiKey"`           // ApiKey
	EndpointUrl      string   `json:"endpointUrl"`      // 推理url
	SupportFileTypes []string `json:"supportFileTypes"` // 支持的文件类型，由 bff-service 从 recommend_model_config.yaml 注入
}

func (cfg *Ocr) Tags() []mp_common.Tag {
	tags := []mp_common.Tag{
		{
			Text: mp_common.TagOcr,
		},
	}
	return tags
}

func (cfg *Ocr) NewReq(req *mp_common.OcrReq) (mp_common.IOcrReq, error) {
	m, err := req.Data()
	if err != nil {
		return nil, err
	}
	return mp_common.NewOcrReq(m), nil
}

func (cfg *Ocr) Ocr(ctx *gin.Context, req mp_common.IOcrReq, headers ...mp_common.Header) (mp_common.IOcrResp, error) {
	ocrReq := req.Data()

	// base64 数据转为 multipart/form-data 调用下游
	b, err := cfg.ocrWithMultipart(ctx, ocrReq, headers...)
	if err != nil {
		return nil, err
	}
	return &ocrResp{raw: string(b), fileName: util.GetStringFromMap(ocrReq, "fileName")}, nil
}

// ocrWithMultipart 将 base64 文件数据通过 multipart/form-data 发送到下游 OCR 服务
func (cfg *Ocr) ocrWithMultipart(ctx *gin.Context, m map[string]interface{}, headers ...mp_common.Header) ([]byte, error) {
	fileData := util.GetStringFromMap(m, "data")
	if fileData == "" {
		return nil, fmt.Errorf("file data is empty")
	}

	// 解析 base64 数据（兼容 data:xxx;base64, 前缀格式）
	base64Str := fileData
	if idx := strings.Index(base64Str, ","); idx >= 0 && strings.HasPrefix(base64Str, "data:") {
		base64Str = base64Str[idx+1:]
	}

	// base64 解码为文件字节
	fileBytes, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return nil, fmt.Errorf("base64 decode err: %v", err)
	}

	fileName := util.GetStringFromMap(m, "fileName")
	// 使用 util.FileData2FileHeader 将字节数据转为 multipart.FileHeader
	fileHeader, err := util.FileData2FileHeader(fileName, fileBytes)
	if err != nil {
		return nil, fmt.Errorf("convert file data to file header err: %v", err)
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("open file header err: %v", err)
	}
	defer func() { _ = file.Close() }()

	if apiKey := cfg.ApiKey; apiKey != "" {
		headers = append(headers, mp_common.Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}

	// 构建 multipart 表单数据
	formData := map[string]string{
		"file_name": fileName,
	}
	if v := util.GetIntFromMap(m, "extract_image"); v != nil {
		formData["extract_image"] = strconv.Itoa(*v)
	}
	if v := util.GetIntFromMap(m, "extract_image_content"); v != nil {
		formData["extract_image_content"] = strconv.Itoa(*v)
	}
	if v := util.GetStringFromMap(m, "model"); v != "" {
		formData["model"] = v
	}
	if v := util.GetIntFromMap(m, "fileType"); v != nil {
		formData["fileType"] = strconv.Itoa(*v)
	}

	request := trace_util.NewResty(ctx).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // 关闭证书校验
		SetTimeout(0).                                             // 关闭请求超时
		R().
		SetContext(ctx).
		SetHeader("Accept", "application/json").
		SetFileReader("file", fileHeader.Filename, file).
		SetMultipartFormData(formData).
		SetDoNotParseResponse(true)
	for _, header := range headers {
		request.SetHeader(header.Key, header.Value)
	}

	resp, err := request.Post(cfg.ocrUrl())
	if err != nil {
		return nil, fmt.Errorf("request %v yuanjing ocr err: %v", cfg.ocrUrl(), err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v yuanjing ocr read response body err: %v", cfg.ocrUrl(), err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v yuanjing ocr http status %v msg: %v", cfg.ocrUrl(), resp.StatusCode(), string(b))
	}
	return b, nil
}

func (cfg *Ocr) ocrUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/rag/model_parser_file")
	return ret
}

// --- ocrResp 下游原生响应结构体，实现 mp_common.IOcrResp 接口 ---

// ocrRawResp 下游（原 pdf-parser）返回的原生响应结构
type ocrRawResp struct {
	Code           string `json:"code"`
	Content        string `json:"content"`
	Message        string `json:"message"`
	Status         string `json:"status"`
	TraceId        string `json:"trace_id"`
	PrefixImageUrl string `json:"prefix_image_url"`
	Version        string `json:"version"`
}

type ocrResp struct {
	raw      string
	fileName string
}

func (resp *ocrResp) String() string {
	return resp.raw
}

func (resp *ocrResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("yuanjing ocr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *ocrResp) ConvertResp() (*mp_common.OcrResp, bool) {
	var raw ocrRawResp
	if err := json.Unmarshal([]byte(resp.raw), &raw); err != nil {
		log.Errorf("yuanjing ocr resp (%v) unmarshal err: %v", resp.raw, err)
		return nil, false
	}

	if err := util.Validate(&raw); err != nil {
		log.Errorf("yuanjing ocr resp validate err: %v", err)
		return nil, false
	}
	log.Infof("yuanjing ocr resp PrefixImageUrl : %v", raw.PrefixImageUrl)
	// 将下游原生响应转换为统一 OcrResp 格式
	target := resp.buildTargetOcrResp(&raw)
	return target, true
}

func (resp *ocrResp) buildTargetOcrResp(raw *ocrRawResp) *mp_common.OcrResp {
	code := 1 // 默认异常
	if raw.Code == "0" || raw.Code == "200" {
		code = 0
	}

	// 从请求参数中获取文件名
	fileName := resp.fileName

	// 从文件名推断文件类型
	fileType := ""
	if fileName != "" {
		lower := strings.ToLower(fileName)
		if strings.HasSuffix(lower, ".pdf") {
			fileType = "pdf"
		} else {
			fileType = "image"
		}
	}

	target := &mp_common.OcrResp{
		Code:    code,
		Message: raw.Message,
		Meta: &mp_common.OcrMeta{
			TraceId:   raw.TraceId,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
		Data: &mp_common.OcrRespData{
			FileInfo: &mp_common.OcrFileInfo{
				FileName: fileName,
				FileType: fileType,
			},
			FullContent: raw.Content,
			OcrResults:  []mp_common.OcrResult{},
		},
		Version:        raw.Version,
		PrefixImageUrl: raw.PrefixImageUrl,
	}
	return target
}
