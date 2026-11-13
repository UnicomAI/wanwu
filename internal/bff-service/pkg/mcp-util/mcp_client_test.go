package mcp_util

import (
	mcpoauth "github.com/UnicomAI/wanwu/pkg/mcp-oauth"
	"net/http"
	"testing"
)

func TestOAuthRelayRetainsTLSAndStrictDialPolicy(t *testing.T) {
	for _, strict := range []bool{true, false} {
		client := newHTTPClientWithHeaders(map[string]string{mcpoauth.RelayHeader: "synthetic-reference"}, strict)
		transport := client.Transport.(*headerTransport).base.(*http.Transport)
		if transport.TLSClientConfig.InsecureSkipVerify {
			t.Fatal("OAuth relay must verify TLS")
		}
		if (transport.DialContext != nil) != strict {
			t.Fatal("strict public-address dial policy lost")
		}
		if (client.CheckRedirect != nil) != strict {
			t.Fatal("strict redirect policy lost")
		}
	}
}
