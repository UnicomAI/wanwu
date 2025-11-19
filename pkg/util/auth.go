package util

import (
	"fmt"

	"github.com/UnicomAI/wanwu/api/proto/common"
	openapi3_util "github.com/UnicomAI/wanwu/pkg/openapi3-util"
)

// api authentication
const (
	AuthTypeNone             = "none"
	AuthTypeAPIKeyQuery      = "api_key_query"
	AuthTypeAPIKeyHeader     = "api_key_header"
	ApiKeyHeaderPrefixBasic  = "basic"
	ApiKeyHeaderPrefixBearer = "bearer"
	ApiKeyHeaderPrefixCustom = "custom"
	ApiKeyHeaderDefault      = "Authorization"
)

type ApiAuthWebRequest struct {
	AuthType           string `json:"authType" validate:"required,oneof='none' 'api_key_query' 'api_key_header'"` // Authentication type None or request header or query parameter
	ApiKeyHeaderPrefix string `json:"apiKeyHeaderPrefix" validate:"omitempty,oneof='basic' 'bearer' 'custom'"`    // Authentication header prefix
	ApiKeyHeader       string `json:"apiKeyHeader"`                                                               // HTTP header name
	ApiKeyQueryParam   string `json:"apiKeyQueryParam"`                                                           // Query parameter name
	ApiKeyValue        string `json:"apiKeyValue"`                                                                // apiKey
}

func (req *ApiAuthWebRequest) Check() error {
	switch req.AuthType {
	case AuthTypeNone:
		return nil
	case AuthTypeAPIKeyQuery:
		if req.ApiKeyQueryParam == "" {
			return fmt.Errorf("apiKeyQueryParam is empty")
		}
		if req.ApiKeyValue == "" {
			return fmt.Errorf("apiKeyValue is empty")
		}
		return nil
	case AuthTypeAPIKeyHeader:
		if req.ApiKeyHeader == "" {
			return fmt.Errorf("apiKeyHeader is empty")
		}
		if req.ApiKeyValue == "" {
			return fmt.Errorf("apiKeyValue is empty")
		}
		return nil
	default:
		return fmt.Errorf("invalid authType: %v", req.AuthType)
	}
}

func (auth *ApiAuthWebRequest) ToOpenapiAuth() (*openapi3_util.Auth, error) {
	if auth == nil || auth.AuthType == "" || auth.AuthType == AuthTypeNone {
		return &openapi3_util.Auth{
			Type: "none",
		}, nil
	}
	if err := auth.Check(); err != nil {
		return nil, err
	}
	ret := &openapi3_util.Auth{Type: "apiKey"}
	switch auth.AuthType {
	case AuthTypeNone:
		return ret, nil
	case AuthTypeAPIKeyQuery:
		ret.In = "query"
		ret.Name = auth.ApiKeyQueryParam
		ret.Value = auth.ApiKeyValue
		return ret, nil
	case AuthTypeAPIKeyHeader:
		ret.In = "header"
		switch auth.ApiKeyHeaderPrefix {
		case ApiKeyHeaderPrefixBasic:
			// FIXME
			ret.Name = auth.ApiKeyHeader
			ret.Value = "Basic " + auth.ApiKeyValue
			return ret, nil
		case ApiKeyHeaderPrefixBearer:
			ret.Name = auth.ApiKeyHeader
			ret.Value = "Bearer " + auth.ApiKeyValue
			return ret, nil
		case ApiKeyHeaderPrefixCustom:
			ret.Name = auth.ApiKeyHeader
			ret.Value = auth.ApiKeyValue
			return ret, nil
		default:
			return nil, fmt.Errorf("invalid apiKeyHeaderPrefix: %v", auth.ApiKeyHeaderPrefix)
		}
	default:
		return nil, fmt.Errorf("invalid authType: %v", auth.AuthType)
	}
}

func ConvertApiAuthWebRequestProto(auth *common.ApiAuthWebRequest) (*openapi3_util.Auth, error) {
	authWeb := &ApiAuthWebRequest{
		AuthType:           auth.GetAuthType(),
		ApiKeyHeaderPrefix: auth.GetApiKeyHeaderPrefix(),
		ApiKeyHeader:       auth.GetApiKeyHeader(),
		ApiKeyQueryParam:   auth.GetApiKeyQueryParam(),
		ApiKeyValue:        auth.GetApiKeyValue(),
	}
	return authWeb.ToOpenapiAuth()
}

func ConvertApiAuthProto(auth *common.ApiAuth) *openapi3_util.Auth {
	return &openapi3_util.Auth{
		Type:  auth.GetAuthType(),
		In:    auth.GetAuthIn(),
		Name:  auth.GetAuthName(),
		Value: auth.GetAuthValue(),
	}
}
