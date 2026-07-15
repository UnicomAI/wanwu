package chat

import (
	"encoding/json"
	"fmt"
	"regexp"
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

// lsLongLineRe 匹配 `ls -l` 长格式输出的一行，捕获文件名（最后一列）。
// 形如 "-rw-r--r-- 1 root root 285741 Jul 10 16:17 鹰.pptx"。
// 权限位 10 字符（含目录 d/链接 l），后跟链接数、属主、属组、大小、日期时间、文件名。
// 文件名可含空格/中文，故日期时间后整体捕获到行尾（再 TrimSpace）。
var lsLongLineRe = regexp.MustCompile(`^[-dlrwxstST]{10}\s+\d+\s+\S+\s+\S+\s+\d+\s+\S+\s+\d+\s+[\d:]{4,5}\s+(.+)$`)

// extractProducedFiles 从已完成的聚合 fragment 树中提取"本次 run 实际产生的产物文件名"。
//
// PPT Agent 等智能体在生成产物后，会用 bash 执行 `ls -l <file>` 确认产物（TOOL_CALL_RESULT
// 返回 ls 长格式输出）。这比"正文子串匹配历史文件名"可靠得多：它直接反映本次 run 操作的文件，
// 不会把正文里偶然出现的多字词（如"日本"）误匹配成历史文件 日本.pptx。
//
// 遍历 fragment 树（含子智能体嵌套），收集所有 toolName=bash 且 result 含 ls -l 行的文件名（basename）。
// 返回去重后的 basename 列表，作为回发工作区文件的强信号白名单。
func extractProducedFiles(topFragments []*wgaFragment) []string {
	seen := make(map[string]struct{})
	var names []string
	var walk func(fs []*wgaFragment)
	walk = func(fs []*wgaFragment) {
		for _, f := range fs {
			if f == nil {
				continue
			}
			if f.kind == fragToolCall && strings.EqualFold(f.toolCallName, "bash") {
				for _, n := range parseLsLongOutput(f.toolCallResult) {
					if n == "" {
						continue
					}
					if _, ok := seen[n]; !ok {
						seen[n] = struct{}{}
						names = append(names, n)
					}
				}
			}
			// 递归子片段（sub_agent 嵌套的 tool_call）
			walk(f.children)
		}
	}
	walk(topFragments)
	return names
}

// parseLsLongOutput 从工具结果文本中解析 `ls -l` 长格式行的文件名（basename）。
// 输入是 TOOL_CALL_RESULT.content 的原始 JSON 文本（常为 JSON 字符串字面量），
// 先按 JSON 字符串 unquote 还原原始文本（保留换行），再逐行匹配权限位开头的长格式行。
// 注意：不能用 cleanToolResult——它会 collapseWhitespace 把多行 ls 输出压成一行，破坏逐行解析。
// 非长格式行（如 ls 报错、其他命令输出）不匹配，天然忽略。
func parseLsLongOutput(raw string) []string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	// TOOL_CALL_RESULT.content 常是 JSON 字符串字面量（如 "\"-rw... 鹰.pptx\\n\""），unquote 还原。
	if len(s) >= 2 && s[0] == '"' {
		if unq, err := jsonUnquote(s); err == nil {
			s = unq
		}
	}
	var names []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if m := lsLongLineRe.FindStringSubmatch(line); len(m) == 2 {
			name := strings.TrimSpace(m[1])
			// "total N" 汇总行不匹配权限位正则，无需特判；空名/符号链接目标跳过
			if name != "" && !strings.HasPrefix(name, "->") {
				names = append(names, name)
			}
		}
	}
	return names
}
