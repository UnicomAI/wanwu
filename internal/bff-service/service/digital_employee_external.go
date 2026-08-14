package service

import (
	"fmt"
	"net/url"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	"github.com/gin-gonic/gin"
)

// errDigitalEmployeeNotReady 数字员工发布对话能力未接入（外部接口未就绪）。
// Phase 2 真正接线后移除。
func errDigitalEmployeeNotReady() error {
	return grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_digital_employee_not_ready", "数字员工发布能力未接入")
}

// --- 外部接口响应结构（对齐《数字员工与万悟交互接口契约》§3.1/3.2 版本最新详情） ---

// digitalEmployeeVersion 数字员工当前发布版本快照
// 单条 latest 的 data；批量 latest 的 entries 每项（字段相同）
type digitalEmployeeVersion struct {
	ID            string                     `json:"id"`            // 版本记录ID
	DhID          string                     `json:"dhId"`          // 数字员工ID
	VersionNo     string                     `json:"versionNo"`     // 版本号
	VersionDesc   string                     `json:"versionDesc"`   // 版本描述
	PublishScope  string                     `json:"publishScope"`  // 发布范围(private/organization/public)
	Creator       digitalEmployeeCreator     `json:"creator"`       // 发布操作人
	CreateTime    int64                      `json:"createTime"`    // 版本创建时间(ms)
	Name          string                     `json:"name"`          // 数字员工名称
	Description   string                     `json:"description"`   // 描述
	Role          string                     `json:"role"`          // 角色设定
	Task          string                     `json:"task"`          // 任务设定
	Workflow      string                     `json:"workflow"`      // 工作流程
	SkillPriority string                     `json:"skillPriority"` // 技能优先级
	Skills        []digitalEmployeeSkill     `json:"skills"`        // 技能配置
	Knowledge     []digitalEmployeeKnowledge `json:"knowledge"`     // 知识网络配置
}

type digitalEmployeeCreator struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type digitalEmployeeKnowledge struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type digitalEmployeeSkill struct {
	SkillID     string `json:"skillId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

// digitalEmployeeInjectInfo 数字员工注入所需字段（通用智能体 @数字员工 与 DE 发布对话共用）
// 两个外部接口（/digital-humans/{id} 非版本、/versions/latest 版本）各自解析响应后映射成本结构，
// 后续注入流程（buildDigitalEmployeeInjection）共用，与来源解耦。
type digitalEmployeeInjectInfo struct {
	Name          string
	Role          string
	Task          string
	Workflow      string
	SkillPriority string
	Knowledge     []digitalEmployeeKnowledge
	Skills        []digitalEmployeeSkill
}

// toDigitalEmployeeInjectInfo 把新接口（/versions/latest）的版本快照映射成注入结构
func toDigitalEmployeeInjectInfo(v *digitalEmployeeVersion) *digitalEmployeeInjectInfo {
	if v == nil {
		return nil
	}
	return &digitalEmployeeInjectInfo{
		Name:          v.Name,
		Role:          v.Role,
		Task:          v.Task,
		Workflow:      v.Workflow,
		SkillPriority: v.SkillPriority,
		Knowledge:     v.Knowledge,
		Skills:        v.Skills,
	}
}

// digitalEmployeeInfoResp 单条 latest 响应（未发布 → data 为 null，契约 §3.1）
type digitalEmployeeInfoResp struct {
	Code int                     `json:"code"`
	Data *digitalEmployeeVersion `json:"data"`
	Msg  string                  `json:"msg"`
}

// digitalEmployeeBatchResp 批量 latest 响应（未发布/不存在的 dh_id 在 entries 缺省，契约 §3.2）
type digitalEmployeeBatchResp struct {
	Code int                      `json:"code"`
	Data digitalEmployeeBatchData `json:"data"`
	Msg  string                   `json:"msg"`
}

type digitalEmployeeBatchData struct {
	Entries []*digitalEmployeeVersion `json:"entries"`
}

// GetDigitalEmployeeInfo 从外部获取数字员工当前发布版本详情（契约 §3.1：GET /versions/latest）。
// 未发布/不存在 → HTTP 200 + data:null → 返回 nil,nil（调用方判定「未发布」）。
// Header 透传 Authorization/X-User-Id/X-Org-Id（workflowHttpReqHeader）+ X-Account-Id。ontology 未开启时返回 nil,nil。
func GetDigitalEmployeeInfo(ctx *gin.Context, userId, orgId, employeeId string) (*digitalEmployeeInjectInfo, error) {
	if config.Cfg().Ontology.Enable == 0 {
		return nil, nil
	}

	endpoint := config.Cfg().Ontology.Endpoint
	infoUri := config.Cfg().Ontology.DigitalEmployeeInfoVersionUri // 含 {dh_id} 占位，只能字符串拼接，不能 url.JoinPath
	requestUrl := endpoint + infoUri

	ret := &digitalEmployeeInfoResp{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("X-Account-Id", userId).
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetPathParam("dh_id", employeeId).
		SetResult(ret).
		Get(requestUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_digital_employee_info", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_digital_employee_info", fmt.Sprintf("[%d] http error", resp.StatusCode()))
	}
	if ret.Data == nil {
		return nil, nil // 未发布/不存在（契约：200 + data:null）
	}
	return toDigitalEmployeeInjectInfo(ret.Data), nil
}

// GetDigitalEmployeeBatchInfo 批量获取数字员工当前发布版本详情（契约 §3.2：POST /versions/latest，body dh_ids）。
// 未发布/不存在的 dh_id 在 entries 缺省（不报错）；返回 dh_id → 版本快照 映射。
func GetDigitalEmployeeBatchInfo(ctx *gin.Context, userId, orgId string, employeeIds []string) (map[string]*digitalEmployeeVersion, error) {
	if config.Cfg().Ontology.Enable == 0 {
		return nil, nil
	}
	if len(employeeIds) == 0 {
		return map[string]*digitalEmployeeVersion{}, nil
	}

	endpoint := config.Cfg().Ontology.Endpoint
	batchUri := config.Cfg().Ontology.DigitalEmployeeBatchLatestUri
	requestUrl, err := url.JoinPath(endpoint, batchUri)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_digital_employee_batch", fmt.Sprintf("build url failed: %v", err))
	}

	ret := &digitalEmployeeBatchResp{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("X-Account-Id", userId).
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetBody(map[string]interface{}{"dh_ids": employeeIds}).
		SetResult(ret).
		Post(requestUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_digital_employee_batch", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_digital_employee_batch", fmt.Sprintf("[%d] http error", resp.StatusCode()))
	}

	result := make(map[string]*digitalEmployeeVersion, len(ret.Data.Entries))
	for _, entry := range ret.Data.Entries {
		if entry == nil || entry.DhID == "" {
			continue
		}
		result[entry.DhID] = entry
	}
	return result, nil
}
