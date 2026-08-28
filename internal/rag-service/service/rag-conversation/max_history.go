package rag_conversation

import (
	"encoding/json"

	rag_service "github.com/UnicomAI/wanwu/api/proto/rag-service"
	"github.com/UnicomAI/wanwu/internal/rag-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// BuildMaxHistory 取知识库/问答库两处「最长上下文」配置的较大值，与前端裁剪口径一致。
// 返回 0 表示不带历史
func BuildMaxHistory(rag *model.RagInfo) int {
	if rag == nil {
		return 0
	}
	maxHistory := int(rag.KnowledgeBaseConfig.MaxHistory)
	if len(rag.QAKnowledgebaseConfig) == 0 {
		return maxHistory
	}
	qaConfig := &rag_service.RagQAKnowledgeBaseConfig{}
	if err := json.Unmarshal([]byte(rag.QAKnowledgebaseConfig), qaConfig); err != nil {
		// 解析失败不影响问答，退化为只看知识库侧配置
		log.Errorf("解析知识问答问答库配置失败，ragId: %s, error: %v", rag.RagID, err)
		return maxHistory
	}
	if qaMaxHistory := int(qaConfig.GetGlobalConfig().GetMaxHistory()); qaMaxHistory > maxHistory {
		return qaMaxHistory
	}
	return maxHistory
}
