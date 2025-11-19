package response

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	mp_common "github.com/UnicomAI/wanwu/pkg/model-provider/mp-common"
)

type CustomPromptIDResp struct {
	CustomPromptID string `json:"customPromptId"` // Custom prompt word ID
}

type CustomPrompt struct {
	CustomPromptIDResp                // Custom prompt word ID
	Avatar             request.Avatar `json:"avatar"`   // icon
	Name               string         `json:"name"`     // name
	Desc               string         `json:"desc"`     // describe
	Prompt             string         `json:"prompt"`   // prompt word
	UpdateAt           string         `json:"updateAt"` // Update time
}

type CustomPromptOpt struct {
	Code     *int                       `json:"code"`     // status code
	Message  string                     `json:"message"`  // Status description
	Response string                     `json:"response"` // Response content
	Finish   int                        `json:"finish"`   // end sign
	Usage    *mp_common.OpenAIRespUsage `json:"usage"`    // Token usage statistics
}
