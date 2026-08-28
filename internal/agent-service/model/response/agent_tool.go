package response

import (
	"encoding/json"

	"github.com/cloudwego/eino/schema"
)

type ToolStep int

const (
	ToolNameStep         ToolStep = 0 //输出工具名阶段
	ToolParamStartStep   ToolStep = 1 //输出工具参数 开始阶段
	ToolParamStep        ToolStep = 2 //输出工具参数阶段
	ToolParamFinishStep  ToolStep = 3 //输出工具参数完成阶段
	ToolResultFinishStep ToolStep = 4 //输出工具结果完成阶段
)

type AgentToolErrResp struct {
	ErrorMsg string `json:"errorMsg"`
}

func ToolErrResp(err error) (string, error) {
	if err == nil {
		return "", nil
	}
	resp := AgentToolErrResp{ErrorMsg: err.Error()}
	marshal, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(marshal), nil
}

type AgentTool struct {
	ToolId    string
	ToolIndex *int
	ToolName  string
	ToolType  int
	Avatar    string
	ToolStep  ToolStep //工具阶段
	Order     int      //工具顺序
	StartTime int64
}

type AgentToolContext struct {
	ToolMap              map[string]*AgentTool
	ToolList             []*AgentTool
	DefaultCurrentToolId string // 默认当前工具Id,手动设置
}

func BuildToolIndex(chatMessage *schema.Message) *int {
	if chatMessage != nil && len(chatMessage.ToolCalls) > 0 {
		return chatMessage.ToolCalls[0].Index
	}
	return nil
}

// FilerToolByStep ,equalCondition为true 则过滤等于此类型的tool，为false 则过滤不等于此类型的tool
func FilerToolByStep(respContext *AgentChatRespContext, step ToolStep, equalCondition bool) []string {
	if len(respContext.AgentToolContext.ToolMap) > 0 {
		var toolIdList []string
		for toolId, tool := range respContext.AgentToolContext.ToolMap {
			if FilterToolByCondition(tool, step, equalCondition) {
				toolIdList = append(toolIdList, toolId)
			}
		}
		return toolIdList
	}
	return nil
}

func FilterToolByCondition(tool *AgentTool, step ToolStep, equalCondition bool) bool {
	if equalCondition {
		return tool.ToolStep == step
	} else {
		return tool.ToolStep != step
	}
}

func NewToolContext() *AgentToolContext {
	return &AgentToolContext{
		ToolMap:  make(map[string]*AgentTool),
		ToolList: make([]*AgentTool, 0),
	}
}

func (a *AgentToolContext) HasTool() bool {
	return len(a.ToolMap) > 0
}

func (a *AgentToolContext) ExistTool(toolId string) bool {
	return a.GetTool(toolId) != nil
}

func (a *AgentToolContext) GetTool(toolId string) *AgentTool {
	return a.ToolMap[toolId]
}

func (a *AgentToolContext) GetCurrentToolId() string {
	if len(a.ToolList) == 0 {
		return a.DefaultCurrentToolId
	}
	return a.ToolList[len(a.ToolList)-1].ToolId
}

func (a *AgentToolContext) GetLastTool() *AgentTool {
	if len(a.ToolList) == 0 {
		return nil
	}
	return a.ToolList[len(a.ToolList)-1]
}

func (a *AgentToolContext) AddToolById(toolId string) *AgentTool {
	tool := &AgentTool{ToolId: toolId, Order: len(a.ToolList)}
	a.AddTool(tool)
	return tool
}

func (a *AgentToolContext) AddTool(tool *AgentTool) {
	a.ToolMap[tool.ToolId] = tool
	a.ToolList = append(a.ToolList, tool)
}

func (a *AgentToolContext) DefaultToolId(toolId string) {
	a.DefaultCurrentToolId = toolId
}

func (a *AgentToolContext) Reset() {
	a.ToolList = []*AgentTool{}
	a.ToolMap = make(map[string]*AgentTool)
	a.DefaultCurrentToolId = ""
}
