package agent_chat_builder

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/UnicomAI/wanwu/internal/agent-service/model"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	"github.com/UnicomAI/wanwu/pkg/log"
	utils "github.com/UnicomAI/wanwu/pkg/util"
	"github.com/cloudwego/eino/schema"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	toolParamsStartFormat = "\n\n```工具参数：\n"
	toolParamsEndFormat   = "\n```\n\n"
	toolEndFormat         = "\n\n<<<\n工具%s调用结果：\n %s \n>>>\n\n"
	toolEndJsonFormat     = "\n\n```工具%s调用结果：\n %s \n```\n\n"

	unknownFileSize = -1 // 未知文件大小

	contentLimitSize = 20 * 1024
	contentLimitText = "\n... (因篇幅过长，后续内容已省略)"
	noCitationText   = "已搜索知识库"
)

type ToolMessageContent struct {
	Content      []string
	SubEventData *response.SubEventData
}

func (t ToolMessageContent) Empty() bool {
	return len(t.Content) == 0 && t.SubEventData == nil
}

type SingleAgentMessageBuilder struct {
}

func NewSingleBuilder() *SingleAgentMessageBuilder {
	return &SingleAgentMessageBuilder{}
}

func (*SingleAgentMessageBuilder) MessageType() MessageType {
	return SingleAgentMessage
}

func (*SingleAgentMessageBuilder) FilterMessage(respContext *response.AgentChatRespContext, chatMessage *schema.Message) bool {
	filter := filterMessage(respContext, chatMessage)
	if filter {
		return true
	}
	if !respContext.ContentOutput {
		messageTool := response.CreateMessageTool(chatMessage, respContext)
		//过滤一些只包含/n的内容
		if len(chatMessage.ReasoningContent) == 0 && !messageTool.ToolMessage() && !util.StopMessage(chatMessage) {
			//本身大于0 trim之后=0
			if len(chatMessage.Content) > 0 && len(strings.TrimSpace(strings.Trim(chatMessage.Content, "\n"))) == 0 {
				return true
			}
		}
	}
	return false
}

func (*SingleAgentMessageBuilder) BuildContent(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message) ([]*response.AgentMessageContent, error) {
	////适配不返回任何工具，但是工具参数结束的消息
	//messageTool := response.CreateMessageTool(chatMessage, respContext)
	//if messageTool.ToolParamsEnd() && !respContext.AgentToolContext.HasTool() && !respContext.MultiAgentContext.AgentChangeStart {
	//	chatMessage.ResponseMeta.FinishReason = "stop"
	//}
	return buildSingleAgentContent(req, respContext, chatMessage)
}

// buildSingleAgentContent
// 最极限情况，多智能体->单智能体->skill->多智能体->单智能体
func buildSingleAgentContent(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message) ([]*response.AgentMessageContent, error) {
	var retContentList []*response.AgentMessageContent
	//构造思考内容
	thinkMessage := respContext.ThinkChatContext.ThinkMessage(chatMessage, respContext)
	if len(thinkMessage) > 0 {
		retContentList = append(retContentList, thinkMessage...)
		if len(chatMessage.ReasoningContent) != 0 { //如果只有思考内容，则直接返回，说明是思考过程
			return retContentList, nil
		}
	}

	//技能消息
	skillMessage, err, directReturn := buildSkillMessage(req, respContext, chatMessage)
	if err != nil {
		return nil, err
	}
	if len(skillMessage) > 0 {
		retContentList = append(retContentList, skillMessage...)
	}
	//直接返回
	if directReturn {
		return retContentList, nil
	}

	content := buildCommonSingleAgentContent(req, respContext, chatMessage)
	if len(content) > 0 {
		retContentList = append(retContentList, content...)
	}
	return retContentList, nil
}

func buildSkillMessage(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message) ([]*response.AgentMessageContent, error, bool) {
	skillBuilder := NewSkillBuilder()
	if !skillBuilder.FilterMessage(respContext, chatMessage) { //是skill消息
		content, err := skillBuilder.BuildContent(req, respContext, chatMessage)
		return content, err, true

	}
	return nil, nil, false
}

func buildCommonSingleAgentContent(req *request.AgentChatContext, respContext *response.AgentChatRespContext, chatMessage *schema.Message) []*response.AgentMessageContent {
	stepsMap, toolIdList := buildToolStep(chatMessage, respContext)
	if len(stepsMap) == 0 { //没有工具处理
		if !respContext.ContentOutput {
			respContext.ContentOutput = true
			respContext.IncreaseOrder()
			respContext.ReplaceContent.Reset()
		}
		// 正文内容：缓存到队列，用于提取 markdown 下载文件
		if len(chatMessage.Content) > 0 {
			respContext.DownloadContext.AddContent(chatMessage.Content)
		}
		return buildNoToolContent(chatMessage, respContext)
	}
	return buildToolContentNewStyle(req, chatMessage, respContext, stepsMap, toolIdList)
}

/*
*
目前工具调用有几种情况做处理
1.正常流式：先输出方法名，在流式分别输出方法对应的参数，再输出调用结果
2.并发流式：如果需要调用同一方法两次，先输出方法名，方法参数，再输出方法名方法参数，再输出结果1，再输出结果2
3.同步请求：请求一个事件，返回一个事件，没有流式
4.同步请求和返回：请求和返回都在同一个事件，没有流式
*/
func buildToolStep(chatMessage *schema.Message, respContext *response.AgentChatRespContext) (map[string][]response.ToolStep, []string) {
	messageTool := response.CreateMessageTool(chatMessage, respContext)
	var toolStepMap = make(map[string][]response.ToolStep)
	//构造toolId
	var toolId = messageTool.ToolId()

	var toolIdList []string
	if messageTool.ToolStart() {
		for _, tool := range chatMessage.ToolCalls {
			newTool := messageTool.NewTool(tool)
			if newTool { //新工具开始
				toolId = tool.ID
			}
			steps := toolStepMap[toolId]
			if len(tool.Function.Name) > 0 {
				steps = append(steps, response.ToolNameStep)
				if newTool {
					steps = append(steps, response.ToolParamStartStep)
				}
			}

			if len(tool.Function.Arguments) > 0 {
				steps = append(steps, response.ToolParamStep)
			}
			if messageTool.ToolParamsEnd() {
				steps = append(steps, response.ToolParamFinishStep)
			}
			if messageTool.ToolEnd() {
				steps = append(steps, response.ToolResultFinishStep)
			}
			toolStepMap[toolId] = steps
			toolIdList = append(toolIdList, toolId)
		}
	} else if messageTool.ToolParamsEnd() {
		steps := toolStepMap[toolId]
		steps = append(steps, response.ToolParamFinishStep)
		toolStepMap[toolId] = steps
		toolIdList = append(toolIdList, toolId)
	} else if messageTool.ToolEnd() {
		steps := toolStepMap[toolId]
		steps = append(steps, response.ToolResultFinishStep)
		toolStepMap[toolId] = steps
		toolIdList = append(toolIdList, toolId)
	}
	return toolStepMap, toolIdList
}

// buildNoToolContent 构造没有工具的内容
// case1：tool 有数据同时content内容；如果此时在工具的输出中还没有输出完，则不输出content的相关内容
// case2：在tool输出前会输出规划内容，但是会重复输出相同的规划内容，所以当内容数字大于10时，同时出现重复数据，则不输出
// case3：正式输出
func buildNoToolContent(chatMessage *schema.Message, respContext *response.AgentChatRespContext) []*response.AgentMessageContent {
	notFinishList := response.FilerToolByStep(respContext, response.ToolResultFinishStep, false)
	if len(notFinishList) > 0 { //在工具期间，不输出任何content内容
		return []*response.AgentMessageContent{}
	}
	toolContext := respContext.AgentToolContext
	//替换内容准备(工具未开始，但是输出了内容, 有的模型会重复输出一样的话)
	if !toolContext.HasTool() {
		var content = chatMessage.Content
		if len(content) == 0 {
			content = chatMessage.ReasoningContent
		}
		if utf8.RuneCountInString(content) > 10 {
			var replaceContent = respContext.ReplaceContentStr
			if len(replaceContent) == 0 {
				replaceContent = respContext.ReplaceContent.String()
			}
			if replaceContent == content {
				respContext.ReplaceContentDone = true
				respContext.ReplaceContentStr = replaceContent
				return []*response.AgentMessageContent{}
			}
		}
		if !respContext.ReplaceContentDone {
			respContext.ReplaceContent.WriteString(content)
		}
	}
	return buildContent(chatMessage)
}

func buildContent(chatMessage *schema.Message) []*response.AgentMessageContent {
	var retContentList []*response.AgentMessageContent
	//构造正常内容
	if len(chatMessage.Content) > 0 || util.StopMessage(chatMessage) {
		retContentList = append(retContentList, &response.AgentMessageContent{
			ContentList: []string{chatMessage.Content},
		})
	}
	return retContentList
}

// buildToolContentNewStyle 构造有工具的内容输出-新样式
// 需要额外判断，如果此次输出的步骤不包含当前任务的步骤，同时之前工具有参数未完成的，则补充个参数结束的内容（处理并发调用工具的情况）
func buildToolContentNewStyle(req *request.AgentChatContext, chatMessage *schema.Message, respContext *response.AgentChatRespContext, stepsMap map[string][]response.ToolStep, toolIdList []string) []*response.AgentMessageContent {
	toolContext := respContext.AgentToolContext
	steps := stepsMap[toolContext.GetCurrentToolId()]
	paramsNotFinishList := response.FilerToolByStep(respContext, response.ToolParamStep, true)
	var toolContentList []*response.AgentMessageContent
	if len(steps) == 0 && len(paramsNotFinishList) > 0 { //是新工具且之前工具处于参数处理未完成状态
		//增加参数处理完成结果，并更改状态
		for _, toolId := range paramsNotFinishList {
			tool := respContext.AgentToolContext.GetTool(toolId)
			if tool == nil {
				continue
			}
			//更改状态
			tool.ToolStep = response.ToolParamFinishStep
			toolContentList = append(toolContentList, &response.AgentMessageContent{
				SubEventData: response.BuildEndTool(tool),
				ContentList:  []string{toolParamsEndFormat},
			})
		}
	}
	//根据step循环构造输出的内容
	for _, toolId := range toolIdList {
		toolSteps := stepsMap[toolId]
		agentTool := respContext.AgentToolContext.GetTool(toolId)
		if agentTool == nil {
			agentTool = &response.AgentTool{ToolId: toolId, Order: respContext.Order, StartTime: time.Now().UnixMilli(), ToolIndex: response.BuildToolIndex(chatMessage)}
			respContext.AgentToolContext.AddTool(agentTool)
		}
		for _, step := range toolSteps {
			agentTool.ToolStep = step
			toolContent := buildNewContentByStep(respContext, req, agentTool, chatMessage, step, toolId)
			if toolContent.Empty() {
				continue
			}
			toolContentList = append(toolContentList, toolContent)
		}
	}
	return toolContentList
}

// buildNewContentByStep 根据当前步骤构造需要输出的内容
func buildNewContentByStep(respContext *response.AgentChatRespContext, req *request.AgentChatContext, agentTool *response.AgentTool, chatMessage *schema.Message, step response.ToolStep, toolId string) *response.AgentMessageContent {
	var subEventData *response.SubEventData
	var contentList []string
	if agentTool.ToolType == response.KnowledgeEventType {
		return buildKnowledgeContentByStep(req, agentTool, chatMessage, step)
	}
	switch step {
	case response.ToolNameStep:
		tool := buildMessageTool(chatMessage, toolId)
		if tool == nil {
			break
		}
		respContext.ContentOutput = false
		respContext.IncreaseOrder()
		agentTool.ToolName = buildToolName(tool.Function.Name, req.ToolMap)
		agentTool.ToolType = response.BuildEventTypeByTool(agentTool)
		agentTool.Avatar = buildToolAvatar(tool.Function.Name, req.ToolMap, agentTool.ToolType)
		subEventData = response.BuildStartTool(agentTool)
		if agentTool.ToolType == response.KnowledgeEventType {
			subEventData.ParentId = req.AgentId()
		}
	case response.ToolParamStartStep:
		contentList = append(contentList, toolParamsStartFormat)
		subEventData = response.BuildProcessTool(agentTool)
	case response.ToolParamStep:
		tool := buildMessageTool(chatMessage, toolId)
		if tool == nil {
			break
		}
		contentList = append(contentList, tool.Function.Arguments)
		subEventData = response.BuildProcessTool(agentTool)
	case response.ToolParamFinishStep:
		contentList = append(contentList, toolParamsEndFormat)
		subEventData = response.BuildProcessTool(agentTool)
	case response.ToolResultFinishStep:
		subEventData = response.BuildEndTool(agentTool)
		if len(chatMessage.Content) > 0 {
			toolResult := processToolResult(respContext, subEventData, chatMessage.Content, toolId)
			contentList = append(contentList, toolResult)
		}
	}
	return &response.AgentMessageContent{
		ContentList:  contentList,
		SubEventData: subEventData,
	}
}

func buildKnowledgeContentByStep(req *request.AgentChatContext, agentTool *response.AgentTool, chatMessage *schema.Message, step response.ToolStep) *response.AgentMessageContent {
	var subEventData *response.SubEventData
	var contentList []string
	switch step {
	case response.ToolNameStep, response.ToolParamStartStep, response.ToolParamStep, response.ToolParamFinishStep:
		break
	case response.ToolResultFinishStep:
		knowledgeContent := buildKnowledgeContent(chatMessage.Content)
		req.KnowledgeHitData = knowledgeContent
		if knowledgeContent != nil && !knowledgeContent.AutoCitation {
			contentList = append(contentList, noCitationText)
		}
		subEventData = response.BuildEndTool(agentTool)
		subEventData.ParentId = req.AgentId()
	}
	return &response.AgentMessageContent{
		ContentList:  contentList,
		SubEventData: subEventData,
	}
}

// buildKnowledgeContent 构造知识内容数据
func buildKnowledgeContent(data string) *model.KnowledgeHitData {
	if len(data) == 0 {
		return nil
	}
	var knowledgeHitData = &model.KnowledgeHitData{}
	err := json.Unmarshal([]byte(data), knowledgeHitData)
	if err != nil {
		return nil
	}
	return knowledgeHitData
}

// buildMessageTool 构造消息工具内容数据
func buildMessageTool(chatMessage *schema.Message, toolId string) *schema.ToolCall {
	switch length := len(chatMessage.ToolCalls); length {
	case 0:
		return nil
	case 1:
		return &chatMessage.ToolCalls[0]
	}

	for _, call := range chatMessage.ToolCalls {
		if call.ID == toolId {
			return &call
		}
	}
	return nil
}

// buildToolName 构建工具名称, 从md5转换成正常的工具名称
func buildToolName(toolName string, toolMap map[string]*request.ToolConfig) string {
	if len(toolMap) == 0 {
		return toolName
	}
	tool := toolMap[toolName]
	if tool == nil {
		return toolName
	}
	return tool.ToolName
}

// buildToolAvatar 构建工具头像
func buildToolAvatar(toolName string, toolMap map[string]*request.ToolConfig, toolEventType int) string {
	if len(toolMap) == 0 {
		return response.BuildDefaultAvatarByType(toolEventType)
	}
	toolConfig := toolMap[toolName]
	if toolConfig == nil || toolConfig.Avatar == "" {
		return response.BuildDefaultAvatarByType(toolEventType)
	}
	return toolConfig.Avatar
}

// processToolResult 处理工具调用结果消息
func processToolResult(respContext *response.AgentChatRespContext, subEventData *response.SubEventData, content string, toolId string) string {
	var toolResult string
	if json.Valid([]byte(content)) {
		// 尝试从JSON中提取文件URL和文件名
		fileList := extractFilesFromJSON(content)
		if len(fileList) > 0 {
			respContext.DownloadContext.AddDownloadFile(toolId, fileList)
		}
		toolResult = fmt.Sprintf(toolEndJsonFormat, "", limitContent(content))
		// 尝试判断是否是错误信息
		errResp := extraErrMessage(content)
		if len(errResp) > 0 {
			subEventData.ErrMessage = errResp
			subEventData.Status = response.EventFailStatus
		}
	} else {
		toolResult = fmt.Sprintf(toolEndFormat, "", limitContent(content))
	}
	return toolResult
}

// limitContent 限制content输出大小
func limitContent(content string) string {
	if len(content) <= contentLimitSize {
		return content
	}
	return content[:contentLimitSize] + contentLimitText
}

// extraErrMessage 获取错误信息
func extraErrMessage(content string) (result string) {
	defer utils.PrintPanicStackWithCall(func(panicOccur bool, recoverError error) {
		if panicOccur {
			result = ""
			return
		}
	})
	if len(content) == 0 {
		return ""
	}
	var mcpResult = &mcp.CallToolResult{}
	err := json.Unmarshal([]byte(content), mcpResult)
	if err == nil {
		if len(mcpResult.Content) > 0 {
			textContent, ok := mcpResult.Content[0].(mcp.TextContent)
			if ok && textContent.Text != "" && textContent.Type == "text" {
				content = textContent.Text
			}
		}
	}
	var errResult = response.AgentToolErrResp{}
	err = json.Unmarshal([]byte(content), &errResult)
	if err != nil {
		return ""
	}
	return errResult.ErrorMsg
}

// extractFilesFromJSON 从JSON内容中提取文件URL和文件名
// 支持多种JSON格式：
// 1. 单个文件: {"output": "http://xxx/file.txt"} 或 {"url": "http://xxx/file.txt"}
// 2. 文件对象: {"file_url": "http://xxx", "file_name": "name.txt"} 或类似字段
// 3. 文件数组: {"files": [...]} 或直接是数组
// 4. 嵌套对象中包含文件URL
func extractFilesFromJSON(content string) []*response.DownloadFileInfo {
	var fileList []*response.DownloadFileInfo

	var data interface{}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return fileList
	}

	// 递归提取文件,最多2层
	extractFilesFromValue(data, 2, &fileList)

	return fileList
}

// extractFilesFromValue 递归从值中提取文件信息
func extractFilesFromValue(data interface{}, depth int, fileList *[]*response.DownloadFileInfo) {
	if depth <= 0 {
		return
	}
	switch v := data.(type) {
	case map[string]interface{}:
		// 尝试从当前对象提取文件
		if info := extractFileInfoFromMap(v); info != nil {
			*fileList = append(*fileList, info)
		}
		// 递归处理所有值
		for _, val := range v {
			depth = depth - 1
			extractFilesFromValue(val, depth, fileList)
		}
	case []interface{}:
		depth = depth - 1
		for _, item := range v {
			extractFilesFromValue(item, depth, fileList)
		}
	}
}

// extractFileInfoFromMap 从map中提取文件信息
func extractFileInfoFromMap(m map[string]interface{}) *response.DownloadFileInfo {
	// 常见的URL字段名
	urlKeys := []string{"output", "url", "file_url", "fileUrl", "file_url_path", "downloadUrl", "download_url", "path", "link"}
	// 常见的文件名字段名
	nameKeys := []string{"file_name", "fileName", "name", "filename", "file", "title"}

	var fileURL, fileName string

	// 提取URL
	for _, key := range urlKeys {
		if val, ok := m[key]; ok {
			if str, ok := val.(string); ok && util.IsValidFileURL(str) {
				fileURL = str
				break
			}
		}
	}

	// 如果没找到URL，尝试查找看起来像URL的字符串值
	if fileURL == "" {
		for _, val := range m {
			if str, ok := val.(string); ok && util.IsValidFileURL(str) && !isLikelyNameField(str) {
				fileURL = str
				break
			}
		}
	}

	// 提取文件名
	for _, key := range nameKeys {
		if val, ok := m[key]; ok {
			if str, ok := val.(string); ok && str != "" {
				log.Infof("nameKeys key %s, val %s", key, str)
				fileName = str
				break
			}
		}
	}

	// 如果找到了URL
	if fileURL != "" {
		// 如果没有文件名，从URL中提取
		if fileName == "" {
			fileName = util.ExtractFileNameFromURL(fileURL)
		}
		return &response.DownloadFileInfo{
			FileName: fileName,
			FilePath: fileURL,
			FileSize: unknownFileSize, // 文件大小未知
			CreateAt: time.Now().Format("2006-01-02 15:04:05"),
		}
	}

	return nil
}

// isLikelyNameField 检查字符串是否可能是名称字段值（而非URL）
func isLikelyNameField(s string) bool {
	// 简单判断：如果字符串很短且不包含常见URL特征
	return len(s) < 10 && !strings.Contains(s, "/") && !strings.Contains(s, ".")
}
