# Ordinary agent tool error recovery

Ordinary agents created by `internal/agent-service/service/single_agent.go` use an Eino ChatModelAgent loop. A tool execution error previously escaped the ToolsNode and terminated the run, preventing the model from inspecting the error and correcting its arguments.

`BuildAgentToolsConfig` now installs a tool middleware that returns execution failures as tool results using the existing `{"errorMsg":"..."}` format. Eino retains the original tool call ID and includes the result in the next model input. The same result also reaches the agent event stream, so a failed tool attempt can be displayed without ending the conversation.

The middleware handles invokable tools, stream initialization errors, and errors encountered while reading a tool stream. Successful chunks remain available. Skill errors use the existing serialized message and completion format, identified by the registered skill ID even when its name is hashed.

An overall run cancellation, an expired run context, or an Eino resumable interrupt remains an error. A timeout confined to one tool can be returned to a still-active model. Stream programming panics remain fatal errors. The existing ChatModelAgent iteration limit still bounds the loop; error recovery does not restart the runner or grant unlimited retries.

MCP business errors (`IsError`) are returned immediately instead of repeating the same call. Transport failures retain the configured retry count, with waits that respond to context cancellation. Exhausted transport errors are handled by the common recovery middleware. Skill and ChatDoc argument decoding failures are now explicit errors instead of being ignored.

This change applies to ordinary agents and their ordinary multi-agent subagents. It does not change the general-agent implementation, model API error handling, tool discovery, or unknown-tool dispatch. Those failures remain separate from execution errors. Model correction quality depends on the error information and the model.

## Validation

```bash
go test -mod=readonly ./internal/agent-service/service/agent-tool/... ./internal/agent-service/service/mcp-client
```

The regression tests use the actual Eino ADK and MCP adapter with local model and server stubs. They verify the failed call result reaches the next model input with its original call ID, corrected arguments are used by a subsequent call, and the run completes in both streaming and non-streaming modes. Additional cases cover partial stream failures, Skill completion messages, parameter decoding, cancellation, interrupts, and MCP retry behavior. They do not measure real-model correction rates or exercise a deployed UI.
