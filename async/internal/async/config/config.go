package config

import (
	"github.com/UnicomAI/wanwu/async/pkg/async/async_component"
	"github.com/UnicomAI/wanwu/async/pkg/async/async_config"
)

type Config struct {
	Log               async_config.Logger
	PendingRun        async_component.IQueue
	PendingDel        async_component.IQueue
	RunMaxConcurrency int
	RunTaskInterval   int // second
}

const (
	TaskHeartbeatInterval int = 30 // second
	RunCheckInterval      int = 1  // second
	DeleteMaxConcurrency  int = 5
	DeleteTaskInterval    int = 3   // second
	CleanTimeout          int = 120 // second
	CleanInterval         int = 60  // second
)
