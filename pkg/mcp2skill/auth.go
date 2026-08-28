package mcp2skill

// Auth type constants — kept in sync with pkg/util.
const (
	AuthTypeNone         = "none"
	AuthTypeAPIKeyQuery  = "api_key_query"
	AuthTypeAPIKeyHeader = "api_key_header"

	ApiKeyHeaderPrefixBasic  = "basic"
	ApiKeyHeaderPrefixBearer = "bearer"
	ApiKeyHeaderPrefixCustom = "custom"

	ApiKeyHeaderDefault = "Authorization"
)

// APIAuthConfig describes API authentication for connecting to an MCP server.
// It is a self-contained copy of the subset of util.ApiAuthWebRequest that
// mcp2skill needs, so that the mcp2skill binary does not pull in heavy
// dependencies (api/proto/common, pkg/log, pkg/openapi3-util).
// The JSON field names are identical to util.ApiAuthWebRequest for compatibility.
type APIAuthConfig struct {
	AuthType           string `json:"authType"`
	ApiKeyHeaderPrefix string `json:"apiKeyHeaderPrefix"`
	ApiKeyHeader       string `json:"apiKeyHeader"`
	ApiKeyQueryParam   string `json:"apiKeyQueryParam"`
	ApiKeyValue        string `json:"apiKeyValue"`
}
