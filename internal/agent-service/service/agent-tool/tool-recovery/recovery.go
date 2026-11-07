// Package tool_recovery returns tool execution failures to the model as results.
package tool_recovery

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// ShouldPropagate keeps run cancellation and resumable interrupts as errors.
// A timeout local to one tool can still be returned to an active agent as a result.
func ShouldPropagate(ctx context.Context, err error) bool {
	if ctx.Err() != nil {
		return true
	}
	_, interrupt := compose.IsInterruptRerunError(err)
	return interrupt
}

// NewMiddleware preserves successful results and converts execution errors with
// formatError. The formatter can retain a tool's existing streaming wire format.
func NewMiddleware(formatError func(name string, err error) (string, error)) compose.ToolMiddleware {
	return compose.ToolMiddleware{
		Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
			return func(ctx context.Context, in *compose.ToolInput) (*compose.ToolOutput, error) {
				out, err := next(ctx, in)
				if err == nil || ShouldPropagate(ctx, err) {
					return out, err
				}
				content, err := formatError(in.Name, err)
				if err != nil {
					return nil, err
				}
				return &compose.ToolOutput{Result: content}, nil
			}
		},
		Streamable: func(next compose.StreamableToolEndpoint) compose.StreamableToolEndpoint {
			return func(ctx context.Context, in *compose.ToolInput) (*compose.StreamToolOutput, error) {
				out, err := next(ctx, in)
				if err != nil {
					if ShouldPropagate(ctx, err) {
						return nil, err
					}
					content, err := formatError(in.Name, err)
					if err != nil {
						return nil, err
					}
					return &compose.StreamToolOutput{Result: schema.StreamReaderFromArray([]string{content})}, nil
				}

				// Initialization can succeed while later Recv calls fail. Handle the
				// error before Eino copies the stream to its events and model state.
				sr, sw := schema.Pipe[string](1)
				go func() {
					defer sw.Close()
					// Stream converters run on this goroutine. Keep programming
					// panics on the error path instead of crashing the process.
					defer func() {
						if p := recover(); p != nil {
							sw.Send("", fmt.Errorf("tool %s stream panicked: %v", in.Name, p))
						}
					}()
					defer out.Result.Close()
					for {
						chunk, err := out.Result.Recv()
						if err == io.EOF {
							return
						}
						if err != nil {
							if ShouldPropagate(ctx, err) {
								sw.Send("", err)
							} else {
								content, formatErr := formatError(in.Name, err)
								sw.Send(content, formatErr)
							}
							return
						}
						if sw.Send(chunk, nil) {
							return
						}
					}
				}()
				return &compose.StreamToolOutput{Result: sr}, nil
			}
		},
	}
}
