package mcpoauth

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func diskStore(t *testing.T, dir string) *FileStore {
	t.Helper()
	if e := os.Chmod(dir, 0700); e != nil {
		t.Fatal(e)
	}
	c, e := NewCodec([]byte(strings.Repeat("k", 32)))
	if e != nil {
		t.Fatal(e)
	}
	s, e := NewFileStore(dir, c)
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestFileStoreRestartIsolationAndRefreshSerialization(t *testing.T) {
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	s := diskStore(t, dir)
	o := Owner{"org", "user"}
	p := profile()
	ctx := context.Background()
	if e := s.Update(ctx, o, p.Resource, func(*Grant) (*Grant, error) {
		return &Grant{AccessToken: "secret-access", RefreshToken: "secret-refresh", Generation: "v1", Profile: p, ExpiresAt: time.Now().Add(time.Hour)}, nil
	}); e != nil {
		t.Fatal(e)
	}
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		b, _ := os.ReadFile(filepath.Join(dir, f.Name()))
		if strings.Contains(string(b), "secret-") {
			t.Fatal("plaintext credential on disk")
		}
	}
	restarted := diskStore(t, dir)
	m := Manager{Store: restarted}
	token, e := m.AccessToken(ctx, o, p)
	if e != nil || token != "secret-access" {
		t.Fatal("restart lost credential", e)
	}
	if _, e = m.AccessToken(ctx, Owner{"other", "user"}, p); !errors.Is(e, ErrUnauthorized) {
		t.Fatal("tenant isolation", e)
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			other := diskStore(t, dir)
			if e := other.Update(ctx, o, p.Resource, func(g *Grant) (*Grant, error) { g.ExpiresAt = g.ExpiresAt.Add(time.Second); return g, nil }); e != nil {
				t.Error(e)
			}
		}()
	}
	wg.Wait()
	if e := m.Disconnect(ctx, o, p.Resource); e == nil {
		t.Fatal("missing revocation endpoint must be reported")
	}
	if _, e = m.AccessToken(ctx, o, p); !errors.Is(e, ErrUnauthorized) {
		t.Fatal("disconnected token usable", e)
	}
}
func TestDisconnectRejectsLateCallback(t *testing.T) {
	dir := t.TempDir()
	os.Chmod(dir, 0700)
	s := diskStore(t, dir)
	m := Manager{Store: s}
	o := Owner{"org", "user"}
	p := profile()
	ctx := context.Background()
	raw, e := m.Begin(ctx, o, p)
	if e != nil {
		t.Fatal(e)
	}
	state := authorizationState(t, raw)
	if e = m.Disconnect(ctx, o, p.Resource); e != nil {
		t.Fatal(e)
	}
	if e = m.Complete(ctx, o, state, "late-code"); !errors.Is(e, ErrInvalidState) {
		t.Fatal("late callback restored grant", e)
	}
}

func authorizationState(t *testing.T, raw string) string {
	t.Helper()
	u, e := url.Parse(raw)
	if e != nil {
		t.Fatal(e)
	}
	return u.Query().Get("state")
}
