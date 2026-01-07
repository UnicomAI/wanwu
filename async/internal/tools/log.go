package tools

import (
	"log"

	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
)

func DefaultLog() async_config.Logger {
	return &defaultLog{}
}

func EmptyLog() async_config.Logger { return &emptyLog{} }

// --- default log ---

type defaultLog struct{}

func (l *defaultLog) Debugf(fmt string, i ...interface{}) {
	log.Printf("[ASYNC][DEBUG] "+fmt, i...)
}

func (l *defaultLog) Infof(fmt string, i ...interface{}) {
	log.Printf("[ASYNC][INFO] "+fmt, i...)
}

func (l *defaultLog) Warnf(fmt string, i ...interface{}) {
	log.Printf("[ASYNC][WARN] "+fmt, i...)
}

func (l *defaultLog) Errorf(fmt string, i ...interface{}) {
	log.Printf("[ASYNC][ERROR] "+fmt, i...)
}

// --- empty log ---

type emptyLog struct{}

func (l *emptyLog) Debugf(fmt string, i ...interface{}) {
}

func (l *emptyLog) Infof(fmt string, i ...interface{}) {
}

func (l *emptyLog) Warnf(fmt string, i ...interface{}) {
}

func (l *emptyLog) Errorf(fmt string, i ...interface{}) {
}
