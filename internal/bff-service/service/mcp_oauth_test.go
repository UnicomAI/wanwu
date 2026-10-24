package service

import (
	"github.com/UnicomAI/wanwu/api/proto/common"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	mcpoauth "github.com/UnicomAI/wanwu/pkg/mcp-oauth"
	"github.com/UnicomAI/wanwu/pkg/util"
	"testing"
)

func TestOAuthReferenceIsInternalOnly(t *testing.T) {
	auth := &common.ApiAuthWebRequest{AuthType: util.AuthTypeAPIKeyHeader, ApiKeyHeader: mcpoauth.RelayHeader, ApiKeyHeaderPrefix: util.ApiKeyHeaderPrefixCustom, ApiKeyValue: "mock-capability"}
	public := toApiAuthResponse(auth)
	if public.AuthType != "oauth" || public.ApiKeyValue != "" || public.ApiKeyHeader != "" {
		t.Fatal("capability exposed to browser")
	}
	internal := toApiAuthPtr(auth)
	if internal.ApiKeyValue != "mock-capability" {
		t.Fatal("runtime capability was stripped")
	}
	if !reservedOAuthAuth(internal, nil) || !reservedOAuthAuth(nil, map[string]string{"x-wanwu-mcp-oauth": "forged"}) {
		t.Fatal("reserved header can be supplied by users")
	}
	static := &common.ApiAuthWebRequest{AuthType: util.AuthTypeAPIKeyHeader, ApiKeyHeader: "Authorization", ApiKeyValue: "legacy-key"}
	if toApiAuthResponse(static).ApiKeyValue != "legacy-key" {
		t.Fatal("legacy auth changed")
	}
}

func TestManagedOwnerCannotBindAnotherUsersGrant(t *testing.T) {
	auth := &common.ApiAuthWebRequest{ApiKeyHeader: mcpoauth.RelayHeader}
	owner := &mcp_service.Identity{UserId: "alice", OrgId: "one"}
	if checkManagedOwner(auth, owner, "alice", "one") != nil {
		t.Fatal("valid owner rejected")
	}
	if checkManagedOwner(auth, owner, "bob", "one") == nil || checkManagedOwner(auth, owner, "alice", "two") == nil {
		t.Fatal("cross-owner grant accepted")
	}
}
