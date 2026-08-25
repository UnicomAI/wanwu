package mp_huoshan

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/util"
)

// MultiModalEmbedding 火山多模态 embedding 配置（字段对齐 yuanjing/qwen）
type MultiModalEmbedding struct {
	ApiKey           string   `json:"apiKey"`
	EndpointUrl      string   `json:"endpointUrl"`
	ContextSize      *int     `json:"contextSize"`
	MaxTextLength    *int64   `json:"maxTextLength"`
	MaxImageSize     *int64   `json:"maxImageSize"`
	MaxVideoClipSize *int64   `json:"maxVideoClipSize"`
	SupportFileTypes []string `json:"supportFileTypes"`
}

func (cfg *MultiModalEmbedding) Tags() []mp_common.Tag {
	tags := []mp_common.Tag{
		{
			Text: mp_common.TagMultiModalEmbedding,
		},
	}
	tags = append(tags, mp_common.GetTagsByContentSize(cfg.ContextSize)...)
	return tags
}

// --- 请求转换 ---
// 火山 input 项格式: {"type":"text","text":"..."} 或 {"type":"image_url","image_url":{"url":"..."}}
// BFF 层已把 item.Image（minio url）转成 data URI base64 字符串，直接放进 image_url.url。
func (cfg *MultiModalEmbedding) NewReq(req *mp_common.MultiModalEmbeddingReq) (mp_common.IMultiModalEmbeddingReq, error) {
	input := make([]map[string]interface{}, 0, len(req.Input))
	for _, item := range req.Input {
		if item.Text != nil && *item.Text != "" {
			input = append(input, map[string]interface{}{
				"type": "text",
				"text": *item.Text,
			})
		}
		if item.Image != nil && *item.Image != "" {
			input = append(input, map[string]interface{}{
				"type":      "image_url",
				"image_url": map[string]string{"url": *item.Image},
			})
		}
		// 火山 doubao-embedding-vision 当前支持 text + image；
		// audio/video 暂不映射（火山多模态 embedding 无对应类型）。
	}
	m := map[string]interface{}{
		"model": req.Model,
		"input": input,
	}
	if req.EncodingFormat != "" {
		m["encoding_format"] = req.EncodingFormat
	}
	if req.Dimensions != nil {
		m["dimensions"] = *req.Dimensions
	}
	if req.Parameters != nil {
		m["parameters"] = req.Parameters
	}
	return mp_common.NewMultiModalEmbeddingReq(m), nil
}

func (cfg *MultiModalEmbedding) MultiModalEmbeddings(ctx context.Context, req mp_common.IMultiModalEmbeddingReq, headers ...mp_common.Header) (mp_common.IMultiModalEmbeddingResp, error) {
	b, err := mp_common.MultiModalEmbeddings(ctx, "huoshan", cfg.ApiKey, cfg.embeddingsUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return &multiModalEmbeddingResp{raw: string(b)}, nil
}

// 火山多模态 embedding 路径是 /embeddings/multimodal（非 /embeddings）
func (cfg *MultiModalEmbedding) embeddingsUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/embeddings/multimodal")
	return ret
}

// --- 响应转换 ---
// 火山响应 data 是单对象 {"embedding":[...],"object":"embedding"}；转成 common 单元素 []EmbeddingData(index=0)
// 注意：字段名不能为 Data，否则与 IMultiModalEmbeddingResp.Data() 方法同名冲突。
type multiModalEmbeddingResp struct {
	raw     string               `json:"-"`
	Id      string               `json:"id"`
	Model   string               `json:"model"`
	Object  string               `json:"object"`
	Created int                  `json:"created"`
	Embed   huoshanEmbeddingData `json:"data"`
	Usage   mp_common.Usage      `json:"usage"`
}

// huoshanEmbeddingData 火山响应的 data 单对象
type huoshanEmbeddingData struct {
	Embedding []float64 `json:"embedding"`
	Object    string    `json:"object"`
}

func (resp *multiModalEmbeddingResp) String() string {
	return resp.raw
}

func (resp *multiModalEmbeddingResp) Data() (map[string]interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("huoshan multimodal embedding resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *multiModalEmbeddingResp) ConvertResp() (*mp_common.MultiModalEmbeddingResp, bool) {
	if err := json.Unmarshal([]byte(resp.raw), resp); err != nil {
		log.Errorf("huoshan multimodal embedding resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	if err := util.Validate(resp); err != nil {
		log.Errorf("huoshan multimodal embedding resp validate err: %v", err)
		return nil, false
	}
	dataObj := resp.Embed.Object
	data := []mp_common.EmbeddingData{
		{
			Object:    &dataObj,
			Embedding: resp.Embed.Embedding,
			Index:     0,
		},
	}
	res := &mp_common.MultiModalEmbeddingResp{
		Model:  resp.Model,
		Data:   data,
		Usage:  resp.Usage,
		Object: &resp.Object,
	}
	if resp.Id != "" {
		res.Id = &resp.Id
	}
	if resp.Created != 0 {
		res.Created = &resp.Created
	}
	return res, true
}
