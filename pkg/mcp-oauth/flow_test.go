package mcpoauth

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

type memoryStore struct {
	mu      sync.Mutex
	pending map[string]Pending
	grants  map[Owner]map[string]*Grant
}

func newStore() *memoryStore {
	return &memoryStore{pending: map[string]Pending{}, grants: map[Owner]map[string]*Grant{}}
}
func (s *memoryStore) PutPending(ctx context.Context, k string, p Pending) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[k] = p
	return nil
}
func (s *memoryStore) Take(ctx context.Context, k string, o Owner) (Pending, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.pending[k]
	if !ok || p.Owner != o {
		return Pending{}, ErrInvalidState
	}
	delete(s.pending, k)
	return p, nil
}
func (s *memoryStore) Update(ctx context.Context, o Owner, r string, f func(*Grant) (*Grant, error)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.grants[o] == nil {
		s.grants[o] = map[string]*Grant{}
	}
	g, e := f(s.grants[o][r])
	if e == nil {
		s.grants[o][r] = g
	}
	return e
}
func (s *memoryStore) Delete(ctx context.Context, o Owner, r string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.grants[o], r)
	return nil
}

type roundTrip func(*http.Request) (*http.Response, error)

func (f roundTrip) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func profile() Profile {
	return Profile{Resource: "https://resource.example/mcp", AuthorizationEndpoint: "https://auth.example/authorize", TokenEndpoint: "https://auth.example/token", ClientID: "test-client", RedirectURI: "https://wanwu.example/callback", Scope: "mcp:tools"}
}
func TestOwnerStateAndPKCE(t *testing.T) {
	ctx := context.Background()
	s := newStore()
	m := Manager{Store: s}
	o := Owner{"org", "alice"}
	p := profile()
	raw, e := m.Begin(ctx, o, p)
	if e != nil {
		t.Fatal(e)
	}
	u, _ := url.Parse(raw)
	q := u.Query()
	if q.Get("code_challenge_method") != "S256" || len(q.Get("code_challenge")) != 43 || q.Get("resource") != p.Resource {
		t.Fatal("invalid PKCE authorization")
	}
	if e = m.Complete(ctx, Owner{"other", "alice"}, q.Get("state"), "code"); !errors.Is(e, ErrInvalidState) {
		t.Fatal("cross-tenant callback accepted")
	}
	calls := 0
	m.Client = &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		_ = r.ParseForm()
		if len(r.Form.Get("code_verifier")) != 43 {
			t.Error("missing verifier")
		}
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"access_token":"access","refresh_token":"refresh","token_type":"Bearer","expires_in":3600}`)), Header: make(http.Header)}, nil
	})}
	if e = m.Complete(ctx, o, q.Get("state"), "code"); e != nil {
		t.Fatal(e)
	}
	if e = m.Complete(ctx, o, q.Get("state"), "code"); !errors.Is(e, ErrInvalidState) {
		t.Fatal("replay accepted")
	}
	if calls != 1 {
		t.Fatal(calls)
	}
	token, e := m.AccessToken(ctx, o, p)
	if e != nil || token != "access" {
		t.Fatal(e)
	}
	if _, e = m.AccessToken(ctx, Owner{"org", "bob"}, p); !errors.Is(e, ErrUnauthorized) {
		t.Fatal("cross-user token accepted")
	}
}
func TestConcurrentRefresh(t *testing.T) {
	ctx := context.Background()
	s := newStore()
	o := Owner{"org", "alice"}
	p := profile()
	s.Update(ctx, o, p.Resource, func(*Grant) (*Grant, error) {
		return &Grant{AccessToken: "old", RefreshToken: "rotate", ExpiresAt: time.Now().Add(-time.Hour)}, nil
	})
	calls := 0
	m := Manager{Store: s, Client: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		calls++
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"access_token":"new","refresh_token":"rotated","token_type":"Bearer","expires_in":3600}`)), Header: make(http.Header)}, nil
	})}}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			v, e := m.AccessToken(ctx, o, p)
			if e != nil || v != "new" {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	if calls != 1 {
		t.Fatalf("refreshes=%d", calls)
	}
}
func TestRejectInsecureProfile(t *testing.T) {
	p := profile()
	p.TokenEndpoint = "http://auth.example/token"
	m := Manager{Store: newStore()}
	if _, e := m.Begin(context.Background(), Owner{"org", "user"}, p); e == nil {
		t.Fatal("insecure endpoint accepted")
	}
}

func TestInvalidGrantDisablesConnection(t *testing.T) {
	ctx := context.Background()
	s := newStore()
	o := Owner{"org", "alice"}
	p := profile()
	s.Update(ctx, o, p.Resource, func(*Grant) (*Grant, error) {
		return &Grant{AccessToken: "old", RefreshToken: "revoked", ExpiresAt: time.Now().Add(-time.Hour), Epoch: "old"}, nil
	})
	m := Manager{Store: s, Client: &http.Client{Transport: roundTrip(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 400, Body: io.NopCloser(strings.NewReader(`{"error":"invalid_grant"}`))}, nil
	})}}
	if _, e := m.AccessToken(ctx, o, p); !errors.Is(e, ErrGrantRevoked) {
		t.Fatal(e)
	}
	authorized, e := m.Status(ctx, o, p.Resource)
	if e != nil || authorized {
		t.Fatal("revoked grant remains authorized", e)
	}
}
