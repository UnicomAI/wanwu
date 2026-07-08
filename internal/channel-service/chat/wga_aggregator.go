package chat

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/channel-service/wanwu"
)

// 本文件是前端 web/src/views/generalAgent/utils/message-aggregator.js 的 Go 移植。
// 把 WGA AG-UI SSE 事件（思考/工具调用/子智能体/工作区/文本/提问）聚合成 fragment 树，
// 再渲染成 markdown 推到钉钉流式卡片，让 IM 用户也能看到完整生成过程。
//
// 与前端的差异：前端每个 fragment 有独立 Vue 组件分轨渲染；钉钉卡片只有一个 markdown 正文，
// 故 renderMarkdown 把 fragment 树扁平化成带 emoji/分隔线的 markdown 文本。

// wgaFragmentKind fragment 类型
type wgaFragmentKind int

const (
	fragReasoning wgaFragmentKind = iota
	fragText
	fragToolCall
	fragActivity
	fragWorkspace
	fragQuestion
)

// wgaFragment 聚合 fragment（对应前端的 message/fragment）
type wgaFragment struct {
	kind      wgaFragmentKind
	content   strings.Builder // reasoning/text 累积文本
	startTime int64           // reasoning/toolCall 耗时起点（ms），0 表示未知
	duration  string          // 完成后耗时（如 "3s"），空表示进行中

	// tool_call
	toolCallID     string
	toolCallName   string
	toolCallArgs   strings.Builder
	toolCallResult string
	finished       bool // tool_call 是否已收到 RESULT

	// activity (sub_agent)
	agentName string
	children  []*wgaFragment // activity 的嵌套 fragments

	// workspace
	fileCount int32
	totalSize int64
	runID     string

	// question
	questionID string
	status     string
	questions  []wanwu.WGAQuestion
}

// wgaEvent 从 SSE 事件映射出的中间结构，喂给聚合器
type wgaEvent struct {
	eventType    string
	delta        string
	toolCallID   string
	toolCallName string
	messageId    string
	timestamp    int64 // ms，0 表示未知
	activityType string
	content      json.RawMessage
}

// wgaAggregator AG-UI 事件聚合器。
// 非线程安全：仅在单个 SSE goroutine（handleWGASSEResponse）内使用。
type wgaAggregator struct {
	topFragments  []*wgaFragment          // 顶层 fragment 列表
	activityStack []*wgaFragment          // 嵌套 sub_agent 栈（栈顶为当前活动）
	toolCallMap   map[string]*wgaFragment // toolCallId → tool_call fragment
}

// newWgaAggregator 创建聚合器
func newWgaAggregator() *wgaAggregator {
	return &wgaAggregator{toolCallMap: make(map[string]*wgaFragment)}
}

// nowMs 当前时间戳（毫秒）
func nowMs() int64 { return time.Now().UnixMilli() }

// addFragment 把 fragment 挂到当前活动（栈顶）的 children，否则挂到顶层
func (a *wgaAggregator) addFragment(f *wgaFragment) {
	if n := len(a.activityStack); n > 0 {
		a.activityStack[n-1].children = append(a.activityStack[n-1].children, f)
	} else {
		a.topFragments = append(a.topFragments, f)
	}
}

// lastFragment 返回当前活动栈顶的最后一个 child，否则返回顶层最后一个 fragment
func (a *wgaAggregator) lastFragment() *wgaFragment {
	if n := len(a.activityStack); n > 0 {
		act := a.activityStack[n-1]
		if len(act.children) > 0 {
			return act.children[len(act.children)-1]
		}
		return nil
	}
	if len(a.topFragments) > 0 {
		return a.topFragments[len(a.topFragments)-1]
	}
	return nil
}

// handleEvent 处理一个 AG-UI 事件，更新聚合状态。返回是否为内容事件（需重渲染）。
func (a *wgaAggregator) handleEvent(ev *wgaEvent) bool {
	ts := ev.timestamp
	if ts == 0 {
		ts = nowMs()
	}
	switch ev.eventType {
	case "REASONING_MESSAGE_START":
		a.addFragment(&wgaFragment{kind: fragReasoning, startTime: ts})
		return true
	case "REASONING_MESSAGE_CONTENT":
		// 末尾 fragment 是 reasoning 则追加；否则兜底新建一个（应对 START 缺失或被中间事件打断）。
		f := a.lastFragment()
		if f == nil || f.kind != fragReasoning {
			f = &wgaFragment{kind: fragReasoning, startTime: ts}
			a.addFragment(f)
		}
		f.content.WriteString(ev.delta)
		return true
	case "REASONING_MESSAGE_END":
		if f := a.lastFragment(); f != nil && f.kind == fragReasoning && f.startTime > 0 {
			f.duration = formatDurationMs(ts - f.startTime)
		}
		return true
	case "TEXT_MESSAGE_START":
		// 显式新建 text fragment（部分 WGA 流可能不发 START，CONTENT 里会兜底新建）
		a.addFragment(&wgaFragment{kind: fragText})
		return true
	case "TEXT_MESSAGE_CONTENT":
		// 末尾已是 text 则追加，否则新建
		f := a.lastFragment()
		if f == nil || f.kind != fragText {
			f = &wgaFragment{kind: fragText}
			a.addFragment(f)
		}
		f.content.WriteString(ev.delta)
		return true
	case "TEXT_MESSAGE_END":
		return false
	case "TOOL_CALL_START":
		f := &wgaFragment{
			kind:         fragToolCall,
			toolCallID:   ev.toolCallID,
			toolCallName: ev.toolCallName,
			startTime:    ts,
		}
		a.toolCallMap[ev.toolCallID] = f
		a.addFragment(f)
		return true
	case "TOOL_CALL_ARGS":
		if f, ok := a.toolCallMap[ev.toolCallID]; ok {
			f.toolCallArgs.WriteString(ev.delta)
		}
		return true
	case "TOOL_CALL_END":
		// 不删 toolCallMap，等 RESULT
		return false
	case "TOOL_CALL_RESULT":
		if f, ok := a.toolCallMap[ev.toolCallID]; ok {
			f.toolCallResult = string(ev.content)
			f.finished = true
			if f.startTime > 0 {
				f.duration = formatDurationMs(ts - f.startTime)
			}
			delete(a.toolCallMap, ev.toolCallID)
		}
		return true
	case "ACTIVITY_SNAPSHOT":
		return a.handleActivitySnapshot(ev, ts)
	}
	return false
}

// handleActivitySnapshot 处理 ACTIVITY_SNAPSHOT 事件
func (a *wgaAggregator) handleActivitySnapshot(ev *wgaEvent, ts int64) bool {
	switch ev.activityType {
	case "sub_agent":
		var c struct {
			AgentName string `json:"agentName"`
			Status    string `json:"status"`
		}
		_ = json.Unmarshal(ev.content, &c)
		if c.Status == "started" {
			a.activityStack = append(a.activityStack, &wgaFragment{
				kind:      fragActivity,
				agentName: c.AgentName,
				startTime: ts,
			})
		} else if c.Status == "finished" {
			if n := len(a.activityStack); n > 0 {
				act := a.activityStack[n-1]
				a.activityStack = a.activityStack[:n-1]
				act.duration = ""
				if act.startTime > 0 {
					act.duration = formatDurationMs(ts - act.startTime)
				}
				if len(a.activityStack) > 0 {
					a.activityStack[len(a.activityStack)-1].children = append(a.activityStack[len(a.activityStack)-1].children, act)
				} else {
					a.topFragments = append(a.topFragments, act)
				}
			}
		}
		return true
	case "workspace":
		var c struct {
			FileCount int32  `json:"fileCount"`
			TotalSize int64  `json:"totalSize"`
			RunID     string `json:"runId"`
		}
		_ = json.Unmarshal(ev.content, &c)
		a.addFragment(&wgaFragment{
			kind:      fragWorkspace,
			fileCount: c.FileCount,
			totalSize: c.TotalSize,
			runID:     c.RunID,
		})
		return true
	case "question":
		var c wanwu.WGAQuestionContent
		_ = json.Unmarshal(ev.content, &c)
		a.addFragment(&wgaFragment{
			kind:       fragQuestion,
			questionID: c.QuestionID,
			status:     c.Status,
			questions:  c.Questions,
		})
		return true
	}
	return false
}

// finalize 流结束时把未关闭的 activity（栈中剩余）挂到顶层
func (a *wgaAggregator) finalize() {
	for len(a.activityStack) > 0 {
		n := len(a.activityStack)
		act := a.activityStack[n-1]
		a.activityStack = a.activityStack[:n-1]
		if len(a.activityStack) > 0 {
			a.activityStack[len(a.activityStack)-1].children = append(a.activityStack[len(a.activityStack)-1].children, act)
		} else {
			a.topFragments = append(a.topFragments, act)
		}
	}
}

// renderMarkdown 把 fragment 树渲染成 markdown（钉钉卡片正文）。
//
// 卡片只展示 TEXT_MESSAGE 段（fragText）和提问提示（fragQuestion），
// 按 API 返回的时间顺序排列；思考/工具调用/子智能体进度/工作区产出等
// 过程类内容不展示到卡片（工作区产物仍通过 sendWorkspaceFiles 以文件形式单独发回 IM）。
// sub_agent（activity）内部发出的 TEXT_MESSAGE 段也按时间顺序并入正文，故递归收集。
// text 段之间用一个空行分隔；末尾多余换行裁掉。
func (a *wgaAggregator) renderMarkdown() string {
	var ordered []*wgaFragment
	for _, f := range a.topFragments {
		collectTextAndQuestion(f, &ordered)
	}
	var b strings.Builder
	for _, f := range ordered {
		switch f.kind {
		case fragText:
			b.WriteString(f.content.String())
			b.WriteString("\n\n")
		case fragQuestion:
			if f.status == "pending" {
				b.WriteString("❓ 智能体需要你确认，请在钉钉回复序号选择（见上方消息）\n\n")
			}
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// collectTextAndQuestion 按出现顺序收集 text 与 question fragment；
// activity 递归进 children（子智能体内部的文本段也属于会话正文），其余 kind 跳过。
func collectTextAndQuestion(f *wgaFragment, out *[]*wgaFragment) {
	switch f.kind {
	case fragText, fragQuestion:
		*out = append(*out, f)
	case fragActivity:
		for _, child := range f.children {
			collectTextAndQuestion(child, out)
		}
	}
}

// formatDurationMs 把毫秒格式化成可读时长（如 "3s"、"1.2s"、"1m5s"）
func formatDurationMs(ms int64) string {
	if ms < 0 {
		return ""
	}
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	sec := float64(ms) / 1000
	if sec < 60 {
		return fmt.Sprintf("%.1fs", sec)
	}
	m := ms / 60000
	s := (ms % 60000) / 1000
	return fmt.Sprintf("%dm%ds", m, s)
}
