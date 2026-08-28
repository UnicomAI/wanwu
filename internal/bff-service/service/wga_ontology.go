package service

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	trace_util "github.com/UnicomAI/wanwu/pkg/trace-util"
	wga_option "github.com/UnicomAI/wanwu/pkg/wga/wga-option"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
)

const (
	defaultDIPAgentName = "本体自动构建专员"
)

type ontologyDigitalEmployeeListResp struct {
	Code int                             `json:"code"`
	Data ontologyDigitalEmployeeListData `json:"data"`
	Msg  string                          `json:"msg"`
}

type ontologyDigitalEmployeeListData struct {
	DigitalEmployees []ontologyDigitalEmployeeBrief `json:"entries"`
	TotalCount       int                            `json:"total_count"`
}

type ontologyDigitalEmployeeBrief struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      int    `json:"status"`
}

func GetGeneralAgentOntologyEmployeeSelect(ctx *gin.Context, userId, orgId, name string) ([]*response.GeneralAgentOntologyEmployee, error) {
	if config.Cfg().Ontology.Enable == 0 {
		return nil, nil
	}

	// 构建URL
	endpoint := config.Cfg().Ontology.Endpoint
	listUri := config.Cfg().Ontology.DigitalEmployeeListUri
	requestUrl, err := url.JoinPath(endpoint, listUri)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_employee_select", fmt.Sprintf("build url failed: %v", err))
	}

	// 构建请求参数
	queryParams := map[string]string{
		"direction": "desc",
		"sort":      "update_time",
		"offset":    "0",
		"limit":     "999", // 默认拉取前999条，满足大部分场景需求
	}
	if name != "" {
		queryParams["name"] = name
	}

	// 发送HTTP请求
	ret := &ontologyDigitalEmployeeListResp{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("X-Account-Id", userId).
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(queryParams).
		SetResult(ret).
		Get(requestUrl)

	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_employee_select", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_employee_select", fmt.Sprintf("[%d] http error", resp.StatusCode()))
	}

	var items []*response.GeneralAgentOntologyEmployee

	// "本体自动构建专员" 固定放在第一个，用于创建知识网络
	items = append(items, &response.GeneralAgentOntologyEmployee{
		ID:     defaultDIPAgentName,
		Name:   defaultDIPAgentName,
		Desc:   "可根据业务需求文档自动化进行领域知识网络构建和更新的数字员工，负责领域知识网络的需求理解、知识网络建模及更新维护操作。",
		Avatar: request.Avatar{Path: "/v1/static/icon/wga-digital-employee-icon.svg"},
	})
	for _, de := range ret.Data.DigitalEmployees {
		if de.Status != 1 { // 仅返回状态为启用的数字员工
			continue
		}
		items = append(items, &response.GeneralAgentOntologyEmployee{
			ID:     de.ID,
			Name:   de.Name,
			Desc:   de.Description,
			Avatar: request.Avatar{Path: "/v1/static/icon/wga-digital-employee-icon.svg"},
		})
	}

	return items, nil
}

// ontologyDigitalEmployeeInfoResp 旧接口 /digital-humans/{digitalEmployeeId}（vega 非版本）响应
// knowledge/skills 直接复用 digitalEmployeeKnowledge/Skill 结构（字段一致）
type ontologyDigitalEmployeeInfoResp struct {
	Code int `json:"code"`
	Data struct {
		ID            string                     `json:"id"`
		Name          string                     `json:"name"`
		Role          string                     `json:"role"`
		Task          string                     `json:"task"`
		Workflow      string                     `json:"workflow"`
		SkillPriority string                     `json:"skillPriority"`
		Knowledge     []digitalEmployeeKnowledge `json:"knowledge"`
		Skills        []digitalEmployeeSkill     `json:"skills"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// GetGeneralAgentOntologyEmployeeInfo 获取数字员工详情（通用智能体 @数字员工 注入用）。
// 走旧接口 /vega/api/digital-humans/{digitalEmployeeId}（vega 非版本），区别于 DE 发布场景的 /versions/latest。
// 返回注入所需字段（digitalEmployeeInjectInfo），与 DE 适配器共用后续注入流程。
func GetGeneralAgentOntologyEmployeeInfo(ctx *gin.Context, userId, orgId, digitalEmployeeId string) (*digitalEmployeeInjectInfo, error) {
	if config.Cfg().Ontology.Enable == 0 {
		return nil, nil
	}

	endpoint := config.Cfg().Ontology.Endpoint
	infoUri := config.Cfg().Ontology.DigitalEmployeeInfoUri // 含 {digitalEmployeeId} 占位，只能字符串拼接，不能 url.JoinPath
	requestUrl := endpoint + infoUri

	ret := &ontologyDigitalEmployeeInfoResp{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("X-Account-Id", userId).
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetPathParam("digitalEmployeeId", digitalEmployeeId).
		SetResult(ret).
		Get(requestUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_employee_info", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_employee_info", fmt.Sprintf("[%d] http error", resp.StatusCode()))
	}
	if ret.Data.ID == "" {
		return nil, nil // 不存在
	}
	return &digitalEmployeeInjectInfo{
		Name:          ret.Data.Name,
		Role:          ret.Data.Role,
		Task:          ret.Data.Task,
		Workflow:      ret.Data.Workflow,
		SkillPriority: ret.Data.SkillPriority,
		Knowledge:     ret.Data.Knowledge,
		Skills:        ret.Data.Skills,
	}, nil
}

// --- internal wga ontology ---

// ontologyKnowledgeNetworkListResp 知识网络列表响应
type ontologyKnowledgeNetworkListResp struct {
	KnowledgeNetworks []ontologyKnowledgeNetwork `json:"entries"`
	TotalCount        int                        `json:"total_count"`
}

// ontologyKnowledgeNetwork 知识网络
type ontologyKnowledgeNetwork struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Comment    string `json:"comment"`
	ModuleType string `json:"module_type"`
}

// getOntologyKnowledgeSelect 获取知识网络选择列表
func getOntologyKnowledgeSelect(ctx *gin.Context, userId, orgId, name string) ([]*response.GeneralAgentResourceSelectItem, error) {
	if config.Cfg().Ontology.Enable == 0 {
		return nil, nil
	}
	// 构建URL
	endpoint := config.Cfg().Ontology.Endpoint
	listUri := config.Cfg().Ontology.KnowledgeNetworkListUri
	requestUrl, err := url.JoinPath(endpoint, listUri)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_knowledge_select", fmt.Sprintf("build url failed: %v", err))
	}

	// 构建请求参数
	queryParams := map[string]string{
		"direction": "desc",
		"sort":      "update_time",
		"offset":    "0",
		"limit":     "999", // 默认拉取前999条，满足大部分场景需求
	}
	if name != "" {
		queryParams["name"] = name
	}

	// 发送HTTP请求
	ret := &ontologyKnowledgeNetworkListResp{}
	resp, err := trace_util.NewResty(ctx).
		R().
		SetContext(ctx.Request.Context()).
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept", "application/json").
		SetHeader("X-Account-Id", userId).
		SetHeaders(workflowHttpReqHeader(ctx)).
		SetQueryParams(queryParams).
		SetResult(ret).
		Get(requestUrl)

	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_knowledge_select", err.Error())
	}
	if resp.StatusCode() >= 300 {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_knowledge_select", fmt.Sprintf("[%d] http error", resp.StatusCode()))
	}

	// 转换为统一的响应格式
	items := make([]*response.GeneralAgentResourceSelectItem, 0, len(ret.KnowledgeNetworks))
	for _, kn := range ret.KnowledgeNetworks {
		items = append(items, &response.GeneralAgentResourceSelectItem{
			ID:     kn.ID,
			Name:   kn.Name,
			Desc:   kn.Comment,
			Type:   kn.ModuleType,
			Avatar: request.Avatar{Path: config.Cfg().DefaultIcon.KnowledgeIcon},
		})
	}

	return items, nil
}

// checkWgaOntologyKnowledgeConfig 校验 wga Ontology 配置（用于更新配置）
func checkWgaOntologyKnowledgeConfig(ctx *gin.Context, userId, orgId string, ontologyKnowledgeList []*assistant_service.WgaConfigOntologyKnowledge) error {
	if len(ontologyKnowledgeList) == 0 {
		return nil
	}

	ontologyIds := make([]string, 0, len(ontologyKnowledgeList))
	for _, o := range ontologyKnowledgeList {
		ontologyIds = append(ontologyIds, o.OntologyKnowledgeId)
	}

	validIds, err := getValidOntologyIds(ctx, userId, orgId, ontologyIds)
	if err != nil {
		return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr, "ontology not found")
	}

	for _, o := range ontologyKnowledgeList {
		if !validIds[o.OntologyKnowledgeId] {
			return grpc_util.ErrorStatus(errs.Code_WgaConfigCheckErr,
				fmt.Sprintf("ontology not found: %s", o.OntologyKnowledgeId))
		}
	}
	return nil
}

// buildWgaOntologyDIPMode 构建 wga Ontology 配置选项（DIP Agent 模式）
//
// DIP Agent 模式：用于创建本体知识网络或使用数字员工执行任务。
// 通过 WithInstruction 覆盖 prompt.md、WithOverallTask 传入数字员工 task、WithSkill 注入技能。
// "本体自动构建专员"（默认数字员工）仅注入 userId 提示，不渲染 prompt 模板。
func buildWgaOntologyDIPMode(ctx *gin.Context, userId, orgId, threadId, runId, text string) ([]wga_option.Option, *schema.Message, error) {

	dipAgentName := defaultDIPAgentName
	if strings.HasPrefix(text, "@") {
		if spaceIndex := strings.Index(text, " "); spaceIndex < 0 {
			// @后面没有空格，整个文本都是数字员工名称
			dipAgentName = text[1:]
		} else if spaceIndex >= 1 {
			// @后面有空格，数字员工名称是@和第一个空格之间的部分
			dipAgentName = text[1:spaceIndex]
		}
		// else 非法的@提及格式，默认使用"本体自动构建专员"
	}

	// --- "本体自动构建专员" ---
	if dipAgentName == defaultDIPAgentName {
		var contentBuilder strings.Builder
		_, _ = fmt.Fprintf(&contentBuilder, "【当前用户ID为：%s】如果技能参数需要 userId、user-id、accountId、x-account-id 等相关信息，可以使用此用户ID进行查询。\n", userId)
		return nil, &schema.Message{
			Role:    schema.System,
			Content: contentBuilder.String(),
		}, nil
	}

	// --- 数字员工 ---
	dipEmployees, err := GetGeneralAgentOntologyEmployeeSelect(ctx, userId, orgId, "")
	if err != nil {
		return nil, nil, err
	}
	var dipEmployee *response.GeneralAgentOntologyEmployee
	for _, employee := range dipEmployees {
		if employee.Name == dipAgentName {
			dipEmployee = employee
			break
		}
	}
	if dipEmployee == nil {
		return nil, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_employee_select", fmt.Sprintf("dip agent (%v) not found", dipAgentName))
	}
	dipAgent, err := GetGeneralAgentOntologyEmployeeInfo(ctx, userId, orgId, dipEmployee.ID)
	if err != nil {
		return nil, nil, err
	}
	if dipAgent == nil { // 未发布/不存在
		return nil, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_employee_info", fmt.Sprintf("dip agent (%v) not found", dipAgentName))
	}

	// 数字员工详情注入（instruction/overallTask/knowledge/skills + 系统提示）
	return buildDigitalEmployeeInjection(ctx, userId, orgId, threadId, runId, dipAgent)
}

// buildWgaDigitalEmployeeMode 构建数字员工发布对话的 Ontology 配置选项
//
// 与 buildWgaOntologyDIPMode（@name 文本解析 + 列表搜名）不同，试用对话的 employeeId 由请求显式给出，
// 直接按 ID 拉取员工详情并注入（instruction/overallTask/knowledge/skills）。
// 统一走 DE 版本化详情接口（GetDigitalEmployeeInfo → /versions/latest）。
func buildWgaDigitalEmployeeMode(ctx *gin.Context, userId, orgId, threadId, runId, employeeId string) ([]wga_option.Option, *schema.Message, error) {
	employeeId = strings.TrimSpace(employeeId)
	if employeeId == "" {
		return nil, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_employee_info", "digital employee id is required")
	}
	dipAgent, err := GetDigitalEmployeeInfo(ctx, userId, orgId, employeeId)
	if err != nil {
		return nil, nil, err
	}
	if dipAgent == nil { // 未发布/不存在
		return nil, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_employee_info", fmt.Sprintf("digital employee (%v) not found", employeeId))
	}
	return buildDigitalEmployeeInjection(ctx, userId, orgId, threadId, runId, dipAgent)
}

// buildDigitalEmployeeInjection 构建数字员工详情注入（instruction + overallTask + knowledge + skills + 系统提示）
// 供 buildWgaOntologyDIPMode（@数字员工）与 buildWgaDigitalEmployeeMode（发布对话）共用，直接消费契约结构 digitalEmployeeVersion。
func buildDigitalEmployeeInjection(ctx *gin.Context, userId, orgId, threadId, runId string, de *digitalEmployeeInjectInfo) ([]wga_option.Option, *schema.Message, error) {
	var contentBuilder strings.Builder
	_, _ = fmt.Fprintf(&contentBuilder, "【当前用户ID为：%s】如果技能参数需要 userId、user-id、accountId、x-account-id 等相关信息，可以使用此用户ID进行查询。\n", userId)

	var opts []wga_option.Option
	// 数字员工的 instruction
	instruction, err := formatWgaOntologyDIPInstruction(de.Name, de.Role, de.Workflow, de.SkillPriority)
	if err != nil {
		return nil, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_ontology_dip_prompt", fmt.Sprintf("render digital employee prompt failed: %v", err))
	}
	opts = append(opts, wga_option.WithInstruction(instruction))
	// 数字员工的 task 作为整体任务传入
	if de.Task != "" {
		opts = append(opts, wga_option.WithOverallTask(de.Task))
	}
	// 数字员工的 knowledge（本体知识网络）
	if len(de.Knowledge) > 0 {
		_, _ = fmt.Fprintf(&contentBuilder, "\n【当前配置的知识网络为：%s】%s", de.Knowledge[0].Name, de.Knowledge[0].Description)
		_, _ = fmt.Fprintf(&contentBuilder, "\n【当前配置的知识网络ID为：%s】如果需要用到本体智能体知识网络，可以使用此知识网络ID作为 kn_id 或 KN ID 等相关参数进行查询。", de.Knowledge[0].ID)
	}
	// 数字员工的 skills
	if len(de.Skills) > 0 {
		_, _ = fmt.Fprintf(&contentBuilder, "\n\n【如果需要，优先使用以下skills】")
		var skillList []*assistant_service.WgaConfigSkill
		for _, skill := range de.Skills {
			if skill.Type == constant.SkillTypeBuiltIn { // 内置ontology skills都挂载到wga-sandbox-ontology-wanwu容器了
				_, _ = fmt.Fprintf(&contentBuilder, "\n- %s: %s", path.Base(skill.SkillID), skill.Description)
			} else {
				_, _ = fmt.Fprintf(&contentBuilder, "\n- %s: %s", skill.Name, skill.Description)
				skillList = append(skillList, &assistant_service.WgaConfigSkill{
					SkillId:   skill.SkillID,
					SkillType: skill.Type,
				})
			}
		}
		skillOpts, err := buildWgaSkillOptions(ctx, userId, orgId, threadId, runId, skillList)
		if err != nil {
			return nil, nil, err
		}
		opts = append(opts, skillOpts...)
	}

	return opts, &schema.Message{
		Role:    schema.System,
		Content: contentBuilder.String(),
	}, nil
}

// buildWgaOntologyNonDIPMode 构建 wga Ontology 配置选项（非 DIP Agent 模式）
//
// 非 DIP Agent 模式：可用于本体知识网络问数，确定知识网络 ID：@提及 > 配置表，无配置则不加载；最多只有一个知识网络 ID
//  1. 系统提示：告知 userId 和知识网络 ID，提示 agent 如何使用这些信息
//  2. 数据查询技能：ontology-core / smart-ask-data / smart-data-analysis / smart-search-tables / smart-data-collection / yj-ocr-parser
func buildWgaOntologyNonDIPMode(ctx *gin.Context, userId, orgId string,
	ontologyKnowledgeMentions, ontologyKnowledgeList []*assistant_service.WgaConfigOntologyKnowledge) ([]wga_option.Option, *schema.Message, error) {

	var ontologyKnowledgeId string
	if len(ontologyKnowledgeMentions) > 0 {
		ontologyKnowledgeId = ontologyKnowledgeMentions[0].OntologyKnowledgeId
	} else if len(ontologyKnowledgeList) > 0 {
		ontologyKnowledgeId = ontologyKnowledgeList[0].OntologyKnowledgeId
	} else {
		return nil, nil, nil
	}

	var contentBuilder strings.Builder
	_, _ = fmt.Fprintf(&contentBuilder, "【当前用户ID为：%s】如果技能参数需要 userId、user-id、accountId、account-id、x-account-id 等相关信息，可以使用此用户ID进行查询。\n", userId)
	_, _ = fmt.Fprintf(&contentBuilder, "\n【当前配置的知识网络ID为：%s】如果需要用到本体智能体知识网络的智能问答能力，可以使用此知识网络ID作为 kn_id 或 KN ID 等相关参数进行查询。", ontologyKnowledgeId)

	// 从配置读取数据查询技能列表，动态注入
	smartSkills := config.Cfg().Ontology.SmartDataSkills
	var opts []wga_option.Option
	if len(smartSkills) > 0 {
		_, _ = contentBuilder.WriteString("\n\n已加载以下数据查询相关技能，可以使用这些技能访问知识网络中的数据：")
		for _, s := range smartSkills {
			_, _ = fmt.Fprintf(&contentBuilder, "\n- %s: %s", s.Name, s.Desc)
			opts = append(opts, wga_option.WithSkill(wga_option.Skill{
				Dir: s.SkillPath,
			}))
		}
	}

	return opts, &schema.Message{
		Role:    schema.System,
		Content: contentBuilder.String(),
	}, nil
}

// getValidOntologyIds 批量获取有效的 Ontology ID 映射
func getValidOntologyIds(ctx *gin.Context, userId, orgId string, ontologyIds []string) (map[string]bool, error) {
	if len(ontologyIds) == 0 {
		return make(map[string]bool), nil
	}

	// 获取所有 ontology 列表
	items, err := getOntologyKnowledgeSelect(ctx, userId, orgId, "")
	if err != nil {
		return nil, err
	}

	// 构建有效 ID 映射
	validIds := make(map[string]bool)
	for _, item := range items {
		validIds[item.ID] = true
	}

	return validIds, nil
}

// formatWgaOntologyDIPInstruction 渲染数字员工 instruction 模板。
// 使用 Jinja2 模板语法，与 wga prompt.md 一致：
//
//	agent_name     ← dipAgent.Name
//	agent_role     ← dipAgent.Role
//	current_time   ← time.Now()
//	skill_priority ← dipAgent.SkillPriority
//	workflow       ← dipAgent.Workflow（非空时注入 # 任务 区块，空时省略）
func formatWgaOntologyDIPInstruction(agentName, agentRole, workflow, skillPriority string) (string, error) {
	tmplStr := config.Cfg().Ontology.DigitalEmployeePromptTemplate
	if tmplStr == "" {
		return "", fmt.Errorf("digital_employee_prompt_template not configured")
	}

	var workflowSection string
	if workflow != "" {
		workflowSection = workflow + "\n\n"
	}

	vs := map[string]any{
		"agent_name":     agentName,
		"agent_role":     agentRole,
		"current_time":   time.Now().Format(time.RFC1123),
		"workflow":       workflowSection,
		"skill_priority": skillPriority,
	}

	rets, err := prompt.FromMessages(schema.Jinja2, schema.SystemMessage(tmplStr)).Format(context.Background(), vs)
	if err != nil {
		return "", fmt.Errorf("render digital employee prompt template failed: %w", err)
	}
	for _, ret := range rets {
		return ret.Content, nil
	}
	return "", nil
}
