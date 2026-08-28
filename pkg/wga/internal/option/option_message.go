// Package option 提供智能体运行选项的内部实现。
package option

import (
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// SystemMessageStrategy 定义 system 消息处理策略。
type SystemMessageStrategy string

const (
	SystemMessageStrategyNone  SystemMessageStrategy = ""      // 不处理（默认）
	SystemMessageStrategyMerge SystemMessageStrategy = "merge" // 合并所有 system 消息到第一位
)

// ExtractSystemMessage 提取所有 role=system 消息的内容并返回剩余消息。
//
// 遍历消息列表，收集所有 role=system 且 Content 非空的消息，将内容用 "\n\n" 拼接；
// 非 system 消息保持原顺序返回。
// 主要用于解决某些模型不接受 system 消息不在第一位的问题。
func ExtractSystemMessage(messages []adk.Message) (systemContent string, otherMessages []adk.Message) {
	var systemParts []string
	otherMessages = make([]adk.Message, 0, len(messages))

	for _, msg := range messages {
		if msg.Role == schema.System && msg.Content != "" {
			systemParts = append(systemParts, msg.Content)
		} else {
			otherMessages = append(otherMessages, msg)
		}
	}

	return strings.Join(systemParts, "\n\n"), otherMessages
}

// WithSystemMessageStrategy 设置 system 消息处理策略。
//
// 某些模型（如 OpenAI 部分版本）要求 role=system 的消息必须位于消息列表第一位，
// 否则会报错或忽略后续 system 消息。启用 SystemMessageStrategyMerge 后，
// 系统会在运行 agent 前提取所有 role=system 的消息，合并后放到指令（instruction）的最前面，
// 确保符合模型要求。
func WithSystemMessageStrategy(strategy SystemMessageStrategy) Option {
	return optionFunc(func(opts *Options) error {
		opts.SystemMessageStrategy = strategy
		return nil
	})
}
