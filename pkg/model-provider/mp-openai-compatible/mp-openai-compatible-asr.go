package mp_openai_compatible

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

type SyncAsr struct {
	ApiKey         string `json:"apiKey"`      // ApiKey
	EndpointUrl    string `json:"endpointUrl"` // 推理url（完整地址，含 /v1/chat/completions）
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
				break
			}
		}
	}
	// audioData 已由 bff handler 转换为 data URL 格式 (data:audio/wav;base64,<base64>)，
	// 上游 audio_url.url 直接接收该 data URL。
	m := map[string]interface{}{
		"model": req.Model,
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "audio_url",
						"audio_url": map[string]interface{}{
							"url": audioData,
						},
					},
				},
			},
		},
	}

	return mp_common.NewSyncAsrReq(m), nil
}

func (cfg *SyncAsr) SyncAsr(ctx context.Context, req mp_common.ISyncAsrReq, headers ...mp_common.Header) (mp_common.ISyncAsrResp, error) {
	b, err := mp_common.SyncAsr(ctx, "openai compatible", cfg.ApiKey, cfg.asrUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return &syncAsrResp{raw: string(b)}, nil
}

func (cfg *SyncAsr) asrUrl() string {
	// endpointUrl 配置为完整地址，不再拼接后缀
	ret, _ := url.JoinPath(cfg.EndpointUrl, "")
	return ret
}

// --- syncAsrResp ---

type syncAsrResp struct {
	raw     string
	ID      string          `json:"id"`
	Object  string          `json:"object"`
	Created int64           `json:"created"`
	Model   string          `json:"model"`
	Choices []syncAsrChoice `json:"choices"`
	Usage   syncAsrUsage    `json:"usage"`
}

type syncAsrChoice struct {
	Index        int            `json:"index"`
	Message      syncAsrMessage `json:"message"`
	FinishReason string         `json:"finish_reason"`
}

type syncAsrMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type syncAsrUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func (resp *syncAsrResp) String() string {
	return resp.raw
}

func (resp *syncAsrResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("openai compatible sync_asr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *syncAsrResp) ConvertResp() (*mp_common.SyncAsrResp, bool) {
	if err := resp.unmarshalRawData(); err != nil {
		return nil, false
	}

	targetResp := &mp_common.SyncAsrResp{
		Code:    0,
		Seconds: 0,
		Choices: make([]mp_common.SyncAsrReqMsgRespChoice, 0),
	}

	if len(resp.Choices) == 0 {
		log.Warnf("openai compatible sync_asr resp Choices is empty")
		return targetResp, true
	}
	firstChoice := resp.Choices[0]

	choice := mp_common.SyncAsrReqMsgRespChoice{
		FinishReason: firstChoice.FinishReason,
		Messages: mp_common.SyncAsrRespMsg{
			Role:    mp_common.MsgRole(firstChoice.Message.Role),
			Content: make([]mp_common.SyncAsrRespMsgC, 0),
		},
	}

	// content 透传，不做清洗
	choice.Messages.Content = append(choice.Messages.Content, mp_common.SyncAsrRespMsgC{
		Text: firstChoice.Message.Content,
	})

	targetResp.Choices = append(targetResp.Choices, choice)
	return targetResp, true
}

func (resp *syncAsrResp) unmarshalRawData() error {
	if resp == nil || resp.raw == "" {
		log.Errorf("openai compatible sync_asr resp raw data is nil or empty")
		return fmt.Errorf("raw data empty")
	}
	if err := json.Unmarshal([]byte(resp.raw), resp); err != nil {
		log.Errorf("openai compatible sync_asr resp (%v) convert to data err: %v", resp.raw, err)
		return err
	}
	return nil
}
