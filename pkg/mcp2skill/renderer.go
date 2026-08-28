package mcp2skill

import (
	"fmt"
	"sort"
	"strings"
)

// Renderer converts SkillDocument IR to output strings.
type Renderer struct{}

// NewRenderer creates a new Renderer.
func NewRenderer() *Renderer {
	return &Renderer{}
}

// RenderSkill renders the top-level SKILL.md content.
func (r *Renderer) RenderSkill(doc SkillDocument) string {
	var b strings.Builder

	// Frontmatter
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", yamlQuote(doc.Meta.Name))
	fmt.Fprintf(&b, "description: %s\n", yamlQuote(doc.Meta.Description))
	b.WriteString("---\n\n")

	// Title
	fmt.Fprintf(&b, "# %s\n\n", doc.Meta.Name)

	if doc.Meta.Description != "" {
		b.WriteString(doc.Meta.Description + "\n\n")
	}

	// Server info
	if doc.ServerInfo.URL != "" {
		b.WriteString("## MCP Server\n\n")
		fmt.Fprintf(&b, "- **URL:** `%s`\n", doc.ServerInfo.URL)
		if doc.ServerInfo.TransportType != "" {
			fmt.Fprintf(&b, "- **Transport:** %s\n", doc.ServerInfo.TransportType)
		}
		b.WriteString("\n")
	}

	// Authentication section
	if strings.Contains(doc.ServerInfo.URL, "<YOUR_KEY>") {
		b.WriteString("## Authentication\n\n")
		b.WriteString("**This MCP server requires an API key.** The server URL contains `<YOUR_KEY>` as a placeholder.\n")
		b.WriteString("You **must** replace `<YOUR_KEY>` with your actual API key before calling any tool.\n\n")
		b.WriteString("Example:\n```\n")
		fmt.Fprintf(&b, "# Replace <YOUR_KEY> with your real key\n")
		fmt.Fprintf(&b, "python3 scripts/mcp_client.py --server-url %s --transport %s --tool <tool-name> --arguments '{...}'\n",
			strings.ReplaceAll(doc.ServerInfo.URL, "<YOUR_KEY>", "YOUR_ACTUAL_KEY"), doc.ServerInfo.TransportType)
		b.WriteString("```\n\n")
	} else if doc.ServerInfo.AuthHeader != "" {
		b.WriteString("## Authentication\n\n")
		b.WriteString("**This MCP server requires authentication via HTTP headers.**\n\n")
		fmt.Fprintf(&b, "You **must** set the following header in every request:\n\n")
		b.WriteString("```\n")
		fmt.Fprintf(&b, "%s\n", doc.ServerInfo.AuthHeader)
		b.WriteString("```\n\n")
		b.WriteString("Example:\n```\n")
		fmt.Fprintf(&b, "# Set the required header\n")
		fmt.Fprintf(&b, "python3 scripts/mcp_client.py --server-url \"%s\" --transport %s --tool <tool-name> --arguments '{...}'\n",
			doc.ServerInfo.URL, doc.ServerInfo.TransportType)
		b.WriteString("```\n\n")
	}

	// How to Use
	b.WriteString("## How to Use This Skill\n\n")
	b.WriteString("This skill provides access to MCP tools via a Python client script.\n\n")
	b.WriteString("**Directory structure:**\n```\n")
	b.WriteString("scripts/\n")
	b.WriteString("└── mcp_client.py    # MCP client for calling tools\n")
	b.WriteString("references/\n")
	b.WriteString("└── operations/      # per-tool detail docs\n")
	b.WriteString("```\n\n")
	b.WriteString("**Usage flow:**\n")
	b.WriteString("1. Find the tool you need in the list below\n")
	b.WriteString("2. Read `references/operations/<tool>.md` for full parameter details\n")
	b.WriteString("3. Call the tool using `python3 scripts/mcp_client.py --tool <name> --arguments '<json>'`\n\n")

	// Python client section
	b.WriteString("## Quick Start\n\n")
	b.WriteString("```bash\n")
	b.WriteString("# List available tools\n")
	if doc.ServerInfo.AuthHeader != "" {
		fmt.Fprintf(&b, "python3 scripts/mcp_client.py --server-url \"%s\" --transport %s --headers '{\"Authorization\": \"<YOUR_TOKEN>\"}' --list\n", doc.ServerInfo.URL, doc.ServerInfo.TransportType)
	} else if strings.Contains(doc.ServerInfo.URL, "<YOUR_KEY>") {
		fmt.Fprintf(&b, "python3 scripts/mcp_client.py --server-url %s --transport %s --list\n", strings.ReplaceAll(doc.ServerInfo.URL, "<YOUR_KEY>", "YOUR_ACTUAL_KEY"), doc.ServerInfo.TransportType)
	} else {
		fmt.Fprintf(&b, "python3 scripts/mcp_client.py --server-url \"%s\" --transport %s --list\n", doc.ServerInfo.URL, doc.ServerInfo.TransportType)
	}
	b.WriteString("\n# Call a tool\n")
	if doc.ServerInfo.AuthHeader != "" {
		fmt.Fprintf(&b, "python3 scripts/mcp_client.py --server-url \"%s\" --transport %s --headers '{\"Authorization\": \"<YOUR_TOKEN>\"}' --tool <tool-name> --arguments '{...}'\n", doc.ServerInfo.URL, doc.ServerInfo.TransportType)
	} else if strings.Contains(doc.ServerInfo.URL, "<YOUR_KEY>") {
		fmt.Fprintf(&b, "python3 scripts/mcp_client.py --server-url %s --transport %s --tool <tool-name> --arguments '{...}'\n", strings.ReplaceAll(doc.ServerInfo.URL, "<YOUR_KEY>", "YOUR_ACTUAL_KEY"), doc.ServerInfo.TransportType)
	} else {
		fmt.Fprintf(&b, "python3 scripts/mcp_client.py --server-url \"%s\" --transport %s --tool <tool-name> --arguments '{\"key\": \"value\"}'\n", doc.ServerInfo.URL, doc.ServerInfo.TransportType)
	}
	b.WriteString("```\n\n")

	// Tools table
	b.WriteString("## Tools\n\n")
	b.WriteString("| Name | Description | Details |\n")
	b.WriteString("|------|-------------|----------|\n")
	for _, tool := range doc.Tools {
		desc := tool.Description
		runes := []rune(desc)
		if len(runes) > 80 {
			desc = string(runes[:77]) + "..."
		}
		fmt.Fprintf(&b, "| `%s` | %s | [View](references/operations/%s.md) |\n",
			tool.Name, desc, toFileName(tool.Name))
	}

	return b.String()
}

// RenderOperation renders a single tool detail markdown.
func (r *Renderer) RenderOperation(doc ToolDocument) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n\n", doc.Name)

	if doc.Description != "" {
		b.WriteString(doc.Description + "\n\n")
	}

	// Annotations
	if doc.Annotations != nil {
		b.WriteString("## Properties\n\n")
		if doc.Annotations.Title != "" {
			fmt.Fprintf(&b, "- **Title:** %s\n", doc.Annotations.Title)
		}
		if doc.Annotations.ReadOnlyHint {
			b.WriteString("- **Read-only:** This tool does not modify its environment\n")
		}
		if doc.Annotations.DestructiveHint {
			b.WriteString("- **Destructive:** This tool may perform destructive operations\n")
		}
		if doc.Annotations.IdempotentHint {
			b.WriteString("- **Idempotent:** Repeated calls with the same arguments have no additional effect\n")
		}
		if doc.Annotations.OpenWorldHint {
			b.WriteString("- **Open-world:** This tool may interact with external entities\n")
		}
		b.WriteString("\n")
	}

	// Parameters
	if len(doc.Parameters) > 0 {
		b.WriteString("## Parameters\n\n")
		requiredSet := make(map[string]bool, len(doc.Required))
		for _, r := range doc.Required {
			requiredSet[r] = true
		}

		b.WriteString("| Name | Type | Required | Description |\n")
		b.WriteString("|------|------|----------|-------------|\n")
		for _, param := range doc.Parameters {
			reqStr := "No"
			if param.Required {
				reqStr = "Yes"
			}
			typeStr := param.Type
			if len(param.Enum) > 0 {
				typeStr = fmt.Sprintf("enum: %s", strings.Join(param.Enum, ", "))
			}
			desc := param.Description
			if param.Default != "" {
				desc += fmt.Sprintf(" (default: %s)", param.Default)
			}
			fmt.Fprintf(&b, "| `%s` | %s | %s | %s |\n", param.Name, typeStr, reqStr, desc)
		}
		b.WriteString("\n")

		// Nested parameters
		for _, param := range doc.Parameters {
			if len(param.Properties) > 0 {
				fmt.Fprintf(&b, "### `%s` Properties\n\n", param.Name)
				b.WriteString("| Name | Type | Required | Description |\n")
				b.WriteString("|------|------|----------|-------------|\n")
				renderNestedProperties(&b, param.Properties, param.Name)
				b.WriteString("\n")
			}
			if param.Items != nil && len(param.Items.Properties) > 0 {
				fmt.Fprintf(&b, "### `%s` Item Properties\n\n", param.Name)
				b.WriteString("| Name | Type | Required | Description |\n")
				b.WriteString("|------|------|----------|-------------|\n")
				renderNestedProperties(&b, param.Items.Properties, param.Name+"[]")
				b.WriteString("\n")
			}
		}
	}

	// Output schema
	if doc.OutputSchema != nil && len(doc.OutputSchema.Properties) > 0 {
		b.WriteString("## Output\n\n")
		b.WriteString("| Name | Type | Description |\n")
		b.WriteString("|------|------|-------------|\n")
		for _, param := range doc.OutputSchema.Properties {
			fmt.Fprintf(&b, "| `%s` | %s | %s |\n", param.Name, param.Type, param.Description)
		}
		b.WriteString("\n")
	}

	// Example usage
	b.WriteString("## Example\n\n")
	fmt.Fprintf(&b, "```bash\npython3 scripts/mcp_client.py --tool %s", doc.Name)
	if len(doc.Parameters) > 0 {
		// Build example arguments with required params
		args := make(map[string]string)
		for _, param := range doc.Parameters {
			if param.Required {
				args[param.Name] = exampleValue(param.Type)
			}
		}
		if len(args) > 0 {
			parts := make([]string, 0, len(args))
			for k, v := range args {
				parts = append(parts, fmt.Sprintf("\"%s\": %s", k, v))
			}
			sort.Strings(parts)
			fmt.Fprintf(&b, " --arguments '{%s}'", strings.Join(parts, ", "))
		}
	}
	b.WriteString("'\n```\n")

	return b.String()
}

func renderNestedProperties(b *strings.Builder, props []ParameterDocument, parent string) {
	for _, prop := range props {
		reqStr := "No"
		if prop.Required {
			reqStr = "Yes"
		}
		typeStr := prop.Type
		if len(prop.Enum) > 0 {
			typeStr = fmt.Sprintf("enum: %s", strings.Join(prop.Enum, ", "))
		}
		desc := prop.Description
		if prop.Default != "" {
			desc += fmt.Sprintf(" (default: %s)", prop.Default)
		}
		fmt.Fprintf(b, "| `%s.%s` | %s | %s | %s |\n", parent, prop.Name, typeStr, reqStr, desc)
	}
}

// RenderMCPClient renders the Python MCP client script.
func (r *Renderer) RenderMCPClient(doc SkillDocument) string {
	var b strings.Builder

	b.WriteString(`#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""MCP Client - Auto-generated by mcp2skill.

Connects to an MCP server and calls tools.

Requirements: mcp==1.27.1
Install: pip install 'mcp[cli]==1.27.1'
"""

import argparse
import asyncio
import json
import sys

try:
    from mcp.client.streamable_http import streamablehttp_client
except ImportError:
    pass
try:
    from mcp.client.sse import sse_client
except ImportError:
    pass
try:
    from mcp.client.session import ClientSession
    from mcp.types import InitializeRequestParams, ClientCapabilities
except ImportError:
    print("Error: mcp==1.27.1 is required. Install with: pip install 'mcp[cli]==1.27.1'", file=sys.stderr)
    sys.exit(1)

`)

	// Generate typed helper functions for each tool
	b.WriteString("# Tool helper functions\n\n")
	for _, tool := range doc.Tools {
		fmt.Fprintf(&b, "def %s(", toPythonFuncName(tool.Name))
		params := make([]string, 0, len(tool.Parameters))
		pythonToOriginal := make(map[string]string) // pythonName -> originalName
		for _, p := range tool.Parameters {
			if p.Required {
				pyName := toPythonParamName(p.Name)
				params = append(params, pyName)
				if pyName != p.Name {
					pythonToOriginal[pyName] = p.Name
				}
			}
		}
		if len(params) > 0 {
			b.WriteString(strings.Join(params, ", "))
			b.WriteString(", ")
		}
		b.WriteString("**kwargs):\n")
		fmt.Fprintf(&b, "    \"\"\"Call the %s tool.\n", tool.Name)
		if tool.Description != "" {
			// Truncate long descriptions at rune boundary to avoid breaking multi-byte UTF-8
			desc := tool.Description
			runes := []rune(desc)
			if len(runes) > 200 {
				desc = string(runes[:197]) + "..."
			}
			fmt.Fprintf(&b, "    %s\n", desc)
		}
		b.WriteString("    \"\"\"\n")
		fmt.Fprintf(&b, "    args = {%s}\n", buildArgsDictWithMapping(tool.Parameters, pythonToOriginal))
		b.WriteString("    args.update(kwargs)\n")
		fmt.Fprintf(&b, "    return call_tool(\"%s\", args)\n\n", tool.Name)
	}

	// Core client functions using mcp SDK
	b.WriteString(`def call_tool(tool_name: str, arguments: dict) -> dict:
    """Call an MCP tool on the server."""
    return asyncio.run(_call_tool(tool_name, arguments))


def list_tools() -> list:
    """List available tools on the MCP server."""
    return asyncio.run(_list_tools())


async def _call_tool(tool_name: str, arguments: dict) -> dict:
    """Async: Call a specific tool on the MCP server."""
    server_url, transport, headers = _get_config()
    async with _create_transport(server_url, transport, headers) as streams:
        read_stream, write_stream = streams[0], streams[1]
        async with ClientSession(read_stream, write_stream) as session:
            await session.initialize()
            result = await session.call_tool(tool_name, arguments)
            content = []
            for item in result.content:
                entry = {"type": item.type}
                if hasattr(item, "text") and item.text is not None:
                    entry["text"] = item.text
                content.append(entry)
            return {"content": content, "isError": result.isError}


async def _list_tools() -> list:
    """Async: List available tools on the MCP server."""
    server_url, transport, headers = _get_config()
    async with _create_transport(server_url, transport, headers) as streams:
        read_stream, write_stream = streams[0], streams[1]
        async with ClientSession(read_stream, write_stream) as session:
            await session.initialize()
            result = await session.list_tools()
            tools = []
            for tool in result.tools:
                tools.append({
                    "name": tool.name,
                    "description": tool.description or "",
                })
            return tools


def _create_transport(server_url: str, transport: str, headers: dict | None = None):
    """Create the appropriate MCP transport context manager."""
    extra_headers = headers or {}
    if transport == "sse":
        return sse_client(server_url, headers=extra_headers)
    return streamablehttp_client(server_url, headers=extra_headers)


_config = {"server_url": None, "transport": "streamable", "headers": None}


def _get_config():
    server_url = _config["server_url"]
    if not server_url:
        raise RuntimeError("MCP client not configured. Call setup_client() first.")
    return server_url, _config["transport"], _config["headers"]


def setup_client(server_url: str, transport: str = "streamable", headers: dict | None = None):
    _config["server_url"] = server_url
    _config["transport"] = transport
    _config["headers"] = headers


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description="MCP Client - Call tools on an MCP server")
    parser.add_argument("--server-url", required=True, help="MCP server URL")
    parser.add_argument("--transport", default="streamable", choices=["sse", "streamable"], help="Transport type")
    parser.add_argument("--headers", default="{}", help="JSON headers for authentication, e.g. '{\"Authorization\": \"Bearer TOKEN\"}'")
    parser.add_argument("--list", action="store_true", help="List available tools")
    parser.add_argument("--tool", help="Tool name to call")
    parser.add_argument("--arguments", default="{}", help="JSON arguments for the tool")

    args = parser.parse_args()
    try:
        req_headers = json.loads(args.headers)
    except json.JSONDecodeError as e:
        print(f"Error: Invalid JSON headers: {e}", file=sys.stderr)
        sys.exit(1)
    setup_client(args.server_url, args.transport, headers=req_headers)

    if args.list:
        tools = list_tools()
        print(json.dumps(tools, indent=2, ensure_ascii=False))
    elif args.tool:
        try:
            tool_args = json.loads(args.arguments)
        except json.JSONDecodeError as e:
            print(f"Error: Invalid JSON arguments: {e}", file=sys.stderr)
            sys.exit(1)
        result = call_tool(args.tool, tool_args)
        print(json.dumps(result, indent=2, ensure_ascii=False))
    else:
        parser.print_help()
`)

	return b.String()
}

func toPythonFuncName(name string) string {
	result := strings.ReplaceAll(name, "-", "_")
	result = strings.ReplaceAll(result, ".", "_")
	// Ensure it starts with a letter
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "tool_" + result
	}
	return result
}

func toPythonParamName(name string) string {
	// Replace hyphens and dots with underscores for valid Python identifiers.
	result := strings.ReplaceAll(name, "-", "_")
	result = strings.ReplaceAll(result, ".", "_")
	// Prefix with "arg_" so all tool parameters are clearly identifiable
	// and can never conflict with Python keywords or builtins.
	return "arg_" + result
}

func buildArgsDictWithMapping(params []ParameterDocument, pythonToOriginal map[string]string) string {
	required := make([]ParameterDocument, 0, len(params))
	for _, p := range params {
		if p.Required {
			required = append(required, p)
		}
	}
	if len(required) == 0 {
		return ""
	}
	parts := make([]string, 0, len(required))
	for _, p := range required {
		pyName := toPythonParamName(p.Name)
		originalName := p.Name
		if mapped, ok := pythonToOriginal[pyName]; ok {
			originalName = mapped
		}
		parts = append(parts, fmt.Sprintf("\"%s\": %s", originalName, pyName))
	}
	return strings.Join(parts, ", ")
}

func exampleValue(typeStr string) string {
	switch typeStr {
	case "string":
		return `"example"`
	case "integer":
		return "0"
	case "number":
		return "0.0"
	case "boolean":
		return "false"
	case "array":
		return "[]"
	case "object":
		return "{}"
	default:
		return `"example"`
	}
}

// yamlQuote wraps a string in YAML double quotes if it contains characters
// that would be misinterpreted by the YAML parser (colons, brackets, etc.).
func yamlQuote(s string) string {
	needsQuoting := false
	for _, c := range s {
		switch c {
		case ':', '{', '}', '[', ']', ',', '&', '*', '#', '?', '|', '-', '<', '>', '=', '!', '%', '@', '`', '"', '\'':
			needsQuoting = true
		}
		if needsQuoting {
			break
		}
	}
	if !needsQuoting && s != "" && s[0] != ' ' && s[len(s)-1] != ' ' {
		return s
	}
	// Escape backslashes and double quotes, then wrap in double quotes.
	escaped := strings.ReplaceAll(s, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
