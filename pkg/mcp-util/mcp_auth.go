package mcp_util

import (
	"fmt"
	net_url "net/url"

	"github.com/UnicomAI/wanwu/api/proto/common"
	"github.com/UnicomAI/wanwu/pkg/util"
)

// 将鉴权中的header参数合并到headers，鉴权的query参数追加url
func MergeMcpParams(url string, auth *common.ApiAuthWebRequest, headers map[string]string) (mergedUrl string, mergedHeaders map[string]string, err error) {
	if (auth == nil || auth.AuthType == "" || auth.AuthType == util.AuthTypeNone) && len(headers) == 0 {
		return url, headers, nil
	}
	mergedHeaders = make(map[string]string)
	mergedUrl = url

	for k, v := range headers {
		mergedHeaders[k] = v
	}

	if auth != nil && auth.AuthType != "" && auth.AuthType != util.AuthTypeNone {
		switch auth.AuthType {
		case util.AuthTypeAPIKeyHeader:
			value := auth.ApiKeyValue
			switch auth.ApiKeyHeaderPrefix {
			case util.ApiKeyHeaderPrefixBasic:
				value = "Basic " + auth.ApiKeyValue
			case util.ApiKeyHeaderPrefixBearer:
				value = "Bearer " + auth.ApiKeyValue
			}
			headerName := auth.ApiKeyHeader
			if headerName == "" {
				headerName = util.ApiKeyHeaderDefault
			}
			if _, ok := mergedHeaders[headerName]; ok {
				return url, mergedHeaders, fmt.Errorf("header %s already exists", headerName)
			}
			mergedHeaders[headerName] = value
		case util.AuthTypeAPIKeyQuery:
			rawUrl, err := net_url.Parse(url)
			if err != nil {
				return url, mergedHeaders, err
			}
			queryParams := rawUrl.Query()
			if queryParams.Has(auth.ApiKeyQueryParam) {
				return url, mergedHeaders, fmt.Errorf("query param %s already exists", auth.ApiKeyQueryParam)
			}
			queryParams.Add(auth.ApiKeyQueryParam, auth.ApiKeyValue)
			rawUrl.RawQuery = queryParams.Encode()
			mergedUrl = rawUrl.String()
		}
	}
	return mergedUrl, mergedHeaders, nil
}
