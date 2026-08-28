package nodes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/request"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	agent_util "github.com/UnicomAI/wanwu/internal/agent-service/pkg/util"
	"github.com/UnicomAI/wanwu/internal/agent-service/service/agent-message-flow/prompt"
	minio_service "github.com/UnicomAI/wanwu/internal/agent-service/service/minio-service"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/cloudwego/eino/schema"
)

const (
	placeholderOfUserInput   = "_user_input"
	placeholderOfChatHistory = "_chat_history"
)

type PromptVariables struct {
	Avs map[string]string
}

func (p *PromptVariables) AssemblePromptVariables(ctx context.Context, reqContext *request.AgentChatContext) (variables map[string]any, err error) {
	req := reqContext.AgentChatReq
	variables = make(map[string]any)

	variables[prompt.PlaceholderOfAgentSystemPrompt] = req.AgentBaseParams.Instruction
	variables[prompt.PlaceholderOfTime] = time.Now().Format("Monday 2006/01/02 15:04:05 -07")
	variables[prompt.PlaceholderOfAgentName] = req.AgentBaseParams.Name

	input, err := buildUserInput(reqContext)
	if err != nil {
		return nil, err
	}
	variables[placeholderOfUserInput] = input

	// Handling conversation history
	if len(req.ModelParams.History) > 0 {
		// Add chat history to variable
		variables[placeholderOfChatHistory] = buildHistory(req.ModelParams.History, req.ModelParams.MaxHistory)
	}

	if p.Avs != nil {
		var memoryVariablesList []string
		for k, v := range p.Avs {
			variables[k] = v
			memoryVariablesList = append(memoryVariablesList, fmt.Sprintf("%s: %s\n", k, v))
		}
		variables[prompt.PlaceholderOfVariables] = memoryVariablesList
	}

	subAgentInfoList := reqContext.AgentChatReq.SubAgentInfoList
	if reqContext.AgentChatReq.MultiAgent && len(subAgentInfoList) > 0 {
		variables[prompt.PlaceholderOfSubAgentCount] = strconv.Itoa(len(subAgentInfoList))
	}

	return variables, nil
}

func buildHistory(history []request.AssistantConversionHistory, maxHistory int) []*schema.Message {
	var historyList []*schema.Message

	// 处理所有历史记录
	for _, conversionHistory := range history {
		query := buildUrlInput(conversionHistory.Query, conversionHistory.UploadFileUrl)
		historyList = append(historyList, schema.UserMessage(query))
		if len(conversionHistory.Response) == 0 {
			continue
		}
		//todo 先不传ToolCall(后续版本考虑传进去)
		historyList = append(historyList, schema.AssistantMessage(conversionHistory.Response, nil))
	}
	if maxHistory <= 0 {
		return historyList
	}
	// 每条记录占用2个位置(问/答)
	maxHistory = maxHistory * 2
	// 只返回最后maxHistory条
	if len(historyList) > maxHistory {
		return historyList[len(historyList)-maxHistory:]
	}
	return historyList
}

func buildUserInput(reqContext *request.AgentChatContext) ([]*schema.Message, error) {
	req := reqContext.AgentChatReq
	agentChatInfo := reqContext.AgentChatInfo
	var input = req.Input

	var messages []*schema.Message
	if agentChatInfo.VisionSupport && agentChatInfo.ImageUpload {
		imageUrl := extraImageUrl(req.UploadFile)
		// 视觉模型，传了图片文件
		var parts []schema.MessageInputPart
		if len(imageUrl) > 0 {
			for _, minioFilePath := range imageUrl {
				message, err := buildFileMessage(minioFilePath)
				if err != nil {
					return nil, err
				}
				parts = append(parts, *message)
			}
		}
		input = buildVisionSupportUrlInput(input, req.UploadFile)
		parts = append(parts, schema.MessageInputPart{
			Type: schema.ChatMessagePartTypeText,
			Text: input,
		})
		messages = append(messages, &schema.Message{
			Role:                  schema.User,
			UserInputMultiContent: parts,
		})
	} else if agentChatInfo.UploadUrl { //非视觉模型，传了url
		input += buildUrlInput(input, req.UploadFile)
		messages = append(messages, schema.UserMessage(input))
	} else {
		messages = append(messages, schema.UserMessage(input))
	}
	return messages, nil
}

func extraImageUrl(fileUrlList []string) (imageUrl []string) {
	if len(fileUrlList) == 0 {
		return nil
	}
	for _, fileUrl := range fileUrlList {
		fileName := agent_util.ExtractFileNameFromURL(fileUrl)
		if agent_util.ImageFile(fileName) {
			imageUrl = append(imageUrl, fileUrl)
		}
	}
	return imageUrl
}

func buildUrlInput(query string, fileUrl []string) string {
	count, urlStr := buildMultiUrlString(fileUrl)
	if count == 0 {
		return query
	}
	if count == 1 {
		return query + "\n用户上传的文档连接为:" + urlStr
	}
	return fmt.Sprintf("%s\n用户上传%d个文档连接为:%s", query, count, urlStr)
}

func buildVisionSupportUrlInput(query string, fileUrl []string) string {
	count, urlStr := buildMultiUrlString(fileUrl)
	if count == 0 {
		return query
	}
	if count == 1 {
		return query + "\n用户上传的文档连接为（无需读取，供工具调用使用）:" + urlStr
	}
	return fmt.Sprintf("%s\n用户上传%d个文档连接为（无需读取，供工具调用使用）:%s", query, count, urlStr)
}

func buildMultiUrlString(fileUrlList []string) (int, string) {
	if len(fileUrlList) == 0 {
		return 0, ""
	}
	fileUrlList = rebuildUrlList(fileUrlList)
	if len(fileUrlList) == 1 {
		return 1, fileUrlList[0]
	}
	marshal, err := json.Marshal(fileUrlList)
	if err != nil {
		return 1, fileUrlList[0]
	}
	return len(fileUrlList), string(marshal)
}

func rebuildUrlList(fileUrl []string) []string {
	var retList []string
	for _, file := range fileUrl {
		retList = append(retList, rebuildUlr(file))
	}
	return retList
}

func rebuildUlr(fileUrl string) (retUrl string) {
	defer func() {
		if err := recover(); err != nil {
			retUrl = fileUrl
			return
		}
	}()
	if strings.Contains(fileUrl, "minio-wanwu") {
		openUrl := config.GetConfig().Minio.DownloadUrl
		if openUrl != "" {
			path, err := replaceURLDomainAndMergePath(fileUrl, openUrl)
			if err == nil {
				return path
			}
		}
	}
	return fileUrl
}

// replaceURLDomainAndMergePath 替换域名，并将新 URL 的路径与原始 URL 的路径拼接。
// originalURL: 原始完整 URL
// newBaseURL:  新基础 URL（从中提取 Scheme、Host 和 Path，Path 会与原始 Path 拼接）
// 返回替换后的 URL 字符串，或错误
func replaceURLDomainAndMergePath(originalURL, newBaseURL string) (string, error) {
	// 解析原始 URL
	orig, err := url.Parse(originalURL)
	if err != nil {
		return "", fmt.Errorf("解析原始 URL 失败: %w", err)
	}

	// 解析新基础 URL
	newBase, err := url.Parse(newBaseURL)
	if err != nil {
		return "", fmt.Errorf("解析新基础 URL 失败: %w", err)
	}

	if newBase.Host == "" {
		return "", errors.New("新基础 URL 必须包含协议和主机（例如 http://example.com）")
	}

	var newSchema = newBase.Scheme
	if newSchema == "" {
		newSchema = orig.Scheme
	}

	// 拼接路径：新路径 + 原始路径
	// 注意处理斜杠，避免双斜杠或缺失斜杠
	newPath := strings.TrimRight(newBase.Path, "/")
	origPath := strings.TrimLeft(orig.Path, "/")
	mergedPath := newPath + "/" + origPath
	// 如果原始路径为空，则直接使用新路径（去除末尾多余斜杠）
	if orig.Path == "" {
		mergedPath = newPath
	}
	// 如果新路径为空，则直接使用原始路径
	if newBase.Path == "" {
		mergedPath = orig.Path
	}

	// 构造新 URL
	newURL := &url.URL{
		Scheme:   newSchema,
		Host:     newBase.Host,
		Path:     mergedPath,
		RawPath:  "", // 让 Go 自动编码；如果需要保留原始编码可自行处理，一般不需要
		RawQuery: orig.RawQuery,
		Fragment: orig.Fragment,
	}

	return newURL.String(), nil
}

// buildFileMessage 构建文件消息
func buildFileMessage(minioFilePath string) (*schema.MessageInputPart, error) {
	//1.下载压缩文件到本地
	var localFilePath = agent_util.BuildFilePath(config.GetConfig().AgentFileConfig.LocalFilePath, util.ExtractExtension(minioFilePath))
	err := minio_service.DownloadFileToLocal(context.Background(), minioFilePath, localFilePath)
	if err != nil {
		return nil, err
	}
	//2.图片转base64
	mimeType, base64, err := agent_util.Img2base64Data(localFilePath)
	if err != nil {
		return nil, err
	}
	return &schema.MessageInputPart{
		Type: schema.ChatMessagePartTypeImageURL,
		Image: &schema.MessageInputImage{
			MessagePartCommon: schema.MessagePartCommon{
				Base64Data: &base64,
				MIMEType:   mimeType,
			},
		},
	}, nil
}
