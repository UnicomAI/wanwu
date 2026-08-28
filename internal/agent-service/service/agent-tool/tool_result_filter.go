package agent_tool

import (
	"context"
	"encoding/json"
	"strings"

	agent_util "github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// NewToolResultFilterMiddleware creates an AgentMiddleware that filters tool result content
// before it is sent to the chat model. The SSE event stream remains unaffected — only the
// model's input (State.Messages) is modified.
//
// How it works:
//   - Before each ChatModel call, the BeforeChatModel hook inspects State.Messages for
//     Tool-role messages whose ToolName starts with the skill prefix.
//   - For each such message, the Content field (a concatenation of JSON-encoded schema.Message
//     chunks) is parsed, filtered using the provided ToolResultFilter, and reconstructed.
//   - Chunks that the filter rejects are removed; only the kept chunks are re-serialized and
//     concatenated back into the ToolMessage.Content.
func NewToolResultFilterMiddleware() adk.AgentMiddleware {
	return adk.AgentMiddleware{
		BeforeChatModel: filterToolResultBeforeModel,
	}
}

// filterToolResultBeforeModel is the BeforeChatModel hook that filters tool result messages
// in the conversation state before they are sent to the model.
// It also:
//   - Merges consecutive system messages at the head into a single system message.
//   - Removes spurious blank system messages (e.g. {"content":"\n","role":"system"}).
func filterToolResultBeforeModel(ctx context.Context, state *adk.ChatModelAgentState) error {
	state.Messages = mergeHeadSystemMessages(state.Messages)

	// Filter skill tool result content
	for i, msg := range state.Messages {
		if msg.Role != schema.Tool {
			continue
		}
		if !strings.HasPrefix(msg.ToolName, agent_util.AgentSkillPrefix) {
			continue
		}
		if msg.Content == "" {
			continue
		}

		filtered, changed := filterToolMessageContent(msg.ToolName, msg.Content)
		if changed {
			state.Messages[i].Content = filtered
		}
	}
	return nil
}

// mergeHeadSystemMessages merges consecutive system messages at the beginning of the message
// list into a single system message, separated by newlines. Blank system messages (whitespace
// only) are dropped during the merge.
func mergeHeadSystemMessages(msgs []*schema.Message) []*schema.Message {
	if len(msgs) == 0 {
		return msgs
	}

	// Find the end of the leading system message run
	end := 0
	for end < len(msgs) && msgs[end].Role == schema.System {
		end++
	}

	// No system messages, or only one — nothing to merge
	if end <= 1 {
		return msgs
	}

	// Collect non-blank contents from the leading system messages
	var parts []string
	for i := 0; i < end; i++ {
		trimmed := strings.TrimSpace(msgs[i].Content)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}

	// All blank — remove all system messages
	if len(parts) == 0 {
		return msgs[end:]
	}

	// Merge into a single system message
	merged := &schema.Message{
		Role:    schema.System,
		Content: strings.Join(parts, "\n"),
	}

	result := make([]*schema.Message, 0, len(msgs)-end+1)
	result = append(result, merged)
	result = append(result, msgs[end:]...)
	return result
}

// filterToolMessageContent parses the concatenated JSON chunks in a ToolMessage's Content,
// applies the active ToolResultFilter, and returns the re-concatenated result.
// Returns the filtered content and whether any chunk was removed.
func filterToolMessageContent(toolName, content string) (filterContent string, hasChange bool) {
	defer util.PrintPanicStackWithCall(func(panicOccur bool, recoverError error) {
		if panicOccur {
			filterContent = content
			hasChange = false
		}
	})
	chunks := parseConcatenatedJSON(content)
	if len(chunks) == 0 {
		return content, false
	}

	var kept []string
	changed := false
	var startKept = false
	for _, chunk := range chunks {
		var msg schema.Message
		if err := json.Unmarshal([]byte(chunk), &msg); err != nil {
			// If we can't parse a chunk, keep it as-is
			log.Warnf("tool_result_filter: failed to parse chunk in tool %s: %v", toolName, err)
			kept = append(kept, chunk)
			continue
		}
		// If we encounter a WGA stop message, start keeping chunks
		if agent_util.WgaStopMessage(&msg) && len(msg.Content) > 0 {
			startKept = true
		}
		if startKept {
			kept = append(kept, msg.Content)
			changed = true
		}
	}

	if !changed {
		return content, false
	}
	return strings.Join(kept, ""), true
}

// parseConcatenatedJSON splits a string containing multiple adjacent JSON objects into
// individual JSON strings. For example: `{"a":1}{"b":2}` -> [`{"a":1}`, `{"b":2}`].
func parseConcatenatedJSON(s string) []string {
	var results []string
	decoder := json.NewDecoder(strings.NewReader(s))
	for decoder.More() {
		offset := decoder.InputOffset()
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			// If parsing fails, treat the rest as a single chunk
			remaining := s[offset:]
			if strings.TrimSpace(remaining) != "" {
				results = append(results, remaining)
			}
			break
		}
		end := decoder.InputOffset()
		results = append(results, s[offset:end])
	}
	return results
}
