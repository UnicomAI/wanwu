package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type GetWorkflowTemplateListResp struct {
	Total        int64                   `json:"total"`
	List         []*WorkflowTemplateInfo `json:"list"`
	DownloadLink WorkflowTemplateURL     `json:"downloadLink"`
}

// WorkflowTemplateDetail workflow template details response
type WorkflowTemplateDetail struct {
	WorkflowTemplateInfo
	Summary  string `json:"summary"`  // Template introduction overview
	Feature  string `json:"feature"`  // Template feature description
	Scenario string `json:"scenario"` // Template application scenarios
	Note     string `json:"note"`     // Things to note
}

// WorkflowTemplateListItem workflow template list item
type WorkflowTemplateInfo struct {
	TemplateId    string         `json:"templateId"`    // Template ID
	Avatar        request.Avatar `json:"avatar"`        // icon
	Name          string         `json:"name"`          // Template name
	Desc          string         `json:"desc"`          // Template description
	Category      string         `json:"category"`      // Template classification
	Author        string         `json:"author"`        // author
	DownloadCount int32          `json:"downloadCount"` // Number of downloads
}

type WorkflowTemplateURL struct {
	Url string `json:"url"`
}
