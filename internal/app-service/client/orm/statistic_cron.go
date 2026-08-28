package orm

import (
	"context"
	"fmt"
	"time"

	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/internal/app-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/robfig/cron/v3"
	"gorm.io/gorm"
)

const (
	cronTaskStatisticSync = "CronTaskStatisticSync"
)

var statisticCronManager *StatisticCronManager

type StatisticCronManager struct {
	ctx  context.Context
	cron *cron.Cron
	db   *gorm.DB
}

func CronInit(ctx context.Context, db *gorm.DB) error {
	statisticCronManager = &StatisticCronManager{
		ctx:  ctx,
		cron: cron.New(),
		db:   db,
	}

	if err := syncAllStatistics(); err != nil {
		return fmt.Errorf("sync statistics err: %v", err)
	}
	entryID, err := statisticCronManager.cron.AddFunc("0 * * * *", cronSyncAllStatistics) // 每小时整点执行
	if err != nil {
		log.Errorf("register cron task (%v) error: %v", cronTaskStatisticSync, err)
		return err
	}
	log.Infof("cron task (%v) registered with entry ID: %d", cronTaskStatisticSync, entryID)

	statisticCronManager.cron.Start()

	return nil
}

func CronStop() {
	if statisticCronManager != nil {
		statisticCronManager.cron.Stop()
		log.Infof("cron tasks stopped")
	}
}

func cronSyncAllStatistics() {
	defer util.PrintPanicStack()
	if err := syncAllStatistics(); err != nil {
		log.Errorf("execute statistics sync err: %v", err)
	}
}

func syncAllStatistics() error {
	ctx := statisticCronManager.ctx
	db := statisticCronManager.db

	now := time.Now().UnixMilli()
	startTs := now - 30*24*time.Hour.Milliseconds()
	dates := util.DateRange(startTs, now)

	for i := len(dates) - 1; i >= 0; i-- {
		date := dates[i]
		// 检查 V2 聚合表是否已有该日期的统计数据；若存在则历史已同步，无需继续向前回填。
		// 检查失败时不参与 early-stop，避免 DB 抖动被当成「无数据」或误停回填。
		hasModel, errModel := checkModelStatsRecordExists(ctx, db, date)
		hasApp, errApp := checkAppStatisticV2RecordExists(ctx, db, date)
		hasApi, errApi := checkAPIKeyStatisticV2RecordExists(ctx, db, date)
		if errModel != nil {
			log.Errorf("check model stats exists date %v err: %v", date, errModel)
		}
		if errApp != nil {
			log.Errorf("check app stats exists date %v err: %v", date, errApp)
		}
		if errApi != nil {
			log.Errorf("check api key stats exists date %v err: %v", date, errApi)
		}
		if err := syncStatisticModelStats(ctx, date, db); err != nil {
			log.Errorf("update statistic model stats date %v err: %v", date, err)
		}
		if err := syncAPIKeyStatisticV2Stats(ctx, date, db); err != nil {
			log.Errorf("update api key statistic v2 date %v err: %v", date, err)
		}
		if err := syncAppStatisticV2Stats(ctx, date, db); err != nil {
			log.Errorf("update app statistic v2 date %v err: %v", date, err)
		}
		if errModel == nil && errApp == nil && errApi == nil && hasModel && hasApp && hasApi {
			log.Infof("found existing record for date %v, stop backward sync", date)
			break
		}
	}
	return nil
}

func checkModelStatsRecordExists(ctx context.Context, db *gorm.DB, date string) (bool, error) {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithDate(date),
	).Apply(db.WithContext(ctx)).Model(&model.StatisticModel{}).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check model stats record exists for date %v err: %v", date, err)
	}
	return count > 0, nil
}

func checkAppStatisticV2RecordExists(ctx context.Context, db *gorm.DB, date string) (bool, error) {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithDate(date),
	).Apply(db.WithContext(ctx)).Model(&model.StatisticApp{}).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check app statistic v2 record exists for date %v err: %v", date, err)
	}
	return count > 0, nil
}

func checkAPIKeyStatisticV2RecordExists(ctx context.Context, db *gorm.DB, date string) (bool, error) {
	var count int64
	if err := sqlopt.SQLOptions(
		sqlopt.WithDate(date),
	).Apply(db.WithContext(ctx)).Model(&model.StatisticApiKey{}).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check api key statistic v2 record exists for date %v err: %v", date, err)
	}
	return count > 0, nil
}
