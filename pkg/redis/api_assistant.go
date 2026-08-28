package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
)

const (
	_dbAssistant     = 9
	convSensitiveKey = "assistant_conversation_sensitive:"
)

var (
	_redisAssistant *client
)

func InitAssistant(ctx context.Context, cfg Config) error {
	if _redisAssistant != nil {
		return fmt.Errorf("redis assistant client already init")
	}
	c, err := newClient(ctx, cfg, _dbAssistant)
	if err != nil {
		return err
	}
	_redisAssistant = c
	return nil
}

func StopAssistant() {
	if _redisAssistant != nil {
		_redisAssistant.Stop()
		_redisAssistant = nil
	}
}

func Assistant() *client {
	return _redisAssistant
}

func StoreSensitiveConversation(conversationId, detailID, sensitiveMessage string) {
	defer util.PrintPanicStack()
	_, err := Assistant().SetEx(context.Background(), buildSensConvKey(conversationId, detailID), sensitiveMessage, 1*time.Minute)
	if err != nil {
		log.Infof("[Assistant] store sensitive conversation %v err: %v", conversationId, err)
	}
}

func GetSensitiveConversation(conversationId, detailID string) string {
	defer util.PrintPanicStack()
	//如果集群架构需要考虑主从延迟情况
	//time.Sleep(100 * time.Millisecond)
	sensitiveMessage, err := Assistant().Get(context.Background(), buildSensConvKey(conversationId, detailID))
	if err != nil {
		log.Infof("[Assistant] store sensitive conversation %v err: %v", conversationId, err)
		sensitiveMessage = ""
	}
	log.Infof("[Assistant] get sensitive conversation %v detailID %v message: %v", conversationId, detailID, sensitiveMessage)
	return sensitiveMessage
}

func buildSensConvKey(conversationId, detailID string) string {
	return convSensitiveKey + conversationId + "_" + detailID
}
