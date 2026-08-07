package orm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/app-service/client/model"
	"github.com/UnicomAI/wanwu/pkg/constant"
	"github.com/UnicomAI/wanwu/pkg/log"
	"github.com/UnicomAI/wanwu/pkg/redis"
	"github.com/UnicomAI/wanwu/pkg/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const luaUpdateAppStatsV2 = `
local key = KEYS[1]
local field = KEYS[2]
local delta = cjson.decode(ARGV[1])
local expire = tonumber(ARGV[2])

local current = redis.call('HGET', key, field)

if current then
	local record = cjson.decode(current)
	record.callCount = record.callCount + delta.callCount
	record.callFailure = record.callFailure + delta.callFailure
	record.streamCount = record.streamCount + delta.streamCount
	record.nonStreamCount = record.nonStreamCount + delta.nonStreamCount
	record.streamFailure = record.streamFailure + delta.streamFailure
	record.nonStreamFailure = record.nonStreamFailure + delta.nonStreamFailure
	record.firstTokenLatency = record.firstTokenLatency + delta.firstTokenLatency
	record.costs = record.costs + delta.costs
	redis.call('HSET', key, field, cjson.encode(record))
else
	redis.call('HSET', key, field, ARGV[1])
end

redis.call('EXPIRE', key, expire)
return 1
`

type AppRecordV2Stats struct {
	Source              string `json:"source"`
	Module              string `json:"module"`
	AppType             string `json:"appType"`
	ModuleCreatorUserID string `json:"moduleCreatorUserId"`
	ModuleCreatorOrgID  string `json:"moduleCreatorOrgId"`
	CallCount           int32  `json:"callCount"`
	CallFailure         int32  `json:"callFailure"`
	StreamCount         int32  `json:"streamCount"`
	NonStreamCount      int32  `json:"nonStreamCount"`
	StreamFailure       int32  `json:"streamFailure"`
	NonStreamFailure    int32  `json:"nonStreamFailure"`
	FirstTokenLatency   int64  `json:"firstTokenLatency"`
	Costs               int64  `json:"costs"`
}

func getRedisAppStatsV2Key(date string) string {
	return fmt.Sprintf("statistic|app|%s", date)
}

func getRedisAppStatsV2Field(userID, orgID, source, module, appID, appType, moduleCreatorUserID, moduleCreatorOrgID string) string {
	return strings.Join([]string{userID, orgID, source, module, appID, appType, moduleCreatorUserID, moduleCreatorOrgID}, "|")
}

func parseRedisAppStatsV2Field(field string) (userID, orgID, source, module, appID, appType, moduleCreatorUserID, moduleCreatorOrgID string, ok bool) {
	parts := strings.Split(field, "|")
	if len(parts) != 8 {
		return "", "", "", "", "", "", "", "", false
	}
	return parts[0], parts[1], parts[2], parts[3], parts[4], parts[5], parts[6], parts[7], true
}

func getRedisAppStatsV2Value(value string) (*AppRecordV2Stats, error) {
	record := &AppRecordV2Stats{}
	if err := json.Unmarshal([]byte(value), record); err != nil {
		return nil, fmt.Errorf("unmarshal app v2 value %v err: %v", value, err)
	}
	return record, nil
}

// RecordAppStatisticV2 记录应用统计 V2（先 Redis 日聚合，再写 app_record 明细）。
// wga/model/skill/knowledge/prompt（板块级）允许 AppId 为空；其余 module 仍要求 AppId。
func (c *Client) RecordAppStatisticV2(ctx context.Context, req *RecordAppStatisticV2Input) *errs.Status {
	if req == nil {
		return toErrStatus("app_statistic_v2_record", "invalid request")
	}
	if !constant.StatisticModuleAllowsEmptyAppID(req.Module) && req.AppID == "" {
		return toErrStatus("app_statistic_v2_record", "invalid request: appId required")
	}
	if err := recordAppStatisticV2ToRedis(ctx, req); err != nil {
		log.Errorf("record app statistic v2 to redis appId=%s err: %v", req.AppID, err)
	}
	if err := c.db.WithContext(ctx).Create(&model.AppRecordV2{
		TraceID:             req.TraceID,
		OrgID:               req.OrgID,
		UserID:              req.UserID,
		Source:              req.Source,
		Module:              req.Module,
		ModuleCreatorUserID: req.ModuleCreatorUserID,
		ModuleCreatorOrgID:  req.ModuleCreatorOrgID,
		AppID:               req.AppID,
		AppType:             req.AppType,
		IsStream:            req.IsStream,
		Costs:               req.Costs,
		FirstTokenLatency:   req.FirstTokenLatency,
		StatusCode:          req.StatusCode,
		RequestBody:         req.RequestBody,
		ResponseBody:        req.ResponseBody,
		FinishReason:        req.FinishReason,
		FailureReason:       req.FailureReason,
		Question:            req.Question,
		Answer:              req.Answer,
	}).Error; err != nil {
		log.Errorf("create app record appId=%s err: %v", req.AppID, err)
	}
	return nil
}

func recordAppStatisticV2ToRedis(ctx context.Context, req *RecordAppStatisticV2Input) error {
	today := util.Time2Date(time.Now().UnixMilli())
	key := getRedisAppStatsV2Key(today)
	field := getRedisAppStatsV2Field(
		req.UserID, req.OrgID, req.Source, req.Module,
		req.AppID, req.AppType, req.ModuleCreatorUserID, req.ModuleCreatorOrgID,
	)

	callFailure, streamCount, streamFailure, nonStreamCount, nonStreamFailure := buildCallCounters(req.StatusCode, req.IsStream)
	ftl, costs := statisticCostsForAvg(req.StatusCode, req.FirstTokenLatency, req.Costs)
	delta := &AppRecordV2Stats{
		Source:              req.Source,
		Module:              req.Module,
		AppType:             req.AppType,
		ModuleCreatorUserID: req.ModuleCreatorUserID,
		ModuleCreatorOrgID:  req.ModuleCreatorOrgID,
		CallCount:           1,
		CallFailure:         int32(callFailure),
		StreamCount:         int32(streamCount),
		StreamFailure:       int32(streamFailure),
		NonStreamCount:      int32(nonStreamCount),
		NonStreamFailure:    int32(nonStreamFailure),
		FirstTokenLatency:   ftl,
		Costs:               costs,
	}
	deltaJSON, err := json.Marshal(delta)
	if err != nil {
		return fmt.Errorf("marshal app statistic v2 delta err: %v", err)
	}
	_, err = redis.App().Eval(ctx, luaUpdateAppStatsV2, []string{key, field}, string(deltaJSON), redisStatsExpireSeconds)
	if err != nil {
		return fmt.Errorf("redis eval app statistic v2 err: %v", err)
	}
	return nil
}

func syncAppStatisticV2Stats(ctx context.Context, date string, db *gorm.DB) error {
	key := getRedisAppStatsV2Key(date)
	resultMap, err := redis.App().HGetAll(ctx, key)
	if err != nil {
		return fmt.Errorf("redis HGetAll key %v failed: %v", key, err)
	}
	for _, item := range resultMap {
		userID, orgID, source, module, appID, appType, moduleCreatorUserID, moduleCreatorOrgID, ok := parseRedisAppStatsV2Field(item.K)
		if !ok {
			log.Errorf("parse app statistic v2 field %v failed", item.K)
			continue
		}
		record, err := getRedisAppStatsV2Value(item.V)
		if err != nil {
			log.Errorf("get app statistic v2 item %v err: %v", item.K, err)
			continue
		}
		if err := upsertAppStatisticV2ByRecord(ctx, db, userID, orgID, source, module, appID, appType, moduleCreatorUserID, moduleCreatorOrgID, date, record); err != nil {
			log.Errorf("upsert app statistic v2 date %v field %v err: %v", date, item.K, err)
		}
	}
	return nil
}

func upsertAppStatisticV2ByRecord(ctx context.Context, db *gorm.DB, userID, orgID, source, module, appID, appType, moduleCreatorUserID, moduleCreatorOrgID, date string, record *AppRecordV2Stats) error {
	stat := &model.StatisticApp{
		OrgID:               orgID,
		UserID:              userID,
		Source:              source,
		Module:              module,
		AppID:               appID,
		AppType:             appType,
		ModuleCreatorUserID: moduleCreatorUserID,
		ModuleCreatorOrgID:  moduleCreatorOrgID,
		Date:                date,
		CallCount:           record.CallCount,
		CallFailure:         record.CallFailure,
		StreamCount:         record.StreamCount,
		NonStreamCount:      record.NonStreamCount,
		StreamFailure:       record.StreamFailure,
		NonStreamFailure:    record.NonStreamFailure,
		FirstTokenLatency:   record.FirstTokenLatency,
		Costs:               record.Costs,
	}
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "org_id"},
			{Name: "user_id"},
			{Name: "source"},
			{Name: "module"},
			{Name: "app_id"},
			{Name: "app_type"},
			{Name: "module_creator_user_id"},
			{Name: "module_creator_org_id"},
			{Name: "date"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"call_count", "call_failure", "stream_count", "non_stream_count",
			"stream_failure", "non_stream_failure",
			"first_token_latency", "costs",
		}),
	}).Create(stat).Error
}
