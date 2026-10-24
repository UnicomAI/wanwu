package mcpoauth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// RelayHeader carries a resource-scoped internal capability, never a provider
// access/refresh token. Treat it as a secret in logs and connector exports.
const RelayHeader = "X-Wanwu-MCP-OAuth"
const RelayResource = "wanwu-mcp-relay-v1"

type Reference struct {
	Epoch    string
	Owner    Owner
	Resource string
	Target   string
}

func (c *Codec) Reference(o Owner, r string) (string, error) {
	return c.ReferenceTarget(o, r, r)
}
func (c *Codec) ReferenceTarget(o Owner, r, target string) (string, error) {
	return c.ReferenceGrant(o, r, target, "")
}
func (c *Codec) ReferenceGrant(o Owner, r, target, epoch string) (string, error) {
	b, _ := json.Marshal(Reference{Owner: o, Resource: r, Target: target, Epoch: epoch})
	sealed, e := c.sealBytes(Owner{}, RelayResource, b)
	return base64.RawURLEncoding.EncodeToString(sealed), e
}
func (c *Codec) ReadReference(raw string) (Reference, error) {
	var ref Reference
	b, e := base64.RawURLEncoding.DecodeString(raw)
	if e != nil {
		return ref, ErrUnauthorized
	}
	b, e = c.openBytes(Owner{}, RelayResource, b)
	if e != nil {
		return ref, e
	}
	if json.Unmarshal(b, &ref) != nil || ref.Owner.Organization == "" || ref.Owner.User == "" || (!validHTTPS(ref.Resource) || !validHTTPS(ref.Target)) {
		return ref, ErrUnauthorized
	}
	return ref, nil
}

// Relay makes credentials available to every existing MCP execution chain,
// including subprocess sandbox runners, without exporting provider tokens.
// Only Streamable HTTP is supported: legacy SSE endpoint-event rewriting is not
// silently attempted. Provider URLs come exclusively from administrator profiles.
type Relay struct {
	Manager *Manager
	Codec   *Codec
	Allowed func(string, string) bool
}

func (r *Relay) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	if req.Method != "POST" && req.Method != "GET" && req.Method != "DELETE" {
		w.WriteHeader(405)
		return
	}
	if req.URL.RawQuery != "" {
		http.Error(w, "unexpected relay query", 400)
		return
	}
	ref, e := r.Codec.ReadReference(req.Header.Get(RelayHeader))
	if e != nil || r.Allowed == nil || !r.Allowed(ref.Resource, ref.Target) {
		http.Error(w, "MCP authorization required", 401)
		return
	}
	var p Profile
	e = r.Manager.Store.Update(req.Context(), ref.Owner, ref.Resource, func(g *Grant) (*Grant, error) {
		if g == nil || g.AccessToken == "" || g.Epoch != ref.Epoch {
			return nil, ErrUnauthorized
		}
		p = g.Profile
		return g, nil
	})
	if e != nil {
		http.Error(w, "MCP authorization required", 401)
		return
	}
	token, e := r.Manager.AccessToken(req.Context(), ref.Owner, p)
	if e != nil {
		http.Error(w, "MCP authorization required", 401)
		return
	}
	e = r.Manager.Store.Update(req.Context(), ref.Owner, ref.Resource, func(g *Grant) (*Grant, error) {
		if g == nil || g.Epoch != ref.Epoch || g.AccessToken == "" {
			return nil, ErrUnauthorized
		}
		return g, nil
	})
	if e != nil {
		http.Error(w, "MCP authorization required", 401)
		return
	}
	target, e := http.NewRequestWithContext(req.Context(), req.Method, ref.Target, http.MaxBytesReader(w, req.Body, 8<<20))
	if e != nil {
		http.Error(w, "invalid MCP configuration", 502)
		return
	}
	for _, key := range []string{"Accept", "Content-Type", "Mcp-Session-Id", "MCP-Protocol-Version", "Last-Event-ID"} {
		if value := req.Header.Get(key); value != "" {
			target.Header.Set(key, value)
		}
	}
	target.Header.Set("Authorization", "Bearer "+token)
	client := r.Manager.httpClient()
	client.Timeout = 0 // request context owns streaming lifetime
	res, e := client.Do(target)
	if e != nil {
		http.Error(w, "MCP upstream unavailable", 502)
		return
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 && res.StatusCode < 400 {
		http.Error(w, "MCP redirect rejected", 502)
		return
	}
	if res.StatusCode == 401 {
		// Do not replay tools/call: invalidate only this access token; next request
		// refreshes under the grant lock. A concurrent successful refresh is retained.
		_ = r.Manager.Store.Update(req.Context(), ref.Owner, ref.Resource, func(g *Grant) (*Grant, error) {
			if g == nil {
				return &Grant{}, nil
			}
			if g.AccessToken == token {
				g.ExpiresAt = time.Time{}
			}
			return g, nil
		})
		http.Error(w, "MCP authorization expired; reconnect or retry", 401)
		return
	}
	for _, key := range []string{"Content-Type", "Mcp-Session-Id", "MCP-Protocol-Version", "Retry-After"} {
		if value := res.Header.Get(key); value != "" {
			w.Header().Set(key, value)
		}
	}
	w.WriteHeader(res.StatusCode)
	if strings.HasPrefix(res.Header.Get("Content-Type"), "text/event-stream") {
		_, _ = io.Copy(flushWriter{w}, res.Body)
	} else {
		_, _ = io.Copy(w, res.Body)
	}
}

type flushWriter struct{ http.ResponseWriter }

func (w flushWriter) Write(b []byte) (int, error) {
	n, e := w.ResponseWriter.Write(b)
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
	return n, e
}

func (m *Manager) Epoch(ctx context.Context, o Owner, r string) (string, error) {
	var epoch string
	e := m.Store.Update(ctx, o, r, func(g *Grant) (*Grant, error) {
		if g == nil || g.AccessToken == "" {
			return nil, ErrUnauthorized
		}
		epoch = g.Epoch
		return g, nil
	})
	return epoch, e
}
