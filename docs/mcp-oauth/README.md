# Remote MCP OAuth (Streamable HTTP)

This feature adds a browser authorization entry in **MCP → custom/imported MCP**.
Users authorize at their provider; they never paste an API key or bearer token.
A separate provider-card change can reuse the same component and APIs.

## Deployment

This first implementation supports **one BFF host**, including multiple BFF
processes using the same **local POSIX directory**. Independent replica volumes,
NFS and multi-host deployments are **not supported**. A database-backed Store
implementation is needed before enabling it in a multi-host deployment. The
Store interface explicitly requires cross-process serialized refresh/delete.

1. Copy `config.example.json` to a private, administrator-managed configuration
   file. Set `WANWU_MCP_OAUTH_CONFIG` to that file's absolute path.
2. Supply a random 32-byte key, base64 encoded, through the deployment secret
   mechanism as `WANWU_MCP_OAUTH_KEY`. Do not commit the key. Preserve the same
   key and volume across restarts. Losing/changing the key requires a new empty store directory and reauthorization;
   old ciphertext cannot be reused without its original key.
3. Mount `storeDirectory` persistently, owned by the BFF service account, mode
   **0700**. Files use **0600**, AES-256-GCM, and owner/resource associated data.
   Include the encrypted volume and key in separate secure backup procedures.
4. Publish the BFF callback listener's
   `/callback/v1/mcp/oauth/relay` through an **HTTPS** reverse-proxy location matching
   `relayUrl`. It must be reachable from the BFF, agent service, workflow execution
   environment and WGA sandbox. Disable proxy buffering for SSE responses, strip
   query/header/body logging on this location, and keep other internal callback
   routes private. The relay authenticates an encrypted capability; it does not
   use the browser's login cookie.
5. Publish the web callback at `<web-base>/oauth/mcp/callback` (normally
   `/aibase/oauth/mcp/callback`). Configure that exact HTTPS URL as `redirectUri`.
   Preserve SPA history fallback and set `Referrer-Policy: no-referrer`; do not log
   callback query strings at the edge. This endpoint is a frontend page; it posts
   the code/state using the initiating user's authenticated Wanwu session.
6. Restart BFF after configuration changes. With no OAuth config, existing MCP
   auth continues working and the optional OAuth panel is hidden.

Provider resource, resource metadata and issuer are administrator allowlists.
Discovery verifies the canonical resource/issuer, S256 and same-origin provider
endpoints. A pre-registered public `clientId` may be supplied; otherwise RFC 7591
registration uses client name **Wanwu**, `token_endpoint_auth_method=none` and the
configured callback. Confidential client authentication is not supported yet.
Each additional `targets` entry is an explicit provider-documented shared-grant
resource, never automatically inferred from matching hostnames.

## User flow

Open Wanwu in the system browser and sign in. Select **一键授权 / Authorize**, then
use the prominent **打开登录授权页 / Open authorization** link. It is a direct link,
not `window.open` after an async request, so popup blocking does not silently
lose the authorization window. Embedded-browser users can copy the login link.
If the system browser is not signed in to Wanwu, the callback offers a sign-in
link and an explicit completion button. Use the same account and organization.
The original panel polls for completion, saves the connector and reports the
result. Expired or cancelled authorization can be restarted. **Save connection**
retries an interrupted save without another login.

**Disconnect** revokes the refresh token (or access token if no refresh token was
issued), removes local access, invalidates prior relay capabilities and pending
callbacks, and retains a tombstone to prevent resurrection. Reauthorization plus
Save updates existing connector references. A failed remote revocation is reported;
local access remains disabled. Calls already in flight may finish. Deleting a
managed MCP also disconnects its grant; for a shared-grant bundle this disables
all of that bundle's resources. The UI warns before deleting or disconnecting.

## Why a managed relay

Wanwu has multiple MCP implementations and a subprocess sandbox that consumes
static HTTP configuration. Resolving a token only once during tool discovery
would leave long sessions using stale credentials. The managed relay applies
refresh and revocation checks to **every request** across these existing clients.

The stored connector uses the existing HTTP-header schema for an **internal,
resource-scoped encrypted capability**, not a provider API key. Provider access
and refresh tokens never leave the BFF credential store/HTTP client. The browser
receives `authType: oauth`, not the capability. Capabilities are also secrets:
protect internal gRPC, sandbox configuration files and backups. OAuth connector
export as standalone Skill is blocked to prevent exporting its capability.

The relay forwards only the explicitly configured target URL and allowlisted MCP
headers, never browser cookies or its own capability. Redirects are rejected.
It preserves streaming responses and MCP session headers. A 401 invalidates the
current access token but does **not replay tools/call**; the next request attempts
refresh. It does not add automatic retries to side-effecting tools. Legacy SSE
OAuth is unsupported; existing SSE/API-key/no-auth connectors are unchanged.

## Validation

Run `go test -race ./pkg/mcp-oauth`. The tests include PKCE/state ownership and
replay, ciphertext tampering, persistence/restart, independent-process file-lock
serialization, concurrent token refresh, disconnect/late callback, trust-boundary
checks, and both MCP SDKs actually listing/calling a mock OAuth-protected tool
through the relay.

Run the Linux BFF/agent/sandbox builds and the web production build in the normal
project toolchain. A Linux cross-compile is not a running-service acceptance test.
Before production rollout, additionally validate a deployed multi-service instance:

- fresh provider login, callback in a signed-in system browser;
- tools/list and tools/call from the UI, normal agent, workflow and sandbox;
- token expiry during a session, concurrent refresh, BFF restart;
- unlink and reauthorize, rejecting old capability copies;
- account/organization mismatch, provider denial, network and save failure.

No real account credentials or query results belong in tests, screenshots or PRs.

### Execution identities

Public tool probing and attaching a managed connector require its owning user and
organization. The normal agent runtime checks the connector against the agent
creator; publishing/sharing that agent therefore delegates use through the agent,
as with its other configured resources. WGA checks the calling user's ownership.
The existing isolated workflow callback returns the runtime capability; it must
remain on the private service network and must not log response bodies. Only the
new relay path is exposed through the public HTTPS reverse proxy.

### Local validation environment notes

The repository specifies Node 22.10.0 and pnpm 9.15.0. Use that package-manager
version and the checked-in lockfile; a newer local pnpm initially rejected the
lockfile and resolved incompatible older direct dependencies without it. This is
a local tooling mismatch, not an OAuth code failure. No dependency or lockfile
changes are part of this feature. UI controller tests can be run with
`node --test web/tests/mcp-oauth.test.cjs` after installing the web dependencies. On a strict pnpm node_modules layout, the
existing build config imports hoisted webpack; use
`NODE_PATH=./node_modules/.pnpm/node_modules` from `web/`. The local newer Node
version also needs `NODE_OPTIONS=--openssl-legacy-provider`.

The existing gopsutil v2 dependency does not compile against the local macOS SDK.
Core OAuth tests run natively; BFF/agent/sandbox compile validation targets Linux.
The added GitHub Actions job runs the BFF ownership/redaction and sandbox config
tests on Linux. Real-provider login and full deployed multi-service acceptance
remain separate release checks; mock protocol tests are not evidence of those.
