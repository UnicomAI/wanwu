package mp_yuanjing

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strings"

	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type SyncAsr struct {
	ApiKey         string `json:"apiKey"`      // ApiKey
	EndpointUrl    string `json:"endpointUrl"` // 推理url
	MaxAsrFileSize *int64 `json:"maxAsrFileSize"`
}

func (cfg *SyncAsr) Tags() []mp_common.Tag {
	tags := []mp_common.Tag{
		{
			Text: mp_common.TagSyncAsr,
		},
	}
	return tags
}

func (cfg *SyncAsr) NewReq(req *mp_common.SyncAsrReq) (mp_common.ISyncAsrReq, error) {
	var targetContent mp_common.SyncAsrReqC
	msg := req.Messages[0]
	for _, content := range msg.Content {
		if content.Type == mp_common.MultiModalTypeMinioUrl || content.Type == mp_common.MultiModalTypeAudio {
			targetContent = content
			break
		}
	}

	if targetContent.Audio.Data == "" {
		return nil, fmt.Errorf("sync_asr 未找到有效音频地址")
	}

	var b map[string]interface{}
	if targetContent.Type == mp_common.MultiModalTypeMinioUrl {
		parts := strings.SplitN(targetContent.Audio.Data, ",", 2)
		if len(parts) < 2 {
			return nil, fmt.Errorf("sync_asr 音频地址格式错误：无分隔符,，data=%s", targetContent.Audio.Data)
		}
		fileBase64 := parts[1]
		fileDataFromBase64, err := base64.StdEncoding.DecodeString(fileBase64)
		if err != nil {
			return nil, fmt.Errorf("filebase64 fileData decode err: %v", err)
		}
		fileHeader, err := util.FileData2FileHeader(targetContent.Audio.FileName, fileDataFromBase64)
		if err != nil {
			return nil, fmt.Errorf("filedata %s to multipart err: %v", targetContent.Audio.FileName, err)
		}
		b = map[string]interface{}{
			"file": fileHeader,
			"config": map[string]interface{}{
				"config": map[string]interface{}{"sessionId": util.GenUUID()},
			},
		}
	} else {
		b = map[string]interface{}{
			"config": map[string]interface{}{
				"config": map[string]interface{}{"url": targetContent.Audio.Data},
			},
		}
	}

	return mp_common.NewSyncAsrReq(b), nil
}

func (cfg *SyncAsr) SyncAsr(ctx context.Context, req mp_common.ISyncAsrReq, headers ...mp_common.Header) (mp_common.ISyncAsrResp, error) {
	b, err := syncAsr(ctx, "yuanjing", cfg.ApiKey, cfg.asrUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return &syncAsrResp{raw: string(b)}, nil
}

func (cfg *SyncAsr) asrUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "")
	return ret
}

func syncAsr(ctx context.Context, provider, apiKey, url string, req map[string]interface{}, headers ...mp_common.Header) ([]byte, error) {
	if _, ok := req["file"]; ok {
		return syncAsrFormData(ctx, "yuanjing", apiKey, url, req, headers...)
	}
	return mp_common.SyncAsr(ctx, "yuanjing", apiKey, url, req, headers...)
}
func syncAsrFormData(ctx context.Context, provider, apiKey, url string, req map[string]interface{}, headers ...mp_common.Header) ([]byte, error) {
	if apiKey != "" {
		headers = append(headers, mp_common.Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}

	fileHeader, ok := req["file"].(*multipart.FileHeader)
	if !ok {
		return nil, fmt.Errorf("req中file字段类型错误，期望*multipart.FileHeader")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("request %v %v sync_asr err: %v", url, provider, err)
	}
	defer func() { _ = file.Close() }()

	configVal, ok := req["config"]
	if !ok {
		return nil, fmt.Errorf("req中缺少config字段")
	}
	configJSON, err := json.Marshal(configVal)
	if err != nil {
		return nil, fmt.Errorf("marshal config failed: %v", err)
	}

	request := trace_util.NewResty(ctx).
		SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true}). // 关闭证书校验
		SetTimeout(0).                                             // 关闭请求超时
		R().
		SetContext(ctx).
		SetHeader("Content-Type", "multipart/form-data").
		SetHeader("Accept", "application/json").
		SetFileReader("file", fileHeader.Filename, file). // 使用直接提取的fileHeader
		SetMultipartField("config", "", "application/json", strings.NewReader(string(configJSON))).
		SetDoNotParseResponse(true)
	for _, header := range headers {
		request.SetHeader(header.Key, header.Value)
	}
	resp, err := request.Post(url)
	if err != nil {
		return nil, fmt.Errorf("request %v %v sync_asr err: %v", url, provider, err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v sync_asr read response body err: %v", url, provider, err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v sync_asr http status %v msg: %v", url, provider, resp.StatusCode(), string(b))
	}

	return b, nil
}

type syncAsrResp struct {
	raw     string
	Status  string        `json:"status"`
	Code    int           `json:"code"`
	Message string        `json:"msg"`
	Uuid    string        `json:"uuid"`
	Result  SyncAsrResult `json:"result"`
}

type SyncAsrResult struct {
	Diarization []DiarizationObj `json:"diarization"`
}

type DiarizationObj struct {
	Start   float32 `json:"start"`
	End     float32 `json:"end"`
	Speaker int     `json:"speaker"`
	Text    string  `json:"text"`
	Trans   string  `json:"trans"`
}

func (resp *syncAsrResp) String() string {
	return resp.raw
}

func (resp *syncAsrResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("yuanjing sync_asr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *syncAsrResp) ConvertResp() (*mp_common.SyncAsrResp, bool) {
	if err := json.Unmarshal([]byte(resp.raw), &resp); err != nil {
		log.Errorf("asr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	if err := util.Validate(resp); err != nil {
		log.Errorf("qwen sync_asr resp validate err: %v", err)
		return nil, false
	}
	targetResp := resp.buildTargetAsrResp()
	return targetResp, true

}

// buildTargetAsrResp 将原生syncAsrResp转换为mp_common.SyncAsrResp
func (resp *syncAsrResp) buildTargetAsrResp() *mp_common.SyncAsrResp {
	if resp == nil {
		return &mp_common.SyncAsrResp{}
	}

	targetResp := &mp_common.SyncAsrResp{
		Code:    resp.Code,
		Seconds: 0,
		Choices: make([]mp_common.SyncAsrReqMsgRespChoice, 0, 1),
	}

	// 处理Seconds：取diarization最后一个元素的end值
	diarLen := len(resp.Result.Diarization)
	if diarLen > 0 {
		lastEnd := resp.Result.Diarization[diarLen-1].End
		targetResp.Seconds = int64(lastEnd)
	}

	choice := mp_common.SyncAsrReqMsgRespChoice{
		Messages: mp_common.SyncAsrRespMsg{
			Role:    mp_common.MsgRoleAssistant,
			Extra:   make(map[string]interface{}),
			Content: make([]mp_common.SyncAsrRespMsgC, 0, 1),
		},
	}

	// 填充Extra字段（可拓展）
	//choice.Messages.Extra["uuid"] = resp.Uuid

	content := mp_common.SyncAsrRespMsgC{
		SegmentedContent: make([]mp_common.SegmentedContent, 0, len(resp.Result.Diarization)),
	}

	// 拼接整体文本（所有分句的Text合并）
	var fullText strings.Builder
	for _, diar := range resp.Result.Diarization {
		fullText.WriteString(diar.Text)
		fullText.WriteString("") // 分句间加分隔（可拓展）

		startTime := fmt.Sprintf("%.2f", diar.Start)
		endTime := fmt.Sprintf("%.2f", diar.End)
		speaker := fmt.Sprintf("%d", diar.Speaker)

		// 构建单条分段内容
		segment := mp_common.SegmentedContent{
			StartTime: startTime,
			EndTime:   endTime,
			Text:      diar.Text,
			Speaker:   speaker,
		}
		content.SegmentedContent = append(content.SegmentedContent, segment)
	}

	// 去除整体文本末尾多余的空格
	content.Text = strings.TrimSpace(fullText.String())

	choice.Messages.Content = append(choice.Messages.Content, content)

	targetResp.Choices = append(targetResp.Choices, choice)

	return targetResp
}
