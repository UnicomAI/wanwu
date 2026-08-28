package mp_huoshan

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/url"

	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/UnicomAI/wanwu/pkg/util"
)

// 需要额外修改config的配置
type SyncAsr struct {
	AppKey         string `json:"appKey"` // AppKey
	AccessKey      string `json:"accessKey"`
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
	var audioData string
	if len(req.Messages) > 0 {
		msg := req.Messages[0]
		for _, content := range msg.Content {
			if content.Type == mp_common.MultiModalTypeAudio || content.Type == mp_common.MultiModalTypeMinioUrl {
				audioData = content.Audio.Data
				if content.Type == mp_common.MultiModalTypeMinioUrl {
					audioData, _ = util.CheckAndRemoveBase64Prefix(content.Audio.Data)
				}
				break
			}
		}
	}
	m := map[string]interface{}{
		"model": req.Model,
		"audio": map[string]interface{}{
			"data": audioData,
		},
	}

	return mp_common.NewSyncAsrReq(m), nil
}

func (cfg *SyncAsr) SyncAsr(ctx context.Context, req mp_common.ISyncAsrReq, headers ...mp_common.Header) (mp_common.ISyncAsrResp, error) {
	cfgHeaders := []mp_common.Header{
		{Key: "X-Api-App-Key", Value: cfg.AppKey},
		{Key: "X-Api-Access-Key", Value: cfg.AccessKey},
		{Key: "X-Api-Resource-Id", Value: "volc.bigasr.auc_turbo"}, // 固定值
		{Key: "X-Api-Request-Id", Value: util.GenUUID()},
		{Key: "X-Api-Sequence", Value: "-1"},
	}
	b, err := syncAsr(ctx, "huoshan", "", cfg.asrUrl(), req.Data(), cfgHeaders...)
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
	if apiKey != "" {
		headers = append(headers, mp_common.Header{
			Key:   "Authorization",
			Value: "Bearer " + apiKey,
		})
	}
	reqjson, _ := json.Marshal(req)
	log.Debugf("huoshan sync_asr req: %v", string(reqjson))
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
		return nil, fmt.Errorf("request %v %v sync_asr err: %v", url, provider, err)
	}
	b, err := io.ReadAll(resp.RawResponse.Body)
	if err != nil {
		return nil, fmt.Errorf("request %v %v sync_asr read response body failed: %v", url, provider, err)
	}
	if resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("request %v %v sync_asr http status %v msg: %v", url, provider, resp.StatusCode(), string(b))
	}
	return b, nil
}

// --- syncAsrResp ---

// 根结构体，对应新的返回体
type syncAsrResp struct {
	raw       string
	AudioInfo syncAsrAudioInfo `json:"audio_info" validate:"required"`
	Result    syncAsrResult    `json:"result" validate:"required"`
}

// 音频信息结构体
type syncAsrAudioInfo struct {
	Duration int `json:"duration"` // 音频时长（毫秒）
}

// 识别结果主结构体
type syncAsrResult struct {
	Additions  syncAsrAdditions   `json:"additions" validate:"required"`
	Text       string             `json:"text" validate:"required"`       // 完整识别文本
	Utterances []syncAsrUtterance `json:"utterances" validate:"required"` // 分句识别结果
}

// 附加信息结构体
type syncAsrAdditions struct {
	Duration string `json:"duration" validate:"required"` // 时长（字符串格式）
}

// 单句识别结果结构体
type syncAsrUtterance struct {
	EndTime   int           `json:"end_time" validate:"required"`   // 结束时间（毫秒）
	StartTime int           `json:"start_time" validate:"required"` // 开始时间（毫秒）
	Text      string        `json:"text" validate:"required"`       // 句子文本
	Words     []syncAsrWord `json:"words" validate:"required"`      // 逐字识别结果
}

// 逐字识别结果结构体
type syncAsrWord struct {
	Confidence int    `json:"confidence" validate:"required"` // 置信度
	EndTime    int    `json:"end_time" validate:"required"`   // 结束时间（毫秒）
	StartTime  int    `json:"start_time" validate:"required"` // 开始时间（毫秒）
	Text       string `json:"text" validate:"required"`       // 单字文本
}

func (resp *syncAsrResp) String() string {
	return resp.raw
}

func (resp *syncAsrResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("huoshan sync_asr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *syncAsrResp) ConvertResp() (*mp_common.SyncAsrResp, bool) {
	if err := resp.unmarshalRawData(); err != nil {
		return nil, false
	}

	if err := util.Validate(resp); err != nil {
		log.Errorf("huoshan sync_asr resp validate err: %v", err)
		return nil, false
	}
	targetResp := resp.buildTargetAsrResp()
	return targetResp, true

}

func (resp *syncAsrResp) unmarshalRawData() error {
	if resp == nil || resp.raw == "" {
		log.Errorf("huoshan sync_asr resp raw data is nil or empty")
		return fmt.Errorf("raw data empty")
	}
	if err := json.Unmarshal([]byte(resp.raw), resp); err != nil {
		log.Errorf("huoshan sync_asr resp (%v) convert to data err: %v", resp.raw, err)
		return err
	}
	return nil
}

func (resp *syncAsrResp) buildTargetAsrResp() *mp_common.SyncAsrResp {
	// 初始化目标结构体
	targetResp := &mp_common.SyncAsrResp{
		Code:    0,
		Seconds: int64(resp.AudioInfo.Duration) / 1000, // 将毫秒转换为秒
	}

	// 构建 Choices 数组（ASR识别结果通常只有一个choice）
	choices := make([]mp_common.SyncAsrReqMsgRespChoice, 0, 1)
	choice := mp_common.SyncAsrReqMsgRespChoice{
		FinishReason: "stop", // 识别完成的原因，通常固定为 "stop"
		Messages: mp_common.SyncAsrRespMsg{
			Role: mp_common.MsgRoleAssistant, // 设置角色，根据实际业务调整
		},
	}

	// 构建 Content 数组（核心识别文本和分段信息）
	content := mp_common.SyncAsrRespMsgC{
		Text: resp.Result.Text, // 完整的识别文本
	}

	// 构建分段内容 SegmentedContent
	segmentedContent := make([]mp_common.SegmentedContent, 0)
	for _, utterance := range resp.Result.Utterances {
		// 遍历每一个分句，转换为 SegmentedContent
		segment := mp_common.SegmentedContent{
			StartTime: fmt.Sprintf("%d", utterance.StartTime), // 转换为字符串格式
			EndTime:   fmt.Sprintf("%d", utterance.EndTime),
			Text:      utterance.Text,
			Speaker:   "", // 如果原结构体无说话人信息，默认空字符串
		}
		segmentedContent = append(segmentedContent, segment)
	}

	content.SegmentedContent = segmentedContent
	choice.Messages.Content = []mp_common.SyncAsrRespMsgC{content}
	choices = append(choices, choice)

	targetResp.Choices = choices

	return targetResp
}
