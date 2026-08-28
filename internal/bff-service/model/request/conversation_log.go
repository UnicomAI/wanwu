package request

type GetConversationLogListRequest struct {
	ConversationLogQuery
	PageSearch
	CommonCheck
}

type ConversationLogQuery struct {
	AppId     string   `json:"appId" form:"appId" validate:"required"`         //业务id 智能体id、知识问答id
	AppType   string   `json:"appType" form:"appType" validate:"required"`     //业务类型 智能体、知识问答
	Name      string   `json:"name" form:"name"`                               //会话标题
	Source    []string `json:"source" form:"source"`                           //来源 web、openapi、webURL、draft
	StartDate string   `json:"startDate" form:"startDate" validate:"required"` //起始日期
	EndDate   string   `json:"endDate" form:"endDate" validate:"required"`     //截至日期
	UserIds   []string `json:"userIds" form:"userIds"`                         //用户id列表
	OrderBy   string   `json:"orderBy" form:"orderBy"`                         //排序字段 avgCosts avgFirstTokenLatency createAt updateAt
	OrderType string   `json:"orderType" form:"orderType"`                     //排序类型 asc des
}

type GetConversationLogDetailRequest struct {
	AppId          string `json:"appId" form:"appId"`                   //业务id 智能体id、知识问答id
	AppType        string `json:"appType" form:"appType"`               //业务类型 智能体、知识问答
	ConversationId string `json:"conversationId" form:"conversationId"` //会话id
	PageSearch
	CommonCheck
}

type GetConversationLogUserSelectRequest struct {
	AppId   string `json:"appId" form:"appId"`     //业务id 智能体id、知识问答id
	AppType string `json:"appType" form:"appType"` //业务类型 智能体、知识问答
	CommonCheck
}

// ConversationLogExportReq 对话日志导出请求
type ConversationLogExportReq struct {
	AppId   string   `json:"appId" form:"appId" validate:"required"`     // 应用id
	AppType string   `json:"appType" form:"appType" validate:"required"` // 应用类型
	LogIds  []string `json:"logIds" form:"logIds"`                       // 对话日志id列表
	CommonCheck
}

// ConversationLogExportRecordListReq 对话日志导出记录列表请求
type ConversationLogExportRecordListReq struct {
	ConversationLogQuery
	PageSearch
	CommonCheck
}

// DeleteConversationLogExportRecordReq 删除对话日志导出记录请求
type DeleteConversationLogExportRecordReq struct {
	ExportRecordIds []string `json:"exportRecordIds" validate:"required"` // 导出记录id列表
	CommonCheck
}
