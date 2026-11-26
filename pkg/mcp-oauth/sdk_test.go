package mcpoauth

import (
	"context"
	"encoding/json"
	think "github.com/ThinkInAIXYZ/go-mcp/client"
	tp "github.com/ThinkInAIXYZ/go-mcp/protocol"
	tt "github.com/ThinkInAIXYZ/go-mcp/transport"
	mark "github.com/mark3labs/mcp-go/client"
	mt "github.com/mark3labs/mcp-go/client/transport"
	mp "github.com/mark3labs/mcp-go/mcp"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// Exercise both SDKs used in Wanwu, against an OAuth-protected mock resource.
// This is protocol integration, not a claim of real QCC account acceptance.
func TestBothMCPClientsThroughOAuthRelay(t *testing.T) {
	t.Run("transparent", func(t *testing.T) { testBothMCPClients(t, false) })
	t.Run("on-demand", func(t *testing.T) { testBothMCPClients(t, true) })
}
func testBothMCPClients(t *testing.T, onDemand bool) {
	expectedTools := 1
	callName := "company"
	args := map[string]interface{}{}
	if onDemand {
		expectedTools = 3
		callName = "mcp_call_tool"
		args = map[string]interface{}{"name": "company", "arguments": map[string]interface{}{}}
	}

	var called atomic.Int32
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer access" {
			w.WriteHeader(401)
			return
		}
		if r.Method == "GET" {
			w.WriteHeader(405)
			return
		}
		if r.Method == "DELETE" {
			w.WriteHeader(204)
			return
		}
		var in struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
		}
		if json.NewDecoder(r.Body).Decode(&in) != nil {
			w.WriteHeader(400)
			return
		}
		if len(in.ID) == 0 {
			w.WriteHeader(202)
			return
		}
		var result interface{}
		switch in.Method {
		case "initialize":
			result = map[string]interface{}{"protocolVersion": "2025-03-26", "capabilities": map[string]interface{}{"tools": map[string]interface{}{}}, "serverInfo": map[string]string{"name": "mock", "version": "1"}}
		case "tools/list":
			result = map[string]interface{}{"tools": []interface{}{map[string]interface{}{"name": "company", "description": "mock company lookup", "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}}}}
		case "tools/call":
			called.Add(1)
			result = map[string]interface{}{"content": []interface{}{map[string]string{"type": "text", "text": "mock result"}}}
		default:
			result = map[string]interface{}{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": in.ID, "result": result})
	}))
	defer upstream.Close()
	p := Profile{Resource: upstream.URL, AuthorizationEndpoint: upstream.URL + "/authorize", TokenEndpoint: upstream.URL + "/token", ClientID: "client", RedirectURI: "https://wanwu.example/callback"}
	store := diskStore(t, t.TempDir())
	owner := Owner{"org", "alice"}
	if e := store.Update(context.Background(), owner, p.Resource, func(*Grant) (*Grant, error) {
		return &Grant{AccessToken: "access", Profile: p, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}); e != nil {
		t.Fatal(e)
	}
	relay := httptest.NewServer(&Relay{Manager: &Manager{Store: store, Client: upstream.Client()}, Codec: store.codec, OnDemand: func(string, string) bool { return onDemand }, Allowed: func(resource, target string) bool { return resource == p.Resource && target == p.Resource }})
	defer relay.Close()
	reference, _ := store.codec.Reference(owner, p.Resource)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	t.Run("probe-sdk", func(t *testing.T) {
		httpClient := &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
			clone := r.Clone(r.Context())
			clone.Header.Set(RelayHeader, reference)
			return http.DefaultTransport.RoundTrip(clone)
		})}
		transport, e := tt.NewStreamableHTTPClientTransport(relay.URL, tt.WithStreamableHTTPClientOptionHTTPClient(httpClient))
		if e != nil {
			t.Fatal(e)
		}
		c, e := think.NewClient(transport)
		if e != nil {
			t.Fatal(e)
		}
		defer c.Close()
		list, e := c.ListTools(ctx)
		if e != nil || len(list.Tools) != expectedTools {
			t.Fatal("list", e)
		}
		result, e := c.CallTool(ctx, tp.NewCallToolRequest(callName, args))
		if e != nil || len(result.Content) != 1 {
			t.Fatal("call", e)
		}
	})
	t.Run("agent-sdk", func(t *testing.T) {
		c, e := mark.NewStreamableHttpClient(relay.URL, mt.WithHTTPHeaders(map[string]string{RelayHeader: reference}))
		if e != nil {
			t.Fatal(e)
		}
		defer c.Close()
		if e = c.Start(ctx); e != nil {
			t.Fatal(e)
		}
		init := mp.InitializeRequest{}
		init.Params.ProtocolVersion = mp.LATEST_PROTOCOL_VERSION
		init.Params.ClientInfo = mp.Implementation{Name: "test", Version: "1"}
		if _, e = c.Initialize(ctx, init); e != nil {
			t.Fatal(e)
		}
		list, e := c.ListTools(ctx, mp.ListToolsRequest{})
		if e != nil || len(list.Tools) != expectedTools {
			t.Fatal("list", e)
		}
		call := mp.CallToolRequest{}
		call.Params.Name = callName
		call.Params.Arguments = args
		result, e := c.CallTool(ctx, call)
		if e != nil || len(result.Content) != 1 {
			t.Fatal("call", e)
		}
	})
	if called.Load() != 2 {
		t.Fatalf("expected 2 actual tools/call requests, got %d", called.Load())
	}
}
