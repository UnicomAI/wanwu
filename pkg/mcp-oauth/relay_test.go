package mcpoauth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRelayRefreshRestartCallAndUnbind(t *testing.T) {
	var calls, refresh atomic.Int32
	upstream := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		calls.Add(1)
		if r.Header.Get("Authorization") != "Bearer fresh" || r.Header.Get(RelayHeader) != "" || r.Header.Get("Cookie") != "" {
			t.Error("credential boundary violated")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Mcp-Session-Id", "session")
		io.WriteString(w, `{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"company result"}]}}`)
	}))
	defer upstream.Close()
	dir := t.TempDir()
	store := diskStore(t, dir)
	codec := store.codec
	o := Owner{"org", "alice"}
	p := Profile{Resource: upstream.URL + "/mcp", AuthorizationEndpoint: upstream.URL + "/authorize", TokenEndpoint: upstream.URL + "/token", RevocationEndpoint: upstream.URL + "/revoke", ClientID: "test", RedirectURI: "https://wanwu.example/aibase/oauth/mcp/callback"}
	e := store.Update(context.Background(), o, p.Resource, func(*Grant) (*Grant, error) {
		return &Grant{Generation: "1", Profile: p, AccessToken: "old", RefreshToken: "old-refresh", ExpiresAt: time.Now().Add(-time.Hour)}, nil
	})
	if e != nil {
		t.Fatal(e)
	}
	manager := &Manager{Store: diskStore(t, dir), Client: upstream.Client()}
	relay := &Relay{Manager: manager, Codec: codec, Allowed: func(resource, target string) bool { return resource == p.Resource && target == p.Resource }}
	ref, _ := codec.Reference(o, p.Resource)
	invoke := func(reference string) *httptest.ResponseRecorder {
		req := httptest.NewRequest("POST", "https://wanwu.example/relay", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"query"}}`))
		req.Header.Set(RelayHeader, reference)
		req.Header.Set("Cookie", "must-not-leak")
		req.Header.Set("Authorization", "must-not-leak")
		out := httptest.NewRecorder()
		relay.ServeHTTP(out, req)
		return out
	}
	result := invoke(ref)
	if result.Code != 200 || !strings.Contains(result.Body.String(), "company result") || result.Header().Get("Mcp-Session-Id") != "session" {
		t.Fatal(result.Code, result.Body.String())
	}
	if refresh.Load() != 1 {
		t.Fatal("refresh missing")
	}
	wrong, _ := codec.Reference(Owner{"other", "alice"}, p.Resource)
	if invoke(wrong).Code != 401 {
		t.Fatal("cross-tenant grant used")
	}
	if invoke(ref+"bad").Code != 401 {
		t.Fatal("tampered reference accepted")
	}
	if e := manager.Disconnect(context.Background(), o, p.Resource); e != nil {
		t.Fatal(e)
	}
	if invoke(ref).Code != 401 {
		t.Fatal("revoked reference remained usable")
	}
	if e := manager.Store.Update(context.Background(), o, p.Resource, func(g *Grant) (*Grant, error) {
		g.AccessToken = "fresh"
		g.Profile = p
		g.ExpiresAt = time.Now().Add(time.Hour)
		return g, nil
	}); e != nil {
		t.Fatal(e)
	}
	if invoke(ref).Code != 401 {
		t.Fatal("old capability resurrected after reauthorization")
	}
	if calls.Load() != 1 {
		t.Fatal("unauthorized upstream call", calls.Load())
	}
}
func TestDiscoveryRejectsUntrustedMetadata(t *testing.T) {
	m := Manager{Client: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		var body interface{}
		if r.URL.Path == "/resource-meta" {
			body = map[string]interface{}{"resource": "https://resource.example/mcp", "authorization_servers": []string{"https://auth.example"}}
		} else {
			body = map[string]interface{}{"issuer": "https://auth.example", "authorization_endpoint": "https://auth.example/authorize", "token_endpoint": "https://attacker.example/token", "code_challenge_methods_supported": []string{"S256"}}
		}
		b, _ := json.Marshal(body)
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(string(b)))}, nil
	})}}
	_, e := m.Discover(context.Background(), Provider{Resource: "https://resource.example/mcp", ResourceMetadataURL: "https://resource.example/resource-meta", Issuer: "https://auth.example", RedirectURI: "https://wanwu.example/callback", ClientID: "known"})
	if e == nil {
		t.Fatal("untrusted token endpoint accepted")
	}
}
