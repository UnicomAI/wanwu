// Integration driver: calls Wanwu's real sandbox runner without starting its DB/UI stack.
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	sandbox "github.com/UnicomAI/wanwu/pkg/wga-sandbox"
	option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	runID := fmt.Sprintf("finsafe-integration-%d", time.Now().UnixNano())
	inputDir, err := os.MkdirTemp("", "wanwu-input-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(inputDir)
	if err := os.WriteFile(filepath.Join(inputDir, "input.txt"), []byte("WANWU_INTEGRATION_OK"), 0600); err != nil {
		return err
	}
	outputDir := filepath.Join(os.TempDir(), runID+"-output")
	defer os.RemoveAll(outputDir)
	// The application passes an existing persistent workspace as OutputDir.
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}
	forbidden := "/var/lib/" + runID + "-forbidden"
	command := "cat input.txt > integration.txt; " +
		"if printf forbidden > " + forbidden + "; then printf UNEXPECTED > denial.txt; " +
		"else printf DENIED > denial.txt; fi; cat integration.txt denial.txt"
	prompt := "Execute the requested test command. WANWU_TEST_COMMAND:" + base64.StdEncoding.EncodeToString([]byte(command))
	model := option.ModelConfig{
		Provider: "fixture", ProviderName: "Local deterministic fixture",
		BaseURL: "http://127.0.0.1:8000/v1", APIKey: "fixture-only", Model: "fixture", ModelName: "Fixture",
	}
	modelDescription := "deterministic local OpenAI-compatible fixture"
	if os.Getenv("WANWU_LIVE_MODEL") == "1" {
		model = option.ModelConfig{
			Provider: "deepseek", ProviderName: "DeepSeek live test",
			BaseURL: os.Getenv("WANWU_LIVE_MODEL_BASE_URL"), APIKey: "local-test-gateway",
			Model: "deepseek-flash", ModelName: "DeepSeek Flash",
		}
		prompt = "This is an authorized sandbox boundary test using synthetic files. Run exactly this script with the bash tool, " +
			"then read integration.txt and denial.txt to verify their contents. Do not change the script. Finally reply DONE.\n```sh\n" + command + "\n```"
		modelDescription = "official DeepSeek Flash via credential-isolating forwarding gateway"
	}
	_, events, err := sandbox.Run(ctx,
		option.WithSandbox(option.SandboxReuse("127.0.0.1")),
		option.WithRunnerType(option.RunnerTypeOpencode),
		option.WithModelConfig(model),
		option.WithRunSession(option.RunSession{ThreadID: runID, RunID: runID}),
		option.WithMessages([]adk.Message{&schema.Message{Role: schema.User, Content: prompt}}),
		option.WithInputDir(inputDir+"/."), option.WithOutputDir(outputDir), option.WithSkipCleanup(true),
	)
	if err != nil {
		return err
	}
	defer sandbox.Cleanup(context.Background(), runID)
	counts := map[string]int{}
	boundaryCommandObserved := false
	for line := range events {
		var event struct {
			Type string `json:"type"`
			Part struct {
				Tool  string `json:"tool"`
				State struct {
					Input json.RawMessage `json:"input"`
				} `json:"state"`
			} `json:"part"`
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			return fmt.Errorf("decode Wanwu event: %w", err)
		}
		counts[event.Type]++
		if event.Type == "tool_use" && event.Part.Tool == "bash" && strings.Contains(string(event.Part.State.Input), forbidden) {
			boundaryCommandObserved = true
		}
		if event.Type == "error" {
			return fmt.Errorf("Wanwu runner error: %s", line)
		}
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if counts["tool_use"] == 0 || counts["text"] == 0 {
		return fmt.Errorf("missing tool/text events: %v", counts)
	}
	if !boundaryCommandObserved {
		return fmt.Errorf("the requested boundary-check Bash command was not observed")
	}
	expected := map[string]string{"integration.txt": "WANWU_INTEGRATION_OK", "denial.txt": "DENIED"}
	err = filepath.WalkDir(outputDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		want, ok := expected[entry.Name()]
		if !ok {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(content)) != want {
			return fmt.Errorf("unexpected artifact %s: %q", entry.Name(), content)
		}
		delete(expected, entry.Name())
		return nil
	})
	if err != nil {
		return err
	}
	if len(expected) != 0 {
		return fmt.Errorf("missing downloaded artifacts: %v", expected)
	}
	if container := os.Getenv("WANWU_LIVE_DOCKER_CONTAINER"); container != "" {
		if err := exec.CommandContext(ctx, "docker", "exec", container, "test", "!", "-e", forbidden).Run(); err != nil {
			return fmt.Errorf("container out-of-scope artifact check failed: %w", err)
		}
	} else if _, err := os.Stat(forbidden); !os.IsNotExist(err) {
		return fmt.Errorf("out-of-scope artifact was created or could not be checked: %v", err)
	}
	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"result": "PASS", "driver": "actual Wanwu pkg/wga-sandbox.Run",
		"model": modelDescription, "events": counts,
		"input_upload": true, "artifact_download": true, "out_of_scope_write_denied": true,
		"boundary_command_observed": boundaryCommandObserved,
	})
}
