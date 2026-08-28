package orm

import (
	"context"
	"time"

	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const (
	cornTaskClientRecordSync = "CornTaskStatisticAtiveClient"
)

var (
	cronManager *CronManager
)

// 定时任务管理器
type CronManager struct {
	ctx  context.Context
	cron *cron.Cron
	db   *gorm.DB
}

// 初始化定时任务
func CronInit(db *gorm.DB) error {
	cronManager = &CronManager{
		ctx:  context.Background(),
		cron: cron.New(),
		db:   db,
	}

	entryID, err := cronManager.cron.AddFunc("*/10 * * * *", executeClientRecordSync)
	if err != nil {
		log.Errorf("register cron task (%v) error: %v", cornTaskClientRecordSync, err)
		return err
	}
	log.Infof("cron task (%v) registered successfully with entry ID: %d", cornTaskClientRecordSync, entryID)

	cronManager.cron.Start()
	return nil
}

// 停止定时任务
func CronStop() {
	if cronManager != nil {
		cronManager.cron.Stop()
		log.Infof("cron tasks stopped")
	}
}

// 执行工作流模板记录同步任务
func executeClientRecordSync() {
	util.PrintPanicStack()

	// 计算活跃客户端数量并存储到新表
	date := util.Time2Date(time.Now().UnixMilli())
	if err := updateActiveDailyStats(cronManager.ctx, cronManager.db, date); err != nil {
		log.Errorf("corn task (%v) calculate active clients error: %v", cornTaskClientRecordSync, err)
		return
	}
}
