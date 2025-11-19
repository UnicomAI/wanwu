package mp_qwen

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
	EndpointUrl string `json:"endpointUrl"` // 推理url [EN] inference url
	ContextSize *int   `json:"contextSize"` // 上下文长度 [EN] context length
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
		"input": map[string]interface{}{
			"documents": req.Documents,
			"query":     req.Query,
		},
	}
	if req.TopN != nil || req.ReturnDocuments != nil {
		parameters := make(map[string]interface{})
		if req.TopN != nil {
			parameters["top_n"] = req.TopN
		}
		if req.ReturnDocuments != nil {
			parameters["return_documents"] = req.ReturnDocuments
		}
		m["parameters"] = parameters
	}
	return mp_common.NewRerankReq(m), nil
}

func (cfg *Rerank) Rerank(ctx context.Context, req mp_common.IRerankReq, headers ...mp_common.Header) (mp_common.IRerankResp, error) {
	b, err := mp_common.Rerank(ctx, "qwen", cfg.ApiKey, cfg.rerankUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return &rerankResp{raw: string(b)}, nil
}

func (cfg *Rerank) rerankUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/services/rerank/text-rerank/text-rerank")
	return ret
}

// --- rerankResp ---

type rerankResp struct {
	raw       string
	Output    rerankRespOutput `json:"output" validate:"required"`
	Usage     mp_common.Usage  `json:"usage" validate:"required"`
	RequestId string           `json:"request_id" validate:"required"`
}

type rerankRespOutput struct {
	Results []mp_common.Result `json:"results" validate:"required,dive"`
}

func (resp *rerankResp) String() string {
	return resp.raw
}

func (resp *rerankResp) Data() (interface{}, bool) {
	ret := make(map[string]interface{})
	if err := json.Unmarshal([]byte(resp.raw), &ret); err != nil {
		log.Errorf("qwen rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}
	return ret, true
}

func (resp *rerankResp) ConvertResp() (*mp_common.RerankResp, bool) {
	if err := json.Unmarshal([]byte(resp.raw), resp); err != nil {
		log.Errorf("qwen rerank resp (%v) convert to data err: %v", resp.raw, err)
		return nil, false
	}

	if err := util.Validate(resp); err != nil {
		log.Errorf("qwen rerank resp validate err: %v", err)
		return nil, false
	}
	res := &mp_common.RerankResp{
		Results:   resp.Output.Results,
		Usage:     resp.Usage,
		RequestId: &resp.RequestId,
	}
	return res, true
}
