package mp_yuanjing

import (
	"context"
	"encoding/json"
	"net/url"

	"github.com/UnicomAI/wanwu/pkg/log"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/UnicomAI/wanwu/pkg/util"
)

type Rerank struct {
	ApiKey      string `json:"apiKey"`      // ApiKey
	EndpointUrl string `json:"endpointUrl"` // inference url
	ContextSize *int   `json:"contextSize"` // context length
}

func (cfg *Rerank) Tags() []mp_common.Tag {
	tags := []mp_common.Tag{
		{
			Text: mp_common.TagRerank,
		},
	}
	tags = append(tags, mp_common.GetTagsByContentSize(cfg.ContextSize)...)
	return tags
}

func (cfg *Rerank) NewReq(req *mp_common.RerankReq) (mp_common.IRerankReq, error) {
	m := map[string]interface{}{
		"model": req.Model,
		"texts": req.Documents,
		"query": req.Query,
	}
	return mp_common.NewRerankReq(m), nil
}

func (cfg *Rerank) Rerank(ctx context.Context, req mp_common.IRerankReq, headers ...mp_common.Header) (mp_common.IRerankResp, error) {
	b, err := mp_common.Rerank(ctx, "yuanjing", cfg.ApiKey, cfg.rerankUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return &rerankResp{raw: string(b)}, nil
}

func (cfg *Rerank) rerankUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/rerank")
	return ret
}

// --- rerankResp ---

type rerankResp struct {
	raw string

	Index    int     `json:"index"`
	Score    float64 `json:"score" validate:"required" `
	Document string  `json:"document" validate:"required"`
}

func (resp *rerankResp) String() string {
	return resp.raw
}

func (resp *rerankResp) Data() (interface{}, bool) {
	ret := []map[string]interface{}{}
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("yuanjing rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *rerankResp) ConvertResp() (*mp_common.RerankResp, bool) {
	var data []map[string]interface{}
	if err := json.Unmarshal([]byte(resp.raw), &data); err != nil {
		log.Errorf("yuanjing rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	var results []mp_common.Result
	for _, item := range data {
		b, err := json.Marshal(item)
		if err != nil {
			log.Errorf("yuanjing rerank resp (%v) item (%v) convert err: %v", resp.raw, item, err)
			return nil, false
		}
		if err = json.Unmarshal(b, resp); err != nil {
			log.Errorf("yuanjing rerank resp (%v) item (%v) unmarshal err: %v", resp.raw, item, err)
			return nil, false
		}

		if err := util.Validate(resp); err != nil {
			log.Errorf("yuanjing rerank resp validate err: %v", err)
			return nil, false
		}

		results = append(results, mp_common.Result{
			Index:          resp.Index,
			RelevanceScore: resp.Score,
			Document: &mp_common.Document{
				Text: resp.Document,
			},
		})
	}

	return &mp_common.RerankResp{
		Results: results,
	}, true
}
