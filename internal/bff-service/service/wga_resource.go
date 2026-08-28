package service

import (
	"regexp"
	"strings"
	"sync"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

// wgaResourceNameRegex 用于从文本中提取 @提及 的资源名称。
// 格式: @name，支持中文、英文字母、数字、下划线和连字符。
// 例如: "@工具名"、"@workflow-1"、"@skill_2"
var wgaResourceNameRegex = regexp.MustCompile(`@([\p{Han}a-zA-Z0-9_-]+)`)

func GetGeneralAgentResourceSelect(ctx *gin.Context, userId, orgId string, name string) ([]*response.GeneralAgentResourceSelectList, error) {
	result := make([]*response.GeneralAgentResourceSelectList, 0, 5)

	// 并发获取六种资源列表
	var wg sync.WaitGroup
	var mcpErr, workflowErr, skillErr, assistantErr, knowledgeErr, ontologyErr error
	var mcpList []response.MCPSelect
	var workflowList []*response.ExplorationAppInfo
	var skillList []*response.SkillInfo
	var assistantList []*response.ExplorationAppInfo
	var knowledgeList *response.KnowledgeListResp
	var ontologyList []*response.GeneralAgentResourceSelectItem

	// 获取 MCP 列表
	wg.Add(1)
	go func() {
		defer util.PrintPanicStack()
		defer wg.Done()
		resp, err := GetMCPSelect(ctx, userId, orgId, name)
		if err != nil {
			mcpErr = err
			return
		}
		if list, ok := resp.List.([]response.MCPSelect); ok {
			mcpList = list
		}
	}()

	// 获取 Workflow 列表
	wg.Add(1)
	go func() {
		defer util.PrintPanicStack()
		defer wg.Done()
		resp, err := GetWorkflowSelect(ctx, userId, orgId, request.GetExplorationAppListRequest{Name: name})
		if err != nil {
			workflowErr = err
			return
		}
		if list, ok := resp.List.([]*response.ExplorationAppInfo); ok {
			workflowList = list
		}
	}()

	// 获取 Skill 列表（custom 已发布 + acquired，不包括 builtin）
	wg.Add(1)
	go func() {
		defer util.PrintPanicStack()
		defer wg.Done()
		// custom、acquired
		var allSkills []*response.SkillInfo
		// 获取已发布的 custom skill
		customResp, err := GetSkillSelect(ctx, userId, orgId, name, constant.SkillTypeCustom, false)
		if err != nil {
			skillErr = err
			return
		}
		if list, ok := customResp.List.([]*response.SkillInfo); ok {
			allSkills = append(allSkills, list...)
		}
		// 获取 acquired skill
		acquiredResp, err := GetSkillSelect(ctx, userId, orgId, name, constant.SkillTypeAcquired, true)
		if err != nil {
			skillErr = err
			return
		}
		if list, ok := acquiredResp.List.([]*response.SkillInfo); ok {
			allSkills = append(allSkills, list...)
		}
		skillList = allSkills
	}()

	// 获取 Assistant 列表
	wg.Add(1)
	go func() {
		defer util.PrintPanicStack()
		defer wg.Done()
		resp, err := GetAssistantSelectSingle(ctx, userId, orgId, request.GetExplorationAppListRequest{Name: name})
		if err != nil {
			assistantErr = err
			return
		}
		if list, ok := resp.List.([]*response.ExplorationAppInfo); ok {
			assistantList = list
		}
	}()

	// 获取 Knowledge 列表
	wg.Add(1)
	go func() {
		defer util.PrintPanicStack()
		defer wg.Done()
		resp, err := SelectKnowledgeList(ctx, userId, orgId, &request.KnowledgeSelectReq{Name: name})
		if err != nil {
			knowledgeErr = err
			return
		}
		knowledgeList = resp
	}()

	// 获取 Ontology 列表
	wg.Add(1)
	go func() {
		defer util.PrintPanicStack()
		defer wg.Done()
		list, err := getOntologyKnowledgeSelect(ctx, userId, orgId, name)
		if err != nil {
			ontologyErr = err
			return
		}
		ontologyList = list
	}()

	wg.Wait()

	// 检查错误
	if mcpErr != nil {
		return nil, mcpErr
	}
	if workflowErr != nil {
		return nil, workflowErr
	}
	if skillErr != nil {
		return nil, skillErr
	}
	if assistantErr != nil {
		return nil, assistantErr
	}
	if knowledgeErr != nil {
		return nil, knowledgeErr
	}
	if ontologyErr != nil {
		return nil, ontologyErr
	}

	// 构建 MCP 列表
	mcpItems := make([]*response.GeneralAgentResourceSelectItem, 0, len(mcpList))
	for _, item := range mcpList {
		mcpItems = append(mcpItems, &response.GeneralAgentResourceSelectItem{
			ID:     item.MCPID,
			Name:   item.Name,
			Desc:   item.Description,
			Avatar: item.Avatar,
			Type:   item.Type,
		})
	}
	result = append(result, &response.GeneralAgentResourceSelectList{
		ListType: "mcp",
		List:     mcpItems,
	})

	// 构建 Workflow 列表
	workflowItems := make([]*response.GeneralAgentResourceSelectItem, 0, len(workflowList))
	for _, item := range workflowList {
		workflowItems = append(workflowItems, &response.GeneralAgentResourceSelectItem{
			ID:     item.AppId,
			Name:   item.Name,
			Desc:   item.Desc,
			Avatar: item.Avatar,
			Type:   item.AppType,
		})
	}
	result = append(result, &response.GeneralAgentResourceSelectList{
		ListType: "workflow",
		List:     workflowItems,
	})

	// 构建 Skill 列表
	skillItems := make([]*response.GeneralAgentResourceSelectItem, 0, len(skillList))
	for _, item := range skillList {
		skillItems = append(skillItems, &response.GeneralAgentResourceSelectItem{
			ID:     item.SkillId,
			Name:   item.SkillName,
			Desc:   item.Desc,
			Avatar: item.Avatar,
			Type:   item.SkillType,
			Author: item.Author,
		})
	}
	result = append(result, &response.GeneralAgentResourceSelectList{
		ListType: "skill",
		List:     skillItems,
	})

	// 构建 Assistant 列表
	assistantItems := make([]*response.GeneralAgentResourceSelectItem, 0, len(assistantList))
	for _, item := range assistantList {
		assistantItems = append(assistantItems, &response.GeneralAgentResourceSelectItem{
			ID:     item.AppId,
			Name:   item.Name,
			Desc:   item.Desc,
			Avatar: item.Avatar,
			Type:   item.AppType,
		})
	}
	result = append(result, &response.GeneralAgentResourceSelectList{
		ListType: "assistant",
		List:     assistantItems,
	})

	// 构建 Knowledge 列表
	knowledgeItems := make([]*response.GeneralAgentResourceSelectItem, 0)
	if knowledgeList != nil {
		knowledgeItems = make([]*response.GeneralAgentResourceSelectItem, 0, len(knowledgeList.KnowledgeList))
		for _, item := range knowledgeList.KnowledgeList {
			knowledgeItems = append(knowledgeItems, &response.GeneralAgentResourceSelectItem{
				ID:     item.KnowledgeId,
				Name:   item.Name,
				Desc:   item.Description,
				Avatar: item.Avatar,
				Type:   "knowledge",
			})
		}
	}
	result = append(result, &response.GeneralAgentResourceSelectList{
		ListType: "knowledge",
		List:     knowledgeItems,
	})

	// 构建 Ontology 列表
	result = append(result, &response.GeneralAgentResourceSelectList{
		ListType: "ontology",
		List:     ontologyList,
	})

	return result, nil
}

// parseResourceMentions 从用户消息中解析 @ 提及的资源名称
// 格式: @资源名称 后面跟空格或消息结束
// 支持: "@mcp1 @workflow2 请帮我处理" -> ["mcp1", "workflow2"]
// 返回: 提取到的资源名称列表（去重后）
func parseWgaResourceMentions(text string) []string {
	// 使用正则提取 @name 格式，支持中文、英文、数字、下划线、连字符
	matches := wgaResourceNameRegex.FindAllStringSubmatch(text, -1)

	// 去重
	seen := make(map[string]bool)
	var names []string
	for _, m := range matches {
		if len(m) > 1 && !seen[m[1]] {
			seen[m[1]] = true
			names = append(names, m[1])
		}
	}

	return names
}

// wgaMentionResources @提及的资源列表
type wgaMentionResources struct {
	McpList       []*assistant_service.WgaConfigMcp
	WorkflowList  []*assistant_service.WgaConfigWorkflow
	SkillList     []*assistant_service.WgaConfigSkill
	AssistantList []*assistant_service.WgaConfigAssistant
	KnowledgeList []*assistant_service.WgaConfigKnowledge
	OntologyList  []*assistant_service.WgaConfigOntologyKnowledge
	// 用于构建系统消息
	McpItems       []*response.GeneralAgentResourceSelectItem
	WorkflowItems  []*response.GeneralAgentResourceSelectItem
	SkillItems     []*response.GeneralAgentResourceSelectItem
	AssistantItems []*response.GeneralAgentResourceSelectItem
	KnowledgeItems []*response.GeneralAgentResourceSelectItem
	OntologyItems  []*response.GeneralAgentResourceSelectItem
}

// hasResources 检查是否有任何资源
func (r *wgaMentionResources) hasResources() bool {
	return len(r.McpItems) > 0 ||
		len(r.WorkflowItems) > 0 ||
		len(r.SkillItems) > 0 ||
		len(r.AssistantItems) > 0 ||
		len(r.KnowledgeItems) > 0 ||
		len(r.OntologyItems) > 0
}

// fetchWgaMentionResources 获取@提及的资源列表，一次性获取所有资源后按 mentionNames 模糊匹配过滤
func fetchWgaMentionResources(ctx *gin.Context, userID, orgID string, mentionNames []string) *wgaMentionResources {
	result := &wgaMentionResources{}

	if len(mentionNames) == 0 {
		return result
	}

	// 获取全部资源（name="" 表示获取全部）
	allResources, err := GetGeneralAgentResourceSelect(ctx, userID, orgID, "")
	if err != nil {
		log.Warnf("[wga] get all resources failed: %v", err)
		return result
	}

	// 构建 mentionNames 小写列表用于模糊匹配
	mentionList := make([]string, 0, len(mentionNames))
	for _, name := range mentionNames {
		if name != "" {
			mentionList = append(mentionList, strings.ToLower(name))
		}
	}

	// 各资源类型的处理函数
	handlers := map[string]func(*response.GeneralAgentResourceSelectItem){
		"mcp": func(item *response.GeneralAgentResourceSelectItem) {
			result.McpList = append(result.McpList, &assistant_service.WgaConfigMcp{
				McpId: item.ID, McpType: item.Type,
			})
			result.McpItems = append(result.McpItems, item)
		},
		"workflow": func(item *response.GeneralAgentResourceSelectItem) {
			result.WorkflowList = append(result.WorkflowList, &assistant_service.WgaConfigWorkflow{
				WorkflowId: item.ID,
			})
			result.WorkflowItems = append(result.WorkflowItems, item)
		},
		"skill": func(item *response.GeneralAgentResourceSelectItem) {
			result.SkillList = append(result.SkillList, &assistant_service.WgaConfigSkill{
				SkillId: item.ID, SkillType: item.Type,
			})
			result.SkillItems = append(result.SkillItems, item)
		},
		"assistant": func(item *response.GeneralAgentResourceSelectItem) {
			result.AssistantList = append(result.AssistantList, &assistant_service.WgaConfigAssistant{
				AssistantId: item.ID, AssistantType: util.Int2Str(constant.AgentCategorySingle),
			})
			result.AssistantItems = append(result.AssistantItems, item)
		},
		"knowledge": func(item *response.GeneralAgentResourceSelectItem) {
			result.KnowledgeList = append(result.KnowledgeList, &assistant_service.WgaConfigKnowledge{
				KnowledgeId: item.ID,
			})
			result.KnowledgeItems = append(result.KnowledgeItems, item)
		},
		"ontology": func(item *response.GeneralAgentResourceSelectItem) {
			result.OntologyList = append(result.OntologyList, &assistant_service.WgaConfigOntologyKnowledge{
				OntologyKnowledgeId: item.ID,
			})
			result.OntologyItems = append(result.OntologyItems, item)
		},
	}

	// 各类型去重 map
	seen := make(map[string]map[string]bool)
	for _, t := range []string{"mcp", "workflow", "skill", "assistant", "knowledge", "ontology"} {
		seen[t] = make(map[string]bool)
	}

	// 遍历匹配并去重
	for _, group := range allResources {
		handler, ok := handlers[group.ListType]
		if !ok {
			continue
		}
		for _, item := range group.List {
			if matchMentionName(item.Name, mentionList) && !seen[group.ListType][item.ID] {
				seen[group.ListType][item.ID] = true
				handler(item)
			}
		}
	}

	return result
}

// matchMentionName 检查资源名称是否匹配任一 mentionName（模糊匹配，忽略大小写）
// mentionList 是预先转换为小写的 mentionNames 列表
func matchMentionName(resourceName string, mentionList []string) bool {
	if resourceName == "" || len(mentionList) == 0 {
		return false
	}

	// 资源名称转小写（只转一次）
	resourceNameLower := strings.ToLower(resourceName)

	// 检查是否有任一 mentionName 被包含在资源名称中
	for _, mentionName := range mentionList {
		if strings.Contains(resourceNameLower, mentionName) {
			return true
		}
	}
	return false
}
