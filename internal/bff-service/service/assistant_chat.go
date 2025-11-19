package service

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/pkg/ahocorasick"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/log"
	sse_util "github.com/UnicomAI/wanwu/pkg/sse-util"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func AssistantConversionStream(ctx *gin.Context, userId, orgId string, req request.ConversionStreamRequest) error {
	// 1. CallAssistantConversationStream
	chatCh, err := CallAssistantConversationStream(ctx, userId, orgId, req)
	if err != nil {
		return err
	}
	// 2. 流式返回结果 [EN] 2. Streaming return results
	_ = sse_util.NewSSEWriter(ctx, fmt.Sprintf("[Agent] %v conversation %v user %v org %v recv", req.AssistantId, req.ConversationId, userId, orgId), sse_util.DONE_MSG).
		WriteStream(chatCh, nil, buildAgentChatRespLineProcessor(), nil)
	return nil
}

func CallAssistantConversationStream(ctx *gin.Context, userId, orgId string, req request.ConversionStreamRequest) (<-chan string, error) {
	// 根据agentID获取敏感词配置 [EN] Get sensitive word configuration based on agentID
	agentInfo, err := assistant.GetAssistantInfo(ctx, &assistant_service.GetAssistantInfoReq{
		AssistantId: req.AssistantId,
	})
	if err != nil {
		return nil, err
	}

	var matchDicts []ahocorasick.DictConfig
	// 如果Enable为true,则处理敏感词 [EN] If Enable is true, handle sensitive words
	if agentInfo.SafetyConfig.GetEnable() {
		var ids []string
		for _, idx := range agentInfo.SafetyConfig.GetSensitiveTable() {
			ids = append(ids, idx.TableId)
		}
		matchDicts, err = BuildSensitiveDict(ctx, ids)
		if err != nil {
			return nil, err
		}
		matchResults, err := ahocorasick.ContentMatch(req.Prompt, matchDicts, true)
		if err != nil {
			return nil, err
		}
		if len(matchResults) > 0 {
			if matchResults[0].Reply != "" {
				return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req", matchResults[0].Reply)
			}
			return nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFSensitiveWordCheck, "bff_sensitive_check_req_default_reply")
		}
	}
	stream, err := assistant.AssistantConversionStream(ctx.Request.Context(), &assistant_service.AssistantConversionStreamReq{
		AssistantId:    req.AssistantId,
		ConversationId: req.ConversationId,
		FileInfo:       transFileInfo(req.FileInfo),
		Trial:          req.Trial,
		Prompt:         req.Prompt,
		Identity: &assistant_service.Identity{
			UserId: userId,
			OrgId:  orgId,
		},
	})
	if err != nil {
		return nil, err
	}

	rawCh := make(chan string, 128)
	go func() {
		defer util.PrintPanicStack()
		defer close(rawCh)
		log.Infof("[Agent] %v conversation %v user %v org %v start, query: %s", req.AssistantId, req.ConversationId, userId, orgId, req.Prompt)
		for {
			s, err := stream.Recv()
			if err == io.EOF {
				log.Infof("[Agent] %v conversation %v user %v org %v stop", req.AssistantId, req.ConversationId, userId, orgId)
				break
			}
			if err != nil {
				log.Errorf("[Agent] %v conversation %v user %v org %v recv err: %v", req.AssistantId, req.ConversationId, userId, orgId, err)
				break
			}
			rawCh <- s.Content
		}
	}()
	if !agentInfo.SafetyConfig.GetEnable() {
		return rawCh, nil
	}
	// 敏感词过滤 [EN] Sensitive word filtering
	outputCh := ProcessSensitiveWords(ctx, rawCh, matchDicts, &agentSensitiveService{})
	return outputCh, nil
}

// transFileInfo 转换文件信息从请求模型到protobuf模型 [EN] transFileInfo converts file information from the request model to the protobuf model
func transFileInfo(fileInfo []request.ConversionStreamFile) []*assistant_service.ConversionStreamFile {
	if len(fileInfo) == 0 {
		return nil
	}
	result := make([]*assistant_service.ConversionStreamFile, 0, len(fileInfo))
	for _, file := range fileInfo {
		result = append(result, &assistant_service.ConversionStreamFile{
			FileName: file.FileName,
			FileSize: file.FileSize,
			FileUrl:  file.FileUrl,
		})
	}
	return result
}

// buildAgentChatRespLineProcessor 构造agent对话结果行处理器 [EN] buildAgentChatRespLineProcessor constructs the agent conversation result line processor
func buildAgentChatRespLineProcessor() func(*gin.Context, string, interface{}) (string, bool, error) {
	return func(c *gin.Context, lineText string, params interface{}) (string, bool, error) {
		if strings.HasPrefix(lineText, "error:") {
			errorText := fmt.Sprintf("data: {\"code\": -1, \"message\": \"%s\"}\n\n", strings.TrimPrefix(lineText, "error:"))
			return errorText, false, nil
		}
		if strings.HasPrefix(lineText, "data:") {
			return lineText + "\n\n", false, nil
		}
		return "data:" + lineText + "\n\n", false, nil
	}
}

// --- agent sensitive ---

type agentSensitiveService struct{}

func (s *agentSensitiveService) serviceType() string {
	return constant.AppTypeAgent
}

// parseContent implements ChatService.
func (s *agentSensitiveService) parseContent(raw string) (id, content string) {
	raw = strings.TrimPrefix(raw, "data:")
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", ""
	}
	resp := struct {
		MsgID    string `json:"msg_id"`
		Response string `json:"response"`
	}{}
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return "", ""
	}
	return resp.MsgID, resp.Response
}

// buildSensitiveResp implements ChatService.
func (s *agentSensitiveService) buildSensitiveResp(id string, content string) []string {
	resp := map[string]interface{}{
		"code":              0,
		"message":           "success",
		"response":          content,
		"gen_file_url_list": []interface{}{},
		"history":           []interface{}{},
		"finish":            1, // Note: The original JSON has "finish" misspelled as "finish"
		"usage": map[string]interface{}{
			"prompt_tokens":     0,
			"completion_tokens": 0,
			"total_tokens":      0,
		},
		"search_list": []interface{}{},
	}
	marshal, _ := json.Marshal(resp)
	return []string{"data: " + string(marshal)}
}
