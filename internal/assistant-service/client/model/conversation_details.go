package model

import (
	"time"
)

type ConversationType string
type SubEventStatus int

const (
	AgentTool         ConversationType = "agentTool"         //主智能体工具
	AgentKnowledge    ConversationType = "agentKnowledge"    //主智能体知识库
	AgentThink        ConversationType = "agentThink"        //主智能体思考
	SubAgent          ConversationType = "subAgent"          //子智能体
	AgentSkill        ConversationType = "agentSkill"        //子智能体
	AgentSubText      ConversationType = "subText"           //智能体skill内容,或者子智能体内容
	SubAgentTool      ConversationType = "subAgentTool"      //子智能体工具
	SubAgentKnowledge ConversationType = "subAgentKnowledge" //子智能体只是库

	EventStartStatus   SubEventStatus = 1 //开始事件
	EventProcessStatus SubEventStatus = 2 //输出中
	EventEndStatus     SubEventStatus = 3 //结束事件
	EventFailStatus    SubEventStatus = 4 //子智能体失败

	FeedBackLike    int32 = 1 //点赞类型
	FeedBackDislike int32 = 2 //点踩类型
)

type FileInfo struct {
	FileName string `json:"fileName"`
	FileSize int64  `json:"fileSize"`
	FileUrl  string `json:"fileUrl"`
}

type AgentFile struct {
	Name     string     `json:"name"`
	Size     int        `json:"size"`
	FileUrl  string     `json:"fileUrl"`
	FileType string     `json:"fileType"`
	Metadata *AgentMeta `json:"metadata"`
}

type AgentMeta struct {
	Desc     string `json:"desc"`
	CreateAt string `json:"createAt"`
	Name     string `json:"name"`
}

type AgentStatistic struct {
	// 统计字段
	StartTime         int64  `json:"startTime"`         // 请求开始时间戳(毫秒)
	FirstTokenLatency int64  `json:"firstTokenLatency"` // 首token延时(毫秒)
	TotalCostTime     int64  `json:"totalCostTime"`     //总消耗时间
	ErrMessage        string `json:"errMessage"`        //是否有报错
	SourceFrom        string `json:"sourceFrom"`        //来源类型：web/openapi/webUrl/draft
	TraceId           string `json:"traceId"`           // 追踪ID
}

func (a *AgentStatistic) SetFirstTokenLatency() {
	if a.FirstTokenLatency == 0 {
		span := time.Now().UnixMilli() - a.StartTime
		if span <= 0 { //处理兜底情况
			span = 1
		}
		a.FirstTokenLatency = span
	}
}

func (a *AgentStatistic) SetTotalCostTime() {
	a.TotalCostTime = time.Now().UnixMilli() - a.StartTime
}

func (a *AgentStatistic) SetErr(err error) {
	if err != nil {
		a.ErrMessage = err.Error()
	}
}

func (a *AgentStatistic) SetTraceId(traceId string) {
	a.TraceId = traceId
}

func (a *AgentStatistic) SetSourceFrom(sourceFrom string) {
	a.SourceFrom = sourceFrom
}

type SubEventData struct {
	Status      SubEventStatus `json:"status"`
	Id          string         `json:"id"`
	ParentId    string         `json:"parentId"`
	Name        string         `json:"name"`
	Profile     string         `json:"profile"`
	TimeCost    string         `json:"timeCost"`
	Order       int            `json:"order"`
	DisplayMode int            `json:"displayMode"`
	ErrMessage  string         `json:"errMessage"`
}

func (s *SubEventData) Copy() *SubEventData {
	return &SubEventData{
		Status:     s.Status,
		Id:         s.Id,
		ParentId:   s.ParentId,
		Name:       s.Name,
		Profile:    s.Profile,
		TimeCost:   s.TimeCost,
		Order:      s.Order,
		ErrMessage: s.ErrMessage,
	}
}

type SubConversationDetail struct {
	BusinessId                string                   `json:"businessId"` //业务id，当conversationType 是AgentTool,SubAgentTool 则是toolId，SubAgent 则是agentId
	ConversationType          ConversationType         `json:"conversationType"`
	Content                   string                   `json:"content"`                   //内容
	Order                     int                      `json:"order"`                     //全局顺序
	SubConversationDetailList []*SubConversationDetail `json:"subConversationDetailList"` //子数据内容，对于多智能体，每个智能体又有多个工具详情，使用此处
	SearchList                string                   `json:"searchList"`
	EventData                 *SubEventData            `json:"eventData"`
}

type ConversationResponse struct {
	Response    string `json:"response"`
	ErrResponse string `json:"errResponse"`
	ErrMessage  string `json:"errMessage"`
	Order       int    `json:"order"`
}

func (c *ConversationResponse) Empty() bool {
	return len(c.Response) == 0 && len(c.ErrResponse) == 0 && len(c.ErrMessage) == 0
}

type ConversationDetails struct {
	Id                        string                   `json:"id"`
	AssistantId               string                   `json:"assistantId"`
	ConversationId            string                   `json:"conversationId"`
	Prompt                    string                   `json:"prompt"`
	SysPrompt                 string                   `json:"sysPrompt"`
	Response                  string                   `json:"response"`
	ResponseList              []*ConversationResponse  `json:"responseList"`
	SubConversationDetailList []*SubConversationDetail `json:"SubConversationDetailList"`
	SearchList                string                   `json:"searchList"`
	FileUrl                   string                   `json:"requestFileUrls"`
	FileSize                  int64                    `json:"fileSize"`
	FileName                  string                   `json:"fileName"`
	FileInfo                  []FileInfo               `json:"fileInfo"`
	UserId                    string                   `json:"userId"`
	OrgId                     string                   `json:"orgId"`
	CreatedAt                 int64                    `json:"createdAt"`
	UpdatedAt                 int64                    `json:"updatedAt"`
	ResponseFiles             []*AgentFile             `json:"responseFiles"`
	Statistic                 *AgentStatistic          `json:"statistic"`
	Deleted                   bool                     `json:"deleted"`         // 逻辑删除标记，true 表示已删除
	Feedback                  int32                    `json:"feedback"`        // 当前反馈状态: 0=无 1=点赞 2=点踩
	FeedbackContent           string                   `json:"feedbackContent"` // 反馈文本内容
}
