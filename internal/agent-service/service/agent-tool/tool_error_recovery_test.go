package agent_tool

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func TestSkillErrorKeepsWireFormat(t *testing.T) {
	for _, name := range []string{"skill-example", "d41d8cd98f00b204e9800998ecf8427e"} {
		t.Run(name, func(t *testing.T) {
			middleware := newToolErrorRecoveryMiddleware([]tool.BaseTool{&skillTool{info: &schema.ToolInfo{Name: name}}})
			run := middleware.Streamable(func(context.Context, *compose.ToolInput) (*compose.StreamToolOutput, error) {
				return nil, errors.New("invalid uploadFileUrl")
			})
			out, err := run(context.Background(), &compose.ToolInput{Name: name})
			if err != nil {
				t.Fatal(err)
			}
			defer out.Result.Close()
			chunk, err := out.Result.Recv()
			if err != nil {
				t.Fatal(err)
			}
			var msg schema.Message
			if err := json.Unmarshal([]byte(chunk), &msg); err != nil {
				t.Fatal(err)
			}
			if msg.Role != schema.Assistant || msg.ResponseMeta == nil || msg.ResponseMeta.FinishReason != "stop" {
				t.Fatalf("invalid skill completion: %+v", msg)
			}
			var failure response.AgentToolErrResp
			if err := json.Unmarshal([]byte(msg.Content), &failure); err != nil {
				t.Fatal(err)
			}
			if failure.ErrorMsg != "invalid uploadFileUrl" {
				t.Fatalf("lost error: %+v", failure)
			}
			filtered, _ := filterToolMessageContent(name, chunk)
			if !strings.Contains(filtered, "invalid uploadFileUrl") {
				t.Fatal("skill filter removed the error")
			}
		})
	}
}

func TestPluginParameterErrorBecomesResult(t *testing.T) {
	plugin := &openAPITool{handler: createHTTPHandler("http://unused.invalid", "/lookup", "get", nil, "application/json", "lookup")}
	run := newToolErrorRecoveryMiddleware(nil).Invokable(func(ctx context.Context, in *compose.ToolInput) (*compose.ToolOutput, error) {
		result, err := plugin.InvokableRun(ctx, in.Arguments)
		return &compose.ToolOutput{Result: result}, err
	})
	out, err := run(context.Background(), &compose.ToolInput{Name: "lookup", Arguments: `{"city":`})
	if err != nil {
		t.Fatal(err)
	}
	var failure response.AgentToolErrResp
	if err := json.Unmarshal([]byte(out.Result), &failure); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(failure.ErrorMsg, "failed to parse arguments") {
		t.Fatalf("lost parameter error: %+v", failure)
	}
}

func TestInvalidSkillArgumentsStopBeforeEnvironmentSetup(t *testing.T) {
	executor := (&SkillExecutor{}).initEnv(`{"uploadFileUrl":42}`)
	if executor.err == nil || executor.runEnv != nil {
		t.Fatalf("invalid arguments were accepted: %+v", executor)
	}
}

func TestInvalidChatDocArguments(t *testing.T) {
	if _, err := buildChatDocParams(`{"upload_file_url":42}`, nil); err == nil {
		t.Fatal("invalid arguments were accepted")
	}
}
