package request

import (
	"fmt"

	skill_var "github.com/UnicomAI/wanwu/pkg/skill-var"
	url_util "github.com/UnicomAI/wanwu/pkg/url-util"
)

// --- Skill Variable ---

// SkillVariable 变量配置。
// 安全约束（与 assistant_service.SkillVariable 注释一致）：
// VariableValue 仅允许沿 BFF -> assistant-svc -> agent-svc -> sandbox (.skill_env.json) 流动；
// 不得进入 LLM 上下文 (system prompt / SKILL.md / 日志 / 错误信息 / SSE 帧)。
type SkillVariable struct {
	Name          string `json:"name"`
	Desc          string `json:"desc"`
	VariableKey   string `json:"variableKey"`
	VariableValue string `json:"variableValue"`
}

// --- Custom Skill ---

type CreateCustomSkillReq struct {
	Avatar Avatar `json:"avatar" form:"avatar"`
	ZipUrl string `json:"zipUrl" form:"zipUrl" validate:"required"`
}

func (c *CreateCustomSkillReq) Check() error {
	if c.ZipUrl != "" {
		if err := url_util.ValidateURL(c.ZipUrl); err != nil {
			return err
		}
	}
	return nil
}

type CustomSkillIDReq struct {
	SkillId string `json:"skillId" validate:"required"`
}

func (c *CustomSkillIDReq) Check() error {
	return nil
}

type DeleteCustomSkillReq struct {
	SkillId string `json:"skillId" validate:"required"`
}

func (c *DeleteCustomSkillReq) Check() error {
	return nil
}

type CheckCustomSkillReq struct {
	ZipUrl string `json:"zipUrl" form:"zipUrl" validate:"required"`
}

func (c *CheckCustomSkillReq) Check() error {
	if c.ZipUrl != "" {
		if err := url_util.ValidateURL(c.ZipUrl); err != nil {
			return err
		}
	}
	return nil
}

type CustomSkillVersionDownloadReq struct {
	SkillId string `form:"skillId" json:"skillId" validate:"required"`
	Version string `form:"version" json:"version" validate:"required"`
}

func (c *CustomSkillVersionDownloadReq) Check() error {
	return nil
}

// --- Acquired Skill ---

type DeleteAcquiredSkillReq struct {
	SkillId string `json:"skillId" validate:"required"`
}

func (r *DeleteAcquiredSkillReq) Check() error {
	return nil
}

// --- Square Skill ---

type ShareSquareSkillReq struct {
	SkillId string `json:"skillId" validate:"required"`
}

func (r *ShareSquareSkillReq) Check() error {
	return nil
}

// --- Skill Config (Builtin/Custom/Acquired 通用) ---

type SkillConfigReq struct {
	SkillId  string        `json:"skillId" validate:"required"`
	Variable SkillVariable `json:"variable" validate:"required"`
}

func (r *SkillConfigReq) Check() error {
	return r.Variable.validate()
}

type UpdateSkillConfigReq struct {
	ID       string        `json:"id" validate:"required"`
	Variable SkillVariable `json:"variable" validate:"required"`
}

func (r *UpdateSkillConfigReq) Check() error {
	return r.Variable.validate()
}

// validate 共享给 SkillConfigReq / UpdateSkillConfigReq 用，避免两处重复。
// 规则的 key 黑名单与 value 长度上限来自 pkg/skill-var 共享模块，与 sandbox 兜底
// 单一事实源，黑名单更新时不会两边漂移。
func (v *SkillVariable) validate() error {
	if v.Name == "" {
		return fmt.Errorf("variable.name is required")
	}
	if err := skill_var.ValidateVariableKey(v.VariableKey); err != nil {
		return err
	}
	if err := skill_var.ValidateVariableValueLen(v.VariableValue); err != nil {
		return err
	}
	return nil
}

type DeleteSkillConfigReq struct {
	ID string `json:"id" validate:"required"`
}

func (r *DeleteSkillConfigReq) Check() error {
	return nil
}

// --- Callback Skill Search ---

type SearchBuiltinSkillListReq struct {
	SkillIdList []string `json:"skillIdList" form:"skillIdList" validate:"required"`
	// UserId / OrgId 用于 mcp.GetBuiltinSkillVars 鉴权（builtin skill var 按 user 配）。
	// 老 caller 不传时回退为不带 vars 的行为，保持向后兼容。
	UserId string `json:"userId" form:"userId"`
	OrgId  string `json:"orgId" form:"orgId"`
	CommonCheck
}

type SearchCustomSkillListReq struct {
	SkillIdList []string `json:"skillIdList" form:"skillIdList" validate:"required"`
	// UserId/OrgId：调用方期望"以哪个用户身份"查询这批 skill 的详情与变量。
	// callback 由 assistant-service 在处理智能体（Assistant）对话时内部回调过来，
	// 此处应传智能体在 assistant 表里的 user_id / org_id（智能体创建者），
	// 而不是发起 HTTP 请求的调用者——发布态智能体常被非创建者调用，
	// 用调用者身份会查不到创建者配置的 per-user skill 变量。
	UserId string `json:"userId" form:"userId"`
	OrgId  string `json:"orgId" form:"orgId"`
	CommonCheck
}

type SearchAcquiredSkillListReq struct {
	SkillIdList []string `json:"skillIdList" form:"skillIdList" validate:"required"`
	// UserId/OrgId：同 SearchCustomSkillListReq 说明，应传智能体创建者身份，
	// 而非 HTTP 调用者身份。
	UserId string `json:"userId" form:"userId"`
	OrgId  string `json:"orgId" form:"orgId"`
	CommonCheck
}
