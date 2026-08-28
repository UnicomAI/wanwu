package mcp2skill

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ThinkInAIXYZ/go-mcp/client"
	"github.com/ThinkInAIXYZ/go-mcp/protocol"
	"github.com/ThinkInAIXYZ/go-mcp/transport"
)

// MCPConfig represents the JSON input configuration for converting an MCP server to a skill.
// It is designed to be stateless and self-contained — no gRPC or service dependencies needed.
//
// Example JSON:
//
//	{
//	  "name": "天气查询streamable",
//	  "description": "根据地点获取当前的天气情况",
//	  "streamableUrl": "http://192.168.0.21:8081/service/api/openapi/v1/mcp/server/streamable",
//	  "sseUrl": "",
//	  "transport": "streamable",
//	  "apiAuth": {
//	    "authType": "api_key_query",
//	    "apiKeyQueryParam": "key",
//	    "apiKeyValue": "ww-xxx"
//	  },
//	  "headers": {
//	    "X-Custom": "value"
//	  }
//	}
type MCPConfig struct {
	// Name is the skill name (e.g. MCP info name). Used as the skill directory name.
	Name string `json:"name"`
	// Description is an optional skill description override.
	// If empty, one is auto-generated from the tool list.
	Description string `json:"description,omitempty"`
	// StreamableUrl is the MCP server streamable HTTP endpoint.
	StreamableUrl string `json:"streamableUrl,omitempty"`
	// SseUrl is the MCP server SSE endpoint.
	SseUrl string `json:"sseUrl,omitempty"`
	// Transport is the transport type: "streamable" or "sse". Defaults to "streamable".
	Transport string `json:"transport,omitempty"`
	// ApiAuth is the optional API authentication configuration.
	ApiAuth *APIAuthConfig `json:"apiAuth,omitempty"`
	// Headers are optional custom HTTP headers to include in requests.
	Headers map[string]string `json:"headers,omitempty"`
}

// ConvertFromConfig reads a JSON config file, connects to the MCP server,
// fetches tools, and writes the skill output.
func ConvertFromConfig(ctx context.Context, configPath string, outputDir string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg MCPConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config JSON: %w", err)
	}

	return ConvertFromMCPConfig(ctx, &cfg, outputDir)
}

// ConvertFromMCPConfig connects to the MCP server described by MCPConfig,
// fetches tools, and writes the skill output.
// ApiAuth and Headers from the config are used for the connection but stripped
// from the generated output — users must fill in their own credentials.
func ConvertFromMCPConfig(ctx context.Context, cfg *MCPConfig, outputDir string) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	// Determine server URL based on transport type.
	transportType := cfg.Transport
	if transportType == "" {
		transportType = "streamable"
	}

	serverURL := cfg.StreamableUrl
	if transportType == "sse" {
		serverURL = cfg.SseUrl
	}
	if serverURL == "" {
		// Fallback: try the other URL
		serverURL = cfg.SseUrl
	}
	if serverURL == "" {
		return fmt.Errorf("no server URL available in config (need streamableUrl or sseUrl)")
	}

	// Merge auth params into connection URL and extract headers.
	connURL, connHeaders, err := mergeAuthParams(serverURL, cfg.ApiAuth, cfg.Headers)
	if err != nil {
		return fmt.Errorf("failed to merge auth params: %w", err)
	}

	// Connect with the full URL (including merged query params) and headers.
	config := MCPServerConfig{
		URL:           connURL,
		TransportType: transportType,
		Headers:       connHeaders,
	}

	// Derive skill name.
	skillName := ""
	if cfg.Name != "" {
		skillName = toFileName(cfg.Name)
	}
	if skillName == "" {
		skillName = "mcp-skill"
	}

	// Build display URL (without credentials) for the generated output.
	displayURL := buildDisplayURL(serverURL, cfg.ApiAuth)

	return ConvertFromServer(ctx, config, ConvertOptions{
		OutputDir:     outputDir,
		SkillName:     skillName,
		Description:   cfg.Description,
		ServerURL:     displayURL,
		TransportType: transportType,
		ApiAuth:       cfg.ApiAuth,
	})
}

// ConvertFromServer connects to an MCP server, fetches tools, and writes
// the skill output to the specified directory.
func ConvertFromServer(ctx context.Context, config MCPServerConfig, opts ConvertOptions) error {
	tools, err := listToolsFromServer(ctx, config)
	if err != nil {
		return fmt.Errorf("failed to list tools from MCP server: %w", err)
	}

	if opts.ServerURL == "" {
		opts.ServerURL = maskURLKey(config.URL)
	}
	if opts.TransportType == "" {
		opts.TransportType = config.TransportType
	}

	return ConvertFromTools(tools, opts)
}

// ConvertFromTools converts a list of MCP tools into the Skill format and
// writes the output files.
func ConvertFromTools(tools []*protocol.Tool, opts ConvertOptions) error {
	skillDoc := Parse(tools, opts.SkillName)

	if opts.Description != "" {
		skillDoc.Meta.Description = opts.Description
	}

	skillDoc.ServerInfo = ServerInfoDocument{
		URL:           opts.ServerURL,
		TransportType: opts.TransportType,
		AuthHeader:    buildAuthHeaderDescription(opts.ApiAuth),
	}

	renderer := NewRenderer()
	return writeSkillOutput(skillDoc, opts.OutputDir, renderer)
}

// ParseToIR parses MCP tools and returns the SkillDocument IR without
// writing any files.
func ParseToIR(tools []*protocol.Tool, skillName string) SkillDocument {
	return Parse(tools, skillName)
}

// =============================================================================
// MCP Server connection
// =============================================================================

var insecureHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
}

// headerTransport is an http.RoundTripper wrapper for injecting custom headers.
type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for key, value := range t.headers {
		req.Header.Set(key, value)
	}
	return t.base.RoundTrip(req)
}

// newHTTPClientWithHeaders creates an http.Client with custom headers injected.
func newHTTPClientWithHeaders(headers map[string]string) *http.Client {
	if len(headers) == 0 {
		return insecureHTTPClient
	}
	return &http.Client{
		Transport: &headerTransport{
			base: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
			headers: headers,
		},
	}
}

func listToolsFromServer(ctx context.Context, config MCPServerConfig) ([]*protocol.Tool, error) {
	if config.URL == "" {
		return nil, fmt.Errorf("MCP server URL is required")
	}

	httpClient := newHTTPClientWithHeaders(config.Headers)

	var transportClient transport.ClientTransport
	var err error

	switch config.TransportType {
	case "streamable":
		transportClient, err = transport.NewStreamableHTTPClientTransport(config.URL,
			transport.WithStreamableHTTPClientOptionHTTPClient(httpClient),
		)
	case "sse":
		transportClient, err = transport.NewSSEClientTransport(config.URL,
			transport.WithSSEClientOptionReceiveTimeout(time.Minute*2),
			transport.WithSSEClientOptionHTTPClient(httpClient),
		)
	default:
		return nil, fmt.Errorf("unsupported transport type: %s (use 'sse' or 'streamable')", config.TransportType)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %w", err)
	}

	mcpClient, err := client.NewClient(transportClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP client: %w", err)
	}
	defer func() { _ = mcpClient.Close() }()

	resp, err := mcpClient.ListTools(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %w", err)
	}

	return resp.Tools, nil
}

// =============================================================================
// URL key masking
// =============================================================================

// sensitiveParams are query parameter names whose values should be masked
// in generated output. Users must fill in their own values before using the skill.
var sensitiveParams = []string{"key", "token", "secret", "apikey", "api_key"}

// maskURLKey replaces sensitive query parameter values with <YOUR_KEY> placeholder
// so users know they need to provide their own credentials.
func maskURLKey(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	if u.RawQuery == "" {
		return rawURL
	}

	// Replace sensitive param values directly in the raw query string
	// to preserve parameter order and avoid URL-encoding < and >.
	result := u.RawQuery
	for _, param := range sensitiveParams {
		// Match the param value: key=<value>& or key=<value> at end
		// Use a simple approach: find key= and replace the value until & or end.
		result = maskParamInQuery(result, param)
	}
	u.RawQuery = result
	return u.String()
}

// maskParamInQuery replaces the value of a sensitive parameter in a query string
// with <YOUR_KEY>, preserving parameter order and other values.
func maskParamInQuery(query, param string) string {
	prefix := param + "="
	start := 0
	for {
		idx := strings.Index(query[start:], prefix)
		if idx == -1 {
			break
		}
		idx += start
		valueStart := idx + len(prefix)
		valueEnd := strings.IndexByte(query[valueStart:], '&')
		if valueEnd == -1 {
			// Last parameter
			query = query[:valueStart] + "<YOUR_KEY>"
			break
		}
		query = query[:valueStart] + "<YOUR_KEY>" + query[valueStart+valueEnd:]
		start = valueStart + len("<YOUR_KEY>")
	}
	return query
}

// =============================================================================
// Auth params merging
// =============================================================================

// mergeAuthParams merges API auth and custom headers into a connection URL
// and headers map. Query auth params are merged into the URL, header auth
// params and custom headers are returned as a combined headers map.
func mergeAuthParams(serverURL string, apiAuth *APIAuthConfig, headers map[string]string) (string, map[string]string, error) {
	connHeaders := make(map[string]string)
	for k, v := range headers {
		connHeaders[k] = v
	}

	if apiAuth == nil || apiAuth.AuthType == "" || apiAuth.AuthType == AuthTypeNone {
		return serverURL, connHeaders, nil
	}

	switch apiAuth.AuthType {
	case AuthTypeAPIKeyQuery:
		rawUrl, err := url.Parse(serverURL)
		if err != nil {
			return "", nil, fmt.Errorf("parse url err: %w", err)
		}
		q := rawUrl.Query()
		q.Set(apiAuth.ApiKeyQueryParam, apiAuth.ApiKeyValue)
		rawUrl.RawQuery = q.Encode()
		return rawUrl.String(), connHeaders, nil

	case AuthTypeAPIKeyHeader:
		value := apiAuth.ApiKeyValue
		switch apiAuth.ApiKeyHeaderPrefix {
		case ApiKeyHeaderPrefixBasic:
			value = "Basic " + apiAuth.ApiKeyValue
		case ApiKeyHeaderPrefixBearer:
			value = "Bearer " + apiAuth.ApiKeyValue
		}
		headerName := apiAuth.ApiKeyHeader
		if headerName == "" {
			headerName = ApiKeyHeaderDefault
		}
		connHeaders[headerName] = value
		return serverURL, connHeaders, nil
	}

	return serverURL, connHeaders, nil
}

// buildDisplayURL builds the URL to display in the generated skill output.
// It strips credential info so users know what they need to fill in.
func buildDisplayURL(serverURL string, apiAuth *APIAuthConfig) string {
	if apiAuth == nil || apiAuth.AuthType == "" || apiAuth.AuthType == AuthTypeNone {
		return maskURLKey(serverURL)
	}

	switch apiAuth.AuthType {
	case AuthTypeAPIKeyQuery:
		// Use maskParamInQuery to replace the query param value with <YOUR_KEY>,
		// preserving the raw query string to avoid URL-encoding angle brackets.
		u, err := url.Parse(serverURL)
		if err != nil {
			return serverURL
		}
		if u.RawQuery == "" {
			// No query params yet, add the placeholder
			u.RawQuery = apiAuth.ApiKeyQueryParam + "=<YOUR_KEY>"
		} else {
			u.RawQuery = maskParamInQuery(u.RawQuery, apiAuth.ApiKeyQueryParam)
		}
		return u.String()
	case AuthTypeAPIKeyHeader:
		// Header auth doesn't affect URL, just mask any existing sensitive params.
		return maskURLKey(serverURL)
	}
	return maskURLKey(serverURL)
}

// buildAuthHeaderDescription builds a human-readable description of header-based auth.
func buildAuthHeaderDescription(apiAuth *APIAuthConfig) string {
	if apiAuth == nil || apiAuth.AuthType == "" || apiAuth.AuthType == AuthTypeNone {
		return ""
	}
	if apiAuth.AuthType == AuthTypeAPIKeyHeader {
		prefix := apiAuth.ApiKeyHeaderPrefix
		if prefix == "" {
			prefix = "custom"
		}
		headerName := apiAuth.ApiKeyHeader
		if headerName == "" {
			headerName = ApiKeyHeaderDefault
		}
		switch prefix {
		case "bearer":
			return headerName + ": Bearer <YOUR_TOKEN>"
		case "basic":
			return headerName + ": Basic <YOUR_CREDENTIALS>"
		default:
			return headerName + ": <YOUR_VALUE>"
		}
	}
	return ""
}

func writeSkillOutput(doc SkillDocument, outputDir string, renderer *Renderer) error {
	skillDir := filepath.Join(outputDir, doc.Meta.Name)
	scriptsDir := filepath.Join(skillDir, "scripts")
	refsDir := filepath.Join(skillDir, "references")
	opsDir := filepath.Join(refsDir, "operations")

	// Create directories.
	dirs := []string{skillDir, scriptsDir, opsDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Write SKILL.md.
	skillMd := renderer.RenderSkill(doc)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMd), 0o644); err != nil {
		return fmt.Errorf("failed to write SKILL.md: %w", err)
	}

	// Write mcp_client.py.
	clientPy := renderer.RenderMCPClient(doc)
	if err := os.WriteFile(filepath.Join(scriptsDir, "mcp_client.py"), []byte(clientPy), 0o644); err != nil {
		return fmt.Errorf("failed to write mcp_client.py: %w", err)
	}

	// Write operation files.
	for _, tool := range doc.Tools {
		opMd := renderer.RenderOperation(tool)
		fileName := toFileName(tool.Name)
		if err := os.WriteFile(filepath.Join(opsDir, fileName+".md"), []byte(opMd), 0o644); err != nil {
			return fmt.Errorf("failed to write operation %s: %w", fileName, err)
		}
	}

	return nil
}
