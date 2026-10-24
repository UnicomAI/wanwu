package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/UnicomAI/wanwu/api/proto/common"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	mcpoauth "github.com/UnicomAI/wanwu/pkg/mcp-oauth"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

type oauthConfiguration struct {
	StoreDirectory string              `json:"storeDirectory"`
	RelayURL       string              `json:"relayUrl"`
	Providers      []mcpoauth.Provider `json:"providers"`
}
type oauthRuntime struct {
	config  oauthConfiguration
	manager *mcpoauth.Manager
	codec   *mcpoauth.Codec
}

var oauthOnce sync.Once
var oauthInstance *oauthRuntime
var oauthInitError error

func oauthService() (*oauthRuntime, error) {
	oauthOnce.Do(func() {
		path := os.Getenv("WANWU_MCP_OAUTH_CONFIG")
		if path == "" {
			oauthInitError = errors.New("MCP OAuth is not configured by the administrator")
			return
		}
		b, e := os.ReadFile(path)
		if e != nil {
			oauthInitError = e
			return
		}
		var cfg oauthConfiguration
		if e = json.Unmarshal(b, &cfg); e != nil {
			oauthInitError = e
			return
		}
		u, e := url.Parse(cfg.RelayURL)
		if e != nil || u.Scheme != "https" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || !strings.HasSuffix(u.Path, "/mcp/oauth/relay") {
			oauthInitError = mcpoauth.ErrInvalidConfig
			return
		}
		key, e := base64.StdEncoding.DecodeString(os.Getenv("WANWU_MCP_OAUTH_KEY"))
		if e != nil {
			oauthInitError = mcpoauth.ErrInvalidConfig
			return
		}
		codec, e := mcpoauth.NewCodec(key)
		if e != nil {
			oauthInitError = e
			return
		}
		store, e := mcpoauth.NewFileStore(cfg.StoreDirectory, codec)
		if e != nil {
			oauthInitError = e
			return
		}
		seen := map[string]bool{}
		for _, p := range cfg.Providers {
			if p.ID == "" || p.Name == "" || seen[p.ID] || seen[p.Resource] {
				oauthInitError = mcpoauth.ErrInvalidConfig
				return
			}
			seen[p.ID] = true
			seen[p.Resource] = true
			for _, target := range p.Targets {
				u, e := url.Parse(target.Resource)
				if e != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" || target.Name == "" {
					oauthInitError = mcpoauth.ErrInvalidConfig
					return
				}
			}
		}
		oauthInstance = &oauthRuntime{cfg, &mcpoauth.Manager{Store: store}, codec}
	})
	return oauthInstance, oauthInitError
}
func (s *oauthRuntime) provider(id string) (mcpoauth.Provider, error) {
	for _, p := range s.config.Providers {
		if p.ID == id {
			return p, nil
		}
	}
	return mcpoauth.Provider{}, mcpoauth.ErrInvalidConfig
}
func MCPProviders() []map[string]string {
	s, e := oauthService()
	if e != nil {
		return []map[string]string{}
	}
	result := []map[string]string{}
	for _, p := range s.config.Providers {
		result = append(result, map[string]string{"id": p.ID, "name": p.Name, "resource": p.Resource})
	}
	return result
}
func MCPPrepareOAuth(ctx *gin.Context, o mcpoauth.Owner, id string) (string, error) {
	s, e := oauthService()
	if e != nil {
		return "", e
	}
	provider, e := s.provider(id)
	if e != nil {
		return "", e
	}
	p, e := s.manager.Discover(ctx.Request.Context(), provider)
	if e != nil {
		return "", e
	}
	return s.manager.Begin(ctx.Request.Context(), o, p)
}
func MCPCompleteOAuth(ctx *gin.Context, o mcpoauth.Owner, state, code string) error {
	s, e := oauthService()
	if e != nil {
		return e
	}
	return s.manager.Complete(ctx.Request.Context(), o, state, code)
}
func MCPOAuthStatus(ctx *gin.Context, o mcpoauth.Owner, id string) (bool, error) {
	s, e := oauthService()
	if e != nil {
		return false, e
	}
	p, e := s.provider(id)
	if e != nil {
		return false, e
	}
	return s.manager.Status(ctx.Request.Context(), o, p.Resource)
}
func MCPDisconnectOAuth(ctx *gin.Context, o mcpoauth.Owner, id string) error {
	s, e := oauthService()
	if e != nil {
		return e
	}
	p, e := s.provider(id)
	if e != nil {
		return e
	}
	return s.manager.Disconnect(ctx.Request.Context(), o, p.Resource)
}

// MCPConnectOAuth saves an ordinary Streamable HTTP MCP configuration pointing
// at the managed relay. Existing agent, workflow and sandbox consumers therefore
// use the same per-request refresh/revocation checks instead of cached tokens.
func MCPConnectOAuth(ctx *gin.Context, o mcpoauth.Owner, id string) error {
	s, e := oauthService()
	if e != nil {
		return e
	}
	p, e := s.provider(id)
	if e != nil {
		return e
	}
	granted, e := s.manager.Status(ctx.Request.Context(), o, p.Resource)
	if e != nil {
		return e
	}
	if !granted {
		return mcpoauth.ErrUnauthorized
	}
	list, e := mcp.GetCustomMCPList(ctx.Request.Context(), &mcp_service.GetCustomMCPListReq{OrgId: o.Organization, UserId: o.User})
	if e != nil {
		return e
	}
	epoch, e := s.manager.Epoch(ctx.Request.Context(), o, p.Resource)
	if e != nil {
		return e
	}
	targets := p.Targets
	if len(targets) == 0 {
		targets = []mcpoauth.Target{{Name: p.Name, Resource: p.Resource}}
	}
	for _, target := range targets {
		found := false
		for _, item := range list.Infos {
			if item.GetApiAuth().GetApiKeyHeader() == mcpoauth.RelayHeader {
				ref, e := s.codec.ReadReference(item.GetApiAuth().GetApiKeyValue())
				if e == nil && ref.Owner == o && ref.Resource == p.Resource && ref.Target == target.Resource {
					if ref.Epoch != epoch || item.StreamableUrl != s.config.RelayURL {
						reference, e := s.codec.ReferenceGrant(o, p.Resource, target.Resource, epoch)
						if e != nil {
							return e
						}
						auth := &common.ApiAuthWebRequest{AuthType: util.AuthTypeAPIKeyHeader, ApiKeyHeader: mcpoauth.RelayHeader, ApiKeyHeaderPrefix: util.ApiKeyHeaderPrefixCustom, ApiKeyValue: reference}
						_, e = mcp.UpdateCustomMCP(ctx.Request.Context(), &mcp_service.UpdateCustomMCPReq{OrgId: o.Organization, UserId: o.User, McpId: item.McpId, Name: item.Info.Name, Desc: item.Info.Desc, From: item.Info.From, StreamableUrl: s.config.RelayURL, Transport: "streamable", AvatarPath: item.AvatarPath, ApiAuth: auth})
						if e != nil {
							return e
						}
					}
					found = true
					break
				}
			}
		}
		if found {
			continue
		}
		reference, e := s.codec.ReferenceGrant(o, p.Resource, target.Resource, epoch)
		if e != nil {
			return e
		}
		_, e = mcp.CreateCustomMCP(ctx.Request.Context(), &mcp_service.CreateCustomMCPReq{OrgId: o.Organization, UserId: o.User, Name: target.Name, Desc: "OAuth · " + target.Resource, From: "OAuth", StreamableUrl: s.config.RelayURL, Transport: "streamable", ApiAuth: &common.ApiAuthWebRequest{AuthType: util.AuthTypeAPIKeyHeader, ApiKeyHeader: mcpoauth.RelayHeader, ApiKeyHeaderPrefix: util.ApiKeyHeaderPrefixCustom, ApiKeyValue: reference}})
		if e != nil {
			return e
		}
	}
	return nil
}

func MCPRelayOAuth(ctx *gin.Context) {
	s, e := oauthService()
	if e != nil {
		http.Error(ctx.Writer, "MCP OAuth unavailable", 503)
		return
	}
	relay := mcpoauth.Relay{Manager: s.manager, Codec: s.codec, Allowed: func(resource, target string) bool {
		for _, p := range s.config.Providers {
			if p.Resource == resource {
				if target == resource {
					return true
				}
				for _, t := range p.Targets {
					if t.Resource == target {
						return true
					}
				}
			}
		}
		return false
	}}
	relay.ServeHTTP(ctx.Writer, ctx.Request)
}
func disconnectManagedMCP(ctx *gin.Context, o mcpoauth.Owner, auth *common.ApiAuthWebRequest) error {
	if auth.GetApiKeyHeader() != mcpoauth.RelayHeader {
		return nil
	}
	s, e := oauthService()
	if e != nil {
		return e
	}
	ref, e := s.codec.ReadReference(auth.GetApiKeyValue())
	if e != nil || ref.Owner != o {
		return mcpoauth.ErrUnauthorized
	}
	return s.manager.Disconnect(ctx.Request.Context(), o, ref.Resource)
}
