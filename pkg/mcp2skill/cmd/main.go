package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/UnicomAI/wanwu/pkg/mcp2skill"
)

var (
	buildTime    string
	buildVersion string
	gitCommitID  string
	gitBranch    string
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "-v", "--version":
		fmt.Printf("mcp2skill %s\n", buildVersion)
		fmt.Printf("  build_time:    %s\n", buildTime)
		fmt.Printf("  git_commit_id: %s\n", gitCommitID)
		fmt.Printf("  git_branch:    %s\n", gitBranch)
		return
	case "-h", "--help":
		printUsage()
		return
	}

	cfg := &mcp2skill.MCPConfig{}
	outputDir := "."
	timeout := 30 * time.Second

	for _, arg := range args {
		switch arg {
		case "-v", "--version":
			fmt.Printf("mcp2skill %s\n", buildVersion)
			return
		case "-h", "--help":
			printUsage()
			return
		default:
			if err := parseKeyValue(cfg, &outputDir, &timeout, arg); err != nil {
				fmt.Fprintf(os.Stderr, "invalid argument: %s (%v)\n", arg, err)
				os.Exit(1)
			}
		}
	}

	if cfg.StreamableUrl == "" && cfg.SseUrl == "" {
		fmt.Fprintln(os.Stderr, "error: must provide streamableUrl or sseUrl")
		printUsage()
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := mcp2skill.ConvertFromMCPConfig(ctx, cfg, outputDir); err != nil {
		fmt.Fprintf(os.Stderr, "conversion failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("done")
}

func parseKeyValue(cfg *mcp2skill.MCPConfig, outputDir *string, timeout *time.Duration, arg string) error {
	idx := -1
	for i, c := range arg {
		if c == '=' {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("expected key=value format")
	}
	key := arg[:idx]
	val := arg[idx+1:]

	switch key {
	case "name":
		cfg.Name = val
	case "description":
		cfg.Description = val
	case "streamableUrl":
		cfg.StreamableUrl = val
	case "sseUrl":
		cfg.SseUrl = val
	case "transport":
		cfg.Transport = val
	case "output":
		*outputDir = val
	case "timeout":
		d, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("invalid timeout: %s", val)
		}
		*timeout = d
	case "headers":
		var headers map[string]string
		if err := json.Unmarshal([]byte(val), &headers); err != nil {
			return fmt.Errorf("invalid headers JSON: %v", err)
		}
		if cfg.Headers == nil {
			cfg.Headers = make(map[string]string)
		}
		for k, v := range headers {
			cfg.Headers[k] = v
		}
	case "apiAuth":
		var auth mcp2skill.APIAuthConfig
		if err := json.Unmarshal([]byte(val), &auth); err != nil {
			return fmt.Errorf("invalid apiAuth JSON: %v", err)
		}
		cfg.ApiAuth = &auth
	default:
		return fmt.Errorf("unknown key: %s", key)
	}
	return nil
}

func printUsage() {
	fmt.Println("mcp2skill - convert MCP server tools to skill format")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  mcp2skill streamableUrl=http://... name=MySkill [key=value...]")
	fmt.Println()
	fmt.Println("Arguments (key=value):")
	fmt.Println("  streamableUrl  MCP server streamable HTTP URL")
	fmt.Println("  sseUrl         MCP server SSE URL")
	fmt.Println("  name           Skill name (auto-detected if empty)")
	fmt.Println("  description    Skill description (auto-generated if empty)")
	fmt.Println("  transport      Transport type: streamable (default) or sse")
	fmt.Println("  output         Output directory (default: .)")
	fmt.Println("  timeout        Connection timeout (default: 30s)")
	fmt.Println("  headers        Custom HTTP headers as JSON object, e.g. '{\"Authorization\":\"Bearer TOKEN\"}'")
	fmt.Println("  apiAuth        API auth config as JSON object, e.g. '{\"authType\":\"api_key_header\",\"apiKeyHeaderPrefix\":\"bearer\",\"apiKeyValue\":\"TOKEN\"}'")
	fmt.Println("  -v, --version  Print version")
	fmt.Println("  -h, --help     Print help")
}
