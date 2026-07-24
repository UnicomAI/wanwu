package mp_qianfan

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/minio"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

type Ocr struct {
	ApiKey           string   `json:"apiKey"`           // ApiKey（千帆 Bearer 鉴权用）
	EndpointUrl      string   `json:"endpointUrl"`      // 完整 OCR 地址，如 https://qianfan.baidubce.com/v2/ocr/paddleocr
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
	if req.FileData == nil || *req.FileData == "" {
		return nil, fmt.Errorf("file data is empty")
	}

	// fileType：0=PDF，1=图片；未传时按文件名推断（pdf -> 0，其余 -> 1）
	fileType := 1
	if req.FileType != nil {
		fileType = *req.FileType
	} else if isPdfFileName(req.FileName) {
		fileType = 0
	}

	// 构建千帆 paddleocr 原生请求体
	m := map[string]interface{}{
		"model":               *req.Model,
		"file":                *req.FileData, // base64 data URL 透传
		"fileType":            fileType,
		"useChartRecognition": true, // 图表识别
		"useDocUnwarping":     true, // 文档矫正
		"useLayoutDetection":  true, // 版面检测
		"layoutNms":           true, // 版面 NMS 去重
		"repetitionPenalty":   1.0,  // 重复惩罚
		"temperature":         0,    // 采样温度
		"topP":                1.0,  // top-p 采样
		"visualize":           true, // 输出可视化结果
	}

	// model：使用请求传入的模型标识（如 paddleocr-vl-0.9b）
	if req.Model != nil && *req.Model != "" {
		m["model"] = *req.Model
	}

	// 模型推理参数（上游传入时覆盖默认值）
	if req.UseChartRecognition != nil {
		m["useChartRecognition"] = *req.UseChartRecognition
	}
	if req.UseDocUnwarping != nil {
		m["useDocUnwarping"] = *req.UseDocUnwarping
	}
	if req.UseLayoutDetection != nil {
		m["useLayoutDetection"] = *req.UseLayoutDetection
	}
	if req.LayoutNms != nil {
		m["layoutNms"] = *req.LayoutNms
	}
	if req.RepetitionPenalty != nil {
		m["repetitionPenalty"] = *req.RepetitionPenalty
	}
	if req.Temperature != nil {
		m["temperature"] = *req.Temperature
	}
	if req.TopP != nil {
		m["topP"] = *req.TopP
	}
	if req.MinPixels != nil {
		m["minPixels"] = *req.MinPixels
	}
	if req.MaxPixels != nil {
		m["maxPixels"] = *req.MaxPixels
	}
	if req.Visualize != nil {
		m["visualize"] = *req.Visualize
	}

	return mp_common.NewOcrReq(m), nil
}

func (cfg *Ocr) Ocr(ctx *gin.Context, req mp_common.IOcrReq, headers ...mp_common.Header) (mp_common.IOcrResp, error) {
	b, err := mp_common.Ocr(ctx, "qianfan", cfg.ApiKey, cfg.ocrUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	// 将 markdown.text 中的图片 key 转存 minio 并替换为 minio_url；失败回退原始响应
	if processed, ok := processOcrImages(ctx.Request.Context(), b); ok {
		b = processed
	}
	return &ocrResp{raw: string(b)}, nil
}

// ocrUrl 返回下游请求地址（千帆原生接口直接使用 endpointUrl，鉴权通过 Authorization Header）
func (cfg *Ocr) ocrUrl() string {
	if u, err := url.Parse(cfg.EndpointUrl); err == nil && u != nil {
		return u.String()
	}
	return cfg.EndpointUrl
}

func isPdfFileName(fileName string) bool {
	return strings.HasSuffix(strings.ToLower(fileName), ".pdf")
}

// --- 文档内图片转存 minio ---

// imgSrcRe 匹配 HTML 中 src 引用，兼容单/双引号，不限定图片 key 前缀。
// group1=引号类型，group2=src 值（图片 key 或绝对 URL）。
// 注意：Go regexp（RE2）不支持反向引用，结尾用字符类 ["'] 匹配任一引号，
// 不强制首尾配对；实际 HTML 不会有 src="x' 这种畸形写法，且仅命中 images map 才替换，不会误伤。
var imgSrcRe = regexp.MustCompile(`src=(["'])([^"'\s]+)["']`)

// processOcrImages 解析千帆原生响应，将各页 markdown.text 中的图片 key 替换为转存 minio 后的 URL，返回新的 JSON 字节。
// 仅在成功响应（ErrorCode==0 且 Result!=nil）时处理；任何阶段失败（unmarshal/marshal）返回 (nil,false)，由调用方回退原始响应。
func processOcrImages(ctx context.Context, raw []byte) ([]byte, bool) {
	// minio 未初始化时跳过，避免 panic（正常运行时必非 nil，此处为防御）
	if minio.FileUpload() == nil {
		log.Warnf("qianfan ocr process images skipped: minio fileupload client not init")
		return nil, false
	}
	var resp ocrRawResp
	if err := json.Unmarshal(raw, &resp); err != nil {
		log.Errorf("qianfan ocr process images unmarshal err: %v", err)
		return nil, false
	}
	if resp.ErrorCode != 0 || resp.Result == nil {
		return nil, false
	}
	// 跨页共享图片缓存：同一 key 多处引用只下载转存一次
	urlCache := make(map[string]string)
	for i := range resp.Result.LayoutParsingResults {
		md := &resp.Result.LayoutParsingResults[i].Markdown
		if len(md.Images) == 0 || strings.TrimSpace(md.Text) == "" {
			continue
		}
		md.Text = replaceImageKeys(ctx, md.Text, md.Images, urlCache)
	}
	out, err := json.Marshal(&resp)
	if err != nil {
		log.Errorf("qianfan ocr process images marshal err: %v", err)
		return nil, false
	}
	return out, true
}

// replaceImageKeys 将 text 中 src 引用的图片 key 替换为 minio_url。
// 仅当 src 值命中 images map 时才替换（已是绝对 URL 的 src 自然跳过）；
// 单张图片下载/转存失败仅记录日志并保留原 src，不影响整体。
func replaceImageKeys(ctx context.Context, text string, images map[string]string, urlCache map[string]string) string {
	return imgSrcRe.ReplaceAllStringFunc(text, func(match string) string {
		sub := imgSrcRe.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		quote, srcValue := sub[1], sub[2]
		qianfanUrl, ok := images[srcValue]
		if !ok {
			// 非图片 key（如已是绝对 URL），保留原样
			return match
		}
		// 命中缓存直接复用
		if minioUrl, ok := urlCache[srcValue]; ok {
			return "src=" + quote + minioUrl + quote
		}
		minioUrl, err := downloadAndUploadImage(ctx, qianfanUrl)
		if err != nil {
			log.Errorf("qianfan ocr replace image (%v -> %v) err: %v", srcValue, qianfanUrl, err)
			return match
		}
		urlCache[srcValue] = minioUrl
		return "src=" + quote + minioUrl + quote
	})
}

// downloadAndUploadImage 下载千帆公网图片字节并转存到 minio，返回 minio 可访问 URL。
func downloadAndUploadImage(ctx context.Context, imgUrl string) (string, error) {
	resp, err := http_client.Default().GetOriResp(ctx, &http_client.HttpRequestParams{
		Url:        imgUrl,
		Timeout:    2 * time.Minute,
		MonitorKey: "qianfan_ocr_image_download",
		LogLevel:   http_client.LogBasic,
	})
	if err != nil {
		return "", fmt.Errorf("download (%v) err: %v", imgUrl, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download (%v) status: %v", imgUrl, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read (%v) body err: %v", imgUrl, err)
	}
	if len(data) == 0 {
		return "", fmt.Errorf("download (%v) empty body", imgUrl)
	}
	// 扩展名优先用响应 content-type，兜底 .png
	ext := extFromContentType(resp.Header.Get("Content-Type"))
	fileName := util.GenUUID() + ext
	minioUrl, _, err := minio.UploadFile(ctx, minio.BucketFileUpload, minio.DirFileNotExpire, fileName, bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("upload to minio err: %v", err)
	}
	return minioUrl, nil
}

// extFromContentType 根据响应 content-type 推断图片扩展名，无法识别时兜底 .png
func extFromContentType(ct string) string {
	switch {
	case strings.Contains(ct, "image/png"):
		return ".png"
	case strings.Contains(ct, "image/jpeg"), strings.Contains(ct, "image/jpg"):
		return ".jpg"
	case strings.Contains(ct, "image/gif"):
		return ".gif"
	case strings.Contains(ct, "image/webp"):
		return ".webp"
	default:
		return ".png"
	}
}

// --- ocrResp 下游原生响应结构体，实现 mp_common.IOcrResp 接口 ---

// ocrRawResp 千帆 paddleocr 原生响应结构
type ocrRawResp struct {
	ID        string        `json:"id"` // 请求 id（如 as-xxxx）
	ErrorCode int           `json:"error_code,omitempty"`
	ErrorMsg  string        `json:"error_msg,omitempty"`
	Result    *ocrRawResult `json:"result,omitempty"`
}

// ocrRawResult 千帆 paddleocr result 结构
type ocrRawResult struct {
	LayoutParsingResults []ocrLayoutParsingResult `json:"layoutParsingResults"`
	DataInfo             *ocrDataInfo             `json:"dataInfo"`
}

// ocrLayoutParsingResult 单页/单文件版面解析结果
type ocrLayoutParsingResult struct {
	PrunedResult ocrPrunedResult `json:"prunedResult"`
	Markdown     ocrMarkdown     `json:"markdown"`
}

// ocrPrunedResult 裁剪后的解析结果
type ocrPrunedResult struct {
	ParsingResList []ocrParsingRes `json:"parsing_res_list"`
}

// ocrParsingRes 单个版面区块
type ocrParsingRes struct {
	BlockLabel   string `json:"block_label"`   // 区块类型：title/text/image/table 等
	BlockContent string `json:"block_content"` // 区块文本内容
	BlockBbox    []int  `json:"block_bbox"`    // 区块坐标 [x1,y1,x2,y2]
	BlockID      int    `json:"block_id"`      // 区块 id
}

// ocrMarkdown markdown 结构化结果
type ocrMarkdown struct {
	Text   string            `json:"text"`   // markdown 全文
	Images map[string]string `json:"images"` // 图片名 -> 图片 URL
}

// ocrDataInfo 输入文件信息
type ocrDataInfo struct {
	Type   string `json:"type"` // image / pdf
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type ocrResp struct {
	raw string
}

func (resp *ocrResp) String() string {
	return resp.raw
}

func (resp *ocrResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("qianfan ocr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *ocrResp) ConvertResp() (*mp_common.OcrResp, bool) {
	var raw ocrRawResp
	if err := json.Unmarshal([]byte(resp.raw), &raw); err != nil {
		log.Errorf("qianfan ocr resp (%v) unmarshal err: %v", resp.raw, err)
		return nil, false
	}

	if raw.ErrorCode != 0 {
		log.Errorf("qianfan ocr resp error_code %v, error_msg: %v", raw.ErrorCode, raw.ErrorMsg)
		return nil, false
	}

	// 将下游原生响应转换为统一 OcrResp 格式
	target := resp.buildTargetOcrResp(&raw)
	return target, true
}

func (resp *ocrResp) buildTargetOcrResp(raw *ocrRawResp) *mp_common.OcrResp {
	code := 0 // 成功（ErrorCode != 0 已在 ConvertResp 提前返回）
	message := "文档处理完成"

	// 拼接全文：优先使用 markdown.text，回退拼接 parsing_res_list 的 block_content
	fullContent := ""
	ocrResults := make([]mp_common.OcrResult, 0)
	totalPages := int64(0)
	fileType := ""

	if raw.Result != nil {
		for i, item := range raw.Result.LayoutParsingResults {
			pageNumber := i + 1
			totalPages++

			// markdown.text 作为该页全文
			if strings.TrimSpace(item.Markdown.Text) != "" {
				if fullContent != "" {
					fullContent += "\n"
				}
				fullContent += item.Markdown.Text
			}

			// parsing_res_list 转为分区块结构化结果
			for _, block := range item.PrunedResult.ParsingResList {
				// 跳过无文本内容的区块（如纯图片）
				if strings.TrimSpace(block.BlockContent) == "" {
					continue
				}
				ocrResults = append(ocrResults, mp_common.OcrResult{
					PageNumber: pageNumber,
					Type:       block.BlockLabel,
					Content:    block.BlockContent,
				})
			}
		}

		// dataInfo.type 优先用于推断文件类型
		if raw.Result.DataInfo != nil && raw.Result.DataInfo.Type != "" {
			fileType = raw.Result.DataInfo.Type
		}
	}

	// 若未从 markdown 取到全文，回退拼接 ocrResults 内容
	if fullContent == "" && len(ocrResults) > 0 {
		var lines []string
		for _, r := range ocrResults {
			if r.Content != "" {
				lines = append(lines, r.Content)
			}
		}
		fullContent = strings.Join(lines, "\n")
	}

	target := &mp_common.OcrResp{
		Code:    code,
		Message: message,
		Meta: &mp_common.OcrMeta{
			RequestId: raw.ID,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		},
		Data: &mp_common.OcrRespData{
			FileInfo: &mp_common.OcrFileInfo{
				FileType:   fileType,
				TotalPages: totalPages,
			},
			FullContent: fullContent,
			OcrResults:  ocrResults,
		},
	}
	return target
}
