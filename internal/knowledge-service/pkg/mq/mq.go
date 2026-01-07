package mq

import (
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
)

func SendMessage(msg interface{}, topic string) error {
	cfg := config.GetConfig()
	if cfg.Kafka.Enabled {
		return sendMessageToKafka(msg, topic)
	} else if cfg.Redis.Enabled {
		return sendMessageToRedis(msg, topic)
	}
	return nil
}
