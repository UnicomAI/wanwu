package mcpoauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func randomString() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
func validHTTPS(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.Hostname() != "" && u.User == nil && u.Fragment == ""
}
func validate(p Profile) error {
	if !validHTTPS(p.Resource) || !validHTTPS(p.AuthorizationEndpoint) || !validHTTPS(p.TokenEndpoint) || p.ClientID == "" {
		return ErrInvalidConfig
	}
	if p.RevocationEndpoint != "" && !validHTTPS(p.RevocationEndpoint) {
		return ErrInvalidConfig
	}
	// Deployment callback is configured by the administrator, not the caller.
	if !validHTTPS(p.RedirectURI) {
		return ErrInvalidConfig
	}
	return nil
}
func (m *Manager) Begin(ctx context.Context, owner Owner, p Profile) (string, error) {
	if owner.Organization == "" || owner.User == "" || m.Store == nil {
		return "", ErrInvalidConfig
	}
	if err := validate(p); err != nil {
		return "", err
	}
	state, err := randomString()
	if err != nil {
		return "", err
	}
	verifier, err := randomString()
	if err != nil {
		return "", err
	}
	hash := sha256.Sum256([]byte(verifier))
	u, _ := url.Parse(p.AuthorizationEndpoint)
	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", p.ClientID)
	q.Set("redirect_uri", p.RedirectURI)
	q.Set("state", state)
	q.Set("code_challenge", base64.RawURLEncoding.EncodeToString(hash[:]))
	q.Set("code_challenge_method", "S256")
	q.Set("resource", p.Resource)
	if p.Scope != "" {
		q.Set("scope", p.Scope)
	}
	u.RawQuery = q.Encode()
	var generation string
	if err := m.Store.Update(ctx, owner, p.Resource, func(g *Grant) (*Grant, error) {
		if g == nil {
			g = &Grant{}
		}
		{
			var e error
			g.Generation, e = randomString()
			if e != nil {
				return nil, e
			}
		}
		if g.Epoch == "" {
			var e error
			g.Epoch, e = randomString()
			if e != nil {
				return nil, e
			}
		}
		g.Pending = true
		generation = g.Generation
		return g, nil
	}); err != nil {
		return "", err
	}
	if err := m.Store.PutPending(ctx, state, Pending{Generation: generation, Owner: owner, Profile: p, Verifier: verifier, ExpiresAt: m.now().Add(10 * time.Minute)}); err != nil {
		return "", err
	}
	return u.String(), nil
}

// Complete must be invoked with the same authenticated owner who began the flow.
// The callback must reject duplicate query parameters before invoking this method.
func (m *Manager) Complete(ctx context.Context, owner Owner, state, code string) error {
	if state == "" || code == "" {
		return ErrInvalidState
	}
	pending, err := m.Store.Take(ctx, state, owner)
	if err != nil || pending.Owner != owner || !m.now().Before(pending.ExpiresAt) {
		return ErrInvalidState
	}
	p := pending.Profile
	if err := validate(p); err != nil {
		return err
	}
	values := url.Values{"grant_type": {"authorization_code"}, "code": {code}, "code_verifier": {pending.Verifier}, "client_id": {p.ClientID}, "redirect_uri": {p.RedirectURI}, "resource": {p.Resource}}
	return m.Store.Update(ctx, owner, p.Resource, func(old *Grant) (*Grant, error) {
		if old == nil || old.Generation != pending.Generation {
			return nil, ErrInvalidState
		}
		grant, err := m.exchange(ctx, p.TokenEndpoint, values)
		if err != nil {
			return nil, err
		}
		grant.Generation = old.Generation
		grant.Epoch = old.Epoch
		grant.Profile = p
		return grant, nil
	})
}
func (m *Manager) httpClient() *http.Client {
	c := http.Client{Timeout: 30 * time.Second}
	if m.Client != nil {
		c = *m.Client
		if c.Timeout == 0 {
			c.Timeout = 30 * time.Second
		}
	}
	// Tokens and authorization codes must never follow redirects.
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse }
	return &c
}
func (m *Manager) exchange(ctx context.Context, endpoint string, values url.Values) (*Grant, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := m.httpClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("OAuth token request failed")
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var failure struct {
			Error string `json:"error"`
		}
		_ = json.NewDecoder(io.LimitReader(res.Body, 4096)).Decode(&failure)
		if res.StatusCode == http.StatusBadRequest && failure.Error == "invalid_grant" {
			return nil, ErrGrantRevoked
		}
		return nil, ErrUnauthorized
	}
	var wire struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int64  `json:"expires_in"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&wire); err != nil {
		return nil, ErrUnauthorized
	}
	if wire.AccessToken == "" || !strings.EqualFold(wire.TokenType, "Bearer") || wire.ExpiresIn <= 0 || wire.ExpiresIn > 31536000 {
		return nil, ErrUnauthorized
	}
	return &Grant{AccessToken: wire.AccessToken, RefreshToken: wire.RefreshToken, TokenType: "Bearer", ExpiresAt: m.now().Add(time.Duration(wire.ExpiresIn) * time.Second)}, nil
}

// AccessToken refreshes within Store.Update so rotating refresh tokens are not
// concurrently redeemed by two requests or two server instances.
func (m *Manager) AccessToken(ctx context.Context, owner Owner, p Profile) (string, error) {
	if err := validate(p); err != nil {
		return "", err
	}
	var token string
	var revoked bool
	err := m.Store.Update(ctx, owner, p.Resource, func(g *Grant) (*Grant, error) {
		if g == nil || g.AccessToken == "" {
			return nil, ErrUnauthorized
		}
		if m.now().Add(30 * time.Second).Before(g.ExpiresAt) {
			token = g.AccessToken
			return g, nil
		}
		if g.RefreshToken == "" {
			return nil, ErrUnauthorized
		}
		next, err := m.exchange(ctx, p.TokenEndpoint, url.Values{"grant_type": {"refresh_token"}, "refresh_token": {g.RefreshToken}, "client_id": {p.ClientID}, "resource": {p.Resource}})
		if errors.Is(err, ErrGrantRevoked) {
			revoked = true
			generation, e := randomString()
			if e != nil {
				return nil, e
			}
			return &Grant{Generation: generation, Epoch: generation}, nil
		}
		if err != nil {
			return nil, err
		}
		if next.RefreshToken == "" {
			next.RefreshToken = g.RefreshToken
		}
		next.Generation = g.Generation
		next.Epoch = g.Epoch
		next.Profile = g.Profile
		next.Pending = g.Pending
		token = next.AccessToken
		return next, nil
	})
	if err == nil && revoked {
		return "", ErrGrantRevoked
	}
	return token, err
}
