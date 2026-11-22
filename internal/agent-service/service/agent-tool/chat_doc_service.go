package agent_tool

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/internal/agent-service/model/response"
	agent_config "github.com/UnicomAI/wanwu/internal/agent-service/pkg/config"
	"github.com/UnicomAI/wanwu/internal/agent-service/pkg/http"
	tool_recovery "github.com/UnicomAI/wanwu/internal/agent-service/service/agent-tool/tool-recovery"
	service_model "github.com/UnicomAI/wanwu/internal/agent-service/service/service-model"
	tokenizer_service "github.com/UnicomAI/wanwu/internal/agent-service/service/tokenizer-service"
	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ChatDocParams struct {
	UploadFileUrl string `json:"upload_file_url"`
	MaxToken      int    `json:"max_token"`
}

// chatDocTool 实现了 tool.InvokableRun 接口
type chatDocTool struct {
	info     *schema.ToolInfo
	chatInfo *service_model.AgentChatInfo
}

// Info 返回工具的元信息
func (t *chatDocTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	marshal, _ := json.Marshal(t.info)
	log.Infof("chatDocTool %v", string(marshal))
	return t.info, nil
}

// InvokableRun 执行工具
func (t *chatDocTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	params, err := buildChatDocParams(argumentsInJSON, t.chatInfo)
	if err != nil {
		return "", err
	}
	doc, err := searchChatDoc(ctx, params)
	if err != nil {
		log.Errorf("chat doc request error,params %s, err: %v", argumentsInJSON, err)
		if tool_recovery.ShouldPropagate(ctx, err) {
			return "", err
		}
		return response.ToolErrResp(err)
	}
	return doc, nil
}

// GetChatDocTool 构建chatDoc技能工具
func GetChatDocTool(chatInfo *service_model.AgentChatInfo, hasChatDoc bool) tool.BaseTool {
	if hasChatDoc {
		toolInfo := buildChatDocToolInfo()
		if toolInfo != nil {
			return &chatDocTool{
				info:     toolInfo,
				chatInfo: chatInfo,
			}
		}
	}
	return nil
}

func buildChatDocToolInfo() *schema.ToolInfo {
	templateConfig := agent_config.GetToolTemplateConfig()
	chatDocConfig, _ := templateConfig.GetToolByID(agent_config.DocParser)
	if chatDocConfig != nil {
		apiSchema, _ := GetEnioToolsFromOpenAPISchema(context.Background(), chatDocConfig)
		if len(apiSchema) > 0 {
			info := apiSchema[0]
			return &schema.ToolInfo{
				Name:        info.Name,
				Desc:        info.Desc,
				ParamsOneOf: info.ParamsOneOf,
			}
		}
	}
	return nil
}

// buildChatDocParams 构造chatDoc请求参数
func buildChatDocParams(argumentsInJSON string, chatInfo *service_model.AgentChatInfo) (*ChatDocParams, error) {
	var chatDocParams = &ChatDocParams{}
	if err := json.Unmarshal([]byte(argumentsInJSON), chatDocParams); err != nil {
		return nil, fmt.Errorf("invalid chatdoc arguments: %w", err)
	}
	chatDocParams.MaxToken = tokenizer_service.TokenLimit(chatInfo)
	return chatDocParams, nil
}

// searchChatDoc 查询chatDoc
func searchChatDoc(ctx context.Context, chatDocParams *ChatDocParams) (string, error) {
	toolServer := agent_config.GetConfig().ToolServer
	url := toolServer.Endpoint + "/v1/doc_parse"
	marshal, err := json.Marshal(chatDocParams)
	if err != nil {
		return "", err
	}
	result, err := http.GetClient().PostJson(ctx, &http_client.HttpRequestParams{
		Url:        url,
		Body:       marshal,
		Timeout:    5 * time.Minute,
		MonitorKey: "search_chat_doc",
		LogLevel:   http_client.LogAll,
	})
	if err != nil {
		return "", err
	}
	return string(result), nil
}
