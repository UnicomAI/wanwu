package response

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	agent_util "github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

const (
	agentSuccessCode = 0
	agentFailCode    = 1
	finish           = 1
	notFinish        = 0
	// 两种实现方案1：堆积所有总结到内存，最后统一正则识别url（空间换时间），2.每个流式事件校验一下，正则识别url（时间换空间）
	//目前采用第二种实现,这个值越大内存消耗越多，cpu消耗越少，反之亦然
	defaultContentQueueSize = 1000
)

type AgentInfo struct {
	Id     string //id
	Name   string //名称
	Avatar string //头像
	Order  int
}

func CreateAgentInfo(name, avatar string) *AgentInfo {
	return &AgentInfo{Id: uuid.New().String(), Name: name, Avatar: avatar}
}

type AgentChatRespContext struct {
	Order             int                //消息的order，每切换一次智能体，order+1
	MultiAgent        bool               //多智能体
	MultiAgentContext *MultiAgentContext //多智能体上下文
	AgentToolContext  *AgentToolContext  //智能体工具上下文
	SkillChatContext  *SkillChatContext  //智能体skill上下文
	ThinkChatContext  *ThinkChatContext  //智能体思考上下文
	DownloadContext   *DownloadContext   //智能体文件下载上下文

	ReplaceContent     strings.Builder // 替换内容，如果出现相同内则则进行替换
	ReplaceContentStr  string          // 替换内容，如果出现相同内则则进行替换
	ReplaceContentDone bool            //替换内容准备完成

	ContentOutput bool //上个事件是否是输出内容
}

func (c *AgentChatRespContext) ResetTool() {
	c.AgentToolContext.Reset()
	c.ReplaceContent = strings.Builder{}
	c.ReplaceContentStr = ""
	c.ReplaceContentDone = false
}

func (c *AgentChatRespContext) IncreaseOrder() {
	c.Order = c.Order + 1
	c.DownloadContext.ClearContent()
}

func NewAgentChatRespContext(multiAgent bool, order int, reqParams *request.AgentChatParams) *AgentChatRespContext {
	var mainAgentName = reqParams.AgentBaseParams.Name
	return &AgentChatRespContext{
		MultiAgent:        multiAgent,
		Order:             order,
		AgentToolContext:  NewToolContext(),
		MultiAgentContext: NewMultiAgentContext(mainAgentName),
		SkillChatContext:  NewSkillChatContext(),
		ThinkChatContext:  NewThinkChatContext(),
		DownloadContext:   NewDownloadContext(),
	}
}

type AgentChatResp struct {
	Code          int             `json:"code"`
	Message       string          `json:"message"`
	Response      string          `json:"response"`
	DetailId      string          `json:"detailId"`
	Order         int             `json:"order"` //顺序
	EventType     AgentEventType  `json:"eventType"`
	EventData     *SubEventData   `json:"eventData"`
	ResponseFiles []*AgentFile    `json:"responseFiles"`
	Finish        int             `json:"finish"`
	Usage         *AgentChatUsage `json:"usage"`
	SearchList    []interface{}   `json:"search_list"`
}

type AgentChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func BuildAgentChatResp(req *request.AgentChatContext, chatMessage *schema.Message, contentList []string, subAgentEventData *SubEventData, notStop bool, respContext *AgentChatRespContext) ([]string, error) {
	var outputList = make([]string, 0)
	if len(contentList) == 0 && subAgentEventData != nil {
		return buildSubAgentEventInfo(req, chatMessage, subAgentEventData, respContext.Order)
	}
	for _, content := range contentList {
		var agentChatResp = AgentChatSuccessResp(req, chatMessage, subAgentEventData, content, notStop, respContext)
		respString, err := buildRespString(agentChatResp)
		if err != nil {
			return nil, err
		}
		outputList = append(outputList, respString)
	}
	return outputList, nil
}

func AgentChatSuccessResp(req *request.AgentChatContext, chatMessage *schema.Message, subAgentEventData *SubEventData, content string, notStop bool, respContext *AgentChatRespContext) *AgentChatResp {
	agentFinish := buildFinish(chatMessage, notStop)
	return &AgentChatResp{
		Code:          agentSuccessCode,
		Message:       "success",
		DetailId:      req.AgentChatReq.DetailId,
		Response:      content,
		EventType:     buildEventType(subAgentEventData),
		EventData:     subAgentEventData,
		SearchList:    buildSearchList(req),
		Finish:        agentFinish,
		Usage:         buildUsage(chatMessage),
		Order:         respContext.Order,
		ResponseFiles: respContext.DownloadContext.ResponseFiles(agentFinish),
	}
}
func AgentChatFailResp(detailId string, err error, respContext *AgentChatRespContext) string {
	resp, errMsg := buildErrMsg(err)
	var agentChatResp = &AgentChatResp{
		Code:     agentFailCode,
		Message:  errMsg,
		Response: resp,
		DetailId: detailId,
		Finish:   finish,
		Order:    respContext.Order,
	}
	respString, err := buildRespString(agentChatResp)
	if err != nil {
		log.Errorf("buildRespString error: %v", err)
		return ""
	}
	return respString
}

func buildErrMsg(err error) (string, string) {
	errMsg := err.Error()
	if strings.Contains(errMsg, "[direct]") { // 判断是否为直接返回错误
		return strings.ReplaceAll(errMsg, "[direct]", ""), ""
	}
	if strings.Contains(errMsg, "chat completions err") { // 判断是否为墨子模型错误
		return "模型调用异常，请检查后重试", errMsg
	}
	return "智能体处理异常，请稍后重试", errMsg
}

func buildRespString(agentChatResp *AgentChatResp) (string, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false) // 关键：禁用 HTML 转义

	if err := encoder.Encode(agentChatResp); err != nil {
		return "", err
	}
	return "data:" + buf.String(), nil
}

func buildFinish(chatMessage *schema.Message, notStop bool) int {
	if notStop {
		return notFinish
	}
	if agent_util.StopMessage(chatMessage) {
		return finish
	}
	if chatMessage.Role == schema.Tool && chatMessage.ToolName == "exit" {
		return finish
	}
	return notFinish
}

func buildUsage(chatMessage *schema.Message) *AgentChatUsage {
	if chatMessage.ResponseMeta != nil && chatMessage.ResponseMeta.Usage != nil {
		usage := chatMessage.ResponseMeta.Usage
		return &AgentChatUsage{
			PromptTokens:     usage.PromptTokens,
			CompletionTokens: usage.CompletionTokens,
			TotalTokens:      usage.TotalTokens,
		}
	}
	return &AgentChatUsage{}
}

func buildSubAgentSearchList(subAgentEventData *SubEventData, req *request.AgentChatContext) []interface{} {
	searchList := buildSearchList(req) //处理单智能体知识库数据
	if len(searchList) > 0 {
		return searchList
	}
	if subAgentEventData == nil || req == nil || len(req.SubAgentMap) == 0 {
		return nil
	}
	config := req.SubAgentMap[subAgentEventData.Name]
	if config == nil || config.AgentChatContext == nil {
		return nil
	}
	return buildSearchList(config.AgentChatContext)
}

func buildSearchList(req *request.AgentChatContext) []interface{} {
	if req.KnowledgeHitData == nil || !req.KnowledgeHitData.AutoCitation {
		return []interface{}{}
	}
	list := req.KnowledgeHitData.SearchList
	var retList = make([]interface{}, 0)
	if len(list) > 0 {
		for _, item := range list {
			retList = append(retList, item)
		}
	}
	return retList
}
