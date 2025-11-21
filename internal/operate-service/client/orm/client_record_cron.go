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

// Scheduled task manager
type CronManager struct {
	ctx  context.Context
	cron *cron.Cron
	db   *gorm.DB
}

// Initialize scheduled tasks
func CronInit(db *gorm.DB) error {
	cronManager = &CronManager{
		ctx:  context.TODO(),
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

// Stop scheduled tasks
func CronStop() {
	if cronManager != nil {
		cronManager.cron.Stop()
		log.Infof("cron tasks stopped")
	}
}

// Execute workflow template record synchronization task
func executeClientRecordSync() {
	util.PrintPanicStack()

	// Count the number of active clients and store it in a new table
	date := util.Time2Date(time.Now().UnixMilli())
	if err := updateActiveDailyStats(cronManager.ctx, cronManager.db, date); err != nil {
		log.Errorf("corn task (%v) calculate active clients error: %v", cornTaskClientRecordSync, err)
		return
	}
}
