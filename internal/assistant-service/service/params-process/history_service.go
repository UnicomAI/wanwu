package params_process

import (
	"encoding/json"
	"strings"

	assistant_service "github.com/UnicomAI/wanwu/api/proto/assistant-service"
	"github.com/UnicomAI/wanwu/internal/assistant-service/client/model"
	"github.com/UnicomAI/wanwu/internal/assistant-service/config"
	"github.com/UnicomAI/wanwu/pkg/es"
	"github.com/UnicomAI/wanwu/pkg/log"
)

const (
	esHistoryIndexPattern = "conversation_detail_infos_*"
)

type HistoryProcess struct {
}

func init() {
	AddServiceContainer(&HistoryProcess{})
}

func (k *HistoryProcess) ServiceType() ServiceType {
	return ConversionHistoryType
}

func (k *HistoryProcess) Prepare(agent *AgentInfo, prepareParams *AgentPrepareParams, clientInfo *ClientInfo, userQueryParams *UserQueryParams) error {
	if userQueryParams == nil || len(userQueryParams.ConversationId) == 0 {
		return nil
	}
	maxHistory := BuildMaxHistory(agent.Assistant)
	if maxHistory == 0 {
		return nil
	}

	fieldConditions := map[string]interface{}{
		"conversationId": userQueryParams.ConversationId,
		"userId.keyword": userQueryParams.QueryUserId,
		"orgId.keyword":  userQueryParams.QueryOrgId,
	}

	documents, _, err := es.Assistant().SearchByFields(userQueryParams.Ctx, esHistoryIndexPattern, fieldConditions, 0, maxHistory, "desc")
	if err != nil {
		log.Warnf("Assistant服务查询历史聊天记录失败，conversationId: %s, userId: %s, error: %v", userQueryParams.ConversationId, userQueryParams.QueryUserId, err)
		return err
	}
	//转换顺序
	var conversationList []*model.ConversationDetails
	for i := len(documents) - 1; i >= 0; i-- {
		doc := documents[i]
		var detail model.ConversationDetails
		if err := json.Unmarshal(doc, &detail); err != nil {
			log.Warnf("Assistant服务解析ES历史聊天记录失败: %v", err)
			continue
		}
		if !checkDetail(&detail) {
			log.Infof("conversation detail fail response %s", doc)
			continue
		}
		conversationList = append(conversationList, &detail)
	}
	prepareParams.ConversionDetailList = conversationList
	return nil
}

func (k *HistoryProcess) Build(assistant *AgentInfo, prepareParams *AgentPrepareParams, agentChatParams *assistant_service.AgentDetail) error {
	var historyList []*assistant_service.ConversionHistory
	if len(prepareParams.ConversionDetailList) > 0 {
		for _, detail := range prepareParams.ConversionDetailList {
			historyList = append(historyList, &assistant_service.ConversionHistory{
				Query:         detail.Prompt,
				UploadFileUrl: extractFileUrlsFromModel(detail.FileInfo),
				Response:      buildConversationResp(detail.Response, detail.ResponseList),
			})
		}
	}
	agentChatParams.ModelParams.History = historyList
	return nil
}

func checkDetail(detail *model.ConversationDetails) bool {
	respLen := len(detail.ResponseList)
	if respLen > 0 {
		resp := detail.ResponseList[respLen-1]
		if len(resp.ErrResponse) > 0 {
			return false
		}
	}
	return true
}

func buildConversationResp(response string, respList []*model.ConversationResponse) string {
	if len(respList) == 0 {
		return response
	}
	var retBuilder = strings.Builder{}
	for _, resp := range respList {
		retBuilder.WriteString(resp.Response)
	}
	return retBuilder.String()
}

func BuildMaxHistory(agent *model.Assistant) int {
	var maxHistory = config.DefaultMaxHistoryLength
	memoryConfigStr := agent.MemoryConfig
	if len(memoryConfigStr) > 0 {
		memoryConfig := &assistant_service.AssistantMemoryConfig{}
		err := json.Unmarshal([]byte(memoryConfigStr), memoryConfig)
		if err != nil {
			//失败不影响智能体整个流程
			log.Errorf("Assistant服务解析智能体记忆配置失败，assistantId: %d, error: %v, memoryConfigRaw: %s", agent.ID, err, agent.MemoryConfig)
		} else {
			maxHistory = int(memoryConfig.MaxHistoryLength)
		}
	}
	return maxHistory
}

// extractFileUrlsFromModel 从model FileInfo中提取所有文件URL
func extractFileUrlsFromModel(fileInfos []model.FileInfo) []string {
	if len(fileInfos) == 0 {
		return nil
	}
	var fileUrls []string
	for _, file := range fileInfos {
		if file.FileUrl != "" {
			fileUrls = append(fileUrls, file.FileUrl)
		}
	}
	return fileUrls
}
