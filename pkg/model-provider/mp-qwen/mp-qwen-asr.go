package mp_qwen

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
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
	m := map[string]interface{}{
		"model": req.Model,
		"input": map[string]interface{}{
			"messages": []map[string]interface{}{
				{
					"role": "user",
					"content": []map[string]interface{}{
						{
							"audio": audioData,
						},
					},
				},
			},
		},
	}

	return mp_common.NewSyncAsrReq(m), nil
}

func (cfg *SyncAsr) SyncAsr(ctx context.Context, req mp_common.ISyncAsrReq, headers ...mp_common.Header) (mp_common.ISyncAsrResp, error) {
	b, err := mp_common.SyncAsr(ctx, "qwen", cfg.ApiKey, cfg.asrUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return &syncAsrResp{raw: string(b)}, nil
}

func (cfg *SyncAsr) asrUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/services/aigc/multimodal-generation/generation")
	return ret
}

// --- syncAsrResp ---

type syncAsrResp struct {
	raw       string
	Output    syncAsrRespOutput `json:"output" validate:"required"`
	Usage     syncAsrUsage      `json:"usage" validate:"required"`
	RequestId string            `json:"request_id" validate:"required"`
}

type syncAsrRespOutput struct {
	Choices []syncAsrRespOptChoice `json:"choices"`
}

type syncAsrRespOptChoice struct {
	FinishReason string            `json:"finish_reason"`
	Message      syncAsrRespOptMsg `json:"message"`
}

type syncAsrRespOptMsg struct {
	Content []syncAsrRespOptMsgCnt `json:"content"`
	Role    string                 `json:"role"`
}

type syncAsrRespOptMsgCnt struct {
	Text string `json:"text"`
}

type syncAsrUsage struct {
	AudioTokens  int64 `json:"audio_tokens"`
	InputTokens  int64 `json:"input_tokens"`
	OutputTokens int64 `json:"output_tokens"`
	Seconds      int64 `json:"seconds"`
	TotalTokens  int64 `json:"total_tokens"`
}

func (resp *syncAsrResp) String() string {
	return resp.raw
}

func (resp *syncAsrResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("qwen sync_asr resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *syncAsrResp) ConvertResp() (*mp_common.SyncAsrResp, bool) {
	if err := resp.unmarshalRawData(); err != nil {
		return nil, false
	}

	if err := util.Validate(resp); err != nil {
		log.Errorf("qwen sync_asr resp validate err: %v", err)
		return nil, false
	}
	targetResp := resp.buildTargetAsrResp()
	return targetResp, true

}

func (resp *syncAsrResp) unmarshalRawData() error {
	if resp == nil || resp.raw == "" {
		log.Errorf("qwen sync_asr resp raw data is nil or empty")
		return fmt.Errorf("raw data empty")
	}
	if err := json.Unmarshal([]byte(resp.raw), resp); err != nil {
		log.Errorf("qwen sync_asr resp (%v) convert to data err: %v", resp.raw, err)
		return err
	}
	return nil
}

func (resp *syncAsrResp) buildTargetAsrResp() *mp_common.SyncAsrResp {
	// 初始化返回结构体，赋默认值，保底不会nil
	target := &mp_common.SyncAsrResp{
		Code:    0,
		Seconds: resp.Usage.Seconds,
		Choices: make([]mp_common.SyncAsrReqMsgRespChoice, 0),
	}

	if len(resp.Output.Choices) == 0 {
		log.Warnf("qwen sync_asr resp Choices is empty")
		return target
	}
	firstChoice := resp.Output.Choices[0]

	choice := mp_common.SyncAsrReqMsgRespChoice{
		FinishReason: firstChoice.FinishReason,
		Messages:     mp_common.SyncAsrRespMsg{},
	}

	msg := mp_common.SyncAsrRespMsg{
		Role:    mp_common.MsgRole(firstChoice.Message.Role),
		Content: make([]mp_common.SyncAsrRespMsgC, 0),
	}

	if len(firstChoice.Message.Content) > 0 {
		msg.Content = append(msg.Content, mp_common.SyncAsrRespMsgC{
			Text: firstChoice.Message.Content[0].Text,
		})
	} else {
		log.Warnf("qwen sync_asr resp message content is empty")
	}

	choice.Messages = msg
	target.Choices = append(target.Choices, choice)

	return target
}
