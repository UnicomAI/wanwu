package knowledge_log

import (
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg"
	"github.com/UnicomAI/wanwu/internal/knowledge-service/pkg/config"
	"github.com/UnicomAI/wanwu/pkg/log"
)

// 打印等级 Panic > Error > Warn > Info > Debug [EN] Print level Panic > Error > Warn > Info > Debug

var knowledgeLog = KnowledgeLog{}

type KnowledgeLog struct {
}

func init() {
	pkg.AddContainer(knowledgeLog)
}

func (c KnowledgeLog) LoadType() string {
	return "knowledge-log-config"
}

func (c KnowledgeLog) Load() error {
	configInfo := config.GetConfig()
	logConfig := configInfo.Log
	return log.InitLog(logConfig.Std, logConfig.Level, logConfig.Logs...)
}

func (c KnowledgeLog) StopPriority() int {
	return pkg.DefaultPriority
}

func (c KnowledgeLog) Stop() error {
	return nil
}
