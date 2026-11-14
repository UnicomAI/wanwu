package opencode

import (
	"encoding/json"
	mcpoauth "github.com/UnicomAI/wanwu/pkg/mcp-oauth"
	"github.com/UnicomAI/wanwu/pkg/util"
	option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"testing"
)

func TestManagedOAuthRelayConfiguration(t *testing.T) {
	config, e := renderConfig(option.ModelConfig{Provider: "mock", Model: "mock"}, []option.MCP{{Name: "oauth-connection", URL: "https://wanwu.example/relay", ApiAuth: &util.ApiAuthWebRequest{AuthType: util.AuthTypeAPIKeyHeader, ApiKeyHeader: mcpoauth.RelayHeader, ApiKeyHeaderPrefix: util.ApiKeyHeaderPrefixCustom, ApiKeyValue: "mock-internal-capability"}}}, false)
	if e != nil {
		t.Fatal(e)
	}
	var wire struct {
		MCP map[string]struct {
			URL     string            `json:"url"`
			Headers map[string]string `json:"headers"`
		} `json:"mcp"`
	}
	if e = json.Unmarshal([]byte(config), &wire); e != nil {
		t.Fatal(e)
	}
	remote := wire.MCP["oauth-connection"]
	if remote.URL != "https://wanwu.example/relay" || remote.Headers[mcpoauth.RelayHeader] != "mock-internal-capability" {
		t.Fatal("sandbox lost managed OAuth configuration")
	}
	if remote.Headers["Authorization"] != "" {
		t.Fatal("provider credential exported to sandbox")
	}
}
