package mcpoauth

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// Disconnect serializes with refresh and code exchange. A pending callback from
// before disconnect cannot restore a revoked connection. Local access is removed
// even if the provider is unavailable; the caller must display that remote
// revocation was not confirmed in that case.
func (m *Manager) Disconnect(ctx context.Context, o Owner, resource string) error {
	var remoteErr error
	err := m.Store.Update(ctx, o, resource, func(g *Grant) (*Grant, error) {
		generation, e := randomString()
		if e != nil {
			return nil, e
		}
		if g != nil && (g.RefreshToken != "" || g.AccessToken != "") {
			if g.Profile.RevocationEndpoint == "" {
				remoteErr = errors.New("local disconnect completed; provider does not advertise revocation")
			} else {
				token, hint := g.RefreshToken, "refresh_token"
				if token == "" {
					token, hint = g.AccessToken, "access_token"
				}
				values := url.Values{"token": {token}, "token_type_hint": {hint}, "client_id": {g.Profile.ClientID}}
				req, e := http.NewRequestWithContext(ctx, http.MethodPost, g.Profile.RevocationEndpoint, strings.NewReader(values.Encode()))
				if e == nil {
					req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					var res *http.Response
					res, e = m.httpClient().Do(req)
					if e == nil {
						res.Body.Close()
						if res.StatusCode != 200 {
							e = ErrUnauthorized
						}
					}
				}
				if e != nil {
					remoteErr = errors.New("local disconnect completed; provider revocation was not confirmed")
				}
			}
		}
		return &Grant{Generation: generation, Epoch: generation}, nil
	})
	if err != nil {
		return err
	}
	return remoteErr
}

// Status never returns any token or PKCE material.
func (m *Manager) Status(ctx context.Context, o Owner, r string) (bool, error) {
	authorized := false
	err := m.Store.Update(ctx, o, r, func(g *Grant) (*Grant, error) {
		if g == nil {
			g = &Grant{}
		}
		authorized = !g.Pending && g.AccessToken != "" && (m.now().Before(g.ExpiresAt) || g.RefreshToken != "")
		return g, nil
	})
	return authorized, err
}
