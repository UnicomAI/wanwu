package orm

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/model"
	"github.com/UnicomAI/wanwu/internal/operate-service/client/orm/sqlopt"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
)

func (c *Client) AddClientRecord(ctx context.Context, clientId string) *err_code.Status {
	// 检查数据库中是否已存在该clientId的记录 [EN] Check whether a record with this clientId already exists in the database
	existingRecord := &model.ClientRecord{}
	nowTs := time.Now().UnixMilli()
	if err := sqlopt.WithClientID(clientId).Apply(c.db).WithContext(ctx).First(existingRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录不存在，创建新记录 [EN] The record does not exist, create a new record
			if err := sqlopt.WithClientID(clientId).Apply(c.db).WithContext(ctx).Create(&model.ClientRecord{
				ClientId:  clientId,
				UpdatedAt: nowTs,
			}).Error; err != nil {
				return toErrStatus("ope_client_record_create", err.Error())
			}
		} else {
			// 其他数据库错误 [EN] Other database errors
			return toErrStatus("ope_client_record_create", err.Error())
		}
	} else {
		// 记录已存在，更新updated_at字段 [EN] The record already exists, update the updated_at field
		if err := c.db.WithContext(ctx).Model(existingRecord).Update("updated_at", nowTs).Error; err != nil {
			return toErrStatus("ope_client_record_create", err.Error())
		}
	}
	return nil
}

func (c *Client) GetClientStatistic(ctx context.Context, startDate, endDate string) (*ClientStatistic, *err_code.Status) {
	if startDate > endDate {
		return nil, toErrStatus("ope_client_statistic_get", fmt.Errorf("startDate %v greater than endDate %v", startDate, endDate).Error())
	}
	// overview
	overview, err := statisticClientOverView(ctx, c.db, startDate, endDate)
	if err != nil {
		return nil, toErrStatus("ope_client_statistic_get", err.Error())
	}
	// trend
	trend, err := statisticClientTrend(ctx, c.db, startDate, endDate)
	if err != nil {
		return nil, toErrStatus("ope_client_statistic_get", err.Error())
	}
	return &ClientStatistic{
		Overview: *overview,
		Trend:    *trend,
	}, nil
}

// --- overview ---

func statisticClientOverView(ctx context.Context, db *gorm.DB, startDate, endDate string) (*ClientOverView, error) {
	cumulativeOverview, err := statisticCumulativeClientOverview(ctx, db, endDate)
	if err != nil {
		return nil, err
	}
	newOverview, err := statisticNewClientOverview(ctx, db, startDate, endDate)
	if err != nil {
		return nil, err
	}
	activeOverview, err := statisticActiveClientOverview(ctx, db, startDate, endDate)
	if err != nil {
		return nil, err
	}
	return &ClientOverView{
		Cumulative: *cumulativeOverview,
		New:        *newOverview,
		Active:     *activeOverview,
	}, nil
}

func statisticCumulativeClientOverview(ctx context.Context, db *gorm.DB, endDate string) (*ClientOverviewItem, error) {
	endTs, err := util.Date2Time(endDate)
	if err != nil {
		return nil, err
	}
	endTs = endTs + 24*time.Hour.Milliseconds()
	// 查询累计client（与时间段无关，所有endTs之前的client总数） [EN] Query the cumulative clients (regardless of the time period, the total number of clients before all endTs)
	var totalCount int64
	if err := db.WithContext(ctx).
		Model(&model.ClientRecord{}).
		Where("created_at < ?", endTs).
		Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("cumulative client overview err: %v", err)
	}
	return &ClientOverviewItem{
		Value: float32(totalCount),
	}, nil
}

// 统计新增client [EN] Statistics new client
func statisticNewClientOverview(ctx context.Context, db *gorm.DB, startDate, endDate string) (*ClientOverviewItem, error) {
	prevPeriod, currPeriod, err := util.PreviousDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	currNewCount, err := statisticNewClient(ctx, db, currPeriod[0], currPeriod[len(currPeriod)-1])
	if err != nil {
		return nil, err
	}
	prevNewCount, err := statisticNewClient(ctx, db, prevPeriod[0], prevPeriod[len(prevPeriod)-1])
	if err != nil {
		return nil, err
	}
	return &ClientOverviewItem{
		Value:            float32(currNewCount),
		PeriodOverPeriod: calculatePoP(float32(currNewCount), float32(prevNewCount)),
	}, nil
}

func statisticNewClient(ctx context.Context, db *gorm.DB, startDate, endDate string) (int64, error) {
	startTs, err := util.Date2Time(startDate)
	if err != nil {
		return 0, err
	}
	endTs, err := util.Date2Time(endDate)
	if err != nil {
		return 0, err
	}
	endTs = endTs + 24*time.Hour.Milliseconds()
	// 查询新增client（创建时间在指定时间段内） [EN] Query new clients (created within the specified time period)
	var newCount int64
	if err := db.WithContext(ctx).
		Model(&model.ClientRecord{}).
		Where("created_at BETWEEN ? AND ?", startTs, endTs).
		Count(&newCount).Error; err != nil {
		return 0, fmt.Errorf("new client overview err: %v", err)
	}
	return newCount, nil
}

// 统计日均活跃client [EN] Statistics of daily active clients
func statisticActiveClientOverview(ctx context.Context, db *gorm.DB, startDate, endDate string) (*ClientOverviewItem, error) {
	prevPeriod, currPeriod, err := util.PreviousDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	currActiveCount, err := statisticActiveClient(ctx, db, currPeriod[0], currPeriod[len(currPeriod)-1])
	if err != nil {
		return nil, err
	}
	prevActiveCount, err := statisticActiveClient(ctx, db, prevPeriod[0], prevPeriod[len(prevPeriod)-1])
	if err != nil {
		return nil, err
	}
	return &ClientOverviewItem{
		Value:            float32(currActiveCount),
		PeriodOverPeriod: calculatePoP(float32(currActiveCount), float32(prevActiveCount)),
	}, nil
}

func statisticActiveClient(ctx context.Context, db *gorm.DB, startDate, endDate string) (float32, error) {
	// 如果时间范围包含今天，则需要先更新今天的活跃用户统计数据 [EN] If the time range includes today, you need to update today's active user statistics first
	startTs, err := util.Date2Time(startDate)
	if err != nil {
		return 0, err
	}
	endTs, err := util.Date2Time(endDate)
	if err != nil {
		return 0, err
	}
	endTs = endTs + 24*time.Hour.Milliseconds()
	nowTs := time.Now().UnixMilli()
	if startTs <= nowTs && nowTs < endTs {
		if err := updateActiveDailyStats(ctx, db, util.Time2Date(nowTs)); err != nil {
			return 0, err
		}
	}
	// 查询活跃client（最后操作时间在指定时间段内） [EN] Query active clients (last operation time is within the specified time period)
	var activeClient model.ClientDailyStats
	if err := sqlopt.SQLOptions(
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
	).Apply(db).WithContext(ctx).Select("SUM(dau_count) as dau_count").First(&activeClient).Error; err != nil {
		return 0, fmt.Errorf("active client overview err: %v", err)
	}
	day := (endTs - startTs) / (24 * time.Hour.Milliseconds())
	dau, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float32(activeClient.DauCount)/float32(day)), 64)
	return float32(dau), nil
}

func updateActiveDailyStats(ctx context.Context, db *gorm.DB, date string) error {
	startTs, err := util.Date2Time(date)
	if err != nil {
		return err
	}
	endTs := startTs + 24*time.Hour.Milliseconds()
	// 查询活跃client（更新时间在指定时间段内） [EN] Query active clients (update time is within the specified time period)
	var activeCount int64
	if err := db.WithContext(ctx).
		Model(&model.ClientRecord{}).
		Where("updated_at BETWEEN ? AND ?", startTs, endTs).
		Count(&activeCount).Error; err != nil {
		return fmt.Errorf("count active client err: %v", err)
	}
	// 更新或插入某一天的活跃统计记录 [EN] Update or insert active statistics records for a certain day
	var existingRecord model.ClientDailyStats
	if err := db.WithContext(ctx).Where("date=?", date).First(&existingRecord).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 记录不存在，创建新记录 [EN] The record does not exist, create a new record
			if err := db.WithContext(ctx).Create(&model.ClientDailyStats{
				Date:     date,
				DauCount: int32(activeCount),
			}).Error; err != nil {
				return fmt.Errorf("create client daily stats err: %v", err)
			}
		} else {
			// 其他数据库错误 [EN] Other database errors
			return err
		}
	} else {
		// 记录已存在，更新dau_count字段 [EN] The record already exists, update the dau_count field
		if err := db.WithContext(ctx).Model(&existingRecord).Updates(map[string]interface{}{
			"dau_count": int32(activeCount),
		}).Error; err != nil {
			return fmt.Errorf("update client daily stats err: %v", err)
		}
	}
	return nil
}

// 计算环比 [EN] Calculate month-on-month
func calculatePoP(current, previous float32) float32 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100 // 避免除以零的错误 [EN] Avoid divide-by-zero errors
	}
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", ((current-previous)/previous)*100), 64)
	return float32(value)
}

// --- trend ---

func statisticClientTrend(ctx context.Context, db *gorm.DB, startDate, endDate string) (*ClientTrend, error) {
	activeTrend, err := statisticActiveClientTrend(ctx, db, startDate, endDate)
	if err != nil {
		return nil, err
	}
	newTrend, err := statisticNewClientTrend(ctx, db, startDate, endDate)
	if err != nil {
		return nil, err
	}
	startTs, err := util.Date2Time(startDate)
	if err != nil {
		return nil, err
	}
	startCumulative, err := statisticCumulativeClientOverview(ctx, db, util.Time2Date(startTs-24*time.Hour.Milliseconds()))
	if err != nil {
		return nil, err
	}
	return &ClientTrend{
		Client: StatisticChart{
			Name: "ope_statistic_client_table",
			Lines: []StatisticChartLine{
				statisticCumulativeClientTrend(startCumulative.Value, newTrend.Items),
				*activeTrend,
				*newTrend,
			},
		},
	}, nil
}

func statisticCumulativeClientTrend(startCumulative float32, newClients []StatisticChartLineItem) StatisticChartLine {
	var values []StatisticChartLineItem
	cumulative := startCumulative
	for _, item := range newClients {
		cumulative += item.Value
		values = append(values, StatisticChartLineItem{
			Key:   item.Key,
			Value: cumulative,
		})
	}
	return StatisticChartLine{
		Name:  "ope_statistic_cumulative_client_line",
		Items: values,
	}
}

func statisticActiveClientTrend(ctx context.Context, db *gorm.DB, startDate, endDate string) (*StatisticChartLine, error) {
	startTs, err := util.Date2Time(startDate)
	if err != nil {
		return nil, err
	}
	endTs, err := util.Date2Time(endDate)
	if err != nil {
		return nil, err
	}
	var stats []*model.ClientDailyStats
	if err := sqlopt.SQLOptions(
		sqlopt.StartDate(startDate),
		sqlopt.EndDate(endDate),
	).Apply(db.WithContext(ctx)).Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("active client trend err: %v", err)
	}
	var values []StatisticChartLineItem
	for _, date := range util.DateRange(startTs, endTs) {
		exist := false
		for _, stat := range stats {
			if stat.Date == date {
				values = append(values, StatisticChartLineItem{
					Key:   stat.Date,
					Value: float32(stat.DauCount),
				})
				exist = true
				break
			}
		}
		if !exist {
			values = append(values, StatisticChartLineItem{
				Key:   date,
				Value: 0,
			})
		}
	}
	return &StatisticChartLine{
		Name:  "ope_statistic_active_client_line",
		Items: values,
	}, nil
}

func statisticNewClientTrend(ctx context.Context, db *gorm.DB, startDate, endDate string) (*StatisticChartLine, error) {
	startTs, err := util.Date2Time(startDate)
	if err != nil {
		return nil, err
	}
	endTs, err := util.Date2Time(endDate)
	if err != nil {
		return nil, err
	}
	var clients []*model.ClientRecord
	if err := db.WithContext(ctx).
		Model(&model.ClientRecord{}).
		Where("created_at BETWEEN ? AND ?", startTs, endTs+24*time.Hour.Milliseconds()).
		Find(&clients).Error; err != nil {
		return nil, fmt.Errorf("new client trend err: %v", err)
	}
	dateNewCount := make(map[string]int)
	for _, client := range clients {
		date := util.Time2Date(client.CreatedAt)
		dateNewCount[date]++
	}
	var values []StatisticChartLineItem
	for _, date := range util.DateRange(startTs, endTs) {
		values = append(values, StatisticChartLineItem{
			Key:   date,
			Value: float32(dateNewCount[date]),
		})
	}
	return &StatisticChartLine{
		Name:  "ope_statistic_new_client_line",
		Items: values,
	}, nil
}
