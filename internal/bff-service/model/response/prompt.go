package response

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

type CustomPromptIDResp struct {
	CustomPromptID string `json:"customPromptId"` // 自定义提示词ID [EN] Custom prompt word ID
}

type CustomPrompt struct {
	CustomPromptIDResp                // 自定义提示词ID [EN] Custom prompt word ID
	Avatar             request.Avatar `json:"avatar"`   // 图标 [EN] icon
	Name               string         `json:"name"`     // 名称 [EN] name
	Desc               string         `json:"desc"`     // 描述 [EN] describe
	Prompt             string         `json:"prompt"`   // 提示词 [EN] prompt word
	UpdateAt           string         `json:"updateAt"` // 更新时间 [EN] Update time
}

type CustomPromptOpt struct {
	Code     *int                       `json:"code"`     // 状态码 [EN] status code
	Message  string                     `json:"message"`  // 状态描述 [EN] Status description
	Response string                     `json:"response"` // 响应内容 [EN] Response content
	Finish   int                        `json:"finish"`   // 结束标志 [EN] end sign
	Usage    *mp_common.OpenAIRespUsage `json:"usage"`    // token使用统计 [EN] Token usage statistics
}
