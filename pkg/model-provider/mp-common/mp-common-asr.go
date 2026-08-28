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

const (
	MultiModalTypeText     = "text"
	MultiModalTypeAudio    = "audio"
	MultiModalTypeMinioUrl = "minio_url"
)

// --- openapi request ---

type SyncAsrReq struct {
	Model    string                 `form:"model" json:"model" validate:"required"`
	Messages []SyncAsrReqMsg        `form:"messages" json:"messages" validate:"required"`
	Extra    map[string]interface{} `form:"extra" json:"extra,omitempty"`
}

type SyncAsrReqMsg struct {
	Content []SyncAsrReqC `form:"content" json:"content" validate:"required"`
	Role    string        `form:"content" json:"role" validate:"required"`
}

type SyncAsrReqC struct {
	Type  string       `form:"type" json:"type" validate:"required"`
	Text  string       `form:"text" json:"text,omitempty"`
	Audio SyncAsrAudio `form:"audio" json:"audio,omitempty"`
}

type SyncAsrAudio struct {
	Data     string `form:"data" json:"data,omitempty"`
	FileName string `form:"fileName" json:"fileName,omitempty"`
}

type AsrConfigOut struct {
	Config AsrConfig `form:"config" json:"config" validate:"required"`
}

type AsrConfig struct {
	SessionId           string  `json:"session_id" validate:"required"`
	AddPunc             int     `json:"add_punc,omitempty"`
	ItnSwitch           int     `json:"itn_switch,omitempty"`
	VadSwitch           int     `json:"vad_switch,omitempty"`
	Diarization         int     `json:"diarization,omitempty"`
	SpkNum              int     `json:"spk_num,omitempty"`
	Translate           int     `json:"translate,omitempty"`
	Sensitive           int     `json:"sensitive,omitempty"`
	Language            int     `json:"language,omitempty"`
	AudioClassification int     `json:"audio_classification,omitempty"`
	DiarizationMode     int     `json:"diarization_mode,omitempty"`
	MaxEndSil           int     `json:"max_end_sil,omitempty"`
	MaxSingleSeg        int     `json:"max_single_seg,omitempty"`
	SpeechNoiseThres    float64 `json:"speech_noise_thres,omitempty"`
}

func (req *SyncAsrReq) Check() error {
	return nil
}

func (req *SyncAsrReq) Data() (map[string]interface{}, error) {
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

type SyncAsrResp struct {
	Code    int                       `json:"code"`
	Choices []SyncAsrReqMsgRespChoice `json:"choices"`
	Seconds int64                     `json:"seconds"`
}

type SyncAsrReqMsgRespChoice struct {
	FinishReason string         `json:"finish_reason,omitempty"`
	Messages     SyncAsrRespMsg `json:"message"`
}

type SyncAsrRespMsg struct {
	Content []SyncAsrRespMsgC      `json:"content"`
	Extra   map[string]interface{} `json:"extra,omitempty"`
	Role    MsgRole                `json:"role"`
}

type SyncAsrRespMsgC struct {
	Text             string             `json:"text"`
	SegmentedContent []SegmentedContent `json:"segmented_content"`
}

type SegmentedContent struct {
	StartTime string `json:"start"`
	EndTime   string `json:"end"`
	Text      string `json:"text"`
	Speaker   string `json:"speaker"`
}

// --- request ---

type ISyncAsrReq interface {
	Data() map[string]interface{}
}

// syncAsrReq implementation of ISyncAsrReq
type syncAsrReq struct {
	data map[string]interface{}
}

func NewSyncAsrReq(data map[string]interface{}) ISyncAsrReq {
	return &syncAsrReq{data: data}
}

func (req *syncAsrReq) Data() map[string]interface{} {
	return req.data
}

// --- response ---

type ISyncAsrResp interface {
	String() string
	Data() (interface{}, bool)
	ConvertResp() (*SyncAsrResp, bool)
}

// syncAsrResp implementation of ISyncAsrResp
type syncAsrResp struct {
	raw string
}

func NewSyncAsrResp(raw string) ISyncAsrResp {
	return &syncAsrResp{raw: raw}
}

func (resp *syncAsrResp) String() string {
	return resp.raw
}

func (resp *syncAsrResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("sync_asr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *syncAsrResp) ConvertResp() (*SyncAsrResp, bool) {
	var ret *SyncAsrResp
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("sync_asr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	log.Infof("sync_asr resp: %v", resp.raw)
	if err := util.Validate(ret); err != nil {
		log.Errorf("sync_asr resp validate err: %v", err)
		return nil, false
	}
	return ret, true
}

// --- sync_asr ---

func SyncAsr(ctx context.Context, provider, apiKey, url string, req map[string]interface{}, headers ...Header) ([]byte, error) {
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
