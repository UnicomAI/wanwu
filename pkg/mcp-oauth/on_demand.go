package mcpoauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	client "github.com/mark3labs/mcp-go/client"
	transport "github.com/mark3labs/mcp-go/client/transport"
	mcp "github.com/mark3labs/mcp-go/mcp"
)

// The model receives only these three definitions. Full provider schemas remain
// server-side until explicitly requested. No catalogue cache crosses owners or
// authorization epochs. Each operation uses a fresh upstream MCP session.
func discoveryTools() []mcp.Tool {
	return []mcp.Tool{
		mcp.NewTool("mcp_search_tools", mcp.WithDescription("Search available tools by name or description. Use short keywords (历史, 专利, 标准), or empty query to browse. Returns summaries only; use mcp_tool_detail before calling."), mcp.WithString("query"), mcp.WithNumber("offset")),
		mcp.NewTool("mcp_tool_detail", mcp.WithDescription("Load the full input schema and description of one discovered tool, on demand."), mcp.WithString("name", mcp.Required())),
		mcp.NewTool("mcp_call_tool", mcp.WithDescription("Execute a discovered tool after reading its definition. Pass its exact name and arguments. Respect its permissions and confirmation requirements; do not retry side effects automatically."), mcp.WithString("name", mcp.Required()), mcp.WithObject("arguments", mcp.Required())),
	}
}

func (r *Relay) serveOnDemand(w http.ResponseWriter, req *http.Request, ref Reference) {
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var in struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      json.RawMessage `json:"id"`
		Method  string          `json:"method"`
		Params  json.RawMessage `json:"params"`
	}
	if json.NewDecoder(http.MaxBytesReader(w, req.Body, 8<<20)).Decode(&in) != nil || in.JSONRPC != "2.0" {
		http.Error(w, "invalid MCP request", 400)
		return
	}
	if len(in.ID) == 0 {
		w.WriteHeader(http.StatusAccepted)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	reply := func(result interface{}, code int, message string) {
		out := map[string]interface{}{"jsonrpc": "2.0", "id": in.ID}
		if code != 0 {
			out["error"] = map[string]interface{}{"code": code, "message": message}
		} else {
			out["result"] = result
		}
		_ = json.NewEncoder(w).Encode(out)
	}
	switch in.Method {
	case "initialize":
		reply(map[string]interface{}{"protocolVersion": "2025-03-26", "serverInfo": map[string]string{"name": "Wanwu on-demand MCP", "version": "1"}, "capabilities": map[string]interface{}{"tools": map[string]interface{}{}}, "instructions": "Discover tools with mcp_search_tools, load selected schemas with mcp_tool_detail, then execute with mcp_call_tool. The remote catalogue is dynamic and may include historical data."}, 0, "")
		return
	case "ping":
		reply(map[string]interface{}{}, 0, "")
		return
	case "tools/list":
		reply(map[string]interface{}{"tools": discoveryTools()}, 0, "")
		return
	case "tools/call":
	default:
		reply(nil, -32601, "Method not supported by on-demand connector")
		return
	}
	var call struct {
		Name      string `json:"name"`
		Arguments struct {
			Query     string                 `json:"query"`
			Offset    int                    `json:"offset"`
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments"`
		} `json:"arguments"`
	}
	if json.Unmarshal(in.Params, &call) != nil || (call.Name != "mcp_search_tools" && call.Name != "mcp_tool_detail" && call.Name != "mcp_call_tool") || call.Arguments.Offset < 0 {
		reply(nil, -32602, "Invalid discovery tool arguments")
		return
	}
	ctx, cancel := context.WithTimeout(req.Context(), 60*time.Second)
	defer cancel()
	// Recheck the grant before EVERY upstream HTTP operation, including execution
	// after catalogue discovery. Never forward a revoked capability or stale token.
	hc := r.Manager.httpClient()
	base := hc.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	hc.Transport = authorizedTransport{relay: r, ref: ref, base: base}
	c, e := client.NewStreamableHttpClient(ref.Target, transport.WithHTTPBasicClient(hc))
	if e != nil {
		reply(nil, -32603, "MCP connection unavailable")
		return
	}
	defer c.Close()
	if e = c.Start(ctx); e != nil {
		reply(nil, -32603, "MCP connection unavailable")
		return
	}
	init := mcp.InitializeRequest{}
	init.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	init.Params.ClientInfo = mcp.Implementation{Name: "Wanwu", Version: "1"}
	if _, e = c.Initialize(ctx, init); e != nil {
		reply(nil, -32603, "MCP initialization failed; check authorization")
		return
	}
	tools, e := readCatalogue(ctx, c)
	if e != nil {
		reply(nil, -32603, "MCP catalogue unavailable or exceeds safety limit")
		return
	}
	if call.Name == "mcp_search_tools" {
		type summary struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		matches := []summary{}
		terms := strings.Fields(strings.ToLower(call.Arguments.Query))
		for _, tool := range tools {
			text := strings.ToLower(tool.Name + " " + tool.Description)
			match := true
			for _, term := range terms {
				if !strings.Contains(text, term) {
					match = false
					break
				}
			}
			if match {
				d := []rune(tool.Description)
				if len(d) > 300 {
					d = d[:300]
				}
				matches = append(matches, summary{tool.Name, string(d)})
			}
		}
		start := call.Arguments.Offset
		if start > len(matches) {
			start = len(matches)
		}
		end := start + 10
		if end > len(matches) {
			end = len(matches)
		}
		payload := map[string]interface{}{"tools": matches[start:end], "totalMatches": len(matches), "catalogueSize": len(tools)}
		if end < len(matches) {
			payload["nextOffset"] = end
		}
		b, _ := json.Marshal(payload)
		reply(mcp.NewToolResultText(string(b)), 0, "")
		return
	}
	for _, tool := range tools {
		if tool.Name == call.Arguments.Name {
			if call.Name == "mcp_tool_detail" {
				b, _ := json.Marshal(tool)
				reply(mcp.NewToolResultText(string(b)), 0, "")
				return
			}
			execute := mcp.CallToolRequest{}
			execute.Params.Name = tool.Name
			execute.Params.Arguments = call.Arguments.Arguments
			result, err := c.CallTool(ctx, execute)
			if err != nil {
				reply(nil, -32603, "MCP tool call failed; do not retry side effects automatically")
				return
			}
			reply(result, 0, "")
			return
		}
	}
	reply(nil, -32602, "Tool not in current authorized catalogue; search again")
}

type authorizedTransport struct {
	relay *Relay
	ref   Reference
	base  http.RoundTripper
}

func (t authorizedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.String() != t.ref.Target {
		return nil, ErrUnauthorized
	}
	var p Profile
	err := t.relay.Manager.Store.Update(req.Context(), t.ref.Owner, t.ref.Resource, func(g *Grant) (*Grant, error) {
		if g == nil || g.AccessToken == "" || g.Epoch != t.ref.Epoch {
			return nil, ErrUnauthorized
		}
		p = g.Profile
		return g, nil
	})
	if err != nil {
		return nil, err
	}
	token, err := t.relay.Manager.AccessToken(req.Context(), t.ref.Owner, p)
	if err != nil {
		return nil, err
	}
	err = t.relay.Manager.Store.Update(req.Context(), t.ref.Owner, t.ref.Resource, func(g *Grant) (*Grant, error) {
		if g == nil || g.AccessToken == "" || g.Epoch != t.ref.Epoch {
			return nil, ErrUnauthorized
		}
		return g, nil
	})
	if err != nil {
		return nil, err
	}
	clone := req.Clone(req.Context())
	clone.Header = req.Header.Clone()
	clone.Header.Set("Authorization", "Bearer "+token)
	res, err := t.base.RoundTrip(clone)
	if err == nil && res.StatusCode == 401 {
		_ = t.relay.Manager.Store.Update(req.Context(), t.ref.Owner, t.ref.Resource, func(g *Grant) (*Grant, error) {
			if g != nil && g.AccessToken == token {
				g.ExpiresAt = time.Time{}
			}
			return g, nil
		})
	}
	return res, err
}
func readCatalogue(ctx context.Context, c *client.Client) ([]mcp.Tool, error) {
	tools := []mcp.Tool{}
	seen := map[string]bool{}
	cursors := map[mcp.Cursor]bool{}
	req := mcp.ListToolsRequest{}
	for page := 0; page < 100; page++ {
		result, err := c.ListTools(ctx, req)
		if err != nil {
			return nil, err
		}
		for _, tool := range result.Tools {
			if tool.Name == "" || seen[tool.Name] {
				return nil, errors.New("ambiguous tool catalogue")
			}
			seen[tool.Name] = true
			tools = append(tools, tool)
		}
		if len(tools) > 5000 {
			return nil, errors.New("catalogue limit")
		}
		if result.NextCursor == "" {
			return tools, nil
		}
		if cursors[result.NextCursor] {
			return nil, errors.New("cursor loop")
		}
		cursors[result.NextCursor] = true
		req.Params.Cursor = result.NextCursor
	}
	return nil, errors.New("catalogue pagination limit")
}
