package params_process

import (
	"context"
	"encoding/json"
	"errors"
	net_url "net/url"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/pkg/constant"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
)

type SkillProcess struct {
}

type BuiltinSkillIdListParams struct {
	SkillIdList []string `json:"skillIdList"`
	// UserId / OrgId 用于 BFF callback 内部调 mcp.GetBuiltinSkillVars（builtin var 按 user 配）。
	// 不传时 BFF 端会跳过 var 填充，保持老行为。custom/acquired 不需要这两个字段，
	// 但共用 struct 体可以保持 callback 入参一致——空字段在 JSON 中自然省略。
	UserId string `json:"userId,omitempty"`
	OrgId  string `json:"orgId,omitempty"`
}

type SkillDetailListResp struct {
	Code int64                  `json:"code"`
	Msg  string                 `json:"msg"`
	Data *SkillDetailListResult `json:"data"`
}

type SkillDetailListResult struct {
	SkillList []*SkillDetail `json:"skillList"`
}

type CustomSkillDetailListResp struct {
	Code int64                        `json:"code"`
	Msg  string                       `json:"msg"`
	Data *CustomSkillDetailListResult `json:"data"`
}

type CustomSkillDetailListResult struct {
	SkillList []*CustomSkillDetail `json:"skillList"`
}

type AcquiredSkillDetailListResp struct {
	Code int64                          `json:"code"`
	Msg  string                         `json:"msg"`
	Data *AcquiredSkillDetailListResult `json:"data"`
}

type AcquiredSkillDetailListResult struct {
	SkillList []*AcquiredSkillDetail `json:"skillList"`
}

type SkillDetail struct {
	SkillId       string           `json:"skillId"`             // 模板ID
	Name          string           `json:"name"`                // 模板名称
	Avatar        request.Avatar   `json:"avatar"`              // 模板头像
	Author        string           `json:"author"`              // 作者
	Desc          string           `json:"desc"`                // 模板描述
	SkillMarkdown string           `json:"skillMarkdown"`       // 模板markdown预览
	SkillPath     string           `json:"skillPath,omitempty"` // markdown地址，内部使用，不要对外
	Variables     []*SkillVariable `json:"variables,omitempty"` // 技能自定义变量（来自 BFF callback）
}

type CustomSkillDetail struct {
	SkillId       string           `json:"skillId"`
	Name          string           `json:"name"`
	Avatar        request.Avatar   `json:"avatar"`
	Author        string           `json:"author"`
	Desc          string           `json:"desc"`
	SkillMarkdown string           `json:"skillMarkdown,omitempty"`
	ObjectPath    string           `json:"objectPath,omitempty"`
	Variables     []*SkillVariable `json:"variables,omitempty"` // 技能自定义变量（来自 BFF callback）
}

type AcquiredSkillDetail struct {
	SkillId    string           `json:"skillId"`
	Name       string           `json:"name"`
	Avatar     request.Avatar   `json:"avatar"`
	Author     string           `json:"author"`
	Desc       string           `json:"desc"`
	ObjectPath string           `json:"objectPath"`
	Variables  []*SkillVariable `json:"variables,omitempty"` // 技能自定义变量（来自 BFF callback）
}

// SkillVariable 与 BFF response.SkillVariable JSON shape 对齐；仅保留下游需要的字段。
// VariableValue 注意：完整数据流为
//
//	assistant-svc(Prepare) → HTTP callback BFF → mcp-svc(GetXxxVars) → BFF
//	  → assistant-svc(Build deserialize) → proto SkillInfo (gRPC) → agent-svc
//	  → wga-sandbox (.skill_env.json) → bash 子进程 env
//
// 不得进入任何回流 LLM 的上下文（system prompt / SKILL.md / 日志 / 错误信息 / SSE 帧）。
type SkillVariable struct {
	Name          string `json:"name"`
	Desc          string `json:"desc"`
	VariableKey   string `json:"variableKey"`
	VariableValue string `json:"variableValue"`
}

func init() {
	AddServiceContainer(&SkillProcess{})
}

func (k *SkillProcess) ServiceType() ServiceType {
	return SkillType
}

func (k *SkillProcess) Prepare(agent *AgentInfo, prepareParams *AgentPrepareParams, clientInfo *ClientInfo, userQueryParams *UserQueryParams) error {
	skills := buildAssistantSkills(agent, clientInfo)
	if len(skills) == 0 {
		return nil
	}

	var builtinSkillIds []string
	var customSkillIds []string
	var acquiredSkillIds []string
	for _, skill := range skills {
		if !skill.Enable {
			continue
		}
		switch skill.SkillType {
		case constant.SkillTypeBuiltIn:
			builtinSkillIds = append(builtinSkillIds, skill.SkillId)
		case constant.SkillTypeCustom:
			customSkillIds = append(customSkillIds, skill.SkillId)
		case constant.SkillTypeAcquired:
			acquiredSkillIds = append(acquiredSkillIds, skill.SkillId)
		}
	}
	ctx := userQueryParams.Ctx
	// 三个 skill callback 统一使用智能体（Assistant）创建者身份（agent.Assistant.UserId/OrgId）
	// 而不是 HTTP 调用者身份（userQueryParams.QueryUserId/QueryOrgId）。
	// 智能体发布后常被非创建者调用，用调用者身份会:
	//   1) Builtin skill 变量按用户隔离存储，查不到创建者配置的变量（此前的真实 bug）;
	//   2) Custom/Acquired 虽然目前 mcp 层按 skill 记录 owner 隐式解析，但显式携带
	//      创建者身份能消除隐式依赖，便于日志审计。
	// 注意：userQueryParams.QueryUserId/QueryOrgId 保持不变，因为会话历史归属等逻辑
	// 仍需要 HTTP 调用者身份，不能污染。
	creatorUserId := agent.Assistant.UserId
	creatorOrgId := agent.Assistant.OrgId

	//获取custom skill详情
	if len(customSkillIds) > 0 {
		customSkillResp, err := SearchCustomSkillList(ctx, &BuiltinSkillIdListParams{
			SkillIdList: customSkillIds,
			UserId:      creatorUserId,
			OrgId:       creatorOrgId,
		})
		if err != nil {
			log.Errorf("Assistant服务获取Custom Skill详情失败，assistantId: %d, error: %v", agent.Assistant.ID, err)
			return err
		}
		if customSkillResp.Data != nil {
			prepareParams.CustomSkillList = customSkillResp.Data.SkillList
		}
	}

	//获取acquired skill详情
	if len(acquiredSkillIds) > 0 {
		acquiredSkillResp, err := SearchAcquiredSkillList(ctx, &BuiltinSkillIdListParams{
			SkillIdList: acquiredSkillIds,
			UserId:      creatorUserId,
			OrgId:       creatorOrgId,
		})
		if err != nil {
			log.Errorf("Assistant服务获取Acquired Skill详情失败，assistantId: %d, error: %v", agent.Assistant.ID, err)
			return err
		}
		if acquiredSkillResp.Data != nil {
			prepareParams.AcquiredSkillList = acquiredSkillResp.Data.SkillList
		}
	}

	// 获取builtin skill详情。带 userId/orgId 让 BFF callback 内部调 mcp 拿 per-user vars。
	if len(builtinSkillIds) > 0 {
		resp, err := SearchBuiltInSkillList(ctx, &BuiltinSkillIdListParams{
			SkillIdList: builtinSkillIds,
			UserId:      creatorUserId,
			OrgId:       creatorOrgId,
		})
		if err != nil {
			log.Errorf("Assistant服务获取BuiltIn Skill详情失败，assistantId: %d, error: %v", agent.Assistant.ID, err)
			return err
		}
		prepareParams.builtinSkillList = resp.Data.SkillList
	}
	return nil
}

func (k *SkillProcess) Build(assistant *AgentInfo, prepareParams *AgentPrepareParams, agentChatParams *assistant_service.AgentDetail) error {
	var skillInfos []*assistant_service.SkillInfo
	if len(prepareParams.CustomSkillList) > 0 {
		for _, detail := range prepareParams.CustomSkillList {
			skillInfos = append(skillInfos, &assistant_service.SkillInfo{
				SkillId:    detail.SkillId,
				SkillType:  constant.SkillTypeCustom,
				Name:       detail.Name,
				Desc:       detail.Desc,
				Avatar:     detail.Avatar.Key,
				ObjectPath: detail.ObjectPath,
				Variables:  toProtoSkillVariables(detail.Variables),
			})
		}
	}
	if len(prepareParams.AcquiredSkillList) > 0 {
		for _, detail := range prepareParams.AcquiredSkillList {
			skillInfos = append(skillInfos, &assistant_service.SkillInfo{
				SkillId:    detail.SkillId,
				SkillType:  constant.SkillTypeAcquired,
				Name:       detail.Name,
				Desc:       detail.Desc,
				Avatar:     detail.Avatar.Key,
				ObjectPath: detail.ObjectPath,
				Variables:  toProtoSkillVariables(detail.Variables),
			})
		}
	}
	if len(prepareParams.builtinSkillList) > 0 {
		for _, skill := range prepareParams.builtinSkillList {
			skillInfos = append(skillInfos, buildBuiltInSkillDetail(skill))
		}
	}
	if agentChatParams.SkillParams == nil {
		agentChatParams.SkillParams = &assistant_service.SkillParams{}
	}
	agentChatParams.SkillParams.SkillList = skillInfos
	return nil
}

// toProtoSkillVariables 把本包反序列化后的 SkillVariable 映射到 proto SkillVariable。
// 仅做字段拷贝；任何过滤交给 sandbox 层兜底。注意不打印 VariableValue。
func toProtoSkillVariables(in []*SkillVariable) []*assistant_service.SkillVariable {
	if len(in) == 0 {
		return nil
	}
	out := make([]*assistant_service.SkillVariable, 0, len(in))
	for _, v := range in {
		if v == nil {
			continue
		}
		out = append(out, &assistant_service.SkillVariable{
			Name:          v.Name,
			Desc:          v.Desc,
			VariableKey:   v.VariableKey,
			VariableValue: v.VariableValue,
		})
	}
	return out
}

// SearchCustomSkillList 批量搜索自定义skill详情
func SearchCustomSkillList(ctx context.Context, params *BuiltinSkillIdListParams) (*CustomSkillDetailListResp, error) {
	skillConfig := config.Cfg().Skill
	if skillConfig.CustomSkillListUri == "" {
		return nil, errors.New("custom skill list uri is empty")
	}
	url, _ := net_url.JoinPath(skillConfig.Endpoint, skillConfig.CustomSkillListUri)
	reqBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	result, err := http_client.Default().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       reqBody,
		Timeout:    time.Minute,
		MonitorKey: "custom_skill_detail_list",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var detailResp CustomSkillDetailListResp
	if err = json.Unmarshal(result, &detailResp); err != nil {
		return nil, err
	}
	if detailResp.Code != 0 {
		return nil, errors.New(detailResp.Msg)
	}
	return &detailResp, nil
}

// SearchAcquiredSkillList 批量搜索我添加skill详情
func SearchAcquiredSkillList(ctx context.Context, params *BuiltinSkillIdListParams) (*AcquiredSkillDetailListResp, error) {
	skillConfig := config.Cfg().Skill
	if skillConfig.AcquiredSkillListUri == "" {
		return nil, errors.New("acquired skill list uri is empty")
	}
	url, _ := net_url.JoinPath(skillConfig.Endpoint, skillConfig.AcquiredSkillListUri)
	reqBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	result, err := http_client.Default().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       reqBody,
		Timeout:    time.Minute,
		MonitorKey: "acquired_skill_detail_list",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var detailResp AcquiredSkillDetailListResp
	if err = json.Unmarshal(result, &detailResp); err != nil {
		return nil, err
	}
	if detailResp.Code != 0 {
		return nil, errors.New(detailResp.Msg)
	}
	return &detailResp, nil
}

// SearchBuiltInSkillList 批量搜索内置skill详情
func SearchBuiltInSkillList(ctx context.Context, params *BuiltinSkillIdListParams) (*SkillDetailListResp, error) {
	skillConfig := config.Cfg().Skill
	url, _ := net_url.JoinPath(skillConfig.Endpoint, skillConfig.BuiltInSkillListUri)
	reqBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	result, err := http_client.Default().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       reqBody,
		Timeout:    time.Minute,
		MonitorKey: "builtin_skill_detail_list",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return nil, err
	}
	var detailResp SkillDetailListResp
	if err = json.Unmarshal(result, &detailResp); err != nil {
		return nil, err
	}
	if detailResp.Code != 0 {
		return nil, errors.New(detailResp.Msg)
	}
	return &detailResp, nil
}

// buildBuiltInSkillDetail 构建内置skill详情
func buildBuiltInSkillDetail(skill *SkillDetail) *assistant_service.SkillInfo {
	return &assistant_service.SkillInfo{
		SkillId:    skill.SkillId,
		SkillType:  constant.SkillTypeBuiltIn,
		Name:       skill.Name,
		Desc:       skill.Desc,
		Avatar:     skill.Avatar.Key,
		ObjectPath: skill.SkillPath,
		Variables:  toProtoSkillVariables(skill.Variables),
	}
}

func buildAssistantSkills(agent *AgentInfo, clientInfo *ClientInfo) []*model.AssistantSkill {
	if agent.Draft {
		list, status := clientInfo.Cli.GetAssistantSkillList(context.Background(), agent.Assistant.ID)
		if status != nil {
			log.Errorf("GetAssistantSkillList error: %v", status)
			return nil
		}
		return list
	}
	var skillList []*model.AssistantSkill
	if agent.AssistantSnapshot.AssistantSkillConfig != "" {
		if err := json.Unmarshal([]byte(agent.AssistantSnapshot.AssistantSkillConfig), &skillList); err != nil {
			log.Errorf("GetAssistantSnapshotSkillList error: %v", err)
			return nil
		}
	}
	return skillList
}
