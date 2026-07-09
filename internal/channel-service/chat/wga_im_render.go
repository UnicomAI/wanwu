package chat

import (
	"encoding/json"
	"fmt"
	"strings"
)

// 本文件把 WGA 聚合器中已完成的 process fragment 渲染成发给 IM（钉钉/微信）的纯文本，
// 仅用于"关键里程碑"下发（Supervisor 委派 transfer / 子智能体 finished / 产物生成）。
// 常规工具调用（glob/read/skill/todowrite/bash 等）与思考过程不下发，避免过程刷屏撞 IM 频控。
// text(正文) 由 chat.go 在 TEXT_MESSAGE_END 时逐段下发，不走这里。
//
// 渲染只用 emoji + 换行，不依赖 markdown 语法（钉钉 sampleText / 微信纯文本都能正常显示）。
// fragment 字段为同包非导出，可直接读取；SSE 处理在单 goroutine 内，无并发问题。

// maxToolResultLineRunes 工具结果简短摘要的长度上限（按 rune 计）。
// 工具结果只有简短（≤该长度）且为纯文本时才显示，否则只留工具名+耗时，永不出现"已截断"。
const maxToolResultLineRunes = 60

// isMilestoneToolCall 判定一个已完成的工具调用是否为"关键里程碑"，需下发到 IM。
// 只下发 Supervisor 委派子智能体这类关键节点（结果含 "transferred to agent"，
// 或工具名含"交给"/"transfer"）；常规工具（glob/read/skill/todowrite/bash/task 等）
// 一律不下发，避免过程刷屏。判定基于净化后的结果文本与工具名。
func isMilestoneToolCall(f *wgaFragment) bool {
	if f == nil || f.kind != fragToolCall {
		return false
	}
	name := strings.ToLower(f.toolCallName)
	if strings.Contains(name, "交给") || strings.Contains(name, "transfer") {
		return true
	}
	result := strings.ToLower(cleanToolResult(f.toolCallResult))
	return strings.Contains(result, "transferred to agent")
}

// renderToolCallLine 🔧 工具名 (耗时)[ → <简短结果>]
// 只有当净化后的结果是简短（≤ maxToolResultLineRunes rune）且无换行的纯文本时才附 "→ 结果"，
// 否则只留头部（路径/skill内容/JSON/长输出一律不显示，且永不出现"已截断"）。
func renderToolCallLine(f *wgaFragment) string {
	name := f.toolCallName
	if name == "" {
		name = "未知工具"
	}
	head := fmt.Sprintf("🔧 %s (%s)", name, durationOrUnknown(f.duration))
	result := cleanToolResult(f.toolCallResult)
	if result == "" || strings.Contains(result, "\n") {
		return head
	}
	if len([]rune(result)) > maxToolResultLineRunes {
		return head // 超长结果不显示，避免"已截断"噪声
	}
	return fmt.Sprintf("%s → %s", head, result)
}

// renderActivityLine 🤖 子智能体: 名称 (耗时)（里程碑直接下发用，子智能体 finished 时调用）
func renderActivityLine(f *wgaFragment) string {
	name := f.agentName
	if name == "" {
		name = "子智能体"
	}
	return fmt.Sprintf("🤖 子智能体: %s (%s)", name, durationOrUnknown(f.duration))
}

// cleanToolResult 把工具结果净化成一行可读摘要。
// WGA 的 TOOL_CALL_RESULT.content 是 SSE content 字段的原始 JSON 文本（常是 JSON 字符串字面量，
// 如 "\"<path>…</path>\""、"[{…}]"），直接显示会带 < 转义、HTML/Skill 标签、换行，在 IM 里很难看。
// 处理顺序：
//  1. 若整体是 JSON 字符串（首字符是 "），先 unquote 出原始文本
//  2. 剥离 <tag>…</tag> 类标签，只留标签内文本（<path>/home/…</path> → /home/…）
//  3. 折叠所有空白（换行/制表符/多空格）为单个空格，压成一行
//  4. TrimSpace；空则返回 ""
func cleanToolResult(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	// 1. 尝试按 JSON 字符串字面量 unquote（处理 "\"...\"" 形态）。
	if len(s) >= 2 && s[0] == '"' {
		if unq, err := jsonUnquote(s); err == nil {
			s = unq
		}
	}
	// 2. 剥离 HTML/标签：去 <…> 但保留标签间文本
	s = stripTags(s)
	// 3. 折叠空白为一行
	s = collapseWhitespace(s)
	return strings.TrimSpace(s)
}

// jsonUnquote 解析 JSON 字符串字面量（含 \" < 等转义）为原始文本。
func jsonUnquote(s string) (string, error) {
	var out string
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return "", err
	}
	return out, nil
}

// stripTags 去除 <tag> / </tag> 形式的标签，保留标签内文本。
// 不引入正则：扫描遇到 '<' 跳到下一个 '>'，其余字符保留。
func stripTags(s string) string {
	var b strings.Builder
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case inTag:
			// 跳过标签内字符
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// collapseWhitespace 把所有连续空白（含换行/制表符）折叠为单个空格。
func collapseWhitespace(s string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' || r == '\v' || r == '\f' {
			if !prevSpace {
				b.WriteByte(' ')
				prevSpace = true
			}
		} else {
			b.WriteRune(r)
			prevSpace = false
		}
	}
	return b.String()
}

// durationOrUnknown 耗时为空（startTime 缺失/算不出）时显示"用时未知"。
func durationOrUnknown(d string) string {
	if d == "" {
		return "用时未知"
	}
	return d
}
