package response

import (
	"encoding/json"
	"strings"
)

type AgentEventType int
type SubEventStatus int

const (
	EventStartStatus   SubEventStatus = 1 //开始事件
	EventProcessStatus SubEventStatus = 2 //输出中
	EventEndStatus     SubEventStatus = 3 //结束事件
	EventFailStatus    SubEventStatus = 4 //子智能体失败
)

type AgentChatResp struct {
	Code          int             `json:"code"`
	Message       string          `json:"message"`
	Response      string          `json:"response"`
	DetailId      string          `json:"detailId"`
	Order         int             `json:"order"` //顺序
	EventType     int             `json:"eventType"`
	EventData     *SubEventData   `json:"eventData"`
	ResponseFiles []*AgentFile    `json:"responseFiles"`
	Finish        int             `json:"finish"`
	Usage         *AgentChatUsage `json:"usage"`
	SearchList    []interface{}   `json:"search_list"`
}

// Compact 合并智能体返回消息
func (a *AgentChatResp) Compact(newMsg *AgentChatResp) *AgentChatResp {
	//处理正文输出请情况
	if a.EventData == nil && newMsg.EventData == nil {
		a.Response = a.Response + newMsg.Response
		return a
	}
	//处理子事件输出中的情况，一个子事件，会拆成3部分，1.开始事件 2.输出中 3.结束事件
	if a.EventData != nil && newMsg.EventData != nil && a.EventData.Id == newMsg.EventData.Id &&
		a.EventData.Status == EventProcessStatus && newMsg.EventData.Status == EventProcessStatus {
		a.Response = a.Response + newMsg.Response
		return a
	}
	return nil
}

func UnmarshalAgentResp(data string) (*AgentChatResp, error) {
	data = strings.TrimPrefix(data, "data:")
	resp := AgentChatResp{}
	err := json.Unmarshal([]byte(data), &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func MarshalAgentResp(data *AgentChatResp) (string, error) {
	marshal, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return "data:" + string(marshal), nil
}

type SubEventData struct {
	Status    SubEventStatus `json:"status"`
	Id        string         `json:"id"`
	EventType int            `json:"eventType"`
	Name      string         `json:"name"`
	Profile   string         `json:"profile"`
	TimeCost  string         `json:"timeCost"`
	ParentId  string         `json:"parentId"`
	Order     int            `json:"order"`
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

type AgentChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
