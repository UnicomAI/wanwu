package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	_dbRag = 10
	// ragConvSensitiveKey 安全护栏回复的暂存前缀。bff 命中时写、rag-service 落库时读，
	// 两侧都要 InitRag 才连得上同一个库
	ragConvSensitiveKey = "rag_conversation_sensitive:"
	// ragSensitiveTTL 只需覆盖"bff 写入 → rag-service 落库"这一小段，与智能体同口径
	ragSensitiveTTL = 1 * time.Minute
)

var (
	_redisRag *client
)

func InitRag(ctx context.Context, cfg Config) error {
	if _redisRag != nil {
		return fmt.Errorf("redis rag client already init")
	}
	c, err := newClient(ctx, cfg, _dbRag)
	if err != nil {
		return err
	}
	_redisRag = c
	return nil
}

func StopRag() {
	if _redisRag != nil {
		_redisRag.Stop()
		_redisRag = nil
	}
}

// Rag 返回rag的redis客户端，初始化失败时为nil
func Rag() *client {
	return _redisRag
}

// RagSensitiveReply 护栏命中时用户在页面上实际看到的内容。
type RagSensitiveReply struct {
	Response  string `json:"response"`  // 已转发的正文 + 护栏回复
	Reasoning string `json:"reasoning"` // 已转发的思考内容
}

// StoreRagSensitiveConversation 暂存本轮问答的护栏结果。
func StoreRagSensitiveConversation(conversationId, detailId string, reply *RagSensitiveReply) {
	defer util.PrintPanicStack()
	if Rag() == nil || conversationId == "" || detailId == "" || reply == nil {
		return
	}
	data, err := json.Marshal(reply)
	if err != nil {
		log.Errorf("[RAG] 序列化安全护栏回复失败，conversationId: %v, err: %v", conversationId, err)
		return
	}
	if _, err := Rag().SetEx(context.Background(), buildRagSensConvKey(conversationId, detailId), string(data), ragSensitiveTTL); err != nil {
		log.Errorf("[RAG] 暂存安全护栏回复失败，conversationId: %v, detailId: %v, err: %v", conversationId, detailId, err)
	}
}

// GetRagSensitiveConversation 取本轮问答的护栏结果，这一轮没命中护栏则返回 nil
func GetRagSensitiveConversation(conversationId, detailId string) *RagSensitiveReply {
	defer util.PrintPanicStack()
	if Rag() == nil || conversationId == "" || detailId == "" {
		return nil
	}
	data, err := Rag().Get(context.Background(), buildRagSensConvKey(conversationId, detailId))
	if err != nil || data == "" {
		return nil
	}
	var reply RagSensitiveReply
	if err := json.Unmarshal([]byte(data), &reply); err != nil {
		log.Errorf("[RAG] 解析安全护栏回复失败，conversationId: %v, err: %v", conversationId, err)
		return nil
	}
	return &reply
}

func buildRagSensConvKey(conversationId, detailId string) string {
	return ragConvSensitiveKey + conversationId + "_" + detailId
}
