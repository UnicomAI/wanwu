package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

// SkillBasicInfo 所有 skill 相关结构体共享的基础字段
type SkillBasicInfo struct {
	SkillId   string         `json:"skillId"`
	Name      string         `json:"name"`
	Avatar    request.Avatar `json:"avatar"`
	Author    string         `json:"author"`
	Desc      string         `json:"desc"`
	OrgId     string         `json:"orgId"`
	UserId    string         `json:"userId"`
	CreatedAt string         `json:"createdAt"`
	UpdatedAt string         `json:"updatedAt"`
}

// SkillVariable 变量配置。
// 安全约束（与 assistant_service.SkillVariable 注释一致）：
// VariableValue 仅允许沿 BFF -> assistant-svc -> agent-svc -> sandbox (.skill_env.json) 流动；
// 不得进入 LLM 上下文 (system prompt / SKILL.md / 日志 / 错误信息 / SSE 帧)。
// 注意 BFF callback HTTP response 会在内部网络携带此字段返回 assistant-svc，gRPC interceptor
// 若 dump request/response body 需要做 redaction
type SkillVariable struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Desc          string `json:"desc"`
	VariableKey   string `json:"variableKey"`
	VariableValue string `json:"variableValue"`
}

// SkillVersionInfo 版本信息（通用）
type SkillVersionInfo struct {
	Version   string `json:"version"`
	Desc      string `json:"desc"`
	UpdatedAt string `json:"updatedAt"`
}

// --- 资源库 内置 skill ---

// BuiltinSkillInfo 资源库/广场-内置skill列表项
type BuiltinSkillInfo struct {
	SkillBasicInfo
	DownloadCount int32 `json:"downloadCount"`
}

// BuiltinSkillDetail 资源库/广场-内置skill详情
type BuiltinSkillDetail struct {
	BuiltinSkillInfo
	SkillMarkdown string           `json:"skillMarkdown"`
	Variables     []*SkillVariable `json:"variables,omitempty"`
}

// --- 资源库 我添加的 skill (acquired) ---

// AcquiredSkillInfo 资源库-我添加的skill列表项
type AcquiredSkillInfo struct {
	SkillBasicInfo
	DownloadCount int32 `json:"downloadCount"`
}

// AcquiredSkillDetail 资源库-我添加的skill详情
type AcquiredSkillDetail struct {
	AcquiredSkillInfo
	SkillMarkdown string           `json:"skillMarkdown"`
	Variables     []*SkillVariable `json:"variables,omitempty"`
}

// --- 广场 skill ---

// SquareSkillInfo 探索广场-skill列表项（内置+共享的混合列表）
type SquareSkillInfo struct {
	SkillBasicInfo
	IsShared bool `json:"isShared"`
}

// --- 广场 共享 skill ---

// SharedSkillInfo 探索广场-共享skill列表项
type SharedSkillInfo struct {
	SkillBasicInfo
	IsShared      bool  `json:"isShared"`
	DownloadCount int32 `json:"downloadCount"`
	AcquiredCount int32 `json:"acquiredCount"`
}

// SharedSkillDetail 探索广场-共享skill详情
type SharedSkillDetail struct {
	SharedSkillInfo
	SkillMarkdown string `json:"skillMarkdown"`
}

// --- 广场/资源库 我发布的 skill（= 自定义 skill）---

// PublishedSkillInfo 我发布的skill列表项（资源库 custom list + 广场 created list 共用）
type PublishedSkillInfo struct {
	SkillBasicInfo
	IsPublished   bool   `json:"isPublished"`
	Version       string `json:"version"`
	PublishType   string `json:"publishType"`
	ThreadID      string `json:"threadId,omitempty"`
	PreviewID     string `json:"previewId,omitempty"`
	DownloadCount int32  `json:"downloadCount"`
	AcquiredCount int32  `json:"acquiredCount"`
}

// PublishedSkillDetail 我发布的skill详情
type PublishedSkillDetail struct {
	PublishedSkillInfo
	Variables     []*SkillVariable `json:"variables,omitempty"`
	SkillMarkdown string           `json:"skillMarkdown,omitempty"`
}

// --- 内部回调用（不对外暴露）---

// SkillDetail 内置skill详情（资源库）
type SkillDetail struct {
	SkillBasicInfo
	SkillMarkdown string           `json:"skillMarkdown"`
	SkillPath     string           `json:"skillPath,omitempty"`
	Variables     []*SkillVariable `json:"variables,omitempty"`
}

// SkillDetailListResp 回调用
type SkillDetailListResp struct {
	SkillList []*SkillDetail `json:"skillList"`
}

// CustomSkillDetailListResp 回调用
type CustomSkillDetailListResp struct {
	SkillList []*CustomSkillListDetail `json:"skillList"`
}

// CustomSkillListDetail 回调用
type CustomSkillListDetail struct {
	SkillBasicInfo
	ObjectPath string           `json:"objectPath,omitempty"`
	Variables  []*SkillVariable `json:"variables,omitempty"`
}

// CallbackAcquiredSkillDetailListResp 回调用
type CallbackAcquiredSkillDetailListResp struct {
	SkillList []*CallbackAcquiredSkillDetail `json:"skillList"`
}

// CallbackAcquiredSkillDetail 回调用
type CallbackAcquiredSkillDetail struct {
	SkillBasicInfo
	ObjectPath string           `json:"objectPath"`
	Variables  []*SkillVariable `json:"variables,omitempty"`
}

// SkillInfo（select 列表用）
type SkillInfo struct {
	SkillBasicInfo
	SkillName string `json:"skillName"`
	SkillType string `json:"skillType"`
}

// CallbackSkillDetail 回调用
type CallbackSkillDetail struct {
	SkillId    string `json:"skillId"`
	SkillType  string `json:"skillType"`
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	Avatar     string `json:"avatar"`
	ObjectPath string `json:"objectPath"`
}

// --- 校验 返回值 ---

// CustomSkillCheckResp 校验自定义skill包返回值
type CustomSkillCheckResp struct {
	Name string `json:"name"`
	Desc string `json:"desc"`
}
