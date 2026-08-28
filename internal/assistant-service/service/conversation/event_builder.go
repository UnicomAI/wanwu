package conversation

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
)

const (
	errorDefaultOrder = 999
)

type SubEventStatus int

const (
	MainAgentEventType    = 0 //单智能体事件/多智能体主智能体
	SubAgentEventType     = 1 //子智能体事件
	KnowledgeEventType    = 2 //知识库事件
	ToolEventType         = 3 //工具事件
	SkillEventType        = 4 //技能事件
	SubAgentToolEventType = 5 //子智能体工具事件
	ThinkingEventType     = 6 //智能体思考事件

	SubTextEventType   = 20 //技能内容事件
	terminationMessage = "本次回答已被终止"
)

var builderMap = make(map[int]EventBuilder)

type ConversationResp struct {
	Order                int
	EventOrder           int
	EventType            int
	EventData            *model.SubEventData
	FullResponseList     []*model.ConversationResponse
	FullResponse         *strings.Builder
	SearchList           *string
	ConversationEventMap map[string]*ConversationResp
	ResponseFiles        []*model.AgentFile
	CurrentData          string
	SensitiveMessage     string
	Error                error
	Statistic            *model.AgentStatistic
}

func CreateConversationResp() *ConversationResp {
	return &ConversationResp{FullResponse: &strings.Builder{}, ConversationEventMap: make(map[string]*ConversationResp),
		Order: -1, Statistic: &model.AgentStatistic{
			StartTime: time.Now().UnixMilli(),
		}}
}

func (cr *ConversationResp) Write(data string, order int) {
	if order != cr.EventOrder {
		resp := &model.ConversationResponse{Response: cr.FullResponse.String(), Order: cr.EventOrder}
		cr.EventOrder = order
		if !resp.Empty() {
			cr.FullResponseList = append(cr.FullResponseList, resp)
		}
		cr.FullResponse.Reset()
	}
	cr.FullResponse.WriteString(data)
}

func (cr *ConversationResp) WriteError(data string, errMessage string, order int) {
	if cr.FullResponse.Len() > 0 {
		resp := &model.ConversationResponse{Response: cr.FullResponse.String(), Order: cr.EventOrder}
		cr.EventOrder = order
		cr.FullResponseList = append(cr.FullResponseList, resp)
		cr.FullResponse.Reset()
		order += 1
	}
	resp := &model.ConversationResponse{ErrMessage: errMessage, ErrResponse: data, Order: order}
	cr.EventOrder = order
	cr.FullResponseList = append(cr.FullResponseList, resp)
	cr.FullResponse.Reset()
}

func (cr *ConversationResp) References() string {
	var searchList string
	if cr.SearchList != nil {
		searchList = *cr.SearchList
	}
	return searchList
}

func (cr *ConversationResp) Response() string {
	var conversationResponse = cr.FullResponse.String()
	if cr.Error != nil {
		//这里面不直接使用stringBuilder 原因是防止Response 被多次调用导致多次生成err
		if len(conversationResponse) > 0 {
			conversationResponse += "\n"
		}
		conversationResponse += terminationMessage
	}
	return conversationResponse
}

func (cr *ConversationResp) SensitiveResponse(sensitiveResponse string) {
	if len(sensitiveResponse) > 0 {
		cr.SensitiveMessage = sensitiveResponse
	}
}

func (cr *ConversationResp) ErrorResponse(err error) {
	cr.Error = err
}
func (cr *ConversationResp) ResponseList() []*model.ConversationResponse {
	var conversationResponse = cr.FullResponse.String()
	if cr.Error != nil {
		//这里面不直接使用stringBuilder 原因是防止Response 被多次调用导致多次生成err
		if len(conversationResponse) > 0 {
			conversationResponse += "\n"
		}
		conversationResponse += terminationMessage
	}
	list := cr.FullResponseList
	var retList []*model.ConversationResponse
	if len(list) > 0 {
		for _, data := range list {
			if data.Order < 0 {
				continue
			}
			retList = append(retList, data)
		}
	}
	if len(conversationResponse) > 0 {
		retList = append(retList, &model.ConversationResponse{Response: conversationResponse, Order: cr.EventOrder})
	}

	return retList
}

type AgentChatResp struct {
	Code          int                 `json:"code"`
	Message       string              `json:"message"`
	Order         int                 `json:"order"`
	Response      string              `json:"response"`
	SearchList    []interface{}       `json:"search_list"`
	Finish        int                 `json:"finish"`
	EventType     int                 `json:"eventType"`
	EventData     *model.SubEventData `json:"eventData"`
	ResponseFiles []*model.AgentFile  `json:"responseFiles"`
}

type EventBuilder interface {
	EventType() int
	Build(conversationResp *ConversationResp, conversation, searchResult string, agentChatResp *AgentChatResp) error
}

func InitBuilder(eventBuilder EventBuilder) {
	builderMap[eventBuilder.EventType()] = eventBuilder
}

func BuildConversationResp(conversationResp *ConversationResp, strLine string) error {
	if strLine == "\n" {
		return nil
	}
	var lastData = conversationResp.CurrentData
	conversationResp.CurrentData = strLine
	if len(lastData) == 0 {
		return nil
	}
	// 记录首token到达时间（首次处理到有效数据时）
	conversationResp.Statistic.SetFirstTokenLatency()
	return buildConversationResp(conversationResp, lastData)
}

func FinishConversationResp(conversationResp *ConversationResp) error {
	if len(conversationResp.CurrentData) > 0 {
		err := buildConversationResp(conversationResp, conversationResp.CurrentData)
		return err
	}
	if conversationResp.Error != nil {
		conversationResp.WriteError("智能体处理异常，请稍后重试", conversationResp.Error.Error(), errorDefaultOrder)
	}
	return nil
}

func buildConversationResp(conversationResp *ConversationResp, strLine string) error {
	conversation, searchResult, agentChatResp := processAgentResp(strLine)
	if agentChatResp == nil {
		return nil
	}
	//下载文件值
	conversationResp.ResponseFiles = agentChatResp.ResponseFiles
	builder := builderMap[agentChatResp.EventType]
	if builder == nil {
		return fmt.Errorf("no builder found event type %d", agentChatResp.EventType)
	}
	return builder.Build(conversationResp, conversation, searchResult, agentChatResp)
}

func processAgentResp(strLine string) (string, string, *AgentChatResp) {
	if len(strLine) >= 5 && strLine[:5] == "data:" {
		jsonStrData := strLine[5:]
		// 解析流式数据，提取response字段和search_list
		var agentChatResp = &AgentChatResp{}
		if err1 := json.Unmarshal([]byte(jsonStrData), agentChatResp); err1 == nil {
			var searchList string
			if len(agentChatResp.SearchList) > 0 {
				marshal, err := json.Marshal(agentChatResp.SearchList)
				if err == nil {
					searchList = string(marshal)
				}
			}
			return agentChatResp.Response, searchList, agentChatResp
		}
	}
	return "", "", nil
}
