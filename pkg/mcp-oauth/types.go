// Package mcpoauth provides the authorization-code state machine used by remote
// MCP clients. Persistence and ownership checks belong to the hosting service.
package mcpoauth

import (
	"context"
	"errors"
	"net/http"
	"time"
)

var (
	ErrGrantRevoked  = errors.New("MCP OAuth grant revoked; reauthorization required")
	ErrUnauthorized  = errors.New("MCP OAuth authorization required")
	ErrInvalidState  = errors.New("invalid or expired MCP OAuth state")
	ErrInvalidConfig = errors.New("invalid MCP OAuth configuration")
)

// Owner must be derived from the authenticated request, never from request JSON.
type Owner struct{ Organization, User string }
type Profile struct {
	Resource              string
	AuthorizationEndpoint string
	TokenEndpoint         string
	RevocationEndpoint    string
	ClientID              string
	RedirectURI           string
	Scope                 string
}
type Grant struct {
	Epoch        string    `json:"epoch"`
	Pending      bool      `json:"pending"`
	Generation   string    `json:"generation"`
	Profile      Profile   `json:"profile"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresAt    time.Time `json:"expires_at"`
}
type Pending struct {
	Generation string
	Owner      Owner
	Profile    Profile
	Verifier   string
	ExpiresAt  time.Time
}

// Store implementations must encrypt grants at rest and serialize Update across
// every process using the same grant. Delete and Update must share the same lock.
// Take consumes a pending state atomically only for the supplied owner.
type Store interface {
	PutPending(context.Context, string, Pending) error
	Take(context.Context, string, Owner) (Pending, error)
	Update(context.Context, Owner, string, func(*Grant) (*Grant, error)) error
	Delete(context.Context, Owner, string) error
}
type Manager struct {
	Store  Store
	Client *http.Client
	Now    func() time.Time
}

func (m *Manager) now() time.Time {
	if m.Now != nil {
		return m.Now()
	}
	return time.Now()
}
