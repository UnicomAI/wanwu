package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	operate_service "github.com/UnicomAI/wanwu/api/proto/operate-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

func GetClientStatistic(ctx *gin.Context, startDate, endDate string) (*response.ClientStatistic, error) {
	// client
	statistic, err := getClientStatistic(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	// Views
	browseOverview, browseTrend, err := getGlobalBrowseStatistic(ctx, startDate, endDate)
	if err != nil {
		return nil, err
	}
	statistic.Overview.Browse = *browseOverview
	statistic.Trend.Browse = *browseTrend

	return &response.ClientStatistic{
		Overview: statistic.Overview,
		Trend:    statistic.Trend,
	}, nil
}

func getClientStatistic(ctx *gin.Context, startDate, endDate string) (*response.ClientStatistic, error) {
	resp, err := operate.GetClientStatistic(ctx, &operate_service.GetClientStatisticReq{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		return nil, err
	}
	return &response.ClientStatistic{
		Overview: response.ClientOverView{
			CumulativeClient: clientOverviewPb2resp(resp.Overview.GetCumulative()),
			AdditionClient:   clientOverviewPb2resp(resp.Overview.GetNew()),
			ActiveClient:     clientOverviewPb2resp(resp.Overview.GetActive()),
		},
		Trend: response.ClientTrend{
			Client: convertStatisticChart(ctx, resp.Trend.Client),
		},
	}, nil
}

func clientOverviewPb2resp(item *operate_service.ClientOverviewItem) response.StatisticOverviewItem {
	value, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", item.Value), 64)
	return response.StatisticOverviewItem{
		Value:            float32(value),
		PeriodOverPeriod: item.PeriodOverperiod,
	}
}

// --- global browse statistic ---

func recordGlobalBrowse(ctx context.Context) error {
	// Use HINCRBY atomicity to increase template downloads
	date := util.Time2Date(time.Now().UnixMilli())
	err := redis.OP().Cli().HIncrBy(ctx, redisGlobalBrowseKey, date, 1).Err()
	if err != nil {
		return fmt.Errorf("redis IncrBy key %v filed %v err: %v", redisGlobalBrowseKey, date, err)
	}
	return nil
}

func getGlobalBrowseStatistic(ctx *gin.Context, startDate, endDate string) (*response.StatisticOverviewItem, *response.StatisticChart, error) {
	// Get a list of dates for the current period and the previous period
	prevDates, currentDates, err := util.PreviousDateRange(startDate, endDate)
	if err != nil {
		return nil, nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_global_browse_stats", fmt.Sprintf("get date range error: %v", err))
	}

	// Get browsing data
	currentBrowseData, err := getBrowseDataFromRedis(ctx.Request.Context(), currentDates)
	if err != nil {
		return nil, nil, err
	}
	prevBrowseData, err := getBrowseDataFromRedis(ctx.Request.Context(), prevDates)
	if err != nil {
		return nil, nil, err
	}

	// Calculate overview data
	overview := calculateGlobalBrowseOverview(currentBrowseData, prevBrowseData)

	// Calculate trend data
	trend := calculateGlobalBrowseTrend(ctx, currentBrowseData, currentDates)

	return &overview, &trend, nil
}

// Get browsing data for multiple dates from Redis
func getBrowseDataFromRedis(ctx context.Context, dates []string) (map[string]int64, error) {
	items, err := redis.OP().HGetAll(ctx, redisGlobalBrowseKey)
	if err != nil {
		return nil, grpc_util.ErrorStatusWithKey(errs.Code_BFFGeneral, "bff_global_browse_stats", fmt.Sprintf("redis HGetAll key %v fields %v err: %v", redisGlobalBrowseKey, dates, err))
	}

	data := make(map[string]int64)
	if len(items) == 0 {
		return data, nil
	}
	for _, date := range dates {
		for _, item := range items {
			if item.K == date {
				data[date] = util.MustI64(item.V)
				break
			}
		}
		// If there is no data for a certain date, the default value is 0
		if _, exist := data[date]; !exist {
			data[date] = 0
		}
	}

	return data, nil
}

// Calculate overview data
func calculateGlobalBrowseOverview(currentData, prevData map[string]int64) response.StatisticOverviewItem {
	// Calculate the total number of views in the current period
	var currentTotal int64
	for _, count := range currentData {
		currentTotal += count
	}

	// Calculate the total number of pageviews in the previous cycle
	var prevTotal int64
	for _, count := range prevData {
		prevTotal += count
	}

	// Calculate month-on-month
	var pop float64
	if prevTotal > 0 {
		pop, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", (float32(currentTotal)-float32(prevTotal))/float32(prevTotal)*100), 64)
	} else if currentTotal > 0 {
		// If the previous period is 0 and there is data in this period, the growth rate is 100%.
		pop = 100
	}

	return response.StatisticOverviewItem{
		Value:            float32(currentTotal),
		PeriodOverPeriod: float32(pop),
	}
}

// Calculate trend data
func calculateGlobalBrowseTrend(ctx *gin.Context, browseData map[string]int64, dates []string) response.StatisticChart {
	var items []response.StatisticChartLineItem
	for _, date := range dates {
		count := browseData[date]
		items = append(items, response.StatisticChartLineItem{
			Key:   date,
			Value: float32(count),
		})
	}
	return response.StatisticChart{
		TableName: gin_util.I18nKey(ctx, "ope_statistic_browse_table"),
		Lines: []response.StatisticChartLine{
			{
				LineName: gin_util.I18nKey(ctx, "ope_statistic_browse_line"),
				Items:    items,
			},
		},
	}
}
