package mp_yuanjing

import (
	"net/url"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
	"github.com/gin-gonic/gin"
)

type PdfParser struct {
	ApiKey      string `json:"apiKey"`      // ApiKey
	EndpointUrl string `json:"endpointUrl"` // inference url
}

func (cfg *PdfParser) Tags() []mp_common.Tag {
	tags := []mp_common.Tag{
		{
			Text: mp_common.TagPdfParser,
		},
	}
	return tags
}
func (cfg *PdfParser) NewReq(req *mp_common.PdfParserReq) (mp_common.IPdfParserReq, error) {
	return mp_common.NewPdfParserReq(req), nil
}

func (cfg *PdfParser) PdfParser(ctx *gin.Context, req mp_common.IPdfParserReq, headers ...mp_common.Header) (mp_common.IPdfParserResp, error) {
	b, err := mp_common.PdfParser(ctx, "yuanjing", cfg.ApiKey, cfg.pdfParserUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return mp_common.NewPdfParserResp(string(b)), nil
}

func (cfg *PdfParser) pdfParserUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/rag/model_parser_file")
	return ret
}
