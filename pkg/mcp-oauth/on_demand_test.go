package mcpoauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestOnDemandCatalogueAndLifecycle(t *testing.T) {
	var calls, refresh atomic.Int32
	var size atomic.Int32
	size.Store(186)
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/token" {
			refresh.Add(1)
			fmt.Fprint(w, `{"access_token":"access","refresh_token":"rotated","token_type":"Bearer","expires_in":3600}`)
			return
		}
		if r.URL.Path == "/revoke" {
			return
		}
		if r.Header.Get("Authorization") != "Bearer access" {
			w.WriteHeader(401)
			return
		}
		if r.Method != "POST" {
			w.WriteHeader(405)
			return
		}
		var in struct {
			ID     json.RawMessage `json:"id"`
			Method string          `json:"method"`
			Params struct {
				Cursor    string                 `json:"cursor"`
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			} `json:"params"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		if len(in.ID) == 0 {
			w.WriteHeader(202)
			return
		}
		var result interface{}
		switch in.Method {
		case "initialize":
			result = map[string]interface{}{"protocolVersion": "2025-03-26", "capabilities": map[string]interface{}{"tools": map[string]interface{}{}}, "serverInfo": map[string]string{"name": "mock", "version": "1"}}
		case "tools/list":
			start, end := 0, 100
			if in.Params.Cursor == "next" {
				start = 100
				end = int(size.Load())
			}
			tools := []interface{}{}
			for i := start; i < end; i++ {
				name := fmt.Sprintf("tool_%03d", i)
				description := "current data"
				if i == 185 {
					name = "get_historical_patents"
					description = "历史专利"
				}
				tools = append(tools, map[string]interface{}{"name": name, "description": description, "inputSchema": map[string]interface{}{"type": "object", "properties": map[string]interface{}{"schema_marker": map[string]string{"type": "string"}}}})
			}
			out := map[string]interface{}{"tools": tools}
			if start == 0 {
				out["nextCursor"] = "next"
			}
			result = out
		case "tools/call":
			calls.Add(1)
			if in.Params.Name != "get_historical_patents" || in.Params.Arguments["schema_marker"] != "selected" {
				t.Error("wrong call", in.Params)
			}
			result = map[string]interface{}{"content": []interface{}{map[string]string{"type": "text", "text": "historical result"}}}
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": in.ID, "result": result})
	}))
	defer upstream.Close()
	dir := t.TempDir()
	store := diskStore(t, dir)
	owner := Owner{"org", "alice"}
	p := Profile{Resource: upstream.URL, TokenEndpoint: upstream.URL + "/token", RevocationEndpoint: upstream.URL + "/revoke", ClientID: "test", AuthorizationEndpoint: upstream.URL + "/authorize", RedirectURI: "https://wanwu.example/callback"}
	_ = store.Update(context.Background(), owner, p.Resource, func(*Grant) (*Grant, error) {
		return &Grant{Profile: p, AccessToken: "expired", RefreshToken: "refresh", Epoch: "one", ExpiresAt: time.Now().Add(-time.Hour)}, nil
	})
	manager := &Manager{Store: diskStore(t, dir), Client: upstream.Client()}
	relay := &Relay{Manager: manager, Codec: store.codec, Allowed: func(a, b string) bool { return a == p.Resource && b == p.Resource }, OnDemand: func(a, b string) bool { return true }}
	ref, _ := store.codec.ReferenceGrant(owner, p.Resource, p.Resource, "one")
	invoke := func(method string, params interface{}) *httptest.ResponseRecorder {
		b, _ := json.Marshal(map[string]interface{}{"jsonrpc": "2.0", "id": 1, "method": method, "params": params})
		req := httptest.NewRequest("POST", "https://wanwu.example/relay", strings.NewReader(string(b)))
		req.Header.Set(RelayHeader, ref)
		w := httptest.NewRecorder()
		relay.ServeHTTP(w, req)
		return w
	}
	call := func(name string, args interface{}) string {
		return invoke("tools/call", map[string]interface{}{"name": name, "arguments": args}).Body.String()
	}
	listed := invoke("tools/list", nil).Body.String()
	if strings.Contains(listed, "schema_marker") || strings.Contains(listed, "tool_185") || !strings.Contains(listed, "mcp_search_tools") {
		t.Fatal(listed)
	}
	found := call("mcp_search_tools", map[string]interface{}{"query": "历史"})
	if !strings.Contains(found, "get_historical_patents") || !strings.Contains(found, "186") || strings.Contains(found, "schema_marker") {
		t.Fatal(found)
	}
	detail := call("mcp_tool_detail", map[string]interface{}{"name": "get_historical_patents"})
	if !strings.Contains(detail, "schema_marker") {
		t.Fatal(detail)
	}
	if calls.Load() != 0 {
		t.Fatal("discovery executed business tool")
	}
	executed := call("mcp_call_tool", map[string]interface{}{"name": "get_historical_patents", "arguments": map[string]string{"schema_marker": "selected"}})
	if !strings.Contains(executed, "historical result") || calls.Load() != 1 {
		t.Fatal(executed)
	}
	size.Store(187)
	found = call("mcp_search_tools", map[string]interface{}{"query": "tool_186"})
	if !strings.Contains(found, "tool_186") || !strings.Contains(found, "187") {
		t.Fatal("hardcoded or stale catalogue", found)
	}
	unknown := call("mcp_call_tool", map[string]interface{}{"name": "not_authorized"})
	if !strings.Contains(unknown, "current authorized catalogue") || calls.Load() != 1 {
		t.Fatal(unknown)
	}
	if refresh.Load() != 1 {
		t.Fatal("refresh not shared", refresh.Load())
	}
	if e := manager.Disconnect(context.Background(), owner, p.Resource); e != nil {
		t.Fatal(e)
	}
	if invoke("tools/list", nil).Code != 401 {
		t.Fatal("revoked capability still works")
	}
}
