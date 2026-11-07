package tool_recovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	einomcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func errorResult(err error) string {
	b, _ := json.Marshal(map[string]string{"errorMsg": err.Error()})
	return string(b)
}

func recoveryMiddleware() compose.ToolMiddleware {
	return NewMiddleware(func(_ string, err error) (string, error) {
		return errorResult(err), nil
	})
}

// Embed the unused MCP methods; the probe needs only discovery and invocation.
type mcpStub struct {
	client.MCPClient
	arguments []string
}

func (s *mcpStub) ListTools(context.Context, mcp.ListToolsRequest) (*mcp.ListToolsResult, error) {
	return &mcp.ListToolsResult{Tools: []mcp.Tool{mcp.NewTool("lookup")}}, nil
}

func (s *mcpStub) CallTool(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args, _ := json.Marshal(req.Params.Arguments)
	s.arguments = append(s.arguments, string(args))
	if !strings.Contains(string(args), "correct") {
		return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: "invalid city; use correct"},
		}}, nil
	}
	return &mcp.CallToolResult{Content: []mcp.Content{
		mcp.TextContent{Type: "text", Text: "lookup succeeded"},
	}}, nil
}

type streamTool struct{ calls int }

func (*streamTool) Info(context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{Name: "lookup", Desc: "probe"}, nil
}

func (s *streamTool) StreamableRun(context.Context, string, ...tool.Option) (*schema.StreamReader[string], error) {
	s.calls++
	if s.calls > 1 {
		return schema.StreamReaderFromArray([]string{"lookup succeeded"}), nil
	}
	sr, sw := schema.Pipe[string](2)
	sw.Send("partial output; ", nil)
	sw.Send("", errors.New("invalid city; use correct"))
	sw.Close()
	return sr, nil
}

// Scripted model: verifies the actual next-turn messages before issuing a
// corrected call. It proves message plumbing, not real model reasoning quality.
type scriptedModel struct {
	calls int
	err   error
}

func (m *scriptedModel) WithTools([]*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return m, nil
}

func (m *scriptedModel) Generate(_ context.Context, msgs []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	m.calls++
	if m.calls > 1 {
		last := msgs[len(msgs)-1]
		want, id := "invalid city", "call-1"
		if m.calls == 3 {
			want, id = "lookup succeeded", "call-2"
		}
		if last.Role != schema.Tool || last.ToolCallID != id || !strings.Contains(last.Content, want) {
			m.err = fmt.Errorf("round %d: expected tool result %s containing %q, got %+v", m.calls, id, want, last)
			return nil, m.err
		}
	}
	if m.calls == 3 {
		return schema.AssistantMessage("completed", nil), nil
	}
	args := `{"city":"wrong"}`
	if m.calls == 2 {
		args = `{"city":"correct"}`
	}
	idx := 0
	return schema.AssistantMessage("", []schema.ToolCall{{ID: fmt.Sprintf("call-%d", m.calls), Index: &idx,
		Type: "function", Function: schema.FunctionCall{Name: "lookup", Arguments: args}}}), nil
}

func (m *scriptedModel) Stream(ctx context.Context, msgs []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	msg, err := m.Generate(ctx, msgs, opts...)
	if err != nil {
		return nil, err
	}
	return schema.StreamReaderFromArray([]*schema.Message{msg}), nil
}

func TestAgentLoop(t *testing.T) {
	for _, stream := range []bool{false, true} {
		for _, kind := range []string{"mcp", "late-stream-error"} {
			for _, recoverTools := range []bool{false, true} {
				t.Run(fmt.Sprintf("stream=%t/tool=%s/recovery=%t", stream, kind, recoverTools), func(t *testing.T) {
					ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
					defer cancel()
					var tools []tool.BaseTool
					stub := &mcpStub{}
					if kind == "mcp" {
						var err error
						tools, err = einomcp.GetTools(ctx, &einomcp.Config{Cli: stub})
						if err != nil {
							t.Fatal(err)
						}
					} else {
						tools = []tool.BaseTool{&streamTool{}}
					}
					cfg := compose.ToolsNodeConfig{Tools: tools}
					if recoverTools {
						cfg.ToolCallMiddlewares = []compose.ToolMiddleware{recoveryMiddleware()}
					}
					cm := &scriptedModel{}
					agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
						Name: "probe", Description: "probe", Model: cm,
						ToolsConfig: adk.ToolsConfig{ToolsNodeConfig: cfg}, MaxIterations: 4,
					})
					if err != nil {
						t.Fatal(err)
					}
					it := adk.NewRunner(ctx, adk.RunnerConfig{Agent: agent, EnableStreaming: stream}).Query(ctx, "lookup")
					var failures []error
					var toolResults []string
					completed := false
					for {
						ev, ok := it.Next()
						if !ok {
							break
						}
						if ev.Err != nil {
							failures = append(failures, ev.Err)
							continue
						}
						if ev.Output == nil || ev.Output.MessageOutput == nil {
							continue
						}
						out := ev.Output.MessageOutput
						msg := out.Message
						if out.IsStreaming {
							msg, err = schema.ConcatMessageStream(out.MessageStream)
							if err != nil {
								failures = append(failures, err)
								continue
							}
						}
						if msg != nil && msg.Role == schema.Tool {
							toolResults = append(toolResults, msg.Content)
						}
						if msg != nil && msg.Role == schema.Assistant && msg.Content == "completed" {
							completed = true
						}
					}
					if cm.err != nil {
						t.Fatal(cm.err)
					}
					if recoverTools {
						if len(failures) != 0 || !completed || cm.calls != 3 || len(toolResults) != 2 {
							t.Fatalf("recovery failed: errors=%v completed=%t modelCalls=%d results=%v", failures, completed, cm.calls, toolResults)
						}
						if kind == "mcp" && (len(stub.arguments) != 2 || !strings.Contains(stub.arguments[1], "correct")) {
							t.Fatalf("arguments were not corrected: %v", stub.arguments)
						}
					} else if len(failures) == 0 || completed || cm.calls != 1 {
						t.Fatalf("expected baseline failure: errors=%v completed=%t modelCalls=%d", failures, completed, cm.calls)
					}
					t.Logf("modelCalls=%d toolResults=%d completed=%t errors=%v", cm.calls, len(toolResults), completed, failures)
				})
			}
		}
	}
}

func TestRecoveryPreservesCanceledRun(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	next := recoveryMiddleware().Invokable(func(context.Context, *compose.ToolInput) (*compose.ToolOutput, error) {
		return nil, context.Canceled
	})
	_, err := next(ctx, &compose.ToolInput{Name: "lookup"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("lost cancellation: %v", err)
	}
}

func TestRecoveryPreservesInterrupt(t *testing.T) {
	interrupt := compose.NewInterruptAndRerunErr("human input required")
	next := recoveryMiddleware().Invokable(func(context.Context, *compose.ToolInput) (*compose.ToolOutput, error) {
		return nil, interrupt
	})
	_, err := next(context.Background(), &compose.ToolInput{Name: "lookup"})
	if err != interrupt {
		t.Fatalf("lost interrupt: %v", err)
	}
}

func TestStreamErrorBoundaries(t *testing.T) {
	for _, immediate := range []bool{false, true} {
		for _, interrupt := range []bool{false, true} {
			t.Run(fmt.Sprintf("immediate=%t/interrupt=%t", immediate, interrupt), func(t *testing.T) {
				failure := error(context.DeadlineExceeded) // The enclosing agent is still active.
				if interrupt {
					failure = compose.NewInterruptAndRerunErr("human input required")
				}
				run := recoveryMiddleware().Streamable(func(context.Context, *compose.ToolInput) (*compose.StreamToolOutput, error) {
					if immediate {
						return nil, failure
					}
					sr, sw := schema.Pipe[string](1)
					sw.Send("", failure)
					sw.Close()
					return &compose.StreamToolOutput{Result: sr}, nil
				})
				out, err := run(context.Background(), &compose.ToolInput{Name: "lookup"})
				if immediate && interrupt {
					if err != failure {
						t.Fatalf("lost interrupt: %v", err)
					}
					return
				}
				if err != nil {
					t.Fatal(err)
				}
				defer out.Result.Close()
				chunk, err := out.Result.Recv()
				if interrupt {
					if err != failure {
						t.Fatalf("lost interrupt: %v", err)
					}
				} else if err != nil || !strings.Contains(chunk, failure.Error()) {
					t.Fatalf("tool-local timeout was not returned: chunk=%q error=%v", chunk, err)
				}
				if _, err := out.Result.Recv(); err != io.EOF {
					t.Fatalf("stream did not finish: %v", err)
				}
			})
		}
	}
}

func TestStreamPanicStaysFatal(t *testing.T) {
	run := recoveryMiddleware().Streamable(func(context.Context, *compose.ToolInput) (*compose.StreamToolOutput, error) {
		sr := schema.StreamReaderWithConvert(schema.StreamReaderFromArray([]string{"chunk"}), func(string) (string, error) {
			panic("broken converter")
		})
		return &compose.StreamToolOutput{Result: sr}, nil
	})
	out, err := run(context.Background(), &compose.ToolInput{Name: "lookup"})
	if err != nil {
		t.Fatal(err)
	}
	defer out.Result.Close()
	if _, err := out.Result.Recv(); err == nil || !strings.Contains(err.Error(), "broken converter") {
		t.Fatalf("lost stream panic: %v", err)
	}
}
