package mcpoauth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Provider is administrator-controlled, never populated from an authorization
// request. Issuer and metadata URLs are an explicit network trust boundary.
type Target struct {
	Name     string `json:"name"`
	Resource string `json:"resource"`
}
type Provider struct {
	OnDemand            bool     `json:"onDemand,omitempty"`
	Targets             []Target `json:"targets,omitempty"`
	ID                  string   `json:"id"`
	Name                string   `json:"name"`
	Resource            string   `json:"resource"`
	ResourceMetadataURL string   `json:"resourceMetadataUrl"`
	Issuer              string   `json:"issuer"`
	ClientID            string   `json:"clientId"`
	RedirectURI         string   `json:"redirectUri"`
	Scope               string   `json:"scope"`
}

func sameOrigin(a, b string) bool {
	u, e := url.Parse(a)
	v, f := url.Parse(b)
	return e == nil && f == nil && u.Scheme == v.Scheme && u.Host == v.Host
}
func (m *Manager) metadata(ctx context.Context, endpoint string, out interface{}) error {
	req, e := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if e != nil {
		return ErrInvalidConfig
	}
	res, e := m.httpClient().Do(req)
	if e != nil {
		return ErrInvalidConfig
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return ErrInvalidConfig
	}
	if json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(out) != nil {
		return ErrInvalidConfig
	}
	return nil
}

// Discover validates resource metadata and the issuer before using its endpoints.
// Public clients either use a pre-registered client ID or RFC 7591 registration.
func (m *Manager) Discover(ctx context.Context, provider Provider) (Profile, error) {
	p := Profile{Resource: provider.Resource, ClientID: provider.ClientID, RedirectURI: provider.RedirectURI, Scope: provider.Scope}
	if !validHTTPS(provider.Resource) || !validHTTPS(provider.Issuer) || !validHTTPS(provider.RedirectURI) || !validHTTPS(provider.ResourceMetadataURL) || !sameOrigin(provider.Resource, provider.ResourceMetadataURL) {
		return p, ErrInvalidConfig
	}
	var resource struct {
		Resource             string   `json:"resource"`
		AuthorizationServers []string `json:"authorization_servers"`
	}
	if e := m.metadata(ctx, provider.ResourceMetadataURL, &resource); e != nil {
		return p, e
	}
	allowed := false
	for _, issuer := range resource.AuthorizationServers {
		if issuer == provider.Issuer {
			allowed = true
		}
	}
	if resource.Resource != provider.Resource || !allowed {
		return p, ErrInvalidConfig
	}
	issuer, _ := url.Parse(provider.Issuer)
	issuer.Path = "/.well-known/oauth-authorization-server" + strings.TrimRight(issuer.Path, "/")
	issuer.RawQuery = ""
	var metadata struct {
		Issuer        string   `json:"issuer"`
		Authorization string   `json:"authorization_endpoint"`
		Token         string   `json:"token_endpoint"`
		Registration  string   `json:"registration_endpoint"`
		Revocation    string   `json:"revocation_endpoint"`
		PKCE          []string `json:"code_challenge_methods_supported"`
	}
	if e := m.metadata(ctx, issuer.String(), &metadata); e != nil {
		return p, e
	}
	if metadata.Issuer != provider.Issuer {
		return p, ErrInvalidConfig
	}
	pkce := false
	for _, method := range metadata.PKCE {
		if method == "S256" {
			pkce = true
		}
	}
	if !pkce {
		return p, ErrInvalidConfig
	}
	for _, endpoint := range []string{metadata.Authorization, metadata.Token, metadata.Revocation, metadata.Registration} {
		if endpoint != "" && (!validHTTPS(endpoint) || !sameOrigin(provider.Issuer, endpoint)) {
			return p, ErrInvalidConfig
		}
	}
	p.AuthorizationEndpoint = metadata.Authorization
	p.TokenEndpoint = metadata.Token
	p.RevocationEndpoint = metadata.Revocation
	if p.ClientID == "" {
		if metadata.Registration == "" {
			return p, ErrInvalidConfig
		}
		b, _ := json.Marshal(map[string]interface{}{"client_name": "Wanwu", "redirect_uris": []string{p.RedirectURI}, "grant_types": []string{"authorization_code", "refresh_token"}, "response_types": []string{"code"}, "token_endpoint_auth_method": "none"})
		req, e := http.NewRequestWithContext(ctx, http.MethodPost, metadata.Registration, bytes.NewReader(b))
		if e != nil {
			return p, ErrInvalidConfig
		}
		req.Header.Set("Content-Type", "application/json")
		res, e := m.httpClient().Do(req)
		if e != nil {
			return p, ErrInvalidConfig
		}
		defer res.Body.Close()
		var registration struct {
			ClientID   string `json:"client_id"`
			AuthMethod string `json:"token_endpoint_auth_method"`
		}
		if res.StatusCode != 201 && res.StatusCode != 200 {
			return p, ErrInvalidConfig
		}
		if json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&registration) != nil || registration.ClientID == "" || (registration.AuthMethod != "" && registration.AuthMethod != "none") {
			return p, ErrInvalidConfig
		}
		p.ClientID = registration.ClientID
	}
	return p, validate(p)
}
