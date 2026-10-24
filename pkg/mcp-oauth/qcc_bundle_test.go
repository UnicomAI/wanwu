package mcpoauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestQCCPresetBundle(t *testing.T) {
	b, e := os.ReadFile("../../docs/mcp-oauth/qcc.example.json")
	if e != nil {
		t.Fatal(e)
	}
	var cfg struct {
		Providers []Provider `json:"providers"`
	}
	if e = json.Unmarshal(b, &cfg); e != nil {
		t.Fatal(e)
	}
	if len(cfg.Providers) != 1 {
		t.Fatal("expected one shared grant")
	}
	p := cfg.Providers[0]
	if p.ID != "qcc" || p.Scope != "mcp:tools" || p.Resource != "https://agent.qcc.com/mcp/company/stream" || len(p.Targets) != 5 {
		t.Fatal("invalid QCC profile")
	}
	names := []string{"company", "risk", "ipr", "operation", "executive"}
	for i, name := range names {
		if p.Targets[i].Resource != "https://agent.qcc.com/mcp/"+name+"/stream" {
			t.Fatal("unexpected resource", p.Targets[i])
		}
	}
}
func TestFiveTargetsShareRefreshAndRevocation(t *testing.T) {
	var refresh, calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			refresh.Add(1)
			w.Header().Set("Content-Type", "application/json")
			io.WriteString(w, `{"access_token":"fresh","refresh_token":"rotated","token_type":"Bearer","expires_in":3600}`)
			return
		}
		if r.URL.Path == "/revoke" {
			w.WriteHeader(200)
			return
		}
		if r.Header.Get("Authorization") != "Bearer fresh" {
			t.Error("missing shared grant")
		}
		calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"mock"}]}}`)
	}))
	defer server.Close()
	ctx := context.Background()
	o := Owner{"org", "user"}
	dir := t.TempDir()
	s := diskStore(t, dir)
	p := Profile{Resource: server.URL + "/company", AuthorizationEndpoint: server.URL + "/authorize", TokenEndpoint: server.URL + "/token", RevocationEndpoint: server.URL + "/revoke", ClientID: "mock", RedirectURI: "https://wanwu.example/callback"}
	if e := s.Update(ctx, o, p.Resource, func(*Grant) (*Grant, error) {
		return &Grant{AccessToken: "expired", RefreshToken: "shared", Profile: p, ExpiresAt: time.Now().Add(-time.Hour), Epoch: "epoch"}, nil
	}); e != nil {
		t.Fatal(e)
	}
	m := &Manager{Store: diskStore(t, dir), Client: server.Client()}
	targets := []string{"company", "risk", "ipr", "operation", "executive"}
	relay := &Relay{Manager: m, Codec: s.codec, Allowed: func(resource, target string) bool {
		if resource != p.Resource {
			return false
		}
		for _, name := range targets {
			if target == server.URL+"/"+name {
				return true
			}
		}
		return false
	}}
	refs := []string{}
	for _, name := range targets {
		ref, e := s.codec.ReferenceGrant(o, p.Resource, server.URL+"/"+name, "epoch")
		if e != nil {
			t.Fatal(e)
		}
		refs = append(refs, ref)
	}
	invoke := func(ref string) int {
		req := httptest.NewRequest("POST", "https://wanwu.example/relay", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call"}`))
		req.Header.Set(RelayHeader, ref)
		w := httptest.NewRecorder()
		relay.ServeHTTP(w, req)
		return w.Code
	}
	var wg sync.WaitGroup
	for _, ref := range refs {
		wg.Add(1)
		go func(ref string) {
			defer wg.Done()
			if status := invoke(ref); status != 200 {
				t.Error(status)
			}
		}(ref)
	}
	wg.Wait()
	if refresh.Load() != 1 || calls.Load() != 5 {
		t.Fatal("bundle did not share refresh", refresh.Load(), calls.Load())
	}
	if e := m.Disconnect(ctx, o, p.Resource); e != nil {
		t.Fatal(e)
	}
	for _, ref := range refs {
		if invoke(ref) != 401 {
			t.Fatal("bundle remained authorized")
		}
	}
	if calls.Load() != 5 {
		t.Fatal("revoked call reached provider")
	}
}
