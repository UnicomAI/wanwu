package wga_sandbox_converter

import (
	"context"

	"github.com/UnicomAI/wanwu/pkg/util"
	wga_sandbox_option "github.com/UnicomAI/wanwu/pkg/wga-sandbox/wga-sandbox-option"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func ConvertToEinoIterator(
	ctx context.Context,
	runnerType wga_sandbox_option.RunnerType,
	outputCh <-chan string,
) *adk.AsyncIterator[*adk.AgentEvent] {
	iterator, generator := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
	conv := NewEinoConverter(runnerType)

	go func() {
		defer util.PrintPanicStack()
		defer generator.Close()

		// Track the current stream writer so we can close it when the message role changes.
		// This follows eino's design: each distinct message (per role/tool) is a separate
		// streaming AgentEvent, allowing flowAgent.genAgentInput() to aggregate each stream
		// into one complete Message via getMessageFromWrappedEvent().
		var (
			curWriter *schema.StreamWriter[*schema.Message]
			curRole   schema.RoleType
			curTool   string
		)

		flush := func() {
			if curWriter != nil {
				curWriter.Close()
				curWriter = nil
			}
		}
		defer flush()

		for {
			select {
			case <-ctx.Done():
				flush()
				return
			case line, ok := <-outputCh:
				if !ok {
					flush()
					return
				}
				msgs, err := conv.Convert(line)
				if err != nil {
					flush()
					generator.Send(&adk.AgentEvent{Err: err})
					continue
				}
				for _, msg := range msgs {
					// Question 消息（Human-in-the-Loop）需要特殊处理：
					// 1. 使用非流式方式（IsStreaming: false），直接通过 Message 字段传递
					// 2. 不走 stream pipe，避免流式管道阻塞导致事件无法及时消费
					// 3. Extra 字段包含 questionType="question" 和 questionData
					if msg.Extra != nil {
						if qt, ok := msg.Extra["questionType"].(string); ok && qt == "question" {
							flush()
							curRole = msg.Role
							curTool = msg.ToolName

							generator.Send(&adk.AgentEvent{
								Output: &adk.AgentOutput{
									MessageOutput: &adk.MessageVariant{
										IsStreaming: false,
										Message:     msg,
										Role:        curRole,
										ToolName:    curTool,
									},
								},
							})
							continue
						}
					}

					// Start a new stream when role changes or a new tool call begins
					if curWriter == nil || msg.Role != curRole || (msg.Role == schema.Tool && msg.ToolName != curTool) {
						flush()
						curRole = msg.Role
						curTool = msg.ToolName

						streamReader, streamWriter := schema.Pipe[*schema.Message](1)
						streamReader.SetAutomaticClose()
						curWriter = streamWriter

						generator.Send(&adk.AgentEvent{
							Output: &adk.AgentOutput{
								MessageOutput: &adk.MessageVariant{
									IsStreaming:   true,
									Message:       msg,
									MessageStream: streamReader,
									Role:          curRole,
									ToolName:      curTool,
								},
							},
						})
					}
					if curWriter.Send(msg, nil) {
						// Stream closed by consumer
						return
					}
				}
			}
		}
	}()

	return iterator
}

func ConvertToEinoIteratorWithError(
	ctx context.Context,
	runnerType wga_sandbox_option.RunnerType,
	err error,
) *adk.AsyncIterator[*adk.AgentEvent] {
	iterator, generator := adk.NewAsyncIteratorPair[*adk.AgentEvent]()

	go func() {
		defer util.PrintPanicStack()
		defer generator.Close()
		generator.Send(&adk.AgentEvent{Err: err})
	}()

	return iterator
}
