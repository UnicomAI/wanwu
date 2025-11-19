package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type GetWorkflowTemplateListResp struct {
	Total        int64                   `json:"total"`
	List         []*WorkflowTemplateInfo `json:"list"`
	DownloadLink WorkflowTemplateURL     `json:"downloadLink"`
}

// WorkflowTemplateDetail 工作流模板详情响应 [EN] WorkflowTemplateDetail workflow template details response
type WorkflowTemplateDetail struct {
	WorkflowTemplateInfo
	Summary  string `json:"summary"`  // 模板介绍概览 [EN] Template introduction overview
	Feature  string `json:"feature"`  // 模板特性说明 [EN] Template feature description
	Scenario string `json:"scenario"` // 模板应用场景 [EN] Template application scenarios
	Note     string `json:"note"`     // 注意事项 [EN] Things to note
}

// WorkflowTemplateListItem 工作流模板列表项 [EN] WorkflowTemplateListItem workflow template list item
type WorkflowTemplateInfo struct {
	TemplateId    string         `json:"templateId"`    // 模板ID [EN] Template ID
	Avatar        request.Avatar `json:"avatar"`        // 图标 [EN] icon
	Name          string         `json:"name"`          // 模板名称 [EN] Template name
	Desc          string         `json:"desc"`          // 模板描述 [EN] Template description
	Category      string         `json:"category"`      // 模板分类 [EN] Template classification
	Author        string         `json:"author"`        // 作者 [EN] author
	DownloadCount int32          `json:"downloadCount"` // 下载次数 [EN] Number of downloads
}

type WorkflowTemplateURL struct {
	Url string `json:"url"`
}
