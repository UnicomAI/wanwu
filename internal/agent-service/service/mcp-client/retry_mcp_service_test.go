package mcp_client

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

type stubClient struct {
	client.MCPClient
	call func() (*mcp.CallToolResult, error)
}

func (*stubClient) Start(context.Context) error { return nil }

func (s *stubClient) CallTool(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return s.call()
}

func TestBusinessErrorIsNotRetried(t *testing.T) {
	calls := 0
	want := &mcp.CallToolResult{IsError: true}
	r := &RetryMcpClient{retry: &Retry{Times: 2}, client: &stubClient{call: func() (*mcp.CallToolResult, error) {
		calls++
		return want, nil
	}}}
	got, err := r.CallTool(context.Background(), mcp.CallToolRequest{})
	if err != nil || got != want || calls != 1 {
		t.Fatalf("result=%v error=%v calls=%d", got, err, calls)
	}
}

func TestTransportRetries(t *testing.T) {
	for _, succeed := range []bool{false, true} {
		t.Run(map[bool]string{false: "exhausted", true: "eventually succeeds"}[succeed], func(t *testing.T) {
			calls := 0
			transportErr := errors.New("connection reset")
			r := &RetryMcpClient{retry: &Retry{Times: 2}, client: &stubClient{call: func() (*mcp.CallToolResult, error) {
				calls++
				if succeed && calls == 2 {
					return &mcp.CallToolResult{}, nil
				}
				return nil, transportErr
			}}}
			got, err := r.CallTool(context.Background(), mcp.CallToolRequest{})
			if succeed {
				if err != nil || got == nil || calls != 2 {
					t.Fatalf("result=%v error=%v calls=%d", got, err, calls)
				}
			} else if !errors.Is(err, transportErr) || calls != 3 {
				t.Fatalf("error=%v calls=%d", err, calls)
			}
		})
	}
}

func TestCancelDuringRetryWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	called := make(chan struct{}, 1)
	r := &RetryMcpClient{retry: &Retry{Times: 2, Span: time.Hour}, client: &stubClient{call: func() (*mcp.CallToolResult, error) {
		called <- struct{}{}
		return nil, errors.New("connection reset")
	}}}
	done := make(chan error, 1)
	go func() {
		_, err := r.CallTool(ctx, mcp.CallToolRequest{})
		done <- err
	}()
	<-called
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("lost cancellation: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("retry wait ignored cancellation")
	}
}

func TestEmptyResultDoesNotPanic(t *testing.T) {
	r := &RetryMcpClient{log: true, client: &stubClient{call: func() (*mcp.CallToolResult, error) {
		return nil, nil
	}}}
	if _, err := r.CallTool(context.Background(), mcp.CallToolRequest{}); err == nil {
		t.Fatal("expected an error for an empty result")
	}
}
