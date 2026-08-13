package orm

import (
	"encoding/json"
	"net/url"
	"strings"
)

// 跳转动作的 msgType 取值（与 actionParams 的 schema 映射表一一对应：各 msgType 的必填字段见
// msgTypeRequiredFields，归一化与占位符规则见 bff 侧 normalizeActionParams）
const (
	MsgTypeAgent     = "agent"
	MsgTypeWorkflow  = "workflow"
	MsgTypeChatflow  = "chatflow"
	MsgTypeRag       = "rag"
	MsgTypeSkill     = "skill"
	MsgTypeKnowledge = "knowledge"
	MsgTypeModel     = "model"
	MsgTypeAbout     = "about"
	MsgTypeCustomURL = "custom_url"
)

// 动作类型
const (
	ActionTypeLink  = "link"  // 当前页直接跳
	ActionTypeBlank = "blank" // 新页面打开
)

// actionParams 的取值约束（字段值长度与嵌套深度的通用规则）
const (
	actionParamMaxValueLen = 512
	actionParamMaxDepth    = 2
)

// Action 跳转动作。content 中每个 $${超链文字}$$ 占位按出现顺序对应 actions[i]。
type Action struct {
	MsgType      string
	ActionType   string
	ActionParams map[string]interface{}
}

// msgTypeRequiredFields 逐 msgType 的必填字段（未知 msgType / 未知字段 → 写入侧拒绝）
var msgTypeRequiredFields = map[string][]string{
	MsgTypeAgent:     {"appId", "appType"},
	MsgTypeWorkflow:  {"appId", "appType"},
	MsgTypeChatflow:  {"appId", "appType"},
	MsgTypeRag:       {"appId", "appType"},
	MsgTypeSkill:     {"appId", "appType"},
	MsgTypeKnowledge: {"knowledgeId"},
	MsgTypeModel:     {"modelId"},
	MsgTypeAbout:     {},
	MsgTypeCustomURL: {"url"},
}

// validActionTypes 允许的 actionType
var validActionTypes = map[string]struct{}{
	ActionTypeLink:  {},
	ActionTypeBlank: {},
}

// sanitizeActions 按各 msgType 的必填字段 schema 校验跳转集合，返回可落库的 JSON 字符串
// 与被剔除的动作数量。
//
// 校验不通过的 action 被剔除而不是整条消息失败——消息主体仍然有价值，
// 对应的 $${}$$ 占位符由前端降级为纯文本展示。
// 空集合落 "[]" 而非 NULL，让读侧不必区分两种"无跳转"。
func sanitizeActions(actions []Action) (string, int) {
	valid := make([]map[string]interface{}, 0, len(actions))
	var dropped int
	for _, a := range actions {
		if !isValidAction(a) {
			dropped++
			continue
		}
		params := a.ActionParams
		if params == nil {
			params = map[string]interface{}{}
		}
		valid = append(valid, map[string]interface{}{
			"msgType":      a.MsgType,
			"actionType":   a.ActionType,
			"actionParams": params,
		})
	}
	b, err := json.Marshal(valid)
	if err != nil {
		return "[]", len(actions)
	}
	return string(b), dropped
}

func isValidAction(a Action) bool {
	required, ok := msgTypeRequiredFields[a.MsgType]
	if !ok {
		return false // 未知 msgType
	}
	if _, ok := validActionTypes[a.ActionType]; !ok {
		return false
	}
	for _, field := range required {
		v, ok := a.ActionParams[field]
		if !ok {
			return false
		}
		s, ok := v.(string)
		if !ok || s == "" {
			return false
		}
	}
	for _, v := range a.ActionParams {
		if !isValidParamValue(v, 1) {
			return false
		}
	}
	// custom_url 只允许 http/https；javascript:/data: 等一律拒绝
	if a.MsgType == MsgTypeCustomURL {
		raw, _ := a.ActionParams["url"].(string)
		if !isAllowedURL(raw) {
			return false
		}
	}
	return true
}

// isValidParamValue 限制字段值长度 ≤512、JSON 深度 ≤2
func isValidParamValue(v interface{}, depth int) bool {
	switch val := v.(type) {
	case string:
		return len(val) <= actionParamMaxValueLen
	case bool, float64, int, int32, int64, nil:
		return true
	case map[string]interface{}:
		if depth >= actionParamMaxDepth {
			return false
		}
		for _, sub := range val {
			if !isValidParamValue(sub, depth+1) {
				return false
			}
		}
		return true
	case []interface{}:
		if depth >= actionParamMaxDepth {
			return false
		}
		for _, sub := range val {
			if !isValidParamValue(sub, depth+1) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func isAllowedURL(raw string) bool {
	if raw == "" {
		return false
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	switch strings.ToLower(u.Scheme) {
	case "http", "https":
		return u.Host != ""
	default:
		return false
	}
}

// ParseActions 把落库的 actions JSON 解析回结构体，供读侧下发。
// 解析失败返回空集合——读路径不因单条脏数据整体失败。
func ParseActions(raw string) []Action {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var items []struct {
		MsgType      string                 `json:"msgType"`
		ActionType   string                 `json:"actionType"`
		ActionParams map[string]interface{} `json:"actionParams"`
	}
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return nil
	}
	ret := make([]Action, 0, len(items))
	for _, it := range items {
		ret = append(ret, Action{
			MsgType:      it.MsgType,
			ActionType:   it.ActionType,
			ActionParams: it.ActionParams,
		})
	}
	return ret
}
