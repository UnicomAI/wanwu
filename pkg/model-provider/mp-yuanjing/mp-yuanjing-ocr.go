package mp_yuanjing

import (
	"net/url"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/gin-gonic/gin"
)

type Ocr struct {
	ApiKey      string `json:"apiKey"`      // ApiKey
	EndpointUrl string `json:"endpointUrl"` // inference url
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
	return mp_common.NewOcrReq(req), nil
}

func (cfg *Ocr) Ocr(ctx *gin.Context, req mp_common.IOcrReq, headers ...mp_common.Header) (mp_common.IOcrResp, error) {
	b, err := mp_common.Ocr(ctx, "yuanjing", cfg.ApiKey, cfg.ocrUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return mp_common.NewOcrResp(string(b)), nil
}

func (cfg *Ocr) ocrUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/unicom-ocr")
	return ret
}
