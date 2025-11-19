package mp_yuanjing

import (
	"context"
	"net/url"

	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

type Gui struct {
	ApiKey      string `json:"apiKey"`      // ApiKey
	EndpointUrl string `json:"endpointUrl"` // inference url
}

func (cfg *Gui) Tags() []mp_common.Tag {
	tags := []mp_common.Tag{
		{
			Text: mp_common.TagGui,
		},
	}
	return tags
}

func (cfg *Gui) NewReq(req *mp_common.GuiReq) (mp_common.IGuiReq, error) {
	return mp_common.NewGuiReq(req), nil
}

func (cfg *Gui) Gui(ctx context.Context, req mp_common.IGuiReq, headers ...mp_common.Header) (mp_common.IGuiResp, error) {
	b, err := mp_common.Gui(ctx, "yuanjing", cfg.ApiKey, cfg.guiUrl(), req.Data(), headers...)
	if err != nil {
		return nil, err
	}
	return mp_common.NewGuiResp(string(b)), nil
}

func (cfg *Gui) guiUrl() string {
	ret, _ := url.JoinPath(cfg.EndpointUrl, "/lmm_gui_agent")
	return ret
}
